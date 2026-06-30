package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type quoteItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type createQuoteRequest struct {
	ClientID      string             `json:"client_id" binding:"required"`
	Items         []quoteItemRequest `json:"items" binding:"required,min=1"`
	Notes         string             `json:"notes"`
	PaymentType   string             `json:"payment_type"`
	Installments  int                `json:"installments"`
	DiscountType  string             `json:"discount_type"`
	DiscountValue float64            `json:"discount_value"`
}

func (req createQuoteRequest) toDomain(userID string) domain.Quote {
	items := make([]domain.QuoteItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = domain.QuoteItem{ProductID: it.ProductID, Quantity: it.Quantity}
	}
	return domain.Quote{
		UserID:        userID,
		ClientID:      req.ClientID,
		Items:         items,
		Notes:         req.Notes,
		PaymentType:   domain.PaymentType(req.PaymentType),
		Installments:  req.Installments,
		DiscountType:  req.DiscountType,
		DiscountValue: domain.NewMoneyFromFloat(req.DiscountValue),
	}
}

// @Summary     Listar orçamentos
// @Tags        Orçamentos
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Quote
// @Failure     500 {object} errorResponse
// @Router      /quotes [get]
func (h *Handler) ListQuotes(c *gin.Context) {
	quotes, err := h.quoteUC.ListQuotes(c.Request.Context(), filterID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quotes)
}

// @Summary     Criar orçamento
// @Tags        Orçamentos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createQuoteRequest true "Dados do orçamento"
// @Success     201 {object} domain.Quote
// @Failure     400 {object} errorResponse
// @Router      /quotes [post]
func (h *Handler) CreateQuote(c *gin.Context) {
	var req createQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "cliente e itens são obrigatórios")
		return
	}
	userID := ""
	if u := ctxUser(c); u != nil {
		userID = u.ID
	}
	q := req.toDomain(userID)
	if err := h.quoteUC.CreateQuote(c.Request.Context(), &q); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, q)
}

// @Summary     Atualizar itens de orçamento
// @Tags        Orçamentos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string              true "ID do orçamento"
// @Param       body body createQuoteRequest true "Novos itens"
// @Success     200 {object} domain.Quote
// @Failure     400 {object} errorResponse
// @Failure     404 {object} errorResponse
// @Router      /quotes/{id}/items [patch]
func (h *Handler) UpdateQuoteItems(c *gin.Context) {
	var req createQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "itens inválidos")
		return
	}
	items := make([]domain.QuoteItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = domain.QuoteItem{ProductID: it.ProductID, Quantity: it.Quantity}
	}
	q, err := h.quoteUC.UpdateQuoteItems(c.Request.Context(), c.Param("id"), items, req.DiscountType, domain.NewMoneyFromFloat(req.DiscountValue))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, q)
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
	q, err := h.quoteUC.GetQuote(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
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
	q, err := h.quoteUC.ApproveQuote(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
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
	q, err := h.quoteUC.DeliverQuote(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
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
	q, err := h.quoteUC.InvoiceQuote(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
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
	q, err := h.quoteUC.CancelQuote(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
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
	if err := h.quoteUC.DeleteQuote(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "orçamento excluído"})
}

type sendQuoteEmailRequest struct {
	Kind      string `json:"kind" binding:"required" example:"orcamento"`
	PDFBase64 string `json:"pdf_base64" binding:"required"`
}

// @Summary     Enviar orçamento/pedido por e-mail
// @Tags        Orçamentos
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do orçamento"
// @Param       body body sendQuoteEmailRequest true "Tipo do documento e PDF em base64"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /quotes/{id}/send-email [post]
func (h *Handler) SendQuoteEmail(c *gin.Context) {
	var body sendQuoteEmailRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "tipo do documento e PDF são obrigatórios")
		return
	}
	if err := h.quoteUC.SendQuoteEmail(c.Request.Context(), c.Param("id"), body.Kind, body.PDFBase64); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "e-mail enviado com sucesso"})
}
