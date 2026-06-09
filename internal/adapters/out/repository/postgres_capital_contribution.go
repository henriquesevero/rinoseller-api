package repository

import (
	"context"
	"errors"

	"rinoseller-api/internal/core/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCapitalContributionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCapitalContributionRepository(db *pgxpool.Pool) *PostgresCapitalContributionRepository {
	return &PostgresCapitalContributionRepository{db: db}
}

func (r *PostgresCapitalContributionRepository) Save(e *domain.CapitalContribution) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO capital_contributions (id, user_id, description, amount, type, hidden, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, e.ID, nullStr(e.UserID), e.Description, e.Amount, e.Type, e.Hidden, e.CreatedAt)
	return err
}

func (r *PostgresCapitalContributionRepository) FindAll(userID string) ([]domain.CapitalContribution, error) {
	var rows interface {
		Next() bool
		Scan(...any) error
		Close()
	}
	var err error

	if userID == "" {
		rows, err = r.db.Query(context.Background(), `
			SELECT id, COALESCE(user_id,''), description, amount, type, hidden, created_at
			FROM capital_contributions ORDER BY created_at DESC
		`)
	} else {
		rows, err = r.db.Query(context.Background(), `
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
		if err := rows.Scan(&e.ID, &e.UserID, &e.Description, &e.Amount, &e.Type, &e.Hidden, &e.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	if result == nil {
		return []domain.CapitalContribution{}, nil
	}
	return result, nil
}

func (r *PostgresCapitalContributionRepository) Delete(id string) error {
	tag, err := r.db.Exec(context.Background(), `DELETE FROM capital_contributions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("aporte não encontrado")
	}
	return nil
}
