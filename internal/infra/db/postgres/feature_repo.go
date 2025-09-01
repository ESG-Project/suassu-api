package postgres

import (
	"context"
	"database/sql"

	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type FeatureRepo struct{ q *sqlc.Queries }

// NewFeatureRepoFrom accepts any sqlc.DBTX (e.g. *sql.DB or *sql.Tx)
func NewFeatureRepoFrom(d dbtx) *FeatureRepo { return &FeatureRepo{q: sqlc.New(d)} }

// NewFeatureRepo for backward compatibility
func NewFeatureRepo(db *sql.DB) *FeatureRepo { return NewFeatureRepoFrom(db) }

// Upsert insere uma nova feature ou não faz nada se o nome já existir.
func (r *FeatureRepo) Upsert(ctx context.Context, name string) error {
	return r.q.UpsertFeature(ctx, name)
}
