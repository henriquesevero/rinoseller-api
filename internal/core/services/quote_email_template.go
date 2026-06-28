package services

import (
	"fmt"
	"strings"

	"rinoseller-api/internal/core/domain"
)

func quoteDocumentCode(id, kind string) string {
	prefix := "ORC"
	if kind == "pedido" {
		prefix = "PED"
	}
	tail := id
	if len(tail) > 6 {
		tail = tail[len(tail)-6:]
	}
	return prefix + "-" + strings.ToUpper(tail)
}

func quoteDocumentNoun(kind string) string {
	if kind == "pedido" {
		return "pedido"
	}
	return "orçamento"
}

func buildQuoteEmailContent(client *domain.Client, kind, docCode, issuerName string) (subject, html string) {
	noun := quoteDocumentNoun(kind)
	subject = fmt.Sprintf("Seu %s — RinoSeller", noun)

	issuedBy := ""
	if issuerName != "" {
		issuedBy = fmt.Sprintf(`
							<p style="margin:12px 0 0;color:#94a3b8;font-size:13px;line-height:1.6;">
								Emitido por <strong style="color:#475569;">%s</strong>.
							</p>`, issuerName)
	}

	html = fmt.Sprintf(`
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
							<p style="margin:0;color:#475569;font-size:15px;line-height:1.6;">
								Segue em anexo seu %s, em PDF.
							</p>
							%s
						</td>
					</tr>

					<tr>
						<td style="padding:0 36px 8px;">
							<p style="margin:0;color:#94a3b8;font-size:12px;line-height:1.6;">
								Qualquer dúvida sobre o %s, estamos à disposição.
							</p>
						</td>
					</tr>

					<tr>
						<td style="padding:24px 36px 32px;border-top:1px solid #f1f5f9;margin-top:8px;">
							<p style="margin:20px 0 0;color:#cbd5e1;font-size:11px;">RinoSeller · documento gerado automaticamente</p>
						</td>
					</tr>

				</table>
			</td>
		</tr>
	</table>
</body>
</html>`,
		client.Name,
		noun,
		issuedBy,
		noun,
	)

	return subject, html
}
