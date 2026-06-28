package httphandler

import (
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

type createBrandCatalogRequest struct {
	BrandName string `json:"brand_name" binding:"required"`
	DriveURL  string `json:"drive_url" binding:"required"`
}

func (req createBrandCatalogRequest) toDomain(userID string) domain.BrandCatalog {
	return domain.BrandCatalog{
		UserID:    userID,
		BrandName: req.BrandName,
		DriveURL:  req.DriveURL,
	}
}

// @Summary     Listar catálogos de marca
// @Tags        Catálogos de Marca
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} domain.BrandCatalog
// @Failure     500 {object} errorResponse
// @Router      /brand-catalogs [get]
func (h *Handler) ListBrandCatalogs(c *gin.Context) {
	catalogs, err := h.brandCatalogUC.ListCatalogs(c.Request.Context(), filterID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, catalogs)
}

// @Summary     Cadastrar catálogo de marca
// @Tags        Catálogos de Marca
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createBrandCatalogRequest true "Dados do catálogo"
// @Success     201 {object} domain.BrandCatalog
// @Failure     400 {object} errorResponse
// @Router      /brand-catalogs [post]
func (h *Handler) CreateBrandCatalog(c *gin.Context) {
	var req createBrandCatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "nome da marca e link são obrigatórios")
		return
	}
	userID := ""
	if u := ctxUser(c); u != nil {
		userID = u.ID
	}
	catalog := req.toDomain(userID)
	if err := h.brandCatalogUC.AddCatalog(c.Request.Context(), &catalog); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, catalog)
}

// @Summary     Excluir catálogo de marca
// @Tags        Catálogos de Marca
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "ID do catálogo"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /brand-catalogs/{id} [delete]
func (h *Handler) DeleteBrandCatalog(c *gin.Context) {
	if err := h.brandCatalogUC.DeleteCatalog(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "catálogo excluído"})
}
