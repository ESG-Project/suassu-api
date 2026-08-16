package specieshttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	appspecies "github.com/ESG-Project/suassu-api/internal/app/species"
	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	speciesdto "github.com/ESG-Project/suassu-api/internal/http/dto/species"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	httpmw "github.com/ESG-Project/suassu-api/internal/http/middleware"
	"github.com/ESG-Project/suassu-api/internal/http/response"
	"github.com/go-chi/chi/v5"
)

// Service define a interface do serviço de Species para a camada HTTP
type Service = appspecies.ServiceInterface

// Middleware é um alias para os middlewares de autorização injetados.
type Middleware = func(http.Handler) http.Handler

// Routes monta o roteador de espécies.
//   - requireCreate: exige permissão Species.create (criar espécie)
//   - requireUpdate: exige permissão Species.update (solicitar alteração)
//   - requireManage: exige permissão ManageSpecies.update (aprovar/recusar/fila)
func Routes(svc Service, requireCreate, requireUpdate, requireManage Middleware) chi.Router {
	r := chi.NewRouter()

	// GET /species - Catálogo oficial (apenas versões aprovadas, global)
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		limit, offset := parsePaging(req)
		list, err := svc.List(req.Context(), limit, offset)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusOK, toResponses(list), nil)
	})

	// GET /species/manage - Espécies visíveis para a empresa do usuário
	// (aprovadas globais + próprias pendentes/recusadas, todas as versões)
	r.Get("/manage", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}
		limit, offset := parsePaging(req)
		list, err := svc.ListVisible(req.Context(), claims.EnterpriseID, limit, offset)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusOK, toResponses(list), nil)
	})

	// GET /species/pending - Fila de aprovação (super-admin)
	r.With(requireManage).Get("/pending", func(w http.ResponseWriter, req *http.Request) {
		limit, offset := parsePaging(req)
		list, err := svc.ListPending(req.Context(), limit, offset)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusOK, toResponses(list), nil)
	})

	// GET /species/{id} - Buscar espécie por ID
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		species, err := svc.GetByID(req.Context(), id)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusOK, speciesdto.ToSpeciesResponse(species), nil)
	})

	// POST /species - Criar nova espécie (nasce PENDING). Exige Species.create.
	r.With(requireCreate).Post("/", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}

		var in speciesdto.CreateSpeciesRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		userID := claims.Subject
		enterpriseID := claims.EnterpriseID
		input := appspecies.CreateInput{
			ScientificName: in.ScientificName,
			Family:         in.Family,
			PopularName:    in.PopularName,
			Habit:          in.Habit,
			CreatedBy:      &userID,
			EnterpriseID:   &enterpriseID,
		}
		applyLegislation(&input, in.Legislation)
		id, err := svc.Create(req.Context(), input)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusCreated, map[string]string{"id": id}, nil)
	})

	// PUT /species/{id} - Solicitar alteração (cria nova versão). Exige Species.update.
	r.With(requireUpdate).Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
		claims, ok := httpmw.ClaimsFromCtx(req.Context())
		if !ok {
			httperr.Handle(w, req, apperr.New(apperr.CodeUnauthorized, "authentication required"))
			return
		}

		id := chi.URLParam(req, "id")
		var in speciesdto.ChangeSpeciesRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		userID := claims.Subject
		enterpriseID := claims.EnterpriseID
		change := appspecies.ChangeInput{
			BaseSpeciesID:  id,
			ScientificName: in.ScientificName,
			Family:         in.Family,
			PopularName:    in.PopularName,
			Habit:          in.Habit,
			CreatedBy:      &userID,
			EnterpriseID:   &enterpriseID,
		}
		applyLegislationChange(&change, in.Legislation)
		newID, err := svc.RequestChange(req.Context(), change)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusCreated, map[string]string{"id": newID}, nil)
	})

	// PATCH /species/{id}/approve - Aprovar espécie pendente (super-admin)
	r.With(requireManage).Patch("/{id}/approve", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		if err := svc.Approve(req.Context(), id); err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "APPROVED"}, nil)
	})

	// PATCH /species/{id}/refuse - Recusar espécie pendente (super-admin)
	r.With(requireManage).Patch("/{id}/refuse", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		if err := svc.Refuse(req.Context(), id); err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "REFUSED"}, nil)
	})

	return r
}

func toResponses(list []*types.SpeciesWithLegislation) []*speciesdto.SpeciesResponse {
	out := make([]*speciesdto.SpeciesResponse, 0, len(list))
	for _, s := range list {
		out = append(out, speciesdto.ToSpeciesResponse(s))
	}
	return out
}

// applyLegislation copia os campos de legislação do DTO para o CreateInput.
// Só o LawScope != "" faz a legislação ser efetivamente criada (ver service).
func applyLegislation(in *appspecies.CreateInput, l *speciesdto.LegislationRequest) {
	if l == nil {
		return
	}
	in.LawScope = l.LawScope
	in.LawID = l.LawID
	in.IsLawActive = l.IsLawActive
	in.SpeciesFormFactor = l.SpeciesFormFactor
	in.IsSpeciesProtected = l.IsSpeciesProtected
	in.SpeciesThreatStatus = l.SpeciesThreatStatus
	in.SpeciesOrigin = l.SpeciesOrigin
	in.SuccessionalEcology = l.SuccessionalEcology
}

// applyLegislationChange copia os campos de legislação do DTO para o ChangeInput.
func applyLegislationChange(in *appspecies.ChangeInput, l *speciesdto.LegislationRequest) {
	if l == nil {
		return
	}
	in.LawScope = l.LawScope
	in.LawID = l.LawID
	in.IsLawActive = l.IsLawActive
	in.SpeciesFormFactor = l.SpeciesFormFactor
	in.IsSpeciesProtected = l.IsSpeciesProtected
	in.SpeciesThreatStatus = l.SpeciesThreatStatus
	in.SpeciesOrigin = l.SpeciesOrigin
	in.SuccessionalEcology = l.SuccessionalEcology
}

func parsePaging(req *http.Request) (limit, offset int32) {
	offset = parseInt32(req.URL.Query().Get("offset"), 0)
	limitParam := req.URL.Query().Get("limit")
	if limitParam == "" || limitParam == "0" {
		return 999999, offset
	}
	limit = parseInt32(limitParam, 50)
	if limit > 10000 {
		limit = 10000
	}
	return limit, offset
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
