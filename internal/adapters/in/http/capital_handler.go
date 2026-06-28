package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type createCapitalContributionRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type"`
}

func (req createCapitalContributionRequest) toDomain(userID string) domain.CapitalContribution {
	return domain.CapitalContribution{
		UserID:      userID,
		Description: req.Description,
		Amount:      domain.NewMoneyFromFloat(req.Amount),
		Type:        domain.ContributionType(req.Type),
	}
}

// @Summary     Listar aportes de capital
// @Tags        Capital
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.CapitalContribution
// @Failure     500 {object} errorResponse
// @Router      /capital-contributions [get]
func (h *Handler) ListCapitalContributions(c *gin.Context) {
	contributions, err := h.capitalUC.ListContributions(c.Request.Context(), filterID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, contributions)
}

// @Summary     Criar aporte de capital
// @Tags        Capital
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createCapitalContributionRequest true "Dados do aporte"
// @Success     201 {object} domain.CapitalContribution
// @Failure     400 {object} errorResponse
// @Router      /capital-contributions [post]
func (h *Handler) CreateCapitalContribution(c *gin.Context) {
	var req createCapitalContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "valor inválido")
		return
	}
	userID := ""
	if u := ctxUser(c); u != nil {
		userID = u.ID
	}
	e := req.toDomain(userID)
	if err := h.capitalUC.AddContribution(c.Request.Context(), &e); err != nil {
		respondError(c, err)
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
	if err := h.capitalUC.DeleteContribution(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "aporte excluído"})
}
