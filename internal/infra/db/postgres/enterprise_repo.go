package postgres

import (
	"context"
	"database/sql"

	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
	domainenterprise "github.com/ESG-Project/suassu-api/internal/domain/enterprise"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type EnterpriseRepo struct{ q *sqlc.Queries }

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
		return nil, err
	}

	enterprise := &domainenterprise.Enterprise{
		ID:          row.ID,
		CNPJ:        row.Cnpj,
		Email:       row.Email,
		Name:        row.Name,
		FantasyName: &row.FantasyName.String,
		Phone:       &row.Phone.String,
		AddressID:   &row.AddressId.String,
	}

	if row.AddressId.Valid {
		enterprise.Address = &domainaddress.Address{
			ID:           row.AddressId.String,
			ZipCode:      row.ZipCode,
			State:        row.State,
			City:         row.City,
			Neighborhood: row.Neighborhood,
			Street:       row.Street,
			Num:          row.Num,
			Latitude:     &row.Latitude.String,
			Longitude:    &row.Longitude.String,
			AddInfo:      &row.AddInfo.String,
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
