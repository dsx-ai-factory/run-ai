/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package access

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authz "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func bptr(b bool) *bool { return &b }

const (
	muaClusterUUID  = "d1405b3f-761a-4d4b-acb6-a00327ca853c"
	permScopesPath  = "/api/v1/authorization/permitted-scopes"
	usersPath       = "/api/v1/users"
	serviceAcctPath = "/api/v1/service-accounts"
	accessRulesPath = "/api/v1/authorization/access-rules"
	rolesPath       = "/api/v2/authorization/roles"
)

// usersSearchHandler answers the GetUsers(search) lookup with one user, so a user subject
// given by id or email resolves to its username.
func usersSearchHandler(id, username string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, []map[string]any{{"id": id, "username": username}})
	}
}

// permTenantHandler grants the listed actions at tenant-wide (system) reach.
func permTenantHandler(grant ...string) http.HandlerFunc {
	allowed := map[string]bool{}
	for _, a := range grant {
		allowed[a] = true
	}
	scope := func(action string) map[string]any {
		sys := allowed[action]
		return map[string]any{"system": sys}
	}
	return func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"create": scope("create"), "read": scope("read"),
			"update": scope("update"), "delete": scope("delete"),
		})
	}
}

func TestManageUserAccessInput_Validate(t *testing.T) {
	roleID := int32(7)
	cases := []struct {
		name    string
		in      manageUserAccessInput
		wantErr string
	}{
		{name: "create_user_happy", in: manageUserAccessInput{Entity: "user", Action: "create", Email: "a@b.com"}},
		{name: "create_sa_happy", in: manageUserAccessInput{Entity: "service-account", Action: "create", Name: "ci-bot"}},
		{name: "delete_user_by_email", in: manageUserAccessInput{Entity: "user", Action: "delete", Email: "a@b.com"}},
		{name: "delete_sa_by_id", in: manageUserAccessInput{Entity: "service-account", Action: "delete", ID: "sa-1"}},
		{name: "create_rule_rolename", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleName: "L1 researcher", ScopeType: "project", ScopeID: "proj-1", ClusterID: muaClusterUUID}},
		{name: "create_rule_roleid_system", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleID: &roleID, ScopeType: "system"}},
		{name: "delete_rule_by_id", in: manageUserAccessInput{Entity: "access-rule", Action: "delete", ID: "42"}},

		{name: "bad_entity", in: manageUserAccessInput{Entity: "team", Action: "create"}, wantErr: "entity must be one of"},
		{name: "application_dropped", in: manageUserAccessInput{Entity: "application", Action: "create", Name: "my-app"}, wantErr: "entity must be one of"},
		{name: "rule_subjecttype_app_rejected", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "app", RoleID: &roleID, ScopeType: "system"}, wantErr: "subjectType must be one of"},
		{name: "bad_action", in: manageUserAccessInput{Entity: "user", Action: "frob"}, wantErr: "action must be one of"},
		{name: "user_create_no_email", in: manageUserAccessInput{Entity: "user", Action: "create"}, wantErr: "email is required"},
		{name: "user_with_name", in: manageUserAccessInput{Entity: "user", Action: "create", Email: "a@b.com", Name: "x"}, wantErr: "use email (not name)"},
		{name: "sa_create_no_name", in: manageUserAccessInput{Entity: "service-account", Action: "create"}, wantErr: "name is required"},
		{name: "sa_with_email", in: manageUserAccessInput{Entity: "service-account", Action: "create", Name: "x", Email: "a@b.com"}, wantErr: "email is only valid for entity=user"},
		{name: "sa_with_names", in: manageUserAccessInput{Entity: "service-account", Action: "create", Name: "x", FirstName: "f"}, wantErr: "firstName and lastName are only valid for entity=user"},
		{name: "identity_with_rule_fields", in: manageUserAccessInput{Entity: "user", Action: "create", Email: "a@b.com", RoleName: "r"}, wantErr: "only valid for entity=access-rule"},
		{name: "user_delete_no_target", in: manageUserAccessInput{Entity: "user", Action: "delete"}, wantErr: "one of id or email"},
		{name: "ignore_missing_on_create", in: manageUserAccessInput{Entity: "user", Action: "create", Email: "a@b.com", IgnoreMissing: true}, wantErr: "ignoreMissing is only valid"},

		{name: "rule_no_subject", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectType: "user", RoleID: &roleID, ScopeType: "system"}, wantErr: "exactly one of subjectId or subjectEmail"},
		{name: "rule_subject_by_email", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectEmail: "a@b.com", SubjectType: "user", RoleID: &roleID, ScopeType: "system"}},
		{name: "rule_subject_email_and_id", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectEmail: "a@b.com", SubjectType: "user", RoleID: &roleID, ScopeType: "system"}, wantErr: "exactly one of subjectId or subjectEmail"},
		{name: "rule_sa_subject_by_name", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectName: "ci-bot", SubjectType: "service-account", RoleID: &roleID, ScopeType: "system"}},
		{name: "rule_user_with_subjectname", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectEmail: "a@b.com", SubjectName: "x", SubjectType: "user", RoleID: &roleID, ScopeType: "system"}, wantErr: "subjectName is for subjectType=service-account"},
		{name: "rule_group_needs_id", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectType: "group", RoleID: &roleID, ScopeType: "system"}, wantErr: "subjectId is required for subjectType=group"},
		{name: "rule_delete_by_tuple", in: manageUserAccessInput{Entity: "access-rule", Action: "delete", SubjectID: "u1", SubjectType: "user", RoleName: "L1 researcher", ScopeType: "system"}},
		{name: "rule_bad_subjecttype", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "robot", RoleID: &roleID, ScopeType: "system"}, wantErr: "subjectType must be one of"},
		{name: "rule_both_roles", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleName: "r", RoleID: &roleID, ScopeType: "system"}, wantErr: "exactly one of roleName or roleId"},
		{name: "rule_no_role", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", ScopeType: "system"}, wantErr: "exactly one of roleName or roleId"},
		{name: "rule_project_no_cluster", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleID: &roleID, ScopeType: "project", ScopeID: "proj-1"}, wantErr: "clusterId or clusterName is required for scopeType=project"},
		{name: "rule_cluster_no_scopeid", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleID: &roleID, ScopeType: "cluster"}, wantErr: "scopeId (cluster ID) or clusterName is required"},
		{name: "rule_cluster_by_name", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleID: &roleID, ScopeType: "cluster", ClusterName: "c1"}},
		{name: "rule_project_by_name", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleID: &roleID, ScopeType: "project", ScopeName: "proj-x", ClusterName: "c1"}},
		{name: "rule_project_name_no_cluster", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleID: &roleID, ScopeType: "project", ScopeName: "proj-x"}, wantErr: "clusterId or clusterName is required"},
		{name: "identity_with_clustername", in: manageUserAccessInput{Entity: "user", Action: "create", Email: "a@b.com", ClusterName: "c1"}, wantErr: "only valid for entity=access-rule"},
		{name: "rule_bad_scopetype", in: manageUserAccessInput{Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user", RoleID: &roleID, ScopeType: "galaxy"}, wantErr: "scopeType must be one of"},
		{name: "rule_delete_no_target", in: manageUserAccessInput{Entity: "access-rule", Action: "delete"}, wantErr: "give its numeric id, or the subject+role+scope tuple"},
		{name: "rule_delete_id_and_tuple", in: manageUserAccessInput{Entity: "access-rule", Action: "delete", ID: "42", RoleName: "r"}, wantErr: "by id alone, or by the subject+role+scope tuple"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestManageUserAccess_CreateUserHappy(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "user-new", "username": "a@b.com", "tempPassword": "s3cr3t",
		})
	})

	out, err := manageUserAccessRun(context.Background(), client,
		manageUserAccessInput{Confirm: true, Entity: "user", Action: "create", Email: "a@b.com", FirstName: "Ada"})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "user-new", out.ID)
	assert.Equal(t, "a@b.com", out.After["username"])
	assert.Equal(t, "s3cr3t", out.After["tempPassword"])
	require.NotEmpty(t, out.Warnings)
	assert.Contains(t, out.Warnings[0], "shown once")
}

func TestManageUserAccess_CreateServiceAccountReturnsSecret(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(serviceAcctPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "sa-new", "name": "ci-bot", "clientId": "ci-bot", "clientSecret": "shhh",
		})
	})

	out, err := manageUserAccessRun(context.Background(), client,
		manageUserAccessInput{Confirm: true, Entity: "service-account", Action: "create", Name: "ci-bot"})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "sa-new", out.ID)
	assert.Equal(t, "shhh", out.After["clientSecret"])
}

func TestManageUserAccess_CreateUserForbidden(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler()) // grant nothing

	_, err := manageUserAccessRun(context.Background(), client,
		manageUserAccessInput{Entity: "user", Action: "create", Email: "a@b.com"})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
}

func TestManageUserAccess_CreateUserDryRun(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("dry-run must not POST")
	})

	out, err := manageUserAccessRun(context.Background(), client,
		manageUserAccessInput{Entity: "user", Action: "create", Email: "a@b.com", DryRun: true})
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.Equal(t, "a@b.com", out.After["username"])
}

func TestManageUserAccess_DeleteUserByEmailIgnoreMissing(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("delete"))
	mux.HandleFunc(usersPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, []map[string]any{}) // no match
	})

	out, err := manageUserAccessRun(context.Background(), client,
		manageUserAccessInput{Entity: "user", Action: "delete", Email: "gone@b.com", IgnoreMissing: true})
	require.NoError(t, err)
	assert.False(t, out.Applied)
	require.NotEmpty(t, out.Warnings)
	assert.Contains(t, out.Warnings[0], "nothing to delete")
}

func TestManageUserAccess_CreateAccessRuleResolvesRole(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath, usersSearchHandler("u1", "u1"))
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 7, "name": "L1 researcher", "description": ""}},
		})
	})
	var posted map[string]any
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &posted))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 101, "subjectId": "u1", "subjectType": "user", "roleId": 7, "roleName": "L1 researcher",
			"scopeId": "proj-1", "scopeType": "project", "scopeName": "ml-team", "clusterId": muaClusterUUID,
		})
	})

	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Confirm: true,
		Entity:  "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user",
		RoleName: "L1 researcher", ScopeType: "project", ScopeID: "proj-1", ClusterID: muaClusterUUID,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "101", out.ID)
	assert.EqualValues(t, 7, posted["roleId"])
	assert.Equal(t, "L1 researcher", out.After["roleName"])
	assert.Empty(t, out.Warnings)
}

func TestManageUserAccess_CreateAccessRuleWarnsDeprecatedRole(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath, usersSearchHandler("u1", "u1"))
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 9, "name": "Legacy admin", "description": "", "deprecated": true}},
		})
	})
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 102, "subjectId": "u1", "subjectType": "user", "roleId": 9, "roleName": "Legacy admin",
			"scopeId": "proj-1", "scopeType": "project", "scopeName": "ml-team", "clusterId": muaClusterUUID,
		})
	})

	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Confirm: true,
		Entity:  "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user",
		RoleName: "Legacy admin", ScopeType: "project", ScopeID: "proj-1", ClusterID: muaClusterUUID,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	require.NotEmpty(t, out.Warnings)
	assert.Contains(t, out.Warnings[0], "deprecated")
}

func TestManageUserAccess_DeleteAccessRuleRequiresConfirm(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("delete"))
	mux.HandleFunc(accessRulesPath+"/42", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 42, "subjectId": "u1", "subjectType": "user", "roleId": 7, "roleName": "L1 researcher",
			"scopeId": "proj-1", "scopeType": "project", "scopeName": "ml-team", "clusterId": muaClusterUUID,
		})
	})

	_, err := manageUserAccessRun(context.Background(), client,
		manageUserAccessInput{Entity: "access-rule", Action: "delete", ID: "42"})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
}

func TestManageUserAccess_DeleteAccessRuleConfirmed(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("delete"))
	deleted := false
	mux.HandleFunc(accessRulesPath+"/42", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": 42, "subjectId": "u1", "subjectType": "user", "roleId": 7, "roleName": "L1 researcher",
				"scopeId": "proj-1", "scopeType": "project", "scopeName": "ml-team", "clusterId": muaClusterUUID,
			})
		case http.MethodDelete:
			deleted = true
			w.WriteHeader(http.StatusOK)
		}
	})

	out, err := manageUserAccessRun(context.Background(), client,
		manageUserAccessInput{Entity: "access-rule", Action: "delete", ID: "42", Confirm: true})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.True(t, deleted)
	assert.Equal(t, "L1 researcher", out.Before["roleName"])
}

func TestManageUserAccess_CreateAccessRuleSubjectByEmail(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, []map[string]any{{"id": "u-9", "username": "subj@b.com"}})
	})
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 7, "name": "L1 researcher", "description": ""}},
		})
	})
	var posted map[string]any
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &posted))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 103, "subjectId": "u-9", "subjectType": "user", "roleId": 7, "roleName": "L1 researcher",
			"scopeId": "proj-1", "scopeType": "project", "scopeName": "ml-team", "clusterId": muaClusterUUID,
		})
	})

	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Confirm: true,
		Entity:  "access-rule", Action: "create", SubjectEmail: "subj@b.com", SubjectType: "user",
		RoleName: "L1 researcher", ScopeType: "project", ScopeID: "proj-1", ClusterID: muaClusterUUID,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	// The subject sent upstream must be the username (email), not the internal UUID "u-9".
	assert.Equal(t, "subj@b.com", posted["subjectId"])
}

// A user access-rule must be keyed by the user's username (email), not the internal UUID.
// Passing the UUID (what identity-create returns) must resolve to the username, otherwise the
// grant is orphaned — it binds to no real principal and never shows under the user.
func TestManageUserAccess_CreateAccessRuleUserSubjectResolvesToUsername(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath+"/2f73d5a6", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"id": "2f73d5a6", "username": "barak@run.ai"})
	})
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 24, "name": "Department administrator", "description": ""}},
		})
	})
	var posted map[string]any
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &posted))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 112, "subjectId": "barak@run.ai", "subjectType": "user", "roleId": 24,
			"roleName": "Department administrator", "scopeId": "1001", "scopeType": "tenant", "scopeName": "runai",
		})
	})

	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Confirm: true,
		Entity:  "access-rule", Action: "create", SubjectID: "2f73d5a6", SubjectType: "user",
		RoleName: "Department administrator", ScopeType: "tenant", ScopeID: "1001",
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "barak@run.ai", posted["subjectId"], "user subject must be the username, not the UUID")
}

func TestManageUserAccess_RoleNameCaseInsensitive(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath, usersSearchHandler("u1", "u1"))
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 4, "name": "System administrator", "description": ""}},
		})
	})
	var posted map[string]any
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &posted))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 401, "subjectId": "u1", "subjectType": "user", "roleId": 4, "roleName": "System administrator",
			"scopeId": "", "scopeType": "system",
		})
	})

	// Caller mis-cases the role ("System Administrator") — should still resolve to id 4.
	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Confirm: true,
		Entity:  "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user",
		RoleName: "System Administrator", ScopeType: "system",
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.EqualValues(t, 4, posted["roleId"])
}

func TestManageUserAccess_UnknownRolePointsToListRoles(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 2, "name": "L1 researcher", "description": ""}},
		})
	})

	_, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Entity: "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user",
		RoleName: "reseacher", ScopeType: "system", DryRun: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
	assert.Contains(t, err.Error(), "list_roles")
}

func TestManageUserAccess_DeleteAccessRuleByTuple(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("delete"))
	mux.HandleFunc(usersPath, usersSearchHandler("u-9", "u-9"))
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 7, "name": "L1 researcher", "description": ""}},
		})
	})
	deleted := false
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"totalRecords": 1, "displayRecords": 1,
			"accessRules": []map[string]any{{
				"id": 55, "subjectId": "u-9", "subjectType": "user", "roleId": 7, "roleName": "L1 researcher",
				"scopeId": "proj-1", "scopeType": "project", "scopeName": "ml-team", "clusterId": muaClusterUUID,
			}},
		})
	})
	mux.HandleFunc(accessRulesPath+"/55", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		deleted = true
		w.WriteHeader(http.StatusOK)
	})

	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Entity: "access-rule", Action: "delete", SubjectID: "u-9", SubjectType: "user",
		RoleName: "L1 researcher", ScopeType: "project", ScopeID: "proj-1", ClusterID: muaClusterUUID, Confirm: true,
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.True(t, deleted)
	assert.Equal(t, "55", out.ID)
}

// Guard against deleting the wrong rule: if the server returns a rule that matches
// role+scope+type+cluster but belongs to a DIFFERENT subject, the tuple delete must not
// match it — otherwise a delete-for-user could remove someone else's grant.
func TestManageUserAccess_DeleteAccessRuleByTupleWrongSubjectNotMatched(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("delete"))
	mux.HandleFunc(usersPath, usersSearchHandler("u-alice", "alice@run.ai"))
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 7, "name": "L1 researcher", "description": ""}},
		})
	})
	deleteCalled := false
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"accessRules": []map[string]any{{
				"id": 77, "subjectId": "bob@run.ai", "subjectType": "user", "roleId": 7, "roleName": "L1 researcher",
				"scopeId": "proj-1", "scopeType": "project", "scopeName": "ml-team", "clusterId": muaClusterUUID,
			}},
		})
	})
	mux.HandleFunc(accessRulesPath+"/77", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCalled = true
		}
		w.WriteHeader(http.StatusOK)
	})

	_, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Entity: "access-rule", Action: "delete", SubjectEmail: "alice@run.ai", SubjectType: "user",
		RoleName: "L1 researcher", ScopeType: "project", ScopeID: "proj-1", ClusterID: muaClusterUUID, Confirm: true,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
	assert.False(t, deleteCalled, "must not delete a rule belonging to a different subject")
}

func TestScopeAllowsTarget(t *testing.T) {
	sys := authz.PermittedScopes{System: bptr(true)}
	tenants := authz.PermittedScopes{Tenants: []string{"t1"}}
	clusters := authz.PermittedScopes{Clusters: []string{"c1"}}
	cases := []struct {
		name      string
		scope     authz.PermittedScopes
		scopeType string
		scopeID   string
		want      bool
	}{
		{"system with system perm", sys, "system", "", true},
		{"system with tenant-only denied", tenants, "system", "", false}, // CR#1: tenant grant must not satisfy system
		{"tenant matches id", tenants, "tenant", "t1", true},
		{"tenant wrong id denied", tenants, "tenant", "t2", false},
		{"tenant via system", sys, "tenant", "t1", true},
		{"cluster matches", clusters, "cluster", "c1", true},
		{"project defers to upstream", authz.PermittedScopes{}, "project", "proj-1", true}, // #8
		{"department defers to upstream", authz.PermittedScopes{}, "department", "dept-1", true},
		{"unknown denied", sys, "galaxy", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, scopeAllowsTarget(tc.scope, tc.scopeType, tc.scopeID, ""))
		})
	}
}

func TestManageUserAccess_CreateAccessRuleByRoleIdResolvesAndWarns(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath, usersSearchHandler("u1", "u1"))
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 9, "name": "Legacy admin", "description": "", "deprecated": true}},
		})
	})
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 201, "subjectId": "u1", "subjectType": "user", "roleId": 9, "roleName": "Legacy admin",
			"scopeId": "", "scopeType": "system",
		})
	})

	roleID := int32(9)
	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Confirm: true,
		Entity:  "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user",
		RoleID: &roleID, ScopeType: "system",
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	require.NotEmpty(t, out.Warnings)
	assert.Contains(t, out.Warnings[0], "deprecated")
}

func TestManageUserAccess_DeleteTupleScopeMismatchNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("delete"))
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 7, "name": "L1 researcher", "description": ""}},
		})
	})
	// Server returns a rule for the subject but at a DIFFERENT scope than requested.
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"accessRules": []map[string]any{{
				"id": 60, "subjectId": "u-9", "subjectType": "user", "roleId": 7, "roleName": "L1 researcher",
				"scopeId": "proj-OTHER", "scopeType": "project", "clusterId": muaClusterUUID,
			}},
		})
	})

	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Entity: "access-rule", Action: "delete", SubjectID: "u-9", SubjectType: "user",
		RoleName: "L1 researcher", ScopeType: "project", ScopeID: "proj-1", ClusterID: muaClusterUUID,
		IgnoreMissing: true,
	})
	require.NoError(t, err)
	assert.False(t, out.Applied) // scopeId mismatch → not matched, no wrong-rule delete
	require.NotEmpty(t, out.Warnings)
	assert.Contains(t, out.Warnings[0], "nothing to delete")
}

func TestManageUserAccess_CreateAccessRuleByClusterName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(usersPath, usersSearchHandler("u1", "u1"))
	mux.HandleFunc("/api/v1/clusters", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, []map[string]any{
			{"uuid": "c-uuid", "name": "my-cluster", "tenantId": 1},
		})
	})
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 7, "name": "L1 researcher", "description": ""}},
		})
	})
	var posted map[string]any
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &posted))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 301, "subjectId": "u1", "subjectType": "user", "roleId": 7, "roleName": "L1 researcher",
			"scopeId": "c-uuid", "scopeType": "cluster", "clusterId": "c-uuid",
		})
	})

	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Confirm: true,
		Entity:  "access-rule", Action: "create", SubjectID: "u1", SubjectType: "user",
		RoleName: "L1 researcher", ScopeType: "cluster", ClusterName: "my-cluster",
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "c-uuid", posted["scopeId"]) // cluster scope's id = resolved cluster
	assert.Equal(t, "c-uuid", posted["clusterId"])
}

func TestListRoles(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{
				{"id": 2, "name": "L1 researcher", "description": "view + submit"},
				{"id": 9, "name": "Legacy admin", "description": "", "deprecated": true},
			},
		})
	})

	out, err := listRolesRun(context.Background(), client, listRolesInput{})
	require.NoError(t, err)
	require.Len(t, out.Roles, 2)
	assert.Equal(t, int32(2), out.Roles[0].ID)
	assert.Equal(t, "L1 researcher", out.Roles[0].Name)
	assert.False(t, out.Roles[0].Deprecated)
	assert.True(t, out.Roles[1].Deprecated)
}

func TestListAccessRules(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "u-9", r.URL.Query().Get("subjectIds"))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"totalRecords": 1, "displayRecords": 1,
			"accessRules": []map[string]any{{
				"id": 5, "subjectId": "u-9", "subjectType": "user", "roleId": 2, "roleName": "L1 researcher",
				"scopeType": "project", "scopeId": "p1", "scopeName": "proj", "clusterId": muaClusterUUID,
			}},
		})
	})

	out, err := listAccessRulesRun(context.Background(), client, listAccessRulesInput{SubjectID: "u-9"})
	require.NoError(t, err)
	require.Len(t, out.AccessRules, 1)
	assert.Equal(t, int32(5), out.AccessRules[0].ID)
	assert.Equal(t, "L1 researcher", out.AccessRules[0].RoleName)
	assert.Equal(t, muaClusterUUID, out.AccessRules[0].ClusterID)
	assert.False(t, out.HasMore)
}

func TestManageUserAccess_CreateAccessRuleSAsubjectUsesClientId(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permTenantHandler("create"))
	mux.HandleFunc(serviceAcctPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, []map[string]any{
			{"id": "sa-uuid-xyz", "name": "ci-bot", "clientId": "ci-bot", "enabled": true},
		})
	})
	mux.HandleFunc(rolesPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"roles": []map[string]any{{"id": 7, "name": "L1 researcher", "description": ""}},
		})
	})
	var posted map[string]any
	mux.HandleFunc(accessRulesPath, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &posted))
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": 501, "subjectId": "ci-bot", "subjectType": "service-account", "roleId": 7, "roleName": "L1 researcher",
			"scopeId": "", "scopeType": "system",
		})
	})

	out, err := manageUserAccessRun(context.Background(), client, manageUserAccessInput{
		Confirm: true,
		Entity:  "access-rule", Action: "create", SubjectName: "ci-bot", SubjectType: "service-account",
		RoleName: "L1 researcher", ScopeType: "system",
	})
	require.NoError(t, err)
	assert.True(t, out.Applied)
	// subjectId must be the clientId (== name), NOT the internal UUID.
	assert.Equal(t, "ci-bot", posted["subjectId"])
}

func TestManageUserAccessToolsLint(t *testing.T) {
	errs := runaimcp.Lint(func(r *runaimcp.Registry) {
		registerManageUserAccess(r, nil)
	})
	assert.Empty(t, errs)
}
