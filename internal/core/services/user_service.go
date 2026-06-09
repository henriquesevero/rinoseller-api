package services

import (
	"errors"
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

func (s *UserService) CreateUser(name, email, password, role string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		ID:           uuid.New().String(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		Active:       true,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Save(u); err != nil {
		return nil, errors.New("e-mail já cadastrado")
	}

	return u, nil
}

func (s *UserService) ListUsers() ([]domain.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) GetUser(id string) (*domain.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) UpdateUser(id, name, email string, active bool) (*domain.User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	u.Name = name
	u.Email = email
	u.Active = active
	if err := s.repo.Update(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) EnsureDefaultAdmin() error {
	has, err := s.repo.HasAnyAdmin()
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	_, err = s.CreateUser("Administrador", "admin@mirra.com", "mirra2024", "admin")
	return err
}
