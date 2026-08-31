/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orm_test

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type projectModel struct {
	gorm.Model
	Name      string `gorm:"column:name"`
	ClusterID string `gorm:"column:cluster_id"`
	Status    string `gorm:"column:status"`
}

func (projectModel) TableName() string { return "projects" }

func TestWithEnricher_SingleEnricher(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM projects WHERE projects.status = ? AND projects.deleted_at IS NULL", "active").
		Returns([]projectModel{
			{Name: "proj-a", Status: "active"},
		})

	statusEnricher := orm.Enricher(func(tx orm.DB) orm.DB {
		return tx.Where("projects.status = ?", "active")
	})

	var results []projectModel
	err := db.Table("projects").WithEnricher(statusEnricher).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestWithEnricher_MultipleEnrichers(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM projects WHERE projects.status = ? AND projects.cluster_id = ? AND projects.deleted_at IS NULL", "active", "cluster-1").
		Returns([]projectModel{
			{Name: "proj-a", Status: "active", ClusterID: "cluster-1"},
		})

	statusEnricher := orm.Enricher(func(tx orm.DB) orm.DB {
		return tx.Where("projects.status = ?", "active")
	})
	clusterEnricher := orm.Enricher(func(tx orm.DB) orm.DB {
		return tx.Where("projects.cluster_id = ?", "cluster-1")
	})

	var results []projectModel
	err := db.Table("projects").WithEnricher(statusEnricher, clusterEnricher).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestWithEnricher_Chainable(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM projects WHERE projects.status = ? AND projects.cluster_id = ? AND projects.deleted_at IS NULL ORDER BY name ASC", "active", "cluster-1").
		Returns([]projectModel{
			{Name: "proj-a", Status: "active", ClusterID: "cluster-1"},
			{Name: "proj-b", Status: "active", ClusterID: "cluster-1"},
		})

	statusEnricher := orm.Enricher(func(tx orm.DB) orm.DB {
		return tx.Where("projects.status = ?", "active")
	})
	clusterEnricher := orm.Enricher(func(tx orm.DB) orm.DB {
		return tx.Where("projects.cluster_id = ?", "cluster-1")
	})

	var results []projectModel
	err := db.Table("projects").
		WithEnricher(statusEnricher).
		WithEnricher(clusterEnricher).
		Order("name ASC").
		Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestWithEnricher_NoEnrichers(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM projects WHERE projects.deleted_at IS NULL").
		Returns([]projectModel{
			{Name: "proj-a"},
		})

	var results []projectModel
	err := db.Table("projects").WithEnricher().Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
}

func TestSelect_Subquery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT name, (SELECT count(*) FROM projects WHERE status = ?) FROM projects WHERE projects.deleted_at IS NULL",
		"active",
	).Returns([]projectModel{
		{Name: "proj-a"},
	})

	subQuery := db.Table("projects").Select("count(*)").Where("status = ?", "active")

	var results []projectModel
	err := db.Table("projects").Select("name, (?)", subQuery).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestTable_Subquery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM (SELECT * FROM projects WHERE status = ?) AS active_projects WHERE active_projects.deleted_at IS NULL",
		"active",
	).Returns([]projectModel{
		{Name: "proj-a", Status: "active"},
	})

	subQuery := db.Table("projects").Where("status = ?", "active")

	var results []projectModel
	err := db.Table("(?) AS active_projects", subQuery).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestWhere_Subquery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM projects WHERE cluster_id = (SELECT cluster_id FROM projects WHERE status = ? ORDER BY name ASC LIMIT ?) AND projects.deleted_at IS NULL",
		"active", 1,
	).Returns([]projectModel{
		{Name: "proj-a", Status: "pending", ClusterID: "cluster-1"},
	})

	subQuery := db.Table("projects").Select("cluster_id").Where("status = ?", "active").Order("name ASC").Limit(1)

	var results []projectModel
	err := db.Table("projects").Where("cluster_id = (?)", subQuery).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestWhere_NestedCondition(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM projects WHERE (status = ? OR cluster_id = ?) AND projects.deleted_at IS NULL",
		"active", "cluster-1",
	).Returns([]projectModel{
		{Name: "proj-a", Status: "active", ClusterID: "cluster-1"},
	})

	condition := db.Where("status = ?", "active").Or("cluster_id = ?", "cluster-1")

	var results []projectModel
	err := db.Table("projects").Where(condition).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestNot_Subquery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM projects WHERE NOT cluster_id = (SELECT cluster_id FROM projects WHERE status = ?) AND projects.deleted_at IS NULL",
		"active",
	).Returns([]projectModel{
		{Name: "proj-a", Status: "pending", ClusterID: "cluster-2"},
	})

	subQuery := db.Table("projects").Select("cluster_id").Where("status = ?", "active")

	var results []projectModel
	err := db.Table("projects").Not("cluster_id = (?)", subQuery).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestNot_NestedCondition(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM projects WHERE NOT (status = ? OR cluster_id = ?) AND projects.deleted_at IS NULL",
		"active", "cluster-1",
	).Returns([]projectModel{
		{Name: "proj-a", Status: "pending", ClusterID: "cluster-2"},
	})

	condition := db.Where("status = ?", "active").Or("cluster_id = ?", "cluster-1")

	var results []projectModel
	err := db.Table("projects").Not(condition).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestOr_Subquery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM projects WHERE (status = ? OR cluster_id = (SELECT cluster_id FROM projects WHERE status = ?)) AND projects.deleted_at IS NULL",
		"pending", "active",
	).Returns([]projectModel{
		{Name: "proj-a", Status: "pending", ClusterID: "cluster-1"},
	})

	subQuery := db.Table("projects").Select("cluster_id").Where("status = ?", "active")

	var results []projectModel
	err := db.Table("projects").Where("status = ?", "pending").Or("cluster_id = (?)", subQuery).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestOr_NestedCondition(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM projects WHERE (name = ? OR (status = ? OR cluster_id = ?)) AND projects.deleted_at IS NULL",
		"proj-a", "active", "cluster-1",
	).Returns([]projectModel{
		{Name: "proj-a", Status: "active", ClusterID: "cluster-1"},
	})

	var results []projectModel
	err := db.Table("projects").Where("name = ?", "proj-a").Or(db.Where("status = ?", "active").Or("cluster_id = ?", "cluster-1")).Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "proj-a", results[0].Name)
}

func TestHaving_Subquery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT cluster_id, count(*) FROM projects WHERE projects.deleted_at IS NULL GROUP BY cluster_id HAVING count(*) > (SELECT count(*) FROM projects WHERE status = ?)",
		"pending",
	).Returns([]projectModel{
		{Name: "proj-a", ClusterID: "cluster-1"},
	})

	var results []projectModel
	err := db.Table("projects").
		Select("cluster_id, count(*)").
		Group("cluster_id").
		Having("count(*) > (?)", db.Table("projects").Select("count(*)").Where("status = ?", "pending")).
		Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
}

func TestHaving_NestedCondition(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT cluster_id, count(*) FROM projects WHERE projects.deleted_at IS NULL GROUP BY cluster_id HAVING count(*) > ? OR cluster_id = ?",
		1, "cluster-1",
	).Returns([]projectModel{
		{Name: "proj-a", ClusterID: "cluster-1"},
	})

	condition := db.Where("count(*) > ?", 1).Or("cluster_id = ?", "cluster-1")

	var results []projectModel
	err := db.Table("projects").
		Select("cluster_id, count(*)").
		Group("cluster_id").
		Having(condition).
		Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
}

func TestJoins_Subquery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT projects.id,projects.created_at,projects.updated_at,projects.deleted_at,projects.name,projects.cluster_id,projects.status "+
			"FROM projects JOIN (SELECT cluster_id FROM projects WHERE status = ?) AS active_clusters ON active_clusters.cluster_id = projects.cluster_id "+
			"WHERE projects.deleted_at IS NULL",
		"active",
	).Returns([]projectModel{
		{Name: "proj-a", ClusterID: "cluster-1"},
	})

	subQuery := db.Table("projects").Select("cluster_id").Where("status = ?", "active")

	var results []projectModel
	err := db.Table("projects").
		Joins("JOIN (?) AS active_clusters ON active_clusters.cluster_id = projects.cluster_id", subQuery).
		Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
}
