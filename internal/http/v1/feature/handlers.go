package featurehttp

import (
	"encoding/json"
	"net/http"

	appfeature "github.com/ESG-Project/suassu-api/internal/app/feature"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	"github.com/go-chi/chi/v5"
)

// createFeatureRequest representa o corpo para criar uma nova feature.
type createFeatureRequest struct {
	Name string `json:"name"`
}

// featureResponse representa a resposta de uma feature (mesmo shape do
// user-crud: id, name).
type featureResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Routes monta o roteador de features.
//
// POST /feature e DELETE /feature/:id não exigem permissão no user-crud
// (gap de RBAC já existente lá) — preservado aqui de propósito para manter o
// contrato idêntico; ver plano de migração, seção "Riscos e cuidados".
func Routes(svc *appfeature.Service) chi.Router {
	r := chi.NewRouter()

	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var in createFeatureRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		f, err := svc.Create(req.Context(), in.Name)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusCreated, featureResponse{ID: f.ID, Name: f.Name})
	})

	r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		if err := svc.Delete(req.Context(), id); err != nil {
			httperr.Handle(w, req, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
