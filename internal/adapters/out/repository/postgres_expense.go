package repository

import (
	"context"
	"fmt"

	"rinoseller-api/internal/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresExpenseRepository struct {
	db *pgxpool.Pool
}

func NewPostgresExpenseRepository(db *pgxpool.Pool) *PostgresExpenseRepository {
	return &PostgresExpenseRepository{db: db}
}

func (r *PostgresExpenseRepository) Save(ctx context.Context, e *domain.Expense) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO expenses (id, user_id, description, supplier, amount, payment_method, due_date, status, paid_at, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, e.ID, nullStr(e.UserID), e.Description, e.Supplier, e.Amount.Float64(), e.PaymentMethod,
		nullTime(e.DueDate), string(e.Status), nullTime(e.PaidAt), e.Notes, e.CreatedAt)
	return err
}

func (r *PostgresExpenseRepository) FindAll(ctx context.Context, userID string) ([]domain.Expense, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if userID == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), description, supplier, amount, payment_method, due_date, status, paid_at, notes, created_at
			FROM expenses ORDER BY created_at DESC
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), description, supplier, amount, payment_method, due_date, status, paid_at, notes, created_at
			FROM expenses WHERE user_id = $1 ORDER BY created_at DESC
		`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Expense
	for rows.Next() {
		e, err := scanExpense(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	if result == nil {
		return []domain.Expense{}, nil
	}
	return result, nil
}

func (r *PostgresExpenseRepository) FindByID(ctx context.Context, id string) (*domain.Expense, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(user_id,''), description, supplier, amount, payment_method, due_date, status, paid_at, notes, created_at
		FROM expenses WHERE id = $1
	`, id)
	e, err := scanExpense(row)
	if err != nil {
		return nil, fmt.Errorf("conta a pagar não encontrada: %w", domain.ErrNotFound)
	}
	return &e, nil
}

func (r *PostgresExpenseRepository) Update(ctx context.Context, e *domain.Expense) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE expenses SET description=$1, supplier=$2, amount=$3, payment_method=$4, due_date=$5,
		                    status=$6, paid_at=$7, notes=$8
		WHERE id=$9
	`, e.Description, e.Supplier, e.Amount.Float64(), e.PaymentMethod, nullTime(e.DueDate),
		string(e.Status), nullTime(e.PaidAt), e.Notes, e.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("conta a pagar não encontrada: %w", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresExpenseRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("conta a pagar não encontrada: %w", domain.ErrNotFound)
	}
	return nil
}

func scanExpense(row rowScanner) (domain.Expense, error) {
	var e domain.Expense
	var amount float64
	var status string
	err := row.Scan(&e.ID, &e.UserID, &e.Description, &e.Supplier, &amount, &e.PaymentMethod,
		&e.DueDate, &status, &e.PaidAt, &e.Notes, &e.CreatedAt)
	if err != nil {
		return domain.Expense{}, err
	}
	e.Amount = domain.NewMoneyFromFloat(amount)
	e.Status = domain.ExpenseStatus(status)
	return e, nil
}
