package httphandler

import (
	"fmt"
	"html"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

var (
	driveIDQueryParam = regexp.MustCompile(`[?&]id=([^&]+)`)
	driveIDPathParam  = regexp.MustCompile(`/d/([^/]+)`)
)

func driveThumbnailURL(driveURL string) string {
	if m := driveIDQueryParam.FindStringSubmatch(driveURL); len(m) > 1 {
		return fmt.Sprintf("https://drive.google.com/thumbnail?id=%s&sz=w800", m[1])
	}
	if m := driveIDPathParam.FindStringSubmatch(driveURL); len(m) > 1 {
		return fmt.Sprintf("https://drive.google.com/thumbnail?id=%s&sz=w800", m[1])
	}
	return ""
}

// @Summary     Página de compartilhamento de catálogo
// @Description Gera uma página com metadados Open Graph (capa, título) para pré-visualização no WhatsApp e redireciona para o PDF no Google Drive.
// @Tags        Catálogos
// @Produce     html
// @Param       id path string true "ID do catálogo"
// @Success     200 {string} string "HTML"
// @Router      /c/{id} [get]
func (h *Handler) ShareBrandCatalog(c *gin.Context) {
	catalog, err := h.brandCatalogUC.GetCatalog(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html lang="pt-BR"><head><meta charset="utf-8"><title>Catálogo não encontrado</title></head><body>Catálogo não encontrado.</body></html>`))
		return
	}

	title := html.EscapeString(fmt.Sprintf("Catálogo %s — RinoSeller", catalog.BrandName))
	desc := "Confira o catálogo digital completo."
	thumb := html.EscapeString(driveThumbnailURL(catalog.DriveURL))
	driveURL := html.EscapeString(catalog.DriveURL)
	brandName := html.EscapeString(catalog.BrandName)

	page := fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:image" content="%s">
<meta property="og:type" content="website">
<meta name="twitter:card" content="summary_large_image">
<meta http-equiv="refresh" content="0; url=%s">
<style>
  body { font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Arial,sans-serif; background:#0a0a0a; color:#e5e7eb; display:flex; align-items:center; justify-content:center; min-height:100vh; margin:0; text-align:center; padding:24px; }
  a { color:#28AEA4; font-weight:600; text-decoration:none; }
</style>
</head>
<body>
  <div>
    <p>Abrindo o catálogo de %s...</p>
    <p><a href="%s">Clique aqui se não for redirecionado automaticamente</a></p>
  </div>
</body>
</html>`, title, title, desc, thumb, driveURL, brandName, driveURL)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}
