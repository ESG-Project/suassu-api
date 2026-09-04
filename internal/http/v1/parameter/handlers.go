package parameterhttp

import (
	"encoding/json"
	"net/http"

	appparameter "github.com/ESG-Project/suassu-api/internal/app/parameter"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	parameterdto "github.com/ESG-Project/suassu-api/internal/http/dto/parameter"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Middleware é um alias para os middlewares de autorização injetados.
type Middleware = func(http.Handler) http.Handler

// Routes monta o roteador de parâmetros, com os mesmos gates do user-crud
// (feature Parameter em read/update). Não há create nem delete: os parâmetros
// nascem no onboarding da empresa e só podem ser editados.
//
// GET /parameter/enterprise responde um array sem envelope — formato que o
// front já consome (models/parameter.ts).
func Routes(svc *appparameter.Service, requireRead, requireUpdate Middleware) chi.Router {
	r := chi.NewRouter()

	r.With(requireRead).Get("/enterprise", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		list, err := svc.List(req.Context(), enterpriseID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, parameterdto.ToParameterResponses(list))
	})

	r.With(requireUpdate).Put("/", func(w http.ResponseWriter, req *http.Request) {
		actorID, enterpriseID, ok := httpmw.Actor(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}

		var in parameterdto.UpdateParameterRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		p, err := svc.Update(req.Context(), actorID, enterpriseID, in.ToInput())
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, parameterdto.ToParameterResponse(p))
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
