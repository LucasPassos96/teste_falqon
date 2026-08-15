package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/LucasPassos96/teste_falqon/backend/internal/config"
)

func newInitConfigCmd() *cobra.Command {
	var (
		out   string
		force bool
	)

	cmd := &cobra.Command{
		Use:   "init-config",
		Short: "Cria um config.yaml com um segredo de sessão novo",
		Long: "Cria um arquivo de configuração a partir do template embutido, já com um\n" +
			"segredo de sessão gerado por crypto/rand.\n\n" +
			"É o que `task setup` chama, para que um clone limpo rode sem edição manual\n" +
			"de arquivo e sem depender de openssl.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Não sobrescrever é o comportamento padrão de propósito: `task
			// setup` roda mais de uma vez ao longo da vida do projeto, e
			// trocar o segredo invalidaria todas as sessões ativas.
			if _, err := os.Stat(out); err == nil && !force {
				cmd.Printf("%s já existe, mantido como está (use --force para sobrescrever)\n", out)
				return nil
			} else if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("verificar %s: %w", out, err)
			}

			content, err := config.RenderExample()
			if err != nil {
				return err
			}

			// 0600: o arquivo contém o segredo de assinatura das sessões.
			if err := os.WriteFile(out, []byte(content), 0o600); err != nil {
				return fmt.Errorf("escrever %s: %w", out, err)
			}

			cmd.Printf("%s criado com um segredo de sessão novo\n", out)
			return nil
		},
	}

	cmd.Flags().StringVar(&out, "out", "config.yaml", "caminho do arquivo a criar")
	cmd.Flags().BoolVar(&force, "force", false, "sobrescreve um config existente, invalidando as sessões ativas")

	return cmd
}
