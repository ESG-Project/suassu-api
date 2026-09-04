package typeproducthttp

import (
	"encoding/json"
	"net/http"

	apptypeproduct "github.com/ESG-Project/suassu-api/internal/app/typeproduct"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	typeproductdto "github.com/ESG-Project/suassu-api/internal/http/dto/typeproduct"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Middleware é um alias para os middlewares de autorização injetados.
type Middleware = func(http.Handler) http.Handler

// Routes monta o roteador de tipos de produto, com os mesmos gates do
// user-crud (feature TypeProduct em read/create/update/delete).
//
// GET /typeProduct/enterprise responde um array sem envelope — formato que o
// front já consome (models/productType.ts).
func Routes(svc *apptypeproduct.Service, requireRead, requireCreate, requireUpdate, requireDelete Middleware) chi.Router {
	r := chi.NewRouter()

	r.With(requireRead).Get("/enterprise", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := httpmw.EnterpriseID(req.Context())
		list, err := svc.List(req.Context(), enterpriseID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, typeproductdto.ToTypeProductResponses(list))
	})

	r.With(requireCreate).Post("/", func(w http.ResponseWriter, req *http.Request) {
		actorID, enterpriseID, ok := httpmw.Actor(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}

		var in typeproductdto.CreateTypeProductRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		t, err := svc.Create(req.Context(), actorID, enterpriseID, in.Type)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusCreated, typeproductdto.ToTypeProductResponse(t))
	})

	r.With(requireUpdate).Put("/", func(w http.ResponseWriter, req *http.Request) {
		actorID, enterpriseID, ok := httpmw.Actor(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}

		var in typeproductdto.UpdateTypeProductRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		t, err := svc.Update(req.Context(), actorID, enterpriseID, in.ID, in.Type)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		writeJSON(w, http.StatusOK, typeproductdto.ToTypeProductResponse(t))
	})

	r.With(requireDelete).Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
		actorID, enterpriseID, ok := httpmw.Actor(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}

		if err := svc.Delete(req.Context(), actorID, enterpriseID, chi.URLParam(req, "id")); err != nil {
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
