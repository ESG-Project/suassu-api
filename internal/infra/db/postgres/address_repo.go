package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type AddressRepo struct{ q *sqlc.Queries }

func NewAddressRepoFrom(d dbtx) *AddressRepo {
	return &AddressRepo{q: sqlc.New(d)}
}

func NewAddressRepo(db *sql.DB) *AddressRepo {
	return &AddressRepo{q: sqlc.New(db)}
}

func (r *AddressRepo) Create(ctx context.Context, a *domainaddress.Address) error {
	return r.q.CreateAddress(ctx, sqlc.CreateAddressParams{
		ID:           a.ID,
		ZipCode:      a.ZipCode,
		State:        a.State,
		City:         a.City,
		Neighborhood: a.Neighborhood,
		Street:       a.Street,
		Num:          a.Num,
		Latitude:     utils.ToNullString(a.Latitude),
		Longitude:    utils.ToNullString(a.Longitude),
		AddInfo:      utils.ToNullString(a.AddInfo),
	})
}

func (r *AddressRepo) FindByDetails(ctx context.Context, params *domainaddress.SearchParams) (*domainaddress.Address, error) {
	row, err := r.q.FindAddressByDetails(ctx, sqlc.FindAddressByDetailsParams{
		ZipCode:      params.ZipCode,
		State:        params.State,
		City:         params.City,
		Neighborhood: params.Neighborhood,
		Street:       params.Street,
		Num:          params.Num,
		Latitude:     utils.ToNullString(params.Latitude),
		Longitude:    utils.ToNullString(params.Longitude),
		AddInfo:      utils.ToNullString(params.AddInfo),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &domainaddress.Address{
		ID:           row.ID,
		ZipCode:      row.ZipCode,
		State:        row.State,
		City:         row.City,
		Neighborhood: row.Neighborhood,
		Street:       row.Street,
		Num:          row.Num,
		Latitude:     utils.FromNullString(row.Latitude),
		Longitude:    utils.FromNullString(row.Longitude),
		AddInfo:      utils.FromNullString(row.AddInfo),
	}, nil
}

// GetByID retorna um endereço como UserAddress para uso na camada de aplicação
func (r *AddressRepo) GetByID(ctx context.Context, addressID string) (*types.UserAddress, error) {
	row, err := r.q.GetAddressByID(ctx, addressID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "address not found")
		}
		return nil, err
	}

	return &types.UserAddress{
		ID:           row.ID,
		ZipCode:      row.ZipCode,
		State:        row.State,
		City:         row.City,
		Neighborhood: row.Neighborhood,
		Street:       row.Street,
		Num:          row.Num,
		Latitude:     utils.FromNullString(row.Latitude),
		Longitude:    utils.FromNullString(row.Longitude),
		AddInfo:      utils.FromNullString(row.AddInfo),
	}, nil
}
