package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"
)

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string, role domain.UserRole) (*domain.User, error) {
	if !role.Valid() {
		role = domain.RoleSeller
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		ID:                 uuid.New().String(),
		Name:               name,
		Email:              email,
		PasswordHash:       string(hash),
		Role:               role,
		Active:             true,
		EmailVerified:      true,
		SubscriptionActive: true,
		CreatedAt:          time.Now(),
	}

	if err := s.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("e-mail já cadastrado: %w", domain.ErrConflict)
	}

	return u, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.FindAll(ctx)
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id, name, email string, active bool) (*domain.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.Name = name
	u.Email = email
	u.Active = active
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) EnsureDefaultAdmin(ctx context.Context) (string, error) {
	has, err := s.repo.HasAnyAdmin(ctx)
	if err != nil {
		return "", err
	}
	if has {
		return "", nil
	}

	password, err := randomPassword()
	if err != nil {
		return "", err
	}
	if _, err := s.CreateUser(ctx, "Administrador", "admin@rinoseller.com", password, domain.RoleAdmin); err != nil {
		return "", err
	}
	return password, nil
}

func randomPassword() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
