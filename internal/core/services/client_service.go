package services

import (
	"context"
	"fmt"
	"time"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

type ClientService struct {
	clientRepo  ports.ClientRepository
	orderRepo   ports.OrderRepository
	paymentRepo ports.ClientPaymentRepository
	quoteRepo   ports.QuoteRepository
}

func NewClientService(clientRepo ports.ClientRepository, orderRepo ports.OrderRepository, paymentRepo ports.ClientPaymentRepository, quoteRepo ports.QuoteRepository) *ClientService {
	return &ClientService{clientRepo: clientRepo, orderRepo: orderRepo, paymentRepo: paymentRepo, quoteRepo: quoteRepo}
}

func (s *ClientService) ListClients(ctx context.Context, userID string) ([]domain.Client, error) {
	return s.clientRepo.FindAll(ctx, userID)
}

func (s *ClientService) CreateClient(ctx context.Context, c *domain.Client) error {
	if c.Name == "" {
		return fmt.Errorf("nome é obrigatório: %w", domain.ErrValidation)
	}
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	return s.clientRepo.Save(ctx, c)
}

func (s *ClientService) GetClient(ctx context.Context, id string) (*domain.Client, error) {
	return s.clientRepo.FindByID(ctx, id)
}

func (s *ClientService) UpdateClient(ctx context.Context, c *domain.Client) error {
	return s.clientRepo.Update(ctx, c)
}

func (s *ClientService) RegisterPayment(ctx context.Context, id, userID string, amount domain.Money, notes string, countAsRevenue bool) error {
	if !amount.IsPositive() {
		return fmt.Errorf("valor inválido: %w", domain.ErrValidation)
	}
	client, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	client.ApplyPayment(amount)
	if err := s.clientRepo.Update(ctx, client); err != nil {
		return err
	}
	return s.paymentRepo.Save(ctx, &domain.ClientPayment{
		ID:             uuid.New().String(),
		ClientID:       id,
		UserID:         userID,
		Amount:         amount,
		Notes:          notes,
		CountAsRevenue: countAsRevenue,
		CreatedAt:      time.Now(),
	})
}

func (s *ClientService) AddDebt(ctx context.Context, id string, amount domain.Money) error {
	if !amount.IsPositive() {
		return fmt.Errorf("valor inválido: %w", domain.ErrValidation)
	}
	client, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	client.AddDebt(amount)
	return s.clientRepo.Update(ctx, client)
}

func (s *ClientService) GetClientOrders(ctx context.Context, id string) ([]domain.Order, error) {
	client, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.orderRepo.FindByClientMatch(ctx, client.Phone.String(), client.Name)
}

func (s *ClientService) GetClientPayments(ctx context.Context, clientID string) ([]domain.ClientPayment, error) {
	return s.paymentRepo.FindByClientID(ctx, clientID)
}

func (s *ClientService) GetAllPayments(ctx context.Context, userID string) ([]domain.ClientPayment, error) {
	return s.paymentRepo.FindAll(ctx, userID)
}

func (s *ClientService) DeleteClient(ctx context.Context, id string) error {
	if err := s.quoteRepo.DeleteByClientID(ctx, id); err != nil {
		return err
	}
	if err := s.ClearClientOrders(ctx, id); err != nil {
		return err
	}
	if err := s.ClearPaymentHistory(ctx, id); err != nil {
		return err
	}
	return s.clientRepo.Delete(ctx, id)
}

func (s *ClientService) ClearPaymentHistory(ctx context.Context, id string) error {
	return s.paymentRepo.DeleteByClientID(ctx, id)
}

func (s *ClientService) ClearClientOrders(ctx context.Context, id string) error {
	orders, err := s.GetClientOrders(ctx, id)
	if err != nil {
		return err
	}
	for _, o := range orders {
		if err := s.orderRepo.Delete(ctx, o.ID); err != nil {
			return err
		}
	}
	return nil
}
