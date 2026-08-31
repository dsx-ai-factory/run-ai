/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/database/models"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/slack"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	sqlInsertSlackCredential = `INSERT INTO slack_credentials (id,app_id,client_id,client_secret,created_at,updated_at) VALUES (?,?,?,?,?,?) ON CONFLICT (id) DO UPDATE SET updated_at=?,app_id=excluded.app_id,client_id=excluded.client_id,client_secret=excluded.client_secret RETURNING id`
	sqlSelectSlackCredential = `SELECT * FROM slack_credentials ORDER BY slack_credentials.id LIMIT ?`
	sqlDeleteSlackCredential = `DELETE FROM slack_credentials WHERE slack_credentials.id = ?`
)

type insertedID struct {
	ID int16
}

func TestSlackCredentials_Save(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectQuery(sqlInsertSlackCredential,
		dbtesting.AnyArg(), "A1", "C1", dbtesting.AnyArg(), dbtesting.AnyTime(), dbtesting.AnyTime(), dbtesting.AnyTime(),
	).Returns([]insertedID{{ID: slackCredentialId}})
	mocker.ExpectTransaction().Commit()

	creds := &slack.SlackCredentials{AppId: "A1", ClientId: "C1", ClientSecret: "secret"}
	require.NoError(t, NewSlackCredentials().Save(context.Background(), creds))
}

func TestSlackCredentials_Save_ReturnsInsertError(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectQuery(sqlInsertSlackCredential).ReturnsError(errors.New("insert failed"))
	mocker.ExpectTransaction().Rollback()

	creds := &slack.SlackCredentials{AppId: "A1", ClientId: "C1", ClientSecret: "secret"}
	require.ErrorContains(t, NewSlackCredentials().Save(context.Background(), creds), "insert failed")
}

func TestSlackCredentials_Get(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectSlackCredential).Returns([]models.SlackCredential{
		{Id: slackCredentialId, AppId: "A1", ClientId: "C1", ClientSecret: dbtesting.EncryptForTest(t, "secret")},
	})

	got, err := NewSlackCredentials().Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, &slack.SlackCredentials{AppId: "A1", ClientId: "C1", ClientSecret: "secret"}, got)
}

func TestSlackCredentials_Get_ReturnsErrMissingAppCredentials(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectQuery(sqlSelectSlackCredential).Returns([]models.SlackCredential{})

	got, err := NewSlackCredentials().Get(context.Background())
	require.Nil(t, got)
	require.ErrorIs(t, err, slack.ErrMissingAppCredentials)
}

func TestSlackCredentials_Delete(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteSlackCredential).Returns(1)
	mocker.ExpectTransaction().Commit()

	require.NoError(t, NewSlackCredentials().Delete(context.Background()))
}

func TestSlackCredentials_Delete_ReturnsDeleteError(t *testing.T) {
	mocker := dbtesting.SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	mocker.ExpectExec(sqlDeleteSlackCredential).ReturnsError(errors.New("delete failed"))
	mocker.ExpectTransaction().Rollback()

	require.ErrorContains(t, NewSlackCredentials().Delete(context.Background()), "delete failed")
}
