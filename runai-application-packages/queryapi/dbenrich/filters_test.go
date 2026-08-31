/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbenrich_test

import (
	"testing"

	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

type item struct {
	ID   int    `gorm:"column:id"`
	Name string `gorm:"column:name"`
	Age  int    `gorm:"column:age"`
}

func (item) TableName() string { return "items" }

func TestFilters_SingleFilter(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items WHERE name = ?", "test").
		Returns([]item{{ID: 1, Name: "test"}})

	var out []item
	err := dbenrich.Filters([]queryapi.Filter{
		{Field: "name", Operator: "=", ValueFormat: "%s", Values: []string{"test"}},
	})(db.Table("items")).Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestFilters_MultipleFilters(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items WHERE name = ? AND age >= ?", "test", "18").
		Returns([]item{})

	var out []item
	err := dbenrich.Filters(
		[]queryapi.Filter{
			{Field: "name", Operator: "=", ValueFormat: "%s", Values: []string{"test"}},
			{Field: "age", Operator: ">=", ValueFormat: "%s", Values: []string{"18"}},
		},
	)(db.Table("items")).Find(&out)
	require.NoError(t, err)
}

func TestFilters_OrValues(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items WHERE (name = ? OR name = ?)", "test", "admin").
		Returns([]item{})

	var out []item
	err := dbenrich.Filters([]queryapi.Filter{
		{Field: "name", Operator: "=", ValueFormat: "%s", Values: []string{"test", "admin"}},
	})(db.Table("items")).Find(&out)
	require.NoError(t, err)
}

func TestFilters_LikeOperators(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM items WHERE name ILIKE ? AND name NOT ILIKE ? AND name ILIKE ? AND name ILIKE ?",
		"%alice%", "%alice%", "alice%", "%alice",
	).Returns([]item{})

	var out []item
	err := dbenrich.Filters(
		[]queryapi.Filter{
			{Field: "name", Operator: "ILIKE", ValueFormat: "%%%s%%", Values: []string{"alice"}},
			{Field: "name", Operator: "NOT ILIKE", ValueFormat: "%%%s%%", Values: []string{"alice"}},
			{Field: "name", Operator: "ILIKE", ValueFormat: "%s%%", Values: []string{"alice"}},
			{Field: "name", Operator: "ILIKE", ValueFormat: "%%%s", Values: []string{"alice"}},
		},
	)(db.Table("items")).Find(&out)
	require.NoError(t, err)
}

func TestFilters_NotEqualOperator(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items WHERE name is distinct from ?", "alice").
		Returns([]item{})

	var out []item
	err := dbenrich.Filters([]queryapi.Filter{
		{Field: "name", Operator: "is distinct from", ValueFormat: "%s", Values: []string{"alice"}},
	})(db.Table("items")).Find(&out)
	require.NoError(t, err)
}

func TestFilters_EmptyValuesSkipped(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items").Returns([]item{})

	var out []item
	err := dbenrich.Filters([]queryapi.Filter{
		{Field: "name", Operator: "=", ValueFormat: "%s", Values: nil},
	})(db.Table("items")).Find(&out)
	require.NoError(t, err)
}

func TestFilters_WithExistingWhere(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items WHERE tenant_id = ? AND (name = ? OR name = ?)", 2, "test", "admin").
		Returns([]item{})

	var out []item
	err := dbenrich.Filters([]queryapi.Filter{
		{Field: "name", Operator: "=", ValueFormat: "%s", Values: []string{"test", "admin"}},
	})(db.Table("items").Where("tenant_id = ?", 2)).Find(&out)
	require.NoError(t, err)
}
