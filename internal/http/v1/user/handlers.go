package userhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"encoding/base64"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	"github.com/ESG-Project/suassu-api/internal/http/dto"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/ESG-Project/suassu-api/internal/http/response"
	"github.com/go-chi/chi/v5"
)

// Service define a interface do serviço de usuários para a camada HTTP
type Service = appuser.ServiceInterface

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

		var after *domainuser.UserCursorKey
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

		out := make([]dto.UserOut, 0, len(users))
		for _, u := range users {
			var aOut *dto.AddressOut
			if u.Address != nil {
				aOut = &dto.AddressOut{
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
			}

			out = append(out, dto.UserOut{
				ID: u.ID, Name: u.Name, Email: u.Email, Document: u.Document,
				Phone: u.Phone, AddressID: u.AddressID, Address: aOut, RoleID: u.RoleID, EnterpriseID: u.EnterpriseID,
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
