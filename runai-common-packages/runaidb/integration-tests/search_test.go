/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrationtests

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/dbenrich"
	"github.com/stretchr/testify/require"
)

type searchDoc struct {
	ID          int    `gorm:"primaryKey;column:id"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
}

func (searchDoc) TableName() string { return "search_docs" }

func seedSearchDocs(t *testing.T) {
	_, err := runaidb.DB().Exec("DELETE FROM test_schema.search_docs")
	require.NoError(t, err)

	docs := []searchDoc{
		{Name: "engineering", Description: "machine learning platform"},
		{Name: "alice", Description: "data scientist"},
		{Name: "bob", Description: "backend engineer"},
	}
	for i := range docs {
		_, err := runaidb.DB().Create(&docs[i])
		require.NoError(t, err)
	}
}

func TestTSSearch_PrefixMatch(t *testing.T) {
	seedSearchDocs(t)

	var out []searchDoc
	err := runaidb.DB().Table("search_docs").
		WithEnricher(dbenrich.TSSearch("eng", "name")).
		Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "engineering", out[0].Name)
}

func TestTSSearch_OrAcrossColumns(t *testing.T) {
	seedSearchDocs(t)

	var out []searchDoc
	err := runaidb.DB().Table("search_docs").
		WithEnricher(dbenrich.TSSearch("backend", "name", "description")).
		Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "bob", out[0].Name)
}

func TestTSSearch_CaseInsensitivePrefix(t *testing.T) {
	seedSearchDocs(t)

	var out []searchDoc
	err := runaidb.DB().Table("search_docs").
		WithEnricher(dbenrich.TSSearch("ENG", "name")).
		Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "engineering", out[0].Name)
}

func TestTSSearch_NoMatch(t *testing.T) {
	seedSearchDocs(t)

	var out []searchDoc
	err := runaidb.DB().Table("search_docs").
		WithEnricher(dbenrich.TSSearch("nonexistentterm", "name", "description")).
		Find(&out)
	require.NoError(t, err)
	require.Empty(t, out)
}
