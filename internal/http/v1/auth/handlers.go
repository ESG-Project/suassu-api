package authhttp

import (
	"context"
	"encoding/json"
	"net/http"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/ESG-Project/suassu-api/internal/http/cookie"
	userdto "github.com/ESG-Project/suassu-api/internal/http/dto/user"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	SignIn(ctx context.Context, in appauth.SignInInput) (appauth.SignInOutput, error)
	Refresh(ctx context.Context, in appauth.RefreshInput) (appauth.RefreshOutput, error)
	Logout(ctx context.Context, userID string) error
	GetMe(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error)
	GetMyPermissions(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error)
	ValidateToken(ctx context.Context, token string) (bool, error)
}

type Handler struct {
	svc       Service
	cookieMgr *cookie.Manager
}

func NewHandler(svc Service, cookieMgr *cookie.Manager) *Handler {
	return &Handler{svc: svc, cookieMgr: cookieMgr}
}

// --------- REGISTRADORES ---------

// RegisterPublic registra rotas públicas de /auth (sem JWT)
func (h *Handler) RegisterPublic(r chi.Router) {
	r.Post("/login", h.login)
	r.Post("/refresh", h.refresh)
	r.Get("/validate-token", h.validateToken)
}

// RegisterPrivate registra rotas privadas de /auth (com JWT)
func (h *Handler) RegisterPrivate(r chi.Router) {
	r.Get("/me", h.me)
	r.Get("/my-permissions", h.myPermissions)
	r.Post("/logout", h.logout)
}

// --------- HANDLERS ---------

// loginResponse é a resposta do endpoint de login (sem refresh token no body)
type loginResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"` // segundos até expirar
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in appauth.SignInInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httperr.Handle(w, r, apperr.New(apperr.CodeInvalid, "invalid request body"))
		return
	}
	out, err := h.svc.SignIn(r.Context(), in)
	if err != nil {
		httperr.Handle(w, r, err)
		return
	}

	// Setar o refresh token como cookie HttpOnly (seguro)
	if out.RefreshToken != "" && out.RefreshExpiresAt != nil {
		h.cookieMgr.SetRefreshToken(w, out.RefreshToken, *out.RefreshExpiresAt)
	}

	// Retornar apenas o access token no body (refresh token vai no cookie)
	resp := loginResponse{
		AccessToken: out.AccessToken,
		ExpiresIn:   out.ExpiresIn,
	}
	writeJSON(w, http.StatusOK, resp)
}

// refreshRequest representa o body da requisição de refresh (fallback se cookie não existir)
type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// refreshResponse é a resposta do endpoint de refresh (sem refresh token no body)
type refreshResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"` // segundos até expirar
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	// Primeiro, tenta ler o refresh token do cookie HttpOnly
	cookieToken, err := h.cookieMgr.GetRefreshToken(r)
	if err == nil && cookieToken != "" {
		refreshToken = cookieToken
	} else {
		// Fallback: tenta ler do body (para compatibilidade)
		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "refresh token is required"))
		return
	}

	out, err := h.svc.Refresh(r.Context(), appauth.RefreshInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		// Se o refresh falhar, limpa o cookie
		h.cookieMgr.ClearRefreshToken(w)
		httperr.Handle(w, r, err)
		return
	}

	// Setar o novo refresh token como cookie HttpOnly
	if out.RefreshToken != "" && out.RefreshExpiresAt != nil {
		h.cookieMgr.SetRefreshToken(w, out.RefreshToken, *out.RefreshExpiresAt)
	}

	// Retornar apenas o access token no body
	resp := refreshResponse{
		AccessToken: out.AccessToken,
		ExpiresIn:   out.ExpiresIn,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpmw.ClaimsFromCtx(r.Context())
	if !ok {
		httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "authentication required"))
		return
	}

	if err := h.svc.Logout(r.Context(), claims.Subject); err != nil {
		httperr.Handle(w, r, err)
		return
	}

	// Limpar o cookie de refresh token
	h.cookieMgr.ClearRefreshToken(w)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpmw.ClaimsFromCtx(r.Context())
	if !ok {
		httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "authentication required"))
		return
	}

	enterpriseID := httpmw.EnterpriseID(r.Context())
	if enterpriseID == "" {
		httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "enterprise ID required"))
		return
	}

	me, err := h.svc.GetMe(r.Context(), claims.Subject, enterpriseID)
	if err != nil {
		httperr.Handle(w, r, err)
		return
	}

	// Converter para DTO HTTP
	meOut := userdto.ToMeOut(me)
	writeJSON(w, http.StatusOK, meOut)
}

func (h *Handler) myPermissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpmw.ClaimsFromCtx(r.Context())
	if !ok {
		httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "authentication required"))
		return
	}

	enterpriseID := httpmw.EnterpriseID(r.Context())
	if enterpriseID == "" {
		httperr.Handle(w, r, apperr.New(apperr.CodeUnauthorized, "enterprise ID required"))
		return
	}

	permissions, err := h.svc.GetMyPermissions(r.Context(), claims.Subject, enterpriseID)
	if err != nil {
		httperr.Handle(w, r, err)
		return
	}

	// Converter para DTO HTTP
	permissionsOut := userdto.ToMyPermissionsOut(permissions)
	writeJSON(w, http.StatusOK, permissionsOut)
}

func (h *Handler) validateToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		httperr.Handle(w, r, apperr.New(apperr.CodeInvalid, "authorization header is required"))
		return
	}

	// Remove "Bearer " prefix if present
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if token == "" {
		httperr.Handle(w, r, apperr.New(apperr.CodeInvalid, "token is required"))
		return
	}

	valid, err := h.svc.ValidateToken(r.Context(), token)
	if err != nil {
		httperr.Handle(w, r, err)
		return
	}

	if valid {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"valid": false})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
