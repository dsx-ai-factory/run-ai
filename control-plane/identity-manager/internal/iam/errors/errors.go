/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

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
	if !ok {
		return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrInternal)
	}

	if keycloakError.Code == http.StatusNotFound || (keycloakError.Code == 0 && strings.Contains(keycloakError.Message, "not found")) {
		return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrNotFound)
	}
	if keycloakError.Code == http.StatusBadRequest || (keycloakError.Code == 0 && strings.Contains(keycloakError.Message, "bad request")) {
		if strings.Contains(err.Error(), "Invalid user credentials") {
			return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrUnauthorized)
		}
		return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrInvalidInput)
	}
	if keycloakError.Code == http.StatusUnauthorized || keycloakError.Code == http.StatusForbidden {
		return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrUnauthorized)
	}
	if keycloakError.Code == http.StatusConflict || (keycloakError.Code == 0 && strings.Contains(keycloakError.Message, "conflict")) {
		return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrConflict)
	}

	return fmt.Errorf("keycloak error: %s: %w", err.Error(), ErrInternal)
}
