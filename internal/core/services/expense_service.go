package services

import (
	"context"
	"fmt"
	"time"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"

	"github.com/google/uuid"
)

type ExpenseService struct {
	expenseRepo ports.ExpenseRepository
}

func NewExpenseService(expenseRepo ports.ExpenseRepository) *ExpenseService {
	return &ExpenseService{expenseRepo: expenseRepo}
}

func (s *ExpenseService) ListExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	return s.expenseRepo.FindAll(ctx, userID)
}

func (s *ExpenseService) CreateExpense(ctx context.Context, e *domain.Expense) error {
	if e.Description == "" || !e.Amount.IsPositive() {
		return fmt.Errorf("descrição e valor são obrigatórios: %w", domain.ErrValidation)
	}
	e.ID = uuid.New().String()
	e.Status = domain.ExpenseStatusPending
	e.PaidAt = nil
	e.CreatedAt = time.Now()
	return s.expenseRepo.Save(ctx, e)
}

func (s *ExpenseService) PayExpense(ctx context.Context, id string) (*domain.Expense, error) {
	e, err := s.expenseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := e.Pay(); err != nil {
		return nil, err
	}
	return e, s.expenseRepo.Update(ctx, e)
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, id string) error {
	return s.expenseRepo.Delete(ctx, id)
}
