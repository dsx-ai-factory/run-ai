/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatasourceListBlocksUnsafeUpdates(t *testing.T) {
	for _, reason := range []string{"unknown_wrapper", "unknown_meta", "unknown_spec", "unknown_root", "system_scope", "missing_update_count"} {
		t.Run(reason, func(t *testing.T) {
			tc := datasourceCases()[0]
			raw := datasourceResponse(tc, 2)
			raw["spec"] = map[string]any{"nfs": tc.spec}
			meta := raw["meta"].(map[string]any)
			switch reason {
			case "unknown_wrapper":
				raw["spec"].(map[string]any)["futureSpec"] = map[string]any{"value": true}
			case "unknown_meta":
				meta["futureField"] = true
			case "unknown_spec":
				tc.spec["futureField"] = true
			case "unknown_root":
				raw["futureField"] = true
			case "system_scope":
				meta["scope"] = string(assets.SCOPE_SYSTEM)
			case "missing_update_count":
				delete(meta, "updateCount")
			}
			item, err := datasourceListItemFromRaw(raw)
			require.NoError(t, err)
			assert.Nil(t, item.UpdateRequest)
			assert.Nil(t, item.ExpectedUpdateCount)
			assert.NotEmpty(t, item.UpdateBlockedReason)
			assert.Equal(t, raw, item.Datasource)
		})
	}
}

func TestDatasourceListAllowsNullableSpecs(t *testing.T) {
	cases := datasourceCases()
	for _, tc := range []datasourceCase{cases[0], cases[5], cases[6]} {
		t.Run(string(tc.kind), func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			raw := datasourceResponse(tc, 2)
			raw["spec"] = map[string]any{tc.specKey: nil}
			mux.HandleFunc("/api/v1/asset/datasource", func(w http.ResponseWriter, r *http.Request) {
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{raw}})
			})
			out, err := listDatasourcesRun(context.Background(), client, listDatasourcesInput{})
			require.NoError(t, err)
			require.Len(t, out.Datasources, 1)
			item := out.Datasources[0]
			assert.Equal(t, raw["spec"], item.Datasource["spec"])
			if tc.kind == assets.ASSETKIND_CONFIG_MAP || tc.kind == assets.ASSETKIND_SECRET_VOLUME {
				require.Empty(t, item.UpdateBlockedReason)
				assert.NotContains(t, item.UpdateRequest, "spec")
				require.NoError(t, (manageDatasourceInput{Action: common.ActionUpdate, Kind: tc.kind, DatasourceID: item.DatasourceID, Request: item.UpdateRequest, ExpectedUpdateCount: item.ExpectedUpdateCount}).Validate())
			} else {
				assert.Nil(t, item.UpdateRequest)
				assert.NotEmpty(t, item.UpdateBlockedReason)
			}
		})
	}
}

func TestListDatasourcesPreparesEveryKind(t *testing.T) {
	for _, tc := range datasourceCases() {
		t.Run(string(tc.kind), func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			raw := datasourceResponse(tc, 2)
			raw["spec"] = map[string]any{tc.specKey: tc.spec}
			mux.HandleFunc("/api/v1/asset/datasource", func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, http.MethodGet, req.Method)
				assert.Equal(t, tc.queryType, req.URL.Query().Get("type"))
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{raw}})
			})
			out, err := listDatasourcesRun(context.Background(), client, listDatasourcesInput{Type: tc.queryType})
			require.NoError(t, err)
			require.Len(t, out.Datasources, 1)
			item := out.Datasources[0]
			assert.Equal(t, tc.kind, item.Kind)
			assert.Equal(t, testDatasourceID, item.DatasourceID)
			require.Empty(t, item.UpdateBlockedReason)
			assert.Equal(t, datasourceRequest(tc, false)["spec"], item.UpdateRequest["spec"])
			in := manageDatasourceInput{Action: common.ActionUpdate, Kind: item.Kind, DatasourceID: item.DatasourceID, Request: item.UpdateRequest, ExpectedUpdateCount: item.ExpectedUpdateCount}
			require.NoError(t, in.Validate())
		})
	}
}

func TestListDatasourcePagination(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	entries := make([]any, 0)
	for _, tc := range datasourceCases() {
		raw := datasourceResponse(tc, 2)
		raw["spec"] = map[string]any{tc.specKey: tc.spec}
		entries = append(entries, raw)
	}
	mux.HandleFunc("/api/v1/asset/datasource", func(w http.ResponseWriter, req *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": entries})
	})
	for _, test := range []struct {
		name   string
		offset int32
		kinds  []assets.AssetKind
		next   int32
	}{
		{"first_page", 0, []assets.AssetKind{assets.ASSETKIND_NFS, assets.ASSETKIND_PVC}, 2},
		{"middle_page", 2, []assets.AssetKind{assets.ASSETKIND_S3, assets.ASSETKIND_GIT}, 4},
		{"last_page", 6, []assets.AssetKind{assets.ASSETKIND_SECRET_VOLUME}, 0},
		{"past_end", 7, []assets.AssetKind{}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			out, err := listDatasourcesRun(context.Background(), client, listDatasourcesInput{Limit: 2, Offset: test.offset})
			require.NoError(t, err)
			require.NotNil(t, out.Datasources)
			kinds := make([]assets.AssetKind, 0, len(out.Datasources))
			for _, item := range out.Datasources {
				kinds = append(kinds, item.Kind)
			}
			assert.Equal(t, test.kinds, kinds)
			assert.Equal(t, int32(len(entries)), out.TotalCount)
			assert.Equal(t, test.next, out.NextOffset)
			assert.Equal(t, test.next != 0, out.HasMore)
		})
	}
}

func TestListDatasourceValidation(t *testing.T) {
	project, zero := int32(7), int32(0)
	for name, in := range map[string]listDatasourcesInput{
		"unknown_type":                  {Type: "unknown"},
		"zero_project":                  {ProjectID: &zero},
		"invalid_multi_id":              {AssetIDs: testDatasourceID + ",invalid"},
		"scope_with_descendants":        {Scope: assets.SCOPE_TENANT, IncludeDescendants: true},
		"data_volume_without_flag":      {Type: "dataVolume"},
		"data_volume_type_conflict":     {DataVolumes: true, ProjectID: &project, Type: "nfs"},
		"data_volume_missing_project":   {DataVolumes: true},
		"data_volume_cluster_filter":    {DataVolumes: true, ProjectID: &project, ClusterID: testDatasourceID},
		"data_volume_department_filter": {DataVolumes: true, ProjectID: &project, DepartmentID: "dept"},
		"compliance_missing_workload":   {ComplyToProject: &project},
		"compliance_missing_project":    {ComplyToWorkloadType: assets.WORKLOADTYPE_WORKSPACE},
		"distributed_missing_replica":   {ComplyToProject: &project, ComplyToWorkloadType: assets.WORKLOADTYPE_DISTRIBUTED},
		"unexpected_replica":            {ComplyToProject: &project, ComplyToWorkloadType: assets.WORKLOADTYPE_WORKSPACE, ComplyToReplicaType: workloads.REPLICATYPE_WORKER},
	} {
		t.Run(name, func(t *testing.T) { require.Error(t, in.Validate()) })
	}
	for name, in := range map[string]listDatasourcesInput{
		"defaults":                            {},
		"system":                              {Scope: assets.SCOPE_SYSTEM},
		"data_volumes_with_other_datasources": {DataVolumes: true, ProjectID: &project},
		"workspace_compliance":                {ComplyToProject: &project, ComplyToWorkloadType: assets.WORKLOADTYPE_WORKSPACE},
	} {
		t.Run(name, func(t *testing.T) { require.NoError(t, in.Validate()) })
	}
}

func TestListDatasourcesForwardsAllAPIOptions(t *testing.T) {
	project := int32(7)
	tests := []struct {
		name  string
		in    listDatasourcesInput
		query url.Values
	}{
		{"filters", listDatasourcesInput{Name: "data", Scope: assets.SCOPE_PROJECT, ClusterID: testDatasourceID, DepartmentID: "dept", ProjectID: &project, AssetIDs: testDatasourceID + "," + testDatasourceID, Type: "nfs"}, url.Values{
			"name": {"data"}, "scope": {"project"}, "clusterId": {testDatasourceID}, "departmentId": {"dept"}, "projectId": {"7"}, "assetIds": {testDatasourceID + "," + testDatasourceID}, "type": {"nfs"},
		}},
		{"compliance_and_info", listDatasourcesInput{ComplyToProject: &project, ComplyToWorkloadType: assets.WORKLOADTYPE_DISTRIBUTED, ComplyToReplicaType: workloads.REPLICATYPE_WORKER, UsageInfo: true, StatusInfo: true, IncludeDescendants: true}, url.Values{
			"complyToProject": {"7"}, "complyToWorkloadType": {"Distributed"}, "complyToReplicaType": {"Worker"}, "usageInfo": {"true"}, "statusInfo": {"true"}, "includeDescendants": {"true"},
		}},
		{"data_volumes", listDatasourcesInput{Type: "dataVolume", DataVolumes: true, ProjectID: &project}, url.Values{"type": {"dataVolume"}, "dataVolumes": {"true"}, "projectId": {"7"}}},
		{"defaults", listDatasourcesInput{}, url.Values{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, test.in.Validate())
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/asset/datasource", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, test.query, r.URL.Query())
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"entries": []any{}})
			})
			out, err := listDatasourcesRun(context.Background(), client, test.in)
			require.NoError(t, err)
			assert.NotNil(t, out.Datasources)
			assert.Empty(t, out.Datasources)
		})
	}
}

func TestListDatasourcesPreservesAPIErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		want   error
	}{
		{"forbidden", http.StatusForbidden, runaimcp.ErrForbidden},
		{"unavailable", http.StatusServiceUnavailable, runaimcp.ErrUpstream},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc("/api/v1/asset/datasource", func(w http.ResponseWriter, r *http.Request) {
				runaiclienttest.WriteJSON(w, tc.status, assets.Error{Code: int32(tc.status), Message: "cannot load project datasources"})
			})
			_, err := listDatasourcesRun(context.Background(), client, listDatasourcesInput{})
			require.ErrorIs(t, err, tc.want)
			assert.ErrorContains(t, err, "cannot load project datasources")
		})
	}
}
