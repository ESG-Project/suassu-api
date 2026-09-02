package legacyuserhttp

import (
	"encoding/json"
	"net/http"

	"github.com/ESG-Project/suassu-api/internal/app/address"
	appadmin "github.com/ESG-Project/suassu-api/internal/app/adminuser"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	legacyuserdto "github.com/ESG-Project/suassu-api/internal/http/dto/legacyuser"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Middleware é um alias para os middlewares de autorização injetados.
type Middleware = func(http.Handler) http.Handler

// RegisterRoutes registra as rotas legadas de usuário (singular /user, mais
// /user-enterprise) diretamente no router recebido — os paths não seguem um
// prefixo comum (ex.: /user-enterprise não é sub-rota de /user), por isso não
// é um sub-router montável como os demais módulos.
//
//   - requireClientRead: Client.read (GET /user/client/enterprise)
//   - requireUserRead: User.read (GET /user/:id, /user-enterprise, /user/import-file/enterprise)
//   - requireCreate: qualquer de Client/Technician/User/Financial em 'create'
//   - requireUpdate: idem em 'update'
//   - requireDelete: idem em 'delete'
func RegisterRoutes(r chi.Router, svc *appadmin.Service, requireClientRead, requireUserRead, requireCreate, requireUpdate, requireDelete Middleware) {
	r.With(requireClientRead).Get("/user/client/enterprise", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		list, err := svc.ListClients(req.Context(), enterpriseID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, legacyuserdto.ToClientListResponses(list))
	})

	r.With(requireUserRead).Get("/user-enterprise", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		list, err := svc.ListUsers(req.Context(), enterpriseID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, legacyuserdto.ToUserListItems(list))
	})

	r.With(requireUserRead).Get("/user/import-file/enterprise", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", `attachment; filename="client_register_template.xlsx"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(clientRegisterTemplateXLSX)
	})

	r.With(requireUserRead).Get("/user/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		enterpriseID := httpmw.EnterpriseID(req.Context())

		out, err := svc.GetByID(req.Context(), enterpriseID, id)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, legacyuserdto.ToUserDetailResponse(out))
	})

	r.With(requireCreate).Post("/user", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}
		enterpriseID := httpmw.EnterpriseID(req.Context())

		var in legacyuserdto.CreateUserRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		out, err := svc.Create(req.Context(), claims.Subject, enterpriseID, createInputFromRequest(in))
		if err != nil {
			httperr.HandlePrivilegeAware(w, req, err)
			return
		}
		writeJSON(w, http.StatusCreated, legacyuserdto.ToUserResponse(out))
	})

	r.With(requireCreate).Post("/user/file", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}
		enterpriseID := httpmw.EnterpriseID(req.Context())

		var body struct {
			Users []legacyuserdto.CreateUserRequest `json:"users"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		items := make([]appadmin.CreateInput, 0, len(body.Users))
		for _, u := range body.Users {
			items = append(items, createInputFromRequest(u))
		}

		if err := svc.CreateMany(req.Context(), claims.Subject, enterpriseID, items); err != nil {
			httperr.HandlePrivilegeAware(w, req, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"message": "Usuários criados com sucesso!"})
	})

	r.With(requireUpdate).Put("/user", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}
		enterpriseID := httpmw.EnterpriseID(req.Context())

		var in legacyuserdto.UpdateUserRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		var addr *address.CreateInput
		if in.HasAddress() {
			addr = &address.CreateInput{
				ZipCode: strVal(in.ZipCode), State: strVal(in.State), City: strVal(in.City),
				Neighborhood: strVal(in.Neighborhood), Street: strVal(in.Street), Num: strVal(in.Num),
				Latitude: in.Latitude, Longitude: in.Longitude, AddInfo: in.AddInfo,
			}
		}

		out, err := svc.Update(req.Context(), claims.Subject, enterpriseID, appadmin.UpdateInput{
			ID: in.ID, Name: in.Name, Document: in.Document, Email: in.Email, Phone: in.Phone,
			RoleIDOrTitle: in.RoleID, Address: addr,
			ProRegister: in.ProRegister, Graduation: in.Graduation, CTF: in.CTF, FantasyName: in.FantasyName,
		})
		if err != nil {
			httperr.HandlePrivilegeAware(w, req, err)
			return
		}
		// Replica o status 201 do user-crud para PUT /user (histórico, não 200).
		writeJSON(w, http.StatusCreated, legacyuserdto.ToUserResponse(out))
	})

	r.With(requireDelete).Delete("/user/{id}", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}
		id := chi.URLParam(req, "id")
		enterpriseID := httpmw.EnterpriseID(req.Context())

		if err := svc.Delete(req.Context(), claims.Subject, enterpriseID, id); err != nil {
			httperr.HandlePrivilegeAware(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Usuário deletado com sucesso"})
	})
}

func createInputFromRequest(in legacyuserdto.CreateUserRequest) appadmin.CreateInput {
	var addr *address.CreateInput
	if in.HasAddress() {
		addr = &address.CreateInput{
			ZipCode: strVal(in.ZipCode), State: strVal(in.State), City: strVal(in.City),
			Neighborhood: strVal(in.Neighborhood), Street: strVal(in.Street), Num: strVal(in.Num),
			Latitude: in.Latitude, Longitude: in.Longitude, AddInfo: in.AddInfo,
		}
	}
	return appadmin.CreateInput{
		Name: in.Name, Document: in.Document, Email: in.Email, Password: in.Password, Phone: in.Phone,
		RoleIDOrTitle: in.RoleID, Address: addr,
		ProRegister: in.ProRegister, Graduation: in.Graduation, CTF: in.CTF, FantasyName: in.FantasyName,
	}
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
