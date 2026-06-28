package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"

	"rinoseller-api/internal/adapters/out/database"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("falha ao conectar ao banco: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, database.MigrationSQL); err != nil {
		log.Fatalf("falha ao executar migração: %v", err)
	}

	log.Println("✓ Migração executada com sucesso")
}
