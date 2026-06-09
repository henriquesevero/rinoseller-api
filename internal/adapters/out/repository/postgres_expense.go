package repository

import (
	"context"
	"errors"

	"rinoseller-api/internal/core/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresExpenseRepository struct {
	db *pgxpool.Pool
}

func NewPostgresExpenseRepository(db *pgxpool.Pool) *PostgresExpenseRepository {
	return &PostgresExpenseRepository{db: db}
}

func (r *PostgresExpenseRepository) Save(e *domain.Expense) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO expenses (id, user_id, description, supplier, amount, payment_method, due_date, status, paid_at, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, e.ID, nullStr(e.UserID), e.Description, e.Supplier, e.Amount, e.PaymentMethod,
		nullTime(e.DueDate), e.Status, nullTime(e.PaidAt), e.Notes, e.CreatedAt)
	return err
}

func (r *PostgresExpenseRepository) FindAll(userID string) ([]domain.Expense, error) {
	var rows interface {
		Next() bool
		Scan(...any) error
		Close()
	}
	var err error

	if userID == "" {
		rows, err = r.db.Query(context.Background(), `
			SELECT id, COALESCE(user_id,''), description, supplier, amount, payment_method, due_date, status, paid_at, notes, created_at
			FROM expenses ORDER BY created_at DESC
		`)
	} else {
		rows, err = r.db.Query(context.Background(), `
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
		var e domain.Expense
		if err := rows.Scan(&e.ID, &e.UserID, &e.Description, &e.Supplier, &e.Amount, &e.PaymentMethod,
			&e.DueDate, &e.Status, &e.PaidAt, &e.Notes, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	if result == nil {
		return []domain.Expense{}, nil
	}
	return result, nil
}

func (r *PostgresExpenseRepository) FindByID(id string) (*domain.Expense, error) {
	var e domain.Expense
	err := r.db.QueryRow(context.Background(), `
		SELECT id, COALESCE(user_id,''), description, supplier, amount, payment_method, due_date, status, paid_at, notes, created_at
		FROM expenses WHERE id = $1
	`, id).Scan(&e.ID, &e.UserID, &e.Description, &e.Supplier, &e.Amount, &e.PaymentMethod,
		&e.DueDate, &e.Status, &e.PaidAt, &e.Notes, &e.CreatedAt)
	if err != nil {
		return nil, errors.New("conta a pagar não encontrada")
	}
	return &e, nil
}

func (r *PostgresExpenseRepository) Update(e *domain.Expense) error {
	tag, err := r.db.Exec(context.Background(), `
		UPDATE expenses SET description=$1, supplier=$2, amount=$3, payment_method=$4, due_date=$5,
		                    status=$6, paid_at=$7, notes=$8
		WHERE id=$9
	`, e.Description, e.Supplier, e.Amount, e.PaymentMethod, nullTime(e.DueDate),
		e.Status, nullTime(e.PaidAt), e.Notes, e.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("conta a pagar não encontrada")
	}
	return nil
}

func (r *PostgresExpenseRepository) Delete(id string) error {
	tag, err := r.db.Exec(context.Background(), `DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("conta a pagar não encontrada")
	}
	return nil
}
