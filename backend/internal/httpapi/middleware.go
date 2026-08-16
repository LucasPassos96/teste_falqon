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
		headers := w.Header()
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// O endpoint público é anônimo: sem teto no corpo, uma única requisição aloca
// memória até derrubar o processo.
const maxBodyBytes = 1 << 20

func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

func requestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// O http.ResponseWriter não expõe o status depois de escrito.
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			inicio := time.Now()

			defer func() {
				log.Info("requisição",
					"metodo", r.Method,
					"rota", r.URL.Path,
					"status", wrapped.Status(),
					"duracao", time.Since(inicio).String(),
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}
