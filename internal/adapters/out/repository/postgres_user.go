package repository

import (
	"context"
	"errors"
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

func (r *PostgresUserRepository) Save(u *domain.User) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO users (id, name, email, password_hash, role, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, u.ID, u.Name, u.Email, u.PasswordHash, u.Role, u.Active, u.CreatedAt)
	return err
}

func (r *PostgresUserRepository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(context.Background(), `
		SELECT id, name, email, password_hash, role, active, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}
	return &u, nil
}

func (r *PostgresUserRepository) FindByID(id string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(context.Background(), `
		SELECT id, name, email, password_hash, role, active, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}
	return &u, nil
}

func (r *PostgresUserRepository) FindAll() ([]domain.User, error) {
	rows, err := r.db.Query(context.Background(), `
		SELECT id, name, email, role, active, created_at FROM users ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Active, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []domain.User{}
	}
	return users, nil
}

func (r *PostgresUserRepository) Update(u *domain.User) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE users SET name=$1, email=$2, password_hash=$3, role=$4, active=$5 WHERE id=$6
	`, u.Name, u.Email, u.PasswordHash, u.Role, u.Active, u.ID)
	return err
}

func (r *PostgresUserRepository) HasAnyAdmin() (bool, error) {
	var count int
	err := r.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM users WHERE role='admin'`).Scan(&count)
	return count > 0, err
}

func (r *PostgresUserRepository) SetResetToken(id, token string, expires time.Time) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE users SET reset_token=$1, reset_expires_at=$2 WHERE id=$3`,
		token, expires, id)
	return err
}

func (r *PostgresUserRepository) FindByResetToken(token string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(context.Background(), `
		SELECT id, name, email, password_hash, role, active, created_at
		FROM users WHERE reset_token=$1 AND reset_expires_at > NOW()
	`, token).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt)
	if err != nil {
		return nil, errors.New("código inválido ou expirado")
	}
	return &u, nil
}

func (r *PostgresUserRepository) ClearResetToken(id string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE users SET reset_token=NULL, reset_expires_at=NULL WHERE id=$1`, id)
	return err
}
