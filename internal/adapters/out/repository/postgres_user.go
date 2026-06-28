package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rinoseller-api/internal/core/domain"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, role, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, u.ID, u.Name, u.Email, u.PasswordHash, string(u.Role), u.Active, u.CreatedAt)
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	var role string
	err := r.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, active, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &u.Active, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", domain.ErrNotFound)
	}
	u.Role = domain.UserRole(role)
	return &u, nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	var role string
	err := r.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, active, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &u.Active, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", domain.ErrNotFound)
	}
	u.Role = domain.UserRole(role)
	return &u, nil
}

func (r *PostgresUserRepository) FindAll(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, email, role, active, created_at FROM users ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var role string
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &role, &u.Active, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.Role = domain.UserRole(role)
		users = append(users, u)
	}
	if users == nil {
		users = []domain.User{}
	}
	return users, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET name=$1, email=$2, password_hash=$3, role=$4, active=$5 WHERE id=$6
	`, u.Name, u.Email, u.PasswordHash, string(u.Role), u.Active, u.ID)
	return err
}

func (r *PostgresUserRepository) HasAnyAdmin(ctx context.Context) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role='admin'`).Scan(&count)
	return count > 0, err
}

func (r *PostgresUserRepository) SetResetToken(ctx context.Context, id, token string, expires time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET reset_token=$1, reset_expires_at=$2 WHERE id=$3`,
		token, expires, id)
	return err
}

func (r *PostgresUserRepository) FindByResetToken(ctx context.Context, token string) (*domain.User, error) {
	var u domain.User
	var role string
	err := r.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, role, active, created_at
		FROM users WHERE reset_token=$1 AND reset_expires_at > NOW()
	`, token).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &u.Active, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("código inválido ou expirado: %w", domain.ErrNotFound)
	}
	u.Role = domain.UserRole(role)
	return &u, nil
}

func (r *PostgresUserRepository) ClearResetToken(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET reset_token=NULL, reset_expires_at=NULL WHERE id=$1`, id)
	return err
}
