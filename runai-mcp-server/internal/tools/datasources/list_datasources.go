/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"
	"fmt"
	"maps"
	"strings"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type listDatasourcesInput struct {
	Name                 string                `json:"name,omitempty" jsonschema:"Filter by datasource name."`
	Scope                assets.Scope          `json:"scope,omitempty" jsonschema:"Scope: system, tenant, cluster, department, or project. tenant includes system presets. Mutually exclusive with includeDescendants."`
	ClusterID            string                `json:"clusterId,omitempty" jsonschema:"Cluster UUID. Includes visible ancestor assets unless scope is set."`
	DepartmentID         string                `json:"departmentId,omitempty" jsonschema:"Department ID. Includes visible ancestor assets unless scope is set."`
	ProjectID            *int32                `json:"projectId,omitempty" jsonschema:"Positive project ID. Includes visible ancestor assets unless scope is set."`
	AssetIDs             string                `json:"assetIds,omitempty" jsonschema:"Filter by one UUID or a comma-separated list of UUIDs."`
	Type                 string                `json:"type,omitempty" jsonschema:"API datasource type: hostPath, nfs, pvc, git, s3, configMap, secretVolume, or dataVolume. dataVolume requires dataVolumes=true."`
	IncludeDescendants   bool                  `json:"includeDescendants,omitempty" jsonschema:"Include assets from descendant scopes. Mutually exclusive with scope."`
	UsageInfo            bool                  `json:"usageInfo,omitempty" jsonschema:"Include usage information in each datasource."`
	StatusInfo           bool                  `json:"statusInfo,omitempty" jsonschema:"Include cluster synchronization status; may include PVC assets pending deletion."`
	DataVolumes          bool                  `json:"dataVolumes,omitempty" jsonschema:"Include read-only data volumes. Requires projectId; excludes clusterId/departmentId and other types."`
	ComplyToProject      *int32                `json:"complyToProject,omitempty" jsonschema:"Positive project ID for compliance information. Requires complyToWorkloadType."`
	ComplyToWorkloadType assets.WorkloadType   `json:"complyToWorkloadType,omitempty" jsonschema:"Compliance workload type: Workspace, Training, Distributed, Inference, or DistributedInference. Requires complyToProject."`
	ComplyToReplicaType  workloads.ReplicaType `json:"complyToReplicaType,omitempty" jsonschema:"Compliance replica type: Worker, Master, or Leader. Required only for distributed workload compliance."`
	Limit                int32                 `json:"limit,omitempty" jsonschema:"Maximum results per page. Defaults to 50, capped at 200."`
	Offset               int32                 `json:"offset,omitempty" jsonschema:"Pagination offset. Use nextOffset from the previous response."`
}

type datasourceListItem struct {
	Kind                assets.AssetKind `json:"kind" jsonschema:"API asset kind to pass unchanged to manage_datasource."`
	DatasourceID        string           `json:"datasourceId"                   jsonschema:"Datasource asset UUID to pass to manage_datasource."`
	Name                string           `json:"name"                            jsonschema:"Datasource name."`
	Scope               assets.Scope     `json:"scope"                           jsonschema:"Datasource scope."`
	Datasource          map[string]any   `json:"datasource"                     jsonschema:"Complete datasource returned by the Assets API, including its type-specific spec wrapper."`
	UpdateRequest       map[string]any   `json:"updateRequest,omitempty"         jsonschema:"Complete kind-specific update request accepted by manage_datasource update. Modify only intended fields and keep all others."` //nolint:lll
	ExpectedUpdateCount *int64           `json:"expectedUpdateCount,omitempty"   jsonschema:"API update count to pass unchanged to manage_datasource update for a best-effort stale-write check."`                          //nolint:lll
	UpdateBlockedReason string           `json:"updateBlockedReason,omitempty"   jsonschema:"Why this datasource cannot be updated. Present only when updateRequest is unavailable."`
}

type listDatasourcesOutput struct {
	Datasources []datasourceListItem `json:"datasources"          jsonschema:"Matching datasources with complete specs and update preparation status."`
	TotalCount  int32                `json:"totalCount"            jsonschema:"Total number of datasources matching the filters across all pages."`
	HasMore     bool                 `json:"hasMore"               jsonschema:"True when more datasources exist beyond the returned page; pass nextOffset back as offset."`
	NextOffset  int32                `json:"nextOffset,omitempty"  jsonschema:"Offset to pass on the next call. Only set when hasMore is true."`
}

func registerListDatasources(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listDatasourcesInput, listDatasourcesOutput]{
		Name: "list_datasources",
		Description: "List datasource assets using Assets API scope, ID, name, type, compliance, usage, and status options with pagination. " +
			"Results include complete specs and, when supported, updateRequest and expectedUpdateCount for manage_datasource; otherwise updateBlockedReason explains why updates are unavailable.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listDatasourcesInput) (listDatasourcesOutput, error) {
			return listDatasourcesRun(ctx, client, in)
		},
	})
}

func listDatasourcesRun(ctx context.Context, client *runaiclient.Client, in listDatasourcesInput) (listDatasourcesOutput, error) {
	entries, err := fetchDatasources(ctx, client, in)
	if err != nil {
		return listDatasourcesOutput{}, err
	}
	page, hasMore, nextOffset := common.Paginate(entries, in.Offset, common.EffectiveLimit(in.Limit))
	items := make([]datasourceListItem, 0, len(page))
	for index, raw := range page {
		item, err := datasourceListItemFromRaw(raw.(map[string]any))
		if err != nil {
			return listDatasourcesOutput{}, fmt.Errorf("list datasources: entry %d: %w", int(in.Offset)+index, err)
		}
		items = append(items, item)
	}
	return listDatasourcesOutput{Datasources: items, TotalCount: int32(len(entries)), HasMore: hasMore, NextOffset: nextOffset}, nil
}

func fetchDatasources(ctx context.Context, client *runaiclient.Client, in listDatasourcesInput) ([]any, error) {
	req := datasourceListRequest(ctx, client, in)
	resp, err := assetResponse(req.Execute())
	if err != nil {
		return nil, fmt.Errorf("list datasources: %w", runaimcp.MapHTTPError(resp, err))
	}
	raw, err := readDatasourceResponse(resp, "list datasources")
	if err != nil {
		return nil, err
	}
	values, ok := raw["entries"].([]any)
	if !ok {
		return nil, fmt.Errorf("list datasources: entries must be an array: %w", runaimcp.ErrUpstream)
	}
	for index, value := range values {
		entry, ok := value.(map[string]any)
		if !ok || entry == nil {
			return nil, fmt.Errorf("list datasources: entry %d must be an object: %w", index, runaimcp.ErrUpstream)
		}
	}
	return values, nil
}

func datasourceListRequest(ctx context.Context, client *runaiclient.Client, in listDatasourcesInput) assets.ApiListDatasourceAssetsRequest {
	req := client.Assets.DatasourcesAPI.ListDatasourceAssets(ctx)
	if in.Name != "" {
		req = req.Name(in.Name)
	}
	if in.Scope != "" {
		req = req.Scope(string(in.Scope))
	}
	if in.ClusterID != "" {
		req = req.ClusterId(in.ClusterID)
	}
	if in.DepartmentID != "" {
		req = req.DepartmentId(in.DepartmentID)
	}
	if in.ProjectID != nil {
		req = req.ProjectId(*in.ProjectID)
	}
	if in.AssetIDs != "" {
		req = req.AssetIds(in.AssetIDs)
	}
	if in.Type != "" {
		req = req.Type_(in.Type)
	}
	if in.ComplyToProject != nil {
		req = req.ComplyToProject(*in.ComplyToProject)
	}
	if in.ComplyToWorkloadType != "" {
		req = req.ComplyToWorkloadType(string(in.ComplyToWorkloadType))
	}
	if in.ComplyToReplicaType != "" {
		req = req.ComplyToReplicaType(string(in.ComplyToReplicaType))
	}
	if in.IncludeDescendants {
		req = req.IncludeDescendants(true)
	}
	if in.UsageInfo {
		req = req.UsageInfo(true)
	}
	if in.StatusInfo {
		req = req.StatusInfo(true)
	}
	if in.DataVolumes {
		req = req.DataVolumes(true)
	}
	return req
}

func datasourceListItemFromRaw(raw map[string]any) (datasourceListItem, error) {
	meta, _ := raw["meta"].(map[string]any)
	kindValue, _ := meta["kind"].(string)
	kind := assets.AssetKind(kindValue)
	spec, ok := raw["spec"].(map[string]any)
	if !ok {
		return datasourceListItem{}, fmt.Errorf("datasource spec must be an object: %w", runaimcp.ErrUpstream)
	}

	normalized := maps.Clone(raw)
	specKey := datasourceSpecKey(kind)
	if specKey != "" {
		normalized["spec"] = spec[specKey]
	}
	snapshot, err := newDatasourceSnapshot(normalized, kind)
	if err != nil {
		return datasourceListItem{}, err
	}
	snapshot = prepareDatasourceUpdate(snapshot)
	item := datasourceListItem{
		Kind: snapshot.kind, DatasourceID: snapshot.id, Name: snapshot.name,
		Scope: snapshot.scope, Datasource: raw, UpdateBlockedReason: snapshot.updateBlockedReason,
	}
	if item.UpdateBlockedReason != "" {
		return item, nil
	}
	for field, value := range spec {
		if field != specKey && value != nil {
			item.UpdateBlockedReason = "the datasource specification contains unsupported fields"
			return item, nil
		}
	}
	item.UpdateRequest = snapshot.updateRequest
	item.ExpectedUpdateCount = snapshot.updateCount
	return item, nil
}

func (in listDatasourcesInput) Validate() error {
	if err := in.validateFilters(); err != nil {
		return err
	}
	if err := in.validateCompliance(); err != nil {
		return err
	}
	return common.ValidatePagination(in.Limit, in.Offset)
}

func (in listDatasourcesInput) validateFilters() error {
	if in.AssetIDs != "" {
		for _, id := range strings.Split(in.AssetIDs, ",") {
			if err := common.ValidateUUID("assetIds", id); err != nil {
				return err
			}
		}
	}
	if in.Name != "" {
		if _, err := common.RequiredString("name", in.Name); err != nil {
			return err
		}
	}
	if in.Scope != "" && !in.Scope.IsValid() {
		return fmt.Errorf("scope must be one of: system, tenant, cluster, department, project")
	}
	if in.DepartmentID != "" {
		if _, err := common.RequiredString("departmentId", in.DepartmentID); err != nil {
			return err
		}
	}
	if in.ProjectID != nil && *in.ProjectID <= 0 {
		return fmt.Errorf("projectId must be positive")
	}
	if in.Scope != "" && in.IncludeDescendants {
		return fmt.Errorf("includeDescendants and scope are mutually exclusive")
	}
	if in.Type != "" {
		if _, ok := datasourceTypes[in.Type]; !ok {
			return fmt.Errorf("type must be one of: hostPath, nfs, pvc, git, s3, configMap, secretVolume, dataVolume")
		}
	}
	return in.validateDataVolumes()
}

func (in listDatasourcesInput) validateDataVolumes() error {
	if in.DataVolumes && in.Type != "" && in.Type != dataVolumeType {
		return fmt.Errorf("dataVolumes=true only supports an omitted type or type=dataVolume")
	}
	if in.Type == dataVolumeType && !in.DataVolumes {
		return fmt.Errorf("type=dataVolume requires dataVolumes=true")
	}
	if in.DataVolumes {
		if in.ProjectID == nil {
			return fmt.Errorf("dataVolumes=true requires projectId")
		}
		if in.ClusterID != "" || in.DepartmentID != "" {
			return fmt.Errorf("dataVolumes=true does not support clusterId or departmentId")
		}
	}
	return nil
}

func (in listDatasourcesInput) validateCompliance() error {
	if (in.ComplyToProject != nil) != (in.ComplyToWorkloadType != "") {
		return fmt.Errorf("complyToProject and complyToWorkloadType must be provided together")
	}
	if in.ComplyToProject != nil && *in.ComplyToProject <= 0 {
		return fmt.Errorf("complyToProject must be positive")
	}
	workload := in.ComplyToWorkloadType
	if workload != "" && !workload.IsValid() {
		return fmt.Errorf("invalid complyToWorkloadType")
	}
	distributed := workload == assets.WORKLOADTYPE_DISTRIBUTED || workload == assets.WORKLOADTYPE_DISTRIBUTED_INFERENCE
	if distributed && in.ComplyToReplicaType == "" {
		return fmt.Errorf("complyToReplicaType is required for distributed workload compliance")
	}
	if !distributed && in.ComplyToReplicaType != "" {
		return fmt.Errorf("complyToReplicaType is only supported for distributed workload compliance")
	}
	if in.ComplyToReplicaType != "" && !in.ComplyToReplicaType.IsValid() {
		return fmt.Errorf("complyToReplicaType must be one of: Worker, Master, Leader")
	}
	return nil
}
