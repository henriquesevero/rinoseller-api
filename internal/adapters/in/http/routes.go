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

	api := r.Group("/api")

	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/forgot-password", h.ForgotPassword)
	api.POST("/auth/reset-password", h.ResetPassword)

	auth := api.Group("")
	auth.Use(AuthMiddleware(authUC))
	{
		auth.GET("/auth/me", h.GetMe)

		admin := auth.Group("")
		admin.Use(AdminOnly())
		{
			admin.GET("/users", h.ListUsers)
			admin.POST("/users", h.CreateUser)
			admin.PATCH("/users/:id", h.UpdateUser)
		}

		auth.GET("/products", h.ListProducts)
		auth.POST("/products", h.CreateProduct)
		auth.PATCH("/products/:id/price", h.UpdatePrice)
		auth.PATCH("/products/:id/stock", h.UpdateStock)
		auth.DELETE("/products/:id", h.DeleteProduct)

		auth.GET("/orders", h.ListOrders)
		auth.POST("/orders", h.CreateOrder)
		auth.DELETE("/orders/:id", h.DeleteOrder)

		auth.GET("/clients", h.ListClients)
		auth.POST("/clients", h.CreateClient)
		auth.GET("/clients/:id", h.GetClient)
		auth.PATCH("/clients/:id", h.UpdateClient)
		auth.DELETE("/clients/:id", h.DeleteClient)
		auth.GET("/clients/:id/orders", h.GetClientOrders)
		auth.DELETE("/clients/:id/orders", h.ClearClientOrders)
		auth.GET("/clients/:id/quotes", h.GetClientQuotes)
		auth.DELETE("/clients/:id/quotes", h.ClearClientQuotes)
		auth.GET("/payments", h.GetAllPayments)
		auth.GET("/clients/:id/payment-history", h.GetClientPayments)
		auth.DELETE("/clients/:id/payment-history", h.ClearClientPayments)
		auth.POST("/clients/:id/payment", h.RegisterPayment)
		auth.POST("/clients/:id/debt", h.AddDebt)

		auth.GET("/quotes", h.ListQuotes)
		auth.POST("/quotes", h.CreateQuote)
		auth.GET("/quotes/:id", h.GetQuote)
		auth.POST("/quotes/:id/approve", h.ApproveQuote)
		auth.POST("/quotes/:id/deliver", h.DeliverQuote)
		auth.POST("/quotes/:id/invoice", h.InvoiceQuote)
		auth.POST("/quotes/:id/cancel", h.CancelQuote)
		auth.DELETE("/quotes/:id", h.DeleteQuote)

		auth.GET("/expenses", h.ListExpenses)
		auth.POST("/expenses", h.CreateExpense)
		auth.POST("/expenses/:id/pay", h.PayExpense)
		auth.DELETE("/expenses/:id", h.DeleteExpense)

		auth.GET("/capital-contributions", h.ListCapitalContributions)
		auth.POST("/capital-contributions", h.CreateCapitalContribution)
		auth.DELETE("/capital-contributions/:id", h.DeleteCapitalContribution)
	}

	return r
}
