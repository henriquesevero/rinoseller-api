package services

import (
	"context"
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

func (s *ProductService) ListProducts(ctx context.Context, userID string) ([]domain.Product, error) {
	return s.repo.FindAll(ctx, userID)
}

func (s *ProductService) CreateProduct(ctx context.Context, p *domain.Product) error {
	p.ID = uuid.New().String()

	if p.IsKit {
		for i, item := range p.KitItems {
			component, err := s.repo.FindByID(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("componente do kit não encontrado: %w", domain.ErrValidation)
			}
			p.KitItems[i].ProductName = component.Name
		}
	} else {
		p.KitItems = nil
	}

	return s.repo.Save(ctx, p)
}

func (s *ProductService) UpdatePrice(ctx context.Context, id string, price domain.Money) error {
	if !price.IsPositive() {
		return fmt.Errorf("preço deve ser maior que zero: %w", domain.ErrValidation)
	}
	return s.repo.UpdatePrice(ctx, id, price)
}

func (s *ProductService) UpdateStock(ctx context.Context, id string, newStock int) error {
	if newStock < 0 {
		return fmt.Errorf("quantidade não pode ser negativa: %w", domain.ErrValidation)
	}
	return s.repo.UpdateStock(ctx, id, newStock)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
