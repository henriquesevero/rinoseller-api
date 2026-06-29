package httphandler

import (
	"net/http"
	"os"
	"rinoseller-api/internal/core/ports"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func allowedOrigins() []string {
	origins := []string{"http://localhost:5173", "http://localhost:3000"}
	if extra := os.Getenv("ALLOWED_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	}
	return origins
}

func SetupRouter(h *Handler, authUC ports.AuthUseCase) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/c/:id", h.ShareBrandCatalog)

	api := r.Group("/api")

	api.POST("/auth/register", h.Register)
	api.POST("/auth/verify-email", h.VerifyEmail)
	api.POST("/auth/resend-verification", h.ResendVerificationEmail)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/forgot-password", h.ForgotPassword)
	api.POST("/auth/reset-password", h.ResetPassword)
	api.POST("/subscriptions/checkout", h.CheckoutSubscription)

	auth := api.Group("")
	auth.Use(AuthMiddleware(authUC))
	{
		// Fora do gate de assinatura: precisa funcionar mesmo com o teste expirado,
		// para a tela de cobrança conseguir mostrar o status e processar o pagamento.
		auth.GET("/auth/me", h.GetMe)
		auth.DELETE("/auth/me", h.DeleteMe)

		gated := auth.Group("")
		gated.Use(SubscriptionRequired(authUC))
		{
			admin := gated.Group("")
			admin.Use(AdminOnly())
			{
				admin.GET("/users", h.ListUsers)
				admin.POST("/users", h.CreateUser)
				admin.PATCH("/users/:id", h.UpdateUser)
			}

			gated.GET("/products", h.ListProducts)
			gated.POST("/products", h.CreateProduct)
			gated.PATCH("/products/:id/price", h.UpdatePrice)
			gated.PATCH("/products/:id/cost-price", h.UpdateCostPrice)
			gated.PATCH("/products/:id/stock", h.UpdateStock)
			gated.DELETE("/products/:id", h.DeleteProduct)

			gated.GET("/orders", h.ListOrders)
			gated.POST("/orders", h.CreateOrder)
			gated.DELETE("/orders/:id", h.DeleteOrder)

			gated.GET("/clients", h.ListClients)
			gated.POST("/clients", h.CreateClient)
			gated.GET("/clients/:id", h.GetClient)
			gated.PATCH("/clients/:id", h.UpdateClient)
			gated.DELETE("/clients/:id", h.DeleteClient)
			gated.GET("/clients/:id/orders", h.GetClientOrders)
			gated.DELETE("/clients/:id/orders", h.ClearClientOrders)
			gated.GET("/clients/:id/quotes", h.GetClientQuotes)
			gated.DELETE("/clients/:id/quotes", h.ClearClientQuotes)
			gated.POST("/clients/:id/send-document-email", h.SendDocumentEmail)
			gated.GET("/payments", h.GetAllPayments)
			gated.GET("/clients/:id/payment-history", h.GetClientPayments)
			gated.DELETE("/clients/:id/payment-history", h.ClearClientPayments)
			gated.POST("/clients/:id/payment", h.RegisterPayment)
			gated.POST("/clients/:id/debt", h.AddDebt)

			gated.GET("/quotes", h.ListQuotes)
			gated.POST("/quotes", h.CreateQuote)
			gated.GET("/quotes/:id", h.GetQuote)
			gated.POST("/quotes/:id/approve", h.ApproveQuote)
			gated.POST("/quotes/:id/deliver", h.DeliverQuote)
			gated.POST("/quotes/:id/invoice", h.InvoiceQuote)
			gated.POST("/quotes/:id/cancel", h.CancelQuote)
			gated.POST("/quotes/:id/send-email", h.SendQuoteEmail)
			gated.DELETE("/quotes/:id", h.DeleteQuote)

			gated.GET("/expenses", h.ListExpenses)
			gated.POST("/expenses", h.CreateExpense)
			gated.POST("/expenses/:id/pay", h.PayExpense)
			gated.DELETE("/expenses/:id", h.DeleteExpense)

			gated.GET("/capital-contributions", h.ListCapitalContributions)
			gated.POST("/capital-contributions", h.CreateCapitalContribution)
			gated.DELETE("/capital-contributions/:id", h.DeleteCapitalContribution)

			gated.GET("/brand-catalogs", h.ListBrandCatalogs)
			gated.POST("/brand-catalogs", h.CreateBrandCatalog)
			gated.DELETE("/brand-catalogs/:id", h.DeleteBrandCatalog)
		}
	}

	return r
}
