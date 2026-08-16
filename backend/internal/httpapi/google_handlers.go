package httpapi

import (
	"context"
	"errors"
	"net/url"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

// StartGoogleLogin redireciona para a tela de consentimento do Google. É
// navegação de navegador: nenhum código do frontend participa do fluxo, e por
// isso o client_secret nunca chega ao bundle.
func (a *API) StartGoogleLogin(_ context.Context, _ gen.StartGoogleLoginRequestObject) (gen.StartGoogleLoginResponseObject, error) {
	// Sem credenciais, volta para o login com o motivo em vez de exibir um
	// JSON cru na barra de endereços.
	if a.google == nil {
		destino := a.publicBaseURL + "/login?erro=google_nao_configurado"
		return gen.StartGoogleLogin302Response{
			Headers: gen.StartGoogleLogin302ResponseHeaders{Location: &destino},
		}, nil
	}

	state, cookie, err := a.google.NewState()
	if err != nil {
		return nil, err
	}

	valor := cookie.String()
	destino := a.google.AuthCodeURL(state)

	return gen.StartGoogleLogin302Response{
		Headers: gen.StartGoogleLogin302ResponseHeaders{
			Location:  &destino,
			SetCookie: &valor,
		},
	}, nil
}

// GoogleCallback valida o state, troca o código e abre a sessão. O destino é
// sempre uma tela do frontend, com o motivo do erro em query param.
func (a *API) GoogleCallback(ctx context.Context, req gen.GoogleCallbackRequestObject) (gen.GoogleCallbackResponseObject, error) {
	if a.google == nil {
		destino := a.publicBaseURL + "/login?erro=google_nao_configurado"
		return gen.GoogleCallback302Response{
			Headers: gen.GoogleCallback302ResponseHeaders{Location: &destino},
		}, nil
	}

	limparState := a.google.ClearedStateCookie().String()

	// ?error=access_denied é a pessoa cancelando na tela do Google.
	if req.Params.Error != nil && *req.Params.Error != "" {
		return a.redirecionarComErro("acesso_negado", limparState), nil
	}

	if req.Params.Code == nil || *req.Params.Code == "" || req.Params.State == nil {
		return a.redirecionarComErro("resposta_invalida", limparState), nil
	}

	stateNoCookie := ""
	if req.Params.OauthState != nil {
		stateNoCookie = *req.Params.OauthState
	}

	user, err := a.google.Complete(ctx, *req.Params.Code, *req.Params.State, stateNoCookie)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrStateMismatch):
			return a.redirecionarComErro("state_invalido", limparState), nil
		case errors.Is(err, auth.ErrEmailNotVerified):
			return a.redirecionarComErro("email_nao_verificado", limparState), nil
		default:
			return nil, err
		}
	}

	sessao, err := a.newSessionCookie(user.ID)
	if err != nil {
		return nil, err
	}

	destino := a.publicBaseURL + "/"
	return gen.GoogleCallback302Response{
		Headers: gen.GoogleCallback302ResponseHeaders{
			Location: &destino,
			// Só o cookie de sessão: `Set-Cookie` não pode ser dobrado num
			// valor separado por vírgula. O cookie de state expira sozinho em
			// 10 minutos e é de uso único.
			SetCookie: &sessao,
		},
	}, nil
}

func (a *API) redirecionarComErro(motivo, limparState string) gen.GoogleCallback302Response {
	destino := a.publicBaseURL + "/login?erro=" + url.QueryEscape(motivo)
	return gen.GoogleCallback302Response{
		Headers: gen.GoogleCallback302ResponseHeaders{
			Location:  &destino,
			SetCookie: &limparState,
		},
	}
}

func strPtr(s string) *string { return &s }
