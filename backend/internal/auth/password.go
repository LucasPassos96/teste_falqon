package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptMaxBytes: o bcrypt ignora tudo depois do byte 72 SEM AVISAR. Sem
	// validar isso, uma senha de 100 caracteres é secretamente uma senha de
	// 72, e duas senhas diferentes com o mesmo prefixo produzem o mesmo hash.
	// Recuso na entrada em vez de truncar em silêncio.
	bcryptMaxBytes = 72

	// Mínimo de comprimento, sem exigência de símbolo ou maiúscula: regras de
	// complexidade empurram o usuário para "Senha@123", previsível e curta.
	// Comprimento é o que importa (NIST 800-63B).
	passwordMinLength = 8
)

var ErrPasswordPolicy = errors.New("senha fora da política")

// hashFake é um hash bcrypt válido de uma senha aleatória, usado para gastar o
// mesmo tempo quando o usuário não existe. Ver CheckDummyPassword.
var hashFake = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")

func ValidatePasswordPolicy(password string) error {
	if len(password) < passwordMinLength {
		return fmt.Errorf("%w: mínimo de %d caracteres", ErrPasswordPolicy, passwordMinLength)
	}
	// len() conta bytes, que é exatamente a unidade que o bcrypt trunca.
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

// CheckDummyPassword existe para o caminho "usuário não encontrado".
//
// Responder a mesma mensagem para e-mail inexistente e senha errada não basta:
// sem usuário eu retornaria em microssegundos, e com usuário gasto ~60ms no
// bcrypt. Essa diferença é mensurável pela rede — é o mesmo vazamento, só que
// por cronômetro. Aqui eu comparo contra um hash fixo e descarto o resultado,
// só para pagar o mesmo custo.
func CheckDummyPassword(password string) {
	_ = bcrypt.CompareHashAndPassword(hashFake, []byte(password))
}
