package email

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"rinoseller-api/internal/core/ports"

	"github.com/resend/resend-go/v2"
)

type ResendSender struct {
	client    *resend.Client
	fromEmail string
}

func NewResendSender(apiKey, fromEmail string) *ResendSender {
	httpClient := &http.Client{Timeout: 60 * time.Second}
	return &ResendSender{client: resend.NewCustomClient(httpClient, apiKey), fromEmail: fromEmail}
}

func (s *ResendSender) Send(ctx context.Context, toEmail, toName, subject, htmlBody string, attachments ...ports.EmailAttachment) error {
	to := toEmail
	if toName != "" {
		to = fmt.Sprintf("%s <%s>", toName, toEmail)
	}

	resendAttachments := make([]*resend.Attachment, 0, len(attachments))
	for _, a := range attachments {
		resendAttachments = append(resendAttachments, &resend.Attachment{
			Filename: a.Filename,
			Content:  a.Content,
		})
	}

	_, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:        s.fromEmail,
		To:          []string{to},
		Subject:     subject,
		Html:        htmlBody,
		Attachments: resendAttachments,
	})
	if err != nil {
		return fmt.Errorf("falha ao enviar e-mail via resend: %w", err)
	}
	return nil
}
