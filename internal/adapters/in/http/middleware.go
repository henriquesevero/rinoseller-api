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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "não autorizado"})
			c.Abort()
			return
		}
		user, err := authUC.ValidateToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido ou expirado"})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := ctxUser(c)
		if user == nil || user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "acesso restrito ao administrador"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ctxUser extrai o usuário autenticado do contexto Gin.
func ctxUser(c *gin.Context) *domain.User {
	v, _ := c.Get("user")
	u, _ := v.(*domain.User)
	return u
}

// filterID retorna o userID para filtrar dados; admin passa "" (vê tudo).
func filterID(c *gin.Context) string {
	u := ctxUser(c)
	if u == nil || u.Role == "admin" {
		return ""
	}
	return u.ID
}
