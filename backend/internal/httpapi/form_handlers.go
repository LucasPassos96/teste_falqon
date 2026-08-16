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
	for _, form := range list {
		g, err := a.toGenForm(form)
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
	for _, campo := range saved {
		gerado, err := toGenField(campo)
		if err != nil {
			return nil, err
		}
		out = append(out, gerado)
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

func (a *API) toGenForm(form forms.Form) (gen.Form, error) {
	id, err := uuid.Parse(form.ID)
	if err != nil {
		return gen.Form{}, err
	}

	out := gen.Form{
		Id:              id,
		Title:           form.Title,
		Description:     form.Description,
		Status:          gen.FormStatus(form.Status),
		SubmissionCount: form.SubmissionCount,
		CreatedAt:       form.CreatedAt,
		UpdatedAt:       form.UpdatedAt,
	}

	// Slug e URL só existem depois da primeira publicação; ficam nulos até lá.
	if form.PublicSlug != "" {
		slug := form.PublicSlug
		url := a.forms.PublicURL(form)
		out.PublicSlug = &slug
		out.PublicUrl = &url
	}

	if form.Fields != nil {
		fields := make([]gen.Field, 0, len(form.Fields))
		for _, campo := range form.Fields {
			gerado, err := toGenField(campo)
			if err != nil {
				return gen.Form{}, err
			}
			fields = append(fields, gerado)
		}
		out.Fields = &fields
	}

	return out, nil
}

func toGenField(campo forms.Field) (gen.Field, error) {
	id, err := uuid.Parse(campo.ID)
	if err != nil {
		return gen.Field{}, err
	}

	config := toGenConfig(campo.Config)
	return gen.Field{
		Id:       id,
		Type:     gen.FieldType(campo.Type),
		Label:    campo.Label,
		HelpText: &campo.HelpText,
		Required: campo.Required,
		Position: campo.Position,
		Config:   &config,
	}, nil
}

func toGenConfig(config forms.FieldConfig) gen.FieldConfig {
	out := gen.FieldConfig{
		MinLength: config.MinLength,
		MaxLength: config.MaxLength,
		Min:       config.Min,
		Max:       config.Max,
	}
	if len(config.Options) > 0 {
		opts := make([]gen.FieldOption, 0, len(config.Options))
		for _, opcao := range config.Options {
			opts = append(opts, gen.FieldOption{Value: opcao.Value, Label: opcao.Label})
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
			for _, opcao := range *in.Config.Options {
				opts = append(opts, forms.Option{Value: opcao.Value, Label: opcao.Label})
			}
			out.Config.Options = opts
		}
	}
	return out
}
