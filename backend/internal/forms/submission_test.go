package forms

import (
	"errors"
	"testing"
)

func ptrInt(v int) *int           { return &v }
func ptrFloat(v float64) *float64 { return &v }

// campo monta uma definição mínima com um ID fixo, para as asserções.
func campo(t FieldType, required bool, c FieldConfig) Field {
	return Field{ID: "f1", Type: t, Label: "Campo", Required: required, Config: c}
}

// TestValidateValue cobre a regra de cada tipo. Table-driven é o formato
// idiomático em Go e é o que permite adicionar um caso numa linha.
func TestValidateValue(t *testing.T) {
	tests := []struct {
		nome      string
		field     Field
		valor     string
		querErro  bool
		querValor string
	}{
		// texto
		{"texto dentro do limite", campo(FieldShortText, false, FieldConfig{MinLength: ptrInt(3), MaxLength: ptrInt(10)}), "abcd", false, "abcd"},
		{"texto curto demais", campo(FieldShortText, false, FieldConfig{MinLength: ptrInt(3)}), "ab", true, ""},
		{"texto longo demais", campo(FieldShortText, false, FieldConfig{MaxLength: ptrInt(3)}), "abcd", true, ""},
		// Este caso existe de propósito: é o que pega a armadilha do len().
		// "ação" tem 4 caracteres e 5 bytes; com len() seria recusado.
		{"acento não conta dobrado", campo(FieldShortText, false, FieldConfig{MaxLength: ptrInt(4)}), "ação", false, "ação"},
		{"emoji conta como 1", campo(FieldShortText, false, FieldConfig{MaxLength: ptrInt(1)}), "🙂", false, "🙂"},
		{"texto longo sem limite", campo(FieldLongText, false, FieldConfig{}), "qualquer coisa", false, "qualquer coisa"},

		// e-mail
		{"e-mail válido em minúsculas", campo(FieldEmail, false, FieldConfig{}), "Ana@Exemplo.COM", false, "ana@exemplo.com"},
		{"e-mail sem arroba", campo(FieldEmail, false, FieldConfig{}), "ana.exemplo.com", true, ""},
		{"e-mail sem domínio", campo(FieldEmail, false, FieldConfig{}), "ana@", true, ""},

		// número
		{"número canonizado", campo(FieldNumber, false, FieldConfig{}), "007", false, "7"},
		{"decimal canonizado", campo(FieldNumber, false, FieldConfig{}), "1.50", false, "1.5"},
		{"negativo aceito sem faixa", campo(FieldNumber, false, FieldConfig{}), "-3", false, "-3"},
		{"não é número", campo(FieldNumber, false, FieldConfig{}), "abc", true, ""},
		{"abaixo do mínimo", campo(FieldNumber, false, FieldConfig{Min: ptrFloat(18)}), "17", true, ""},
		{"no limite do mínimo", campo(FieldNumber, false, FieldConfig{Min: ptrFloat(18)}), "18", false, "18"},
		{"acima do máximo", campo(FieldNumber, false, FieldConfig{Max: ptrFloat(10)}), "11", true, ""},

		// select
		{"opção existente", campo(FieldSelect, false, FieldConfig{Options: []Option{{Value: "a", Label: "A"}}}), "a", false, "a"},
		{"opção inexistente", campo(FieldSelect, false, FieldConfig{Options: []Option{{Value: "a", Label: "A"}}}), "b", true, ""},
		// O visitante manda o value, não o label: mandar o label deve falhar.
		{"mandou o label em vez do value", campo(FieldSelect, false, FieldConfig{Options: []Option{{Value: "a", Label: "Básico"}}}), "Básico", true, ""},

		// checkbox
		{"checkbox marcado", campo(FieldCheckbox, false, FieldConfig{}), "true", false, "true"},
		{"checkbox opcional desmarcado", campo(FieldCheckbox, false, FieldConfig{}), "false", false, "false"},
		{"checkbox obrigatório desmarcado", campo(FieldCheckbox, true, FieldConfig{}), "false", true, ""},
		{"checkbox com lixo", campo(FieldCheckbox, false, FieldConfig{}), "sim", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			got, err := validateValue(tt.field, tt.valor)

			if tt.querErro {
				if err == nil {
					t.Fatalf("queria erro, obtive valor %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("não queria erro, obtive: %v", err)
			}
			if got != tt.querValor {
				t.Errorf("valor = %q, quero %q", got, tt.querValor)
			}
		})
	}
}

func TestObrigatorioVazio(t *testing.T) {
	f := campo(FieldShortText, true, FieldConfig{})

	_, err := ValidateSubmission([]Field{f}, []Answer{{FieldID: "f1", Value: "   "}})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("queria ValidationError, obtive %v", err)
	}
	if len(ve.FieldErrors) != 1 || ve.FieldErrors[0].Message != "Campo obrigatório" {
		t.Errorf("erros = %+v", ve.FieldErrors)
	}
}

// Campo obrigatório ausente do payload precisa cair na mesma regra de vazio —
// é o caso que se perde se a validação iterar sobre as respostas em vez de
// sobre a definição.
func TestObrigatorioAusenteDoPayload(t *testing.T) {
	f := campo(FieldShortText, true, FieldConfig{})

	_, err := ValidateSubmission([]Field{f}, nil)

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("queria ValidationError, obtive %v", err)
	}
	if len(ve.FieldErrors) != 1 || ve.FieldErrors[0].FieldID != "f1" {
		t.Errorf("erros = %+v", ve.FieldErrors)
	}
}

// Campo opcional vazio não pode ser reprovado pelas regras de faixa.
func TestOpcionalVazioPulaAsRegras(t *testing.T) {
	f := campo(FieldShortText, false, FieldConfig{MinLength: ptrInt(5)})

	got, err := ValidateSubmission([]Field{f}, []Answer{{FieldID: "f1", Value: ""}})
	if err != nil {
		t.Fatalf("não queria erro: %v", err)
	}
	if len(got) != 1 || got[0].Value != "" {
		t.Errorf("normalizado = %+v", got)
	}
}

func TestCampoDeOutroFormularioERejeitado(t *testing.T) {
	f := campo(FieldShortText, false, FieldConfig{})

	_, err := ValidateSubmission([]Field{f}, []Answer{
		{FieldID: "f1", Value: "ok"},
		{FieldID: "intruso", Value: "x"},
	})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("queria ValidationError, obtive %v", err)
	}
	if len(ve.FieldErrors) != 1 || ve.FieldErrors[0].FieldID != "intruso" {
		t.Errorf("erros = %+v", ve.FieldErrors)
	}
}

// Todos os erros de uma vez, para o visitante corrigir numa passada só.
func TestAcumulaTodosOsErros(t *testing.T) {
	fields := []Field{
		{ID: "a", Type: FieldShortText, Label: "A", Required: true},
		{ID: "b", Type: FieldEmail, Label: "B", Required: true},
		{ID: "c", Type: FieldNumber, Label: "C", Required: true},
	}

	_, err := ValidateSubmission(fields, []Answer{
		{FieldID: "b", Value: "nao-e-email"},
		{FieldID: "c", Value: "abc"},
	})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("queria ValidationError, obtive %v", err)
	}
	if len(ve.FieldErrors) != 3 {
		t.Errorf("queria 3 erros (a vazio, b inválido, c inválido), obtive %d: %+v",
			len(ve.FieldErrors), ve.FieldErrors)
	}
}

// O rótulo é copiado para a resposta no momento do envio.
func TestNormalizadoCarregaOLabel(t *testing.T) {
	fields := []Field{{ID: "f1", Type: FieldShortText, Label: "Telefone", Required: true}}

	got, err := ValidateSubmission(fields, []Answer{{FieldID: "f1", Value: " 11999 "}})
	if err != nil {
		t.Fatalf("não queria erro: %v", err)
	}
	if got[0].FieldLabel != "Telefone" {
		t.Errorf("label = %q, quero %q", got[0].FieldLabel, "Telefone")
	}
	// O trim também é responsabilidade de quem valida.
	if got[0].Value != "11999" {
		t.Errorf("valor = %q, quero %q", got[0].Value, "11999")
	}
}
