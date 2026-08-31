/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Mocker interface {
	ExpectExec(query string, args ...any) ExecExpectation
	ExpectQuery(query string, args ...any) QueryExpectation
	ExpectTransaction() TransactionExpectation
	ExpectClose()
}

type ExecExpectation interface {
	Returns(rowsAffected int64)
	ReturnsError(err error)
}

type QueryExpectation interface {
	Returns(rows any)
	ReturnsError(err error)
}

type TransactionExpectation interface {
	Begin()
	Commit()
	Rollback()
}

type ErrorTranslatorFunc func(error) error

func (f ErrorTranslatorFunc) Translate(err error) error {
	return f(err)
}

func NewGormMocker(t TestingT) (*gorm.DB, Mocker) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(queryMatcher))
	if err != nil {
		t.Errorf("sqlmock: failed to create mock: %v", err)
		return nil, nil
	}

	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB}),
		&gorm.Config{PrepareStmt: false},
	)
	if err != nil {
		t.Errorf("sqlmock: failed to open GORM DB: %v", err)
		_ = sqlDB.Close()
		return nil, nil
	}

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("sqlmock: unfulfilled expectations: %v", err)
		}
		_ = sqlDB.Close()
	})

	return gormDB, &queryMocker{mock: mock}
}

func NewMocker(t TestingT) (orm.DB, Mocker) {
	gormDB, mocker := NewGormMocker(t)
	if gormDB == nil || mocker == nil {
		return nil, nil
	}
	return orm.NewDB(gormDB, ErrorTranslatorFunc(func(err error) error { return err })), mocker
}

type queryMocker struct {
	mock sqlmock.Sqlmock
}

func (m *queryMocker) ExpectClose() {
	m.mock.ExpectClose()
}

func (m *queryMocker) ExpectTransaction() TransactionExpectation {
	return &transactionExpectation{mock: m.mock}
}

func (m *queryMocker) ExpectExec(query string, args ...any) ExecExpectation {
	exp := m.mock.ExpectExec(query)
	if len(args) > 0 {
		exp.WithArgs(toDriverValues(args)...)
	}
	return &execExpectation{exp: exp}
}

func (m *queryMocker) ExpectQuery(query string, args ...any) QueryExpectation {
	exp := m.mock.ExpectQuery(query)
	if len(args) > 0 {
		exp.WithArgs(toDriverValues(args)...)
	}
	return &queryExpectation{exp: exp}
}

type transactionExpectation struct {
	mock sqlmock.Sqlmock
}

func (t *transactionExpectation) Begin() {
	t.mock.ExpectBegin()
}

func (t *transactionExpectation) Commit() {
	t.mock.ExpectCommit()
}

func (t *transactionExpectation) Rollback() {
	t.mock.ExpectRollback()
}

type execExpectation struct {
	exp *sqlmock.ExpectedExec
}

func (e *execExpectation) Returns(rowsAffected int64) {
	e.exp.WillReturnResult(sqlmock.NewResult(0, rowsAffected))
}

func (e *execExpectation) ReturnsError(err error) {
	e.exp.WillReturnError(err)
}

type queryExpectation struct {
	exp *sqlmock.ExpectedQuery
}

func (q *queryExpectation) Returns(rows any) {
	q.exp.WillReturnRows(convertToMockRows(rows))
}

func (q *queryExpectation) ReturnsError(err error) {
	q.exp.WillReturnError(err)
}

var dollarPlaceholder = regexp.MustCompile(`\$\d+`)

var queryMatcher sqlmock.QueryMatcherFunc = func(expected, actualSQL string) error {
	// The actualSQL parameter has $1, $2, etc. placeholders, but our convention is to use ? placeholders so we replace them.
	normalized := dollarPlaceholder.ReplaceAllString(
		strings.Join(strings.Fields(strings.ReplaceAll(actualSQL, `"`, ``)), " "),
		"?",
	)
	if !strings.EqualFold(normalized, expected) {
		return fmt.Errorf("sqlmock: query %q did not match expected %q", normalized, expected)
	}
	return nil
}

func toDriverValues(args []any) []driver.Value {
	vals := make([]driver.Value, len(args))
	for i, a := range args {
		vals[i] = a
	}
	return vals
}

type fieldEntry struct {
	column         string
	index          []int
	jsonSerialized bool // jsonSerialized is true for columns tagged gorm:"serializer:json"
}

// collectFields walks a struct type recursively, flattening anonymous (embedded) structs,
// and returns one fieldEntry per leaf column with its reflect index path for fast value access.
func collectFields(t reflect.Type, ns schema.NamingStrategy, path []int) []fieldEntry {
	var entries []fieldEntry
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Tag.Get("gorm") == "-" {
			continue
		}
		fieldPath := append(append([]int{}, path...), i)
		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		col := ""
		jsonSerialized := false
		for _, part := range strings.Split(f.Tag.Get("gorm"), ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "column:") {
				col = strings.TrimPrefix(part, "column:")
			}
			if part == "serializer:json" {
				jsonSerialized = true
			}
		}
		if !jsonSerialized && ft.Kind() == reflect.Struct {
			if f.Anonymous {
				entries = append(entries, collectFields(ft, ns, fieldPath)...)
			}
			continue
		}
		// If no column name is found, use the field name to get the column name. For example, "CreatedBy" -> "created_by".
		if col == "" {
			col = ns.ColumnName("", f.Name)
		}
		entries = append(entries, fieldEntry{column: col, index: fieldPath, jsonSerialized: jsonSerialized})
	}
	return entries
}

// convertToMockRows reflects over a slice of structs and builds sqlmock.Rows with columns derived from gorm tags.
func convertToMockRows(rows any) *sqlmock.Rows {
	sliceVal := reflect.ValueOf(rows)
	if sliceVal.Kind() != reflect.Slice {
		panic(fmt.Sprintf("sqlmock: convertToMockRows expects a slice, got %T", rows))
	}
	if sliceVal.Len() == 0 {
		return sqlmock.NewRows(nil)
	}

	elemType := sliceVal.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	fields := collectFields(elemType, schema.NamingStrategy{}, nil)
	columns := make([]string, len(fields))
	for i, f := range fields {
		columns[i] = f.column
	}

	mockRows := sqlmock.NewRows(columns)
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		vals := make([]driver.Value, len(fields))
		for j, f := range fields {
			raw := elem.FieldByIndex(f.index).Interface()
			if f.jsonSerialized {
				b, err := json.Marshal(raw)
				if err != nil {
					panic(fmt.Sprintf("sqlmock: marshalling json column %q: %v", f.column, err))
				}
				vals[j] = b
			} else {
				vals[j] = raw
			}
		}
		mockRows.AddRow(vals...)
	}
	return mockRows
}
