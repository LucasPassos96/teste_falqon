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

// ValidationError acumula TODOS os erros da submissão em vez de parar no
// primeiro: o visitante corrige o formulário inteiro numa passada, em vez de
// descobrir um problema por vez. É uma decisão de UX que virou decisão de tipo
// — por isso é um slice, não um erro único.
type ValidationError struct {
	FieldErrors []FieldError
}

// Implementar Error() faz *ValidationError satisfazer a interface `error`. Em
// Go isso é implícito: não se declara "implements".
func (e *ValidationError) Error() string {
	return fmt.Sprintf("submissão inválida: %d campo(s) com erro", len(e.FieldErrors))
}

// ValidateSubmission confere as respostas contra a definição do formulário e
// devolve os valores normalizados, prontos para persistir.
//
// Não depende de HTTP nem de banco: recebe dados, devolve dados. É o que a
// torna trivial de testar, e é onde mora a regra que o desafio exige que rode
// no servidor.
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

	// field_id que não pertence ao formulário é REJEITADO, não ignorado.
	// Ignorar em silêncio esconde bug de integração — e esconderia também uma
	// tentativa de sondar quais ids existem.
	for id := range given {
		if !known[id] {
			errs = append(errs, FieldError{FieldID: id, Message: "Campo não pertence a este formulário"})
		}
	}

	normalized := make([]Answer, 0, len(fields))

	// Percorre a DEFINIÇÃO, não o payload. Iterando sobre as respostas, um
	// campo obrigatório simplesmente omitido pelo cliente passaria
	// despercebido; percorrendo a definição, ausência e vazio caem na mesma
	// regra.
	for _, f := range fields {
		// Ler chave inexistente de um map em Go devolve o zero value ("") sem
		// erro, então campo ausente já chega aqui como vazio.
		raw := strings.TrimSpace(given[f.ID])

		if raw == "" {
			if f.Required {
				errs = append(errs, FieldError{f.ID, "Campo obrigatório"})
				continue
			}
			// Opcional e vazio: não aplica as demais regras. Sem isto, um campo
			// de texto opcional com mínimo de 3 caracteres seria impossível de
			// deixar em branco.
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

// validateValue aplica a regra do tipo e devolve o valor canonizado.
//
// Quem valida também canoniza: o service persiste o retorno da validação,
// nunca o payload original, então é impossível gravar algo que não passou por
// aqui.
func validateValue(f Field, raw string) (string, error) {
	switch f.Type {

	case FieldShortText, FieldLongText:
		// RuneCountInString e não len(): len conta BYTES. len("ação") devolve
		// 5, não 4. Um limite de 10 caracteres implementado com len recusaria
		// um nome perfeitamente válido — bug que passa despercebido em teste
		// feito só com ASCII.
		n := utf8.RuneCountInString(raw)
		if f.Config.MinLength != nil && n < *f.Config.MinLength {
			return "", fmt.Errorf("Mínimo de %d caracteres", *f.Config.MinLength)
		}
		if f.Config.MaxLength != nil && n > *f.Config.MaxLength {
			return "", fmt.Errorf("Máximo de %d caracteres", *f.Config.MaxLength)
		}
		return raw, nil

	case FieldEmail:
		// ParseAddress valida conforme a RFC 5322 e vem na biblioteca padrão.
		// Regex de e-mail é o exemplo canônico de código que parece certo e
		// recusa endereço legítimo.
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
		// Canoniza: "007" vira "7", "1.50" vira "1.5".
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
			// "required" num checkbox significa "precisa estar marcado" — é a
			// semântica de "aceito os termos", não de "responda alguma coisa".
			if f.Required {
				return "", fmt.Errorf("Campo obrigatório")
			}
			return "false", nil
		}
		return "", fmt.Errorf("Valor inválido")

	default:
		// Tipo desconhecido é bug nosso, não erro do visitante. O CHECK do
		// banco já impediria chegar aqui.
		return "", fmt.Errorf("Tipo de campo não suportado")
	}
}
