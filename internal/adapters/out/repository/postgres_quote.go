package repository

import (
	"context"
	"encoding/json"
	"errors"

	"rinoseller-api/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func marshalKitItems(items []domain.KitItem) string {
	if len(items) == 0 {
		return `[]`
	}
	b, err := json.Marshal(items)
	if err != nil {
		return `[]`
	}
	return string(b)
}

func unmarshalKitItems(raw []byte) []domain.KitItem {
	var items []domain.KitItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

type PostgresQuoteRepository struct {
	db *pgxpool.Pool
}

func NewPostgresQuoteRepository(db *pgxpool.Pool) *PostgresQuoteRepository {
	return &PostgresQuoteRepository{db: db}
}

func (r *PostgresQuoteRepository) Save(q *domain.Quote) error {
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO quotes (id, user_id, client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, q.ID, nullStr(q.UserID), q.ClientID, q.ClientName, q.Total, q.Status, q.Notes,
		q.PaymentType, q.Installments, q.CreatedAt, q.ApprovedAt, q.InvoicedAt)
	if err != nil {
		return err
	}

	for _, item := range q.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO quote_items (id, quote_id, product_id, product_name, quantity, unit_price, subtotal, kit_items)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, uuid.New().String(), q.ID, item.ProductID, item.ProductName, item.Quantity, item.UnitPrice, item.Subtotal, marshalKitItems(item.KitItems))
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// FindAll retorna orçamentos do usuário; userID="" retorna todos (admin).
func (r *PostgresQuoteRepository) FindAll(userID string) ([]domain.Quote, error) {
	if userID == "" {
		return r.query(context.Background(), `
			SELECT id, COALESCE(user_id,''), client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at
			FROM quotes ORDER BY created_at DESC
		`)
	}
	return r.query(context.Background(), `
		SELECT id, COALESCE(user_id,''), client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at
		FROM quotes WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
}

func (r *PostgresQuoteRepository) FindByClientID(clientID string) ([]domain.Quote, error) {
	return r.query(context.Background(), `
		SELECT id, COALESCE(user_id,''), client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at
		FROM quotes WHERE client_id = $1 ORDER BY created_at DESC
	`, clientID)
}

func (r *PostgresQuoteRepository) query(ctx context.Context, sql string, args ...any) ([]domain.Quote, error) {
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quoteMap := make(map[string]*domain.Quote)
	var ids []string

	for rows.Next() {
		var q domain.Quote
		if err := rows.Scan(&q.ID, &q.UserID, &q.ClientID, &q.ClientName, &q.Total, &q.Status, &q.Notes,
			&q.PaymentType, &q.Installments, &q.CreatedAt, &q.ApprovedAt, &q.InvoicedAt); err != nil {
			return nil, err
		}
		q.Items = []domain.QuoteItem{}
		quoteMap[q.ID] = &q
		ids = append(ids, q.ID)
	}

	if len(ids) == 0 {
		return []domain.Quote{}, nil
	}

	itemRows, err := r.db.Query(ctx, `
		SELECT quote_id, product_id, product_name, quantity, unit_price, subtotal, kit_items
		FROM quote_items WHERE quote_id = ANY($1)
	`, ids)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var quoteID string
		var item domain.QuoteItem
		var kitItemsRaw string
		if err := itemRows.Scan(&quoteID, &item.ProductID, &item.ProductName, &item.Quantity, &item.UnitPrice, &item.Subtotal, &kitItemsRaw); err != nil {
			return nil, err
		}
		item.KitItems = unmarshalKitItems([]byte(kitItemsRaw))
		if q, ok := quoteMap[quoteID]; ok {
			q.Items = append(q.Items, item)
		}
	}

	result := make([]domain.Quote, 0, len(ids))
	for _, id := range ids {
		result = append(result, *quoteMap[id])
	}
	return result, nil
}

func (r *PostgresQuoteRepository) FindByID(id string) (*domain.Quote, error) {
	ctx := context.Background()
	var q domain.Quote
	err := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(user_id,''), client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at
		FROM quotes WHERE id = $1
	`, id).Scan(&q.ID, &q.UserID, &q.ClientID, &q.ClientName, &q.Total, &q.Status, &q.Notes,
		&q.PaymentType, &q.Installments, &q.CreatedAt, &q.ApprovedAt, &q.InvoicedAt)
	if err != nil {
		return nil, errors.New("orçamento não encontrado")
	}

	itemRows, err := r.db.Query(ctx, `
		SELECT product_id, product_name, quantity, unit_price, subtotal, kit_items
		FROM quote_items WHERE quote_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	q.Items = []domain.QuoteItem{}
	for itemRows.Next() {
		var item domain.QuoteItem
		var kitItemsRaw string
		if err := itemRows.Scan(&item.ProductID, &item.ProductName, &item.Quantity, &item.UnitPrice, &item.Subtotal, &kitItemsRaw); err != nil {
			return nil, err
		}
		item.KitItems = unmarshalKitItems([]byte(kitItemsRaw))
		q.Items = append(q.Items, item)
	}

	return &q, nil
}

func (r *PostgresQuoteRepository) Update(q *domain.Quote) error {
	tag, err := r.db.Exec(context.Background(), `
		UPDATE quotes SET status=$1, approved_at=$2, invoiced_at=$3 WHERE id=$4
	`, q.Status, q.ApprovedAt, q.InvoicedAt, q.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("orçamento não encontrado")
	}
	return nil
}

func (r *PostgresQuoteRepository) DeleteByClientID(clientID string) error {
	_, err := r.db.Exec(context.Background(), `DELETE FROM quotes WHERE client_id = $1`, clientID)
	return err
}

func (r *PostgresQuoteRepository) Delete(id string) error {
	tag, err := r.db.Exec(context.Background(), `DELETE FROM quotes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("orçamento não encontrado")
	}
	return nil
}
