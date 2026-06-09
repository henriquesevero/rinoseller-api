package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authUC    ports.AuthUseCase
	userUC    ports.UserUseCase
	productUC ports.ProductUseCase
	orderUC   ports.OrderUseCase
	clientUC  ports.ClientUseCase
	quoteUC   ports.QuoteUseCase
	expenseUC ports.ExpenseUseCase
	capitalUC ports.CapitalContributionUseCase
}

func NewHandler(
	authUC ports.AuthUseCase,
	userUC ports.UserUseCase,
	productUC ports.ProductUseCase,
	orderUC ports.OrderUseCase,
	clientUC ports.ClientUseCase,
	quoteUC ports.QuoteUseCase,
	expenseUC ports.ExpenseUseCase,
	capitalUC ports.CapitalContributionUseCase,
) *Handler {
	return &Handler{authUC: authUC, userUC: userUC, productUC: productUC, orderUC: orderUC, clientUC: clientUC, quoteUC: quoteUC, expenseUC: expenseUC, capitalUC: capitalUC}
}

// ── Request body types (used by swagger annotations) ─────────────────────────

type loginRequest struct {
	Email    string `json:"email" example:"admin@rinoseller.com"`
	Password string `json:"password" example:"123456"`
}

type registerRequest struct {
	Name     string `json:"name" example:"João Silva"`
	Email    string `json:"email" example:"joao@email.com"`
	Password string `json:"password" example:"123456"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" example:"joao@email.com"`
}

type resetPasswordRequest struct {
	Token    string `json:"token" example:"abc123"`
	Password string `json:"password" example:"novaSenha123"`
}

type createUserRequest struct {
	Name     string `json:"name" example:"Maria"`
	Email    string `json:"email" example:"maria@email.com"`
	Password string `json:"password" example:"123456"`
	Role     string `json:"role" example:"seller"`
}

type updateUserRequest struct {
	Name   string `json:"name" example:"Maria Silva"`
	Email  string `json:"email" example:"maria@email.com"`
	Active bool   `json:"active" example:"true"`
}

type updatePriceRequest struct {
	Price float64 `json:"price" example:"29.90"`
}

type updateStockRequest struct {
	Quantity int `json:"quantity" example:"50"`
}

type registerPaymentRequest struct {
	Amount         float64 `json:"amount" example:"150.00"`
	Notes          string  `json:"notes" example:"Pagamento parcial"`
	CountAsRevenue bool    `json:"count_as_revenue" example:"true"`
}

type addDebtRequest struct {
	Amount float64 `json:"amount" example:"200.00"`
}

type messageResponse struct {
	Message string `json:"message" example:"operação realizada com sucesso"`
}

type errorResponse struct {
	Error string `json:"error" example:"mensagem de erro"`
}

type authResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  domain.User `json:"user"`
}

// ── Auth ──────────────────────────────────────────────────────────────────────

// @Summary     Registrar conta
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body registerRequest true "Dados de registro"
// @Success     201 {object} authResponse
// @Failure     400 {object} errorResponse
// @Failure     409 {object} errorResponse
// @Router      /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var body struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "preencha todos os campos"})
		return
	}
	if len(body.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a senha deve ter ao menos 6 caracteres"})
		return
	}

	user, err := h.userUC.CreateUser(body.Name, body.Email, body.Password, "seller")
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	token, _, err := h.authUC.Login(body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{"user": user})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

// @Summary     Esqueci a senha
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       body body forgotPasswordRequest true "E-mail"
// @Success     200 {object} map[string]string
// @Failure     400 {object} errorResponse
// @Router      /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "informe o e-mail"})
		return
	}
	token, err := h.authUC.ForgotPassword(body.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
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
	var body struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "preencha todos os campos"})
		return
	}
	if len(body.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a senha deve ter ao menos 6 caracteres"})
		return
	}
	if err := h.authUC.ResetPassword(body.Token, body.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "senha redefinida com sucesso"})
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
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "e-mail e senha são obrigatórios"})
		return
	}
	token, user, err := h.authUC.Login(body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
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

// ── Users (admin) ─────────────────────────────────────────────────────────────

// @Summary     Listar usuários
// @Tags        Usuários
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.User
// @Failure     500 {object} errorResponse
// @Router      /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.userUC.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	var body struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Role != "admin" && body.Role != "seller" {
		body.Role = "seller"
	}
	user, err := h.userUC.CreateUser(body.Name, body.Email, body.Password, body.Role)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
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
	var body struct {
		Name   string `json:"name" binding:"required"`
		Email  string `json:"email" binding:"required"`
		Active bool   `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.userUC.UpdateUser(id, body.Name, body.Email, body.Active)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// ── Products ──────────────────────────────────────────────────────────────────

// @Summary     Listar produtos
// @Tags        Produtos
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Product
// @Failure     500 {object} errorResponse
// @Router      /products [get]
func (h *Handler) ListProducts(c *gin.Context) {
	products, err := h.productUC.ListProducts(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// @Summary     Criar produto
// @Tags        Produtos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.Product true "Dados do produto"
// @Success     201 {object} domain.Product
// @Failure     400 {object} errorResponse
// @Router      /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var p domain.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if u := ctxUser(c); u != nil {
		p.UserID = u.ID
	}
	if err := h.productUC.CreateProduct(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// @Summary     Atualizar preço
// @Tags        Produtos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string          true "ID do produto"
// @Param       body body updatePriceRequest true "Novo preço"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Router      /products/{id}/price [patch]
func (h *Handler) UpdatePrice(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Price float64 `json:"price" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "preço inválido: deve ser maior que zero"})
		return
	}
	if err := h.productUC.UpdatePrice(id, body.Price); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "preço atualizado com sucesso"})
}

// @Summary     Atualizar estoque
// @Tags        Produtos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string           true "ID do produto"
// @Param       body body updateStockRequest true "Nova quantidade"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Router      /products/{id}/stock [patch]
func (h *Handler) UpdateStock(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Quantity < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantidade não pode ser negativa"})
		return
	}
	if err := h.productUC.UpdateStock(id, body.Quantity); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "estoque atualizado com sucesso"})
}

// @Summary     Excluir produto
// @Tags        Produtos
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do produto"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	if err := h.productUC.DeleteProduct(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "produto excluído"})
}

// ── Orders ────────────────────────────────────────────────────────────────────

// @Summary     Listar pedidos
// @Tags        Pedidos
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Order
// @Failure     500 {object} errorResponse
// @Router      /orders [get]
func (h *Handler) ListOrders(c *gin.Context) {
	orders, err := h.orderUC.ListOrders(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// @Summary     Criar pedido
// @Tags        Pedidos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.Order true "Dados do pedido"
// @Success     201 {object} domain.Order
// @Failure     400 {object} errorResponse
// @Router      /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var order domain.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(order.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pedido deve ter ao menos um item"})
		return
	}
	if u := ctxUser(c); u != nil {
		order.UserID = u.ID
	}
	if err := h.orderUC.CreateOrder(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// @Summary     Excluir pedido
// @Tags        Pedidos
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do pedido"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /orders/{id} [delete]
func (h *Handler) DeleteOrder(c *gin.Context) {
	if err := h.orderUC.DeleteOrder(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pedido excluído"})
}

// ── Clients ───────────────────────────────────────────────────────────────────

// @Summary     Listar clientes
// @Tags        Clientes
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Client
// @Failure     500 {object} errorResponse
// @Router      /clients [get]
func (h *Handler) ListClients(c *gin.Context) {
	clients, err := h.clientUC.ListClients(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, clients)
}

// @Summary     Criar cliente
// @Tags        Clientes
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.Client true "Dados do cliente"
// @Success     201 {object} domain.Client
// @Failure     400 {object} errorResponse
// @Router      /clients [post]
func (h *Handler) CreateClient(c *gin.Context) {
	var client domain.Client
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if client.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nome é obrigatório"})
		return
	}
	if u := ctxUser(c); u != nil {
		client.UserID = u.ID
	}
	if err := h.clientUC.CreateClient(&client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	client, err := h.clientUC.GetClient(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
// @Param       body body domain.Client true "Dados atualizados"
// @Success     200 {object} domain.Client
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Router      /clients/{id} [patch]
func (h *Handler) UpdateClient(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.clientUC.GetClient(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing.ID = id
	if err := h.clientUC.UpdateClient(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
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
	id := c.Param("id")
	if err := h.quoteUC.ClearClientQuotes(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.clientUC.ClearClientOrders(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.clientUC.ClearPaymentHistory(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.clientUC.DeleteClient(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cliente excluído"})
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
	orders, err := h.clientUC.GetClientOrders(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
	if err := h.clientUC.ClearClientOrders(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pedidos do catálogo removidos"})
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
	quotes, err := h.quoteUC.GetClientQuotes(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
	if err := h.quoteUC.ClearClientQuotes(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pedidos removidos"})
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
	var body struct {
		Amount         float64 `json:"amount"`
		Notes          string  `json:"notes"`
		CountAsRevenue bool    `json:"count_as_revenue"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valor inválido"})
		return
	}
	userID := ""
	if u := ctxUser(c); u != nil {
		userID = u.ID
	}
	if err := h.clientUC.RegisterPayment(id, userID, body.Amount, body.Notes, body.CountAsRevenue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	client, _ := h.clientUC.GetClient(id)
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
	payments, err := h.clientUC.GetClientPayments(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	if err := h.clientUC.ClearPaymentHistory(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "histórico de recebimentos limpo"})
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
	var body struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valor inválido"})
		return
	}
	if err := h.clientUC.AddDebt(id, body.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	client, _ := h.clientUC.GetClient(id)
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
	payments, err := h.clientUC.GetAllPayments(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payments)
}

// ── Quotes ────────────────────────────────────────────────────────────────────

// @Summary     Listar orçamentos
// @Tags        Orçamentos
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Quote
// @Failure     500 {object} errorResponse
// @Router      /quotes [get]
func (h *Handler) ListQuotes(c *gin.Context) {
	quotes, err := h.quoteUC.ListQuotes(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quotes)
}

// @Summary     Criar orçamento
// @Tags        Orçamentos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.Quote true "Dados do orçamento"
// @Success     201 {object} domain.Quote
// @Failure     400 {object} errorResponse
// @Router      /quotes [post]
func (h *Handler) CreateQuote(c *gin.Context) {
	var q domain.Quote
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if q.ClientID == "" || len(q.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cliente e itens são obrigatórios"})
		return
	}
	if u := ctxUser(c); u != nil {
		q.UserID = u.ID
	}
	if err := h.quoteUC.CreateQuote(&q); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, q)
}

// @Summary     Buscar orçamento
// @Tags        Orçamentos
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do orçamento"
// @Success     200 {object} domain.Quote
// @Failure     404 {object} errorResponse
// @Router      /quotes/{id} [get]
func (h *Handler) GetQuote(c *gin.Context) {
	q, err := h.quoteUC.GetQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

// @Summary     Aprovar orçamento
// @Tags        Orçamentos
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do orçamento"
// @Success     200 {object} domain.Quote
// @Failure     400 {object} errorResponse
// @Router      /quotes/{id}/approve [post]
func (h *Handler) ApproveQuote(c *gin.Context) {
	q, err := h.quoteUC.ApproveQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

// @Summary     Marcar como entregue
// @Tags        Orçamentos
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do orçamento"
// @Success     200 {object} domain.Quote
// @Failure     400 {object} errorResponse
// @Router      /quotes/{id}/deliver [post]
func (h *Handler) DeliverQuote(c *gin.Context) {
	q, err := h.quoteUC.DeliverQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

// @Summary     Faturar orçamento
// @Tags        Orçamentos
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do orçamento"
// @Success     200 {object} domain.Quote
// @Failure     400 {object} errorResponse
// @Router      /quotes/{id}/invoice [post]
func (h *Handler) InvoiceQuote(c *gin.Context) {
	q, err := h.quoteUC.InvoiceQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

// @Summary     Cancelar orçamento
// @Tags        Orçamentos
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do orçamento"
// @Success     200 {object} domain.Quote
// @Failure     400 {object} errorResponse
// @Router      /quotes/{id}/cancel [post]
func (h *Handler) CancelQuote(c *gin.Context) {
	q, err := h.quoteUC.CancelQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

// @Summary     Excluir orçamento
// @Tags        Orçamentos
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do orçamento"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /quotes/{id} [delete]
func (h *Handler) DeleteQuote(c *gin.Context) {
	if err := h.quoteUC.DeleteQuote(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "orçamento excluído"})
}

// ── Expenses (Contas a Pagar) ─────────────────────────────────────────────────

// @Summary     Listar despesas
// @Tags        Despesas
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Expense
// @Failure     500 {object} errorResponse
// @Router      /expenses [get]
func (h *Handler) ListExpenses(c *gin.Context) {
	expenses, err := h.expenseUC.ListExpenses(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, expenses)
}

// @Summary     Criar despesa
// @Tags        Despesas
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.Expense true "Dados da despesa"
// @Success     201 {object} domain.Expense
// @Failure     400 {object} errorResponse
// @Router      /expenses [post]
func (h *Handler) CreateExpense(c *gin.Context) {
	var e domain.Expense
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if e.Description == "" || e.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "descrição e valor são obrigatórios"})
		return
	}
	if u := ctxUser(c); u != nil {
		e.UserID = u.ID
	}
	if err := h.expenseUC.CreateExpense(&e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

// @Summary     Pagar despesa
// @Tags        Despesas
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID da despesa"
// @Success     200 {object} domain.Expense
// @Failure     400 {object} errorResponse
// @Router      /expenses/{id}/pay [post]
func (h *Handler) PayExpense(c *gin.Context) {
	expense, err := h.expenseUC.PayExpense(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, expense)
}

// @Summary     Excluir despesa
// @Tags        Despesas
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID da despesa"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /expenses/{id} [delete]
func (h *Handler) DeleteExpense(c *gin.Context) {
	if err := h.expenseUC.DeleteExpense(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "conta excluída"})
}

// ── Capital Contributions (Aportes de Capital) ───────────────────────────────

// @Summary     Listar aportes de capital
// @Tags        Capital
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.CapitalContribution
// @Failure     500 {object} errorResponse
// @Router      /capital-contributions [get]
func (h *Handler) ListCapitalContributions(c *gin.Context) {
	contributions, err := h.capitalUC.ListContributions(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contributions)
}

// @Summary     Criar aporte de capital
// @Tags        Capital
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.CapitalContribution true "Dados do aporte"
// @Success     201 {object} domain.CapitalContribution
// @Failure     400 {object} errorResponse
// @Router      /capital-contributions [post]
func (h *Handler) CreateCapitalContribution(c *gin.Context) {
	var e domain.CapitalContribution
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if e.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valor inválido"})
		return
	}
	if u := ctxUser(c); u != nil {
		e.UserID = u.ID
	}
	if err := h.capitalUC.AddContribution(&e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, e)
}

// @Summary     Excluir aporte de capital
// @Tags        Capital
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do aporte"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /capital-contributions/{id} [delete]
func (h *Handler) DeleteCapitalContribution(c *gin.Context) {
	if err := h.capitalUC.DeleteContribution(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "aporte excluído"})
}
