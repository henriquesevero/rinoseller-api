package payment

import (
	"context"
	"log"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

// MockAsaasGateway simula a integração com o Asaas (https://www.asaas.com): aprova qualquer
// cobrança sem se comunicar com nenhum serviço externo. Quando a integração real acontecer,
// um novo adapter implementando ports.PaymentGateway substitui este aqui sem tocar no resto
// da aplicação (use cases e handlers dependem só da interface).
type MockAsaasGateway struct{}

func NewMockAsaasGateway() *MockAsaasGateway {
	return &MockAsaasGateway{}
}

func (g *MockAsaasGateway) Charge(_ context.Context, customerEmail, customerName string, plan domain.Plan, _ ports.CardDetails) (*ports.CheckoutResult, error) {
	log.Printf("[asaas-mock] cobrança simulada aprovada — cliente=%s (%s) plano=%s", customerName, customerEmail, plan)
	return &ports.CheckoutResult{ID: uuid.New().String(), Status: "paid"}, nil
}
