/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbenrich_test

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaidb/dbenrich"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

type doc struct {
	ID   int    `gorm:"column:id"`
	Name string `gorm:"column:name"`
	Body string `gorm:"column:body"`
}

func (doc) TableName() string { return "docs" }

func TestTSSearch_SingleColumn(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM docs WHERE (name @@ to_tsquery('english', ? || ':*'))", "alice").
		Returns([]doc{{ID: 1, Name: "alice"}})

	var out []doc
	err := db.Table("docs").WithEnricher(dbenrich.TSSearch("alice", "name")).Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestTSSearch_MultipleColumns(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM docs WHERE (name @@ to_tsquery('english', ? || ':*') OR body @@ to_tsquery('english', ? || ':*'))",
		"alice", "alice",
	).Returns([]doc{{ID: 1, Name: "alice"}})

	var out []doc
	err := db.Table("docs").WithEnricher(dbenrich.TSSearch("alice", "name", "body")).Find(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestTSSearch_SanitizesSpecialCharacters(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM docs WHERE (name @@ to_tsquery('english', ? || ':*'))", "test-foo").
		Returns([]doc{})

	var out []doc
	err := db.Table("docs").WithEnricher(dbenrich.TSSearch("test foo!#$", "name")).Find(&out)
	require.NoError(t, err)
}

func TestTSSearch_NoColumns_NoWhereClause(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM docs").Returns([]doc{})

	var out []doc
	err := db.Table("docs").WithEnricher(dbenrich.TSSearch("alice")).Find(&out)
	require.NoError(t, err)
}

func TestTSSearch_EmptyAfterSanitize_NoWhereClause(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM docs").Returns([]doc{})

	var out []doc
	err := db.Table("docs").WithEnricher(dbenrich.TSSearch("!#$%", "name")).Find(&out)
	require.NoError(t, err)
}
