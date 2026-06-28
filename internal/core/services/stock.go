package services

import (
	"context"
	"fmt"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"
)

func applyStockMovements(ctx context.Context, productRepo ports.ProductRepository, product *domain.Product, quantity int) (insufficient bool, err error) {
	for _, mv := range product.ResolveStockMovements(quantity) {
		target := product
		if mv.ProductID != product.ID {
			component, err := productRepo.FindByID(ctx, mv.ProductID)
			if err != nil {
				return false, fmt.Errorf("componente %s não encontrado: %w", mv.ProductID, domain.ErrValidation)
			}
			target = component
		}
		newStock := target.StockQuantity - mv.Quantity
		if newStock < 0 {
			insufficient = true
			newStock = 0
		}
		if err := productRepo.UpdateStock(ctx, mv.ProductID, newStock); err != nil {
			return false, fmt.Errorf("erro ao atualizar estoque do produto %s: %w", mv.ProductID, err)
		}
	}
	return insufficient, nil
}
