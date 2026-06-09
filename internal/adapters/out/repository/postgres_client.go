package repository

import (
	"context"
	"errors"
	"time"

	"rinoseller-api/internal/core/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClientRepository struct {
	db *pgxpool.Pool
}

func NewPostgresClientRepository(db *pgxpool.Pool) *PostgresClientRepository {
	return &PostgresClientRepository{db: db}
}

func (r *PostgresClientRepository) Save(c *domain.Client) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO clients (id, user_id, name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, c.ID, nullStr(c.UserID), c.Name, c.Phone, c.Email, c.Notes, c.CpfCnpj, c.Address, c.Debt, c.DebtLimit,
		c.PaymentCycleDays, c.PaymentCycleAmount, c.CreatedAt)
	return err
}

func (r *PostgresClientRepository) FindAll(userID string) ([]domain.Client, error) {
	var rows interface{ Next() bool; Scan(...any) error; Close() }
	var err error

	if userID == "" {
		rows, err = r.db.Query(context.Background(), `
			SELECT id, COALESCE(user_id,''), name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at
			FROM clients ORDER BY name
		`)
	} else {
		rows, err = r.db.Query(context.Background(), `
			SELECT id, COALESCE(user_id,''), name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at
			FROM clients WHERE user_id = $1 ORDER BY name
		`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Client
	for rows.Next() {
		var c domain.Client
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Notes, &c.CpfCnpj, &c.Address, &c.Debt, &c.DebtLimit,
			&c.PaymentCycleDays, &c.PaymentCycleAmount, &c.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	if result == nil {
		return []domain.Client{}, nil
	}
	return result, nil
}

func (r *PostgresClientRepository) FindByID(id string) (*domain.Client, error) {
	var c domain.Client
	err := r.db.QueryRow(context.Background(), `
		SELECT id, COALESCE(user_id,''), name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at
		FROM clients WHERE id = $1
	`, id).Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Notes, &c.CpfCnpj, &c.Address, &c.Debt, &c.DebtLimit,
		&c.PaymentCycleDays, &c.PaymentCycleAmount, &c.CreatedAt)
	if err != nil {
		return nil, errors.New("cliente não encontrado")
	}
	return &c, nil
}

func (r *PostgresClientRepository) FindByPhone(phone string) (*domain.Client, error) {
	var c domain.Client
	err := r.db.QueryRow(context.Background(), `
		SELECT id, COALESCE(user_id,''), name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at
		FROM clients WHERE phone = $1 LIMIT 1
	`, phone).Scan(&c.ID, &c.UserID, &c.Name, &c.Phone, &c.Email, &c.Notes, &c.CpfCnpj, &c.Address, &c.Debt, &c.DebtLimit,
		&c.PaymentCycleDays, &c.PaymentCycleAmount, &c.CreatedAt)
	if err != nil {
		return nil, errors.New("cliente não encontrado")
	}
	return &c, nil
}

func (r *PostgresClientRepository) Update(c *domain.Client) error {
	tag, err := r.db.Exec(context.Background(), `
		UPDATE clients SET name=$1, phone=$2, email=$3, notes=$4, cpf_cnpj=$5, address=$6, debt=$7, debt_limit=$8,
		                   payment_cycle_days=$9, payment_cycle_amount=$10
		WHERE id=$11
	`, c.Name, c.Phone, c.Email, c.Notes, c.CpfCnpj, c.Address, c.Debt, c.DebtLimit,
		c.PaymentCycleDays, c.PaymentCycleAmount, c.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("cliente não encontrado")
	}
	return nil
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *PostgresClientRepository) Delete(id string) error {
	tag, err := r.db.Exec(context.Background(), `DELETE FROM clients WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("cliente não encontrado")
	}
	return nil
}

func (r *PostgresClientRepository) SavePayment(p *domain.ClientPayment) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO client_payments (id, client_id, user_id, amount, notes, count_as_revenue, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, p.ID, p.ClientID, nullStr(p.UserID), p.Amount, p.Notes, p.CountAsRevenue, p.CreatedAt)
	return err
}

func (r *PostgresClientRepository) FindPaymentsByClientID(clientID string) ([]domain.ClientPayment, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, client_id, COALESCE(user_id,''), amount, notes, COALESCE(count_as_revenue, false), created_at
		FROM client_payments WHERE client_id = $1 ORDER BY created_at DESC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.ClientPayment
	for rows.Next() {
		var p domain.ClientPayment
		if err := rows.Scan(&p.ID, &p.ClientID, &p.UserID, &p.Amount, &p.Notes, &p.CountAsRevenue, &p.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if result == nil {
		return []domain.ClientPayment{}, nil
	}
	return result, nil
}

func (r *PostgresClientRepository) DeletePaymentsByClientID(clientID string) error {
	_, err := r.db.Exec(context.Background(), `DELETE FROM client_payments WHERE client_id = $1`, clientID)
	return err
}

func (r *PostgresClientRepository) FindAllPayments(userID string) ([]domain.ClientPayment, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, client_id, COALESCE(user_id,''), amount, notes, COALESCE(count_as_revenue, false), created_at
		FROM client_payments WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.ClientPayment
	for rows.Next() {
		var p domain.ClientPayment
		if err := rows.Scan(&p.ID, &p.ClientID, &p.UserID, &p.Amount, &p.Notes, &p.CountAsRevenue, &p.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if result == nil {
		return []domain.ClientPayment{}, nil
	}
	return result, nil
}

// ── helpers para timestamps opcionais ────────────────────────────────────────

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
