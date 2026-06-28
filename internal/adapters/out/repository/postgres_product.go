package repository

import (
	"context"
	"fmt"

	"rinoseller-api/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProductRepository(db *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Save(ctx context.Context, p *domain.Product) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO products (id, user_id, name, category, price, cost_price, stock_quantity, code, is_kit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			name           = EXCLUDED.name,
			category       = EXCLUDED.category,
			price          = EXCLUDED.price,
			cost_price     = EXCLUDED.cost_price,
			stock_quantity = EXCLUDED.stock_quantity,
			code           = EXCLUDED.code,
			is_kit         = EXCLUDED.is_kit
	`, p.ID, nullStr(p.UserID), p.Name, p.Category, p.Price.Float64(), p.CostPrice.Float64(), p.StockQuantity, p.Code, p.IsKit)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `DELETE FROM kit_items WHERE kit_id = $1`, p.ID); err != nil {
		return err
	}
	for _, ki := range p.KitItems {
		_, err = tx.Exec(ctx, `
			INSERT INTO kit_items (id, kit_id, product_id, product_name, quantity)
			VALUES ($1, $2, $3, $4, $5)
		`, uuid.New().String(), p.ID, ki.ProductID, ki.ProductName, ki.Quantity)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresProductRepository) findKitItemsByKitIDs(ctx context.Context, kitIDs []string) (map[string][]domain.KitItem, error) {
	if len(kitIDs) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT kit_id, product_id, product_name, quantity FROM kit_items WHERE kit_id = ANY($1)
	`, kitIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itemsByKitID := make(map[string][]domain.KitItem)
	for rows.Next() {
		var kitID string
		var ki domain.KitItem
		if err := rows.Scan(&kitID, &ki.ProductID, &ki.ProductName, &ki.Quantity); err != nil {
			return nil, err
		}
		itemsByKitID[kitID] = append(itemsByKitID[kitID], ki)
	}
	return itemsByKitID, nil
}

func (r *PostgresProductRepository) findKitItems(ctx context.Context, kitID string) ([]domain.KitItem, error) {
	itemsByKitID, err := r.findKitItemsByKitIDs(ctx, []string{kitID})
	if err != nil {
		return nil, err
	}
	return itemsByKitID[kitID], nil
}

func (r *PostgresProductRepository) FindAll(ctx context.Context, userID string) ([]domain.Product, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if userID == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), name, category, price, cost_price, stock_quantity, code, is_kit
			FROM products ORDER BY name
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), name, category, price, cost_price, stock_quantity, code, is_kit
			FROM products WHERE user_id = $1 ORDER BY name
		`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Product
	for rows.Next() {
		var p domain.Product
		var price, costPrice float64
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Category, &price, &costPrice, &p.StockQuantity, &p.Code, &p.IsKit); err != nil {
			return nil, err
		}
		p.Price = domain.NewMoneyFromFloat(price)
		p.CostPrice = domain.NewMoneyFromFloat(costPrice)
		result = append(result, p)
	}
	if result == nil {
		return []domain.Product{}, nil
	}

	var kitIDs []string
	for _, p := range result {
		if p.IsKit {
			kitIDs = append(kitIDs, p.ID)
		}
	}
	itemsByKitID, err := r.findKitItemsByKitIDs(ctx, kitIDs)
	if err != nil {
		return nil, err
	}
	for i := range result {
		result[i].KitItems = itemsByKitID[result[i].ID]
	}
	return result, nil
}

func (r *PostgresProductRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	var price, costPrice float64
	err := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(user_id,''), name, category, price, cost_price, stock_quantity, code, is_kit
		FROM products WHERE id = $1
	`, id).Scan(&p.ID, &p.UserID, &p.Name, &p.Category, &price, &costPrice, &p.StockQuantity, &p.Code, &p.IsKit)
	if err != nil {
		return nil, fmt.Errorf("produto não encontrado: %w", domain.ErrNotFound)
	}
	p.Price = domain.NewMoneyFromFloat(price)
	p.CostPrice = domain.NewMoneyFromFloat(costPrice)
	if p.IsKit {
		items, err := r.findKitItems(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.KitItems = items
	}
	return &p, nil
}

func (r *PostgresProductRepository) UpdateStock(ctx context.Context, id string, newStock int) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET stock_quantity=$1 WHERE id=$2`, newStock, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("produto não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresProductRepository) UpdatePrice(ctx context.Context, id string, price domain.Money) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET price=$1 WHERE id=$2`, price.Float64(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("produto não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("produto não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}
