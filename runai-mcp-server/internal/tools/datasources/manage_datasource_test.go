/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDatasourceID = "11111111-1111-4111-8111-111111111111"

type datasourceCase struct {
	kind                         assets.AssetKind
	endpoint, specKey, queryType string
	spec                         map[string]any
}

func datasourceCases() []datasourceCase {
	return []datasourceCase{
		{assets.ASSETKIND_NFS, "nfs", "nfs", "nfs", map[string]any{"path": "/data", "server": "nfs.example", "mountPath": "/mnt"}},
		{assets.ASSETKIND_PVC, "pvc", "pvc", "pvc", map[string]any{"path": "/mnt", "claimName": "data", "existingPvc": true}},
		{assets.ASSETKIND_S3, "s3", "s3", "s3", map[string]any{"bucket": "data", "path": "/mnt"}},
		{assets.ASSETKIND_GIT, "git", "git", "git", map[string]any{"repository": "https://example.com/repo.git", "path": "/mnt"}},
		{assets.ASSETKIND_HOST_PATH, "host-path", "hostPath", "hostPath", map[string]any{"path": "/data", "mountPath": "/mnt"}},
		{assets.ASSETKIND_CONFIG_MAP, "config-map", "config_map", "configMap", map[string]any{"configMap": "config", "mountPath": "/mnt", "defaultMode": "0644"}},
		{assets.ASSETKIND_SECRET_VOLUME, "secrets", "secret", "secretVolume", map[string]any{"credentialAssetId": testDatasourceID, "mountPath": "/mnt"}},
	}
}

func datasourceCreateMeta(tc datasourceCase) map[string]any {
	meta := map[string]any{"name": "data", "scope": string(assets.SCOPE_TENANT)}
	if tc.kind == assets.ASSETKIND_CONFIG_MAP || tc.kind == assets.ASSETKIND_SECRET_VOLUME {
		meta["scope"] = string(assets.SCOPE_PROJECT)
		meta["projectId"] = int32(7)
	}
	return meta
}

func datasourceResponse(tc datasourceCase, count int64) map[string]any {
	meta := datasourceCreateMeta(tc)
	meta["id"] = testDatasourceID
	meta["kind"] = string(tc.kind)
	meta["updateCount"] = count
	return map[string]any{"meta": meta, "spec": tc.spec}
}

func datasourceRequest(tc datasourceCase, create bool) map[string]any {
	meta := map[string]any{"name": "data"}
	if create {
		meta = datasourceCreateMeta(tc)
	}

	raw := map[string]any{"meta": meta, "spec": tc.spec}
	if !create {
		switch tc.kind {
		case assets.ASSETKIND_CONFIG_MAP, assets.ASSETKIND_SECRET_VOLUME:
			delete(raw, "spec")
		case assets.ASSETKIND_PVC:
			raw["spec"] = map[string]any{"path": tc.spec["path"]}
		}
	}
	return raw
}

func TestDatasourceLifecycle(t *testing.T) {
	for _, tc := range datasourceCases() {
		for _, action := range []common.WriteAction{common.ActionCreate, common.ActionUpdate, common.ActionDelete} {
			t.Run(string(tc.kind)+"/"+string(action), func(t *testing.T) {
				client, mux := runaiclienttest.NewServer(t)
				collectionPath := "/api/v1/asset/datasource/" + tc.endpoint
				path := collectionPath + "/" + testDatasourceID
				method := http.MethodPut
				if action == common.ActionCreate {
					path = collectionPath
					method = http.MethodPost
				}
				if action == common.ActionDelete {
					method = http.MethodDelete
				}
				request := datasourceRequest(tc, action == common.ActionCreate)
				if action == common.ActionUpdate {
					request["meta"].(map[string]any)["name"] = "updated"
				}
				expectedJSON, err := json.Marshal(request)
				require.NoError(t, err)
				var writes atomic.Int32
				handler := func(w http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodGet {
						assert.Equal(t, collectionPath+"/"+testDatasourceID, req.URL.Path)
						response := datasourceResponse(tc, 2)
						if action == common.ActionDelete && tc.kind == assets.ASSETKIND_NFS {
							response["spec"] = nil
						}
						runaiclienttest.WriteJSON(w, http.StatusOK, response)
						return
					}
					writes.Add(1)
					assert.Equal(t, path, req.URL.Path)
					assert.Equal(t, method, req.Method)
					if action == common.ActionDelete {
						w.WriteHeader(http.StatusAccepted)
						return
					}
					body, err := io.ReadAll(req.Body)
					if !assert.NoError(t, err) {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					assert.JSONEq(t, string(expectedJSON), string(body))
					response := datasourceResponse(tc, 3)
					if action == common.ActionUpdate {
						response["meta"].(map[string]any)["name"] = "updated"
					}
					runaiclienttest.WriteJSON(w, http.StatusAccepted, response)
				}
				mux.HandleFunc(collectionPath, handler)
				mux.HandleFunc(collectionPath+"/"+testDatasourceID, handler)
				in := manageDatasourceInput{Action: action, Kind: tc.kind, Confirm: true}
				if action != common.ActionCreate {
					in.DatasourceID = testDatasourceID
				}
				if action != common.ActionDelete {
					in.Request = request
				}
				if action == common.ActionUpdate {
					count := int64(2)
					in.ExpectedUpdateCount = &count
				}
				require.NoError(t, in.Validate())
				out, err := manageDatasourceRun(context.Background(), client, in)
				require.NoError(t, err)
				require.True(t, out.Applied)
				assert.Equal(t, testDatasourceID, out.DatasourceID)
				assert.Empty(t, out.Warnings)
				assert.Equal(t, int32(1), writes.Load())
				if action == common.ActionDelete {
					assert.Nil(t, out.Datasource)
				} else {
					assert.Equal(t, tc.spec, out.Datasource["spec"])
					assert.Equal(t, request["meta"].(map[string]any)["name"], out.Name)
				}
			})
		}
	}
}

func TestDatasourceRejectsStaleUpdates(t *testing.T) {
	for _, changeAfterConfirmation := range []bool{false, true} {
		t.Run(map[bool]string{false: "already_stale", true: "changed_after_confirmation"}[changeAfterConfirmation], func(t *testing.T) {
			tc := datasourceCases()[0]
			client, mux := runaiclienttest.NewServer(t)
			var changed atomic.Bool
			changed.Store(!changeAfterConfirmation)
			mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
				if !assert.Equal(t, http.MethodGet, r.Method, "a stale update must never write") {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				count := int64(2)
				if changed.Load() {
					count = 3
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, datasourceResponse(tc, count))
			})
			elicitor := &datasourceTestElicitor{action: "accept", onElicit: func() { changed.Store(true) }}
			ctx := runaimcp.WithSession(context.Background(), elicitor)
			count := int64(2)
			_, err := manageDatasourceRun(ctx, client, manageDatasourceInput{Action: common.ActionUpdate, Kind: tc.kind, DatasourceID: testDatasourceID, Request: datasourceRequest(tc, false), ExpectedUpdateCount: &count, Confirm: true})
			require.ErrorIs(t, err, runaimcp.ErrConflict)
			assert.Equal(t, changeAfterConfirmation, elicitor.params != nil)
		})
	}
}

func TestDatasourceValidation(t *testing.T) {
	tc := datasourceCases()[0]
	count := int64(2)
	tests := []struct {
		name    string
		in      manageDatasourceInput
		wantErr string
	}{
		{"unknown_kind", manageDatasourceInput{Action: common.ActionCreate, Kind: "unknown", Request: datasourceRequest(tc, true)}, "kind"},
		{"unsupported_api_kind", manageDatasourceInput{Action: common.ActionCreate, Kind: assets.ASSETKIND_COMPUTE, Request: datasourceRequest(tc, true)}, "kind"},
		{"invalid_scope_filter", manageDatasourceInput{Action: common.ActionDelete, Kind: assets.ASSETKIND_NFS, Name: "data", Scope: assets.SCOPE_PROJECT, ScopeName: "project==other"}, "scopeName"},
		{"unknown_action", manageDatasourceInput{Action: "edit", Kind: assets.ASSETKIND_NFS}, "action"},
		{"missing_request", manageDatasourceInput{Action: common.ActionCreate, Kind: assets.ASSETKIND_NFS}, "request"},
		{"missing_id", manageDatasourceInput{Action: common.ActionDelete, Kind: assets.ASSETKIND_NFS}, "datasourceId"},
		{"bad_id", manageDatasourceInput{Action: common.ActionDelete, Kind: assets.ASSETKIND_NFS, DatasourceID: "invalid"}, "datasourceId"},
		{"missing_count", manageDatasourceInput{Action: common.ActionUpdate, Kind: assets.ASSETKIND_NFS, DatasourceID: testDatasourceID, Request: datasourceRequest(tc, false)}, "expectedUpdateCount"},
		{"delete_request", manageDatasourceInput{Action: common.ActionDelete, Kind: assets.ASSETKIND_NFS, DatasourceID: testDatasourceID, Request: datasourceRequest(tc, false)}, "request"},
		{"create_with_id", manageDatasourceInput{Action: common.ActionCreate, Kind: assets.ASSETKIND_NFS, DatasourceID: testDatasourceID, Request: datasourceRequest(tc, true)}, "datasourceId"},
		{"id_and_name", manageDatasourceInput{Action: common.ActionDelete, Kind: tc.kind, DatasourceID: testDatasourceID, Name: "data"}, "mutually exclusive"},
		{"blank_name", manageDatasourceInput{Action: common.ActionDelete, Kind: tc.kind, Name: " "}, "name"},
		{"scope_without_name", manageDatasourceInput{Action: common.ActionDelete, Kind: tc.kind, DatasourceID: testDatasourceID, Scope: assets.SCOPE_PROJECT}, "require name"},
		{"scope_name_without_scope", manageDatasourceInput{Action: common.ActionDelete, Kind: tc.kind, Name: "data", ScopeName: "research"}, "scope"},
		{"two_scope_selectors", manageDatasourceInput{Action: common.ActionDelete, Kind: tc.kind, Name: "data", Scope: assets.SCOPE_PROJECT, ScopeID: "7", ScopeName: "research"}, "mutually exclusive"},
		{"invalid_project_scope_id", manageDatasourceInput{Action: common.ActionDelete, Kind: tc.kind, Name: "data", Scope: assets.SCOPE_PROJECT, ScopeID: "invalid"}, "scopeId"},
		{"scope_name_filter_injection", manageDatasourceInput{Action: common.ActionDelete, Kind: tc.kind, Name: "data", Scope: assets.SCOPE_PROJECT, ScopeName: "name,other"}, "scopeName"},
		{"create_name_selector", manageDatasourceInput{Action: common.ActionCreate, Kind: tc.kind, Name: "data", Request: datasourceRequest(tc, true)}, "request.meta"},
		{"ignore_missing_on_update", manageDatasourceInput{Action: common.ActionUpdate, Kind: assets.ASSETKIND_NFS, DatasourceID: testDatasourceID, ExpectedUpdateCount: &count, Request: datasourceRequest(tc, false), IgnoreMissing: true}, "ignoreMissing"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { require.ErrorContains(t, test.in.Validate(), test.wantErr) })
	}

	t.Run("system_create", func(t *testing.T) {
		raw := datasourceRequest(tc, true)
		raw["meta"].(map[string]any)["scope"] = string(assets.SCOPE_SYSTEM)
		require.ErrorContains(t, (manageDatasourceInput{Action: common.ActionCreate, Kind: tc.kind, Request: raw}).Validate(), "read-only")
	})

	t.Run("unknown_request_field", func(t *testing.T) {
		raw := datasourceRequest(tc, true)
		raw["spec"].(map[string]any)["futureField"] = true
		require.ErrorContains(t, (manageDatasourceInput{Action: common.ActionCreate, Kind: tc.kind, Request: raw}).Validate(), "unknown field")
	})
}

func TestDatasourceRejectedWritesReturnErrors(t *testing.T) {
	for _, action := range []common.WriteAction{common.ActionCreate, common.ActionUpdate} {
		t.Run(string(action), func(t *testing.T) {
			tc := datasourceCases()[0]
			client, mux := runaiclienttest.NewServer(t)
			handler := func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodGet {
					runaiclienttest.WriteJSON(w, http.StatusOK, datasourceResponse(tc, 2))
					return
				}
				runaiclienttest.WriteJSON(w, http.StatusConflict, assets.Error{Code: http.StatusConflict, Message: "asset conflict"})
			}
			mux.HandleFunc("/api/v1/asset/datasource/nfs", handler)
			mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, handler)
			in := manageDatasourceInput{Action: action, Kind: tc.kind, Request: datasourceRequest(tc, action == common.ActionCreate), Confirm: true}
			if action == common.ActionUpdate {
				count := int64(2)
				in.ExpectedUpdateCount = &count
				in.DatasourceID = testDatasourceID
			}
			out, err := manageDatasourceRun(context.Background(), client, in)
			require.ErrorIs(t, err, runaimcp.ErrConflict)
			assert.Contains(t, err.Error(), "asset conflict")
			assert.False(t, out.Applied)
		})
	}
}

func TestDatasourceDeleteGuards(t *testing.T) {
	for _, scope := range []string{"system", "missing"} {
		t.Run(scope, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
				if !assert.Equal(t, http.MethodGet, r.Method) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if scope == "missing" {
					runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "missing"})
					return
				}
				raw := datasourceResponse(datasourceCases()[0], 2)
				raw["meta"].(map[string]any)["scope"] = string(assets.SCOPE_SYSTEM)
				runaiclienttest.WriteJSON(w, http.StatusOK, raw)
			})
			out, err := manageDatasourceRun(context.Background(), client, manageDatasourceInput{Action: common.ActionDelete, Kind: assets.ASSETKIND_NFS, DatasourceID: testDatasourceID, IgnoreMissing: true, Confirm: true})
			if scope == "system" {
				require.ErrorIs(t, err, runaimcp.ErrForbidden)
			} else {
				require.NoError(t, err)
				assert.False(t, out.Applied)
				require.NotEmpty(t, out.Warnings)
			}
		})
	}
}

func TestDatasourceRejectsImmutableSpecUpdates(t *testing.T) {
	for _, tc := range datasourceCases() {
		if tc.kind != assets.ASSETKIND_PVC && tc.kind != assets.ASSETKIND_CONFIG_MAP && tc.kind != assets.ASSETKIND_SECRET_VOLUME {
			continue
		}
		t.Run(string(tc.kind), func(t *testing.T) {
			count := int64(2)
			raw := datasourceRequest(tc, true)
			delete(raw["meta"].(map[string]any), "scope")
			in := manageDatasourceInput{Action: common.ActionUpdate, Kind: tc.kind, DatasourceID: testDatasourceID, Request: raw, ExpectedUpdateCount: &count}
			require.ErrorContains(t, in.Validate(), "unknown field")
		})
	}
}

func TestDatasourceResponseUpdateCounterPrecision(t *testing.T) {
	tc := datasourceCases()[0]
	client, mux := runaiclienttest.NewServer(t)
	const count = int64(9007199254740993)
	mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, datasourceResponse(tc, count))
	})
	snapshot, err := getDatasource(context.Background(), client, tc.kind, testDatasourceID)
	require.NoError(t, err)
	require.NotNil(t, snapshot.updateCount)
	assert.Equal(t, count, *snapshot.updateCount)
}

type datasourceTestElicitor struct {
	action   string
	onElicit func()
	params   *mcp.ElicitParams
}

func (e *datasourceTestElicitor) Elicit(_ context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error) {
	e.params = params
	if e.onElicit != nil {
		e.onElicit()
	}
	return &mcp.ElicitResult{Action: e.action}, nil
}

func (*datasourceTestElicitor) InitializeParams() *mcp.InitializeParams {
	return &mcp.InitializeParams{Capabilities: &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{}}}
}

func (*datasourceTestElicitor) ID() string { return "datasource-test" }

func TestDatasourceDeleteRejectsChangeDuringApproval(t *testing.T) {
	tc := datasourceCases()[0]
	client, mux := runaiclienttest.NewServer(t)
	var changed, deleted atomic.Bool
	mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleted.Store(true)
			w.WriteHeader(http.StatusAccepted)
			return
		}
		raw := datasourceResponse(tc, 2)
		if changed.Load() {
			raw["meta"].(map[string]any)["name"] = "changed-after-prompt"
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, raw)
	})
	elicitor := &datasourceTestElicitor{action: "accept", onElicit: func() { changed.Store(true) }}
	ctx := runaimcp.WithSession(context.Background(), elicitor)
	_, err := manageDatasourceRun(ctx, client, manageDatasourceInput{Action: common.ActionDelete, Kind: assets.ASSETKIND_NFS, DatasourceID: testDatasourceID, Confirm: true})
	require.ErrorIs(t, err, runaimcp.ErrConflict)
	assert.False(t, deleted.Load())
}

func TestDatasourceInteractiveApprovalCannotBeBypassed(t *testing.T) {
	for _, action := range []common.WriteAction{common.ActionUpdate, common.ActionDelete} {
		for _, answer := range []string{"decline", "accept"} {
			t.Run(string(action)+"/"+answer, func(t *testing.T) {
				tc := datasourceCases()[0]
				client, mux := runaiclienttest.NewServer(t)
				var writes atomic.Int32
				mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet {
						writes.Add(1)
					}
					if r.Method == http.MethodDelete {
						w.WriteHeader(http.StatusAccepted)
						return
					}
					runaiclienttest.WriteJSON(w, http.StatusOK, datasourceResponse(tc, 2))
				})
				count := int64(2)
				in := manageDatasourceInput{Action: action, Kind: assets.ASSETKIND_NFS, DatasourceID: testDatasourceID, Confirm: true}
				if action == common.ActionUpdate {
					in.ExpectedUpdateCount = &count
					in.Request = datasourceRequest(tc, false)
				}
				elicitor := &datasourceTestElicitor{action: answer}
				_, err := manageDatasourceRun(runaimcp.WithSession(context.Background(), elicitor), client, in)
				require.NotNil(t, elicitor.params)
				if answer == "accept" {
					require.NoError(t, err)
					assert.Equal(t, int32(1), writes.Load())
				} else {
					require.ErrorIs(t, err, runaimcp.ErrDeclined)
					assert.Zero(t, writes.Load())
				}
			})
		}
	}
}

func TestDatasourceUpdateApprovalNamesChangedFields(t *testing.T) {
	tc := datasourceCases()[0]
	client, mux := runaiclienttest.NewServer(t)
	tc.spec["path"] = "/" + strings.Repeat("directory/", 12)
	mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, datasourceResponse(tc, 2))
	})
	spec := map[string]any{}
	for key, value := range tc.spec {
		spec[key] = value
	}
	spec["server"] = "new-nfs.example"
	count := int64(2)
	elicitor := &datasourceTestElicitor{action: "decline"}
	_, err := manageDatasourceRun(runaimcp.WithSession(context.Background(), elicitor), client, manageDatasourceInput{
		Action: common.ActionUpdate, Kind: assets.ASSETKIND_NFS, DatasourceID: testDatasourceID, ExpectedUpdateCount: &count,
		Request: map[string]any{"meta": map[string]any{"name": "data"}, "spec": spec},
	})
	require.ErrorIs(t, err, runaimcp.ErrDeclined)
	require.NotNil(t, elicitor.params)
	assert.Contains(t, elicitor.params.Message, "spec.server")
	assert.Contains(t, elicitor.params.Message, "new-nfs.example")
}

func TestPVCUpdateDoesNotResizeUnchangedSize(t *testing.T) {
	for _, requestedSize := range []string{"1Gi", "2Gi"} {
		t.Run(requestedSize, func(t *testing.T) {
			tc := datasourceCases()[1]
			tc.spec = map[string]any{"path": "/mnt", "claimName": "data", "existingPvc": false, "claimInfo": map[string]any{"size": "1Gi", "storageClass": "standard", "addedAttrValues": []any{}}}
			client, mux := runaiclienttest.NewServer(t)
			var puts atomic.Int32
			mux.HandleFunc("/api/v1/asset/datasource/pvc/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPut {
					puts.Add(1)
					var body map[string]any
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					claim := body["spec"].(map[string]any)["claimInfo"].(map[string]any)
					if requestedSize == "1Gi" {
						assert.NotContains(t, claim, "size", "unchanged sizes trigger unnecessary resize validation")
					} else {
						assert.Equal(t, requestedSize, claim["size"])
					}
					assert.Contains(t, claim, "addedAttrValues")
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, datasourceResponse(tc, 2))
			})
			raw := datasourceResponse(tc, 2)
			snapshot, err := newDatasourceSnapshot(raw, assets.ASSETKIND_PVC)
			require.NoError(t, err)
			snapshot = prepareDatasourceUpdate(snapshot)
			require.Empty(t, snapshot.updateBlockedReason)
			claim := snapshot.updateRequest["spec"].(map[string]any)["claimInfo"].(map[string]any)
			claim["size"] = requestedSize
			in := manageDatasourceInput{Action: common.ActionUpdate, Kind: assets.ASSETKIND_PVC, DatasourceID: testDatasourceID, ExpectedUpdateCount: snapshot.updateCount, Request: snapshot.updateRequest, Confirm: true}
			require.NoError(t, in.Validate())
			_, err = manageDatasourceRun(context.Background(), client, in)
			require.NoError(t, err)
			assert.Equal(t, int32(1), puts.Load())
			assert.Equal(t, requestedSize, claim["size"], "the caller's request must remain unchanged")
		})
	}
}

func TestDatasourceAcceptedMutationWithInvalidResponse(t *testing.T) {
	valid, err := json.Marshal(datasourceResponse(datasourceCases()[0], 2))
	require.NoError(t, err)
	for _, action := range []common.WriteAction{common.ActionCreate, common.ActionUpdate} {
		for _, response := range []struct{ name, body string }{
			{"invalid_json", "not-json"},
			{"null", "null"},
			{"missing_metadata", `{"meta":{},"spec":{}}`},
			{"trailing_json", string(valid) + `{}`},
			{"unchanged_name", string(valid)},
		} {
			if response.name == "unchanged_name" && action != common.ActionUpdate {
				continue
			}
			t.Run(string(action)+"/"+response.name, func(t *testing.T) {
				tc := datasourceCases()[0]
				client, mux := runaiclienttest.NewServer(t)
				var writes atomic.Int32
				handler := func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodGet {
						runaiclienttest.WriteJSON(w, http.StatusOK, datasourceResponse(tc, 2))
						return
					}
					writes.Add(1)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusAccepted)
					_, err := w.Write([]byte(response.body))
					assert.NoError(t, err)
				}
				mux.HandleFunc("/api/v1/asset/datasource/nfs", handler)
				mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, handler)
				in := manageDatasourceInput{Action: action, Kind: assets.ASSETKIND_NFS, Request: datasourceRequest(tc, action == common.ActionCreate), Confirm: true}
				if response.name == "unchanged_name" {
					in.Request["meta"].(map[string]any)["name"] = "renamed"
				}
				if action == common.ActionUpdate {
					count := int64(2)
					in.ExpectedUpdateCount = &count
					in.DatasourceID = testDatasourceID
				}
				require.NoError(t, in.Validate())
				out, err := manageDatasourceRun(context.Background(), client, in)
				require.NoError(t, err, "a successful HTTP status must not suggest retrying an already accepted mutation")
				require.True(t, out.Applied)
				require.NotEmpty(t, out.Warnings)
				assert.Contains(t, out.Warnings[0], "Do not retry")
				assert.Nil(t, out.Datasource, "an old or unverified snapshot must not be presented as the mutation result")
				assert.Equal(t, int32(1), writes.Load())
			})
		}
	}
}

func TestDatasourceMCPRoundTrip(t *testing.T) {
	for _, readOnly := range []bool{false, true} {
		t.Run(map[bool]string{false: "write_enabled", true: "read_only"}[readOnly], func(t *testing.T) {
			ctx := context.Background()
			client, mux := runaiclienttest.NewServer(t)
			tc := datasourceCases()[0]
			var writes atomic.Int32
			mux.HandleFunc("/api/v1/asset/datasource", func(w http.ResponseWriter, r *http.Request) {
				raw := datasourceResponse(tc, 2)
				raw["spec"] = map[string]any{tc.specKey: tc.spec}
				future := map[string]any{"meta": map[string]any{"id": "external-id", "kind": "future-kind", "scope": "future-scope"}, "spec": "future-spec"}
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{raw, future}})
			})
			mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
				response := datasourceResponse(tc, 2)
				if r.Method != http.MethodGet {
					writes.Add(1)
					var request assets.NFSUpdateRequest
					if err := json.NewDecoder(r.Body).Decode(&request); !assert.NoError(t, err) {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					response["meta"].(map[string]any)["name"] = request.Meta.Name
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, response)
			})
			srv := mcp.NewServer(&mcp.Implementation{Name: "datasource-test", Version: "1"}, nil)
			registry := runaimcp.New(srv, runaimcp.Options{Logger: zerolog.Nop(), ReadOnly: readOnly})
			Register(registry, client)
			clientTransport, serverTransport := mcp.NewInMemoryTransports()
			serverSession, err := srv.Connect(ctx, serverTransport, nil)
			require.NoError(t, err)
			defer serverSession.Close()
			session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, clientTransport, nil)
			require.NoError(t, err)
			defer session.Close()
			definitions, err := session.ListTools(ctx, &mcp.ListToolsParams{})
			require.NoError(t, err)
			names := make([]string, 0, len(definitions.Tools))
			for _, tool := range definitions.Tools {
				names = append(names, tool.Name)
			}
			require.Contains(t, names, "list_datasources")
			listed, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_datasources", Arguments: map[string]any{}})
			require.NoError(t, err)
			require.False(t, listed.IsError)
			var list listDatasourcesOutput
			decodeDatasourceToolResult(t, listed, &list)
			require.Len(t, list.Datasources, 2)
			blocked := list.Datasources[1]
			assert.Equal(t, "future-kind", blocked.Kind)
			assert.Equal(t, "future-scope", blocked.Scope)
			assert.NotEmpty(t, blocked.UpdateBlockedReason)
			assert.Nil(t, blocked.UpdateRequest)
			assert.Nil(t, blocked.ExpectedUpdateCount)
			item := list.Datasources[0]
			require.Equal(t, string(assets.ASSETKIND_NFS), item.Kind)
			require.NotNil(t, item.ExpectedUpdateCount)
			if readOnly {
				require.NotContains(t, names, "manage_datasource")
				return
			}
			require.Contains(t, names, "manage_datasource")
			request := item.UpdateRequest
			request["meta"].(map[string]any)["name"] = "renamed"
			input := manageDatasourceInput{Action: common.ActionUpdate, Kind: assets.AssetKind(item.Kind), Name: item.Name, ExpectedUpdateCount: item.ExpectedUpdateCount, Request: request}
			denied, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "manage_datasource", Arguments: input})
			require.NoError(t, err)
			require.True(t, denied.IsError)
			assert.Zero(t, writes.Load())
			input.Confirm = true
			applied, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "manage_datasource", Arguments: input})
			require.NoError(t, err)
			require.False(t, applied.IsError)
			var updated manageDatasourceOutput
			decodeDatasourceToolResult(t, applied, &updated)
			require.True(t, updated.Applied)
			assert.Equal(t, "renamed", updated.Name)
			assert.Empty(t, updated.Warnings)
			assert.Equal(t, int32(1), writes.Load())
		})
	}
}

func decodeDatasourceToolResult(t *testing.T, result *mcp.CallToolResult, out any) {
	t.Helper()
	for _, content := range result.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			require.NoError(t, json.Unmarshal([]byte(text.Text), out))
			return
		}
	}
	t.Fatal("tool returned no JSON text content")
}

func TestDatasourceCreateVerifiesScopeIdentity(t *testing.T) {
	for _, tc := range []struct {
		scope            assets.Scope
		field            string
		requested, other any
	}{
		{assets.SCOPE_PROJECT, "projectId", int32(7), int32(8)},
		{assets.SCOPE_DEPARTMENT, "departmentId", "department", "other-department"},
		{assets.SCOPE_CLUSTER, "clusterId", testDatasourceID, "22222222-2222-4222-8222-222222222222"},
	} {
		for _, matches := range []bool{true, false} {
			name := "different_target"
			if matches {
				name = "same_target"
			}
			t.Run(string(tc.scope)+"/"+name, func(t *testing.T) {
				client, mux := runaiclienttest.NewServer(t)
				request := datasourceRequest(datasourceCases()[0], true)
				meta := request["meta"].(map[string]any)
				meta["scope"], meta[tc.field] = string(tc.scope), tc.requested
				mux.HandleFunc("/api/v1/asset/datasource/nfs", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodPost, r.Method)
					raw := datasourceResponse(datasourceCases()[0], 1)
					returned := raw["meta"].(map[string]any)
					returned["scope"], returned[tc.field] = string(tc.scope), tc.requested
					if !matches {
						returned[tc.field] = tc.other
					}
					runaiclienttest.WriteJSON(w, http.StatusCreated, raw)
				})
				in := manageDatasourceInput{Action: common.ActionCreate, Kind: assets.ASSETKIND_NFS, Request: request}
				require.NoError(t, in.Validate())
				out, err := manageDatasourceRun(context.Background(), client, in)
				require.NoError(t, err)
				assert.True(t, out.Applied)
				if matches {
					assert.Empty(t, out.Warnings)
					assert.NotNil(t, out.Datasource)
				} else {
					assert.NotEmpty(t, out.Warnings)
					assert.Nil(t, out.Datasource)
				}
			})
		}
	}
}

func TestDatasourceManageByName(t *testing.T) {
	for _, target := range []struct {
		label         string
		scope         assets.Scope
		id, scopeName string
	}{
		{"project_name", assets.SCOPE_PROJECT, "", "project"},
		{"project_id", assets.SCOPE_PROJECT, "7", ""},
		{"department_name", assets.SCOPE_DEPARTMENT, "", "department"},
		{"department_id", assets.SCOPE_DEPARTMENT, "dept", ""},
		{"cluster_name", assets.SCOPE_CLUSTER, "", "cluster"},
		{"cluster_id", assets.SCOPE_CLUSTER, testDatasourceID, ""},
		{"without_scope", "", "", ""},
	} {
		for _, action := range []common.WriteAction{common.ActionUpdate, common.ActionDelete} {
			t.Run(target.label+"/"+string(action), func(t *testing.T) {
				client, mux := runaiclienttest.NewServer(t)
				tc := datasourceCases()[0]
				raw := datasourceResponse(tc, 2)
				meta := raw["meta"].(map[string]any)
				if target.scope != "" {
					meta["scope"] = string(target.scope)
				}
				query := url.Values{"name": {"data"}, "type": {"nfs"}}
				if target.scope != "" {
					query.Set("scope", string(target.scope))
				}
				path, key, id := "", "", ""
				switch target.scope {
				case assets.SCOPE_PROJECT:
					meta["projectId"] = int32(7)
					query.Set("projectId", "7")
					path, key, id = "/api/v1/org-unit/projects", "projects", "7"
				case assets.SCOPE_DEPARTMENT:
					meta["departmentId"] = "dept"
					query.Set("departmentId", "dept")
					path, key, id = "/api/v1/org-unit/departments", "departments", "dept"
				case assets.SCOPE_CLUSTER:
					meta["clusterId"] = testDatasourceID
					query.Set("clusterId", testDatasourceID)
					path, key, id = "/api/v1/clusters/minimal", "clusters", testDatasourceID
				}
				if target.scopeName != "" {
					mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
						assert.Equal(t, "name=="+target.scopeName, r.URL.Query().Get("filterBy"))
						assert.Equal(t, "2", r.URL.Query().Get("limit"))
						entry := map[string]any{"id": id}
						if target.scope == assets.SCOPE_CLUSTER {
							entry = map[string]any{"uuid": id, "name": target.scopeName}
						}
						runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{key: []any{entry}})
					})
				}
				mux.HandleFunc("/api/v1/asset/datasource", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, query, r.URL.Query())
					listed := maps.Clone(raw)
					listed["spec"] = map[string]any{"nfs": tc.spec}
					other := maps.Clone(listed)
					otherMeta := maps.Clone(meta)
					otherMeta["name"] = "data-backup"
					other["meta"] = otherMeta
					entries := []any{other, listed}
					if target.scope != "" {
						ancestor := maps.Clone(listed)
						ancestorMeta := maps.Clone(meta)
						ancestorMeta["scope"] = string(assets.SCOPE_TENANT)
						ancestor["meta"] = ancestorMeta
						foreign := maps.Clone(listed)
						foreignMeta := maps.Clone(meta)
						foreignMeta["id"] = "22222222-2222-4222-8222-222222222222"
						switch target.scope {
						case assets.SCOPE_PROJECT:
							foreignMeta["projectId"] = int32(8)
						case assets.SCOPE_DEPARTMENT:
							foreignMeta["departmentId"] = "other-department"
						case assets.SCOPE_CLUSTER:
							foreignMeta["clusterId"] = "22222222-2222-4222-8222-222222222222"
						}
						foreign["meta"] = foreignMeta
						entries = append(entries, ancestor, foreign)
					}
					runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": entries})
				})
				var writes atomic.Int32
				mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet {
						writes.Add(1)
					}
					if r.Method == http.MethodDelete {
						w.WriteHeader(http.StatusAccepted)
						return
					}
					response := maps.Clone(raw)
					returnedMeta := maps.Clone(meta)
					if r.Method == http.MethodPut {
						returnedMeta["name"] = "renamed"
					}
					response["meta"] = returnedMeta
					runaiclienttest.WriteJSON(w, http.StatusOK, response)
				})
				in := manageDatasourceInput{Kind: tc.kind, Action: action, Name: "data", Scope: target.scope, ScopeID: target.id, ScopeName: target.scopeName, Confirm: true}
				if action == common.ActionUpdate {
					count := int64(2)
					in.ExpectedUpdateCount = &count
					in.Request = datasourceRequest(tc, false)
					in.Request["meta"].(map[string]any)["name"] = "renamed"
				}
				require.NoError(t, in.Validate())
				out, err := manageDatasourceRun(context.Background(), client, in)
				require.NoError(t, err)
				assert.True(t, out.Applied)
				assert.Empty(t, out.Warnings)
				assert.Equal(t, testDatasourceID, out.DatasourceID)
				assert.Equal(t, int32(1), writes.Load())
			})
		}
	}
}

func TestDatasourceNameLookupGuards(t *testing.T) {
	for _, scenario := range []string{"ambiguous", "missing", "ignored_missing", "renamed", "confirmation_required", "scope_missing", "scope_ambiguous", "scope_empty_id", "project_zero", "project_overflow", "cluster_invalid", "malformed_entry"} {
		t.Run(scenario, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			tc := datasourceCases()[0]
			if strings.HasPrefix(scenario, "scope_") {
				mux.HandleFunc("/api/v1/clusters/minimal", func(w http.ResponseWriter, r *http.Request) {
					clusters := []any{}
					if scenario == "scope_ambiguous" {
						clusters = []any{map[string]any{"uuid": testDatasourceID}, map[string]any{"uuid": "22222222-2222-4222-8222-222222222222"}}
					}
					if scenario == "scope_empty_id" {
						clusters = []any{map[string]any{"uuid": ""}}
					}
					runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"clusters": clusters})
				})
			}
			mux.HandleFunc("/api/v1/asset/datasource", func(w http.ResponseWriter, r *http.Request) {
				if scenario == "cluster_invalid" {
					assert.Equal(t, "invalid", r.URL.Query().Get("clusterId"))
					runaiclienttest.WriteJSON(w, http.StatusBadRequest, assets.Error{Code: http.StatusBadRequest, Message: "invalid cluster ID"})
					return
				}
				assert.False(t, strings.HasPrefix(scenario, "scope_") || scenario == "project_overflow", "scope resolution errors must not fall back to a broader asset search")
				raw := datasourceResponse(tc, 2)
				raw["spec"] = map[string]any{"nfs": tc.spec}
				if scenario == "project_zero" {
					assert.Equal(t, "0", r.URL.Query().Get("projectId"))
					meta := raw["meta"].(map[string]any)
					meta["scope"], meta["projectId"] = string(assets.SCOPE_PROJECT), int32(7)
				}
				entries := []any{raw}
				if scenario == "ambiguous" {
					other := maps.Clone(raw)
					meta := maps.Clone(raw["meta"].(map[string]any))
					meta["id"] = "22222222-2222-4222-8222-222222222222"
					other["meta"] = meta
					entries = append(entries, other)
				}
				if scenario == "malformed_entry" {
					entries = []any{"not-an-object"}
				}
				if scenario == "missing" || scenario == "ignored_missing" {
					entries = []any{}
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": entries})
			})
			var writes atomic.Int32
			mux.HandleFunc("/api/v1/asset/datasource/nfs/"+testDatasourceID, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					writes.Add(1)
				}
				raw := datasourceResponse(tc, 2)
				if scenario == "renamed" {
					raw["meta"].(map[string]any)["name"] = "other"
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, raw)
			})
			in := manageDatasourceInput{Kind: tc.kind, Action: common.ActionDelete, Name: "data", Confirm: scenario != "confirmation_required", IgnoreMissing: scenario == "ignored_missing"}
			if strings.HasPrefix(scenario, "scope_") {
				in.Scope, in.ScopeName = assets.SCOPE_CLUSTER, "cluster"
			}
			if scenario == "cluster_invalid" {
				in.Scope, in.ScopeID = assets.SCOPE_CLUSTER, "invalid"
			}
			if strings.HasPrefix(scenario, "project_") {
				in.Scope, in.ScopeID = assets.SCOPE_PROJECT, "0"
				if scenario == "project_overflow" {
					in.ScopeID = "2147483648"
				}
			}
			require.NoError(t, in.Validate())
			out, err := manageDatasourceRun(context.Background(), client, in)
			switch scenario {
			case "ambiguous":
				require.ErrorIs(t, err, runaimcp.ErrValidation)
				assert.ErrorContains(t, err, "ambiguous_datasource")
			case "missing", "scope_missing", "project_zero":
				require.ErrorIs(t, err, runaimcp.ErrNotFound)
			case "ignored_missing":
				require.NoError(t, err)
				assert.NotEmpty(t, out.Warnings)
			case "scope_ambiguous", "project_overflow", "cluster_invalid":
				require.ErrorIs(t, err, runaimcp.ErrValidation)
			case "scope_empty_id", "malformed_entry":
				require.ErrorIs(t, err, runaimcp.ErrUpstream)
			case "renamed":
				require.ErrorIs(t, err, runaimcp.ErrConflict)
			case "confirmation_required":
				require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
			}
			assert.False(t, out.Applied)
			assert.Zero(t, writes.Load())
		})
	}
}

func TestDatasourceDelegatesPayloadValidationToAPI(t *testing.T) {
	cases := datasourceCases()
	for _, test := range []struct {
		name string
		tc   datasourceCase
		meta map[string]any
		spec map[string]any
	}{
		{"project_id", cases[0], map[string]any{"scope": "project", "projectId": 0}, map[string]any{"path": nil}},
		{"pvc_claim_conflict", cases[1], nil, map[string]any{"existingPvc": true, "claimInfo": map[string]any{"size": "1Gi"}}},
		{"config_map_mode", cases[5], nil, map[string]any{"defaultMode": "invalid"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := datasourceRequest(test.tc, true)
			maps.Copy(raw["meta"].(map[string]any), test.meta)
			raw["spec"] = test.spec
			in := manageDatasourceInput{Kind: test.tc.kind, Action: common.ActionCreate, Request: raw}
			require.NoError(t, in.Validate())
			client, mux := runaiclienttest.NewServer(t)
			var writes atomic.Int32
			mux.HandleFunc("/api/v1/asset/datasource/"+test.tc.endpoint, func(w http.ResponseWriter, r *http.Request) {
				writes.Add(1)
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				expected, err := json.Marshal(raw)
				assert.NoError(t, err)
				assert.JSONEq(t, string(expected), string(body))
				runaiclienttest.WriteJSON(w, http.StatusBadRequest, assets.Error{Code: http.StatusBadRequest, Message: "datasource rejected by API"})
			})
			out, err := manageDatasourceRun(context.Background(), client, in)
			require.ErrorContains(t, err, "datasource rejected by API")
			assert.False(t, out.Applied)
			assert.Equal(t, int32(1), writes.Load())
		})
	}
}

func TestDatasourceCreateByScopeName(t *testing.T) {
	for _, test := range []struct {
		name          string
		scope         assets.Scope
		path, entries string
		matches       []any
		idField       string
		id            any
		wantErr       error
	}{
		{"cluster", assets.SCOPE_CLUSTER, "/api/v1/clusters/minimal", "clusters", []any{map[string]any{"uuid": testDatasourceID}}, "clusterId", testDatasourceID, nil},
		{"project", assets.SCOPE_PROJECT, "/api/v1/org-unit/projects", "projects", []any{map[string]any{"id": "7"}}, "projectId", int32(7), nil},
		{"department", assets.SCOPE_DEPARTMENT, "/api/v1/org-unit/departments", "departments", []any{map[string]any{"id": "dept"}}, "departmentId", "dept", nil},
		{"missing_cluster", assets.SCOPE_CLUSTER, "/api/v1/clusters/minimal", "clusters", []any{}, "", nil, runaimcp.ErrNotFound},
		{"ambiguous_cluster", assets.SCOPE_CLUSTER, "/api/v1/clusters/minimal", "clusters", []any{map[string]any{"uuid": testDatasourceID}, map[string]any{"uuid": "22222222-2222-4222-8222-222222222222"}}, "", nil, runaimcp.ErrValidation},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(test.path, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "name==scope-name", r.URL.Query().Get("filterBy"))
				assert.Equal(t, "2", r.URL.Query().Get("limit"))
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{test.entries: test.matches})
			})
			tc := datasourceCases()[0]
			var writes atomic.Int32
			mux.HandleFunc("/api/v1/asset/datasource/nfs", func(w http.ResponseWriter, r *http.Request) {
				writes.Add(1)
				assert.Equal(t, http.MethodPost, r.Method)
				var request map[string]any
				if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&request)) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				meta := request["meta"].(map[string]any)
				assert.EqualValues(t, test.id, meta[test.idField])
				response := datasourceResponse(tc, 1)
				returned := response["meta"].(map[string]any)
				returned["scope"], returned[test.idField] = string(test.scope), test.id
				runaiclienttest.WriteJSON(w, http.StatusAccepted, response)
			})
			raw := datasourceRequest(tc, true)
			meta := raw["meta"].(map[string]any)
			meta["scope"] = string(test.scope)
			in := manageDatasourceInput{Action: common.ActionCreate, Kind: tc.kind, ScopeName: "scope-name", Request: raw}
			require.NoError(t, in.Validate())
			out, err := manageDatasourceRun(context.Background(), client, in)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				assert.Zero(t, writes.Load())
				assert.False(t, out.Applied)
			} else {
				require.NoError(t, err)
				assert.True(t, out.Applied)
				assert.Empty(t, out.Warnings)
				assert.Equal(t, int32(1), writes.Load())
				assert.NotContains(t, meta, test.idField, "scope resolution must not mutate the caller's request")
				meta[test.idField] = test.id
				require.ErrorContains(t, in.Validate(), "mutually exclusive")
			}
		})
	}
}
