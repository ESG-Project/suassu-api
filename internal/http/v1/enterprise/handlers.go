package enterprisehttp

import (
	"context"
	"encoding/json"
	"net/http"

	appenterprise "github.com/ESG-Project/suassu-api/internal/app/enterprise"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainenterprise "github.com/ESG-Project/suassu-api/internal/domain/enterprise"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	"github.com/ESG-Project/suassu-api/internal/http/response"
	"github.com/go-chi/chi/v5"
)

// Service interface defines the methods for the enterprise service
type Service interface {
	Create(ctx context.Context, in appenterprise.CreateInput) (string, error)
	GetByID(ctx context.Context, id string) (*domainenterprise.Enterprise, error)
	Update(ctx context.Context, in appenterprise.UpdateInput) error
}

// Routes sets up the routes for the enterprise service
func Routes(svc Service) chi.Router {
	r := chi.NewRouter()

	// POST /enterprises
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var in appenterprise.CreateInput
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		id, err := svc.Create(req.Context(), in)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}
		response.JSON(w, http.StatusCreated, map[string]string{"id": id}, nil)
	})

	// GET /enterprises/{id}
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		if id == "" {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "enterprise id is required"))
			return
		}

		enterprise, err := svc.GetByID(req.Context(), id)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		type addressOut struct {
			ID           string  `json:"id"`
			State        string  `json:"state"`
			ZipCode      string  `json:"zipCode"`
			City         string  `json:"city"`
			Neighborhood string  `json:"neighborhood"`
			Street       string  `json:"street"`
			Num          string  `json:"num"`
			Latitude     *string `json:"latitude,omitempty"`
			Longitude    *string `json:"longitude,omitempty"`
			AddInfo      *string `json:"addInfo,omitempty"`
		}

		type enterpriseOut struct {
			ID          string      `json:"id"`
			CNPJ        string      `json:"cnpj"`
			Email       string      `json:"email"`
			Name        string      `json:"name"`
			FantasyName *string     `json:"fantasyName,omitempty"`
			Phone       *string     `json:"phone,omitempty"`
			AddressID   *string     `json:"addressId,omitempty"`
			Address     *addressOut `json:"address,omitempty"`
		}

		var aOut *addressOut
		if enterprise.Address != nil {
			aOut = &addressOut{
				ID:           enterprise.Address.ID,
				ZipCode:      enterprise.Address.ZipCode,
				State:        enterprise.Address.State,
				City:         enterprise.Address.City,
				Neighborhood: enterprise.Address.Neighborhood,
				Street:       enterprise.Address.Street,
				Num:          enterprise.Address.Num,
				Latitude:     enterprise.Address.Latitude,
				Longitude:    enterprise.Address.Longitude,
				AddInfo:      enterprise.Address.AddInfo,
			}
		}

		out := enterpriseOut{
			ID:          enterprise.ID,
			CNPJ:        enterprise.CNPJ,
			Email:       enterprise.Email,
			Name:        enterprise.Name,
			FantasyName: enterprise.FantasyName,
			Phone:       enterprise.Phone,
			AddressID:   enterprise.AddressID,
			Address:     aOut,
		}

		response.JSON(w, http.StatusOK, out, nil)
	})

	// PUT /enterprises
	r.Put("/", func(w http.ResponseWriter, req *http.Request) {
		// 1. Decodificar o corpo inteiro em um struct que contém o ID
		var in appenterprise.UpdateInput
		if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		// 2. Validar e acessar o ID a partir do struct
		// (Assumindo que UpdateInput agora tem um campo ID)
		if in.ID == "" {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "enterprise id is required in body"))
			return
		}

		// 3. Chamar o serviço
		// A assinatura do serviço Update precisaria ser ajustada
		err := svc.Update(req.Context(), in)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		response.JSON(w, http.StatusNoContent, domainenterprise.Enterprise{}, nil)
	})

	return r
}
