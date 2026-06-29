package httphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type sendDocumentEmailRequest struct {
	Subject   string `json:"subject" binding:"required" example:"Tabela de Preços — Elements"`
	Message   string `json:"message" binding:"required" example:"Segue em anexo nossa tabela de preços atualizada."`
	Filename  string `json:"filename" binding:"required" example:"tabela-elements.pdf"`
	PDFBase64 string `json:"pdf_base64" binding:"required"`
}

// @Summary     Enviar documento por e-mail para um cliente
// @Description Envia um PDF qualquer (ex: tabela de preços) por e-mail, em anexo, para o cliente informado.
// @Tags        Clients
// @Accept      json
// @Produce     json
// @Param       id   path string true "ID do cliente"
// @Param       body body sendDocumentEmailRequest true "Assunto, mensagem e PDF em base64"
// @Success     200 {object} messageResponse
// @Failure     400 {object} errorResponse
// @Router      /clients/{id}/send-document-email [post]
func (h *Handler) SendDocumentEmail(c *gin.Context) {
	var body sendDocumentEmailRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, "preencha assunto, mensagem, nome do arquivo e o PDF")
		return
	}
	if err := h.documentEmailUC.SendDocumentEmail(c.Request.Context(), c.Param("id"), body.Subject, body.Message, body.Filename, body.PDFBase64); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "documento enviado por e-mail com sucesso"})
}
