package forms

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	slugAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	// 12 caracteres em base62 dão ~71 bits, longe do enumerável.
	slugLength = 12
)

// NewSlug gera o identificador público de um formulário. crypto/rand e não
// math/rand: o segundo permitiria prever os slugs seguintes.
func NewSlug() (string, error) {
	max := big.NewInt(int64(len(slugAlphabet)))
	buf := make([]byte, slugLength)

	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("gerar slug: %w", err)
		}
		buf[i] = slugAlphabet[n.Int64()]
	}
	return string(buf), nil
}
