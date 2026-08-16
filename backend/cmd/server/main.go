// Command server é o binário do backend do Form Builder.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/LucasPassos96/teste_falqon/backend/internal/cli"
)

func main() {
	// Ctrl+C vira cancelamento de contexto, o que permite o desligamento
	// gracioso.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// O cobra já imprimiu a mensagem; aqui só o código de saída.
	if err := cli.Execute(ctx); err != nil {
		os.Exit(1)
	}
}
