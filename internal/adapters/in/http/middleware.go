package httphandler

import (
	"net/http"
	"strings"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authUC ports.AuthUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, errorResponse{Error: "não autorizado"})
			c.Abort()
			return
		}
		user, err := authUC.ValidateToken(c.Request.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.JSON(http.StatusUnauthorized, errorResponse{Error: "token inválido ou expirado"})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func SubscriptionRequired(authUC ports.AuthUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := ctxUser(c)
		if user == nil {
			c.Next()
			return
		}
		if err := authUC.CheckAccess(c.Request.Context(), user.ID); err != nil {
			respondError(c, err)
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := ctxUser(c)
		if user == nil || !user.Role.IsAdmin() {
			c.JSON(http.StatusForbidden, errorResponse{Error: "acesso restrito ao administrador"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func ctxUser(c *gin.Context) *domain.User {
	v, _ := c.Get("user")
	u, _ := v.(*domain.User)
	return u
}

func filterID(c *gin.Context) string {
	u := ctxUser(c)
	if u == nil || u.Role.IsAdmin() {
		return ""
	}
	return u.ID
}
