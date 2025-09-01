package userhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	appuser "github.com/ESG-Project/suassu-api/internal/app/user"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainuser "github.com/ESG-Project/suassu-api/internal/domain/user"
	userdto "github.com/ESG-Project/suassu-api/internal/http/dto/user"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/ESG-Project/suassu-api/internal/http/pagination"
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

		limit := parseInt32(req.URL.Query().Get("limit"), 50)
		if limit > 1000 {
			limit = 1000
		}

		cursorStr := req.URL.Query().Get("cursor")

		var after *domainuser.UserCursorKey
		if cursorStr != "" {
			if err := pagination.Decode(cursorStr, &after); err != nil {
				httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid cursor format"))
				return
			}
		}

		users, pageInfo, err := svc.List(req.Context(), enterpriseID, limit, after)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		out := make([]userdto.UserOut, 0, len(users))
		for _, u := range users {
			out = append(out, *userdto.ToUserOut(&u))
		}

		var nextCursor *string
		if pageInfo.Next != nil {
			if encoded, err := pagination.Encode(pageInfo.Next); err == nil {
				nextCursor = &encoded
			}
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
