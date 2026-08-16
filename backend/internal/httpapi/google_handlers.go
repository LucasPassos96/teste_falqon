package httpapi

import (
	"context"
	"errors"
	"net/url"

	"github.com/LucasPassos96/teste_falqon/backend/internal/auth"
	"github.com/LucasPassos96/teste_falqon/backend/internal/httpapi/gen"
)

// StartGoogleLogin redireciona para a tela de consentimento do Google.
//
// Responde 302 e não JSON: é navegação de navegador. O frontend só precisa de
// um link apontando para cá — nenhum código TypeScript participa do fluxo, e é
// por isso que o client_secret nunca chega ao bundle.
func (a *API) StartGoogleLogin(_ context.Context, _ gen.StartGoogleLoginRequestObject) (gen.StartGoogleLoginResponseObject, error) {
	// Sem credenciais, manda de volta para a tela de login com o motivo. O
	// visitante nunca vê um JSON cru na barra de endereços, e o app segue
	// utilizável por e-mail e senha.
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

// GoogleCallback fecha o fluxo: valida o state, troca o código e abre a sessão.
//
// Também responde 302 — o destino final é sempre uma tela do frontend, com o
// motivo do erro em query param quando algo falha. Isso mantém o visitante
// dentro da aplicação em vez de deixá-lo olhando um JSON no navegador.
func (a *API) GoogleCallback(ctx context.Context, req gen.GoogleCallbackRequestObject) (gen.GoogleCallbackResponseObject, error) {
	if a.google == nil {
		destino := a.publicBaseURL + "/login?erro=google_nao_configurado"
		return gen.GoogleCallback302Response{
			Headers: gen.GoogleCallback302ResponseHeaders{Location: &destino},
		}, nil
	}

	limparState := a.google.ClearedStateCookie().String()

	// O Google devolve ?error=access_denied quando a pessoa clica em cancelar
	// na tela de consentimento. Não é falha, é desistência.
	if req.Params.Error != nil && *req.Params.Error != "" {
		return a.redirecionarComErro("acesso_negado", limparState), nil
	}

	if req.Params.Code == nil || *req.Params.Code == "" || req.Params.State == nil {
		return a.redirecionarComErro("resposta_invalida", limparState), nil
	}

	// O cookie chega tipado porque está declarado como parâmetro na spec.
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
			// Falha inesperada (rede, Google fora do ar) vira erro genérico na
			// tela; o detalhe fica no log.
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
			// Só o cookie de sessão. `Set-Cookie` é o único cabeçalho que NÃO
			// pode ser dobrado em um valor separado por vírgula, e o tipo
			// gerado carrega um valor só — juntar os dois produziria um
			// cabeçalho malformado que o navegador descartaria inteiro.
			//
			// O cookie de state fica para trás e expira sozinho em 10 minutos.
			// Não é problema: ele é de uso único, o valor já foi conferido, e
			// o código de autorização do Google também é de uso único, então
			// não há o que reaproveitar.
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
