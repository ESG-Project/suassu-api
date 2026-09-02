package postgres

import (
	"context"
	"database/sql"

	domainclient "github.com/ESG-Project/suassu-api/internal/domain/client"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type ClientRepo struct{ q *sqlc.Queries }

func NewClientRepoFrom(d dbtx) *ClientRepo { return &ClientRepo{q: sqlc.New(d)} }
func NewClientRepo(db *sql.DB) *ClientRepo { return NewClientRepoFrom(db) }

func (r *ClientRepo) Create(ctx context.Context, c *domainclient.Client) error {
	_, err := r.q.CreateClient(ctx, sqlc.CreateClientParams{
		ID:          c.ID,
		FantasyName: utils.ToNullString(c.FantasyName),
		UserId:      c.UserID,
	})
	return err
}

// Upsert cria o registro de Client se ainda não existir, ou atualiza o
// fantasyName se já existir (usado no update administrativo de usuário,
// quando o papel do usuário muda para "Cliente" e ele ainda não tinha um).
func (r *ClientRepo) Upsert(ctx context.Context, c *domainclient.Client) error {
	_, err := r.q.UpsertClientByUserID(ctx, sqlc.UpsertClientByUserIDParams{
		ID:          c.ID,
		FantasyName: utils.ToNullString(c.FantasyName),
		UserId:      c.UserID,
	})
	return err
}

func (r *ClientRepo) GetByUserID(ctx context.Context, userID string) (*domainclient.Client, error) {
	row, err := r.q.GetClientByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &domainclient.Client{
		ID:          row.ID,
		FantasyName: utils.FromNullString(row.FantasyName),
		UserID:      row.UserId,
	}, nil
}

// ClientListRow é uma linha de ListClientsByEnterprise (Client + User + Address + Role).
type ClientListRow struct {
	ClientID     string
	FantasyName  *string
	UserID       string
	Name         string
	Email        string
	Phone        *string
	Document     string
	EnterpriseID string
	ZipCode      *string
	State        *string
	City         *string
	Neighborhood *string
	Street       *string
	Num          *string
	RoleID       *string
	RoleTitle    *string
}

func (r *ClientRepo) ListByEnterprise(ctx context.Context, enterpriseID string) ([]ClientListRow, error) {
	rows, err := r.q.ListClientsByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}
	out := make([]ClientListRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, ClientListRow{
			ClientID:     row.ClientID,
			FantasyName:  utils.FromNullString(row.FantasyName),
			UserID:       row.ID,
			Name:         row.Name,
			Email:        row.Email,
			Phone:        utils.FromNullString(row.Phone),
			Document:     row.Document,
			EnterpriseID: row.EnterpriseID,
			ZipCode:      utils.FromNullString(row.ZipCode),
			State:        utils.FromNullString(row.State),
			City:         utils.FromNullString(row.City),
			Neighborhood: utils.FromNullString(row.Neighborhood),
			Street:       utils.FromNullString(row.Street),
			Num:          utils.FromNullString(row.Num),
			RoleID:       utils.FromNullString(row.RoleID),
			RoleTitle:    utils.FromNullString(row.RoleTitle),
		})
	}
	return out, nil
}
