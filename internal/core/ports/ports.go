package ports

import (
	"time"

	"rinoseller-api/internal/core/domain"
)

// --- Output Ports (Driven) ---

type UserRepository interface {
	Save(u *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id string) (*domain.User, error)
	FindAll() ([]domain.User, error)
	Update(u *domain.User) error
	HasAnyAdmin() (bool, error)
	SetResetToken(id, token string, expires time.Time) error
	FindByResetToken(token string) (*domain.User, error)
	ClearResetToken(id string) error
}

type ProductRepository interface {
	Save(p *domain.Product) error
	FindAll(userID string) ([]domain.Product, error)
	FindByID(id string) (*domain.Product, error)
	UpdateStock(id string, newStock int) error
	UpdatePrice(id string, price float64) error
	Delete(id string) error
}

type OrderRepository interface {
	Save(o *domain.Order) error
	FindAll(userID string) ([]domain.Order, error)
	FindByClientMatch(phone, name string) ([]domain.Order, error)
	Delete(id string) error
}

type ClientRepository interface {
	Save(c *domain.Client) error
	FindAll(userID string) ([]domain.Client, error)
	FindByID(id string) (*domain.Client, error)
	FindByPhone(phone string) (*domain.Client, error)
	Update(c *domain.Client) error
	Delete(id string) error
	SavePayment(p *domain.ClientPayment) error
	FindPaymentsByClientID(clientID string) ([]domain.ClientPayment, error)
	FindAllPayments(userID string) ([]domain.ClientPayment, error)
	DeletePaymentsByClientID(clientID string) error
}

type CapitalContributionRepository interface {
	Save(e *domain.CapitalContribution) error
	FindAll(userID string) ([]domain.CapitalContribution, error)
	Delete(id string) error
}

type ExpenseRepository interface {
	Save(e *domain.Expense) error
	FindAll(userID string) ([]domain.Expense, error)
	FindByID(id string) (*domain.Expense, error)
	Update(e *domain.Expense) error
	Delete(id string) error
}

type QuoteRepository interface {
	Save(q *domain.Quote) error
	FindAll(userID string) ([]domain.Quote, error)
	FindByID(id string) (*domain.Quote, error)
	Update(q *domain.Quote) error
	FindByClientID(clientID string) ([]domain.Quote, error)
	Delete(id string) error
	DeleteByClientID(clientID string) error
}


// --- Input Ports / Use Cases (Driver) ---

type AuthUseCase interface {
	Login(email, password string) (token string, user *domain.User, err error)
	ValidateToken(token string) (*domain.User, error)
	ForgotPassword(email string) (resetToken string, err error)
	ResetPassword(token, newPassword string) error
}

type UserUseCase interface {
	CreateUser(name, email, password, role string) (*domain.User, error)
	ListUsers() ([]domain.User, error)
	GetUser(id string) (*domain.User, error)
	UpdateUser(id, name, email string, active bool) (*domain.User, error)
	EnsureDefaultAdmin() error
}

type ProductUseCase interface {
	ListProducts(userID string) ([]domain.Product, error)
	CreateProduct(p *domain.Product) error
	UpdatePrice(id string, price float64) error
	UpdateStock(id string, newStock int) error
	DeleteProduct(id string) error
}

type OrderUseCase interface {
	CreateOrder(order *domain.Order) error
	ListOrders(userID string) ([]domain.Order, error)
	DeleteOrder(id string) error
}

type ClientUseCase interface {
	ListClients(userID string) ([]domain.Client, error)
	CreateClient(c *domain.Client) error
	GetClient(id string) (*domain.Client, error)
	UpdateClient(c *domain.Client) error
	RegisterPayment(id, userID string, amount float64, notes string, countAsRevenue bool) error
	AddDebt(id string, amount float64) error
	GetClientOrders(id string) ([]domain.Order, error)
	GetClientPayments(clientID string) ([]domain.ClientPayment, error)
	GetAllPayments(userID string) ([]domain.ClientPayment, error)
	DeleteClient(id string) error
	ClearPaymentHistory(id string) error
	ClearClientOrders(id string) error
}

type CapitalContributionUseCase interface {
	ListContributions(userID string) ([]domain.CapitalContribution, error)
	AddContribution(e *domain.CapitalContribution) error
	DeleteContribution(id string) error
}

type ExpenseUseCase interface {
	ListExpenses(userID string) ([]domain.Expense, error)
	CreateExpense(e *domain.Expense) error
	PayExpense(id string) (*domain.Expense, error)
	DeleteExpense(id string) error
}

type QuoteUseCase interface {
	ListQuotes(userID string) ([]domain.Quote, error)
	CreateQuote(q *domain.Quote) error
	GetQuote(id string) (*domain.Quote, error)
	ApproveQuote(id string) (*domain.Quote, error)
	DeliverQuote(id string) (*domain.Quote, error)
	InvoiceQuote(id string) (*domain.Quote, error)
	CancelQuote(id string) (*domain.Quote, error)
	GetClientQuotes(clientID string) ([]domain.Quote, error)
	DeleteQuote(id string) error
	ClearClientQuotes(clientID string) error
}

