package services

import (
	"context"
	"encoding/base64"
	"fmt"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"
)

type DocumentEmailService struct {
	clientRepo  ports.ClientRepository
	emailSender ports.EmailSender
}

func NewDocumentEmailService(clientRepo ports.ClientRepository, emailSender ports.EmailSender) *DocumentEmailService {
	return &DocumentEmailService{clientRepo: clientRepo, emailSender: emailSender}
}

func (s *DocumentEmailService) SendDocumentEmail(ctx context.Context, clientID, subject, message, filename, pdfBase64 string) error {
	client, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		return err
	}
	if client.Email == "" {
		return fmt.Errorf("cliente não tem e-mail cadastrado: %w", domain.ErrValidation)
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(pdfBase64)
	if err != nil {
		return fmt.Errorf("PDF inválido: %w", domain.ErrValidation)
	}

	html := buildDocumentEmailContent(client.Name, message)

	return s.emailSender.Send(ctx, client.Email, client.Name, subject, html, ports.EmailAttachment{
		Filename: filename,
		Content:  pdfBytes,
	})
}

func buildDocumentEmailContent(clientName, message string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Arial,sans-serif;">
	<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:32px 16px;">
		<tr>
			<td align="center">
				<table width="100%%" cellpadding="0" cellspacing="0" style="max-width:560px;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 2px 12px rgba(0,0,0,0.06);">

					<tr>
						<td style="background:#0f172a;padding:28px 36px;">
							<span style="color:#ffffff;font-size:20px;font-weight:800;letter-spacing:.02em;">RinoSeller</span>
							<div style="height:3px;width:36px;background:#28AEA4;border-radius:2px;margin-top:10px;"></div>
						</td>
					</tr>

					<tr>
						<td style="padding:36px 36px 28px;">
							<p style="margin:0 0 16px;color:#0f172a;font-size:16px;line-height:1.6;">
								Olá, <strong>%s</strong>!
							</p>
							<p style="margin:0;color:#475569;font-size:15px;line-height:1.6;white-space:pre-line;">
								%s
							</p>
						</td>
					</tr>

					<tr>
						<td style="padding:24px 36px 32px;border-top:1px solid #f1f5f9;margin-top:8px;">
							<p style="margin:0;color:#cbd5e1;font-size:11px;">RinoSeller · documento gerado automaticamente</p>
						</td>
					</tr>

				</table>
			</td>
		</tr>
	</table>
</body>
</html>`, clientName, message)
}
