package services

import (
	"context"
	"fmt"
	"time"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

type CapitalContributionService struct {
	repo ports.CapitalContributionRepository
}

func NewCapitalContributionService(repo ports.CapitalContributionRepository) *CapitalContributionService {
	return &CapitalContributionService{repo: repo}
}

func (s *CapitalContributionService) ListContributions(ctx context.Context, userID string) ([]domain.CapitalContribution, error) {
	return s.repo.FindAll(ctx, userID)
}

func (s *CapitalContributionService) AddContribution(ctx context.Context, e *domain.CapitalContribution) error {
	if !e.Amount.IsPositive() {
		return fmt.Errorf("valor inválido: %w", domain.ErrValidation)
	}
	e.ID = uuid.New().String()
	if e.Type != domain.ContributionRetirada {
		e.Type = domain.ContributionAporte
	}
	e.CreatedAt = time.Now()
	return s.repo.Save(ctx, e)
}

func (s *CapitalContributionService) DeleteContribution(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
