package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type clientRequest struct {
	Name               string  `json:"name" binding:"required"`
	Phone              string  `json:"phone"`
	Email              string  `json:"email"`
	Notes              string  `json:"notes"`
	CpfCnpj            string  `json:"cpf_cnpj"`
	Address            string  `json:"address"`
	DebtLimit          float64 `json:"debt_limit"`
	PaymentCycleDays   int     `json:"payment_cycle_days"`
	PaymentCycleAmount float64 `json:"payment_cycle_amount"`
}

func (req clientRequest) toDomain() (domain.Client, error) {
	phone, err := domain.NewPhone(req.Phone)
	if err != nil {
		return domain.Client{}, err
	}
	cpfCnpj, err := domain.NewCpfCnpj(req.CpfCnpj)
	if err != nil {
		return domain.Client{}, err
	}
	return domain.Client{
		Name:               req.Name,
		Phone:              phone,
		Email:              req.Email,
		Notes:              req.Notes,
		CpfCnpj:            cpfCnpj,
		Address:            req.Address,
		DebtLimit:          domain.NewMoneyFromFloat(req.DebtLimit),
		PaymentCycleDays:   req.PaymentCycleDays,
		PaymentCycleAmount: domain.NewMoneyFromFloat(req.PaymentCycleAmount),
	}, nil
}

type registerPaymentRequest struct {
	Amount         float64 `json:"amount" example:"150.00"`
	Notes          string  `json:"notes" example:"Pagamento parcial"`
	CountAsRevenue bool    `json:"count_as_revenue" example:"true"`
}

type addDebtRequest struct {
	Amount float64 `json:"amount" example:"200.00"`
}

// @Summary     Listar clientes
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Client
// @Failure     500 {object} errorResponse
// @Router      /clients [get]
func (h *Handler) ListClients(c *gin.Context) {
	clients, err := h.clientUC.ListClients(c.Request.Context(), filterID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clients)
}

// @Summary     Criar cliente
// @Tags        Clientes
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body clientRequest true "Dados do cliente"
// @Success     201 {object} domain.Client
// @Failure     400 {object} errorResponse
// @Router      /clients [post]
func (h *Handler) CreateClient(c *gin.Context) {
	var req clientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	client, err := req.toDomain()
	if err != nil {
		respondError(c, err)
		return
	}
	if u := ctxUser(c); u != nil {
		client.UserID = u.ID
	}
	if err := h.clientUC.CreateClient(c.Request.Context(), &client); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, client)
}

// @Summary     Buscar cliente
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do cliente"
// @Success     200 {object} domain.Client
// @Failure     404 {object} errorResponse
// @Router      /clients/{id} [get]
func (h *Handler) GetClient(c *gin.Context) {
	client, err := h.clientUC.GetClient(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, client)
}

// @Summary     Atualizar cliente
// @Tags        Clientes
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string        true "ID do cliente"
// @Param       body body clientRequest true "Dados atualizados"
// @Success     200 {object} domain.Client
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Router      /clients/{id} [patch]
func (h *Handler) UpdateClient(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	existing, err := h.clientUC.GetClient(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	var req clientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	updated, err := req.toDomain()
	if err != nil {
		respondError(c, err)
		return
	}

	updated.ID = existing.ID
	updated.UserID = existing.UserID
	updated.CreatedAt = existing.CreatedAt
	updated.Debt = existing.Debt

	if err := h.clientUC.UpdateClient(ctx, &updated); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// @Summary     Excluir cliente
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do cliente"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /clients/{id} [delete]
func (h *Handler) DeleteClient(c *gin.Context) {
	if err := h.clientUC.DeleteClient(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "cliente excluído"})
}

// @Summary     Pedidos do cliente
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do cliente"
// @Success     200 {array} domain.Order
// @Failure     404 {object} errorResponse
// @Router      /clients/{id}/orders [get]
func (h *Handler) GetClientOrders(c *gin.Context) {
	orders, err := h.clientUC.GetClientOrders(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, orders)
}

// @Summary     Limpar pedidos do cliente
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do cliente"
// @Success     200 {object} messageResponse
// @Failure     500 {object} errorResponse
// @Router      /clients/{id}/orders [delete]
func (h *Handler) ClearClientOrders(c *gin.Context) {
	if err := h.clientUC.ClearClientOrders(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "pedidos do catálogo removidos"})
}

// @Summary     Orçamentos do cliente
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do cliente"
// @Success     200 {array} domain.Quote
// @Failure     404 {object} errorResponse
// @Router      /clients/{id}/quotes [get]
func (h *Handler) GetClientQuotes(c *gin.Context) {
	quotes, err := h.quoteUC.GetClientQuotes(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quotes)
}

// @Summary     Limpar orçamentos do cliente
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do cliente"
// @Success     200 {object} messageResponse
// @Failure     500 {object} errorResponse
// @Router      /clients/{id}/quotes [delete]
func (h *Handler) ClearClientQuotes(c *gin.Context) {
	if err := h.quoteUC.ClearClientQuotes(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "pedidos removidos"})
}

// @Summary     Registrar recebimento
// @Tags        Clientes
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string               true "ID do cliente"
// @Param       body body registerPaymentRequest true "Dados do recebimento"
// @Success     200 {object} domain.Client
// @Failure     400 {object} errorResponse
// @Router      /clients/{id}/payment [post]
func (h *Handler) RegisterPayment(c *gin.Context) {
	id := c.Param("id")
	var body registerPaymentRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Amount <= 0 {
		badRequest(c, "valor inválido")
		return
	}
	ctx := c.Request.Context()
	userID := ""
	if u := ctxUser(c); u != nil {
		userID = u.ID
	}
	amount := domain.NewMoneyFromFloat(body.Amount)
	if err := h.clientUC.RegisterPayment(ctx, id, userID, amount, body.Notes, body.CountAsRevenue); err != nil {
		respondError(c, err)
		return
	}
	client, err := h.clientUC.GetClient(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, client)
}

// @Summary     Histórico de recebimentos do cliente
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do cliente"
// @Success     200 {array} domain.ClientPayment
// @Failure     500 {object} errorResponse
// @Router      /clients/{id}/payment-history [get]
func (h *Handler) GetClientPayments(c *gin.Context) {
	payments, err := h.clientUC.GetClientPayments(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, payments)
}

// @Summary     Limpar histórico de recebimentos
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do cliente"
// @Success     200 {object} messageResponse
// @Failure     500 {object} errorResponse
// @Router      /clients/{id}/payment-history [delete]
func (h *Handler) ClearClientPayments(c *gin.Context) {
	if err := h.clientUC.ClearPaymentHistory(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "histórico de recebimentos limpo"})
}

// @Summary     Adicionar dívida
// @Tags        Clientes
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string      true "ID do cliente"
// @Param       body body addDebtRequest true "Valor da dívida"
// @Success     200 {object} domain.Client
// @Failure     400 {object} errorResponse
// @Router      /clients/{id}/debt [post]
func (h *Handler) AddDebt(c *gin.Context) {
	id := c.Param("id")
	var body addDebtRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Amount <= 0 {
		badRequest(c, "valor inválido")
		return
	}
	ctx := c.Request.Context()
	if err := h.clientUC.AddDebt(ctx, id, domain.NewMoneyFromFloat(body.Amount)); err != nil {
		respondError(c, err)
		return
	}
	client, err := h.clientUC.GetClient(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, client)
}

// @Summary     Todos os recebimentos
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.ClientPayment
// @Failure     500 {object} errorResponse
// @Router      /payments [get]
func (h *Handler) GetAllPayments(c *gin.Context) {
	payments, err := h.clientUC.GetAllPayments(c.Request.Context(), filterID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, payments)
}
