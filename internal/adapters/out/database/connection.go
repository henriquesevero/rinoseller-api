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
var MigrationSQL string

const (
	maxPoolConnsBehindTransactionPooler = 10
	minPoolConnsBehindTransactionPooler = 2
)

func disablePreparedStatementsForTransactionPoolerCompatibility(config *pgxpool.Config) {
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
}

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL não configurada")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao parsear DATABASE_URL: %w", err)
	}
	disablePreparedStatementsForTransactionPoolerCompatibility(config)

	config.MaxConns = maxPoolConnsBehindTransactionPooler
	config.MinConns = minPoolConnsBehindTransactionPooler
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar pool de conexão: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("banco inacessível: %w", err)
	}

	return pool, nil
}
