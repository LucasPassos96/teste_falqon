package httpapi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

// publicOperations lê da própria spec quais operações declararam
// `security: []`.
//
// Isto é o que impede a declaração de segurança do OpenAPI de ser decorativa:
// a lista não é escrita à mão em lugar nenhum, é derivada do contrato. Rota
// nova sem `security: []` nasce protegida, e abrir uma exige uma linha visível
// no diff da spec.
func publicOperations(spec *openapi3.T) map[string]bool {
	public := make(map[string]bool)
	for _, item := range spec.Paths.Map() {
		for _, op := range item.Operations() {
			// Nil significa "herda o security global". Não-nil e vazio é a
			// declaração explícita de rota aberta.
			if op.Security != nil && len(*op.Security) == 0 {
				public[op.OperationID] = true
			}
		}
	}
	return public
}

// requireSession é um middleware do strict server: ele recebe o operationID,
// e é isso que permite decidir com base na spec em vez de repetir o padrão de
// rota aqui.
func requireSession(svc *auth.Service, public map[string]bool) gen.StrictMiddlewareFunc {
	return func(next gen.StrictHandlerFunc, operationID string) gen.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			if public[operationID] {
				return next(ctx, w, r, request)
			}

			cookie, err := r.Cookie(auth.SessionCookieName)
			if err != nil {
				return nil, auth.ErrInvalidToken
			}

			userID, err := svc.ValidateSession(cookie.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", auth.ErrInvalidToken, err)
			}

			// Respostas autenticadas não podem ficar no cache do navegador nem
			// de um proxy: sem isto, o botão voltar recupera dado privado
			// depois do logout.
			w.Header().Set("Cache-Control", "no-store")

			return next(auth.WithUserID(ctx, userID), w, r, request)
		}
	}
}
