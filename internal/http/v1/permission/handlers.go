package permissionhttp

import (
	"encoding/json"
	"net/http"

	appperm "github.com/ESG-Project/suassu-api/internal/app/permission"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	permissiondto "github.com/ESG-Project/suassu-api/internal/http/dto/permission"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Service define a interface do serviço de Permission para a camada HTTP.
type Service = appperm.ServiceInterface

// Middleware é um alias para os middlewares de autorização injetados.
type Middleware = func(http.Handler) http.Handler

// Routes monta o roteador de permissões (Permission).
//   - requireRead: exige permissão Permission.read
//   - requireCreate: exige permissão Permission.create
//   - requireUpdate: exige permissão Permission.update
//   - requireDelete: exige permissão Permission.delete
//
// Todas as respostas seguem o mesmo shape sem envelope do user-crud (não há
// consumidor no front hoje, mas mantém o contrato caso exista outro cliente).
func Routes(svc Service, requireRead, requireCreate, requireUpdate, requireDelete Middleware) chi.Router {
	r := chi.NewRouter()

	r.With(requireRead).Get("/enterprise", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		list, err := svc.ListByEnterprise(req.Context(), enterpriseID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, permissiondto.ToPermissionResponses(list))
	})

	r.With(requireCreate).Post("/", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}
		enterpriseID := httpmw.EnterpriseID(req.Context())

		var in permissiondto.CreatePermissionRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		p, err := svc.Create(req.Context(), claims.Subject, enterpriseID, appperm.CreateInput{
			FeatureID: in.FeatureID,
			RoleID:    in.RoleID,
			Create:    in.Create,
			Read:      in.Read,
			Update:    in.Update,
			Delete:    in.Erase,
		})
		if err != nil {
			httperr.HandlePrivilegeAware(w, req, err)
			return
		}
		writeJSON(w, http.StatusCreated, permissiondto.ToPermissionResponse(p))
	})

	r.With(requireUpdate).Put("/", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}
		enterpriseID := httpmw.EnterpriseID(req.Context())

		var in permissiondto.UpdatePermissionRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		p, err := svc.Update(req.Context(), claims.Subject, enterpriseID, appperm.UpdateInput{
			ID:        in.ID,
			FeatureID: in.FeatureID,
			RoleID:    in.RoleID,
			Create:    in.Create,
			Read:      in.Read,
			Update:    in.Update,
			Delete:    in.Erase,
		})
		if err != nil {
			httperr.HandlePrivilegeAware(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, permissiondto.ToPermissionResponse(p))
	})

	r.With(requireDelete).Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		enterpriseID := httpmw.EnterpriseID(req.Context())

		if err := svc.Delete(req.Context(), id, enterpriseID); err != nil {
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
