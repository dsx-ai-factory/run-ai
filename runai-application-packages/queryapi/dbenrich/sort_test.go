/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbenrich_test

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

func TestSortBy_Asc(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items ORDER BY name ASC NULLS FIRST").Returns([]item{})

	var out []item
	err := dbenrich.SortBy("name", "asc")(db.Table("items")).Find(&out)
	require.NoError(t, err)
}

func TestSortBy_Desc(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items ORDER BY created_at DESC NULLS LAST").Returns([]item{})

	var out []item
	err := dbenrich.SortBy("created_at", "desc")(db.Table("items")).Find(&out)
	require.NoError(t, err)
}

func TestSortBy_EmptyColumnNoOrder(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items").Returns([]item{})

	var out []item
	err := dbenrich.SortBy("", "asc")(db.Table("items")).Find(&out)
	require.NoError(t, err)
}

func TestSortBy_ColumnTemplateWithPlaceholder(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM items ORDER BY CASE WHEN phase = 'Active' THEN 0 ELSE 1 END ASC NULLS FIRST").
		Returns([]item{})

	var out []item
	err := dbenrich.SortBy("CASE WHEN phase = 'Active' THEN 0 ELSE 1 END %s", "asc")(db.Table("items")).Find(&out)
	require.NoError(t, err)
}
