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

type runaiReadDB struct {
	inner DB
}

func NewReadDB(gormDB *gorm.DB, dialector gorm.ErrorTranslator) ReadDB {
	return wrapReadDB(NewDB(gormDB, dialector))
}

func AsReadDB(db DB) ReadDB {
	return wrapReadDB(db)
}

func wrapReadDB(db DB) ReadDB {
	if db == nil {
		return nil
	}
	return &runaiReadDB{inner: db}
}

func (db *runaiReadDB) Actual() *gorm.DB {
	return db.inner.Actual()
}

func (db *runaiReadDB) Close() error {
	return db.inner.Close()
}

func (db *runaiReadDB) WithContext(ctx context.Context) ReadDB {
	return wrapReadDB(db.inner.WithContext(ctx))
}

func (db *runaiReadDB) Session(config *Session) ReadDB {
	return wrapReadDB(db.inner.Session(config))
}

func (db *runaiReadDB) Model(value interface{}) ReadDB {
	return wrapReadDB(db.inner.Model(value))
}

func (db *runaiReadDB) Clauses(conds ...clause.Expression) ReadDB {
	return wrapReadDB(db.inner.Clauses(conds...))
}

func (db *runaiReadDB) Table(name string, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Table(name, args...))
}

func (db *runaiReadDB) Distinct(args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Distinct(args...))
}

func (db *runaiReadDB) Select(query interface{}, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Select(query, args...))
}

func (db *runaiReadDB) Omit(columns ...string) ReadDB {
	return wrapReadDB(db.inner.Omit(columns...))
}

func (db *runaiReadDB) Where(query interface{}, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Where(query, args...))
}

func (db *runaiReadDB) Not(query interface{}, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Not(query, args...))
}

func (db *runaiReadDB) Or(query interface{}, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Or(query, args...))
}

func (db *runaiReadDB) Joins(query string, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Joins(query, args...))
}

func (db *runaiReadDB) InnerJoins(query string, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.InnerJoins(query, args...))
}

func (db *runaiReadDB) Group(name string) ReadDB {
	return wrapReadDB(db.inner.Group(name))
}

func (db *runaiReadDB) Having(query interface{}, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Having(query, args...))
}

func (db *runaiReadDB) Order(value interface{}) ReadDB {
	return wrapReadDB(db.inner.Order(value))
}

func (db *runaiReadDB) Limit(limit int) ReadDB {
	return wrapReadDB(db.inner.Limit(limit))
}

func (db *runaiReadDB) Offset(offset int) ReadDB {
	return wrapReadDB(db.inner.Offset(offset))
}

func (db *runaiReadDB) Preload(query string, args ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Preload(query, args...))
}

func (db *runaiReadDB) Unscoped() ReadDB {
	return wrapReadDB(db.inner.Unscoped())
}

func (db *runaiReadDB) Raw(sql string, values ...interface{}) ReadDB {
	return wrapReadDB(db.inner.Raw(sql, values...))
}

func (db *runaiReadDB) WithEnricher(enrichers ...Enricher) ReadDB {
	return wrapReadDB(db.inner.WithEnricher(enrichers...))
}

func (db *runaiReadDB) First(dest interface{}, conds ...interface{}) error {
	return db.inner.First(dest, conds...)
}

func (db *runaiReadDB) Take(dest interface{}, conds ...interface{}) error {
	return db.inner.Take(dest, conds...)
}

func (db *runaiReadDB) Last(dest interface{}, conds ...interface{}) error {
	return db.inner.Last(dest, conds...)
}

func (db *runaiReadDB) Find(dest interface{}, conds ...interface{}) error {
	return db.inner.Find(dest, conds...)
}

func (db *runaiReadDB) Count(count *int64) error {
	return db.inner.Count(count)
}

func (db *runaiReadDB) Pluck(column string, dest interface{}) error {
	return db.inner.Pluck(column, dest)
}

func (db *runaiReadDB) Row() *sql.Row {
	return db.inner.Row()
}

func (db *runaiReadDB) Rows() (*sql.Rows, error) {
	return db.inner.Rows()
}
