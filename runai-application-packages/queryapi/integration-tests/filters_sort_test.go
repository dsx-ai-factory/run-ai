/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integration_tests

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/require"
)

type item struct {
	ID        int       `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at;type:date"`
}

func (item) TableName() string { return "items" }

var testMapping = queryapi.APIFieldsMapping{
	"name":      queryapi.NewFieldMapping("name"),
	"createdAt": queryapi.NewFieldMapping("created_at", queryapi.DateValidator{}),
}

func seedItems(t *testing.T) {
	t.Helper()

	_, err := runaidb.DB().Exec("DELETE FROM test_schema.items")
	require.NoError(t, err)

	items := []item{
		{Name: "alice", CreatedAt: mustParseDate(t, "2023-06-01")},
		{Name: "bob", CreatedAt: mustParseDate(t, "2024-01-01")},
		{Name: "charlie", CreatedAt: mustParseDate(t, "2022-12-31")},
	}
	for i := range items {
		_, err := runaidb.DB().Create(&items[i])
		require.NoError(t, err)
	}
}

func mustParseDate(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse("2006-01-02", value)
	require.NoError(t, err)
	return parsed
}

func queryItems(t *testing.T, filterBy []string, sortField, sortOrder string) []item {
	t.Helper()

	parser := queryapi.NewParser(testMapping)
	filters, err := parser.Filters(filterBy)
	require.NoError(t, err)

	sortColumn, err := parser.GetMapping(sortField)
	require.NoError(t, err)

	var out []item
	enricher := dbenrich.Filters(filters)
	sorter := dbenrich.SortBy(sortColumn, sortOrder)
	err = runaidb.DB().Table("items").WithEnricher(enricher).WithEnricher(sorter).Find(&out)
	require.NoError(t, err)

	return out
}

func TestFilters_OrValues(t *testing.T) {
	seedItems(t)

	out := queryItems(t, []string{"name==alice||bob"}, "name", "asc")
	require.Len(t, out, 2)
	require.Equal(t, "alice", out[0].Name)
	require.Equal(t, "bob", out[1].Name)
}

func TestFilters_DateComparison(t *testing.T) {
	seedItems(t)

	out := queryItems(t, []string{"createdAt>=2023-01-01"}, "createdAt", "asc")
	require.Len(t, out, 2)
	require.Equal(t, "alice", out[0].Name)
	require.Equal(t, "bob", out[1].Name)
}

func TestFilters_Contains(t *testing.T) {
	seedItems(t)

	parser := queryapi.NewParser(testMapping)
	filters, err := parser.Filters([]string{"name=@ali"})
	require.NoError(t, err)

	var out []item
	err = runaidb.DB().Table("items").WithEnricher(dbenrich.Filters(filters)).Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "alice", out[0].Name)
}

func TestFilters_NotEqual(t *testing.T) {
	seedItems(t)

	parser := queryapi.NewParser(testMapping)
	filters, err := parser.Filters([]string{"name!=alice"})
	require.NoError(t, err)

	var out []item
	err = runaidb.DB().Table("items").WithEnricher(dbenrich.Filters(filters)).Order("name ASC").Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, "bob", out[0].Name)
	require.Equal(t, "charlie", out[1].Name)
}

func TestFiltersAndSort_Combined(t *testing.T) {
	seedItems(t)

	out := queryItems(
		t,
		[]string{"name==alice||bob", "createdAt>=2023-01-01"},
		"createdAt",
		"desc",
	)
	require.Len(t, out, 2)
	require.Equal(t, "bob", out[0].Name)
	require.Equal(t, "alice", out[1].Name)
}

func TestSortBy_DescNullsLast(t *testing.T) {
	seedItems(t)

	parser := queryapi.NewParser(testMapping)
	sortColumn, err := parser.GetMapping("createdAt")
	require.NoError(t, err)

	var out []item
	err = runaidb.DB().Table("items").WithEnricher(dbenrich.SortBy(sortColumn, "desc")).Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 3)
	require.Equal(t, "bob", out[0].Name)
	require.Equal(t, "alice", out[1].Name)
	require.Equal(t, "charlie", out[2].Name)
}
