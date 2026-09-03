/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	"fmt"

	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
)

type migrationRunner interface {
	Migrate(version uint) error
	Up() error
	Down() error
	Force(version int) error
}

type migrationStrategy interface {
	execute(version *uint) error
}

func newMigrationStrategy(mig migrationRunner, direction dbconfig.MigrationDirection) (migrationStrategy, error) {
	switch direction {
	case dbconfig.MigrationDirectionUp:
		return &upStrategy{mig: mig}, nil
	case dbconfig.MigrationDirectionDown:
		return &downStrategy{mig: mig}, nil
	case dbconfig.MigrationDirectionForce:
		return &forceStrategy{mig: mig}, nil
	default:
		return nil, fmt.Errorf("unknown migration direction %q", direction)
	}
}

type upStrategy struct {
	mig migrationRunner
}

func (u *upStrategy) execute(version *uint) error {
	if version != nil {
		return u.mig.Migrate(*version)
	}
	return u.mig.Up()
}

type downStrategy struct {
	mig migrationRunner
}

func (d *downStrategy) execute(version *uint) error {
	if version != nil {
		return d.mig.Migrate(*version)
	}
	return d.mig.Down()
}

type forceStrategy struct {
	mig migrationRunner
}

func (f *forceStrategy) execute(version *uint) error {
	if version == nil {
		return fmt.Errorf("force migration requires a target version")
	}
	return f.mig.Force(int(*version)) //nolint:gosec // G115: value is bounded/non-negative
}
