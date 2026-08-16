package httpapi

import (
	"context"

	"github.com/google/uuid"

	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

func (a *API) GetPublicForm(ctx context.Context, req gen.GetPublicFormRequestObject) (gen.GetPublicFormResponseObject, error) {
	form, err := a.public.GetPublished(ctx, req.Slug)
	if err != nil {
		return nil, err
	}

	// PublicForm é um schema distinto de Form: não carrega id interno, dono,
	// contagem de respostas nem timestamps.
	fields := make([]gen.PublicField, 0, len(form.Fields))
	for _, f := range form.Fields {
		pf, err := toPublicField(f)
		if err != nil {
			return nil, err
		}
		fields = append(fields, pf)
	}

	return gen.GetPublicForm200JSONResponse{
		Title:       form.Title,
		Description: form.Description,
		Fields:      fields,
	}, nil
}

func (a *API) SubmitPublicForm(ctx context.Context, req gen.SubmitPublicFormRequestObject) (gen.SubmitPublicFormResponseObject, error) {
	answers := make([]forms.Answer, 0, len(req.Body.Answers))
	for _, in := range req.Body.Answers {
		answers = append(answers, forms.Answer{
			FieldID: in.FieldId.String(),
			Value:   in.Value,
		})
	}

	if _, err := a.public.Submit(ctx, req.Slug, answers); err != nil {
		return nil, err
	}
	return gen.SubmitPublicForm201Response{}, nil
}

func (a *API) ListSubmissions(ctx context.Context, req gen.ListSubmissionsRequestObject) (gen.ListSubmissionsResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}

	// A posse do formulário é resolvida ANTES de listar as respostas. Sem
	// isto, a proteção do recurso pai não valeria nada: bastaria conhecer o id
	// de um formulário alheio para ler tudo que responderam nele.
	form, err := a.forms.Get(ctx, req.FormId.String(), ownerID)
	if err != nil {
		return nil, err
	}

	limit, offset := 50, 0
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	if req.Params.Offset != nil {
		offset = *req.Params.Offset
	}

	list, total, err := a.public.ListSubmissions(ctx, form.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]gen.Submission, 0, len(list))
	for _, s := range list {
		g, err := toGenSubmission(s)
		if err != nil {
			return nil, err
		}
		items = append(items, g)
	}

	return gen.ListSubmissions200JSONResponse{Items: items, Total: total}, nil
}

func toPublicField(f forms.Field) (gen.PublicField, error) {
	id, err := uuid.Parse(f.ID)
	if err != nil {
		return gen.PublicField{}, err
	}
	config := toGenConfig(f.Config)
	return gen.PublicField{
		Id:       id,
		Type:     gen.FieldType(f.Type),
		Label:    f.Label,
		HelpText: &f.HelpText,
		Required: f.Required,
		Config:   &config,
	}, nil
}

func toGenSubmission(s forms.Submission) (gen.Submission, error) {
	id, err := uuid.Parse(s.ID)
	if err != nil {
		return gen.Submission{}, err
	}

	answers := make([]gen.SubmissionAnswer, 0, len(s.Answers))
	for _, a := range s.Answers {
		fieldID, err := uuid.Parse(a.FieldID)
		if err != nil {
			return gen.Submission{}, err
		}
		answers = append(answers, gen.SubmissionAnswer{
			FieldId:    fieldID,
			FieldLabel: a.FieldLabel,
			Value:      a.Value,
		})
	}

	return gen.Submission{
		Id:          id,
		SubmittedAt: s.SubmittedAt,
		Answers:     answers,
	}, nil
}
