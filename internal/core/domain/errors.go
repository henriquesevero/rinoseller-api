package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("recurso não encontrado")
	ErrConflict   = errors.New("conflito com o estado atual do recurso")
	ErrValidation = errors.New("dados inválidos")
	ErrForbidden  = errors.New("operação não permitida")

	ErrInvalidCredentials = fmt.Errorf("credenciais inválidas: %w", ErrValidation)
	ErrAccountInactive    = fmt.Errorf("conta inativa: %w", ErrForbidden)
	ErrEmailNotVerified   = fmt.Errorf("e-mail não verificado: %w", ErrForbidden)
	ErrEmailTaken         = fmt.Errorf("e-mail já cadastrado: %w", ErrConflict)
)
