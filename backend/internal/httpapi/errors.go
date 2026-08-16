package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

// fieldError é o item de field_errors da spec; o alias evita repetir o struct
// anônimo que o gerador produz.
type fieldError = struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// requestErrorHandler trata o payload que nem chegou a virar uma requisição
// válida: JSON malformado, campo ausente, formato recusado no unmarshal. É
// sempre erro do cliente, então 422 e não 500. A mensagem é genérica porque o
// erro cru do decodificador descreve a estrutura interna do payload.
func requestErrorHandler(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Info("payload rejeitado",
			"rota", r.URL.Path,
			"erro", err,
			"request_id", middleware.GetReqID(r.Context()),
		)
		writeValidationError(w, "Payload inválido", nil)
	}
}

// errorHandler traduz erro de domínio em resposta HTTP. É o único lugar que
// decide status e mensagem, e o que impede erro de banco de vazar: o detalhe
// vai para o log, o cliente recebe uma frase genérica.
func errorHandler(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		var ve *forms.ValidationError
		if errors.As(err, &ve) {
			writeSubmissionErrors(w, ve)
			return
		}

		status, code, message, field := classify(err)

		if status >= http.StatusInternalServerError {
			log.Error("erro interno",
				"erro", err,
				"rota", r.URL.Path,
				"request_id", middleware.GetReqID(r.Context()),
			)
		}

		// 422 usa o envelope ValidationError; os demais usam Error. São dois
		// schemas distintos na spec.
		if status == http.StatusUnprocessableEntity {
			writeValidationError(w, message, &fieldError{Field: field, Message: message})
			return
		}

		writeJSON(w, status, gen.Error{Code: code, Message: message})
	}
}

// writeSubmissionErrors devolve todos os campos reprovados de uma vez.
func writeSubmissionErrors(w http.ResponseWriter, ve *forms.ValidationError) {
	errs := make([]fieldError, 0, len(ve.FieldErrors))
	for _, fe := range ve.FieldErrors {
		errs = append(errs, fieldError{Field: fe.FieldID, Message: fe.Message})
	}

	writeJSON(w, http.StatusUnprocessableEntity, gen.ValidationError{
		Code:        "validation_failed",
		Message:     "A submissão contém campos inválidos",
		FieldErrors: errs,
	})
}

// writeValidationError sempre emite field_errors, obrigatório no schema.
func writeValidationError(w http.ResponseWriter, message string, fe *fieldError) {
	errs := []fieldError{}
	if fe != nil {
		errs = append(errs, *fe)
	}
	writeJSON(w, http.StatusUnprocessableEntity, gen.ValidationError{
		Code:        "validation_failed",
		Message:     message,
		FieldErrors: errs,
	})
}

func classify(err error) (status int, code, message, field string) {
	switch {
	case errors.Is(err, auth.ErrInvalidToken):
		return http.StatusUnauthorized, "unauthorized", "Sessão ausente ou inválida", ""

	case errors.Is(err, auth.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid_credentials", "E-mail ou senha inválidos", ""

	case errors.Is(err, auth.ErrEmailTaken):
		return http.StatusConflict, "email_taken", "Já existe uma conta com este e-mail", ""

	// Sessão apontando para usuário que não existe mais é sessão inválida,
	// não recurso ausente.
	case errors.Is(err, auth.ErrUserNotFound):
		return http.StatusUnauthorized, "unauthorized", "Sessão ausente ou inválida", ""

	case errors.Is(err, auth.ErrPasswordPolicy):
		return http.StatusUnprocessableEntity, "validation_failed", err.Error(), "password"

	case errors.Is(err, auth.ErrInvalidEmail):
		return http.StatusUnprocessableEntity, "validation_failed", "E-mail inválido", "email"

	case errors.Is(err, auth.ErrNameRequired):
		return http.StatusUnprocessableEntity, "validation_failed", "Nome é obrigatório", "name"

	// 404 e não 403 para formulário alheio: 403 confirmaria que ele existe.
	case errors.Is(err, forms.ErrFormNotFound):
		return http.StatusNotFound, "form_not_found", "Formulário não encontrado", ""

	case errors.Is(err, forms.ErrFormPublished):
		return http.StatusConflict, "form_published",
			"Despublique o formulário para alterar os campos", ""

	case errors.Is(err, forms.ErrNoFields):
		return http.StatusConflict, "form_without_fields",
			"Adicione ao menos um campo antes de publicar", ""

	case errors.Is(err, forms.ErrTitleRequired):
		return http.StatusUnprocessableEntity, "validation_failed", "Título é obrigatório", "title"

	case errors.Is(err, forms.ErrInvalidField),
		errors.Is(err, forms.ErrTooManyFields),
		errors.Is(err, forms.ErrFieldsRequired):
		return http.StatusUnprocessableEntity, "validation_failed", err.Error(), "fields"

	default:
		return http.StatusInternalServerError, "internal_error", "Erro interno", ""
	}
}
