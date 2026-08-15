// Package httpapi monta e opera o servidor HTTP.
package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/config"
	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// Run sobe o servidor e só retorna quando ctx é cancelado (SIGINT/SIGTERM) ou
// o servidor falha.
func Run(ctx context.Context, cfg *config.Config, authSvc *auth.Service, formSvc *forms.Service, log *slog.Logger) error {
	router, err := NewRouter(cfg, authSvc, formSvc, log)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
		// O zero value do http.Server não tem timeout nenhum: uma conexão que
		// envia o cabeçalho byte a byte segura uma goroutine para sempre.
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Info("servidor iniciado", "address", cfg.Address, "public_base_url", cfg.PublicBaseURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err

	case <-ctx.Done():
		log.Info("encerrando")
		// WithoutCancel: o prazo do desligamento não pode herdar o
		// cancelamento que acabou de acontecer, senão o Shutdown aborta na hora.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("desligamento forçado", "erro", err)
			return err
		}
		return <-serveErr
	}
}
