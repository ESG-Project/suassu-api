package authhttp

import (
	"context"
	"encoding/json"
	"net/http"

	appauth "github.com/ESG-Project/suassu-api/internal/app/auth"
	"github.com/go-chi/chi/v5"
)

type service interface {
	SignIn(ctx context.Context, in appauth.SignInInput) (appauth.SignInOutput, error)
}

func Routes(svc *appauth.Service) chi.Router {
	r := chi.NewRouter()

	r.Post("/login", func(w http.ResponseWriter, req *http.Request) {
		var in appauth.SignInInput
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		out, err := svc.SignIn(req.Context(), in)
		if err != nil {
			// por segurança, não detalhe se foi email ou senha
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, out)
	})

	// opcional: endpoint /auth/me que apenas retorna o token (ou carrega o usuário)
	r.Get("/me", func(w http.ResponseWriter, req *http.Request) {
		// TODO: Implementar middleware de autenticação para este endpoint
		// Por enquanto, retorna erro de não implementado
		http.Error(w, "not implemented", http.StatusNotImplemented)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
