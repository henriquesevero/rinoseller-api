package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type kitItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type createProductRequest struct {
	Code          string           `json:"code"`
	Name          string           `json:"name" binding:"required"`
	Category      string           `json:"category"`
	Brand         string           `json:"brand"`
	Price         float64          `json:"price"`
	CostPrice     float64          `json:"cost_price"`
	StockQuantity int              `json:"stock_quantity"`
	IsKit         bool             `json:"is_kit"`
	KitItems      []kitItemRequest `json:"kit_items,omitempty"`
}

type updatePriceRequest struct {
	Price float64 `json:"price" example:"29.90"`
}

type updateStockRequest struct {
	Quantity int `json:"quantity" example:"50"`
}

func (req createProductRequest) toDomain(userID string) domain.Product {
	kitItems := make([]domain.KitItem, len(req.KitItems))
	for i, ki := range req.KitItems {
		kitItems[i] = domain.KitItem{ProductID: ki.ProductID, Quantity: ki.Quantity}
	}
	return domain.Product{
		UserID:        userID,
		Code:          req.Code,
		Name:          req.Name,
		Category:      req.Category,
		Brand:         req.Brand,
		Price:         domain.NewMoneyFromFloat(req.Price),
		CostPrice:     domain.NewMoneyFromFloat(req.CostPrice),
		StockQuantity: req.StockQuantity,
		IsKit:         req.IsKit,
		KitItems:      kitItems,
	}
}

// @Summary     Listar produtos
// @Tags        Produtos
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Product
// @Failure     500 {object} errorResponse
// @Router      /products [get]
func (h *Handler) ListProducts(c *gin.Context) {
	products, err := h.productUC.ListProducts(c.Request.Context(), filterID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, products)
}

// @Summary     Criar produto
// @Tags        Produtos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createProductRequest true "Dados do produto"
// @Success     201 {object} domain.Product
// @Failure     400 {object} errorResponse
// @Router      /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}
	userID := ""
	if u := ctxUser(c); u != nil {
		userID = u.ID
	}
	p := req.toDomain(userID)
	if err := h.productUC.CreateProduct(c.Request.Context(), &p); err != nil {
		respondError(c, err)
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
	var body updatePriceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "preço inválido")
		return
	}
	if err := h.productUC.UpdatePrice(c.Request.Context(), id, domain.NewMoneyFromFloat(body.Price)); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "preço atualizado com sucesso"})
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
	var body updateStockRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}
	if err := h.productUC.UpdateStock(c.Request.Context(), id, body.Quantity); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "estoque atualizado com sucesso"})
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
	if err := h.productUC.DeleteProduct(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "produto excluído"})
}
