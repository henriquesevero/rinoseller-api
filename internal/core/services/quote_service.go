package services

import (
	"errors"
	"fmt"
	"time"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

type QuoteService struct {
	quoteRepo   ports.QuoteRepository
	productRepo ports.ProductRepository
	clientRepo  ports.ClientRepository
}

func NewQuoteService(quoteRepo ports.QuoteRepository, productRepo ports.ProductRepository, clientRepo ports.ClientRepository) *QuoteService {
	return &QuoteService{
		quoteRepo:   quoteRepo,
		productRepo: productRepo,
		clientRepo:  clientRepo,
	}
}

func (s *QuoteService) ListQuotes(userID string) ([]domain.Quote, error) {
	return s.quoteRepo.FindAll(userID)
}

func (s *QuoteService) CreateQuote(q *domain.Quote) error {
	client, err := s.clientRepo.FindByID(q.ClientID)
	if err != nil {
		return fmt.Errorf("cliente não encontrado: %w", err)
	}
	if client.DebtLimit > 0 && client.Debt >= client.DebtLimit {
		return fmt.Errorf("cliente \"%s\" atingiu o limite de dívida (%.2f de %.2f) e não pode receber novos pedidos ou orçamentos", client.Name, client.Debt, client.DebtLimit)
	}
	q.ClientName = client.Name

	var total float64
	for i, item := range q.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("produto %s não encontrado: %w", item.ProductID, err)
		}
		q.Items[i].ProductName = product.Name
		q.Items[i].UnitPrice = product.Price
		q.Items[i].Subtotal = product.Price * float64(item.Quantity)
		if product.IsKit {
			q.Items[i].KitItems = product.KitItems
		}
		total += q.Items[i].Subtotal
	}

	q.ID = uuid.New().String()
	q.Total = total
	q.Status = "Aguardando Aprovação"
	q.CreatedAt = time.Now()

	return s.quoteRepo.Save(q)
}

func (s *QuoteService) GetQuote(id string) (*domain.Quote, error) {
	return s.quoteRepo.FindByID(id)
}

func (s *QuoteService) ApproveQuote(id string) (*domain.Quote, error) {
	q, err := s.quoteRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if q.Status != "Aguardando Aprovação" {
		return nil, errors.New("apenas orçamentos aguardando aprovação podem ser aprovados")
	}
	now := time.Now()
	q.Status = "Aprovado"
	q.ApprovedAt = &now
	return q, s.quoteRepo.Update(q)
}

// DeliverQuote marca como Entregue e desconta estoque — é o status final do fluxo,
// para todos os tipos de pagamento. O pedido só pode ser entregue após o faturamento;
// para pedidos com pagamento Personalizado/gradual, a entrega confirma a quitação da
// dívida que foi lançada no momento do faturamento.
func (s *QuoteService) DeliverQuote(id string) (*domain.Quote, error) {
	q, err := s.quoteRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if q.Status != "Faturado" && q.Status != "Faturado Gradual" {
		return nil, errors.New("apenas pedidos faturados podem ser marcados como entregues")
	}
	now := time.Now()
	q.Status = "Entregue"
	q.InvoicedAt = &now

	for _, item := range q.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			continue
		}

		if product.IsKit {
			for _, kitItem := range product.KitItems {
				component, err := s.productRepo.FindByID(kitItem.ProductID)
				if err != nil {
					continue
				}
				newStock := component.StockQuantity - (kitItem.Quantity * item.Quantity)
				if newStock < 0 {
					newStock = 0
				}
				_ = s.productRepo.UpdateStock(component.ID, newStock)
			}
			continue
		}

		newStock := product.StockQuantity - item.Quantity
		if newStock < 0 {
			newStock = 0
		}
		_ = s.productRepo.UpdateStock(product.ID, newStock)
	}

	return q, s.quoteRepo.Update(q)
}

// InvoiceQuote marca como Faturado a partir de "Aprovado", para todos os tipos de pagamento.
// Pedidos com pagamento Personalizado/gradual lançam a dívida no cliente neste momento — o
// pagamento ocorre depois, de forma gradual, e é confirmado quando o pedido é entregue.
func (s *QuoteService) InvoiceQuote(id string) (*domain.Quote, error) {
	q, err := s.quoteRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if q.Status != "Aprovado" {
		return nil, errors.New("apenas pedidos aprovados podem ser marcados como faturados")
	}
	now := time.Now()
	if q.PaymentType == "Personalizado" {
		q.Status = "Faturado Gradual"
		if q.ClientID != "" {
			if client, err := s.clientRepo.FindByID(q.ClientID); err == nil {
				client.Debt += q.Total
				_ = s.clientRepo.Update(client)
			}
		}
	} else {
		q.Status = "Faturado"
	}
	q.InvoicedAt = &now

	return q, s.quoteRepo.Update(q)
}

func (s *QuoteService) CancelQuote(id string) (*domain.Quote, error) {
	q, err := s.quoteRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if q.Status != "Aguardando Aprovação" {
		return nil, errors.New("apenas orçamentos aguardando aprovação podem ser cancelados")
	}
	q.Status = "Cancelado"
	return q, s.quoteRepo.Update(q)
}

func (s *QuoteService) GetClientQuotes(clientID string) ([]domain.Quote, error) {
	return s.quoteRepo.FindByClientID(clientID)
}

func (s *QuoteService) DeleteQuote(id string) error {
	return s.quoteRepo.Delete(id)
}

func (s *QuoteService) ClearClientQuotes(clientID string) error {
	return s.quoteRepo.DeleteByClientID(clientID)
}
