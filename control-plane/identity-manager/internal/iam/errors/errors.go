package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v13"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
)

func HandleKeyCloakError(err error) error {
	if err == nil {
		return nil
	}
	keycloakError, ok := err.(*gocloak.APIError)
	if ok {
		switch keycloakError.Code {
		case http.StatusNotFound:
			return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrNotFound)
		case http.StatusConflict:
			return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrConflict)
		case http.StatusBadRequest:
			if strings.Contains(err.Error(), "Invalid user credentials") {
				return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrUnauthorized)
			}
			return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrInvalidInput)
		case http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrUnauthorized)
		default:
			return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrInternal)
		}
	} else {
		return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrInternal)
	}
}
