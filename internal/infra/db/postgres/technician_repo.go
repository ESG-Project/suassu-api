package postgres

import (
	"context"
	"database/sql"

	domaintechnician "github.com/ESG-Project/suassu-api/internal/domain/technician"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type TechnicianRepo struct{ q *sqlc.Queries }

func NewTechnicianRepoFrom(d dbtx) *TechnicianRepo { return &TechnicianRepo{q: sqlc.New(d)} }
func NewTechnicianRepo(db *sql.DB) *TechnicianRepo { return NewTechnicianRepoFrom(db) }

func (r *TechnicianRepo) Create(ctx context.Context, t *domaintechnician.Technician) error {
	_, err := r.q.CreateTechnician(ctx, sqlc.CreateTechnicianParams{
		ID:          t.ID,
		ProRegister: utils.ToNullString(t.ProRegister),
		Graduation:  utils.ToNullString(t.Graduation),
		Ctf:         utils.ToNullString(t.CTF),
		UserId:      t.UserID,
	})
	return err
}

// Upsert cria o registro de Technician se ainda não existir, ou atualiza os
// campos se já existir (usado no update administrativo de usuário, quando o
// papel do usuário muda para "Técnico" e ele ainda não tinha um).
func (r *TechnicianRepo) Upsert(ctx context.Context, t *domaintechnician.Technician) error {
	_, err := r.q.UpsertTechnicianByUserID(ctx, sqlc.UpsertTechnicianByUserIDParams{
		ID:          t.ID,
		ProRegister: utils.ToNullString(t.ProRegister),
		Graduation:  utils.ToNullString(t.Graduation),
		Ctf:         utils.ToNullString(t.CTF),
		UserId:      t.UserID,
	})
	return err
}

func (r *TechnicianRepo) GetByUserID(ctx context.Context, userID string) (*domaintechnician.Technician, error) {
	row, err := r.q.GetTechnicianByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &domaintechnician.Technician{
		ID:          row.ID,
		ProRegister: utils.FromNullString(row.ProRegister),
		Graduation:  utils.FromNullString(row.Graduation),
		CTF:         utils.FromNullString(row.Ctf),
		UserID:      row.UserId,
	}, nil
}
