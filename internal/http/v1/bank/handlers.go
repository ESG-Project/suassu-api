package bankhttp

import (
	"encoding/json"
	"net/http"

	appbank "github.com/ESG-Project/suassu-api/internal/app/bank"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	bankdto "github.com/ESG-Project/suassu-api/internal/http/dto/bank"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Middleware é um alias para os middlewares de autorização injetados.
type Middleware = func(http.Handler) http.Handler

// RegisterRoutes registra as rotas de banco diretamente no router recebido —
// e não como um sub-router montado — porque GET /all-banks não fica sob o
// prefixo /bank e precisa conviver com ele.
//
// Gates de permissão idênticos aos do user-crud: /all-banks usa a feature
// Bank; as demais usam EnterpriseBank.
//
//   - GET    /all-banks        Bank.read
//   - GET    /bank/enterprise  EnterpriseBank.read
//   - POST   /bank             EnterpriseBank.create
//   - DELETE /bank/{id}        EnterpriseBank.delete
//
// Os erros saem no formato legado ({"error": "<mensagem>"}) porque
// models/bank.ts lê `data.error` como string. Ver httperr.HandleLegacy.
func RegisterRoutes(r chi.Router, svc *appbank.Service, requireCatalogRead, requireRead, requireCreate, requireDelete Middleware) {
	r.With(requireCatalogRead).Get("/all-banks", func(w http.ResponseWriter, req *http.Request) {
		catalog, err := svc.ListCatalog(req.Context())
		if err != nil {
			httperr.HandleLegacy(w, req, err)
			return
		}
		writeRaw(w, http.StatusOK, catalog)
	})

	r.Route("/bank", func(b chi.Router) {
		b.With(requireRead).Get("/enterprise", func(w http.ResponseWriter, req *http.Request) {
			enterpriseID := httpmw.EnterpriseID(req.Context())
			rows, err := svc.ListByEnterprise(req.Context(), enterpriseID)
			if err != nil {
				httperr.HandleLegacy(w, req, err)
				return
			}
			writeJSON(w, http.StatusOK, bankdto.ToEnterpriseBankListItems(rows))
		})

		b.With(requireCreate).Post("/", func(w http.ResponseWriter, req *http.Request) {
			actorID, enterpriseID, ok := httpmw.Actor(req.Context())
			if !ok {
				httperr.HandleLegacy(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
				return
			}

			var in bankdto.CreateBankRequest
			if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
				httperr.HandleLegacy(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
				return
			}

			eb, err := svc.Link(req.Context(), actorID, enterpriseID, in.Bank.Code, in.Bank.Name)
			if err != nil {
				httperr.HandleLegacy(w, req, err)
				return
			}
			writeJSON(w, http.StatusCreated, bankdto.ToEnterpriseBankResponse(eb))
		})

		b.With(requireDelete).Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
			actorID, enterpriseID, ok := httpmw.Actor(req.Context())
			if !ok {
				httperr.HandleLegacy(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
				return
			}

			if err := svc.Unlink(req.Context(), actorID, chi.URLParam(req, "id"), enterpriseID); err != nil {
				httperr.HandleLegacy(w, req, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeRaw(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
