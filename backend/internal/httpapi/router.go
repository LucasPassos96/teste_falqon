package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/config"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

func NewRouter(cfg *config.Config, authSvc *auth.Service, log *slog.Logger) (http.Handler, error) {
	spec, err := gen.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("carregar spec embutida: %w", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger(log))
	r.Use(middleware.Recoverer)
	r.Use(securityHeaders)

	api := &API{auth: authSvc, publicBaseURL: cfg.PublicBaseURL}

	handler := gen.NewStrictHandlerWithOptions(
		api,
		[]gen.StrictMiddlewareFunc{
			requireSession(authSvc, publicOperations(spec)),
		},
		gen.StrictHTTPServerOptions{
			// Dois tratadores distintos de propósito: falha ao decodificar o
			// pedido é erro do cliente (422); falha ao produzir a resposta é
			// erro de domínio, classificado caso a caso.
			RequestErrorHandlerFunc:  requestErrorHandler(log),
			ResponseErrorHandlerFunc: errorHandler(log),
		},
	)

	// As rotas de /api/v1 são registradas pelo roteador gerado, não à mão.
	gen.HandlerWithOptions(handler, gen.ChiServerOptions{
		BaseURL:    "/api/v1",
		BaseRouter: r,
	})

	// A spec que gerou este binário, servida pelo próprio binário.
	r.Get("/openapi.json", handleSpec)

	return r, nil
}

func handleSpec(w http.ResponseWriter, _ *http.Request) {
	spec, err := gen.GetSwagger()
	if err != nil {
		http.Error(w, "spec indisponível", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, spec)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
