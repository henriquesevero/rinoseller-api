package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

type ResendSender struct {
	client    *resend.Client
	fromEmail string
}

func NewResendSender(apiKey, fromEmail string) *ResendSender {
	return &ResendSender{client: resend.NewClient(apiKey), fromEmail: fromEmail}
}

func (s *ResendSender) Send(ctx context.Context, toEmail, toName, subject, htmlBody string) error {
	to := toEmail
	if toName != "" {
		to = fmt.Sprintf("%s <%s>", toName, toEmail)
	}
	_, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	})
	if err != nil {
		return fmt.Errorf("falha ao enviar e-mail via resend: %w", err)
	}
	return nil
}
