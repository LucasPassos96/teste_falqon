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

// UserRepository é declarada aqui, no pacote que CONSOME, e não no que
// implementa. A dependência aponta para dentro: `auth` não sabe que existe
// SQLite, e trocar de banco não toca nesta camada.
type UserRepository interface {
	Create(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
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

// Login devolve o mesmo erro para e-mail inexistente e senha errada.
//
// Se o app diz "usuário não encontrado" num caso e "senha incorreta" no outro,
// o formulário de login vira uma API de consulta: dá para descobrir quem tem
// conta. Uma mensagem só, um status só — e o mesmo tempo de resposta.
func (s *Service) Login(ctx context.Context, email, password string) (User, error) {
	normalized, err := normalizeEmail(email)
	if err != nil {
		// E-mail malformado também não pode revelar nada: mesma resposta.
		CheckDummyPassword(password)
		return User{}, ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Paga o custo do bcrypt mesmo sem usuário, senão a diferença de
			// tempo entrega quem tem conta.
			CheckDummyPassword(password)
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}

	// Usuário criado só pelo Google não tem hash: não pode entrar por senha.
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

// normalizeEmail valida pela RFC 5322 e devolve em minúsculas, para que o
// UNIQUE do banco não deixe passar "A@b.com" e "a@b.com" como contas
// diferentes.
func normalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	// ParseAddress em vez de regex: regex de e-mail é o exemplo canônico de
	// código que parece certo e recusa endereço legítimo.
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidEmail, email)
	}
	return strings.ToLower(addr.Address), nil
}
