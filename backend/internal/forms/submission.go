package forms

import (
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Answer struct {
	FieldID    string
	FieldLabel string
	Value      string
}

type Submission struct {
	ID          string
	FormID      string
	SubmittedAt time.Time
	Answers     []Answer
}

// FieldError aponta o campo e o que há de errado com ele.
type FieldError struct {
	FieldID string
	Message string
}

// ValidationError acumula todos os erros da submissão em vez de parar no
// primeiro, para o visitante corrigir o formulário numa passada só.
type ValidationError struct {
	FieldErrors []FieldError
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("submissão inválida: %d campo(s) com erro", len(e.FieldErrors))
}

// ValidateSubmission confere as respostas contra a definição do formulário e
// devolve os valores normalizados, prontos para persistir. Não depende de HTTP
// nem de banco.
func ValidateSubmission(fields []Field, answers []Answer) ([]Answer, error) {
	given := make(map[string]string, len(answers))
	for _, a := range answers {
		given[a.FieldID] = a.Value
	}

	known := make(map[string]bool, len(fields))
	for _, f := range fields {
		known[f.ID] = true
	}

	var errs []FieldError

	// field_id alheio é rejeitado, não ignorado: ignorar em silêncio esconde
	// bug de integração.
	for id := range given {
		if !known[id] {
			errs = append(errs, FieldError{FieldID: id, Message: "Campo não pertence a este formulário"})
		}
	}

	normalized := make([]Answer, 0, len(fields))

	// Percorre a definição, não o payload: assim um campo obrigatório omitido
	// pelo cliente cai na mesma regra de vazio, em vez de passar despercebido.
	for _, f := range fields {
		raw := strings.TrimSpace(given[f.ID])

		if raw == "" {
			if f.Required {
				errs = append(errs, FieldError{f.ID, "Campo obrigatório"})
				continue
			}
			// Opcional e vazio pula as demais regras, senão um campo com
			// mínimo de caracteres seria impossível de deixar em branco.
			normalized = append(normalized, Answer{FieldID: f.ID, FieldLabel: f.Label, Value: ""})
			continue
		}

		value, err := validateValue(f, raw)
		if err != nil {
			errs = append(errs, FieldError{f.ID, err.Error()})
			continue
		}
		normalized = append(normalized, Answer{FieldID: f.ID, FieldLabel: f.Label, Value: value})
	}

	if len(errs) > 0 {
		return nil, &ValidationError{FieldErrors: errs}
	}
	return normalized, nil
}

// validateValue aplica a regra do tipo e devolve o valor canonizado. O service
// persiste este retorno, nunca o payload original.
func validateValue(f Field, raw string) (string, error) {
	switch f.Type {

	case FieldShortText, FieldLongText:
		// RuneCountInString e não len(): len conta bytes, e len("ação") é 5.
		n := utf8.RuneCountInString(raw)
		if f.Config.MinLength != nil && n < *f.Config.MinLength {
			return "", fmt.Errorf("Mínimo de %d caracteres", *f.Config.MinLength)
		}
		if f.Config.MaxLength != nil && n > *f.Config.MaxLength {
			return "", fmt.Errorf("Máximo de %d caracteres", *f.Config.MaxLength)
		}
		return raw, nil

	case FieldEmail:
		addr, err := mail.ParseAddress(raw)
		if err != nil {
			return "", fmt.Errorf("E-mail inválido")
		}
		return strings.ToLower(addr.Address), nil

	case FieldNumber:
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return "", fmt.Errorf("Informe um número válido")
		}
		if f.Config.Min != nil && n < *f.Config.Min {
			return "", fmt.Errorf("Valor mínimo: %g", *f.Config.Min)
		}
		if f.Config.Max != nil && n > *f.Config.Max {
			return "", fmt.Errorf("Valor máximo: %g", *f.Config.Max)
		}
		// Canoniza "007" para "7" e "1.50" para "1.5".
		return strconv.FormatFloat(n, 'f', -1, 64), nil

	case FieldSelect:
		for _, opt := range f.Config.Options {
			if opt.Value == raw {
				return raw, nil
			}
		}
		return "", fmt.Errorf("Opção inválida")

	case FieldCheckbox:
		switch raw {
		case "true":
			return "true", nil
		case "false":
			// Num checkbox, obrigatório significa "precisa estar marcado".
			if f.Required {
				return "", fmt.Errorf("Campo obrigatório")
			}
			return "false", nil
		}
		return "", fmt.Errorf("Valor inválido")

	default:
		return "", fmt.Errorf("Tipo de campo não suportado")
	}
}
