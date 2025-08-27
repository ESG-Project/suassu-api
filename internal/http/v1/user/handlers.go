package userhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	Create(ctx context.Context, enterpriseID string, in appuser.CreateInput) (string, error)
	List(ctx context.Context, enterpriseID string, limit, offset int32) ([]domainuser.User, error)
}

func Routes(svc Service) chi.Router {
	r := chi.NewRouter()

	// POST /users
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		if enterpriseID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var in appuser.CreateInput
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		in.EnterpriseID = enterpriseID

		id, err := svc.Create(req.Context(), enterpriseID, in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": id})
	})

	// GET /users?limit=50&offset=0
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		if enterpriseID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		limit := parseInt32(req.URL.Query().Get("limit"), 50)
		offset := parseInt32(req.URL.Query().Get("offset"), 0)

		users, err := svc.List(req.Context(), enterpriseID, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		type userOut struct {
			ID           string  `json:"id"`
			Name         string  `json:"name"`
			Email        string  `json:"email"`
			Document     string  `json:"document"`
			Phone        *string `json:"phone,omitempty"`
			AddressID    *string `json:"addressId,omitempty"`
			RoleID       *string `json:"roleId,omitempty"`
			EnterpriseID string  `json:"enterpriseId"`
		}

		out := make([]userOut, 0, len(users))
		for _, u := range users {
			out = append(out, userOut{
				ID: u.ID, Name: u.Name, Email: u.Email, Document: u.Document,
				Phone: u.Phone, AddressID: u.AddressID, RoleID: u.RoleID, EnterpriseID: u.EnterpriseID,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"items":  out,
			"limit":  limit,
			"offset": offset,
			"count":  len(out),
		})
	})

	return r
}

func parseInt32(s string, def int32) int32 {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return int32(v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
