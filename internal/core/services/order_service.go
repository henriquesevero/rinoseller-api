package services

import (
	"context"
	"fmt"
	"time"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

type OrderService struct {
	orderRepo   ports.OrderRepository
	productRepo ports.ProductRepository
	clientRepo  ports.ClientRepository
}

func NewOrderService(orderRepo ports.OrderRepository, productRepo ports.ProductRepository, clientRepo ports.ClientRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		clientRepo:  clientRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *domain.Order) error {
	if order.ClientPhone != "" {
		if client, err := s.clientRepo.FindByPhone(ctx, order.ClientPhone); err == nil {
			if err := client.EnsureCanReceiveNewOrder(); err != nil {
				return err
			}
		}
	}

	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.Status = domain.OrderStatusReady

	var total domain.Money
	for i, item := range order.Items {
		product, err := s.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("produto %s não encontrado: %w", item.ProductID, domain.ErrValidation)
		}

		order.Items[i].ProductName = product.Name
		order.Items[i].Price = product.Price
		total = total.Add(product.Price.MulQty(item.Quantity))

		insufficient, err := applyStockMovements(ctx, s.productRepo, product, item.Quantity)
		if err != nil {
			return err
		}
		if insufficient {
			order.Status = domain.OrderStatusBackorder
		}
	}

	order.Total = total
	return s.orderRepo.Save(ctx, order)
}

func (s *OrderService) ListOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	return s.orderRepo.FindAll(ctx, userID)
}

func (s *OrderService) DeleteOrder(ctx context.Context, id string) error {
	return s.orderRepo.Delete(ctx, id)
}
