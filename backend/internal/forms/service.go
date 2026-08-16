package forms

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Repository é a fronteira com a persistência.
//
// Não existe `GetForm(ctx, formID)`: todo método que alcança um formulário
// exige o dono, e a query carrega `WHERE id = ? AND user_id = ?`. A checagem
// de posse não pode ser esquecida porque não dá para chamar o método sem ela.
type Repository interface {
	Create(ctx context.Context, f Form) error
	ListOwnedBy(ctx context.Context, ownerID string) ([]Form, error)
	GetOwnedBy(ctx context.Context, formID, ownerID string) (Form, error)
	UpdateOwnedBy(ctx context.Context, f Form) error
	DeleteOwnedBy(ctx context.Context, formID, ownerID string) error
	ReplaceFields(ctx context.Context, formID string, fields []Field) error
}

type Service struct {
	repo Repository
	// Base do link público, vinda da configuração.
	publicBaseURL string
}

func NewService(repo Repository, publicBaseURL string) *Service {
	return &Service{repo: repo, publicBaseURL: strings.TrimSuffix(publicBaseURL, "/")}
}

// PublicURL devolve o link do formulário, ou vazio se ainda não publicado.
func (s *Service) PublicURL(f Form) string {
	if f.PublicSlug == "" {
		return ""
	}
	return s.publicBaseURL + "/f/" + f.PublicSlug
}

func (s *Service) List(ctx context.Context, ownerID string) ([]Form, error) {
	return s.repo.ListOwnedBy(ctx, ownerID)
}

func (s *Service) Get(ctx context.Context, formID, ownerID string) (Form, error) {
	return s.repo.GetOwnedBy(ctx, formID, ownerID)
}

func (s *Service) Create(ctx context.Context, ownerID, title, description string) (Form, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Form{}, ErrTitleRequired
	}

	now := time.Now().UTC()
	form := Form{
		ID:          uuid.NewString(),
		UserID:      ownerID,
		Title:       title,
		Description: strings.TrimSpace(description),
		Status:      StatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, form); err != nil {
		return Form{}, err
	}
	return form, nil
}

// Update altera título e descrição. Ponteiro nil significa "não mexer", que é
// diferente de "definir como vazio".
func (s *Service) Update(ctx context.Context, formID, ownerID string, title, description *string) (Form, error) {
	form, err := s.repo.GetOwnedBy(ctx, formID, ownerID)
	if err != nil {
		return Form{}, err
	}

	if title != nil {
		novoTitulo := strings.TrimSpace(*title)
		if novoTitulo == "" {
			return Form{}, ErrTitleRequired
		}
		form.Title = novoTitulo
	}
	if description != nil {
		form.Description = strings.TrimSpace(*description)
	}
	form.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateOwnedBy(ctx, form); err != nil {
		return Form{}, err
	}
	return form, nil
}

func (s *Service) Delete(ctx context.Context, formID, ownerID string) error {
	return s.repo.DeleteOwnedBy(ctx, formID, ownerID)
}

// ReplaceFields substitui a lista inteira de campos: o array recebido é a
// verdade e o índice é a posição.
//
// Os IDs dos campos mudam a cada save, o que é irrelevante porque a operação
// só é permitida em draft, quando ainda não há respostas apontando para eles.
func (s *Service) ReplaceFields(ctx context.Context, formID, ownerID string, inputs []FieldInput) ([]Field, error) {
	form, err := s.repo.GetOwnedBy(ctx, formID, ownerID)
	if err != nil {
		return nil, err
	}

	// Editar a estrutura com respostas existentes deixaria resposta órfã,
	// invalidaria respostas antigas ou mudaria o tipo de um valor já gravado.
	if form.IsPublished() {
		return nil, ErrFormPublished
	}

	if err := ValidateFieldDefinitions(inputs); err != nil {
		return nil, err
	}

	fields := make([]Field, 0, len(inputs))
	for i, in := range inputs {
		fields = append(fields, Field{
			ID:       uuid.NewString(),
			FormID:   form.ID,
			Type:     in.Type,
			Label:    strings.TrimSpace(in.Label),
			HelpText: strings.TrimSpace(in.HelpText),
			Required: in.Required,
			Position: i,
			Config:   normalizeConfig(in.Type, in.Config),
		})
	}

	if err := s.repo.ReplaceFields(ctx, form.ID, fields); err != nil {
		return nil, err
	}
	return fields, nil
}

// Publish gera o slug na primeira publicação e o preserva nas seguintes —
// regenerar mataria todo link já distribuído.
func (s *Service) Publish(ctx context.Context, formID, ownerID string) (Form, error) {
	form, err := s.repo.GetOwnedBy(ctx, formID, ownerID)
	if err != nil {
		return Form{}, err
	}

	if len(form.Fields) == 0 {
		return Form{}, ErrNoFields
	}

	if form.PublicSlug == "" {
		slug, err := NewSlug()
		if err != nil {
			return Form{}, err
		}
		form.PublicSlug = slug
	}

	now := time.Now().UTC()
	form.Status = StatusPublished
	form.PublishedAt = &now
	form.UpdatedAt = now

	if err := s.repo.UpdateOwnedBy(ctx, form); err != nil {
		return Form{}, err
	}
	return form, nil
}

// Unpublish volta para draft mantendo o slug, para republicar reviver o mesmo
// link.
func (s *Service) Unpublish(ctx context.Context, formID, ownerID string) (Form, error) {
	form, err := s.repo.GetOwnedBy(ctx, formID, ownerID)
	if err != nil {
		return Form{}, err
	}

	form.Status = StatusDraft
	form.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateOwnedBy(ctx, form); err != nil {
		return Form{}, err
	}
	return form, nil
}

// normalizeConfig descarta a configuração que não faz sentido para o tipo.
func normalizeConfig(t FieldType, c FieldConfig) FieldConfig {
	switch t {
	case FieldShortText, FieldLongText:
		return FieldConfig{MinLength: c.MinLength, MaxLength: c.MaxLength}
	case FieldNumber:
		return FieldConfig{Min: c.Min, Max: c.Max}
	case FieldSelect:
		return FieldConfig{Options: c.Options}
	default:
		return FieldConfig{}
	}
}
