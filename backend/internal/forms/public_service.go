package forms

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PublicRepository é a fronteira do fluxo anônimo. Separada de Repository de
// propósito: aqui não existe ownerID, porque o visitante não tem sessão — a
// credencial é o slug.
type PublicRepository interface {
	// GetPublishedBySlug devolve ErrFormNotFound também quando o formulário
	// existe mas está em draft.
	GetPublishedBySlug(ctx context.Context, slug string) (Form, error)
	SaveSubmission(ctx context.Context, s Submission) error
	ListSubmissions(ctx context.Context, formID string, limit, offset int) ([]Submission, int, error)
}

type PublicService struct {
	repo PublicRepository
}

func NewPublicService(repo PublicRepository) *PublicService {
	return &PublicService{repo: repo}
}

func (s *PublicService) GetPublished(ctx context.Context, slug string) (Form, error) {
	return s.repo.GetPublishedBySlug(ctx, slug)
}

// Submit valida a resposta contra a definição vinda do banco e persiste.
//
// O cliente envia apenas pares field_id + value. Obrigatoriedade, tipo, faixa
// e opções são sempre relidos daqui — não há como o visitante mandar
// `required: false` nem inventar um campo.
func (s *PublicService) Submit(ctx context.Context, slug string, answers []Answer) (Submission, error) {
	form, err := s.repo.GetPublishedBySlug(ctx, slug)
	if err != nil {
		return Submission{}, err
	}

	// Persiste o retorno da validação, nunca o payload original: assim é
	// impossível gravar um valor que não passou pelo validador.
	normalized, err := ValidateSubmission(form.Fields, answers)
	if err != nil {
		return Submission{}, err
	}

	submission := Submission{
		ID:          uuid.NewString(),
		FormID:      form.ID,
		SubmittedAt: time.Now().UTC(),
		Answers:     normalized,
	}

	if err := s.repo.SaveSubmission(ctx, submission); err != nil {
		return Submission{}, err
	}
	return submission, nil
}

// ListSubmissions é do admin, mas vive aqui porque compartilha o repositório.
// A posse do formulário é resolvida ANTES por quem chama — senão a proteção do
// recurso pai não valeria nada.
func (s *PublicService) ListSubmissions(ctx context.Context, formID string, limit, offset int) ([]Submission, int, error) {
	return s.repo.ListSubmissions(ctx, formID, limit, offset)
}
