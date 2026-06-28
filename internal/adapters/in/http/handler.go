package httphandler

import "rinoseller-api/internal/core/ports"

type Handler struct {
	authUC         ports.AuthUseCase
	userUC         ports.UserUseCase
	productUC      ports.ProductUseCase
	orderUC        ports.OrderUseCase
	clientUC       ports.ClientUseCase
	quoteUC        ports.QuoteUseCase
	expenseUC      ports.ExpenseUseCase
	capitalUC      ports.CapitalContributionUseCase
	brandCatalogUC ports.BrandCatalogUseCase
}

func NewHandler(
	authUC ports.AuthUseCase,
	userUC ports.UserUseCase,
	productUC ports.ProductUseCase,
	orderUC ports.OrderUseCase,
	clientUC ports.ClientUseCase,
	quoteUC ports.QuoteUseCase,
	expenseUC ports.ExpenseUseCase,
	capitalUC ports.CapitalContributionUseCase,
	brandCatalogUC ports.BrandCatalogUseCase,
) *Handler {
	return &Handler{authUC: authUC, userUC: userUC, productUC: productUC, orderUC: orderUC, clientUC: clientUC, quoteUC: quoteUC, expenseUC: expenseUC, capitalUC: capitalUC, brandCatalogUC: brandCatalogUC}
}

type messageResponse struct {
	Message string `json:"message" example:"operação realizada com sucesso"`
}

type errorResponse struct {
	Error string `json:"error" example:"mensagem de erro"`
	Code  string `json:"code,omitempty" example:"email_not_verified"`
}
