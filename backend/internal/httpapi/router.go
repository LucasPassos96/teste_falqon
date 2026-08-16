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
	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

func NewRouter(cfg *config.Config, authSvc *auth.Service, formSvc *forms.Service, publicSvc *forms.PublicService, googleAuth *auth.GoogleAuth, log *slog.Logger) (http.Handler, error) {
	spec, err := gen.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("carregar spec embutida: %w", err)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(requestLogger(log))
	router.Use(middleware.Recoverer)
	router.Use(securityHeaders)
	router.Use(limitBody)

	api := &API{auth: authSvc, forms: formSvc, public: publicSvc, google: googleAuth, publicBaseURL: cfg.PublicBaseURL}

	handler := gen.NewStrictHandlerWithOptions(
		api,
		[]gen.StrictMiddlewareFunc{
			requireSession(authSvc, publicOperations(spec)),
		},
		gen.StrictHTTPServerOptions{
			// Falha ao decodificar o pedido é erro do cliente; falha ao
			// produzir a resposta é erro de domínio.
			RequestErrorHandlerFunc:  requestErrorHandler(log),
			ResponseErrorHandlerFunc: errorHandler(log),
		},
	)

	gen.HandlerWithOptions(handler, gen.ChiServerOptions{
		BaseURL:    "/api/v1",
		BaseRouter: router,
	})

	router.Get("/openapi.json", handleSpec)

	return router, nil
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
