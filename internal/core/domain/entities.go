package domain

import (
	"fmt"
	"time"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleSeller UserRole = "seller"
)

func (r UserRole) IsAdmin() bool {
	return r == RoleAdmin
}

func (r UserRole) Valid() bool {
	return r == RoleAdmin || r == RoleSeller
}

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type KitItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

type Product struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id,omitempty"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	Brand         string    `json:"brand"`
	Price         Money     `json:"price"`
	CostPrice     Money     `json:"cost_price"`
	StockQuantity int       `json:"stock_quantity"`
	IsKit         bool      `json:"is_kit"`
	KitItems      []KitItem `json:"kit_items,omitempty"`
}

type StockMovement struct {
	ProductID string
	Quantity  int
}

func (p Product) ResolveStockMovements(quantity int) []StockMovement {
	if !p.IsKit {
		return []StockMovement{{ProductID: p.ID, Quantity: quantity}}
	}
	movements := make([]StockMovement, 0, len(p.KitItems))
	for _, ki := range p.KitItems {
		movements = append(movements, StockMovement{ProductID: ki.ProductID, Quantity: ki.Quantity * quantity})
	}
	return movements
}

type OrderStatus string

const (
	OrderStatusReady     OrderStatus = "Pronta-Entrega"
	OrderStatusBackorder OrderStatus = "Encomenda"
)

type OrderItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	Price       Money  `json:"price"`
}

type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id,omitempty"`
	ClientName  string      `json:"client_name"`
	ClientPhone string      `json:"client_phone"`
	Items       []OrderItem `json:"items"`
	Total       Money       `json:"total"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
}

type QuoteStatus string

const (
	QuoteStatusAwaitingApproval QuoteStatus = "Aguardando Aprovação"
	QuoteStatusApproved         QuoteStatus = "Aprovado"
	QuoteStatusInvoiced         QuoteStatus = "Faturado"
	QuoteStatusInvoicedGradual  QuoteStatus = "Faturado Gradual"
	QuoteStatusDelivered        QuoteStatus = "Entregue"
	QuoteStatusCancelled        QuoteStatus = "Cancelado"
)

type PaymentType string

const PaymentTypeCustom PaymentType = "Personalizado"

type QuoteItem struct {
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   Money     `json:"unit_price"`
	Subtotal    Money     `json:"subtotal"`
	KitItems    []KitItem `json:"kit_items,omitempty"`
}

type Quote struct {
	ID           string      `json:"id"`
	UserID       string      `json:"user_id,omitempty"`
	ClientID     string      `json:"client_id"`
	ClientName   string      `json:"client_name"`
	Items        []QuoteItem `json:"items"`
	Total        Money       `json:"total"`
	Status       QuoteStatus `json:"status"`
	Notes        string      `json:"notes"`
	PaymentType  PaymentType `json:"payment_type"`
	Installments int         `json:"installments"`
	CreatedAt    time.Time   `json:"created_at"`
	ApprovedAt   *time.Time  `json:"approved_at,omitempty"`
	InvoicedAt   *time.Time  `json:"invoiced_at,omitempty"`
	DeliveredAt  *time.Time  `json:"delivered_at,omitempty"`
}

func (q *Quote) Approve() error {
	if q.Status != QuoteStatusAwaitingApproval {
		return fmt.Errorf("apenas orçamentos aguardando aprovação podem ser aprovados: %w", ErrValidation)
	}
	now := time.Now()
	q.Status = QuoteStatusApproved
	q.ApprovedAt = &now
	return nil
}

func (q *Quote) Cancel() error {
	if q.Status != QuoteStatusAwaitingApproval {
		return fmt.Errorf("apenas orçamentos aguardando aprovação podem ser cancelados: %w", ErrValidation)
	}
	q.Status = QuoteStatusCancelled
	return nil
}

func (q *Quote) Invoice() error {
	if q.Status != QuoteStatusApproved {
		return fmt.Errorf("apenas orçamentos aprovados podem ser faturados: %w", ErrValidation)
	}
	now := time.Now()
	if q.IsGradualPayment() {
		q.Status = QuoteStatusInvoicedGradual
	} else {
		q.Status = QuoteStatusInvoiced
	}
	q.InvoicedAt = &now
	return nil
}

func (q *Quote) Deliver() error {
	if q.Status != QuoteStatusInvoiced && q.Status != QuoteStatusInvoicedGradual {
		return fmt.Errorf("apenas pedidos faturados podem ser marcados como entregues: %w", ErrValidation)
	}
	now := time.Now()
	q.Status = QuoteStatusDelivered
	q.DeliveredAt = &now
	return nil
}

func (q *Quote) IsGradualPayment() bool {
	return q.PaymentType == PaymentTypeCustom
}

type Client struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id,omitempty"`
	Name               string    `json:"name"`
	Phone              Phone     `json:"phone"`
	Email              string    `json:"email"`
	Notes              string    `json:"notes"`
	CpfCnpj            CpfCnpj   `json:"cpf_cnpj"`
	Address            string    `json:"address"`
	Debt               Money     `json:"debt"`
	DebtLimit          Money     `json:"debt_limit"`
	PaymentCycleDays   int       `json:"payment_cycle_days"`
	PaymentCycleAmount Money     `json:"payment_cycle_amount"`
	CreatedAt          time.Time `json:"created_at"`
}

func (c Client) EnsureCanReceiveNewOrder() error {
	if c.DebtLimit.IsPositive() && c.Debt.GreaterOrEqual(c.DebtLimit) {
		return fmt.Errorf("cliente \"%s\" atingiu o limite de dívida (%s de %s) e não pode receber novos pedidos ou orçamentos: %w",
			c.Name, c.Debt, c.DebtLimit, ErrValidation)
	}
	return nil
}

func (c *Client) ApplyPayment(amount Money) {
	c.Debt = c.Debt.Sub(amount).AtLeastZero()
}

func (c *Client) AddDebt(amount Money) {
	c.Debt = c.Debt.Add(amount)
}

type ClientPayment struct {
	ID             string    `json:"id"`
	ClientID       string    `json:"client_id"`
	UserID         string    `json:"user_id,omitempty"`
	Amount         Money     `json:"amount"`
	Notes          string    `json:"notes"`
	CountAsRevenue bool      `json:"count_as_revenue"`
	CreatedAt      time.Time `json:"created_at"`
}

type ContributionType string

const (
	ContributionAporte   ContributionType = "aporte"
	ContributionRetirada ContributionType = "retirada"
)

type CapitalContribution struct {
	ID          string           `json:"id"`
	UserID      string           `json:"user_id,omitempty"`
	Description string           `json:"description"`
	Amount      Money            `json:"amount"`
	Type        ContributionType `json:"type"`
	Hidden      bool             `json:"hidden"`
	CreatedAt   time.Time        `json:"created_at"`
}

type BrandCatalog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	BrandName string    `json:"brand_name"`
	DriveURL  string    `json:"drive_url"`
	CreatedAt time.Time `json:"created_at"`
}

type ExpenseStatus string

const (
	ExpenseStatusPending ExpenseStatus = "Pendente"
	ExpenseStatusPaid    ExpenseStatus = "Pago"
)

type Expense struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id,omitempty"`
	Description   string        `json:"description"`
	Supplier      string        `json:"supplier"`
	Amount        Money         `json:"amount"`
	PaymentMethod string        `json:"payment_method"`
	DueDate       *time.Time    `json:"due_date,omitempty"`
	Status        ExpenseStatus `json:"status"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	Notes         string        `json:"notes"`
	CreatedAt     time.Time     `json:"created_at"`
}

func (e *Expense) Pay() error {
	if e.Status == ExpenseStatusPaid {
		return fmt.Errorf("esta conta já está paga: %w", ErrConflict)
	}
	now := time.Now()
	e.Status = ExpenseStatusPaid
	e.PaidAt = &now
	return nil
}
