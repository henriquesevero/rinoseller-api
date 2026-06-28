package services

import (
	"context"
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

func (s *QuoteService) ListQuotes(ctx context.Context, userID string) ([]domain.Quote, error) {
	return s.quoteRepo.FindAll(ctx, userID)
}

func (s *QuoteService) CreateQuote(ctx context.Context, q *domain.Quote) error {
	if q.ClientID == "" || len(q.Items) == 0 {
		return fmt.Errorf("cliente e itens são obrigatórios: %w", domain.ErrValidation)
	}

	client, err := s.clientRepo.FindByID(ctx, q.ClientID)
	if err != nil {
		return fmt.Errorf("cliente não encontrado: %w", domain.ErrValidation)
	}
	if err := client.EnsureCanReceiveNewOrder(); err != nil {
		return err
	}
	q.ClientName = client.Name

	var total domain.Money
	for i, item := range q.Items {
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("produto %s não encontrado: %w", item.ProductID, domain.ErrValidation)
		}
		q.Items[i].ProductName = product.Name
		q.Items[i].UnitPrice = product.Price
		q.Items[i].Subtotal = product.Price.MulQty(item.Quantity)
		if product.IsKit {
			q.Items[i].KitItems = product.KitItems
		}
		total = total.Add(q.Items[i].Subtotal)
	}

	q.ID = uuid.New().String()
	q.Total = total
	q.Status = domain.QuoteStatusAwaitingApproval
	q.CreatedAt = time.Now()

	return s.quoteRepo.Save(ctx, q)
}

func (s *QuoteService) GetQuote(ctx context.Context, id string) (*domain.Quote, error) {
	return s.quoteRepo.FindByID(ctx, id)
}

func (s *QuoteService) ApproveQuote(ctx context.Context, id string) (*domain.Quote, error) {
	q, err := s.quoteRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := q.Approve(); err != nil {
		return nil, err
	}
	return q, s.quoteRepo.Update(ctx, q)
}

func (s *QuoteService) DeliverQuote(ctx context.Context, id string) (*domain.Quote, error) {
	q, err := s.quoteRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := q.Deliver(); err != nil {
		return nil, err
	}

	s.deductStockBestEffort(ctx, q)

	return q, s.quoteRepo.Update(ctx, q)
}

func (s *QuoteService) deductStockBestEffort(ctx context.Context, q *domain.Quote) {
	for _, item := range q.Items {
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			continue
		}
		_, _ = applyStockMovements(ctx, s.productRepo, product, item.Quantity)
	}
}

func (s *QuoteService) InvoiceQuote(ctx context.Context, id string) (*domain.Quote, error) {
	q, err := s.quoteRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := q.Invoice(); err != nil {
		return nil, err
	}

	if q.IsGradualPayment() {
		s.chargeGradualPaymentDebt(ctx, q)
	}

	return q, s.quoteRepo.Update(ctx, q)
}

func (s *QuoteService) chargeGradualPaymentDebt(ctx context.Context, q *domain.Quote) {
	if q.ClientID == "" {
		return
	}
	client, err := s.clientRepo.FindByID(ctx, q.ClientID)
	if err != nil {
		return
	}
	client.AddDebt(q.Total)
	_ = s.clientRepo.Update(ctx, client)
}

func (s *QuoteService) CancelQuote(ctx context.Context, id string) (*domain.Quote, error) {
	q, err := s.quoteRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := q.Cancel(); err != nil {
		return nil, err
	}
	return q, s.quoteRepo.Update(ctx, q)
}

func (s *QuoteService) GetClientQuotes(ctx context.Context, clientID string) ([]domain.Quote, error) {
	return s.quoteRepo.FindByClientID(ctx, clientID)
}

func (s *QuoteService) DeleteQuote(ctx context.Context, id string) error {
	return s.quoteRepo.Delete(ctx, id)
}

func (s *QuoteService) ClearClientQuotes(ctx context.Context, clientID string) error {
	return s.quoteRepo.DeleteByClientID(ctx, clientID)
}
