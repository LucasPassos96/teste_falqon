// Package forms concentra o modelo e as regras de negócio dos formulários.
package forms

import (
	"errors"
	"time"
)

type FieldType string

const (
	FieldShortText FieldType = "short_text"
	FieldLongText  FieldType = "long_text"
	FieldEmail     FieldType = "email"
	FieldNumber    FieldType = "number"
	FieldSelect    FieldType = "select"
	FieldCheckbox  FieldType = "checkbox"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
)

var (
	ErrFormNotFound   = errors.New("formulário não encontrado")
	ErrFormPublished  = errors.New("formulário publicado não pode ter a estrutura alterada")
	ErrNoFields       = errors.New("formulário sem campos não pode ser publicado")
	ErrTitleRequired  = errors.New("título é obrigatório")
	ErrInvalidField   = errors.New("definição de campo inválida")
	ErrTooManyFields  = errors.New("campos demais")
	ErrFieldsRequired = errors.New("a lista de campos é obrigatória")
)

// MaxFields evita que um único PUT insira dezenas de milhares de linhas.
const MaxFields = 50

type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Ponteiros para distinguir "não configurado" de "configurado como zero" — em
// Go um int não inicializado vale 0.
type FieldConfig struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Options   []Option `json:"options,omitempty"`
}

type Field struct {
	ID       string
	FormID   string
	Type     FieldType
	Label    string
	HelpText string
	Required bool
	Position int
	Config   FieldConfig
}

// FieldInput é o que o builder envia: a posição é o índice no array.
type FieldInput struct {
	Type     FieldType
	Label    string
	HelpText string
	Required bool
	Config   FieldConfig
}

type Form struct {
	ID          string
	UserID      string
	Title       string
	Description string
	Status      Status
	// Vazio até a primeira publicação; nunca muda depois.
	PublicSlug      string
	PublishedAt     *time.Time
	SubmissionCount int
	Fields          []Field
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (f Form) IsPublished() bool { return f.Status == StatusPublished }
