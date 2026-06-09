package repository

import (
	"context"
	"errors"

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

func (r *PostgresOrderRepository) Save(o *domain.Order) error {
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, user_id, client_name, client_phone, total, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, o.ID, nullStr(o.UserID), o.ClientName, o.ClientPhone, o.Total, o.Status, o.CreatedAt)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_id, product_name, quantity, price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), o.ID, item.ProductID, item.ProductName, item.Quantity, item.Price)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// FindAll retorna pedidos do usuário; userID="" retorna todos (admin).
func (r *PostgresOrderRepository) FindAll(userID string) ([]domain.Order, error) {
	ctx := context.Background()

	var rows interface {
		Next() bool
		Scan(...any) error
		Close()
	}
	var err error

	if userID == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), client_name, client_phone, total, status, created_at
			FROM orders ORDER BY created_at DESC
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), client_name, client_phone, total, status, created_at
			FROM orders WHERE user_id = $1 ORDER BY created_at DESC
		`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[string]*domain.Order)
	var ids []string

	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ClientName, &o.ClientPhone, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
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
		if err := itemRows.Scan(&orderID, &item.ProductID, &item.ProductName, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
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

func (r *PostgresOrderRepository) Delete(id string) error {
	tag, err := r.db.Exec(context.Background(), `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("pedido não encontrado")
	}
	return nil
}
