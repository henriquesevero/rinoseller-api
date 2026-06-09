package services

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"
)

type Claims struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo  ports.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo ports.UserRepository) *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "mirra-secret-key-change-in-production"
	}
	return &AuthService{userRepo: userRepo, jwtSecret: []byte(secret)}
}

func (s *AuthService) Login(email, password string) (string, *domain.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, errors.New("credenciais inválidas")
	}
	if !user.Active {
		return "", nil, errors.New("usuário inativo")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("credenciais inválidas")
	}

	claims := Claims{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenStr, user, nil
}

func (s *AuthService) ForgotPassword(email string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Retorna sem erro para não revelar se o e-mail existe
		return "", nil
	}
	token := fmt.Sprintf("%06d", rand.Intn(1000000))
	expires := time.Now().Add(30 * time.Minute)
	if err := s.userRepo.SetResetToken(user.ID, token, expires); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) ResetPassword(token, newPassword string) error {
	user, err := s.userRepo.FindByResetToken(token)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	return s.userRepo.ClearResetToken(user.ID)
}

func (s *AuthService) ValidateToken(tokenStr string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("token inválido")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("token inválido")
	}

	return &domain.User{
		ID:    claims.UserID,
		Name:  claims.Name,
		Email: claims.Email,
		Role:  claims.Role,
	}, nil
}
