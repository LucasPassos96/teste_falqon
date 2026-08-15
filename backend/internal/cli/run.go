package cli

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/LucasPassos96/teste_falqon/backend/internal/config"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Sobe o servidor HTTP",
		Example: "  server run\n" +
			"  server run --address=localhost:9000\n" +
			"  server run --config=./config.yaml\n" +
			"  FB_ADDRESS=localhost:9000 server run",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(configPath, cmd.Flags())
			if err != nil {
				return err
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))

			return httpapi.Run(cmd.Context(), cfg, logger)
		},
	}

	config.RegisterFlags(cmd.Flags())

	return cmd
}
