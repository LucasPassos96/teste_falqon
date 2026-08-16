package httpapi

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/forms"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

// API implementa a interface gerada a partir de api/openapi.yaml. Cada
// operationId da spec vira um método aqui, e o código não compila enquanto
// faltar algum.
type API struct {
	auth   *auth.Service
	forms  *forms.Service
	public *forms.PublicService
	// google é nil quando não há credenciais configuradas; os handlers
	// respondem 501 nesse caso.
	google *auth.GoogleAuth
	// publicBaseURL decide se o cookie sai com Secure e, adiante, monta o link
	// público do formulário.
	publicBaseURL string
}

// Falha em tempo de compilação se API deixar de satisfazer o contrato.
var _ gen.StrictServerInterface = (*API)(nil)

func (a *API) GetHealth(_ context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	return gen.GetHealth200JSONResponse{Status: "ok"}, nil
}

func (a *API) Register(ctx context.Context, req gen.RegisterRequestObject) (gen.RegisterResponseObject, error) {
	user, err := a.auth.Register(ctx, req.Body.Name, string(req.Body.Email), req.Body.Password)
	if err != nil {
		return nil, err
	}

	// Cadastro já autentica: obrigar um login logo depois de criar a conta é
	// atrito sem ganho de segurança.
	cookie, err := a.newSessionCookie(user.ID)
	if err != nil {
		return nil, err
	}

	body, err := toGenUser(user)
	if err != nil {
		return nil, err
	}

	return gen.Register201JSONResponse{
		Body:    body,
		Headers: gen.Register201ResponseHeaders{SetCookie: &cookie},
	}, nil
}

func (a *API) Login(ctx context.Context, req gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	user, err := a.auth.Login(ctx, string(req.Body.Email), req.Body.Password)
	if err != nil {
		return nil, err
	}

	cookie, err := a.newSessionCookie(user.ID)
	if err != nil {
		return nil, err
	}

	body, err := toGenUser(user)
	if err != nil {
		return nil, err
	}

	return gen.Login200JSONResponse{
		Body:    body,
		Headers: gen.Login200ResponseHeaders{SetCookie: &cookie},
	}, nil
}

func (a *API) Logout(_ context.Context, _ gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	cookie := auth.ClearedSessionCookie(a.publicBaseURL).String()
	return gen.Logout204Response{
		Headers: gen.Logout204ResponseHeaders{SetCookie: &cookie},
	}, nil
}

func (a *API) GetCurrentUser(ctx context.Context, _ gen.GetCurrentUserRequestObject) (gen.GetCurrentUserResponseObject, error) {
	// O identificador vem do context, alimentado pelo middleware a partir do
	// cookie assinado — nunca do payload.
	userID, ok := auth.UserIDFrom(ctx)
	if !ok {
		return nil, auth.ErrInvalidToken
	}

	user, err := a.auth.CurrentUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	body, err := toGenUser(user)
	if err != nil {
		return nil, err
	}
	return gen.GetCurrentUser200JSONResponse(body), nil
}

func (a *API) newSessionCookie(userID string) (string, error) {
	token, expiresAt, err := a.auth.IssueSession(userID)
	if err != nil {
		return "", err
	}
	return auth.SessionCookie(token, expiresAt, a.publicBaseURL).String(), nil
}

func toGenUser(u auth.User) (gen.User, error) {
	id, err := uuid.Parse(u.ID)
	if err != nil {
		// Só acontece se um ID inválido tiver entrado no banco: é bug nosso,
		// não erro do cliente, e vira 500 com o detalhe no log.
		return gen.User{}, fmt.Errorf("id de usuário inválido %q: %w", u.ID, err)
	}
	return gen.User{
		Id:    id,
		Email: types.Email(u.Email),
		Name:  u.Name,
	}, nil
}
