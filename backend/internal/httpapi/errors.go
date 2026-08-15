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

// fieldError é o item de field_errors declarado na spec. O gerador produz um
// struct anônimo dentro de ValidationError, então este alias existe só para
// montá-lo sem repetir a declaração inteira.
type fieldError = struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// requestErrorHandler trata o payload que nem chegou a virar uma requisição
// válida: JSON malformado, campo obrigatório ausente, formato recusado pelo
// tipo gerado (um `format: email` da spec vira validação no unmarshal).
//
// Isso é sempre erro do cliente. Cair no tratador genérico devolveria 500 —
// mentindo sobre de quem é a culpa e transformando entrada inválida em alarme
// de indisponibilidade no monitoramento.
//
// A mensagem é genérica de propósito: o erro cru do decodificador descreve a
// estrutura interna do payload.
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

// errorHandler traduz erro de domínio em resposta HTTP.
//
// É o único lugar que decide status e mensagem, e é o que garante que erro de
// banco nunca vaza para o cliente: mensagem crua entrega nome de tabela,
// coluna e às vezes o SQL inteiro. O detalhe vai para o log com o request ID;
// o cliente recebe uma frase genérica.
func errorHandler(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		status, code, message, field := classify(err)

		if status >= http.StatusInternalServerError {
			log.Error("erro interno",
				"erro", err,
				"rota", r.URL.Path,
				"request_id", middleware.GetReqID(r.Context()),
			)
		}

		// O 422 usa o envelope ValidationError, com field_errors; os demais
		// usam Error. São dois schemas distintos na spec e precisam bater.
		if status == http.StatusUnprocessableEntity {
			writeValidationError(w, message, &fieldError{Field: field, Message: message})
			return
		}

		writeJSON(w, status, gen.Error{Code: code, Message: message})
	}
}

// writeValidationError sempre emite field_errors, que é obrigatório no schema.
// Um array vazio satisfaz o contrato quando não há um campo identificável.
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
		// Mesma resposta para e-mail inexistente e senha errada: mensagens
		// diferentes transformariam o login numa API de consulta de contas.
		return http.StatusUnauthorized, "invalid_credentials", "E-mail ou senha inválidos", ""

	case errors.Is(err, auth.ErrEmailTaken):
		return http.StatusConflict, "email_taken", "Já existe uma conta com este e-mail", ""

	case errors.Is(err, auth.ErrUserNotFound):
		// A sessão aponta para um usuário que não existe mais: do ponto de
		// vista do cliente é uma sessão inválida, não um recurso ausente.
		return http.StatusUnauthorized, "unauthorized", "Sessão ausente ou inválida", ""

	// As mensagens abaixo são seguras de mostrar: dizem o que corrigir, sem
	// revelar nada sobre o sistema.
	case errors.Is(err, auth.ErrPasswordPolicy):
		return http.StatusUnprocessableEntity, "validation_failed", err.Error(), "password"

	case errors.Is(err, auth.ErrInvalidEmail):
		return http.StatusUnprocessableEntity, "validation_failed", "E-mail inválido", "email"

	case errors.Is(err, auth.ErrNameRequired):
		return http.StatusUnprocessableEntity, "validation_failed", "Nome é obrigatório", "name"

	// 404 e não 403 para formulário de outro usuário: 403 significa "existe,
	// mas não é seu", o que já é informação. O repositório devolve o mesmo
	// erro para inexistente e para alheio, então nem o handler consegue
	// distinguir os dois.
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
