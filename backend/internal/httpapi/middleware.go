package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// securityHeaders: nosniff impede o navegador de tratar uma resposta JSON como
// HTML; DENY bloqueia clickjacking na página pública do formulário;
// no-referrer evita vazar o slug para sites externos.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func requestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// O http.ResponseWriter não expõe o status depois de escrito.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				log.Info("requisição",
					"metodo", r.Method,
					"rota", r.URL.Path,
					"status", ww.Status(),
					"duracao", time.Since(start).String(),
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
