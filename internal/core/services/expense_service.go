package services

import (
	"errors"
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

func (s *ExpenseService) ListExpenses(userID string) ([]domain.Expense, error) {
	return s.expenseRepo.FindAll(userID)
}

func (s *ExpenseService) CreateExpense(e *domain.Expense) error {
	e.ID = uuid.New().String()
	e.Status = "Pendente"
	e.PaidAt = nil
	e.CreatedAt = time.Now()
	return s.expenseRepo.Save(e)
}

func (s *ExpenseService) PayExpense(id string) (*domain.Expense, error) {
	e, err := s.expenseRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if e.Status == "Pago" {
		return nil, errors.New("esta conta já está paga")
	}
	now := time.Now()
	e.Status = "Pago"
	e.PaidAt = &now
	return e, s.expenseRepo.Update(e)
}

func (s *ExpenseService) DeleteExpense(id string) error {
	return s.expenseRepo.Delete(id)
}
