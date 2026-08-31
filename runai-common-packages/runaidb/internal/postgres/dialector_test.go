/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	pgerrors "github.com/run-ai/runai/runai-common-packages/runaidb/internal/postgres/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTranslatingDialector_ReadOnlyTransaction(t *testing.T) {
	d := translatingDialector{Dialector: postgres.New(postgres.Config{})}
	err := d.Translate(&pgconn.PgError{Code: "25006"})
	require.ErrorIs(t, err, pgerrors.ErrReadOnlyConnection)
}

func TestTranslatingDialector_DuplicatedKey(t *testing.T) {
	d := translatingDialector{Dialector: postgres.New(postgres.Config{})}
	err := d.Translate(&pgconn.PgError{Code: "23505"})
	require.ErrorIs(t, err, gorm.ErrDuplicatedKey)
}

func TestTranslatingDialector_PreservesOtherErrors(t *testing.T) {
	d := translatingDialector{}
	original := errors.New("connection refused")
	err := d.Translate(original)
	require.Same(t, original, err)
}
