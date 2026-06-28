package repository

import (
	"context"
	"fmt"

	"rinoseller-api/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(ctx context.Context, o *domain.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, user_id, client_name, client_phone, total, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, o.ID, nullStr(o.UserID), o.ClientName, o.ClientPhone, o.Total.Float64(), string(o.Status), o.CreatedAt)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_id, product_name, quantity, price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), o.ID, item.ProductID, item.ProductName, item.Quantity, item.Price.Float64())
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) FindAll(ctx context.Context, userID string) ([]domain.Order, error) {
	if userID == "" {
		return r.query(ctx, `
			SELECT id, COALESCE(user_id,''), client_name, client_phone, total, status, created_at
			FROM orders ORDER BY created_at DESC
		`)
	}
	return r.query(ctx, `
		SELECT id, COALESCE(user_id,''), client_name, client_phone, total, status, created_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
}

func (r *PostgresOrderRepository) FindByClientMatch(ctx context.Context, phone, name string) ([]domain.Order, error) {
	return r.query(ctx, `
		SELECT id, COALESCE(user_id,''), client_name, client_phone, total, status, created_at
		FROM orders
		WHERE ($1 <> '' AND client_phone = $1) OR ($1 = '' AND LOWER(client_name) = LOWER($2))
		ORDER BY created_at DESC
	`, phone, name)
}

func (r *PostgresOrderRepository) query(ctx context.Context, sql string, args ...any) ([]domain.Order, error) {
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[string]*domain.Order)
	var ids []string

	for rows.Next() {
		var o domain.Order
		var total float64
		var status string
		if err := rows.Scan(&o.ID, &o.UserID, &o.ClientName, &o.ClientPhone, &total, &status, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Total = domain.NewMoneyFromFloat(total)
		o.Status = domain.OrderStatus(status)
		o.Items = []domain.OrderItem{}
		orderMap[o.ID] = &o
		ids = append(ids, o.ID)
	}

	if len(ids) == 0 {
		return []domain.Order{}, nil
	}

	itemRows, err := r.db.Query(ctx, `
		SELECT order_id, product_id, product_name, quantity, price
		FROM order_items WHERE order_id = ANY($1)
	`, ids)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var orderID string
		var item domain.OrderItem
		var price float64
		if err := itemRows.Scan(&orderID, &item.ProductID, &item.ProductName, &item.Quantity, &price); err != nil {
			return nil, err
		}
		item.Price = domain.NewMoneyFromFloat(price)
		if o, ok := orderMap[orderID]; ok {
			o.Items = append(o.Items, item)
		}
	}

	result := make([]domain.Order, 0, len(ids))
	for _, id := range ids {
		result = append(result, *orderMap[id])
	}
	return result, nil
}

func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("pedido não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}
