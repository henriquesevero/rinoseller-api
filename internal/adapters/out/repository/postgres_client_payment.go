package repository

import (
	"context"

	"rinoseller-api/internal/core/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClientPaymentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresClientPaymentRepository(db *pgxpool.Pool) *PostgresClientPaymentRepository {
	return &PostgresClientPaymentRepository{db: db}
}

func (r *PostgresClientPaymentRepository) Save(ctx context.Context, p *domain.ClientPayment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO client_payments (id, client_id, user_id, amount, notes, count_as_revenue, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, p.ID, p.ClientID, nullStr(p.UserID), p.Amount.Float64(), p.Notes, p.CountAsRevenue, p.CreatedAt)
	return err
}

func (r *PostgresClientPaymentRepository) FindByClientID(ctx context.Context, clientID string) ([]domain.ClientPayment, error) {
	return r.query(ctx, `
		SELECT id, client_id, COALESCE(user_id,''), amount, notes, COALESCE(count_as_revenue, false), created_at
		FROM client_payments WHERE client_id = $1 ORDER BY created_at DESC
	`, clientID)
}

func (r *PostgresClientPaymentRepository) FindAll(ctx context.Context, userID string) ([]domain.ClientPayment, error) {
	return r.query(ctx, `
		SELECT id, client_id, COALESCE(user_id,''), amount, notes, COALESCE(count_as_revenue, false), created_at
		FROM client_payments WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
}

func (r *PostgresClientPaymentRepository) DeleteByClientID(ctx context.Context, clientID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM client_payments WHERE client_id = $1`, clientID)
	return err
}

func (r *PostgresClientPaymentRepository) query(ctx context.Context, sql string, arg string) ([]domain.ClientPayment, error) {
	rows, err := r.db.Query(ctx, sql, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.ClientPayment
	for rows.Next() {
		var p domain.ClientPayment
		var amount float64
		if err := rows.Scan(&p.ID, &p.ClientID, &p.UserID, &amount, &p.Notes, &p.CountAsRevenue, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Amount = domain.NewMoneyFromFloat(amount)
		result = append(result, p)
	}
	if result == nil {
		return []domain.ClientPayment{}, nil
	}
	return result, nil
}
