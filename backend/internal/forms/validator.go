package forms

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxLabelLength    = 200
	maxHelpTextLength = 500
	maxOptions        = 100
)

// ValidateFieldDefinitions confere a definição que o ADMIN está salvando.
//
// É diferente da validação de submissão (que confere a resposta de um
// visitante contra esta definição). Esta existe para o banco nunca guardar
// config incoerente: sem ela, o admin salva um `select` sem opções, publica, e
// o visitante recebe um campo impossível de preencher — erro que só aparece em
// produção, com o link já distribuído.
func ValidateFieldDefinitions(inputs []FieldInput) error {
	if len(inputs) > MaxFields {
		return fmt.Errorf("%w: máximo de %d campos, recebi %d", ErrTooManyFields, MaxFields, len(inputs))
	}

	for i, in := range inputs {
		if err := validateFieldDefinition(in); err != nil {
			// O índice é a única referência estável: o campo ainda não tem ID.
			return fmt.Errorf("campo %d: %w", i+1, err)
		}
	}
	return nil
}

func validateFieldDefinition(in FieldInput) error {
	if strings.TrimSpace(in.Label) == "" {
		return fmt.Errorf("%w: rótulo é obrigatório", ErrInvalidField)
	}
	// RuneCountInString e não len(): len conta BYTES, e "ação" tem 4
	// caracteres mas 5 bytes. Um limite implementado com len recusaria texto
	// perfeitamente válido em português.
	if utf8.RuneCountInString(in.Label) > maxLabelLength {
		return fmt.Errorf("%w: rótulo acima de %d caracteres", ErrInvalidField, maxLabelLength)
	}
	if utf8.RuneCountInString(in.HelpText) > maxHelpTextLength {
		return fmt.Errorf("%w: texto de ajuda acima de %d caracteres", ErrInvalidField, maxHelpTextLength)
	}

	switch in.Type {
	case FieldShortText, FieldLongText:
		return validateLengthConfig(in.Config)
	case FieldNumber:
		return validateRangeConfig(in.Config)
	case FieldSelect:
		return validateOptions(in.Config)
	case FieldEmail, FieldCheckbox:
		// Não aceitam configuração; o que vier é ignorado na persistência.
		return nil
	default:
		return fmt.Errorf("%w: tipo %q não suportado", ErrInvalidField, in.Type)
	}
}

func validateLengthConfig(c FieldConfig) error {
	if c.MinLength != nil && *c.MinLength < 0 {
		return fmt.Errorf("%w: comprimento mínimo negativo", ErrInvalidField)
	}
	if c.MaxLength != nil && *c.MaxLength < 1 {
		return fmt.Errorf("%w: comprimento máximo precisa ser positivo", ErrInvalidField)
	}
	// Sem esta checagem o campo fica impossível de preencher: nenhum texto
	// satisfaz mínimo 10 e máximo 5 ao mesmo tempo.
	if c.MinLength != nil && c.MaxLength != nil && *c.MinLength > *c.MaxLength {
		return fmt.Errorf("%w: comprimento mínimo (%d) maior que o máximo (%d)",
			ErrInvalidField, *c.MinLength, *c.MaxLength)
	}
	return nil
}

func validateRangeConfig(c FieldConfig) error {
	if c.Min != nil && c.Max != nil && *c.Min > *c.Max {
		return fmt.Errorf("%w: valor mínimo (%g) maior que o máximo (%g)",
			ErrInvalidField, *c.Min, *c.Max)
	}
	return nil
}

func validateOptions(c FieldConfig) error {
	if len(c.Options) == 0 {
		return fmt.Errorf("%w: select precisa de ao menos uma opção", ErrInvalidField)
	}
	if len(c.Options) > maxOptions {
		return fmt.Errorf("%w: máximo de %d opções", ErrInvalidField, maxOptions)
	}

	vistos := make(map[string]bool, len(c.Options))
	for _, opt := range c.Options {
		if strings.TrimSpace(opt.Value) == "" || strings.TrimSpace(opt.Label) == "" {
			return fmt.Errorf("%w: opção com valor ou rótulo vazio", ErrInvalidField)
		}
		// Valor duplicado tornaria a resposta ambígua na hora de exibir.
		if vistos[opt.Value] {
			return fmt.Errorf("%w: opção %q duplicada", ErrInvalidField, opt.Value)
		}
		vistos[opt.Value] = true
	}
	return nil
}
