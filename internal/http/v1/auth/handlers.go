package authhttp

import (
	"context"
	"encoding/json"
	"net/http"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	SignIn(ctx context.Context, in appauth.SignInInput) (appauth.SignInOutput, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// --------- REGISTRADORES ---------

// RegisterPublic registra rotas públicas de /auth (sem JWT)
func (h *Handler) RegisterPublic(r chi.Router) {
	r.Post("/login", h.login)
}

// RegisterPrivate registra rotas privadas de /auth (com JWT)
func (h *Handler) RegisterPrivate(r chi.Router) {
	r.Get("/me", h.me)
	// r.Post("/logout", h.logout)   // futuro
	// r.Post("/refresh", h.refresh) // futuro
}

// --------- HANDLERS ---------

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in appauth.SignInInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	out, err := h.svc.SignIn(r.Context(), in)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpmw.ClaimsFromCtx(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sub":          claims.Subject,
		"email":        claims.Email,
		"name":         claims.Name,
		"enterpriseId": claims.EnterpriseID,
		"roleId":       claims.RoleID,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
