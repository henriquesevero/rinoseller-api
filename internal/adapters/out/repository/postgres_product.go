package repository

import (
	"context"
	"errors"

	"rinoseller-api/internal/core/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProductRepository(db *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Save(p *domain.Product) error {
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

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
	`, p.ID, nullStr(p.UserID), p.Name, p.Category, p.Price, p.CostPrice, p.StockQuantity, p.Code, p.IsKit)
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

func (r *PostgresProductRepository) findKitItems(ctx context.Context, kitID string) ([]domain.KitItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT product_id, product_name, quantity FROM kit_items WHERE kit_id = $1
	`, kitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.KitItem
	for rows.Next() {
		var ki domain.KitItem
		if err := rows.Scan(&ki.ProductID, &ki.ProductName, &ki.Quantity); err != nil {
			return nil, err
		}
		items = append(items, ki)
	}
	return items, nil
}

func (r *PostgresProductRepository) FindAll(userID string) ([]domain.Product, error) {
	var (
		rows interface {
			Next() bool
			Scan(...any) error
			Close()
		}
		err error
	)
	if userID == "" {
		rows, err = r.db.Query(context.Background(), `
			SELECT id, COALESCE(user_id,''), name, category, price, cost_price, stock_quantity, code, is_kit
			FROM products ORDER BY name
		`)
	} else {
		rows, err = r.db.Query(context.Background(), `
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
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Category, &p.Price, &p.CostPrice, &p.StockQuantity, &p.Code, &p.IsKit); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	if result == nil {
		return []domain.Product{}, nil
	}

	ctx := context.Background()
	for i := range result {
		if !result[i].IsKit {
			continue
		}
		items, err := r.findKitItems(ctx, result[i].ID)
		if err != nil {
			return nil, err
		}
		result[i].KitItems = items
	}
	return result, nil
}

func (r *PostgresProductRepository) FindByID(id string) (*domain.Product, error) {
	ctx := context.Background()
	var p domain.Product
	err := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(user_id,''), name, category, price, cost_price, stock_quantity, code, is_kit
		FROM products WHERE id = $1
	`, id).Scan(&p.ID, &p.UserID, &p.Name, &p.Category, &p.Price, &p.CostPrice, &p.StockQuantity, &p.Code, &p.IsKit)
	if err != nil {
		return nil, errors.New("produto não encontrado")
	}
	if p.IsKit {
		items, err := r.findKitItems(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.KitItems = items
	}
	return &p, nil
}

func (r *PostgresProductRepository) UpdateStock(id string, newStock int) error {
	tag, err := r.db.Exec(context.Background(), `UPDATE products SET stock_quantity=$1 WHERE id=$2`, newStock, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("produto não encontrado")
	}
	return nil
}

func (r *PostgresProductRepository) UpdatePrice(id string, price float64) error {
	tag, err := r.db.Exec(context.Background(), `UPDATE products SET price=$1 WHERE id=$2`, price, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("produto não encontrado")
	}
	return nil
}

func (r *PostgresProductRepository) Delete(id string) error {
	tag, err := r.db.Exec(context.Background(), `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("produto não encontrado")
	}
	return nil
}
