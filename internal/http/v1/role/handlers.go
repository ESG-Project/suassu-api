package rolehttp

import (
	"encoding/json"
	"net/http"

	approle "github.com/ESG-Project/suassu-api/internal/app/role"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	roledto "github.com/ESG-Project/suassu-api/internal/http/dto/role"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Service define a interface do serviço de Role para a camada HTTP.
type Service = approle.ServiceInterface

// Middleware é um alias para os middlewares de autorização injetados.
type Middleware = func(http.Handler) http.Handler

// Routes monta o roteador de papéis (Role).
//   - requireRead: exige permissão Role.read (GET /role/enterprise)
//   - requireAssignable: qualquer uma das features Client/Technician/User/Financial
//     em 'read' (GET /role/assignable) — gate deliberadamente permissivo: a
//     resposta já vem filtrada ao que o próprio ator pode atribuir, então não
//     expõe nada além disso. Ver approle.Service.ListAssignable.
//   - requireCreate: exige permissão Role.create
//   - requireDelete: exige permissão Role.delete
//
// GET /role/enterprise e GET /role/assignable respondem sem envelope (array
// no corpo), replicando o contrato herdado do user-crud — o front já consome
// essas duas rotas nesse formato hoje.
func Routes(svc Service, requireRead, requireAssignable, requireCreate, requireDelete Middleware) chi.Router {
	r := chi.NewRouter()

	r.With(requireRead).Get("/enterprise", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		roles, err := svc.List(req.Context(), enterpriseID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, roledto.ToRoleResponses(roles))
	})

	r.With(requireAssignable).Get("/assignable", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}
		enterpriseID := httpmw.EnterpriseID(req.Context())
		roles, err := svc.ListAssignable(req.Context(), claims.Subject, enterpriseID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, roledto.ToRoleResponses(roles))
	})

	r.With(requireCreate).Post("/", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())

		var in roledto.CreateRoleRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		role, err := svc.Create(req.Context(), enterpriseID, in.Title)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusCreated, roledto.ToRoleResponse(role))
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
