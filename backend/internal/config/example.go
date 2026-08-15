package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	_ "embed"
)

// exampleYAML é o template de configuração, embutido no binário para que
// `server init-config` funcione rodado de qualquer diretório. É o único
// arquivo de exemplo do projeto: não há uma segunda cópia para divergir.
//
//go:embed config.example.yaml
var exampleYAML string

// secretBytes: 48 bytes viram 64 caracteres em base64, bem acima do mínimo de
// 32 exigido no boot.
const secretBytes = 48

// NewSecret gera um segredo de sessão.
//
// crypto/rand e não math/rand: o segundo é determinístico e previsível a
// partir de algumas saídas observadas, o que aqui significaria conseguir
// forjar a sessão de qualquer usuário.
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
		// O placeholder some se alguém editar o template sem perceber. Sem
		// esta checagem, o config sairia com o valor de exemplo e o servidor
		// se recusaria a subir, com a causa longe do sintoma.
		return "", fmt.Errorf("placeholder %q não encontrado no template", jwtSecretPlaceholder)
	}
	return rendered, nil
}
