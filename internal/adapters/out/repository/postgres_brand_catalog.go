package repository

import (
	"context"
	"fmt"

	"rinoseller-api/internal/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBrandCatalogRepository struct {
	db *pgxpool.Pool
}

func NewPostgresBrandCatalogRepository(db *pgxpool.Pool) *PostgresBrandCatalogRepository {
	return &PostgresBrandCatalogRepository{db: db}
}

func (r *PostgresBrandCatalogRepository) Save(ctx context.Context, c *domain.BrandCatalog) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO brand_catalogs (id, user_id, brand_name, drive_url, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, c.ID, nullStr(c.UserID), c.BrandName, c.DriveURL, c.CreatedAt)
	return err
}

func (r *PostgresBrandCatalogRepository) FindAll(ctx context.Context, userID string) ([]domain.BrandCatalog, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if userID == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), brand_name, drive_url, created_at
			FROM brand_catalogs ORDER BY brand_name
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), brand_name, drive_url, created_at
			FROM brand_catalogs WHERE user_id = $1 ORDER BY brand_name
		`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.BrandCatalog
	for rows.Next() {
		var c domain.BrandCatalog
		if err := rows.Scan(&c.ID, &c.UserID, &c.BrandName, &c.DriveURL, &c.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	if result == nil {
		return []domain.BrandCatalog{}, nil
	}
	return result, nil
}

func (r *PostgresBrandCatalogRepository) FindByID(ctx context.Context, id string) (*domain.BrandCatalog, error) {
	var c domain.BrandCatalog
	err := r.db.QueryRow(ctx, `
		SELECT id, COALESCE(user_id,''), brand_name, drive_url, created_at
		FROM brand_catalogs WHERE id = $1
	`, id).Scan(&c.ID, &c.UserID, &c.BrandName, &c.DriveURL, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("catálogo não encontrado: %w", domain.ErrNotFound)
	}
	return &c, nil
}

func (r *PostgresBrandCatalogRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM brand_catalogs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("catálogo não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}
