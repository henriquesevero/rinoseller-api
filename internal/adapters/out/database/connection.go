package database

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migration.sql
var migrationSQL string

func NewPool() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL não configurada")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
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
