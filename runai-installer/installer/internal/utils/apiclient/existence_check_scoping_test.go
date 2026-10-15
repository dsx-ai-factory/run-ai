/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	runaiclient "github.com/dsx-ai-factory/run-ai/runai-installer/clients/pkg/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient returns a RunAI client pointed at a test server that answers the
// health check and delegates every other request to handler. The handlers below
// record the raw query so the test can assert what the client asked for, not
// only what it did with the answer.
func newTestClient(t *testing.T, handler http.HandlerFunc) *runaiclient.RunAIClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_version" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":"test"}`))

			return
		}

		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	client, err := runaiclient.NewClient(srv.URL, runaiclient.Credentials{Token: "test-token"})
	require.NoError(t, err)

	return client
}

func jsonResponse(t *testing.T, body string) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

// A project with a different name is not the project that was asked for: the
// caller uses a result to decide the project already exists and skips creating
// it, so an unrelated project must not be returned.
func TestGetProjectByNameRejectsAnUnrelatedProject(t *testing.T) {
	var rawQuery string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		jsonResponse(t, `{"projects":[{"id":"alpha-id","name":"alpha","clusterId":"cluster-a"}]}`)(w, r)
	})

	project, _, err := NewOrgUnitServiceClientFromRunAI(client).
		GetProjectByName(context.Background(), "cluster-a", "test-project")

	require.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "test-project")
	assert.Contains(t, rawQuery, "clusterId", "the lookup must be scoped to the cluster")
	assert.Contains(t, rawQuery, "name", "the lookup must be scoped to the project name")
}

func TestGetProjectByNameReturnsTheMatchingProject(t *testing.T) {
	client := newTestClient(t, jsonResponse(t,
		`{"projects":[{"id":"test-project-id","name":"test-project","clusterId":"cluster-a"}]}`))

	project, _, err := NewOrgUnitServiceClientFromRunAI(client).
		GetProjectByName(context.Background(), "cluster-a", "test-project")

	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, "test-project-id", project.GetId())
}

// An access rule the user holds elsewhere is not the rule the installer would
// create, so it must not be reported as already existing.
func TestAccessRuleExistsRejectsARuleForAnotherScope(t *testing.T) {
	var rawQuery string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		jsonResponse(t, `{"accessRules":[{"subjectId":"user@example.com","roleId":7,"scopeId":"other-project","scopeType":"project","clusterId":"cluster-a"}]}`)(w, r)
	})

	exists, err := NewAuthServiceClient(client, nil).
		AccessRuleExists(context.Background(), "user@example.com", 7, "cluster-a", "target-project")

	require.NoError(t, err)
	assert.False(t, exists)
	assert.Contains(t, rawQuery, "roleId", "the lookup must be scoped to the role")
	assert.Contains(t, rawQuery, "scopeId", "the lookup must be scoped to the project")
	assert.Contains(t, rawQuery, "clusterId", "the lookup must be scoped to the cluster")
}

func TestAccessRuleExistsRejectsTheSameScopeInAnotherCluster(t *testing.T) {
	client := newTestClient(t, jsonResponse(t,
		`{"accessRules":[{"subjectId":"user@example.com","roleId":7,"scopeId":"target-project","scopeType":"project","clusterId":"other-cluster"}]}`))

	exists, err := NewAuthServiceClient(client, nil).
		AccessRuleExists(context.Background(), "user@example.com", 7, "cluster-a", "target-project")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestAccessRuleExistsFindsTheMatchingRule(t *testing.T) {
	client := newTestClient(t, jsonResponse(t,
		`{"accessRules":[{"subjectId":"user@example.com","roleId":7,"scopeId":"target-project","scopeType":"project","clusterId":"cluster-a"}]}`))

	exists, err := NewAuthServiceClient(client, nil).
		AccessRuleExists(context.Background(), "user@example.com", 7, "cluster-a", "target-project")

	require.NoError(t, err)
	assert.True(t, exists)
}

// "subjectId=@x" is a contains match, so otheruser@example.com answers a lookup
// for user@example.com. The rule has to be matched on the subject itself.
func TestAccessRuleExistsRejectsARuleForAnotherSubject(t *testing.T) {
	var rawQuery string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		jsonResponse(t, `{"accessRules":[{"subjectId":"otheruser@example.com","roleId":7,"scopeId":"target-project","scopeType":"project","clusterId":"cluster-a"}]}`)(w, r)
	})

	exists, err := NewAuthServiceClient(client, nil).
		AccessRuleExists(context.Background(), "user@example.com", 7, "cluster-a", "target-project")

	require.NoError(t, err)
	assert.False(t, exists)
	assert.Contains(t, rawQuery, "subjectId", "the lookup must be scoped to the subject")
	assert.Contains(t, rawQuery, "scopeType", "the lookup must be scoped to the project scope type")
}

func TestAccessRuleExistsRejectsAnotherScopeTypeOnTheSameId(t *testing.T) {
	client := newTestClient(t, jsonResponse(t,
		`{"accessRules":[{"subjectId":"user@example.com","roleId":7,"scopeId":"target-project","scopeType":"department","clusterId":"cluster-a"}]}`))

	exists, err := NewAuthServiceClient(client, nil).
		AccessRuleExists(context.Background(), "user@example.com", 7, "cluster-a", "target-project")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestAccessRuleExistsRejectsARuleWithoutACluster(t *testing.T) {
	client := newTestClient(t, jsonResponse(t,
		`{"accessRules":[{"subjectId":"user@example.com","roleId":7,"scopeId":"target-project","scopeType":"project"}]}`))

	exists, err := NewAuthServiceClient(client, nil).
		AccessRuleExists(context.Background(), "user@example.com", 7, "cluster-a", "target-project")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGetProjectByNameRejectsAProjectFromAnotherCluster(t *testing.T) {
	client := newTestClient(t, jsonResponse(t,
		`{"projects":[{"id":"other-id","name":"test-project","clusterId":"cluster-b"}]}`))

	project, _, err := NewOrgUnitServiceClientFromRunAI(client).
		GetProjectByName(context.Background(), "cluster-a", "test-project")

	require.Error(t, err)
	assert.Nil(t, project)
}
