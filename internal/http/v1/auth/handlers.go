package authhttp

import (
	"context"
	"encoding/json"
	"net/http"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/ESG-Project/suassu-api/internal/http/dto"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	SignIn(ctx context.Context, in appauth.SignInInput) (appauth.SignInOutput, error)
	GetMe(ctx context.Context, userID string, enterpriseID string) (*dto.MeOut, error)
	GetMyPermissions(ctx context.Context, userID string, enterpriseID string) (*dto.MyPermissionsOut, error)
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
	writeJSON(w, http.StatusOK, me)
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

	me, err := h.svc.GetMyPermissions(r.Context(), claims.Subject, enterpriseID)
	if err != nil {
		httperr.Handle(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, me)

}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
