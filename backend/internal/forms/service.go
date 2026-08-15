package forms

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Repository é a fronteira com a persistência.
//
// Repare no que NÃO existe aqui: não há `GetForm(ctx, formID)`. Todo método
// que alcança um formulário exige o dono, e a query correspondente carrega
// `WHERE id = ? AND user_id = ?`.
//
// IDOR — trocar o ID na URL e ler o dado de outra pessoa — é a falha mais
// comum nesse tipo de aplicação, e acontece porque alguém esqueceu um `if` num
// handler entre vinte. Se o método simplesmente não pode ser chamado sem o
// dono, não há o que esquecer: o compilador vira o revisor.
type Repository interface {
	Create(ctx context.Context, f Form) error
	ListOwnedBy(ctx context.Context, ownerID string) ([]Form, error)
	GetOwnedBy(ctx context.Context, formID, ownerID string) (Form, error)
	UpdateOwnedBy(ctx context.Context, f Form) error
	DeleteOwnedBy(ctx context.Context, formID, ownerID string) error
	// ReplaceFields apaga e reinsere os campos numa transação só.
	ReplaceFields(ctx context.Context, formID string, fields []Field) error
}

type Service struct {
	repo Repository
	// publicBaseURL monta a URL pública a partir da configuração, para o link
	// gerado estar correto em qualquer ambiente.
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
		Status:      StatusDraft, // nasce em draft, sempre
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, form); err != nil {
		return Form{}, err
	}
	return form, nil
}

// Update altera título e descrição. Ponteiro nil significa "não mexer",
// distinguindo isso de "definir como vazio" — é um PATCH, não um PUT.
func (s *Service) Update(ctx context.Context, formID, ownerID string, title, description *string) (Form, error) {
	form, err := s.repo.GetOwnedBy(ctx, formID, ownerID)
	if err != nil {
		return Form{}, err
	}

	if title != nil {
		t := strings.TrimSpace(*title)
		if t == "" {
			return Form{}, ErrTitleRequired
		}
		form.Title = t
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

// ReplaceFields substitui a lista inteira de campos.
//
// O builder é uma unidade de edição, não uma sequência de operações: o usuário
// reordena, renomeia, remove e só então salva. Com CRUD por campo seria
// preciso sincronizar a cada interação ou manter um diff. Aqui o array
// recebido É a verdade e o índice é a posição — quatro endpoints a menos.
//
// Custo: os IDs dos campos mudam a cada save. Irrelevante porque só é
// permitido em draft, quando ainda não existem respostas apontando para eles.
func (s *Service) ReplaceFields(ctx context.Context, formID, ownerID string, inputs []FieldInput) ([]Field, error) {
	form, err := s.repo.GetOwnedBy(ctx, formID, ownerID)
	if err != nil {
		return nil, err
	}

	// Travar a estrutura depois de publicar elimina, com uma checagem de
	// status, três problemas: campo removido deixa resposta órfã, campo novo
	// obrigatório invalida respostas antigas retroativamente, e tipo alterado
	// deixa "abc" num campo que virou numérico. A solução completa seria
	// versionar a definição; está documentada como limitação consciente.
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
			Position: i, // o índice no array é a ordem de exibição
			Config:   normalizeConfig(in.Type, in.Config),
		})
	}

	if err := s.repo.ReplaceFields(ctx, form.ID, fields); err != nil {
		return nil, err
	}
	return fields, nil
}

// Publish gera o slug na primeira publicação e o PRESERVA nas seguintes.
//
// Regenerar mataria todo link já distribuído: alguém que despublicasse para
// corrigir um rótulo e republicasse descobriria, tarde, que o link no e-mail
// enviado para trezentas pessoas parou de funcionar.
func (s *Service) Publish(ctx context.Context, formID, ownerID string) (Form, error) {
	form, err := s.repo.GetOwnedBy(ctx, formID, ownerID)
	if err != nil {
		return Form{}, err
	}

	// Publicar sem campos geraria um link para uma página vazia.
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

// Unpublish volta para draft mantendo o slug, para que republicar reviva o
// mesmo link.
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

// normalizeConfig descarta a configuração que não faz sentido para o tipo, em
// vez de guardar lixo que o validador de submissão teria de ignorar depois.
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
