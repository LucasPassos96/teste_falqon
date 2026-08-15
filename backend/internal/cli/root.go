// Package cli monta a árvore de comandos do binário `server`.
package cli

import (
	"context"

	"github.com/spf13/cobra"
)

var configPath string

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "server",
		Short: "Form Builder — API em Go",
		Long: "Backend do Form Builder.\n\n" +
			"Configuração por arquivo, variável de ambiente ou flag,\n" +
			"nesta ordem de precedência: flag > env (FB_*) > arquivo > default.",
		// Sem isto, erro em tempo de execução despeja o help inteiro junto.
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&configPath, "config", "",
		"caminho do arquivo de configuração (por padrão procura ./config.yaml)")

	root.AddCommand(newRunCmd())

	return root
}

func Execute(ctx context.Context) error {
	return NewRootCmd().ExecuteContext(ctx)
}
