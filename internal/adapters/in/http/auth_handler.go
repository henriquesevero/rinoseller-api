package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required" example:"admin@rinoseller.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

type registerRequest struct {
	Name     string `json:"name" binding:"required" example:"João Silva"`
	Email    string `json:"email" binding:"required" example:"joao@email.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" example:"joao@email.com"`
}

type resetPasswordRequest struct {
	Token    string `json:"token" example:"abc123"`
	Password string `json:"password" example:"novaSenha123"`
}

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required" example:"abc123"`
}

type resendVerificationRequest struct {
	Email string `json:"email" binding:"required" example:"joao@email.com"`
}

type authResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  domain.User `json:"user"`
}

// @Summary     Registrar conta
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body registerRequest true "Dados de registro"
// @Success     201 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Failure     409 {object} errorResponse
// @Router      /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var body registerRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "preencha todos os campos")
		return
	}
	if len(body.Password) < 6 {
		badRequest(c, "a senha deve ter ao menos 6 caracteres")
		return
	}

	if _, err := h.authUC.Register(c.Request.Context(), body.Name, body.Email, body.Password); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, messageResponse{Message: "conta criada! verifique seu e-mail para ativá-la"})
}

// @Summary     Confirmar e-mail
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body verifyEmailRequest true "Token de verificação"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	var body verifyEmailRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "token é obrigatório")
		return
	}
	if err := h.authUC.VerifyEmail(c.Request.Context(), body.Token); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "e-mail confirmado com sucesso"})
}

// @Summary     Reenviar e-mail de confirmação
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body resendVerificationRequest true "E-mail"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /auth/resend-verification [post]
func (h *Handler) ResendVerificationEmail(c *gin.Context) {
	var body resendVerificationRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Email == "" {
		badRequest(c, "informe o e-mail")
		return
	}
	if err := h.authUC.ResendVerificationEmail(c.Request.Context(), body.Email); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "se o e-mail estiver cadastrado e pendente de confirmação, enviaremos um novo link"})
}

// @Summary     Esqueci a senha
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body forgotPasswordRequest true "E-mail"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var body forgotPasswordRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Email == "" {
		badRequest(c, "informe o e-mail")
		return
	}
	if err := h.authUC.ForgotPassword(c.Request.Context(), body.Email); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "se o e-mail estiver cadastrado, enviaremos um link de recuperação"})
}

// @Summary     Redefinir senha
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body resetPasswordRequest true "Token e nova senha"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var body resetPasswordRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Token == "" || body.Password == "" {
		badRequest(c, "preencha todos os campos")
		return
	}
	if len(body.Password) < 6 {
		badRequest(c, "a senha deve ter ao menos 6 caracteres")
		return
	}
	if err := h.authUC.ResetPassword(c.Request.Context(), body.Token, body.Password); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "senha redefinida com sucesso"})
}

// @Summary     Login
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body loginRequest true "Credenciais"
// @Success     200 {object} authResponse
// @Failure     400 {object} errorResponse
// @Failure     401 {object} errorResponse
// @Router      /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var body loginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "e-mail e senha são obrigatórios")
		return
	}
	token, user, err := h.authUC.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, authResponse{Token: token, User: *user})
}

// @Summary     Meu perfil
// @Tags        Auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} domain.User
// @Router      /auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	c.JSON(http.StatusOK, ctxUser(c))
}
