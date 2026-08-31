/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	"fmt"
	"testing"

	"github.com/aws/smithy-go/ptr"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/postgres/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewMigrationStrategy_UnknownDirection(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	_, err := newMigrationStrategy(mig, dbconfig.MigrationDirection("invalid"))
	require.EqualError(t, err, `unknown migration direction "invalid"`)
}

func TestNewMigrationStrategy_ReturnsCorrectStrategy(t *testing.T) {
	tests := []struct {
		direction dbconfig.MigrationDirection
		wantType  migrationStrategy
	}{
		{dbconfig.MigrationDirectionUp, &upStrategy{}},
		{dbconfig.MigrationDirectionDown, &downStrategy{}},
		{dbconfig.MigrationDirectionForce, &forceStrategy{}},
	}
	for _, tt := range tests {
		t.Run(string(tt.direction), func(t *testing.T) {
			mig := mocks.NewMigrationRunner(t)
			s, err := newMigrationStrategy(mig, tt.direction)
			require.NoError(t, err)
			require.IsType(t, tt.wantType, s)
		})
	}
}

func TestUpStrategy_NoVersion_CallsUp(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Up").Return(nil)
	err := (&upStrategy{mig: mig}).execute(nil)
	require.NoError(t, err)
}

func TestUpStrategy_WithVersion_CallsMigrate(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Migrate", uint(5)).Return(nil)
	err := (&upStrategy{mig: mig}).execute(ptr.Uint(5))
	require.NoError(t, err)
}

func TestUpStrategy_Error(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Up").Return(fmt.Errorf("up failed"))
	err := (&upStrategy{mig: mig}).execute(nil)
	require.EqualError(t, err, "up failed")
}

func TestDownStrategy_NoVersion_CallsDown(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Down").Return(nil)
	err := (&downStrategy{mig: mig}).execute(nil)
	require.NoError(t, err)
}

func TestDownStrategy_WithVersion_CallsMigrate(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Migrate", uint(3)).Return(nil)
	err := (&downStrategy{mig: mig}).execute(ptr.Uint(3))
	require.NoError(t, err)
}

func TestDownStrategy_Error(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Down").Return(fmt.Errorf("down failed"))
	err := (&downStrategy{mig: mig}).execute(nil)
	require.EqualError(t, err, "down failed")
}

func TestForceStrategy_NoVersion_ReturnsError(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	err := (&forceStrategy{mig: mig}).execute(nil)
	require.EqualError(t, err, "force migration requires a target version")
}

func TestForceStrategy_WithVersion_CallsForce(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Force", 7).Return(nil)
	err := (&forceStrategy{mig: mig}).execute(ptr.Uint(7))
	require.NoError(t, err)
}

func TestForceStrategy_Error(t *testing.T) {
	mig := mocks.NewMigrationRunner(t)
	mig.On("Force", 7).Return(fmt.Errorf("force failed"))
	err := (&forceStrategy{mig: mig}).execute(ptr.Uint(7))
	require.EqualError(t, err, "force failed")
}
