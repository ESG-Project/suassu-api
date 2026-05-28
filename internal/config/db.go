package config

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenPostgres abre conexão com Postgres (driver pgx), aplica pool e faz ping com timeout.
func OpenPostgres(ctx context.Context, cfg *Config) (*sql.DB, error) {
	// Garantir que o search_path inclui 'public' para encontrar todas as tabelas
	dsn := cfg.DBDSN
	if !strings.Contains(dsn, "search_path") {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		dsn = dsn + separator + "search_path=public"
	}

	db, err := sql.Open("pgx", dsn)
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

	// Garantir que o search_path está configurado para 'public'
	setCtx, setCancel := context.WithTimeout(ctx, 5*time.Second)
	defer setCancel()
	if _, err := db.ExecContext(setCtx, "SET search_path = public"); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
