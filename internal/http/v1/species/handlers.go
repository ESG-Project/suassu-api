package specieshttp

import (
	"net/http"
	"strconv"

	appspecies "github.com/ESG-Project/suassu-api/internal/app/species"
	speciesdto "github.com/ESG-Project/suassu-api/internal/http/dto/species"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	"github.com/ESG-Project/suassu-api/internal/http/response"
	"github.com/go-chi/chi/v5"
)

// Service define a interface do serviço de Species para a camada HTTP
type Service = appspecies.ServiceInterface

func Routes(svc Service) chi.Router {
	r := chi.NewRouter()

	// GET /species?limit=0&offset=0 - Listar espécies
	// Se limit não for especificado ou for 0, retorna todas as espécies
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		limitParam := req.URL.Query().Get("limit")
		offset := parseInt32(req.URL.Query().Get("offset"), 0)

		var limit int32
		if limitParam == "" || limitParam == "0" {
			// Sem limite - retorna todas as espécies
			limit = 999999
		} else {
			limit = parseInt32(limitParam, 50)
			if limit > 10000 {
				limit = 10000
			}
		}

		list, err := svc.List(req.Context(), limit, offset)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		speciesList := make([]*speciesdto.SpeciesResponse, 0, len(list))
		for _, s := range list {
			speciesList = append(speciesList, speciesdto.ToSpeciesResponse(s))
		}

		response.JSON(w, http.StatusOK, speciesList, nil)
	})

	// GET /species/{id} - Buscar espécie por ID
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		species, err := svc.GetByID(req.Context(), id)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		out := speciesdto.ToSpeciesResponse(species)
		response.JSON(w, http.StatusOK, out, nil)
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
