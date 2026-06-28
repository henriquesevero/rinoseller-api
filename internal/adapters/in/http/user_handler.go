package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type createUserRequest struct {
	Name     string `json:"name" binding:"required" example:"Maria"`
	Email    string `json:"email" binding:"required" example:"maria@email.com"`
	Password string `json:"password" binding:"required" example:"123456"`
	Role     string `json:"role" example:"seller"`
}

type updateUserRequest struct {
	Name   string `json:"name" example:"Maria Silva"`
	Email  string `json:"email" example:"maria@email.com"`
	Active bool   `json:"active" example:"true"`
}

// @Summary     Listar usuários
// @Tags        Usuários
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.User
// @Failure     500 {object} errorResponse
// @Router      /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.userUC.ListUsers(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

// @Summary     Criar usuário
// @Tags        Usuários
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createUserRequest true "Dados do usuário"
// @Success     201 {object} domain.User
// @Failure     400 {object} errorResponse
// @Failure     409 {object} errorResponse
// @Router      /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var body createUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}
	user, err := h.userUC.CreateUser(c.Request.Context(), body.Name, body.Email, body.Password, domain.UserRole(body.Role))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

// @Summary     Atualizar usuário
// @Tags        Usuários
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string          true "ID do usuário"
// @Param       body body updateUserRequest true "Dados atualizados"
// @Success     200 {object} domain.User
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Router      /users/{id} [patch]
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var body updateUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}
	user, err := h.userUC.UpdateUser(c.Request.Context(), id, body.Name, body.Email, body.Active)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}
