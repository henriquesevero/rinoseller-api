package repository

import (
	"context"
	"fmt"

	"rinoseller-api/internal/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCapitalContributionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCapitalContributionRepository(db *pgxpool.Pool) *PostgresCapitalContributionRepository {
	return &PostgresCapitalContributionRepository{db: db}
}

func (r *PostgresCapitalContributionRepository) Save(ctx context.Context, e *domain.CapitalContribution) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO capital_contributions (id, user_id, description, amount, type, hidden, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, e.ID, nullStr(e.UserID), e.Description, e.Amount.Float64(), string(e.Type), e.Hidden, e.CreatedAt)
	return err
}

func (r *PostgresCapitalContributionRepository) FindAll(ctx context.Context, userID string) ([]domain.CapitalContribution, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if userID == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), description, amount, type, hidden, created_at
			FROM capital_contributions ORDER BY created_at DESC
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, COALESCE(user_id,''), description, amount, type, hidden, created_at
			FROM capital_contributions WHERE user_id = $1 ORDER BY created_at DESC
		`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.CapitalContribution
	for rows.Next() {
		var e domain.CapitalContribution
		var amount float64
		var cType string
		if err := rows.Scan(&e.ID, &e.UserID, &e.Description, &amount, &cType, &e.Hidden, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Amount = domain.NewMoneyFromFloat(amount)
		e.Type = domain.ContributionType(cType)
		result = append(result, e)
	}
	if result == nil {
		return []domain.CapitalContribution{}, nil
	}
	return result, nil
}

func (r *PostgresCapitalContributionRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM capital_contributions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("aporte não encontrado: %w", domain.ErrNotFound)
	}
	return nil
}
