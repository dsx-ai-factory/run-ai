/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package errors

import (
	"errors"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "404 Not Found error",
			err:      errors.New("404 Not Found"),
			expected: true,
		},
		{
			name:     "not found lowercase",
			err:      errors.New("resource not found"),
			expected: true,
		},
		{
			name:     "NOT FOUND uppercase",
			err:      errors.New("resource NOT FOUND"),
			expected: true,
		},
		{
			name:     "Not Found mixed case",
			err:      errors.New("resource Not Found"),
			expected: true,
		},
		{
			name:     "non-matching error",
			err:      errors.New("bad request"),
			expected: false,
		},
		{
			name:     "empty error message",
			err:      errors.New(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsBadRequest(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "400 Bad Request error",
			err:      errors.New("400 Bad Request"),
			expected: true,
		},
		{
			name:     "bad request lowercase",
			err:      errors.New("bad request error"),
			expected: true,
		},
		{
			name:     "BAD REQUEST uppercase",
			err:      errors.New("BAD REQUEST error"),
			expected: true,
		},
		{
			name:     "Bad Request mixed case",
			err:      errors.New("bad Request error"),
			expected: true,
		},
		{
			name:     "non-matching error",
			err:      errors.New("not found"),
			expected: false,
		},
		{
			name:     "empty error message",
			err:      errors.New(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBadRequest(tt.err)
			if result != tt.expected {
				t.Errorf("IsBadRequest() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "401 Unauthorized error",
			err:      errors.New("401 Unauthorized"),
			expected: true,
		},
		{
			name:     "unauthorized lowercase",
			err:      errors.New("unauthorized access"),
			expected: true,
		},
		{
			name:     "UNAUTHORIZED uppercase",
			err:      errors.New("UNAUTHORIZED access"),
			expected: true,
		},
		{
			name:     "Unauthorized mixed case",
			err:      errors.New("Unauthorized access"),
			expected: true,
		},
		{
			name:     "non-matching error",
			err:      errors.New("not found"),
			expected: false,
		},
		{
			name:     "empty error message",
			err:      errors.New(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUnauthorized(tt.err)
			if result != tt.expected {
				t.Errorf("IsUnauthorized() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsForbidden(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "403 Forbidden error",
			err:      errors.New("403 Forbidden"),
			expected: true,
		},
		{
			name:     "forbidden lowercase",
			err:      errors.New("forbidden access"),
			expected: true,
		},
		{
			name:     "FORBIDDEN uppercase",
			err:      errors.New("FORBIDDEN access"),
			expected: true,
		},
		{
			name:     "Forbidden mixed case",
			err:      errors.New("Forbidden access"),
			expected: true,
		},
		{
			name:     "non-matching error",
			err:      errors.New("not found"),
			expected: false,
		},
		{
			name:     "empty error message",
			err:      errors.New(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsForbidden(tt.err)
			if result != tt.expected {
				t.Errorf("IsForbidden() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsInternalServer(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "500 Internal Server Error",
			err:      errors.New("500 Internal Server Error"),
			expected: true,
		},
		{
			name:     "internal server error lowercase",
			err:      errors.New("internal server error occurred"),
			expected: true,
		},
		{
			name:     "INTERNAL SERVER ERROR uppercase",
			err:      errors.New("INTERNAL SERVER ERROR occurred"),
			expected: true,
		},
		{
			name:     "Internal Server Error mixed case",
			err:      errors.New("Internal Server Error occurred"),
			expected: true,
		},
		{
			name:     "non-matching error",
			err:      errors.New("not found"),
			expected: false,
		},
		{
			name:     "empty error message",
			err:      errors.New(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInternalServerError(tt.err)
			if result != tt.expected {
				t.Errorf("IsInternalServerError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsServiceUnavailable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "503 Service Unavailable",
			err:      errors.New("503 Service Unavailable"),
			expected: true,
		},
		{
			name:     "service unavailable lowercase",
			err:      errors.New("service unavailable temporarily"),
			expected: true,
		},
		{
			name:     "SERVICE UNAVAILABLE uppercase",
			err:      errors.New("SERVICE UNAVAILABLE temporarily"),
			expected: true,
		},
		{
			name:     "Service Unavailable mixed case",
			err:      errors.New("Service Unavailable temporarily"),
			expected: true,
		},
		{
			name:     "non-matching error",
			err:      errors.New("not found"),
			expected: false,
		},
		{
			name:     "empty error message",
			err:      errors.New(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServiceUnavailable(tt.err)
			if result != tt.expected {
				t.Errorf("IsServiceUnavailable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsErrorHelper tests the isError helper function directly
func TestIsErrorHelper(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		pattern  string
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			pattern:  "not found",
			expected: false,
		},
		{
			name:     "pattern match",
			err:      errors.New("resource not found"),
			pattern:  "not found",
			expected: true,
		},
		{
			name:     "pattern match with numbers",
			err:      errors.New("404 Not Found"),
			pattern:  "not found",
			expected: true,
		},
		{
			name:     "pattern match with different text",
			err:      errors.New("unauthorized access"),
			pattern:  "unauthorized",
			expected: true,
		},
		{
			name:     "no pattern match",
			err:      errors.New("bad request"),
			pattern:  "not found",
			expected: false,
		},
		{
			name:     "case insensitive match",
			err:      errors.New("NOT FOUND"),
			pattern:  "not found",
			expected: true,
		},
		{
			name:     "empty pattern",
			err:      errors.New("some error"),
			pattern:  "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isError(tt.err, tt.pattern)
			if result != tt.expected {
				t.Errorf("isError() = %v, want %v", result, tt.expected)
			}
		})
	}
}
