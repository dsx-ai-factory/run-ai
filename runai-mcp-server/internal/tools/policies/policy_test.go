/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policies

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	policy "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPolicyID = "6fced230-c208-4b73-98a8-964fc1a32787"

func testGVKPolicy() map[string]any {
	return map[string]any{
		"id":           testPolicyID,
		"workloadType": map[string]any{"group": "apps", "version": "v1", "kind": "Deployment"},
		"scope":        "project",
		"projectId":    "123",
		"rules":        []any{testGVKPolicyRule()},
	}
}

func testGVKPolicyRule() map[string]any {
	return map[string]any{
		"field":        "image",
		"specSelector": "container",
		"enforcements": []any{map[string]any{"required": true}},
	}
}

func testLegacyPolicy() map[string]any {
	return map[string]any{
		"meta": map[string]any{
			"id":        testPolicyID,
			"name":      "test",
			"scope":     "tenant",
			"createdBy": "test",
			"updatedBy": "test",
			"createdAt": "2026-10-18T10:00:00Z",
			"updatedAt": "2026-10-18T10:00:00Z",
		},
		"policy": map[string]any{"defaults": map[string]any{}, "rules": map[string]any{}},
	}
}

func testLegacyPolicyInput(kind, action string) manageWorkloadPolicyInput {
	in := manageWorkloadPolicyInput{PolicyVersion: "legacy", WorkloadType: kind, Action: action, Confirm: true}
	if action == actionDelete {
		in.Scope = "tenant"
	} else {
		in.Request = map[string]any{"meta": map[string]any{"name": "test", "scope": "tenant"}, "policy": map[string]any{}}
	}
	return in
}

func TestManageLegacyPolicy(t *testing.T) {
	for kind, path := range map[string]string{
		"training":             "trainings",
		"workspace":            "workspaces",
		"distributed":          "distributed",
		"inference":            "inferences",
		"distributedInference": "distributed-inferences",
		"nimService":           "nim-services",
		"fallback":             "fallback",
	} {
		for action, method := range map[string]string{
			actionCreate: http.MethodPut,
			actionUpdate: http.MethodPatch,
			actionDelete: http.MethodDelete,
		} {
			t.Run(kind+"/"+action, func(t *testing.T) {
				// 1. Set up the API and input.
				c, mux := runaiclienttest.NewServer(t)
				writes := 0
				mux.HandleFunc("/api/v2/policy/"+path, func(w http.ResponseWriter, r *http.Request) {
					writes++
					assert.Equal(t, method, r.Method)
					if action == actionDelete {
						w.WriteHeader(http.StatusNoContent)
						return
					}
					var body map[string]any
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					assert.Equal(t, "test", body["meta"].(map[string]any)["name"])
					if action == actionUpdate {
						assert.Equal(t, []any{"compute.gpu"}, body["reset"])
						assert.NotContains(t, body, "policy")
					}
					runaiclienttest.WriteJSON(w, http.StatusOK, testLegacyPolicy())
				})
				in := testLegacyPolicyInput(kind, action)
				in.Confirm = action != actionCreate
				if action == actionUpdate {
					delete(in.Request, "policy")
					in.Request["reset"] = []any{"compute.gpu"}
				}

				// 2. Validate and call the tool.
				require.NoError(t, in.Validate())
				out, err := manageWorkloadPolicyRun(context.Background(), c, in)

				// 3. Check the result.
				require.NoError(t, err)
				if action == actionDelete {
					assert.Equal(t, map[string]any{}, out)
				} else {
					assert.Equal(t, testLegacyPolicy(), out)
				}
				assert.Equal(t, 1, writes)
			})
		}
	}
}

func TestManageGVKPolicy(t *testing.T) {
	for _, tc := range []struct {
		action, method, path string
	}{
		{actionCreate, http.MethodPost, "/api/v1/policies"},
		{actionUpdate, http.MethodPatch, "/api/v1/policies/" + testPolicyID},
		{actionDelete, http.MethodDelete, "/api/v1/policies/" + testPolicyID},
	} {
		t.Run(tc.action, func(t *testing.T) {
			// 1. Set up the API and input.
			c, mux := runaiclienttest.NewServer(t)
			writes := 0
			mux.HandleFunc(tc.path, func(w http.ResponseWriter, r *http.Request) {
				writes++
				assert.Equal(t, tc.method, r.Method)
				if tc.action == actionDelete {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, testGVKPolicy())
			})
			in := manageWorkloadPolicyInput{
				Action: tc.action, PolicyVersion: versionGVK, PolicyID: testPolicyID, Confirm: true,
			}
			switch tc.action {
			case actionCreate:
				in.Confirm = false
				in.PolicyID = ""
				in.Request = testGVKPolicy()
				delete(in.Request, "id")
			case actionUpdate:
				in.Request = map[string]any{"description": "updated"}
			}

			// 2. Call the tool.
			out, err := manageWorkloadPolicyRun(context.Background(), c, in)

			// 3. Check the result.
			require.NoError(t, err)
			if tc.action == actionDelete {
				assert.Equal(t, map[string]any{}, out)
			} else {
				assert.Equal(t, testGVKPolicy(), out)
			}
			assert.Equal(t, 1, writes)
		})
	}
}

func TestPolicyUpdateAndDeleteNeedConfirmation(t *testing.T) {
	for _, version := range []string{versionLegacy, versionGVK} {
		for _, action := range []string{actionUpdate, actionDelete} {
			t.Run(version+"/"+action, func(t *testing.T) {
				// 1. Set up the API and input.
				c, mux := runaiclienttest.NewServer(t)
				calls := 0
				mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
					calls++
					w.WriteHeader(http.StatusInternalServerError)
				})
				in := testLegacyPolicyInput("training", action)
				in.Confirm = false
				if version == versionGVK {
					in = manageWorkloadPolicyInput{PolicyVersion: versionGVK, Action: action, PolicyID: testPolicyID}
					if action == actionUpdate {
						in.Request = map[string]any{"description": "new"}
					}
				}

				// 2. Validate and call the tool.
				require.NoError(t, in.Validate())
				out, err := manageWorkloadPolicyRun(context.Background(), c, in)

				// 3. Check the result.
				assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
				assert.Nil(t, out)
				assert.Zero(t, calls, "confirmation must not make API requests")
			})
		}
	}
}

func TestListPolicies(t *testing.T) {
	for _, tc := range []struct {
		name, version string
		hasPolicies   bool
	}{
		{"legacy policies found", versionLegacy, true},
		{"legacy no policies found", versionLegacy, false},
		{"GVK policies found", versionGVK, true},
		{"GVK no policies found", versionGVK, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Set up the API and input.
			c, mux := runaiclienttest.NewServer(t)
			path, key := "/api/v2/policy", "policies"
			policy := map[string]any{"type": "training", "meta": testLegacyPolicy()["meta"]}
			if tc.version == versionGVK {
				path, key, policy = "/api/v1/policies", "items", testGVKPolicy()
			}
			entries := []map[string]any{}
			if tc.hasPolicies {
				entries = append(entries, policy)
			}
			expected := listWorkloadPoliciesOutput{Policies: []map[string]any{}, Items: []map[string]any{}}
			if tc.version == versionLegacy {
				expected.Policies = entries
			} else {
				expected.Items = entries
			}
			mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				if tc.version == versionGVK {
					assert.Equal(t, "50", r.URL.Query().Get("limit"))
				} else {
					assert.Equal(t, "false", r.URL.Query().Get("includeGvkPolicies"))
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{key: entries})
			})

			// 2. Call the tool.
			out, err := listWorkloadPoliciesRun(context.Background(), c, listWorkloadPoliciesInput{PolicyVersion: tc.version})

			// 3. Check the result.
			require.NoError(t, err)
			assert.Equal(t, expected, out)
		})
	}
}

func TestListPoliciesReturnsEmptyArraysForNullAPIResults(t *testing.T) {
	for _, tc := range []struct {
		version, path, key string
	}{
		{versionLegacy, "/api/v2/policy", "policies"},
		{versionGVK, "/api/v1/policies", "items"},
	} {
		t.Run(tc.version, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(tc.path, func(w http.ResponseWriter, _ *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{tc.key: nil})
			})

			out, err := listWorkloadPoliciesRun(context.Background(), client, listWorkloadPoliciesInput{PolicyVersion: tc.version})

			require.NoError(t, err)
			assert.NotNil(t, out.Items)
			assert.NotNil(t, out.Policies)
			assert.Empty(t, out.Items)
			assert.Empty(t, out.Policies)
		})
	}
}

func TestListLegacyPolicyWorkloadFilter(t *testing.T) {
	for _, tc := range []struct{ input, query string }{
		{"training", "Training"}, {"workspace", "Workspace"},
		{"distributed", "Distributed"}, {"inference", "Inference"},
		{"distributedInference", "DistributedInference"}, {"nimService", "NimService"},
		{"fallback", "Fallback"}, {"Training", "Training"}, {"unknown", "unknown"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v2/policy", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.query, r.URL.Query().Get("workloadType"))
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{})
			})
			_, err := listWorkloadPoliciesRun(context.Background(), client, listWorkloadPoliciesInput{
				PolicyVersion: versionLegacy, WorkloadType: tc.input,
			})
			require.NoError(t, err)
		})
	}
}

func TestPolicyInputValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*manageWorkloadPolicyInput)
	}{
		{"invalid action", func(in *manageWorkloadPolicyInput) { in.Action = "edit" }},
		{"invalid workload type", func(in *manageWorkloadPolicyInput) { in.WorkloadType = "unknown" }},
		{"invalid policy version", func(in *manageWorkloadPolicyInput) { in.PolicyVersion = "new" }},
		{"unknown request field", func(in *manageWorkloadPolicyInput) { in.Request["typo"] = true }},
		{"unknown default field", func(in *manageWorkloadPolicyInput) {
			in.Request["policy"] = map[string]any{"defaults": map[string]any{"typo": true}}
		}},
		{"missing request", func(in *manageWorkloadPolicyInput) { in.Request = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := testLegacyPolicyInput("training", "update")
			tc.mutate(&in)
			require.Error(t, in.Validate())
		})
	}
}

func TestPolicyRequestPresenceFollowsAPI(t *testing.T) {
	emptyUpdate := manageWorkloadPolicyInput{
		Action: actionUpdate, PolicyVersion: versionGVK, PolicyID: testPolicyID,
		Request: map[string]any{},
	}
	require.NoError(t, emptyUpdate.Validate(), "the API decides whether an explicitly provided empty update is meaningful")

	deleteWithBody := manageWorkloadPolicyInput{
		Action: actionDelete, PolicyVersion: versionGVK, PolicyID: testPolicyID,
		Request: map[string]any{},
	}
	require.ErrorContains(t, deleteWithBody.Validate(), "must not be set")
}

func TestLegacyPolicyRejectsTopLevelScope(t *testing.T) {
	for _, action := range []string{actionCreate, actionUpdate} {
		for _, field := range []string{"scope", fieldClusterID, fieldDepartmentID, fieldProjectID} {
			t.Run(action+"/"+field, func(t *testing.T) {
				in := testLegacyPolicyInput("training", action)
				require.NoError(t, in.Validate())
				switch field {
				case "scope":
					in.Scope = "project"
				case fieldClusterID:
					in.ClusterID = testPolicyID
				case fieldDepartmentID:
					in.DepartmentID = "department-1"
				case fieldProjectID:
					in.ProjectID = "123"
				}
				require.ErrorContains(t, in.Validate(), "request.meta")
			})
		}
	}
}

func TestListPoliciesAPIError(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  listWorkloadPoliciesInput
		status int
	}{
		{name: "service error", input: listWorkloadPoliciesInput{PolicyVersion: versionGVK}, status: http.StatusServiceUnavailable},
		{name: "invalid query", input: listWorkloadPoliciesInput{
			PolicyVersion: versionGVK, FilterBy: []string{"bad filter"}, SortBy: "bad sort", SortOrder: "bad order", Limit: 501, Offset: -1,
		}, status: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Set up the API and input.
			c, mux := runaiclienttest.NewServer(t)
			calls := 0
			mux.HandleFunc("/api/v1/policies", func(w http.ResponseWriter, r *http.Request) {
				calls++
				if tc.status == http.StatusBadRequest {
					assert.Equal(t, "bad filter", r.URL.Query().Get("filterBy"))
					assert.Equal(t, "bad sort", r.URL.Query().Get("sortBy"))
					assert.Equal(t, "bad order", r.URL.Query().Get("sortOrder"))
					assert.Equal(t, "501", r.URL.Query().Get("limit"))
					assert.Equal(t, "-1", r.URL.Query().Get("offset"))
				}
				runaiclienttest.WriteJSON(w, tc.status, map[string]any{"message": http.StatusText(tc.status)})
			})
			require.NoError(t, tc.input.Validate(), "query values are validated by the API")

			// 2. Call the tool.
			_, err := listWorkloadPoliciesRun(context.Background(), c, tc.input)

			// 3. Check the result.
			if tc.status == http.StatusBadRequest {
				assert.ErrorIs(t, err, runaimcp.ErrValidation)
			} else {
				assert.ErrorIs(t, err, runaimcp.ErrUpstream)
			}
			assert.Equal(t, 1, calls)
		})
	}
}

func TestGetGVKPolicy(t *testing.T) {
	// 1. Set up the API and input.
	c, mux := runaiclienttest.NewServer(t)
	body := testGVKPolicy()
	body["description"] = "policy description"
	body["futureField"] = true
	mux.HandleFunc("/api/v1/policies/"+testPolicyID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, body)
	})
	in := listWorkloadPoliciesInput{PolicyVersion: versionGVK, PolicyID: testPolicyID}
	require.NoError(t, in.Validate())

	// 2. Call the tool.
	out, err := listWorkloadPoliciesRun(context.Background(), c, in)

	// 3. Check the result.
	require.NoError(t, err)
	assert.Equal(t, []map[string]any{body}, out.Items, "return the API policy without filtering its fields")
	assert.Empty(t, out.Policies)
}

func TestLegacyProjectIDMustBeNumber(t *testing.T) {
	in := testLegacyPolicyInput("inference", "update")
	in.Request["meta"] = map[string]any{"name": "test", "scope": "project", "projectId": float64(123)}
	require.NoError(t, in.Validate())
	in.Request["meta"].(map[string]any)["projectId"] = "123"
	require.Error(t, in.Validate())
}

func TestReplaceGVKPolicyRules(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rules  []any
		status int
	}{
		{name: "rules accepted", rules: []any{testGVKPolicyRule()}, status: http.StatusOK},
		{name: "empty rules rejected by API", rules: []any{}, status: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Set up the API and input.
			c, mux := runaiclienttest.NewServer(t)
			writes := 0
			mux.HandleFunc("/api/v1/policies/"+testPolicyID, func(w http.ResponseWriter, r *http.Request) {
				writes++
				assert.Equal(t, http.MethodPut, r.Method)
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, tc.rules, body["rules"])
				if tc.status == http.StatusBadRequest {
					runaiclienttest.WriteJSON(w, tc.status, map[string]any{"message": "policy must contain rules"})
					return
				}
				runaiclienttest.WriteJSON(w, tc.status, testGVKPolicy())
			})

			// 2. Call the tool.
			out, err := manageWorkloadPolicyRun(context.Background(), c, manageWorkloadPolicyInput{
				Action: actionUpdate, PolicyVersion: versionGVK, PolicyID: testPolicyID,
				ReplaceRules: true, Request: map[string]any{"rules": tc.rules}, Confirm: true,
			})

			// 3. Check the result.
			if tc.status == http.StatusOK {
				require.NoError(t, err)
				assert.Equal(t, testGVKPolicy(), out)
			} else {
				assert.ErrorIs(t, err, runaimcp.ErrValidation)
				assert.Nil(t, out)
			}
			assert.Equal(t, 1, writes, "the API must receive and validate the update")
		})
	}
}

func TestPolicyScopeValidation(t *testing.T) {
	for _, tc := range []struct {
		name, scope, cluster, department, project string
		valid                                     bool
	}{
		{name: "tenant needs no ID", scope: "tenant", valid: true},
		{name: "cluster with UUID", scope: "cluster", cluster: testPolicyID, valid: true},
		{name: "department with ID", scope: "department", department: testPolicyID, valid: true},
		{name: "project with numeric ID", scope: "project", project: "123", valid: true},
		{name: "system scope is unsupported", scope: "system"},
		{name: "project ID is missing", scope: "project"},
		{name: "cluster ID is not a UUID", scope: "cluster", cluster: "bad"},
		{name: "department accepts an extra project ID", scope: "department", department: testPolicyID, project: "123", valid: true},
		{name: "tenant accepts an extra cluster ID", scope: "tenant", cluster: testPolicyID, valid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := common.ValidatePolicyScope(tc.scope, tc.cluster, tc.department, tc.project)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGVKPolicyEnforcements(t *testing.T) {
	for name, enforcements := range map[string][]any{
		"empty":         {},
		"required":      {map[string]any{"required": true}},
		"blocked":       {map[string]any{"blocked": false}},
		"default":       {map[string]any{"default": map[string]any{"value": "image:v1"}}},
		"set":           {map[string]any{"set": map[string]any{"nested": map[string]any{"value": "fixed"}}}},
		"match":         {map[string]any{"match": "^image-"}},
		"options":       {map[string]any{"options": []any{"one", "two"}}},
		"range":         {map[string]any{"range": map[string]any{"min": float64(1), "max": float64(8)}}},
		"quantityRange": {map[string]any{"quantityRange": map[string]any{"min": "1Gi", "max": "8Gi"}}},
		"elements":      {map[string]any{"elements": map[string]any{"min": float64(1)}}},
		"items":         {map[string]any{"items": map[string]any{"max": float64(2)}}},
		"length":        {map[string]any{"length": map[string]any{"min": float64(1)}}},
	} {
		t.Run(name, func(t *testing.T) {
			raw := map[string]any{"rules": []any{map[string]any{"field": "image", "specSelector": "container", "enforcements": enforcements}}}
			var request policy.UpdateGvkPolicyReq
			require.NoError(t, decodeGVKPolicyRequest(raw, &request))
			// Encoding again must preserve every enforcement setting.
			encoded, err := json.Marshal(request)
			require.NoError(t, err)
			var actual map[string]any
			require.NoError(t, json.Unmarshal(encoded, &actual))
			assert.Equal(t, raw, actual)
		})
	}
	for name, enforcement := range map[string]any{
		"rejects unknown setting":     map[string]any{"typo": true},
		"rejects unknown range field": map[string]any{"range": map[string]any{"typo": true}},
		"rejects multiple settings":   map[string]any{"required": true, "blocked": true},
	} {
		t.Run(name, func(t *testing.T) {
			var request policy.UpdateGvkPolicyReq
			require.Error(t, decodeGVKPolicyRequest(map[string]any{"rules": []any{map[string]any{"field": "image", "specSelector": "container", "enforcements": []any{enforcement}}}}, &request))
		})
	}
}

func TestGVKPolicyNullRulesFollowAPIBehavior(t *testing.T) {
	var request policy.UpdateGvkPolicyReq
	require.NoError(t, decodeGVKPolicyRequest(map[string]any{
		"description": "updated",
		"rules":       nil,
	}, &request))
	assert.Nil(t, request.Rules)
}

type decliningPolicySession struct {
	calls   int
	message string
}

func (f *decliningPolicySession) Elicit(_ context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error) {
	f.calls++
	f.message = params.Message
	return &mcp.ElicitResult{Action: "decline"}, nil
}
func (*decliningPolicySession) InitializeParams() *mcp.InitializeParams {
	return &mcp.InitializeParams{Capabilities: &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{}}}
}
func (*decliningPolicySession) ID() string { return "policy-test" }

func TestPolicyDeclineOverridesConfirm(t *testing.T) {
	// 1. Set up the API and input.
	c, mux := runaiclienttest.NewServer(t)
	calls := 0
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	})
	elicitor := &decliningPolicySession{}

	// 2. Call the tool.
	out, err := manageWorkloadPolicyRun(runaimcp.WithSession(context.Background(), elicitor), c, manageWorkloadPolicyInput{Action: "delete", PolicyVersion: "gvk", PolicyID: testPolicyID, Confirm: true})

	// 3. Check the result.
	assert.ErrorIs(t, err, runaimcp.ErrDeclined)
	assert.Nil(t, out)
	assert.Equal(t, 1, elicitor.calls)
	assert.Zero(t, calls)
}

func TestLegacyPolicyConfirmationShowsScope(t *testing.T) {
	in := testLegacyPolicyInput("training", actionUpdate)
	in.Request["meta"] = map[string]any{
		"name": strings.Repeat("policy", 20), "scope": "project",
		"clusterId": testPolicyID, "departmentId": "department-1", "projectId": float64(123),
	}
	session := &decliningPolicySession{}
	err := confirmPolicyOperation(runaimcp.WithSession(context.Background(), session), in)
	require.ErrorIs(t, err, runaimcp.ErrDeclined)
	assert.Contains(t, session.message, "scope: project")
	assert.Contains(t, session.message, "projectId: 123")
	assert.Contains(t, session.message, "departmentId: department-1")
	assert.Contains(t, session.message, "clusterId: "+testPolicyID)
}

func TestGVKPolicyConfirmationIdentifiesTarget(t *testing.T) {
	session := &decliningPolicySession{}
	err := confirmPolicyOperation(runaimcp.WithSession(context.Background(), session), manageWorkloadPolicyInput{
		Action: actionDelete, PolicyVersion: versionGVK, PolicyID: testPolicyID,
	})
	require.ErrorIs(t, err, runaimcp.ErrDeclined)
	assert.Contains(t, session.message, `workload policy "GVK" (`+testPolicyID+`)`)
	assert.NotContains(t, session.message, "request:")
	assert.NotContains(t, session.message, "replaceRules:")
	assert.NotContains(t, session.message, "approveClusterDeletion:")
}

func TestLegacyPolicyScopeIsPreserved(t *testing.T) {
	for _, action := range []string{actionCreate, actionUpdate} {
		t.Run(action, func(t *testing.T) {
			// 1. Set up the API and input.
			c, mux := runaiclienttest.NewServer(t)
			meta := map[string]any{"name": "project policy", "scope": "project", "projectId": float64(4500002), "departmentId": "dept", "clusterId": testPolicyID}
			body := map[string]any{"meta": meta, "policy": map[string]any{}}
			writes := 0
			mux.HandleFunc("/api/v2/policy/trainings", func(w http.ResponseWriter, r *http.Request) {
				writes++
				assert.Equal(t, map[string]string{actionCreate: http.MethodPut, actionUpdate: http.MethodPatch}[action], r.Method)
				var actual map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&actual))
				assert.Equal(t, meta, actual["meta"])
				runaiclienttest.WriteJSON(w, 200, body)
			})

			// 2. Call the tool.
			out, err := manageWorkloadPolicyRun(context.Background(), c, manageWorkloadPolicyInput{
				Action: action, PolicyVersion: versionLegacy, WorkloadType: "training", Request: body, Confirm: true,
			})

			// 3. Check the result.
			require.NoError(t, err)
			assert.Equal(t, body, out)
			assert.Equal(t, 1, writes)
		})
	}
}

func TestListGVKPoliciesPaginationAndSorting(t *testing.T) {
	// 1. Set up the API and input.
	c, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc("/api/v1/policies", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2", r.URL.Query().Get("limit"))
		assert.Equal(t, "4", r.URL.Query().Get("offset"))
		assert.Equal(t, "updatedAt", r.URL.Query().Get("sortBy"))
		assert.Equal(t, "desc", r.URL.Query().Get("sortOrder"))
		runaiclienttest.WriteJSON(w, 200, map[string]any{"items": []any{testGVKPolicy(), testGVKPolicy()}})
	})
	in := listWorkloadPoliciesInput{PolicyVersion: versionGVK, Limit: 2, Offset: 4, SortBy: "updatedAt", SortOrder: "desc"}
	require.NoError(t, in.Validate())

	// 2. Call the tool.
	out, err := listWorkloadPoliciesRun(context.Background(), c, in)

	// 3. Check the result.
	require.NoError(t, err)
	assert.Len(t, out.Items, 2)
}

func TestManagePolicyAPIErrors(t *testing.T) {
	for _, version := range []string{versionLegacy, versionGVK} {
		for _, status := range []int{http.StatusForbidden, http.StatusNotFound} {
			t.Run(version+"/"+http.StatusText(status), func(t *testing.T) {
				// 1. Set up the API and input.
				c, mux := runaiclienttest.NewServer(t)
				path := "/api/v2/policy/trainings"
				in := testLegacyPolicyInput("training", actionUpdate)
				if version == versionGVK {
					path = "/api/v1/policies/" + testPolicyID
					in = manageWorkloadPolicyInput{
						Action: actionUpdate, PolicyVersion: versionGVK, PolicyID: testPolicyID,
						Request: map[string]any{"description": "updated"}, Confirm: true,
					}
				}
				calls := 0
				mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
					calls++
					assert.Equal(t, http.MethodPatch, r.Method)
					runaiclienttest.WriteJSON(w, status, map[string]any{"message": http.StatusText(status)})
				})

				// 2. Validate and call the tool.
				require.NoError(t, in.Validate())
				out, err := manageWorkloadPolicyRun(context.Background(), c, in)
				want := runaimcp.ErrForbidden
				if status == http.StatusNotFound {
					want = runaimcp.ErrNotFound
				}

				// 3. Check the result.
				assert.ErrorIs(t, err, want)
				assert.Nil(t, out)
				assert.Equal(t, 1, calls)
			})
		}
	}
}

func TestLegacyPolicyAssetsArePreserved(t *testing.T) {
	for _, tc := range []struct {
		name, kind, path string
		assets           any
	}{
		{name: "no assets provided", kind: "training", path: "trainings"},
		{name: "training assets", kind: "training", path: "trainings", assets: []any{testPolicyID}},
		{name: "distributed master assets", kind: "distributed", path: "distributed", assets: map[string]any{"master": []any{testPolicyID}}},
		{name: "distributed inference worker assets", kind: "distributedInference", path: "distributed-inferences", assets: map[string]any{"worker": []any{testPolicyID}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Set up the API and input.
			c, mux := runaiclienttest.NewServer(t)
			writes := 0
			mux.HandleFunc("/api/v2/policy/"+tc.path, func(w http.ResponseWriter, r *http.Request) {
				writes++
				assert.Equal(t, http.MethodPatch, r.Method)
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				settings := body["policy"].(map[string]any)
				if tc.assets == nil {
					assert.NotContains(t, settings, "imposedAssets", "the tool must not insert existing assets")
				} else {
					assert.Equal(t, tc.assets, settings["imposedAssets"])
				}
				runaiclienttest.WriteJSON(w, 200, testLegacyPolicy())
			})
			in := testLegacyPolicyInput(tc.kind, actionUpdate)
			if tc.assets != nil {
				in.Request["policy"].(map[string]any)["imposedAssets"] = tc.assets
			}

			// 2. Call the tool.
			out, err := manageWorkloadPolicyRun(context.Background(), c, in)

			// 3. Check the result.
			require.NoError(t, err)
			assert.Equal(t, testLegacyPolicy(), out)
			assert.Equal(t, 1, writes)
		})
	}
}

func TestGVKPolicyFilterExpressionsArePreserved(t *testing.T) {
	for _, filters := range [][]string{
		{"scope==project", "type==/Pod/v1||apps/Deployment/v1"},
		{"scope!=tenant", "scope!=cluster"},
		{"type=@/Deployment/", "type=^apps/", "type=$/v1"},
	} {
		t.Run(strings.Join(filters, ","), func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/policies", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, strings.Join(filters, ","), r.URL.Query().Get("filterBy"))
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"items": []any{}})
			})
			_, err := listWorkloadPoliciesRun(context.Background(), client, listWorkloadPoliciesInput{PolicyVersion: versionGVK, FilterBy: filters})
			require.NoError(t, err)
		})
	}
}
