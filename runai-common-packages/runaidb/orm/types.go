/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orm

import (
	"context"
	"database/sql"

	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
	"gorm.io/gorm"
)

// GORM core types.
type (
	Session = gorm.Session
)

// Enricher is a composable, reusable query-building function applied via WithEnricher.
type Enricher func(DB) DB

// DB is the main database interface composing all operation groups.
type DB interface {
	dbBasic[DB]
	Transactioner
	QueryBuilder[DB]
	QueryWriterBuilder
	QueryReader
	QueryWriter
}

type ReadDB interface {
	dbBasic[ReadDB]
	QueryBuilder[ReadDB]
	QueryReader
}

type dbBasic[T any] interface {
	Actual() *gorm.DB
	WithContext(ctx context.Context) T
	Session(config *Session) T
	Close() error
}

type Transactioner interface {
	Transaction(fc func(tx DB) error, opts ...*sql.TxOptions) error
	SavePoint(name string) DB
	RollbackTo(name string) DB
	Begin() DB
	Commit() error
	Rollback() error
}

type QueryWriterBuilder interface {
	Assign(attrs ...interface{}) DB
}

type QueryBuilder[T any] interface {
	Model(value interface{}) T
	Table(name string, args ...interface{}) T
	Clauses(conds ...clause.Expression) T
	Where(query interface{}, args ...interface{}) T
	Not(query interface{}, args ...interface{}) T
	Or(query interface{}, args ...interface{}) T
	Select(query interface{}, args ...interface{}) T
	Omit(columns ...string) T
	Joins(query string, args ...interface{}) T
	InnerJoins(query string, args ...interface{}) T
	Preload(query string, args ...interface{}) T
	Group(name string) T
	Having(query interface{}, args ...interface{}) T
	Order(value interface{}) T
	Limit(limit int) T
	Offset(offset int) T
	Distinct(args ...interface{}) T
	Unscoped() T
	Raw(sql string, values ...interface{}) T
	WithEnricher(enrichers ...Enricher) T
}

type QueryReader interface {
	First(dest interface{}, conds ...interface{}) error
	Take(dest interface{}, conds ...interface{}) error
	Last(dest interface{}, conds ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	Count(count *int64) error
	Pluck(column string, dest interface{}) error
	Row() *sql.Row
	Rows() (*sql.Rows, error)
}

type QueryWriter interface {
	Create(value interface{}) (int64, error)
	CreateInBatches(value interface{}, batchSize int) (int64, error)
	Save(value interface{}) (int64, error)
	FirstOrInit(dest interface{}, conds ...interface{}) (int64, error)
	FirstOrCreate(dest interface{}, conds ...interface{}) (int64, error)
	Update(column string, value interface{}) (int64, error)
	Updates(values interface{}) (int64, error)
	UpdateColumn(column string, value interface{}) (int64, error)
	UpdateColumns(values interface{}) (int64, error)
	Delete(value interface{}, conds ...interface{}) (int64, error)
	Scan(dest interface{}) (int64, error)
	Exec(sql string, values ...interface{}) (int64, error)
}
