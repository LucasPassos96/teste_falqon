package forms

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PublicRepository é a fronteira do fluxo anônimo: aqui não existe ownerID,
// porque a credencial do visitante é o próprio slug.
type PublicRepository interface {
	// Devolve ErrFormNotFound também quando o formulário existe mas é draft.
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

// Submit valida a resposta contra a definição vinda do banco e persiste. O
// cliente envia apenas pares field_id + value; as regras são sempre relidas
// daqui.
func (s *PublicService) Submit(ctx context.Context, slug string, answers []Answer) (Submission, error) {
	form, err := s.repo.GetPublishedBySlug(ctx, slug)
	if err != nil {
		return Submission{}, err
	}

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
// A posse do formulário é resolvida antes, por quem chama.
func (s *PublicService) ListSubmissions(ctx context.Context, formID string, limit, offset int) ([]Submission, int, error) {
	return s.repo.ListSubmissions(ctx, formID, limit, offset)
}
