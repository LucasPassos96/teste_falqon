package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

func NewRouter(log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger(log))
	r.Use(middleware.Recoverer)
	r.Use(securityHeaders)

	// As rotas de /api/v1 são registradas pelo roteador gerado, não à mão.
	gen.HandlerWithOptions(
		gen.NewStrictHandler(&API{}, nil),
		gen.ChiServerOptions{BaseURL: "/api/v1", BaseRouter: r},
	)

	// A spec que gerou este binário, servida pelo próprio binário.
	r.Get("/openapi.json", handleSpec)

	return r
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
