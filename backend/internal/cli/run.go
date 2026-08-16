package cli

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/config"
	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi"
	"github.com/LucasPassos96/teste_falqon/backend/internal/storage/sqlite"
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

			db, err := sqlite.Open(cfg.Database.Path)
			if err != nil {
				return err
			}
			defer db.Close()

			// Migration no boot para um clone limpo subir sem passo manual.
			if err := sqlite.Migrate(db); err != nil {
				return err
			}
			version, err := sqlite.Version(db)
			if err != nil {
				return err
			}
			logger.Info("banco pronto", "path", cfg.Database.Path, "migration", version)

			authSvc := auth.NewService(
				sqlite.NewUserRepo(db),
				auth.NewTokenIssuer(cfg.Auth.JWTSecret, cfg.Auth.SessionTTL),
			)

			formSvc := forms.NewService(sqlite.NewFormRepo(db), cfg.PublicBaseURL)

			publicSvc := forms.NewPublicService(sqlite.NewSubmissionRepo(db))

			googleAuth := auth.NewGoogleAuth(
				cfg.Auth.Google.ClientID,
				cfg.Auth.Google.ClientSecret,
				cfg.Auth.Google.RedirectURL,
				sqlite.NewUserRepo(db),
				strings.HasPrefix(cfg.PublicBaseURL, "https://"),
			)
			if googleAuth == nil {
				logger.Warn("login com Google desabilitado: credenciais não configuradas")
			}

			return httpapi.Run(cmd.Context(), cfg, authSvc, formSvc, publicSvc, googleAuth, logger)
		},
	}

	config.RegisterFlags(cmd.Flags())

	return cmd
}
