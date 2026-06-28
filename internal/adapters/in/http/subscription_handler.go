package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type checkoutRequest struct {
	Email       string `json:"email" binding:"required" example:"joao@email.com"`
	Plan        string `json:"plan" binding:"required" example:"professional"`
	HolderName  string `json:"holder_name" binding:"required" example:"João Silva"`
	Number      string `json:"number" binding:"required" example:"4444444444444444"`
	ExpiryMonth string `json:"expiry_month" binding:"required" example:"12"`
	ExpiryYear  string `json:"expiry_year" binding:"required" example:"2030"`
	CVV         string `json:"cvv" binding:"required" example:"123"`
}

// @Summary     Assinar um plano (checkout simulado)
// @Description Simula a cobrança via Asaas e ativa a assinatura do usuário. Não requer login — usado tanto no fluxo de cadastro quanto a partir da tela de cobrança após o período de teste.
// @Tags        Subscriptions
// @Accept      json
// @Produce     json
// @Param       body body checkoutRequest true "Dados do plano e do cartão"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /subscriptions/checkout [post]
func (h *Handler) CheckoutSubscription(c *gin.Context) {
	var body checkoutRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "preencha todos os campos do pagamento")
		return
	}

	card := ports.CardDetails{
		HolderName:  body.HolderName,
		Number:      body.Number,
		ExpiryMonth: body.ExpiryMonth,
		ExpiryYear:  body.ExpiryYear,
		CVV:         body.CVV,
	}
	if err := h.subscriptionUC.Checkout(c.Request.Context(), body.Email, domain.Plan(body.Plan), card); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "assinatura ativada com sucesso"})
}
