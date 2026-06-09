package services

import (
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

func (s *CapitalContributionService) ListContributions(userID string) ([]domain.CapitalContribution, error) {
	return s.repo.FindAll(userID)
}

func (s *CapitalContributionService) AddContribution(e *domain.CapitalContribution) error {
	e.ID = uuid.New().String()
	if e.Type != "retirada" {
		e.Type = "aporte"
	}
	e.CreatedAt = time.Now()
	return s.repo.Save(e)
}

func (s *CapitalContributionService) DeleteContribution(id string) error {
	return s.repo.Delete(id)
}
