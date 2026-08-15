package auth

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const SessionCookieName = "session"

// contextKey é um tipo próprio para a chave do context. String crua colidiria
// silenciosamente com a chave de outro pacote.
type contextKey struct{}

var userIDKey contextKey

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFrom devolve o dono da sessão. É a ÚNICA origem aceitável do
// identificador do usuário: se ele viesse do corpo da requisição, o atacante
// escolheria de quem é o recurso.
func UserIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}

// SessionCookie monta o cookie da sessão.
//
// Devolve o cookie em vez de escrevê-lo porque o Set-Cookie está declarado na
// spec: o handler o entrega como um campo tipado da resposta, e o strict
// server escreve o cabeçalho.
//
//   - HttpOnly tira o token do alcance do JavaScript, então uma XSS não
//     consegue exfiltrá-lo.
//   - SameSite=Lax faz o navegador não enviar o cookie em requisição de
//     escrita vinda de outro site: é a defesa principal contra CSRF.
//   - Secure é condicional de propósito. Forçado em http de desenvolvimento, o
//     navegador descartaria o cookie EM SILÊNCIO e o login pareceria quebrado
//     sem nenhum erro visível.
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

// ClearedSessionCookie expira o cookie no navegador. Como o JWT não tem
// revogação no servidor, o logout é exatamente isto: o token continua válido
// até o `exp`, mas some do navegador. É a limitação consciente do desenho.
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
