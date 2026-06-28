package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type orderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type createOrderRequest struct {
	ClientName  string             `json:"client_name"`
	ClientPhone string             `json:"client_phone"`
	Items       []orderItemRequest `json:"items" binding:"required,min=1"`
}

func (req createOrderRequest) toDomain(userID string) domain.Order {
	items := make([]domain.OrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = domain.OrderItem{ProductID: it.ProductID, Quantity: it.Quantity}
	}
	return domain.Order{
		UserID:      userID,
		ClientName:  req.ClientName,
		ClientPhone: req.ClientPhone,
		Items:       items,
	}
}

// @Summary     Listar pedidos
// @Tags        Pedidos
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Order
// @Failure     500 {object} errorResponse
// @Router      /orders [get]
func (h *Handler) ListOrders(c *gin.Context) {
	orders, err := h.orderUC.ListOrders(c.Request.Context(), filterID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, orders)
}

// @Summary     Criar pedido
// @Tags        Pedidos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createOrderRequest true "Dados do pedido"
// @Success     201 {object} domain.Order
// @Failure     400 {object} errorResponse
// @Router      /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "pedido deve ter ao menos um item")
		return
	}
	userID := ""
	if u := ctxUser(c); u != nil {
		userID = u.ID
	}
	order := req.toDomain(userID)
	if err := h.orderUC.CreateOrder(c.Request.Context(), &order); err != nil {
		respondError(c, err)
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
	if err := h.orderUC.DeleteOrder(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "pedido excluído"})
}
