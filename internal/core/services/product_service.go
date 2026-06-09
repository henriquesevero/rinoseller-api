package services

import (
	"fmt"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

type ProductService struct {
	repo ports.ProductRepository
}

func NewProductService(repo ports.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) ListProducts(userID string) ([]domain.Product, error) {
	return s.repo.FindAll(userID)
}

func (s *ProductService) CreateProduct(p *domain.Product) error {
	p.ID = uuid.New().String()

	if p.IsKit {
		for i, item := range p.KitItems {
			component, err := s.repo.FindByID(item.ProductID)
			if err != nil {
				return fmt.Errorf("componente do kit não encontrado: %w", err)
			}
			p.KitItems[i].ProductName = component.Name
		}
	} else {
		p.KitItems = nil
	}

	return s.repo.Save(p)
}

func (s *ProductService) UpdatePrice(id string, price float64) error {
	return s.repo.UpdatePrice(id, price)
}

func (s *ProductService) UpdateStock(id string, newStock int) error {
	return s.repo.UpdateStock(id, newStock)
}

func (s *ProductService) DeleteProduct(id string) error {
	return s.repo.Delete(id)
}
