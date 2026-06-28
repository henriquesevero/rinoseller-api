package services

import (
	"context"
	"fmt"
	"time"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

type BrandCatalogService struct {
	repo ports.BrandCatalogRepository
}

func NewBrandCatalogService(repo ports.BrandCatalogRepository) *BrandCatalogService {
	return &BrandCatalogService{repo: repo}
}

func (s *BrandCatalogService) ListCatalogs(ctx context.Context, userID string) ([]domain.BrandCatalog, error) {
	return s.repo.FindAll(ctx, userID)
}

func (s *BrandCatalogService) AddCatalog(ctx context.Context, c *domain.BrandCatalog) error {
	if c.BrandName == "" || c.DriveURL == "" {
		return fmt.Errorf("nome da marca e link são obrigatórios: %w", domain.ErrValidation)
	}
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	return s.repo.Save(ctx, c)
}

func (s *BrandCatalogService) GetCatalog(ctx context.Context, id string) (*domain.BrandCatalog, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *BrandCatalogService) DeleteCatalog(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
