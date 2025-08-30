package postgres

import (
	"context"
	"database/sql"

	"github.com/ESG-Project/suassu-api/internal/app/address"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainaddress "github.com/ESG-Project/suassu-api/internal/domain/address"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type AddressRepo struct{ q *sqlc.Queries }

// Construtor que aceita qualquer sqlc.DBTX (ex.: *sql.DB ou *sql.Tx)
func NewAddressRepoFrom(d dbtx) *AddressRepo { return &AddressRepo{q: sqlc.New(d)} }

// Compatibilidade com construtor anterior
func NewAddressRepo(db *sql.DB) *AddressRepo { return NewAddressRepoFrom(db) }

func (r *AddressRepo) FindByDetails(ctx context.Context, address *address.CreateInput) (*domainaddress.Address, error) {
	row, err := r.q.FindAddressByDetails(ctx, sqlc.FindAddressByDetailsParams{
		ZipCode:      address.ZipCode,
		State:        address.State,
		City:         address.City,
		Neighborhood: address.Neighborhood,
		Street:       address.Street,
		Num:          address.Num,
		Latitude:     utils.ToNullString(address.Latitude),
		Longitude:    utils.ToNullString(address.Longitude),
		AddInfo:      utils.ToNullString(address.AddInfo),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "address not found")
		}
		return nil, err
	}

	out := domainaddress.NewAddress(
		row.ID,
		row.ZipCode,
		row.State,
		row.City,
		row.Neighborhood,
		row.Street,
		row.Num,
	)

	if row.Latitude.Valid {
		out.SetLatitude(&row.Latitude.String)
	}
	if row.Longitude.Valid {
		out.SetLongitude(&row.Longitude.String)
	}
	if row.AddInfo.Valid {
		out.SetAddInfo(&row.AddInfo.String)
	}

	return out, nil
}

func (r *AddressRepo) Create(ctx context.Context, u *domainaddress.Address) error {
	return r.q.CreateAddress(ctx, sqlc.CreateAddressParams{
		ID:           u.ID,
		ZipCode:      u.ZipCode,
		State:        u.State,
		City:         u.City,
		Neighborhood: u.Neighborhood,
		Street:       u.Street,
		Num:          u.Num,
		Latitude:     utils.ToNullString(u.Latitude),
		Longitude:    utils.ToNullString(u.Longitude),
		AddInfo:      utils.ToNullString(u.AddInfo),
	})
}
