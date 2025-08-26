package userhttp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	Create(ctx context.Context, in appuser.CreateInput) (string, error)
	GetByEmail(ctx context.Context, email string) (*domainuser.User, error)
	List(ctx context.Context, limit, offset int32) ([]*domainuser.User, error)
}

func Routes(svc Service) chi.Router {
	r := chi.NewRouter()

	// POST /users
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var in appuser.CreateInput
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		id, err := svc.Create(req.Context(), in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": id}) // Mudar para 201
	})

	// GET /users?limit=50&offset=0
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		limit := parseInt32(req.URL.Query().Get("limit"), 50)
		offset := parseInt32(req.URL.Query().Get("offset"), 0)

		users, err := svc.List(req.Context(), limit, offset)
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

	// GET /users/by-email?email=foo@bar.com
	r.Get("/by-email", func(w http.ResponseWriter, req *http.Request) {
		email := req.URL.Query().Get("email")
		u, err := svc.GetByEmail(req.Context(), email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || err.Error() == "no rows in result set" {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id": u.ID, "name": u.Name, "email": u.Email,
			"document": u.Document, "phone": u.Phone,
			"addressId": u.AddressID, "roleId": u.RoleID, "enterpriseId": u.EnterpriseID,
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
