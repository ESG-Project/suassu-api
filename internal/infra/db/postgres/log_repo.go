package postgres

import (
	"context"
	"database/sql"

	domainauditlog "github.com/ESG-Project/suassu-api/internal/domain/auditlog"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type LogRepo struct{ q *sqlc.Queries }

func NewLogRepoFrom(d dbtx) *LogRepo { return &LogRepo{q: sqlc.New(d)} }
func NewLogRepo(db *sql.DB) *LogRepo { return NewLogRepoFrom(db) }

func (r *LogRepo) Create(ctx context.Context, l *domainauditlog.Log) error {
	return r.q.CreateLog(ctx, sqlc.CreateLogParams{
		ID:           l.ID,
		Tag:          utils.ToNullString(l.ActorID),
		EnterpriseId: l.EnterpriseID,
		Description:  l.Description,
		CreatedAt:    l.CreatedAt,
	})
}
