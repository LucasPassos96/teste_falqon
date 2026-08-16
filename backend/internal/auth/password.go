package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// O bcrypt ignora tudo depois do byte 72 sem avisar, então recusamos em
	// vez de truncar em silêncio.
	bcryptMaxBytes    = 72
	passwordMinLength = 8
)

var ErrPasswordPolicy = errors.New("senha fora da política")

var hashFake = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")

func ValidatePasswordPolicy(password string) error {
	if len(password) < passwordMinLength {
		return fmt.Errorf("%w: mínimo de %d caracteres", ErrPasswordPolicy, passwordMinLength)
	}
	if len(password) > bcryptMaxBytes {
		return fmt.Errorf("%w: máximo de %d bytes", ErrPasswordPolicy, bcryptMaxBytes)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePasswordPolicy(password); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("gerar hash: %w", err)
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// CheckDummyPassword gasta o mesmo tempo de um bcrypt real no caminho
// "usuário não encontrado", para que a diferença de tempo de resposta não
// revele quais e-mails têm conta.
func CheckDummyPassword(password string) {
	_ = bcrypt.CompareHashAndPassword(hashFake, []byte(password))
}
