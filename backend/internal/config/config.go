// Package config carrega a configuração do servidor.
// Precedência: flag > variável de ambiente (FB_*) > arquivo > default.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const envPrefix = "FB"

type Config struct {
	Address string `mapstructure:"address"`
	// PublicBaseURL é a origem do frontend, usada para montar o link público
	// de um formulário publicado.
	PublicBaseURL string   `mapstructure:"public_base_url"`
	Database      Database `mapstructure:"database"`
}

type Database struct {
	Path string `mapstructure:"path"`
}

var defaults = map[string]any{
	"address":         "localhost:8080",
	"public_base_url": "http://localhost:5173",
	"database.path":   "./formbuilder.db",
}

// flagFor mapeia chave de config para nome de flag: snake_case no YAML,
// kebab-case na linha de comando.
var flagFor = map[string]string{
	"address":         "address",
	"public_base_url": "public-base-url",
	"database.path":   "db-path",
}

// RegisterFlags declara as flags de configuração. Os defaults ficam no mapa
// `defaults` para não existirem em dois lugares.
func RegisterFlags(fs *pflag.FlagSet) {
	fs.String("address", "", `endereço de escuta (default "localhost:8080")`)
	fs.String("public-base-url", "", `URL base do frontend (default "http://localhost:5173")`)
	fs.String("db-path", "", `caminho do arquivo SQLite (default "./formbuilder.db")`)
}

func Load(path string, fs *pflag.FlagSet) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	for key, def := range defaults {
		v.SetDefault(key, def)
		// BindEnv explícito em vez de AutomaticEnv, que não alimenta o
		// Unmarshal de forma confiável.
		if err := v.BindEnv(key); err != nil {
			return nil, fmt.Errorf("vincular env %q: %w", key, err)
		}
	}

	if fs != nil {
		for key, name := range flagFor {
			if f := fs.Lookup(name); f != nil {
				if err := v.BindPFlag(key, f); err != nil {
					return nil, fmt.Errorf("vincular flag %q: %w", name, err)
				}
			}
		}
	}

	if err := readFile(v, path); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("interpretar configuração: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// readFile trata a diferença entre pedir um arquivo e procurar um: com
// --config explícito, arquivo ausente é erro; sem a flag, é normal.
func readFile(v *viper.Viper, path string) error {
	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("ler o arquivo de configuração %q: %w", path, err)
		}
		return nil
	}

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	var notFound viper.ConfigFileNotFoundError
	if err := v.ReadInConfig(); err != nil && !errors.As(err, &notFound) {
		return fmt.Errorf("ler o arquivo de configuração: %w", err)
	}
	return nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		return errors.New("address é obrigatório")
	}
	u, err := url.Parse(c.PublicBaseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("public_base_url precisa ser uma URL absoluta (ex.: http://localhost:5173), recebi %q", c.PublicBaseURL)
	}
	if strings.TrimSpace(c.Database.Path) == "" {
		return errors.New("database.path é obrigatório")
	}
	return nil
}
