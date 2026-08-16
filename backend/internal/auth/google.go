package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/google/uuid"
)

const (
	// StateCookieName guarda o anti-CSRF do fluxo, com vida curta.
	StateCookieName = "oauth_state"
	stateTTL        = 10 * time.Minute
	// stateBytes: 32 bytes de crypto/rand, o mesmo padrão do segredo de sessão.
	stateBytes = 32

	userInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
)

var (
	ErrGoogleNotConfigured = errors.New("login com Google não está configurado")
	ErrStateMismatch       = errors.New("parâmetro state inválido")
	ErrEmailNotVerified    = errors.New("e-mail não verificado no Google")
	ErrGoogleDenied        = errors.New("acesso negado no Google")
)

// GoogleAuth encapsula o fluxo Authorization Code.
type GoogleAuth struct {
	oauth  *oauth2.Config
	users  UserRepository
	secure bool
}

// NewGoogleAuth devolve nil quando não há credenciais. O chamador trata nil
// como "não configurado" e responde 501 — o resto do app segue funcionando.
func NewGoogleAuth(clientID, clientSecret, redirectURL string, users UserRepository, secure bool) *GoogleAuth {
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil
	}

	return &GoogleAuth{
		oauth: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			// RedirectURL vem da configuração e é FIXO, registrado no console
			// do Google. Nunca de query param: aceitar destino do cliente
			// viraria open redirect, e pior — o código de autorização seria
			// entregue no domínio do atacante.
			RedirectURL: redirectURL,
			Scopes:      []string{"openid", "email", "profile"},
			Endpoint:    google.Endpoint,
		},
		users:  users,
		secure: secure,
	}
}

// NewState gera o parâmetro state e o cookie que o guarda.
//
// É o anti-CSRF do fluxo. Sem ele existe o *login CSRF*: o atacante inicia o
// login com a conta Google dele, captura o `code` do callback e faz a vítima
// abrir aquela URL — a vítima termina logada NA CONTA DO ATACANTE e passa a
// cadastrar dados nela, que o atacante depois lê. O state amarra o callback ao
// navegador que iniciou o fluxo.
func (g *GoogleAuth) NewState() (string, *http.Cookie, error) {
	buf := make([]byte, stateBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, fmt.Errorf("gerar state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(buf)

	return state, &http.Cookie{
		Name:     StateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   int(stateTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   g.secure,
	}, nil
}

// AuthCodeURL monta o destino da tela de consentimento.
func (g *GoogleAuth) AuthCodeURL(state string) string {
	return g.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// ClearedStateCookie expira o cookie de state depois do callback: ele é de uso
// único e não tem por que sobreviver ao fluxo.
func (g *GoogleAuth) ClearedStateCookie() *http.Cookie {
	return &http.Cookie{
		Name:     StateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   g.secure,
	}
}

// googleProfile é o subconjunto do userinfo que nos interessa.
type googleProfile struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

// Complete valida o state, troca o código por token e resolve o usuário.
func (g *GoogleAuth) Complete(ctx context.Context, code, stateFromQuery, stateFromCookie string) (User, error) {
	// Comparação antes de qualquer chamada de rede: não trocamos o código se o
	// fluxo não começou neste navegador.
	if stateFromCookie == "" || stateFromQuery != stateFromCookie {
		return User{}, ErrStateMismatch
	}

	// A troca do código acontece AQUI, no servidor. O client_secret só existe
	// no backend: o navegador nunca vê e o bundle do frontend nunca contém.
	token, err := g.oauth.Exchange(ctx, code)
	if err != nil {
		return User{}, fmt.Errorf("trocar código por token: %w", err)
	}

	profile, err := g.fetchProfile(ctx, token)
	if err != nil {
		return User{}, err
	}

	// O item mais importante desta seção, porque o desenho vincula conta
	// Google a conta de senha PELO E-MAIL. Sem esta checagem: alguém cria uma
	// conta Google declarando um e-mail que não é dele, entra no app, o app
	// encontra o usuário com aquele e-mail e entrega a conta. O Google só
	// marca email_verified para Gmail ou domínio verificado no Workspace, então
	// esta flag é o que separa vinculação de sequestro.
	if !profile.EmailVerified {
		return User{}, ErrEmailNotVerified
	}

	email, err := normalizeEmail(profile.Email)
	if err != nil {
		return User{}, err
	}

	return g.upsert(ctx, profile, email)
}

func (g *GoogleAuth) fetchProfile(ctx context.Context, token *oauth2.Token) (googleProfile, error) {
	client := g.oauth.Client(ctx, token)
	client.Timeout = 10 * time.Second

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return googleProfile{}, fmt.Errorf("consultar perfil no Google: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return googleProfile{}, fmt.Errorf("perfil do Google devolveu %d", resp.StatusCode)
	}

	var profile googleProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return googleProfile{}, fmt.Errorf("interpretar perfil do Google: %w", err)
	}
	if profile.Sub == "" || profile.Email == "" {
		return googleProfile{}, errors.New("perfil do Google incompleto")
	}
	return profile, nil
}

// upsert resolve as três situações possíveis: usuário já vinculado, usuário
// existente com senha (vincula), e usuário novo.
func (g *GoogleAuth) upsert(ctx context.Context, profile googleProfile, email string) (User, error) {
	existente, err := g.users.FindByEmail(ctx, email)

	switch {
	case err == nil:
		// Conta com este e-mail já existe. Vinculamos em vez de bloquear:
		// bloquear gera o pior atendimento possível — "já existe conta com
		// esse e-mail, mas não te digo como entrar". A alternativa mais rígida
		// seria exigir login por senha primeiro e vincular o Google de dentro
		// da conta; é o que eu faria com dado sensível.
		if existente.GoogleID == "" {
			existente.GoogleID = profile.Sub
			existente.UpdatedAt = time.Now().UTC()
			if err := g.users.LinkGoogleID(ctx, existente.ID, profile.Sub); err != nil {
				return User{}, err
			}
		}
		return existente, nil

	case errors.Is(err, ErrUserNotFound):
		now := time.Now().UTC()
		novo := User{
			ID:    uuid.NewString(),
			Email: email,
			Name:  nomeOuEmail(profile.Name, email),
			// PasswordHash vazio: usuário só-Google não entra por senha, e o
			// Login já recusa quem não tem hash.
			GoogleID:  profile.Sub,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := g.users.Create(ctx, novo); err != nil {
			return User{}, err
		}
		return novo, nil

	default:
		return User{}, err
	}
}

func nomeOuEmail(nome, email string) string {
	if n := strings.TrimSpace(nome); n != "" {
		return n
	}
	return email
}
