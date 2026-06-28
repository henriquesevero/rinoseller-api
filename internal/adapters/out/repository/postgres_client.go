package repository

import (
	"context"
	"fmt"

	"rinoseller-api/internal/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClientRepository struct {
	db *pgxpool.Pool
}

func NewPostgresClientRepository(db *pgxpool.Pool) *PostgresClientRepository {
	return &PostgresClientRepository{db: db}
}

func (r *PostgresClientRepository) Save(ctx context.Context, c *domain.Client) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO clients (id, user_id, name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, c.ID, nullStr(c.UserID), c.Name, c.Phone.String(), c.Email, c.Notes, c.CpfCnpj.String(), c.Address,
		c.Debt.Float64(), c.DebtLimit.Float64(), c.PaymentCycleDays, c.PaymentCycleAmount.Float64(), c.CreatedAt)
	return err
}

func (r *PostgresClientRepository) FindAll(ctx context.Context, userID string) ([]domain.Client, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if userID == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at
			FROM clients ORDER BY name
		`)
	} else {
		rows, err = r.db.Query(ctx, `
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
		c, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	if result == nil {
		return []domain.Client{}, nil
	}
	return result, nil
}

func (r *PostgresClientRepository) FindByID(ctx context.Context, id string) (*domain.Client, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(user_id,''), name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at
		FROM clients WHERE id = $1
	`, id)
	c, err := scanClient(row)
	if err != nil {
		return nil, fmt.Errorf("cliente não encontrado: %w", domain.ErrNotFound)
	}
	return &c, nil
}

func (r *PostgresClientRepository) FindByPhone(ctx context.Context, phone string) (*domain.Client, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(user_id,''), name, phone, email, notes, cpf_cnpj, address, debt, debt_limit, payment_cycle_days, payment_cycle_amount, created_at
		FROM clients WHERE phone = $1 LIMIT 1
	`, phone)
	c, err := scanClient(row)
	if err != nil {
		return nil, fmt.Errorf("cliente não encontrado: %w", domain.ErrNotFound)
	}
	return &c, nil
}

func (r *PostgresClientRepository) Update(ctx context.Context, c *domain.Client) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE clients SET name=$1, phone=$2, email=$3, notes=$4, cpf_cnpj=$5, address=$6, debt=$7, debt_limit=$8,
		                   payment_cycle_days=$9, payment_cycle_amount=$10
		WHERE id=$11
	`, c.Name, c.Phone.String(), c.Email, c.Notes, c.CpfCnpj.String(), c.Address, c.Debt.Float64(), c.DebtLimit.Float64(),
		c.PaymentCycleDays, c.PaymentCycleAmount.Float64(), c.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("cliente não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresClientRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM clients WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("cliente não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanClient(row rowScanner) (domain.Client, error) {
	var c domain.Client
	var phone, cpfCnpj string
	var debt, debtLimit, paymentCycleAmount float64
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &phone, &c.Email, &c.Notes, &cpfCnpj, &c.Address,
		&debt, &debtLimit, &c.PaymentCycleDays, &paymentCycleAmount, &c.CreatedAt)
	if err != nil {
		return domain.Client{}, err
	}
	c.Phone = domain.NewPhoneUnchecked(phone)
	c.CpfCnpj = domain.NewCpfCnpjUnchecked(cpfCnpj)
	c.Debt = domain.NewMoneyFromFloat(debt)
	c.DebtLimit = domain.NewMoneyFromFloat(debtLimit)
	c.PaymentCycleAmount = domain.NewMoneyFromFloat(paymentCycleAmount)
	return c, nil
}
