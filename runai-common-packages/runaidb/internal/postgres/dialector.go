/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgres

import (
	pgerrors "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/internal/postgres/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type postgresErrorTranslator interface {
	Translate(error) error
}

type translatingDialector struct {
	gorm.Dialector
}

func newTranslatingDialector(dsn string) translatingDialector {
	return translatingDialector{
		Dialector: postgres.New(postgres.Config{
			DSN: dsn,
		}),
	}
}

func (d translatingDialector) Translate(err error) error {
	if translator, ok := d.Dialector.(postgresErrorTranslator); ok {
		err = translator.Translate(err)
	}
	return pgerrors.Translate(err)
}

func (d translatingDialector) Apply(config *gorm.Config) error {
	type applicator interface{ Apply(*gorm.Config) error }
	if a, ok := d.Dialector.(applicator); ok {
		return a.Apply(config)
	}
	return nil
}

func (d translatingDialector) SavePoint(tx *gorm.DB, name string) error {
	if sp, ok := d.Dialector.(gorm.SavePointerDialectorInterface); ok {
		return sp.SavePoint(tx, name)
	}
	return gorm.ErrUnsupportedDriver
}

func (d translatingDialector) RollbackTo(tx *gorm.DB, name string) error {
	if sp, ok := d.Dialector.(gorm.SavePointerDialectorInterface); ok {
		return sp.RollbackTo(tx, name)
	}
	return gorm.ErrUnsupportedDriver
}
