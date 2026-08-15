package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
)

type FormRepo struct {
	db *sql.DB
}

func NewFormRepo(db *sql.DB) *FormRepo {
	return &FormRepo{db: db}
}

var _ forms.Repository = (*FormRepo)(nil)

func (r *FormRepo) Create(ctx context.Context, f forms.Form) error {
	const q = `
		INSERT INTO forms (id, user_id, title, description, status, public_slug, published_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, q,
		f.ID, f.UserID, f.Title, f.Description, string(f.Status),
		nullIfEmpty(f.PublicSlug), nullTime(f.PublishedAt),
		formatTime(f.CreatedAt), formatTime(f.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("criar formulário: %w", err)
	}
	return nil
}

// ListOwnedBy traz a lista sem os campos: a tela de listagem não os usa, e
// carregá-los seria uma query por formulário à toa.
func (r *FormRepo) ListOwnedBy(ctx context.Context, ownerID string) ([]forms.Form, error) {
	const q = `
		SELECT f.id, f.user_id, f.title, f.description, f.status, f.public_slug,
		       f.published_at, f.created_at, f.updated_at,
		       (SELECT COUNT(*) FROM submissions s WHERE s.form_id = f.id)
		FROM forms f
		WHERE f.user_id = ?
		ORDER BY f.created_at DESC`

	rows, err := r.db.QueryContext(ctx, q, ownerID)
	if err != nil {
		return nil, fmt.Errorf("listar formulários: %w", err)
	}
	defer rows.Close()

	list := make([]forms.Form, 0)
	for rows.Next() {
		f, err := scanForm(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	// rows.Err() pega o erro que interrompeu a iteração; sem isto, uma falha
	// no meio do cursor viraria "lista curta" em silêncio.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar formulários: %w", err)
	}
	return list, nil
}

// GetOwnedBy é a única forma de alcançar um formulário. O `AND user_id = ?`
// não é uma checagem que o handler pode esquecer: ele está na query.
func (r *FormRepo) GetOwnedBy(ctx context.Context, formID, ownerID string) (forms.Form, error) {
	const q = `
		SELECT f.id, f.user_id, f.title, f.description, f.status, f.public_slug,
		       f.published_at, f.created_at, f.updated_at,
		       (SELECT COUNT(*) FROM submissions s WHERE s.form_id = f.id)
		FROM forms f
		WHERE f.id = ? AND f.user_id = ?`

	form, err := scanForm(r.db.QueryRowContext(ctx, q, formID, ownerID))
	if err != nil {
		return forms.Form{}, err
	}

	form.Fields, err = r.fieldsOf(ctx, form.ID)
	if err != nil {
		return forms.Form{}, err
	}
	return form, nil
}

func (r *FormRepo) UpdateOwnedBy(ctx context.Context, f forms.Form) error {
	const q = `
		UPDATE forms
		SET title = ?, description = ?, status = ?, public_slug = ?, published_at = ?, updated_at = ?
		WHERE id = ? AND user_id = ?`

	res, err := r.db.ExecContext(ctx, q,
		f.Title, f.Description, string(f.Status),
		nullIfEmpty(f.PublicSlug), nullTime(f.PublishedAt), formatTime(f.UpdatedAt),
		f.ID, f.UserID,
	)
	if err != nil {
		return fmt.Errorf("atualizar formulário: %w", err)
	}
	return requireOneRow(res, "atualizar formulário")
}

func (r *FormRepo) DeleteOwnedBy(ctx context.Context, formID, ownerID string) error {
	// Campos, submissões e respostas somem por ON DELETE CASCADE — que só
	// funciona porque o pragma foreign_keys está ligado no DSN.
	const q = `DELETE FROM forms WHERE id = ? AND user_id = ?`

	res, err := r.db.ExecContext(ctx, q, formID, ownerID)
	if err != nil {
		return fmt.Errorf("remover formulário: %w", err)
	}
	return requireOneRow(res, "remover formulário")
}

// ReplaceFields apaga tudo e reinsere, numa transação.
//
// A transação é o que torna o "substitui a lista inteira" seguro: se a
// inserção falhar no meio, o rollback devolve os campos antigos. Sem ela, um
// erro deixaria o formulário sem campo nenhum.
func (r *FormRepo) ReplaceFields(ctx context.Context, formID string, fields []forms.Field) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	// Rollback depois de um Commit bem-sucedido é no-op, então este defer é
	// seguro e cobre todo caminho de erro abaixo.
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM form_fields WHERE form_id = ?`, formID); err != nil {
		return fmt.Errorf("limpar campos: %w", err)
	}

	const insert = `
		INSERT INTO form_fields (id, form_id, type, label, help_text, required, position, config)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	for _, f := range fields {
		config, err := json.Marshal(f.Config)
		if err != nil {
			return fmt.Errorf("serializar config do campo %q: %w", f.Label, err)
		}
		_, err = tx.ExecContext(ctx, insert,
			f.ID, formID, string(f.Type), f.Label, f.HelpText, f.Required, f.Position, string(config),
		)
		if err != nil {
			return fmt.Errorf("inserir campo %q: %w", f.Label, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar transação: %w", err)
	}
	return nil
}

func (r *FormRepo) fieldsOf(ctx context.Context, formID string) ([]forms.Field, error) {
	const q = `
		SELECT id, form_id, type, label, help_text, required, position, config
		FROM form_fields WHERE form_id = ? ORDER BY position`

	rows, err := r.db.QueryContext(ctx, q, formID)
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

// scanner cobre *sql.Row e *sql.Rows, que têm o mesmo Scan mas não uma
// interface comum na biblioteca padrão.
type scanner interface {
	Scan(dest ...any) error
}

func scanForm(s scanner) (forms.Form, error) {
	var (
		f                    forms.Form
		status               string
		slug                 sql.NullString
		publishedAt          sql.NullString
		createdAt, updatedAt string
	)

	err := s.Scan(&f.ID, &f.UserID, &f.Title, &f.Description, &status, &slug,
		&publishedAt, &createdAt, &updatedAt, &f.SubmissionCount)
	if errors.Is(err, sql.ErrNoRows) {
		// Não encontrado e não é seu são indistinguíveis aqui de propósito: o
		// handler devolve 404 nos dois casos, e 404 não confirma existência.
		return forms.Form{}, forms.ErrFormNotFound
	}
	if err != nil {
		return forms.Form{}, fmt.Errorf("ler formulário: %w", err)
	}

	f.Status = forms.Status(status)
	f.PublicSlug = slug.String
	if publishedAt.Valid {
		if t, err := time.Parse(time.RFC3339, publishedAt.String); err == nil {
			f.PublishedAt = &t
		}
	}
	f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	f.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return f, nil
}

// requireOneRow transforma "não afetou nada" em ErrFormNotFound. Sem isto, um
// UPDATE num formulário de outro usuário seria um sucesso silencioso que não
// mudou nada, e o cliente receberia 200.
func requireOneRow(res sql.Result, acao string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", acao, err)
	}
	if n == 0 {
		return forms.ErrFormNotFound
	}
	return nil
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTime(*t)
}
