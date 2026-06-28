package httphandler

import (
	"net/http"
	"time"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type createExpenseRequest struct {
	Description   string     `json:"description" binding:"required"`
	Supplier      string     `json:"supplier"`
	Amount        float64    `json:"amount" binding:"required,gt=0"`
	PaymentMethod string     `json:"payment_method"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	Notes         string     `json:"notes"`
}

func (req createExpenseRequest) toDomain(userID string) domain.Expense {
	return domain.Expense{
		UserID:        userID,
		Description:   req.Description,
		Supplier:      req.Supplier,
		Amount:        domain.NewMoneyFromFloat(req.Amount),
		PaymentMethod: req.PaymentMethod,
		DueDate:       req.DueDate,
		Notes:         req.Notes,
	}
}

// @Summary     Listar despesas
// @Tags        Despesas
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.Expense
// @Failure     500 {object} errorResponse
// @Router      /expenses [get]
func (h *Handler) ListExpenses(c *gin.Context) {
	expenses, err := h.expenseUC.ListExpenses(c.Request.Context(), filterID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, expenses)
}

// @Summary     Criar despesa
// @Tags        Despesas
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createExpenseRequest true "Dados da despesa"
// @Success     201 {object} domain.Expense
// @Failure     400 {object} errorResponse
// @Router      /expenses [post]
func (h *Handler) CreateExpense(c *gin.Context) {
	var req createExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "descrição e valor são obrigatórios")
		return
	}
	userID := ""
	if u := ctxUser(c); u != nil {
		userID = u.ID
	}
	e := req.toDomain(userID)
	if err := h.expenseUC.CreateExpense(c.Request.Context(), &e); err != nil {
		respondError(c, err)
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
	expense, err := h.expenseUC.PayExpense(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
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
	if err := h.expenseUC.DeleteExpense(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "conta excluída"})
}
