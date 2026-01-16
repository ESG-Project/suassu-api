package cookie

import (
	"net/http"
	"time"
)

const (
	// RefreshTokenCookieName é o nome do cookie que armazena o refresh token
	RefreshTokenCookieName = "refresh_token"
)

// Config contém as configurações para cookies
type Config struct {
	Domain string // domínio do cookie (ex: ".suassu.com")
	Secure bool   // true para HTTPS only
}

// Manager gerencia operações de cookie
type Manager struct {
	config Config
}

// NewManager cria um novo gerenciador de cookies
func NewManager(cfg Config) *Manager {
	return &Manager{config: cfg}
}

// SetRefreshToken define o cookie HttpOnly com o refresh token
func (m *Manager) SetRefreshToken(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     "/", // disponível em todas as rotas
		Domain:   m.config.Domain,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,                    // NÃO acessível via JavaScript
		Secure:   m.config.Secure,         // HTTPS only em produção
		SameSite: http.SameSiteStrictMode, // Proteção contra CSRF
	})
}

// GetRefreshToken obtém o refresh token do cookie
func (m *Manager) GetRefreshToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshTokenCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearRefreshToken remove o cookie de refresh token
func (m *Manager) ClearRefreshToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		Domain:   m.config.Domain,
		Expires:  time.Unix(0, 0), // expira imediatamente
		MaxAge:   -1,              // remove o cookie
		HttpOnly: true,
		Secure:   m.config.Secure,
		SameSite: http.SameSiteStrictMode,
	})
}
