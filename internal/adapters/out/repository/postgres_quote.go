package repository

import (
	"context"
	"encoding/json"
	"fmt"

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

func (r *PostgresQuoteRepository) Save(ctx context.Context, q *domain.Quote) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO quotes (id, user_id, client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at, delivered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, q.ID, nullStr(q.UserID), q.ClientID, q.ClientName, q.Total.Float64(), string(q.Status), q.Notes,
		string(q.PaymentType), q.Installments, q.CreatedAt, q.ApprovedAt, q.InvoicedAt, q.DeliveredAt)
	if err != nil {
		return err
	}

	for _, item := range q.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO quote_items (id, quote_id, product_id, product_name, quantity, unit_price, subtotal, kit_items)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, uuid.New().String(), q.ID, item.ProductID, item.ProductName, item.Quantity, item.UnitPrice.Float64(), item.Subtotal.Float64(), marshalKitItems(item.KitItems))
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresQuoteRepository) FindAll(ctx context.Context, userID string) ([]domain.Quote, error) {
	if userID == "" {
		return r.query(ctx, `
			SELECT id, COALESCE(user_id,''), client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at, delivered_at
			FROM quotes ORDER BY created_at DESC
		`)
	}
	return r.query(ctx, `
		SELECT id, COALESCE(user_id,''), client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at, delivered_at
		FROM quotes WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
}

func (r *PostgresQuoteRepository) FindByClientID(ctx context.Context, clientID string) ([]domain.Quote, error) {
	return r.query(ctx, `
		SELECT id, COALESCE(user_id,''), client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at, delivered_at
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
		q, err := scanQuote(rows)
		if err != nil {
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
		item, err := scanQuoteItem(itemRows, &quoteID)
		if err != nil {
			return nil, err
		}
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

func (r *PostgresQuoteRepository) FindByID(ctx context.Context, id string) (*domain.Quote, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(user_id,''), client_id, client_name, total, status, notes, payment_type, installments, created_at, approved_at, invoiced_at, delivered_at
		FROM quotes WHERE id = $1
	`, id)
	q, err := scanQuote(row)
	if err != nil {
		return nil, fmt.Errorf("orçamento não encontrado: %w", domain.ErrNotFound)
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
		item, err := scanQuoteItem(itemRows, nil)
		if err != nil {
			return nil, err
		}
		q.Items = append(q.Items, item)
	}

	return &q, nil
}

func (r *PostgresQuoteRepository) Update(ctx context.Context, q *domain.Quote) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE quotes SET status=$1, approved_at=$2, invoiced_at=$3, delivered_at=$4 WHERE id=$5
	`, string(q.Status), q.ApprovedAt, q.InvoicedAt, q.DeliveredAt, q.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("orçamento não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresQuoteRepository) DeleteByClientID(ctx context.Context, clientID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM quotes WHERE client_id = $1`, clientID)
	return err
}

func (r *PostgresQuoteRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM quotes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("orçamento não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}

func scanQuote(row rowScanner) (domain.Quote, error) {
	var q domain.Quote
	var total float64
	var status, paymentType string
	err := row.Scan(&q.ID, &q.UserID, &q.ClientID, &q.ClientName, &total, &status, &q.Notes,
		&paymentType, &q.Installments, &q.CreatedAt, &q.ApprovedAt, &q.InvoicedAt, &q.DeliveredAt)
	if err != nil {
		return domain.Quote{}, err
	}
	q.Total = domain.NewMoneyFromFloat(total)
	q.Status = domain.QuoteStatus(status)
	q.PaymentType = domain.PaymentType(paymentType)
	return q, nil
}

func scanQuoteItem(row rowScanner, quoteID *string) (domain.QuoteItem, error) {
	var item domain.QuoteItem
	var unitPrice, subtotal float64
	var kitItemsRaw string
	var err error
	if quoteID != nil {
		err = row.Scan(quoteID, &item.ProductID, &item.ProductName, &item.Quantity, &unitPrice, &subtotal, &kitItemsRaw)
	} else {
		err = row.Scan(&item.ProductID, &item.ProductName, &item.Quantity, &unitPrice, &subtotal, &kitItemsRaw)
	}
	if err != nil {
		return domain.QuoteItem{}, err
	}
	item.UnitPrice = domain.NewMoneyFromFloat(unitPrice)
	item.Subtotal = domain.NewMoneyFromFloat(subtotal)
	item.KitItems = unmarshalKitItems([]byte(kitItemsRaw))
	return item, nil
}
