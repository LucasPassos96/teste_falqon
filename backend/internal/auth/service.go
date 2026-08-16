// Package auth concentra senha, sessão e as regras de autenticação.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailTaken         = errors.New("e-mail já cadastrado")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrUserNotFound       = errors.New("usuário não encontrado")
	ErrInvalidEmail       = errors.New("e-mail inválido")
	ErrNameRequired       = errors.New("nome é obrigatório")
)

type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string // vazio para usuário que só entra pelo Google
	GoogleID     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository é declarada aqui, no pacote que consome, e não no que
// implementa: `auth` não sabe que existe SQLite.
type UserRepository interface {
	Create(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
	LinkGoogleID(ctx context.Context, userID, googleID string) error
}

type Service struct {
	users  UserRepository
	tokens *TokenIssuer
}

func NewService(users UserRepository, tokens *TokenIssuer) *Service {
	return &Service{users: users, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, name, email, password string) (User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return User{}, ErrNameRequired
	}

	email, err := normalizeEmail(email)
	if err != nil {
		return User{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	now := time.Now().UTC()
	user := User{
		ID:           uuid.NewString(),
		Email:        email,
		Name:         name,
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return User{}, err
	}
	return user, nil
}

// Login devolve o mesmo erro, e gasta o mesmo tempo, para e-mail inexistente
// e senha errada — senão o login vira uma API de consulta de contas.
func (s *Service) Login(ctx context.Context, email, password string) (User, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		CheckDummyPassword(password)
		return User{}, ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			CheckDummyPassword(password)
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}

	// Usuário só-Google não tem hash e não entra por senha.
	if user.PasswordHash == "" {
		CheckDummyPassword(password)
		return User{}, ErrInvalidCredentials
	}

	if !CheckPassword(user.PasswordHash, password) {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) CurrentUser(ctx context.Context, userID string) (User, error) {
	return s.users.FindByID(ctx, userID)
}

// IssueSession emite o token da sessão e diz quando ele expira.
func (s *Service) IssueSession(userID string) (string, time.Time, error) {
	return s.tokens.Issue(userID)
}

func (s *Service) ValidateSession(token string) (string, error) {
	return s.tokens.Validate(token)
}

// normalizeEmail valida e devolve em minúsculas, para o UNIQUE do banco não
// deixar passar "A@b.com" e "a@b.com" como contas diferentes.
func normalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidEmail, email)
	}
	return strings.ToLower(addr.Address), nil
}
