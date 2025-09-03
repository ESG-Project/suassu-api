package authhttp

import (
	"context"
	"encoding/json"
	"net/http"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	userdto "github.com/ESG-Project/suassu-api/internal/http/dto/user"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	SignIn(ctx context.Context, in appauth.SignInInput) (appauth.SignInOutput, error)
	GetMe(ctx context.Context, userID string, enterpriseID string) (*types.UserWithDetails, error)
	GetMyPermissions(ctx context.Context, userID string, enterpriseID string) (*types.UserPermissions, error)
	ValidateToken(ctx context.Context, token string) (bool, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// --------- REGISTRADORES ---------

// RegisterPublic registra rotas públicas de /auth (sem JWT)
func (h *Handler) RegisterPublic(r chi.Router) {
	r.Post("/login", h.login)
	r.Get("/validate-token", h.validateToken)
}

// RegisterPrivate registra rotas privadas de /auth (com JWT)
func (h *Handler) RegisterPrivate(r chi.Router) {
	r.Get("/me", h.me)
	r.Get("/my-permissions", h.myPermissions)
	// r.Post("/logout", h.logout)   // futuro
	// r.Post("/refresh", h.refresh) // futuro
}

// --------- HANDLERS ---------

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
	writeJSON(w, http.StatusOK, out)
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
