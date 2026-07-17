package enterprisehttp

import (
	"context"
	"encoding/json"
	"net/http"

	appaddress "github.com/ESG-Project/suassu-api/internal/app/address"
	appenterprise "github.com/ESG-Project/suassu-api/internal/app/enterprise"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainenterprise "github.com/ESG-Project/suassu-api/internal/domain/enterprise"
	"github.com/ESG-Project/suassu-api/internal/http/httperr"
	"github.com/ESG-Project/suassu-api/internal/http/response"
	"github.com/go-chi/chi/v5"
)

// Service interface defines the methods for the enterprise service
type Service interface {
	Create(ctx context.Context, in appenterprise.CreateInput) (enterpriseID string, userID string, err error)
	GetByID(ctx context.Context, id string) (*domainenterprise.Enterprise, error)
	Update(ctx context.Context, in appenterprise.UpdateInput) error
}

// Routes sets up the routes for the enterprise service
func Routes(svc Service) chi.Router {
	r := chi.NewRouter()

	// POST /enterprises - creates enterprise with products, parameters, roles, permissions and admin user
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		var reqBody struct {
			Enterprise struct {
				Name        string  `json:"name"`
				CNPJ        string  `json:"cnpj"`
				Email       string  `json:"email"`
				FantasyName *string `json:"fantasyName,omitempty"`
				Phone       *string `json:"phone,omitempty"`
				Address     *struct {
					State        string  `json:"state"`
					ZipCode      string  `json:"zipCode"`
					City         string  `json:"city"`
					Neighborhood string  `json:"neighborhood"`
					Street       string  `json:"street"`
					Num          string  `json:"num"`
					Latitude     *string `json:"latitude,omitempty"`
					Longitude    *string `json:"longitude,omitempty"`
					AddInfo      *string `json:"addInfo,omitempty"`
				} `json:"address,omitempty"`
			} `json:"enterprise"`
			User struct {
				Name     string  `json:"name"`
				Email    string  `json:"email"`
				Password string  `json:"password"`
				Document string  `json:"document"`
				Phone    *string `json:"phone,omitempty"`
			} `json:"user"`
		}

		if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
			httperr.Handle(w, req, apperr.New(apperr.CodeInvalid, "invalid body"))
			return
		}

		// Convert to CreateInput
		input := appenterprise.CreateInput{
			Name:        reqBody.Enterprise.Name,
			CNPJ:        reqBody.Enterprise.CNPJ,
			Email:       reqBody.Enterprise.Email,
			FantasyName: reqBody.Enterprise.FantasyName,
			Phone:       reqBody.Enterprise.Phone,
			User: appenterprise.UserInput{
				Name:     reqBody.User.Name,
				Email:    reqBody.User.Email,
				Password: reqBody.User.Password,
				Document: reqBody.User.Document,
				Phone:    reqBody.User.Phone,
			},
		}

		// Handle address if provided
		if reqBody.Enterprise.Address != nil {
			input.Address = &appaddress.CreateInput{
				State:        reqBody.Enterprise.Address.State,
				ZipCode:      reqBody.Enterprise.Address.ZipCode,
				City:         reqBody.Enterprise.Address.City,
				Neighborhood: reqBody.Enterprise.Address.Neighborhood,
				Street:       reqBody.Enterprise.Address.Street,
				Num:          reqBody.Enterprise.Address.Num,
				Latitude:     reqBody.Enterprise.Address.Latitude,
				Longitude:    reqBody.Enterprise.Address.Longitude,
				AddInfo:      reqBody.Enterprise.Address.AddInfo,
			}
		}

		enterpriseID, userID, err := svc.Create(req.Context(), input)
		if err != nil {
			httperr.Handle(w, req, err)
			return
		}

		response.JSON(w, http.StatusCreated, map[string]string{
			"enterpriseId": enterpriseID,
			"userId":       userID,
			"message":      "Enterprise created successfully with default configuration",
		}, nil)
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
