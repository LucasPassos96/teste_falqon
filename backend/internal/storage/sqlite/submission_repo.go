package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
)

func newID() string { return uuid.NewString() }

type SubmissionRepo struct {
	db *sql.DB
}

func NewSubmissionRepo(db *sql.DB) *SubmissionRepo {
	return &SubmissionRepo{db: db}
}

var _ forms.PublicRepository = (*SubmissionRepo)(nil)

// GetPublishedBySlug exige status = 'published' na query, então formulário em
// draft cai no mesmo ErrFormNotFound de inexistente.
func (r *SubmissionRepo) GetPublishedBySlug(ctx context.Context, slug string) (forms.Form, error) {
	const query = `
		SELECT id, user_id, title, description, status, public_slug, published_at, created_at, updated_at, 0
		FROM forms
		WHERE public_slug = ? AND status = 'published'`

	form, err := scanForm(r.db.QueryRowContext(ctx, query, slug))
	if err != nil {
		return forms.Form{}, err
	}

	form.Fields, err = r.fieldsOf(ctx, form.ID)
	if err != nil {
		return forms.Form{}, err
	}
	return form, nil
}

// SaveSubmission grava a resposta e seus valores numa transação.
func (r *SubmissionRepo) SaveSubmission(ctx context.Context, submissao forms.Submission) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO submissions (id, form_id, submitted_at) VALUES (?, ?, ?)`,
		submissao.ID, submissao.FormID, formatTime(submissao.SubmittedAt),
	)
	if err != nil {
		return fmt.Errorf("criar submissão: %w", err)
	}

	const insert = `
		INSERT INTO submission_answers (id, submission_id, field_id, field_label, value)
		VALUES (?, ?, ?, ?, ?)`

	for _, resposta := range submissao.Answers {
		// field_label é gravado agora, não lido por join depois: é o snapshot
		// do que o visitante viu.
		if _, err := tx.ExecContext(ctx, insert,
			newID(), submissao.ID, resposta.FieldID, resposta.FieldLabel, resposta.Value); err != nil {
			return fmt.Errorf("gravar resposta do campo %s: %w", resposta.FieldID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar transação: %w", err)
	}
	return nil
}

// ListSubmissions devolve resposta página e o total, que vem de query separada porque
// COUNT com LIMIT contaria só resposta página.
func (r *SubmissionRepo) ListSubmissions(ctx context.Context, formID string, limit, offset int) ([]forms.Submission, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM submissions WHERE form_id = ?`, formID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar respostas: %w", err)
	}

	const query = `
		SELECT id, submitted_at
		FROM submissions
		WHERE form_id = ?
		ORDER BY submitted_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, formID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listar respostas: %w", err)
	}
	defer rows.Close()

	list := make([]forms.Submission, 0)
	ids := make([]string, 0)
	for rows.Next() {
		var (
			submissao   forms.Submission
			submittedAt string
		)
		if err := rows.Scan(&submissao.ID, &submittedAt); err != nil {
			return nil, 0, fmt.Errorf("ler resposta: %w", err)
		}
		submissao.FormID = formID
		submissao.SubmittedAt, _ = time.Parse(time.RFC3339, submittedAt)
		submissao.Answers = make([]forms.Answer, 0)
		list = append(list, submissao)
		ids = append(ids, submissao.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterar respostas: %w", err)
	}

	if err := r.attachAnswers(ctx, list, ids); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// attachAnswers busca as respostas da página inteira numa query só, evitando
// o N+1.
func (r *SubmissionRepo) attachAnswers(ctx context.Context, list []forms.Submission, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Os placeholders são montados pela quantidade de ids; os valores seguem
	// bindados.
	placeholders := make([]byte, 0, len(ids)*2)
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
		args = append(args, id)
	}

	query := `SELECT submission_id, field_id, field_label, value
	      FROM submission_answers
	      WHERE submission_id IN (` + string(placeholders) + `)`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("listar valores das respostas: %w", err)
	}
	defer rows.Close()

	porSubmissao := make(map[string][]forms.Answer, len(ids))
	for rows.Next() {
		var (
			submissionID string
			resposta     forms.Answer
		)
		if err := rows.Scan(&submissionID, &resposta.FieldID, &resposta.FieldLabel, &resposta.Value); err != nil {
			return fmt.Errorf("ler valor: %w", err)
		}
		porSubmissao[submissionID] = append(porSubmissao[submissionID], resposta)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterar valores: %w", err)
	}

	for i := range list {
		if answers, ok := porSubmissao[list[i].ID]; ok {
			list[i].Answers = answers
		}
	}
	return nil
}

func (r *SubmissionRepo) fieldsOf(ctx context.Context, formID string) ([]forms.Field, error) {
	const query = `
		SELECT id, form_id, type, label, help_text, required, position, config
		FROM form_fields WHERE form_id = ? ORDER BY position`

	rows, err := r.db.QueryContext(ctx, query, formID)
	if err != nil {
		return nil, fmt.Errorf("listar campos: %w", err)
	}
	defer rows.Close()

	list := make([]forms.Field, 0)
	for rows.Next() {
		var (
			f          forms.Field
			fieldType  string
			configJSON string
		)
		if err := rows.Scan(&f.ID, &f.FormID, &fieldType, &f.Label, &f.HelpText,
			&f.Required, &f.Position, &configJSON); err != nil {
			return nil, fmt.Errorf("ler campo: %w", err)
		}
		f.Type = forms.FieldType(fieldType)
		if err := json.Unmarshal([]byte(configJSON), &f.Config); err != nil {
			return nil, fmt.Errorf("interpretar config do campo %s: %w", f.ID, err)
		}
		list = append(list, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar campos: %w", err)
	}
	return list, nil
}
