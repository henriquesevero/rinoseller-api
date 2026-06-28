package domain

import "errors"

var (
	ErrNotFound   = errors.New("recurso não encontrado")
	ErrConflict   = errors.New("conflito com o estado atual do recurso")
	ErrValidation = errors.New("dados inválidos")
	ErrForbidden  = errors.New("operação não permitida")
)
