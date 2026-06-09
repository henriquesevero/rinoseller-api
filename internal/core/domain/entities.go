package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // "admin" | "seller"
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

type QuoteItem struct {
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Subtotal    float64   `json:"subtotal"`
	KitItems    []KitItem `json:"kit_items,omitempty"`
}

type Quote struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id,omitempty"`
	ClientID    string      `json:"client_id"`
	ClientName  string      `json:"client_name"`
	Items       []QuoteItem `json:"items"`
	Total       float64     `json:"total"`
	Status      string      `json:"status"`
	Notes       string      `json:"notes"`
	PaymentType string      `json:"payment_type"`
	Installments int        `json:"installments"`
	CreatedAt   time.Time   `json:"created_at"`
	ApprovedAt  *time.Time  `json:"approved_at,omitempty"`
	InvoicedAt  *time.Time  `json:"invoiced_at,omitempty"`
}

type Client struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id,omitempty"`
	Name               string    `json:"name"`
	Phone              string    `json:"phone"`
	Email              string    `json:"email"`
	Notes              string    `json:"notes"`
	CpfCnpj            string    `json:"cpf_cnpj"`
	Address            string    `json:"address"`
	Debt               float64   `json:"debt"`
	DebtLimit          float64   `json:"debt_limit"`
	PaymentCycleDays   int       `json:"payment_cycle_days"`
	PaymentCycleAmount float64   `json:"payment_cycle_amount"`
	CreatedAt          time.Time `json:"created_at"`
}

type ClientPayment struct {
	ID             string    `json:"id"`
	ClientID       string    `json:"client_id"`
	UserID         string    `json:"user_id,omitempty"`
	Amount         float64   `json:"amount"`
	Notes          string    `json:"notes"`
	CountAsRevenue bool      `json:"count_as_revenue"`
	CreatedAt      time.Time `json:"created_at"`
}

// CapitalContribution representa uma movimentação manual de capital: um aporte (entrada)
// ou uma retirada (saída) de dinheiro feita pelo dono do negócio.
type CapitalContribution struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id,omitempty"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"` // "aporte" ou "retirada"
	Hidden      bool      `json:"hidden"` // não aparece no histórico (ex: ajuste de zerar capital)
	CreatedAt   time.Time `json:"created_at"`
}

// Expense representa uma conta a pagar (compra de produtos de distribuidores, despesas etc).
type Expense struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id,omitempty"`
	Description   string     `json:"description"`
	Supplier      string     `json:"supplier"`
	Amount        float64    `json:"amount"`
	PaymentMethod string     `json:"payment_method"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	Status        string     `json:"status"` // "Pendente" | "Pago"
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	Notes         string     `json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Product struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id,omitempty"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	Price         float64   `json:"price"`
	CostPrice     float64   `json:"cost_price"`
	StockQuantity int       `json:"stock_quantity"`
	IsKit         bool      `json:"is_kit"`
	KitItems      []KitItem `json:"kit_items,omitempty"`
}

type KitItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id,omitempty"`
	ClientName  string      `json:"client_name"`
	ClientPhone string      `json:"client_phone"`
	Items       []OrderItem `json:"items"`
	Total       float64     `json:"total"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
}
