package config

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenPostgres abre conexão com Postgres (driver pgx), aplica pool e faz ping com timeout.
func OpenPostgres(ctx context.Context, cfg *Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxIdleTime(time.Duration(cfg.DBConnMaxIdleMS) * time.Millisecond)
	db.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifeMS) * time.Millisecond)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
