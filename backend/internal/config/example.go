package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	_ "embed"
)

// Embutido para que `server init-config` funcione de qualquer diretório.
//
//go:embed config.example.yaml
var exampleYAML string

const secretBytes = 48

// NewSecret gera um segredo de sessão com crypto/rand — math/rand é
// determinístico e permitiria forjar sessões.
func NewSecret() (string, error) {
	buf := make([]byte, secretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("gerar segredo: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

// RenderExample devolve o template com um segredo novo no lugar do
// placeholder, pronto para virar o config.yaml de quem acabou de clonar.
func RenderExample() (string, error) {
	secret, err := NewSecret()
	if err != nil {
		return "", err
	}

	rendered := strings.Replace(exampleYAML, jwtSecretPlaceholder, secret, 1)
	if rendered == exampleYAML {
		return "", fmt.Errorf("placeholder %q não encontrado no template", jwtSecretPlaceholder)
	}
	return rendered, nil
}
