package specimenhttp

import (
	"encoding/json"
	"net/http"

	appspecimen "github.com/ESG-Project/suassu-api/internal/app/specimen"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	specimendto "github.com/ESG-Project/suassu-api/internal/http/dto/specimen"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	"github.com/ESG-Project/suassu-api/internal/http/response"
	"github.com/go-chi/chi/v5"
)

// Service define a interface do serviço de Specimen para a camada HTTP
type Service = appspecimen.ServiceInterface

func Routes(svc Service) chi.Router {
	r := chi.NewRouter()

	// POST /specimens - Criar novo specimen
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var in specimendto.CreateSpecimenRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		createInput := appspecimen.CreateInput{
			Portion:         in.Portion,
			Height:          in.Height,
			Cap1:            in.Cap1,
			Cap2:            in.Cap2,
			Cap3:            in.Cap3,
			Cap4:            in.Cap4,
			Cap5:            in.Cap5,
			Cap6:            in.Cap6,
			RegisterDate:    in.RegisterDate,
			PhytoAnalysisID: in.PhytoAnalysisID,
			SpecieID:        in.SpecieID,
		}

		id, err := svc.Create(req.Context(), createInput)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		response.JSON(w, http.StatusCreated, map[string]string{"id": id}, nil)
	})

	// GET /specimens/{id} - Buscar specimen por ID
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		specimen, err := svc.GetByID(req.Context(), id)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		out := specimendto.ToSpecimenResponse(specimen)
		response.JSON(w, http.StatusOK, out, nil)
	})

	// PUT /specimens/{id} - Atualizar specimen
	r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		var in specimendto.UpdateSpecimenRequest
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		updateInput := appspecimen.UpdateInput{
			Portion:      in.Portion,
			Height:       in.Height,
			Cap1:         in.Cap1,
			Cap2:         in.Cap2,
			Cap3:         in.Cap3,
			Cap4:         in.Cap4,
			Cap5:         in.Cap5,
			Cap6:         in.Cap6,
			RegisterDate: in.RegisterDate,
			SpecieID:     in.SpecieID,
		}

		if err := svc.Update(req.Context(), id, updateInput); err != nil {
			httperr.Handle(w, req, err)
			return
		}

		response.JSON(w, http.StatusOK, map[string]string{"message": "updated"}, nil)
	})

	// DELETE /specimens/{id} - Deletar specimen
	r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		if err := svc.Delete(req.Context(), id); err != nil {
			httperr.Handle(w, req, err)
			return
		}

		response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"}, nil)
	})

	return r
}

