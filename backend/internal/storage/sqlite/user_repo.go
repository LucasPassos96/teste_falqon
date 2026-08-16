package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

var _ auth.UserRepository = (*UserRepo)(nil)

func (r *UserRepo) Create(ctx context.Context, u auth.User) error {
	const query = `
		INSERT INTO users (id, email, name, password_hash, google_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.Email, u.Name,
		nullIfEmpty(u.PasswordHash),
		nullIfEmpty(u.GoogleID),
		formatTime(u.CreatedAt), formatTime(u.UpdatedAt),
	)
	if err != nil {
		// O UNIQUE do banco é a autoridade: um SELECT antes do INSERT teria
		// janela de corrida.
		if isUniqueViolation(err) {
			return auth.ErrEmailTaken
		}
		return fmt.Errorf("criar usuário: %w", err)
	}
	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (auth.User, error) {
	const query = `
		SELECT id, email, name, password_hash, google_id, created_at, updated_at
		FROM users WHERE email = ?`
	return r.scanOne(r.db.QueryRowContext(ctx, query, email))
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (auth.User, error) {
	const query = `
		SELECT id, email, name, password_hash, google_id, created_at, updated_at
		FROM users WHERE id = ?`
	return r.scanOne(r.db.QueryRowContext(ctx, query, id))
}

func (r *UserRepo) LinkGoogleID(ctx context.Context, userID, googleID string) error {
	const query = `UPDATE users SET google_id = ?, updated_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, googleID, formatTime(time.Now().UTC()), userID)
	if err != nil {
		if isUniqueViolation(err) {
			return auth.ErrEmailTaken
		}
		return fmt.Errorf("vincular conta Google: %w", err)
	}
	return nil
}

func (r *UserRepo) scanOne(row *sql.Row) (auth.User, error) {
	var (
		u                  auth.User
		hash, googleID     sql.NullString
		createdAt, updated string
	)

	err := row.Scan(&u.ID, &u.Email, &u.Name, &hash, &googleID, &createdAt, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrUserNotFound
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("ler usuário: %w", err)
	}

	u.PasswordHash = hash.String
	u.GoogleID = googleID.String
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updated)

	return u, nil
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// nullIfEmpty grava NULL em vez de string vazia: no SQLite vários NULL
// convivem num índice único, mas vários "" colidiriam.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
