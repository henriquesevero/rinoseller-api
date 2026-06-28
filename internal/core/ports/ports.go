package ports

import (
	"context"
	"time"

	"rinoseller-api/internal/core/domain"
)

type UserRepository interface {
	Save(ctx context.Context, u *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindAll(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, u *domain.User) error
	HasAnyAdmin(ctx context.Context) (bool, error)
	SetResetToken(ctx context.Context, id, token string, expires time.Time) error
	FindByResetToken(ctx context.Context, token string) (*domain.User, error)
	ClearResetToken(ctx context.Context, id string) error
}

type ProductRepository interface {
	Save(ctx context.Context, p *domain.Product) error
	FindAll(ctx context.Context, userID string) ([]domain.Product, error)
	FindByID(ctx context.Context, id string) (*domain.Product, error)
	UpdateStock(ctx context.Context, id string, newStock int) error
	UpdatePrice(ctx context.Context, id string, price domain.Money) error
	Delete(ctx context.Context, id string) error
}

type OrderRepository interface {
	Save(ctx context.Context, o *domain.Order) error
	FindAll(ctx context.Context, userID string) ([]domain.Order, error)
	FindByClientMatch(ctx context.Context, phone, name string) ([]domain.Order, error)
	Delete(ctx context.Context, id string) error
}

type ClientRepository interface {
	Save(ctx context.Context, c *domain.Client) error
	FindAll(ctx context.Context, userID string) ([]domain.Client, error)
	FindByID(ctx context.Context, id string) (*domain.Client, error)
	FindByPhone(ctx context.Context, phone string) (*domain.Client, error)
	Update(ctx context.Context, c *domain.Client) error
	Delete(ctx context.Context, id string) error
}

type ClientPaymentRepository interface {
	Save(ctx context.Context, p *domain.ClientPayment) error
	FindByClientID(ctx context.Context, clientID string) ([]domain.ClientPayment, error)
	FindAll(ctx context.Context, userID string) ([]domain.ClientPayment, error)
	DeleteByClientID(ctx context.Context, clientID string) error
}

type CapitalContributionRepository interface {
	Save(ctx context.Context, e *domain.CapitalContribution) error
	FindAll(ctx context.Context, userID string) ([]domain.CapitalContribution, error)
	Delete(ctx context.Context, id string) error
}

type ExpenseRepository interface {
	Save(ctx context.Context, e *domain.Expense) error
	FindAll(ctx context.Context, userID string) ([]domain.Expense, error)
	FindByID(ctx context.Context, id string) (*domain.Expense, error)
	Update(ctx context.Context, e *domain.Expense) error
	Delete(ctx context.Context, id string) error
}

type QuoteRepository interface {
	Save(ctx context.Context, q *domain.Quote) error
	FindAll(ctx context.Context, userID string) ([]domain.Quote, error)
	FindByID(ctx context.Context, id string) (*domain.Quote, error)
	Update(ctx context.Context, q *domain.Quote) error
	FindByClientID(ctx context.Context, clientID string) ([]domain.Quote, error)
	Delete(ctx context.Context, id string) error
	DeleteByClientID(ctx context.Context, clientID string) error
}

type AuthUseCase interface {
	Login(ctx context.Context, email, password string) (token string, user *domain.User, err error)
	ValidateToken(ctx context.Context, token string) (*domain.User, error)
	ForgotPassword(ctx context.Context, email string) (resetToken string, err error)
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type UserUseCase interface {
	CreateUser(ctx context.Context, name, email, password string, role domain.UserRole) (*domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, id, name, email string, active bool) (*domain.User, error)
	EnsureDefaultAdmin(ctx context.Context) (generatedPassword string, err error)
}

type ProductUseCase interface {
	ListProducts(ctx context.Context, userID string) ([]domain.Product, error)
	CreateProduct(ctx context.Context, p *domain.Product) error
	UpdatePrice(ctx context.Context, id string, price domain.Money) error
	UpdateStock(ctx context.Context, id string, newStock int) error
	DeleteProduct(ctx context.Context, id string) error
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	ListOrders(ctx context.Context, userID string) ([]domain.Order, error)
	DeleteOrder(ctx context.Context, id string) error
}

type ClientUseCase interface {
	ListClients(ctx context.Context, userID string) ([]domain.Client, error)
	CreateClient(ctx context.Context, c *domain.Client) error
	GetClient(ctx context.Context, id string) (*domain.Client, error)
	UpdateClient(ctx context.Context, c *domain.Client) error
	RegisterPayment(ctx context.Context, id, userID string, amount domain.Money, notes string, countAsRevenue bool) error
	AddDebt(ctx context.Context, id string, amount domain.Money) error
	GetClientOrders(ctx context.Context, id string) ([]domain.Order, error)
	GetClientPayments(ctx context.Context, clientID string) ([]domain.ClientPayment, error)
	GetAllPayments(ctx context.Context, userID string) ([]domain.ClientPayment, error)
	DeleteClient(ctx context.Context, id string) error
	ClearPaymentHistory(ctx context.Context, id string) error
	ClearClientOrders(ctx context.Context, id string) error
}

type CapitalContributionUseCase interface {
	ListContributions(ctx context.Context, userID string) ([]domain.CapitalContribution, error)
	AddContribution(ctx context.Context, e *domain.CapitalContribution) error
	DeleteContribution(ctx context.Context, id string) error
}

type ExpenseUseCase interface {
	ListExpenses(ctx context.Context, userID string) ([]domain.Expense, error)
	CreateExpense(ctx context.Context, e *domain.Expense) error
	PayExpense(ctx context.Context, id string) (*domain.Expense, error)
	DeleteExpense(ctx context.Context, id string) error
}

type QuoteUseCase interface {
	ListQuotes(ctx context.Context, userID string) ([]domain.Quote, error)
	CreateQuote(ctx context.Context, q *domain.Quote) error
	GetQuote(ctx context.Context, id string) (*domain.Quote, error)
	ApproveQuote(ctx context.Context, id string) (*domain.Quote, error)
	DeliverQuote(ctx context.Context, id string) (*domain.Quote, error)
	InvoiceQuote(ctx context.Context, id string) (*domain.Quote, error)
	CancelQuote(ctx context.Context, id string) (*domain.Quote, error)
	GetClientQuotes(ctx context.Context, clientID string) ([]domain.Quote, error)
	DeleteQuote(ctx context.Context, id string) error
	ClearClientQuotes(ctx context.Context, clientID string) error
}
