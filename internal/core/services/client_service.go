package services

import (
	"fmt"
	"strings"
	"time"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

type ClientService struct {
	clientRepo ports.ClientRepository
	orderRepo  ports.OrderRepository
}

func NewClientService(clientRepo ports.ClientRepository, orderRepo ports.OrderRepository) *ClientService {
	return &ClientService{clientRepo: clientRepo, orderRepo: orderRepo}
}

func (s *ClientService) ListClients(userID string) ([]domain.Client, error) {
	return s.clientRepo.FindAll(userID)
}

func (s *ClientService) CreateClient(c *domain.Client) error {
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	return s.clientRepo.Save(c)
}

func (s *ClientService) GetClient(id string) (*domain.Client, error) {
	return s.clientRepo.FindByID(id)
}

func (s *ClientService) UpdateClient(c *domain.Client) error {
	return s.clientRepo.Update(c)
}

func (s *ClientService) RegisterPayment(id, userID string, amount float64, notes string, countAsRevenue bool) error {
	client, err := s.clientRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("cliente não encontrado: %w", err)
	}
	client.Debt -= amount
	if client.Debt < 0 {
		client.Debt = 0
	}
	if err := s.clientRepo.Update(client); err != nil {
		return err
	}
	return s.clientRepo.SavePayment(&domain.ClientPayment{
		ID:             uuid.New().String(),
		ClientID:       id,
		UserID:         userID,
		Amount:         amount,
		Notes:          notes,
		CountAsRevenue: countAsRevenue,
		CreatedAt:      time.Now(),
	})
}

func (s *ClientService) AddDebt(id string, amount float64) error {
	client, err := s.clientRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("cliente não encontrado: %w", err)
	}
	client.Debt += amount
	return s.clientRepo.Update(client)
}

func (s *ClientService) GetClientOrders(id string) ([]domain.Order, error) {
	client, err := s.clientRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("cliente não encontrado: %w", err)
	}
	allOrders, err := s.orderRepo.FindAll("")
	if err != nil {
		return nil, err
	}
	result := make([]domain.Order, 0)
	for _, o := range allOrders {
		if client.Phone != "" && o.ClientPhone == client.Phone {
			result = append(result, o)
		} else if client.Phone == "" && strings.EqualFold(o.ClientName, client.Name) {
			result = append(result, o)
		}
	}
	return result, nil
}

func (s *ClientService) GetClientPayments(clientID string) ([]domain.ClientPayment, error) {
	return s.clientRepo.FindPaymentsByClientID(clientID)
}

func (s *ClientService) GetAllPayments(userID string) ([]domain.ClientPayment, error) {
	return s.clientRepo.FindAllPayments(userID)
}

func (s *ClientService) DeleteClient(id string) error {
	return s.clientRepo.Delete(id)
}

func (s *ClientService) ClearPaymentHistory(id string) error {
	return s.clientRepo.DeletePaymentsByClientID(id)
}

// ClearClientOrders remove os pedidos do catálogo associados ao cliente (mesmo critério de GetClientOrders).
func (s *ClientService) ClearClientOrders(id string) error {
	orders, err := s.GetClientOrders(id)
	if err != nil {
		return err
	}
	for _, o := range orders {
		if err := s.orderRepo.Delete(o.ID); err != nil {
			return err
		}
	}
	return nil
}
