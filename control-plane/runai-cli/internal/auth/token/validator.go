/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package token

import (
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
)

type Validator struct {
	expirationDelta time.Duration
}

func NewTokenValidator(timeDelta time.Duration) *Validator {
	return &Validator{expirationDelta: timeDelta}
}

func (tv *Validator) Validate(token string) (bool, error) {
	if token == "" {
		return false, nil
	}
	claims, err := auth.GetClaims(token)
	if err != nil {
		return false, err
	}
	return claims.ExpiresAt.After(time.Now().Add(tv.expirationDelta)), nil
}
