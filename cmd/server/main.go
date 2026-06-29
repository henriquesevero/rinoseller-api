// @title           RinoSeller API
// @version         1.0
// @description     API do sistema RinoSeller — gestão de vendas, clientes, produtos, orçamentos e finanças.
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Informe: Bearer {token}
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	_ "rinoseller-api/docs"
	httphandler "rinoseller-api/internal/adapters/in/http"
	"rinoseller-api/internal/adapters/out/database"
	"rinoseller-api/internal/adapters/out/email"
	"rinoseller-api/internal/adapters/out/payment"
	"rinoseller-api/internal/adapters/out/repository"
	"rinoseller-api/internal/core/services"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco: %v", err)
	}
	defer db.Close()
	log.Println("✓ Conectado ao banco de dados")

	userRepo := repository.NewPostgresUserRepository(db)
	productRepo := repository.NewPostgresProductRepository(db)
	orderRepo := repository.NewPostgresOrderRepository(db)
	clientRepo := repository.NewPostgresClientRepository(db)
	paymentRepo := repository.NewPostgresClientPaymentRepository(db)
	quoteRepo := repository.NewPostgresQuoteRepository(db)
	expenseRepo := repository.NewPostgresExpenseRepository(db)
	capitalRepo := repository.NewPostgresCapitalContributionRepository(db)
	brandCatalogRepo := repository.NewPostgresBrandCatalogRepository(db)

	resendAPIKey := os.Getenv("RESEND_API_KEY")
	resendFromEmail := os.Getenv("RESEND_FROM_EMAIL")
	frontendURL := os.Getenv("FRONTEND_URL")
	if resendAPIKey == "" || resendFromEmail == "" || frontendURL == "" {
		log.Fatal("RESEND_API_KEY, RESEND_FROM_EMAIL e FRONTEND_URL são obrigatórias")
	}
	emailSender := email.NewResendSender(resendAPIKey, resendFromEmail)
	paymentGateway := payment.NewMockAsaasGateway()
	registrationCode := os.Getenv("REGISTRATION_ACCESS_CODE")
	if registrationCode == "" {
		log.Println("⚠ REGISTRATION_ACCESS_CODE não configurada — cadastro público está aberto sem trava")
	}

	authService, err := services.NewAuthService(userRepo, emailSender, frontendURL, registrationCode)
	if err != nil {
		log.Fatalf("Falha ao iniciar serviço de autenticação: %v", err)
	}
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo, clientRepo)
	clientService := services.NewClientService(clientRepo, orderRepo, paymentRepo, quoteRepo)
	quoteService := services.NewQuoteService(quoteRepo, productRepo, clientRepo, userRepo, emailSender)
	expenseService := services.NewExpenseService(expenseRepo)
	capitalService := services.NewCapitalContributionService(capitalRepo)
	brandCatalogService := services.NewBrandCatalogService(brandCatalogRepo)
	subscriptionService := services.NewSubscriptionService(userRepo, paymentGateway)
	documentEmailService := services.NewDocumentEmailService(clientRepo, emailSender)

	if password, err := userService.EnsureDefaultAdmin(ctx); err != nil {
		log.Printf("⚠ Aviso ao criar admin padrão: %v", err)
	} else if password != "" {
		log.Printf("✓ Admin padrão criado: admin@rinoseller.com / senha temporária: %s (troque-a após o primeiro login)", password)
	}

	handler := httphandler.NewHandler(authService, userService, productService, orderService, clientService, quoteService, expenseService, capitalService, brandCatalogService, subscriptionService, documentEmailService)
	router := httphandler.SetupRouter(handler, authService)

	srv := newServer(router)
	runWithGracefulShutdown(srv)
}

func newServer(handler http.Handler) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
}

func runWithGracefulShutdown(srv *http.Server) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("✓ Servidor RinoSeller iniciado em %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Falha ao iniciar o servidor: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Encerrando servidor...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Erro ao encerrar servidor: %v", err)
	}
	log.Println("✓ Servidor encerrado")
}
