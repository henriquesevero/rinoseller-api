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

// ── Auth ──────────────────────────────────────────────────────────────────────

// Register cria uma conta de vendedor (acesso público).
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
	// token vazio = e-mail não encontrado, mas não revelamos isso
	c.JSON(http.StatusOK, gin.H{"token": token})
}

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

func (h *Handler) GetMe(c *gin.Context) {
	c.JSON(http.StatusOK, ctxUser(c))
}

// ── Users (admin) ─────────────────────────────────────────────────────────────

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.userUC.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

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

func (h *Handler) ListProducts(c *gin.Context) {
	products, err := h.productUC.ListProducts(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

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

// ── Orders ────────────────────────────────────────────────────────────────────

func (h *Handler) ListOrders(c *gin.Context) {
	orders, err := h.orderUC.ListOrders(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

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

// ── Clients ───────────────────────────────────────────────────────────────────

func (h *Handler) ListClients(c *gin.Context) {
	clients, err := h.clientUC.ListClients(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, clients)
}

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

func (h *Handler) GetClient(c *gin.Context) {
	client, err := h.clientUC.GetClient(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, client)
}

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

func (h *Handler) GetClientPayments(c *gin.Context) {
	payments, err := h.clientUC.GetClientPayments(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payments)
}

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

func (h *Handler) GetClientOrders(c *gin.Context) {
	orders, err := h.clientUC.GetClientOrders(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *Handler) ClearClientPayments(c *gin.Context) {
	if err := h.clientUC.ClearPaymentHistory(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "histórico de recebimentos limpo"})
}

func (h *Handler) ClearClientOrders(c *gin.Context) {
	if err := h.clientUC.ClearClientOrders(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pedidos do catálogo removidos"})
}

func (h *Handler) ClearClientQuotes(c *gin.Context) {
	if err := h.quoteUC.ClearClientQuotes(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pedidos removidos"})
}

func (h *Handler) GetClientQuotes(c *gin.Context) {
	quotes, err := h.quoteUC.GetClientQuotes(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quotes)
}

func (h *Handler) GetAllPayments(c *gin.Context) {
	payments, err := h.clientUC.GetAllPayments(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payments)
}

// ── Quotes ────────────────────────────────────────────────────────────────────

func (h *Handler) ListQuotes(c *gin.Context) {
	quotes, err := h.quoteUC.ListQuotes(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quotes)
}

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

func (h *Handler) GetQuote(c *gin.Context) {
	q, err := h.quoteUC.GetQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func (h *Handler) ApproveQuote(c *gin.Context) {
	q, err := h.quoteUC.ApproveQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func (h *Handler) DeliverQuote(c *gin.Context) {
	q, err := h.quoteUC.DeliverQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func (h *Handler) InvoiceQuote(c *gin.Context) {
	q, err := h.quoteUC.InvoiceQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func (h *Handler) CancelQuote(c *gin.Context) {
	q, err := h.quoteUC.CancelQuote(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	if err := h.productUC.DeleteProduct(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "produto excluído"})
}

func (h *Handler) DeleteClient(c *gin.Context) {
	id := c.Param("id")
	// Remove vínculos antes de excluir o cliente para não violar chaves estrangeiras
	// (ex.: quotes.client_id referencia clients.id sem cascade).
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

func (h *Handler) DeleteQuote(c *gin.Context) {
	if err := h.quoteUC.DeleteQuote(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "orçamento excluído"})
}

func (h *Handler) DeleteOrder(c *gin.Context) {
	if err := h.orderUC.DeleteOrder(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pedido excluído"})
}

// ── Expenses (Contas a Pagar) ─────────────────────────────────────────────────

func (h *Handler) ListExpenses(c *gin.Context) {
	expenses, err := h.expenseUC.ListExpenses(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, expenses)
}

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

func (h *Handler) PayExpense(c *gin.Context) {
	expense, err := h.expenseUC.PayExpense(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, expense)
}

func (h *Handler) DeleteExpense(c *gin.Context) {
	if err := h.expenseUC.DeleteExpense(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "conta excluída"})
}

// ── Capital Contributions (Aportes de Capital) ───────────────────────────────

func (h *Handler) ListCapitalContributions(c *gin.Context) {
	contributions, err := h.capitalUC.ListContributions(filterID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contributions)
}

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

func (h *Handler) DeleteCapitalContribution(c *gin.Context) {
	if err := h.capitalUC.DeleteContribution(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "aporte excluído"})
}

