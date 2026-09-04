/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaihttp

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/stretchr/testify/require"
)

const defaultCode = http.StatusServiceUnavailable

func loadErrorDetails(t *testing.T, goError error) (int32, string, string) {
	t.Helper()

	code, message, details := int32(defaultCode), "", ""
	LoadErrorDetails(&code, &message, &details, goError)
	return code, message, details
}

func TestLoadErrorDetails_NilErrorLeavesResultUntouched(t *testing.T) {
	code, message, details := loadErrorDetails(t, nil)

	require.Equal(t, int32(defaultCode), code)
	require.Empty(t, message)
	require.Empty(t, details)
}

// gorm replaces the driver error with a bare sentinel, so there is nothing to unwrap.
func TestLoadErrorDetails_OrmSentinels(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		code    int32
		details string
	}{
		{"record not found", orm.ErrRecordNotFound, http.StatusNotFound, "requested resource does not exist"},
		{"duplicated key", orm.ErrDuplicatedKey, http.StatusConflict, orm.ErrDuplicatedKey.Error()},
		{"foreign key violated", orm.ErrForeignKeyViolated, http.StatusPreconditionFailed, orm.ErrForeignKeyViolated.Error()},
		{"check constraint violated", orm.ErrCheckConstraintViolated, http.StatusConflict, orm.ErrCheckConstraintViolated.Error()},
		{"read only connection", orm.ErrReadOnlyConnection, http.StatusServiceUnavailable, orm.ErrReadOnlyConnection.Error()},
		{"undefined table", orm.ErrUndefinedTable, http.StatusInternalServerError, orm.ErrUndefinedTable.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, message, details := loadErrorDetails(t, tt.err)

			require.Equal(t, tt.code, code)
			require.Empty(t, message)
			require.Equal(t, tt.details, details)

			wrappedCode, _, wrappedDetails := loadErrorDetails(t, fmt.Errorf("query failed: %w", tt.err))
			require.Equal(t, tt.code, wrappedCode)
			require.Equal(t, tt.details, wrappedDetails)
		})
	}
}

func TestLoadErrorDetails_SentinelWrappingPgError(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		pgCode   string
		code     int32
	}{
		{"read only connection", orm.ErrReadOnlyConnection, "25006", http.StatusServiceUnavailable},
		{"undefined table", orm.ErrUndefinedTable, "42P01", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: tt.pgCode, Message: "relation \"internal_table\" does not exist"}
			code, _, details := loadErrorDetails(t, fmt.Errorf("%w: %w", tt.sentinel, pgErr))

			require.Equal(t, tt.code, code)
			require.Equal(t, tt.sentinel.Error(), details)
		})
	}
}

func TestLoadErrorDetails_UntranslatedPgErrors(t *testing.T) {
	tests := []struct {
		name    string
		pgErr   *pgconn.PgError
		code    int32
		message string
		details string
	}{
		{
			name:    "duplicate key",
			pgErr:   &pgconn.PgError{Code: ErrDuplicateKey, Detail: "Key (name)=(a) already exists."},
			code:    http.StatusConflict,
			details: "Key (name)=(a) already exists.",
		},
		{
			name:    "invalid syntax",
			pgErr:   &pgconn.PgError{Code: ErrInvalidSyntax, Message: "invalid input syntax for type uuid"},
			code:    http.StatusBadRequest,
			details: "invalid input syntax for type uuid",
		},
		{
			name:    "foreign key",
			pgErr:   &pgconn.PgError{Code: ErrForeignKey, ConstraintName: "workloads_project_fk"},
			code:    http.StatusPreconditionFailed,
			details: "The request invalidates workloads_project_fk dependency",
		},
		{
			name:    "trigger failed",
			pgErr:   &pgconn.PgError{Code: ErrTriggerFailed, Message: "project is not empty"},
			code:    http.StatusConflict,
			message: "project is not empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, message, details := loadErrorDetails(t, fmt.Errorf("exec failed: %w", tt.pgErr))

			require.Equal(t, tt.code, code)
			require.Equal(t, tt.message, message)
			require.Equal(t, tt.details, details)
		})
	}
}

// Raw Postgres text can expose schema internals, so an unmapped SQL state reports nothing.
func TestLoadErrorDetails_UnmappedPgErrorKeepsCodeAndHidesRawError(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "42501", Message: "permission denied for table tenants"}
	code, message, details := loadErrorDetails(t, pgErr)

	require.Equal(t, int32(defaultCode), code)
	require.Empty(t, message)
	require.Empty(t, details)
}

func TestLoadErrorDetails_UnknownErrorKeepsCode(t *testing.T) {
	code, message, details := loadErrorDetails(t, errors.New("connection refused"))

	require.Equal(t, int32(defaultCode), code)
	require.Empty(t, message)
	require.Equal(t, "connection refused", details)
}
