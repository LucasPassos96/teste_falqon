package httpapi

import (
	"context"

	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

// API implementa a interface gerada a partir de api/openapi.yaml. Cada
// operationId da spec vira um método aqui, e o código não compila enquanto
// faltar algum.
type API struct{}

// Falha em tempo de compilação se API deixar de satisfazer o contrato.
var _ gen.StrictServerInterface = (*API)(nil)

func (a *API) GetHealth(_ context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	return gen.GetHealth200JSONResponse{Status: "ok"}, nil
}
