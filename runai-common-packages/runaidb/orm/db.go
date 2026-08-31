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

type runaiDB struct {
	gormDB          *gorm.DB
	errorTranslator gorm.ErrorTranslator
}

func NewDB(gormDB *gorm.DB, errorTranslator gorm.ErrorTranslator) DB {
	if errorTranslator == nil {
		errorTranslator = noopErrorTranslator{}
	}
	return &runaiDB{gormDB, errorTranslator}
}

type noopErrorTranslator struct{}

func (noopErrorTranslator) Translate(err error) error {
	return err
}

// Expr creates a raw SQL expression, mirroring gorm.Expr.
func Expr(expr string, args ...interface{}) clause.Expr {
	return clause.Expr{SQL: expr, Vars: args}
}

func (db *runaiDB) Actual() *gorm.DB {
	return db.gormDB
}

func (db *runaiDB) Close() error {
	if db.gormDB == nil {
		return nil
	}
	sqlDB, err := db.gormDB.DB()
	if err != nil {
		return db.errorTranslator.Translate(err)
	}
	return db.errorTranslator.Translate(sqlDB.Close())
}

func (db *runaiDB) WithContext(ctx context.Context) DB {
	return &runaiDB{db.gormDB.WithContext(ctx), db.errorTranslator}
}

func (db *runaiDB) Session(config *Session) DB {
	return &runaiDB{db.gormDB.Session(config), db.errorTranslator}
}

// ---- Transactioner ----

func (db *runaiDB) Transaction(fc func(tx DB) error, opts ...*sql.TxOptions) error {
	return db.errorTranslator.Translate(db.gormDB.Transaction(func(gormTx *gorm.DB) error {
		return fc(&runaiDB{gormTx, db.errorTranslator})
	}, opts...))
}

func (db *runaiDB) SavePoint(name string) DB {
	return &runaiDB{db.gormDB.SavePoint(name), db.errorTranslator}
}

func (db *runaiDB) RollbackTo(name string) DB {
	return &runaiDB{db.gormDB.RollbackTo(name), db.errorTranslator}
}

func (db *runaiDB) Begin() DB {
	return &runaiDB{db.gormDB.Begin(), db.errorTranslator}
}

func (db *runaiDB) Commit() error {
	return db.errorTranslator.Translate(db.gormDB.Commit().Error)
}

func (db *runaiDB) Rollback() error {
	return db.errorTranslator.Translate(db.gormDB.Rollback().Error)
}

// ---- QueryBuilder ----

func (db *runaiDB) Model(value interface{}) DB {
	return &runaiDB{db.gormDB.Model(value), db.errorTranslator}
}

func (db *runaiDB) Clauses(conds ...clause.Expression) DB {
	return &runaiDB{db.gormDB.Clauses(conds...), db.errorTranslator}
}

func (db *runaiDB) Table(name string, args ...interface{}) DB {
	return &runaiDB{db.gormDB.Table(name, unwrapDBArgs(args)...), db.errorTranslator}
}

func (db *runaiDB) Distinct(args ...interface{}) DB {
	return &runaiDB{db.gormDB.Distinct(args...), db.errorTranslator}
}

func (db *runaiDB) Select(query interface{}, args ...interface{}) DB {
	return &runaiDB{db.gormDB.Select(query, unwrapDBArgs(args)...), db.errorTranslator}
}

func (db *runaiDB) Omit(columns ...string) DB {
	return &runaiDB{db.gormDB.Omit(columns...), db.errorTranslator}
}

func (db *runaiDB) Where(query interface{}, args ...interface{}) DB {
	return &runaiDB{db.gormDB.Where(unwrapQuery(query), unwrapDBArgs(args)...), db.errorTranslator}
}

func (db *runaiDB) Not(query interface{}, args ...interface{}) DB {
	return &runaiDB{db.gormDB.Not(unwrapQuery(query), unwrapDBArgs(args)...), db.errorTranslator}
}

func (db *runaiDB) Or(query interface{}, args ...interface{}) DB {
	return &runaiDB{db.gormDB.Or(unwrapQuery(query), unwrapDBArgs(args)...), db.errorTranslator}
}

func (db *runaiDB) Joins(query string, args ...interface{}) DB {
	return &runaiDB{db.gormDB.Joins(query, unwrapDBArgs(args)...), db.errorTranslator}
}

func (db *runaiDB) InnerJoins(query string, args ...interface{}) DB {
	return &runaiDB{db.gormDB.InnerJoins(query, unwrapDBArgs(args)...), db.errorTranslator}
}

func (db *runaiDB) Group(name string) DB {
	return &runaiDB{db.gormDB.Group(name), db.errorTranslator}
}

func (db *runaiDB) Having(query interface{}, args ...interface{}) DB {
	return &runaiDB{db.gormDB.Having(unwrapQuery(query), unwrapDBArgs(args)...), db.errorTranslator}
}

func (db *runaiDB) Order(value interface{}) DB {
	return &runaiDB{db.gormDB.Order(value), db.errorTranslator}
}

func (db *runaiDB) Limit(limit int) DB {
	return &runaiDB{db.gormDB.Limit(limit), db.errorTranslator}
}

func (db *runaiDB) Offset(offset int) DB {
	return &runaiDB{db.gormDB.Offset(offset), db.errorTranslator}
}

func (db *runaiDB) Preload(query string, args ...interface{}) DB {
	return &runaiDB{db.gormDB.Preload(query, args...), db.errorTranslator}
}

func (db *runaiDB) Assign(attrs ...interface{}) DB {
	return &runaiDB{db.gormDB.Assign(attrs...), db.errorTranslator}
}

func (db *runaiDB) Unscoped() DB {
	return &runaiDB{db.gormDB.Unscoped(), db.errorTranslator}
}

func (db *runaiDB) Raw(sql string, values ...interface{}) DB {
	return &runaiDB{db.gormDB.Raw(sql, unwrapDBArgs(values)...), db.errorTranslator}
}

func (db *runaiDB) WithEnricher(enrichers ...Enricher) DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, len(enrichers))
	for i, e := range enrichers {
		scopes[i] = func(tx *gorm.DB) *gorm.DB {
			return e(NewDB(tx, db.errorTranslator)).Actual()
		}
	}
	return &runaiDB{db.gormDB.Scopes(scopes...), db.errorTranslator}
}

// ---- QueryReader ----

func (db *runaiDB) First(dest interface{}, conds ...interface{}) error {
	return db.errorTranslator.Translate(db.gormDB.First(dest, conds...).Error)
}

func (db *runaiDB) Take(dest interface{}, conds ...interface{}) error {
	return db.errorTranslator.Translate(db.gormDB.Take(dest, conds...).Error)
}

func (db *runaiDB) Last(dest interface{}, conds ...interface{}) error {
	return db.errorTranslator.Translate(db.gormDB.Last(dest, conds...).Error)
}

func (db *runaiDB) Find(dest interface{}, conds ...interface{}) error {
	return db.errorTranslator.Translate(db.gormDB.Find(dest, conds...).Error)
}

func (db *runaiDB) Count(count *int64) error {
	return db.errorTranslator.Translate(db.gormDB.Count(count).Error)
}

func (db *runaiDB) Pluck(column string, dest interface{}) error {
	return db.errorTranslator.Translate(db.gormDB.Pluck(column, dest).Error)
}

func (db *runaiDB) Row() *sql.Row {
	return db.gormDB.Row()
}

func (db *runaiDB) Rows() (*sql.Rows, error) {
	rows, err := db.gormDB.Rows()
	return rows, db.errorTranslator.Translate(err)
}

// ---- QueryWriter ----

func (db *runaiDB) Create(value interface{}) (int64, error) {
	r := db.gormDB.Create(value)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) CreateInBatches(value interface{}, batchSize int) (int64, error) {
	r := db.gormDB.CreateInBatches(value, batchSize)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) Save(value interface{}) (int64, error) {
	r := db.gormDB.Save(value)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) FirstOrInit(dest interface{}, conds ...interface{}) (int64, error) {
	r := db.gormDB.FirstOrInit(dest, conds...)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) FirstOrCreate(dest interface{}, conds ...interface{}) (int64, error) {
	r := db.gormDB.FirstOrCreate(dest, conds...)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) Update(column string, value interface{}) (int64, error) {
	r := db.gormDB.Update(column, value)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) Updates(values interface{}) (int64, error) {
	r := db.gormDB.Updates(values)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) UpdateColumn(column string, value interface{}) (int64, error) {
	r := db.gormDB.UpdateColumn(column, value)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) UpdateColumns(values interface{}) (int64, error) {
	r := db.gormDB.UpdateColumns(values)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) Delete(value interface{}, conds ...interface{}) (int64, error) {
	r := db.gormDB.Delete(value, conds...)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) Scan(dest interface{}) (int64, error) {
	r := db.gormDB.Scan(dest)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

func (db *runaiDB) Exec(sql string, values ...interface{}) (int64, error) {
	r := db.gormDB.Exec(sql, values...)
	return r.RowsAffected, db.errorTranslator.Translate(r.Error)
}

// unwrapDBArgs replaces any DB in args with its underlying *gorm.DB, since
// GORM type-asserts subquery args as *gorm.DB wherever it accepts them.
func unwrapDBArgs(args []interface{}) []interface{} {
	for i, arg := range args {
		if db, ok := arg.(DB); ok {
			args[i] = db.Actual()
		}
	}
	return args
}

func unwrapQuery(query interface{}) interface{} {
	if db, ok := query.(DB); ok {
		return db.Actual()
	}
	return query
}
