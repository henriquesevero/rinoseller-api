package services

import (
	"context"
	"fmt"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"
)

type SubscriptionService struct {
	userRepo ports.UserRepository
	gateway  ports.PaymentGateway
}

func NewSubscriptionService(userRepo ports.UserRepository, gateway ports.PaymentGateway) *SubscriptionService {
	return &SubscriptionService{userRepo: userRepo, gateway: gateway}
}

func (s *SubscriptionService) Checkout(ctx context.Context, email string, plan domain.Plan, card ports.CardDetails) error {
	if !plan.Selectable() || plan == domain.PlanTrial {
		return fmt.Errorf("plano inválido: %w", domain.ErrValidation)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	result, err := s.gateway.Charge(ctx, user.Email, user.Name, plan, card)
	if err != nil {
		return err
	}
	if result.Status != "paid" {
		return domain.ErrPaymentFailed
	}

	return s.userRepo.ActivateSubscription(ctx, user.ID, plan)
}
