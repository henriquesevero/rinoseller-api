package services

import (
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

// CreateOrder aplica as regras de negócio: subtrai estoque e determina o status do pedido.
// Se qualquer item não tiver estoque suficiente, o pedido inteiro é marcado como "Encomenda".
func (s *OrderService) CreateOrder(order *domain.Order) error {
	if order.ClientPhone != "" {
		if client, err := s.clientRepo.FindByPhone(order.ClientPhone); err == nil {
			if client.DebtLimit > 0 && client.Debt >= client.DebtLimit {
				return fmt.Errorf("cliente \"%s\" atingiu o limite de dívida (%.2f de %.2f) e não pode receber novos pedidos ou orçamentos", client.Name, client.Debt, client.DebtLimit)
			}
		}
	}

	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.Status = "Pronta-Entrega"

	var total float64
	for i, item := range order.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("produto %s não encontrado: %w", item.ProductID, err)
		}

		order.Items[i].ProductName = product.Name
		order.Items[i].Price = product.Price
		total += product.Price * float64(item.Quantity)

		newStock := product.StockQuantity - item.Quantity
		if newStock < 0 {
			order.Status = "Encomenda"
			newStock = 0
		}
		if err := s.productRepo.UpdateStock(product.ID, newStock); err != nil {
			return fmt.Errorf("erro ao atualizar estoque do produto %s: %w", product.ID, err)
		}
	}

	order.Total = total
	return s.orderRepo.Save(order)
}

func (s *OrderService) ListOrders(userID string) ([]domain.Order, error) {
	return s.orderRepo.FindAll(userID)
}

func (s *OrderService) DeleteOrder(id string) error {
	return s.orderRepo.Delete(id)
}
