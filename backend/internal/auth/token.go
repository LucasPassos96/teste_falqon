package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	issuer = "form-builder"
	// signingMethod está fixo e é conferido no parse. Ver Validate.
	signingAlg = "HS256"
)

var ErrInvalidToken = errors.New("token inválido")

type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), ttl: ttl}
}

// Issue emite um token contendo apenas o ID do usuário.
//
// JWT é ASSINADO, não criptografado: qualquer pessoa com o token faz base64 do
// payload e lê o conteúdo. A assinatura garante que ninguém alterou, não que
// ninguém leu. Por isso não vai e-mail, nome nem papel — só o identificador.
func (t *TokenIssuer) Issue(userID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(t.ttl)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	})

	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("assinar token: %w", err)
	}
	return signed, expiresAt, nil
}

// Validate confere a assinatura e devolve o ID do usuário.
func (t *TokenIssuer) Validate(raw string) (string, error) {
	claims := &jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(raw, claims,
		func(token *jwt.Token) (any, error) {
			// Segunda barreira, dentro da keyfunc: confere o TIPO do método
			// antes de entregar o segredo.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritmo inesperado: %v", token.Header["alg"])
			}
			return t.secret, nil
		},
		// A barreira principal. A biblioteca entrega o token já parseado e
		// deixa você devolver a chave; se você devolver o segredo sem olhar o
		// header `alg`, quem escolhe o algoritmo é o atacante — ele manda
		// `alg: none` (sem assinatura) ou troca HMAC por RSA, assinando com a
		// chave pública. É o abuso de JWT mais explorado que existe.
		jwt.WithValidMethods([]string{signingAlg}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if claims.Subject == "" {
		return "", fmt.Errorf("%w: sem subject", ErrInvalidToken)
	}
	return claims.Subject, nil
}
