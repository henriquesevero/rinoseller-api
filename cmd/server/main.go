package main

import (
	"log"

	"github.com/joho/godotenv"

	httphandler "rinoseller-api/internal/adapters/in/http"
	"rinoseller-api/internal/adapters/out/database"
	"rinoseller-api/internal/adapters/out/repository"
	"rinoseller-api/internal/core/services"
)

func main() {
	_ = godotenv.Load()

	db, err := database.NewPool()
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco: %v", err)
	}
	defer db.Close()
	log.Println("✓ Conectado ao Supabase PostgreSQL")

	userRepo    := repository.NewPostgresUserRepository(db)
	productRepo := repository.NewPostgresProductRepository(db)
	orderRepo   := repository.NewPostgresOrderRepository(db)
	clientRepo  := repository.NewPostgresClientRepository(db)
	quoteRepo   := repository.NewPostgresQuoteRepository(db)
	expenseRepo := repository.NewPostgresExpenseRepository(db)
	capitalRepo := repository.NewPostgresCapitalContributionRepository(db)

	authService    := services.NewAuthService(userRepo)
	userService    := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)
	orderService   := services.NewOrderService(orderRepo, productRepo, clientRepo)
	clientService  := services.NewClientService(clientRepo, orderRepo)
	quoteService   := services.NewQuoteService(quoteRepo, productRepo, clientRepo)
	expenseService := services.NewExpenseService(expenseRepo)
	capitalService := services.NewCapitalContributionService(capitalRepo)

	if err := userService.EnsureDefaultAdmin(); err != nil {
		log.Printf("⚠ Aviso ao criar admin padrão: %v", err)
	}

	handler := httphandler.NewHandler(authService, userService, productService, orderService, clientService, quoteService, expenseService, capitalService)
	router  := httphandler.SetupRouter(handler, authService)

	log.Println("✓ Servidor KoravHub iniciado em http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}
