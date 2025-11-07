package phytoanalysishttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	appphyto "github.com/ESG-Project/suassu-api/internal/app/phytoanalysis"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	phytodto "github.com/ESG-Project/suassu-api/internal/http/dto/phytoanalysis"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	"github.com/ESG-Project/suassu-api/internal/http/response"
	"github.com/go-chi/chi/v5"
)

// Service define a interface do serviço de PhytoAnalysis para a camada HTTP
type Service = appphyto.ServiceInterface

func Routes(svc Service) chi.Router {
	r := chi.NewRouter()

	// POST /phyto-analyses - Criar nova análise
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var in phytodto.CreatePhytoAnalysisRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		// Converter DTO para input do serviço
		specimens := make([]appphyto.SpecimenInput, 0, len(in.Specimens))
		for _, s := range in.Specimens {
			specimens = append(specimens, appphyto.SpecimenInput{
				Portion:             s.Portion,
				Height:              s.Height,
				Cap1:                s.Cap1,
				Cap2:                s.Cap2,
				Cap3:                s.Cap3,
				Cap4:                s.Cap4,
				Cap5:                s.Cap5,
				Cap6:                s.Cap6,
				AverageDap:          s.AverageDap,
				BasalArea:           s.BasalArea,
				Volume:              s.Volume,
				RegisterDate:        s.RegisterDate,
				SpecieID:            s.SpecieID,
				ScientificName:      s.ScientificName,
				Family:              s.Family,
				PopularName:         s.PopularName,
				LawScope:            s.LawScope,
				LawID:               s.LawID,
				IsLawActive:         s.IsLawActive,
				SpeciesFormFactor:   s.SpeciesFormFactor,
				IsSpeciesProtected:  s.IsSpeciesProtected,
				SpeciesThreatStatus: s.SpeciesThreatStatus,
				SpeciesOrigin:       s.SpeciesOrigin,
			})
		}

		createInput := appphyto.CreateInput{
			Title:           in.Title,
			InitialDate:     in.InitialDate,
			PortionQuantity: in.PortionQuantity,
			PortionArea:     in.PortionArea,
			TotalArea:       in.TotalArea,
			SampledArea:     in.SampledArea,
			Description:     in.Description,
			ProjectID:       in.ProjectID,
			Specimens:       specimens,
		}

		id, err := svc.Create(req.Context(), createInput)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		response.JSON(w, http.StatusCreated, map[string]string{"id": id}, nil)
	})

	// GET /phyto-analyses/project/:projectId - Listar análises por projeto
	r.Get("/project/{projectId}", func(w http.ResponseWriter, req *http.Request) {
		projectID := chi.URLParam(req, "projectId")

		list, err := svc.ListByProject(req.Context(), projectID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		phytoList := make([]*phytodto.PhytoAnalysisResponse, 0, len(list))
		for _, p := range list {
			phytoList = append(phytoList, phytodto.ToPhytoAnalysisResponse(p))
		}

		response.JSON(w, http.StatusOK, phytoList, nil)
	})

	// GET /phyto-analyses/enterprise/:enterpriseId - Listar análises por enterprise
	r.Get("/enterprise/{enterpriseId}", func(w http.ResponseWriter, req *http.Request) {
		enterpriseID := chi.URLParam(req, "enterpriseId")

		list, err := svc.ListByEnterprise(req.Context(), enterpriseID)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		phytoList := make([]*phytodto.PhytoAnalysisResponse, 0, len(list))
		for _, p := range list {
			phytoList = append(phytoList, phytodto.ToPhytoAnalysisResponse(p))
		}

		response.JSON(w, http.StatusOK, phytoList, nil)
	})

	// GET /phyto-analyses/:id - Buscar análise por ID
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		phyto, err := svc.GetWithSpecimens(req.Context(), id)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		out := phytodto.ToPhytoAnalysisCompleteResponse(phyto)
		response.JSON(w, http.StatusOK, out, nil)
	})

	// GET /phyto-analyses?limit=50&offset=0&projectId=xxx
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		projectID := req.URL.Query().Get("projectId")

		var phytoList []*phytodto.PhytoAnalysisResponse

		if projectID != "" {
			// Listar por projeto
			list, err := svc.ListByProject(req.Context(), projectID)
			if err != nil {
				httperr.Handle(w, req, err)
				return
			}

			for _, p := range list {
				phytoList = append(phytoList, phytodto.ToPhytoAnalysisResponse(p))
			}
		} else {
			// Listar todas com paginação
			limit := parseInt32(req.URL.Query().Get("limit"), 50)
			offset := parseInt32(req.URL.Query().Get("offset"), 0)

			if limit > 1000 {
				limit = 1000
			}

			list, err := svc.ListAll(req.Context(), limit, offset)
			if err != nil {
				httperr.Handle(w, req, err)
				return
			}

			for _, p := range list {
				phytoList = append(phytoList, phytodto.ToPhytoAnalysisResponse(p))
			}
		}

		response.JSON(w, http.StatusOK, phytoList, nil)
	})

	// PUT /phyto-analyses/:id - Atualizar análise
	r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		var in phytodto.UpdatePhytoAnalysisRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		updateInput := appphyto.UpdateInput{
			Title:           in.Title,
			InitialDate:     in.InitialDate,
			PortionQuantity: in.PortionQuantity,
			PortionArea:     in.PortionArea,
			TotalArea:       in.TotalArea,
			SampledArea:     in.SampledArea,
			Description:     in.Description,
		}

		if err := svc.Update(req.Context(), id, updateInput); err != nil {
			httperr.Handle(w, req, err)
			return
		}

		response.JSON(w, http.StatusOK, map[string]string{"message": "updated"}, nil)
	})

	// DELETE /phyto-analyses/:id - Deletar análise
	r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		if err := svc.Delete(req.Context(), id); err != nil {
			httperr.Handle(w, req, err)
			return
		}

		response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"}, nil)
	})

	// Rotas para specimens
	r.Route("/{phytoId}/specimens", func(r chi.Router) {
		// GET /phyto-analyses/:phytoId/specimens - Listar specimens
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			phytoID := chi.URLParam(req, "phytoId")

			// Buscar análise completa
			phyto, err := svc.GetWithSpecimens(req.Context(), phytoID)
			if err != nil {
				httperr.Handle(w, req, err)
				return
			}

			// Converter specimens
			specimens := make([]phytodto.SpecimenResponse, 0, len(phyto.Specimens))
			for _, s := range phyto.Specimens {
				specimens = append(specimens, phytodto.SpecimenResponse{
					ID:             s.ID,
					Portion:        s.Portion,
					Height:         s.Height,
					Cap1:           s.Cap1,
					Cap2:           s.Cap2,
					Cap3:           s.Cap3,
					Cap4:           s.Cap4,
					Cap5:           s.Cap5,
					Cap6:           s.Cap6,
					AverageDap:     s.AverageDap,
					BasalArea:      s.BasalArea,
					Volume:         s.Volume,
					RegisterDate:   s.RegisterDate,
					SpecieID:       s.SpecieID,
					ScientificName: s.ScientificName,
					Family:         s.Family,
					PopularName:    s.PopularName,
				})
			}

			response.JSON(w, http.StatusOK, specimens, nil)
		})
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
