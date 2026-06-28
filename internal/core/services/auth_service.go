package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"rinoseller-api/internal/core/domain"
	"rinoseller-api/internal/core/ports"
)

type Claims struct {
	UserID string          `json:"user_id"`
	Name   string          `json:"name"`
	Email  string          `json:"email"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo    ports.UserRepository
	emailSender ports.EmailSender
	jwtSecret   []byte
	frontendURL string
}

func NewAuthService(userRepo ports.UserRepository, emailSender ports.EmailSender, frontendURL string) (*AuthService, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET não configurada")
	}
	return &AuthService{userRepo: userRepo, emailSender: emailSender, jwtSecret: []byte(secret), frontendURL: frontendURL}, nil
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:            uuid.New().String(),
		Name:          name,
		Email:         email,
		PasswordHash:  string(hash),
		Role:          domain.RoleSeller,
		Active:        true,
		EmailVerified: false,
		CreatedAt:     time.Now(),
	}
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("e-mail já cadastrado: %w", domain.ErrConflict)
	}

	token := uuid.New().String()
	if err := s.userRepo.SetVerificationToken(ctx, user.ID, token); err != nil {
		return nil, err
	}

	verifyLink := fmt.Sprintf("%s/verificar-email?token=%s", s.frontendURL, token)
	subject := "Confirme seu e-mail — RinoSeller"
	body := fmt.Sprintf(`
		<p>Olá, %s!</p>
		<p>Falta só um passo para ativar sua conta RinoSeller: confirme seu e-mail clicando no link abaixo.</p>
		<p><a href="%s">%s</a></p>
		<p>Se você não criou essa conta, pode ignorar este e-mail.</p>
	`, user.Name, verifyLink, verifyLink)
	if err := s.emailSender.Send(ctx, user.Email, user.Name, subject, body); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.userRepo.FindByVerificationToken(ctx, token)
	if err != nil {
		return err
	}
	return s.userRepo.MarkEmailVerified(ctx, user.ID)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("credenciais inválidas: %w", domain.ErrValidation)
	}
	if !user.Active {
		return "", nil, fmt.Errorf("usuário inativo: %w", domain.ErrForbidden)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("credenciais inválidas: %w", domain.ErrValidation)
	}
	if !user.EmailVerified {
		return "", nil, fmt.Errorf("confirme seu e-mail antes de fazer login: %w", domain.ErrForbidden)
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

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil
	}
	token := fmt.Sprintf("%06d", rand.Intn(1000000))
	expires := time.Now().Add(30 * time.Minute)
	if err := s.userRepo.SetResetToken(ctx, user.ID, token, expires); err != nil {
		return err
	}

	resetLink := fmt.Sprintf("%s/redefinir-senha?code=%s", s.frontendURL, token)
	subject := "Recuperação de senha — RinoSeller"
	body := fmt.Sprintf(`
		<p>Olá, %s!</p>
		<p>Recebemos uma solicitação para redefinir a senha da sua conta RinoSeller.</p>
		<p>Use o código abaixo, ou clique no link para criar uma nova senha:</p>
		<p style="font-size:24px;font-weight:bold;letter-spacing:4px;">%s</p>
		<p><a href="%s">%s</a></p>
		<p>O código expira em 30 minutos. Se você não solicitou isso, pode ignorar este e-mail.</p>
	`, user.Name, token, resetLink, resetLink)

	return s.emailSender.Send(ctx, user.Email, user.Name, subject, body)
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.userRepo.FindByResetToken(ctx, token)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}
	return s.userRepo.ClearResetToken(ctx, user.ID)
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token inválido: %w", domain.ErrForbidden)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("token inválido: %w", domain.ErrForbidden)
	}

	return &domain.User{
		ID:    claims.UserID,
		Name:  claims.Name,
		Email: claims.Email,
		Role:  claims.Role,
	}, nil
}
