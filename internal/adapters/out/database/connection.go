package database

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migration.sql
var migrationSQL string

func NewPool() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL não configurada")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao parsear DATABASE_URL: %w", err)
	}
	// Supabase transaction pooler (porta 6543) não suporta prepared statements
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	// Tuning do pool: o transaction pooler do Supabase já multiplexa conexões,
	// então mantemos um pool enxuto para não esgotar o limite do pooler.
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar pool de conexão: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("banco inacessível: %w", err)
	}

	if _, err := pool.Exec(context.Background(), migrationSQL); err != nil {
		return nil, fmt.Errorf("falha ao executar migração: %w", err)
	}

	return pool, nil
}
