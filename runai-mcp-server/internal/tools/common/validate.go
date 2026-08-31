/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func ValidateUUID(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	parsed, err := uuid.Parse(value)
	if err != nil || !strings.EqualFold(parsed.String(), value) {
		return fmt.Errorf("%s must be a UUID (got %q)", field, value)
	}
	return nil
}

func ValidateNumericString(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if _, err := strconv.ParseUint(value, 10, 64); err != nil {
		return fmt.Errorf("%s must be a numeric string (got %q)", field, value)
	}
	return nil
}

func RequiredString(field string, value any) (string, error) {
	s, ok := value.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	return s, nil
}

func IsFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func ParseFiniteFloat(s string) (float64, bool) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || !IsFinite(f) || f < 0 {
		return 0, false
	}
	return f, true
}
