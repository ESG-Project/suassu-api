package userhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"encoding/base64"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/ESG-Project/suassu-api/internal/http/response"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	Create(ctx context.Context, enterpriseID string, in appuser.CreateInput) (string, error)
	List(ctx context.Context, enterpriseID string, limit int32, after *postgres.UserCursorKey) ([]domainuser.User, *postgres.PageInfo, error)
}

func Routes(svc Service) chi.Router {
	r := chi.NewRouter()

	// POST /users
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		if enterpriseID == "" {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "enterprise ID required"))
			return
		}

		var in appuser.CreateInput
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		in.EnterpriseID = enterpriseID

		id, err := svc.Create(req.Context(), enterpriseID, in)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusCreated, map[string]string{"id": id}, nil)
	})

	// GET /users?limit=50&cursor=...
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		if enterpriseID == "" {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "enterprise ID required"))
			return
		}

		limit := parseInt32(req.URL.Query().Get("limit"), 50)
		if limit > 1000 {
			limit = 1000
		}

		cursorStr := req.URL.Query().Get("cursor")

		var after *postgres.UserCursorKey
		if cursorStr != "" {
			// Decodificar cursor base64 JSON
			decoded, err := base64.StdEncoding.DecodeString(cursorStr)
			if err != nil {
				httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid cursor format"))
				return
			}
			if err := json.Unmarshal(decoded, &after); err != nil {
				httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid cursor data"))
				return
			}
		}

		users, pageInfo, err := svc.List(req.Context(), enterpriseID, limit, after)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		type addressOut struct {
			ID           string  `json:"id"`
			State        string  `json:"state"`
			ZipCode      string  `json:"zipCode"`
			City         string  `json:"city"`
			Neighborhood string  `json:"neighborhood"`
			Street       string  `json:"street"`
			Num          string  `json:"num"`
			Latitude     *string `json:"latitude,omitempty"`
			Longitude    *string `json:"longitude,omitempty"`
			AddInfo      *string `json:"addInfo,omitempty"`
		}

		type userOut struct {
			ID           string      `json:"id"`
			Name         string      `json:"name"`
			Email        string      `json:"email"`
			Document     string      `json:"document"`
			Phone        *string     `json:"phone,omitempty"`
			AddressID    *string     `json:"addressId,omitempty"`
			Address      *addressOut `json:"address,omitempty"`
			RoleID       *string     `json:"roleId,omitempty"`
			EnterpriseID string      `json:"enterpriseId"`
		}

		out := make([]userOut, 0, len(users))
		for _, u := range users {
			aOut := addressOut{
				ID:           u.Address.ID,
				ZipCode:      u.Address.ZipCode,
				State:        u.Address.State,
				City:         u.Address.City,
				Neighborhood: u.Address.Neighborhood,
				Street:       u.Address.Street,
				Num:          u.Address.Num,
				Latitude:     u.Address.Latitude,
				Longitude:    u.Address.Longitude,
				AddInfo:      u.Address.AddInfo,
			}
			out = append(out, userOut{
				ID: u.ID, Name: u.Name, Email: u.Email, Document: u.Document,
				Phone: u.Phone, AddressID: u.AddressID, Address: &aOut, RoleID: u.RoleID, EnterpriseID: u.EnterpriseID,
			})
		}

		var nextCursor *string
		if pageInfo.Next != nil {
			cursorData, _ := json.Marshal(pageInfo.Next)
			encoded := base64.StdEncoding.EncodeToString(cursorData)
			nextCursor = &encoded
		}

		meta := response.MetaCursor{
			Limit:      int(limit),
			NextCursor: nextCursor,
			HasMore:    pageInfo.HasMore,
		}

		response.JSON(w, http.StatusOK, out, meta)
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
