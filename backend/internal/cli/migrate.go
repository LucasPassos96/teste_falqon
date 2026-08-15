package cli

import (
	"github.com/spf13/cobra"

	"github.com/LucasPassos96/teste_falqon/backend/internal/config"
	"github.com/LucasPassos96/teste_falqon/backend/internal/storage/sqlite"
)

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Aplica as migrations pendentes",
		Long: "Aplica as migrations pendentes ao banco configurado.\n\n" +
			"O comando é idempotente: rodar duas vezes seguidas não faz nada na segunda.\n" +
			"O servidor também aplica as migrations no boot, então este comando existe\n" +
			"para preparar o banco sem subir a API.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(configPath, cmd.Flags())
			if err != nil {
				return err
			}

			db, err := sqlite.Open(cfg.Database.Path)
			if err != nil {
				return err
			}
			defer db.Close()

			if err := sqlite.Migrate(db); err != nil {
				return err
			}

			version, err := sqlite.Version(db)
			if err != nil {
				return err
			}

			cmd.Printf("Banco %s na versão %d\n", cfg.Database.Path, version)
			return nil
		},
	}

	config.RegisterFlags(cmd.Flags())

	return cmd
}
