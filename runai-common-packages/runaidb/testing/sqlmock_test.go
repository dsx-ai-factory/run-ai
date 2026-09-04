/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing_test

import (
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
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

func TestNewMocker(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)
	require.NotNil(t, db)
	require.NotNil(t, mocker)
}

func TestExpectQuery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM projects WHERE (projects.cluster_id = ? AND projects.status = ?) AND projects.deleted_at IS NULL ORDER BY name ASC", "cluster-1", "active").
		Returns([]projectModel{
			{Name: "proj-a", ClusterID: "cluster-1", Status: "active"},
			{Name: "proj-b", ClusterID: "cluster-1", Status: "active"},
		})

	var results []projectModel
	err := db.Table("projects").
		Where(projectModel{ClusterID: "cluster-1", Status: "active"}).
		Order("name ASC").
		Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, "proj-a", results[0].Name)
	require.Equal(t, "proj-b", results[1].Name)
}

func TestExpectQuery_UpdateReturning_BehavesLikeQuery(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectTransaction().Begin()
	mocker.ExpectQuery(
		"UPDATE projects SET updated_at=? WHERE (projects.cluster_id = ? AND projects.status = ?) AND projects.deleted_at IS NULL RETURNING *",
		dbtesting.AnyTime(), "cluster-1", "active",
	).Returns([]projectModel{{Name: "proj-a", ClusterID: "cluster-1", Status: "active"}})
	mocker.ExpectTransaction().Commit()

	p := &projectModel{}
	_, err := db.Table("projects").
		Where(projectModel{ClusterID: "cluster-1", Status: "active"}).
		Clauses(clause.Returning{}).
		Updates(p)
	require.NoError(t, err)
	require.Equal(t, "proj-a", p.Name)
	require.Equal(t, "cluster-1", p.ClusterID)
	require.Equal(t, "active", p.Status)
}

func TestExpectQuery_Error(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM projects WHERE (projects.cluster_id = ? AND projects.status = ?) AND projects.deleted_at IS NULL", "cluster-1", "active").
		ReturnsError(errors.New("db error"))

	var results []projectModel
	err := db.Table("projects").
		Where(projectModel{ClusterID: "cluster-1", Status: "active"}).
		Find(&results)
	require.EqualError(t, err, "db error")
}

func TestExpectExec(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec("UPDATE projects SET deleted_at=? WHERE (projects.cluster_id = ? AND projects.status = ?) AND projects.deleted_at IS NULL", dbtesting.AnyTime(), "cluster-1", "active").Returns(2)
	mocker.ExpectTransaction().Commit()

	rowsAffected, err := db.Table("projects").Where(projectModel{ClusterID: "cluster-1", Status: "active"}).Delete(&projectModel{})
	require.Equal(t, int64(2), rowsAffected)
	require.NoError(t, err)
}

func TestExpectExec_Unscoped_Delete(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec("DELETE FROM projects WHERE projects.cluster_id = ?", "cluster-1").Returns(1)
	mocker.ExpectTransaction().Commit()

	rowsAffected, err := db.Table("projects").Where(projectModel{ClusterID: "cluster-1"}).Unscoped().Delete(&projectModel{})
	require.Equal(t, int64(1), rowsAffected)
	require.NoError(t, err)
}

func TestExpectExec_SkipDefaultTransaction(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectExec("DELETE FROM projects WHERE projects.cluster_id = ?", "cluster-1").Returns(1)

	_, err := db.Session(&gorm.Session{SkipDefaultTransaction: true}).
		Table("projects").
		Where(projectModel{ClusterID: "cluster-1"}).
		Unscoped().
		Delete(&projectModel{})
	require.NoError(t, err)
}

func TestExpectExec_WithSavepoint(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec("SAVEPOINT sp1").Returns(0)
	mocker.ExpectExec("DELETE FROM projects WHERE projects.cluster_id = ?", "cluster-1").Returns(1)
	mocker.ExpectTransaction().Commit()

	err := db.Transaction(func(tx orm.DB) error {
		tx = tx.SavePoint("sp1")
		require.NoError(t, tx.Actual().Error)

		_, err := tx.Table("projects").
			Where(projectModel{ClusterID: "cluster-1"}).
			Unscoped().
			Delete(&projectModel{})
		return err
	})
	require.NoError(t, err)
}

func TestExpectExec_AnyArg(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectExec("UPDATE users SET name = ? WHERE id = ?", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)

	rows, err := db.Exec("UPDATE users SET name = ? WHERE id = ?", "Bob", 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)
}

func TestExpectExec_Error(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec("UPDATE projects SET deleted_at=? WHERE (projects.cluster_id = ? AND projects.status = ?) AND projects.deleted_at IS NULL", dbtesting.AnyTime(), "cluster-1", "active").
		ReturnsError(errors.New("db error"))
	mocker.ExpectTransaction().Rollback()

	_, err := db.Table("projects").Where(projectModel{ClusterID: "cluster-1", Status: "active"}).Delete(&projectModel{})
	require.EqualError(t, err, "db error")
}

func TestExpectClose(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectClose()

	err := db.Close()
	require.NoError(t, err)
}

type testUser struct {
	ID   uint   `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

func TestExpectQuery_Raw(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM users WHERE id = ?", 1).
		Returns([]testUser{{ID: 1, Name: "Alice"}})

	var user testUser
	_, err := db.Raw("SELECT * FROM users WHERE id = ?", 1).Scan(&user)
	require.NoError(t, err)
	require.Equal(t, uint(1), user.ID)
	require.Equal(t, "Alice", user.Name)
}

func TestExpectQuery_Raw_Empty(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT * FROM users").Returns([]testUser{})

	var users []testUser
	_, err := db.Raw("SELECT * FROM users").Scan(&users)
	require.NoError(t, err)
	require.Empty(t, users)
}

func TestExpectExec_Raw(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectExec("UPDATE users SET name = ? WHERE id = ?", "Bob", 1).Returns(1)

	rows, err := db.Exec("UPDATE users SET name = ? WHERE id = ?", "Bob", 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)
}

func TestQueryMatcher_CaseInsensitive(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("select * from users").Returns([]testUser{{ID: 1, Name: "Alice"}})

	var users []testUser
	_, err := db.Raw("SELECT * FROM users").Scan(&users)
	require.NoError(t, err)
}

type serviceLabels struct {
	Team string `json:"team"`
	Tier string `json:"tier"`
}

type servicePort struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type serviceModel struct {
	gorm.Model
	Name   string        `gorm:"column:name"`
	Labels serviceLabels `gorm:"column:labels;serializer:json"`
	Ports  []servicePort `gorm:"column:ports;serializer:json"`
}

func (serviceModel) TableName() string { return "services" }

// serializer:json columns (a struct and a slice-of-struct) round-trip through
// the mock: convertToMockRows marshals them to JSON bytes so gorm's serializer
// can unmarshal them back into the struct on scan.
func TestExpectQuery_JsonSerializerColumns(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery(
		"SELECT * FROM services WHERE name = ? AND services.deleted_at IS NULL",
		"svc-a",
	).Returns([]serviceModel{{
		Name:   "svc-a",
		Labels: serviceLabels{Team: "core", Tier: "gold"},
		Ports:  []servicePort{{Name: "http", Port: 80}, {Name: "grpc", Port: 8080}},
	}})

	var results []serviceModel
	err := db.Table("services").Where("name = ?", "svc-a").Find(&results)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, serviceLabels{Team: "core", Tier: "gold"}, results[0].Labels)
	require.Equal(t, []servicePort{{Name: "http", Port: 80}, {Name: "grpc", Port: 8080}}, results[0].Ports)
}
