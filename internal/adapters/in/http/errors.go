package httphandler

import (
	"errors"
	"net/http"

	"rinoseller-api/internal/core/domain"

	"github.com/gin-gonic/gin"
)

func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := ""
	switch {
	case errors.Is(err, domain.ErrEmailNotVerified):
		status, code = http.StatusForbidden, "email_not_verified"
	case errors.Is(err, domain.ErrAccountInactive):
		status, code = http.StatusForbidden, "account_inactive"
	case errors.Is(err, domain.ErrEmailTaken):
		status, code = http.StatusConflict, "email_already_registered"
	case errors.Is(err, domain.ErrInvalidCredentials):
		status, code = http.StatusBadRequest, "invalid_credentials"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
	}
	c.JSON(status, errorResponse{Error: err.Error(), Code: code})
}

func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, errorResponse{Error: message})
}
