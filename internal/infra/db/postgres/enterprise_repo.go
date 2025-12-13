package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
	domainenterprise "github.com/ESG-Project/suassu-api/internal/domain/enterprise"
	domainparameter "github.com/ESG-Project/suassu-api/internal/domain/parameter"
	domainproduct "github.com/ESG-Project/suassu-api/internal/domain/product"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

// Structs auxiliares para unmarshaling do JSON retornado pelo banco
type productJSON struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	SuggestedValue *string `json:"suggestedValue"`
	EnterpriseID   string  `json:"enterpriseId"`
	ParameterID    *string `json:"parameterId"`
	Deliverable    bool    `json:"deliverable"`
	TypeProductID  *string `json:"typeProductId"`
	IsDefault      bool    `json:"isDefault"`
}

type parameterJSON struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Value        *string `json:"value"`
	EnterpriseID string  `json:"enterpriseId"`
	IsDefault    bool    `json:"isDefault"`
}

type EnterpriseRepo struct{ q *sqlc.Queries }

func NewEnterpriseRepoFrom(d dbtx) *EnterpriseRepo {
	return &EnterpriseRepo{q: sqlc.New(d)}
}

func NewEnterpriseRepo(db *sql.DB) *EnterpriseRepo {
	return &EnterpriseRepo{q: sqlc.New(db)}
}

func (r *EnterpriseRepo) Create(ctx context.Context, e *domainenterprise.Enterprise) error {
	return r.q.CreateEnterprise(ctx, sqlc.CreateEnterpriseParams{
		ID:          e.ID,
		Cnpj:        e.CNPJ,
		AddressId:   utils.ToNullString(e.AddressID),
		Email:       e.Email,
		FantasyName: utils.ToNullString(e.FantasyName),
		Name:        e.Name,
		Phone:       utils.ToNullString(e.Phone),
	})
}

func (r *EnterpriseRepo) GetByID(ctx context.Context, id string) (*domainenterprise.Enterprise, error) {
	row, err := r.q.GetEnterpriseByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "enterprise not found")
		}
		return nil, err
	}

	enterprise := &domainenterprise.Enterprise{
		ID:          row.ID,
		CNPJ:        row.Cnpj,
		Email:       row.Email,
		Name:        row.Name,
		FantasyName: utils.FromNullString(row.FantasyName),
		Phone:       utils.FromNullString(row.Phone),
		AddressID:   utils.FromNullString(row.AddressId),
	}

	// Processar endereço
	if row.AddressId.Valid {
		addr := &domainaddress.Address{
			ID:           row.AddressId.String,
			ZipCode:      row.ZipCode,
			State:        row.State,
			City:         row.City,
			Neighborhood: row.Neighborhood,
			Street:       row.Street,
			Num:          row.Num,
		}
		if row.Latitude.Valid {
			addr.SetLatitude(&row.Latitude.String)
		}
		if row.Longitude.Valid {
			addr.SetLongitude(&row.Longitude.String)
		}
		if row.AddInfo.Valid {
			addr.SetAddInfo(&row.AddInfo.String)
		}
		enterprise.Address = addr
	}

	// Processar products (JSON array)
	if row.Products != nil {
		if productsBytes, ok := row.Products.([]byte); ok && len(productsBytes) > 0 {
			var productsJSON []productJSON
			if err := json.Unmarshal(productsBytes, &productsJSON); err == nil {
				products := make([]domainproduct.Product, 0, len(productsJSON))
				for _, pj := range productsJSON {
					products = append(products, domainproduct.Product{
						ID:             pj.ID,
						Name:           pj.Name,
						SuggestedValue: pj.SuggestedValue,
						EnterpriseID:   pj.EnterpriseID,
						ParameterID:    pj.ParameterID,
						Deliverable:    pj.Deliverable,
						TypeProductID:  pj.TypeProductID,
						IsDefault:      pj.IsDefault,
					})
				}
				enterprise.Products = products
			}
		}
	}

	// Processar parameters (JSON array)
	if row.Parameters != nil {
		if parametersBytes, ok := row.Parameters.([]byte); ok && len(parametersBytes) > 0 {
			var parametersJSON []parameterJSON
			if err := json.Unmarshal(parametersBytes, &parametersJSON); err == nil {
				parameters := make([]domainparameter.Parameter, 0, len(parametersJSON))
				for _, pj := range parametersJSON {
					parameters = append(parameters, domainparameter.Parameter{
						ID:           pj.ID,
						Title:        pj.Title,
						Value:        pj.Value,
						EnterpriseID: pj.EnterpriseID,
						IsDefault:    pj.IsDefault,
					})
				}
				enterprise.Parameters = parameters
			}
		}
	}

	return enterprise, nil
}

func (r *EnterpriseRepo) Update(ctx context.Context, e *domainenterprise.Enterprise) error {
	return r.q.UpdateEnterprise(ctx, sqlc.UpdateEnterpriseParams{
		ID:          e.ID,
		Cnpj:        e.CNPJ,
		AddressId:   utils.ToNullString(e.AddressID),
		Email:       e.Email,
		FantasyName: utils.ToNullString(e.FantasyName),
		Name:        e.Name,
		Phone:       utils.ToNullString(e.Phone),
	})
}
