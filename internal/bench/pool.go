package bench

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func DSN() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://bench:bench@localhost:5432/bench?sslmode=disable"
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func NewStdlibPool(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("pgx", DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(envInt("MAX_CONNS", 16))
	db.SetMaxIdleConns(envInt("MIN_CONNS", 4))
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func NewPGXPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(DSN())
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = int32(envInt("MAX_CONNS", 16))
	cfg.MinConns = int32(envInt("MIN_CONNS", 4))
	return pgxpool.NewWithConfig(ctx, cfg)
}
