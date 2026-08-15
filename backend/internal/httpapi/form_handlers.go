package httpapi

import (
	"context"

	"github.com/google/uuid"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

func (a *API) ListForms(ctx context.Context, _ gen.ListFormsRequestObject) (gen.ListFormsResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}

	list, err := a.forms.List(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	out := make([]gen.Form, 0, len(list))
	for _, f := range list {
		g, err := a.toGenForm(f)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return gen.ListForms200JSONResponse(out), nil
}

func (a *API) CreateForm(ctx context.Context, req gen.CreateFormRequestObject) (gen.CreateFormResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}

	var description string
	if req.Body.Description != nil {
		description = *req.Body.Description
	}

	form, err := a.forms.Create(ctx, ownerID, req.Body.Title, description)
	if err != nil {
		return nil, err
	}

	g, err := a.toGenForm(form)
	if err != nil {
		return nil, err
	}
	return gen.CreateForm201JSONResponse(g), nil
}

func (a *API) GetForm(ctx context.Context, req gen.GetFormRequestObject) (gen.GetFormResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}

	form, err := a.forms.Get(ctx, req.FormId.String(), ownerID)
	if err != nil {
		return nil, err
	}

	g, err := a.toGenForm(form)
	if err != nil {
		return nil, err
	}
	return gen.GetForm200JSONResponse(g), nil
}

func (a *API) UpdateForm(ctx context.Context, req gen.UpdateFormRequestObject) (gen.UpdateFormResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}

	// Os campos chegam como ponteiro porque é PATCH: nil significa "não
	// mexer", que é diferente de "definir como vazio".
	form, err := a.forms.Update(ctx, req.FormId.String(), ownerID, req.Body.Title, req.Body.Description)
	if err != nil {
		return nil, err
	}

	g, err := a.toGenForm(form)
	if err != nil {
		return nil, err
	}
	return gen.UpdateForm200JSONResponse(g), nil
}

func (a *API) DeleteForm(ctx context.Context, req gen.DeleteFormRequestObject) (gen.DeleteFormResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.forms.Delete(ctx, req.FormId.String(), ownerID); err != nil {
		return nil, err
	}
	return gen.DeleteForm204Response{}, nil
}

func (a *API) ReplaceFormFields(ctx context.Context, req gen.ReplaceFormFieldsRequestObject) (gen.ReplaceFormFieldsResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return nil, forms.ErrFieldsRequired
	}

	inputs := make([]forms.FieldInput, 0, len(*req.Body))
	for _, in := range *req.Body {
		inputs = append(inputs, toDomainFieldInput(in))
	}

	saved, err := a.forms.ReplaceFields(ctx, req.FormId.String(), ownerID, inputs)
	if err != nil {
		return nil, err
	}

	out := make([]gen.Field, 0, len(saved))
	for _, f := range saved {
		g, err := toGenField(f)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return gen.ReplaceFormFields200JSONResponse(out), nil
}

func (a *API) PublishForm(ctx context.Context, req gen.PublishFormRequestObject) (gen.PublishFormResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}

	form, err := a.forms.Publish(ctx, req.FormId.String(), ownerID)
	if err != nil {
		return nil, err
	}

	g, err := a.toGenForm(form)
	if err != nil {
		return nil, err
	}
	return gen.PublishForm200JSONResponse(g), nil
}

func (a *API) UnpublishForm(ctx context.Context, req gen.UnpublishFormRequestObject) (gen.UnpublishFormResponseObject, error) {
	ownerID, err := a.ownerID(ctx)
	if err != nil {
		return nil, err
	}

	form, err := a.forms.Unpublish(ctx, req.FormId.String(), ownerID)
	if err != nil {
		return nil, err
	}

	g, err := a.toGenForm(form)
	if err != nil {
		return nil, err
	}
	return gen.UnpublishForm200JSONResponse(g), nil
}

// ownerID lê o dono da sessão. É a única origem aceitável: vindo do payload, o
// atacante escolheria de quem é o formulário.
func (a *API) ownerID(ctx context.Context) (string, error) {
	id, ok := auth.UserIDFrom(ctx)
	if !ok {
		return "", auth.ErrInvalidToken
	}
	return id, nil
}

func (a *API) toGenForm(f forms.Form) (gen.Form, error) {
	id, err := uuid.Parse(f.ID)
	if err != nil {
		return gen.Form{}, err
	}

	out := gen.Form{
		Id:              id,
		Title:           f.Title,
		Description:     f.Description,
		Status:          gen.FormStatus(f.Status),
		SubmissionCount: f.SubmissionCount,
		CreatedAt:       f.CreatedAt,
		UpdatedAt:       f.UpdatedAt,
	}

	// Slug e URL só existem depois da primeira publicação; ficam nulos até lá.
	if f.PublicSlug != "" {
		slug := f.PublicSlug
		url := a.forms.PublicURL(f)
		out.PublicSlug = &slug
		out.PublicUrl = &url
	}

	if f.Fields != nil {
		fields := make([]gen.Field, 0, len(f.Fields))
		for _, field := range f.Fields {
			g, err := toGenField(field)
			if err != nil {
				return gen.Form{}, err
			}
			fields = append(fields, g)
		}
		out.Fields = &fields
	}

	return out, nil
}

func toGenField(f forms.Field) (gen.Field, error) {
	id, err := uuid.Parse(f.ID)
	if err != nil {
		return gen.Field{}, err
	}

	config := toGenConfig(f.Config)
	return gen.Field{
		Id:       id,
		Type:     gen.FieldType(f.Type),
		Label:    f.Label,
		HelpText: &f.HelpText,
		Required: f.Required,
		Position: f.Position,
		Config:   &config,
	}, nil
}

func toGenConfig(c forms.FieldConfig) gen.FieldConfig {
	out := gen.FieldConfig{
		MinLength: c.MinLength,
		MaxLength: c.MaxLength,
		Min:       c.Min,
		Max:       c.Max,
	}
	if len(c.Options) > 0 {
		opts := make([]gen.FieldOption, 0, len(c.Options))
		for _, o := range c.Options {
			opts = append(opts, gen.FieldOption{Value: o.Value, Label: o.Label})
		}
		out.Options = &opts
	}
	return out
}

func toDomainFieldInput(in gen.FieldInput) forms.FieldInput {
	out := forms.FieldInput{
		Type:     forms.FieldType(in.Type),
		Label:    in.Label,
		Required: in.Required,
	}
	if in.HelpText != nil {
		out.HelpText = *in.HelpText
	}
	if in.Config != nil {
		out.Config = forms.FieldConfig{
			MinLength: in.Config.MinLength,
			MaxLength: in.Config.MaxLength,
			Min:       in.Config.Min,
			Max:       in.Config.Max,
		}
		if in.Config.Options != nil {
			opts := make([]forms.Option, 0, len(*in.Config.Options))
			for _, o := range *in.Config.Options {
				opts = append(opts, forms.Option{Value: o.Value, Label: o.Label})
			}
			out.Config.Options = opts
		}
	}
	return out
}
