package auth

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const SessionCookieName = "session"

// contextKey é um tipo próprio para evitar colisão com chaves de outros pacotes.
type contextKey struct{}

var userIDKey contextKey

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFrom devolve o dono da sessão. É a única origem aceitável do
// identificador do usuário — nunca o corpo da requisição.
func UserIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}

// SessionCookie monta o cookie da sessão: HttpOnly contra XSS, SameSite=Lax
// contra CSRF. Secure é condicional porque, forçado em http, o navegador
// descartaria o cookie em silêncio e o login pareceria quebrado.
func SessionCookie(token string, expiresAt time.Time, publicBaseURL string) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(publicBaseURL, "https://"),
	}
}

// ClearedSessionCookie expira o cookie no navegador. O token em si continua
// válido até o `exp` — o JWT não tem revogação no servidor.
func ClearedSessionCookie(publicBaseURL string) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(publicBaseURL, "https://"),
	}
}
