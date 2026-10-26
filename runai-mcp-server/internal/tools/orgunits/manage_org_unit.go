/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orgunits

import (
	"context"
	"fmt"
	"strings"

	orgunit "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type orgUnitQuota struct {
	NodePoolName          string   `json:"nodePoolName"                 jsonschema:"Node pool this quota line applies to, by name; resolved to its ID."`
	GPUDeserved           *float64 `json:"gpuDeserved,omitempty"        jsonschema:"Guaranteed GPU quota on this node pool. Required for each quota line on create."`
	GPULimit              *float64 `json:"gpuLimit,omitempty"           jsonschema:"Max GPU allocation (deserved + over-quota) on this node pool. Omit to leave at the upstream default."`
	GPUOverQuotaWeight    *float64 `json:"gpuOverQuotaWeight,omitempty" jsonschema:"Over-quota share weight within the rank (1-20)."`
	CPUDeserved           *float64 `json:"cpuDeserved,omitempty"        jsonschema:"Guaranteed CPU quota in cores on this node pool."`
	CPULimit              *float64 `json:"cpuLimit,omitempty"           jsonschema:"Max CPU allocation in cores on this node pool."`
	CPUOverQuotaWeight    *float64 `json:"cpuOverQuotaWeight,omitempty" jsonschema:"CPU over-quota share weight within the rank (1-20)."`
	MemoryDeserved        *float64 `json:"memoryDeserved,omitempty"     jsonschema:"Guaranteed CPU-memory quota in GiB on this node pool."`
	MemoryLimit           *float64 `json:"memoryLimit,omitempty"        jsonschema:"Max CPU-memory allocation in GiB on this node pool."`
	MemoryOverQuotaWeight *float64 `json:"memoryOverQuotaWeight,omitempty" jsonschema:"CPU-memory over-quota share weight within the rank (1-20)."`
	Rank                  string   `json:"rank,omitempty"               jsonschema:"Scheduling rank of this org-unit on the node pool (e.g. High, Medium, Low)."`
}

type orgUnitSchedulingRules struct {
	InteractiveJobTimeLimitSeconds       *int32 `json:"interactiveJobTimeLimitSeconds,omitempty"       jsonschema:"Max wall-clock runtime for interactive jobs, in seconds."`
	InteractiveJobMaxIdleDurationSeconds *int32 `json:"interactiveJobMaxIdleDurationSeconds,omitempty" jsonschema:"Max idle time before an interactive job is stopped, in seconds."`
	InteractiveJobPreemptIdleDuration    *int32 `json:"interactiveJobPreemptIdleDurationSeconds,omitempty" jsonschema:"Idle time before a preemptible interactive job is preempted, in seconds."`
	TrainingJobMaxIdleDurationSeconds    *int32 `json:"trainingJobMaxIdleDurationSeconds,omitempty"    jsonschema:"Max idle time before a training job is stopped, in seconds."`
	TrainingJobTimeLimitSeconds          *int32 `json:"trainingJobTimeLimitSeconds,omitempty"          jsonschema:"Max wall-clock runtime for training jobs, in seconds."`
}

type orgUnitSpec struct {
	Description           *string                 `json:"description,omitempty"           jsonschema:"Free-text description of the org-unit."`
	ParentID              *string                 `json:"parentId,omitempty"              jsonschema:"Project-only, create-only: ID of the parent department. Cannot be changed via update."`
	RequestedNamespace    *string                 `json:"requestedNamespace,omitempty"    jsonschema:"Project-only, create-only: custom Kubernetes namespace. Defaults to runai-<name>."`
	EnforceRunaiScheduler *bool                   `json:"enforceRunaiScheduler,omitempty" jsonschema:"Project-only: require all pods in the namespace to use the Run:ai scheduler (cluster v2.21+)."`
	DefaultNodePools      []string                `json:"defaultNodePools,omitempty"      jsonschema:"Ordered node pool names defining the default scheduling preference for workloads."`
	SchedulingRules       *orgUnitSchedulingRules `json:"schedulingRules,omitempty"       jsonschema:"Job time and idle limits applied to workloads in this org-unit."`
	Quota                 []orgUnitQuota          `json:"quota,omitempty"                 jsonschema:"Per-node-pool quota lines. On update they merge by node pool; pools you omit keep their current quota."`
}

type manageOrgUnitInput struct {
	Action        string       `json:"action"                  jsonschema:"What to do. One of: create, update, delete."`
	OrgType       string       `json:"orgType"                 jsonschema:"Org-unit kind. One of: project, department."`
	OrgUnitID     string       `json:"orgUnitId,omitempty"     jsonschema:"Org-unit ID. Preferred for update/delete. Mutually exclusive with name."`
	Name          string       `json:"name,omitempty"          jsonschema:"Org-unit name. Required for create; resolved to an ID otherwise. Mutually exclusive with orgUnitId."`
	ClusterID     string       `json:"clusterId,omitempty"     jsonschema:"Cluster the org-unit lives in (UUID). Required for create; also disambiguates name resolution."`
	Spec          *orgUnitSpec `json:"spec,omitempty"          jsonschema:"Desired-state spec. Required for create. On update only the fields you set change; omitted ones are preserved."`
	DryRun        bool         `json:"dryRun,omitempty"        jsonschema:"When true, preview the change (before/after/diff) without mutating anything upstream."`
	IgnoreMissing bool         `json:"ignoreMissing,omitempty" jsonschema:"Delete-only. When true, deleting an org-unit that no longer exists is a successful no-op instead of not_found."`
	Confirm       bool         `json:"confirm,omitempty"       jsonschema:"Required for a non-dry-run update or delete when the client can't prompt; otherwise returns confirmation_required with the snapshot."`
}

func (in manageOrgUnitInput) Validate() error {
	switch common.WriteAction(in.Action) {
	case common.ActionCreate, common.ActionUpdate, common.ActionDelete:
	default:
		return fmt.Errorf("action must be one of: create, update, delete")
	}
	if in.OrgType != common.OrgTypeProject && in.OrgType != common.OrgTypeDepartment {
		return fmt.Errorf("orgType must be one of: %s, %s", common.OrgTypeProject, common.OrgTypeDepartment)
	}
	if in.OrgUnitID != "" && in.Name != "" {
		return fmt.Errorf("orgUnitId and name are mutually exclusive")
	}
	for field, v := range map[string]string{"name": in.Name, "orgUnitId": in.OrgUnitID, "clusterId": in.ClusterID} {
		if err := common.ValidateNoCommas(field, v); err != nil {
			return err
		}
	}

	switch common.WriteAction(in.Action) {
	case common.ActionCreate:
		if in.Name == "" {
			return fmt.Errorf("name is required for create")
		}
		if in.OrgUnitID != "" {
			return fmt.Errorf("orgUnitId must not be set for create")
		}
		if in.ClusterID == "" {
			return fmt.Errorf("clusterId is required for create")
		}
		if in.Spec == nil {
			return fmt.Errorf("spec is required for create")
		}
	case common.ActionUpdate:
		if in.OrgUnitID == "" && in.Name == "" {
			return fmt.Errorf("one of orgUnitId or name is required for update")
		}
		if in.Spec == nil {
			return fmt.Errorf("spec is required for update")
		}
	case common.ActionDelete:
		if in.OrgUnitID == "" && in.Name == "" {
			return fmt.Errorf("one of orgUnitId or name is required for delete")
		}
		if in.Spec != nil {
			return fmt.Errorf("spec must not be set for delete")
		}
	}

	if in.IgnoreMissing && common.WriteAction(in.Action) != common.ActionDelete {
		return fmt.Errorf("ignoreMissing is only valid with action=delete")
	}
	return in.validateSpec()
}

func (in manageOrgUnitInput) validateSpec() error {
	if in.Spec == nil {
		return nil
	}
	if in.OrgType == common.OrgTypeDepartment {
		if in.Spec.ParentID != nil {
			return fmt.Errorf("spec.parentId is only valid for orgType=project")
		}
		if in.Spec.RequestedNamespace != nil {
			return fmt.Errorf("spec.requestedNamespace is only valid for orgType=project")
		}
		if in.Spec.EnforceRunaiScheduler != nil {
			return fmt.Errorf("spec.enforceRunaiScheduler is only valid for orgType=project")
		}
	}
	if in.Spec.ParentID != nil && common.WriteAction(in.Action) == common.ActionUpdate {
		return fmt.Errorf("spec.parentId cannot be changed via update; it is create-only")
	}
	if in.Spec.RequestedNamespace != nil && common.WriteAction(in.Action) == common.ActionUpdate {
		return fmt.Errorf("spec.requestedNamespace cannot be changed via update; it is create-only")
	}
	for i, q := range in.Spec.Quota {
		if strings.TrimSpace(q.NodePoolName) == "" {
			return fmt.Errorf("spec.quota[%d].nodePoolName is required", i)
		}
		if err := common.ValidateFilterValue(fmt.Sprintf("spec.quota[%d].nodePoolName", i), q.NodePoolName); err != nil {
			return err
		}
		if common.WriteAction(in.Action) == common.ActionCreate && q.GPUDeserved == nil {
			return fmt.Errorf("spec.quota[%d].gpuDeserved is required on create", i)
		}
		if err := validateQuotaNumbers(i, q); err != nil {
			return err
		}
	}
	return validateSchedulingRules(in.Spec.SchedulingRules)
}

// validateQuotaNumbers rejects negative deserved/limit values and over-quota weights
// outside the documented 1-20 range before they reach the upstream API.
func validateQuotaNumbers(i int, q orgUnitQuota) error {
	for name, v := range map[string]*float64{
		"gpuDeserved":    q.GPUDeserved,
		"gpuLimit":       q.GPULimit,
		"cpuDeserved":    q.CPUDeserved,
		"cpuLimit":       q.CPULimit,
		"memoryDeserved": q.MemoryDeserved,
		"memoryLimit":    q.MemoryLimit,
	} {
		if v != nil && *v < 0 {
			return fmt.Errorf("spec.quota[%d].%s must not be negative", i, name)
		}
	}
	for name, w := range map[string]*float64{
		"gpuOverQuotaWeight":    q.GPUOverQuotaWeight,
		"cpuOverQuotaWeight":    q.CPUOverQuotaWeight,
		"memoryOverQuotaWeight": q.MemoryOverQuotaWeight,
	} {
		if w != nil && (*w < 1 || *w > 20) {
			return fmt.Errorf("spec.quota[%d].%s must be between 1 and 20", i, name)
		}
	}
	return nil
}

func validateSchedulingRules(sr *orgUnitSchedulingRules) error {
	if sr == nil {
		return nil
	}
	for field, v := range map[string]*int32{
		"interactiveJobTimeLimitSeconds":           sr.InteractiveJobTimeLimitSeconds,
		"interactiveJobMaxIdleDurationSeconds":     sr.InteractiveJobMaxIdleDurationSeconds,
		"interactiveJobPreemptIdleDurationSeconds": sr.InteractiveJobPreemptIdleDuration,
		"trainingJobMaxIdleDurationSeconds":        sr.TrainingJobMaxIdleDurationSeconds,
		"trainingJobTimeLimitSeconds":              sr.TrainingJobTimeLimitSeconds,
	} {
		if v != nil && *v < 0 {
			return fmt.Errorf("spec.schedulingRules.%s must not be negative", field)
		}
	}
	return nil
}

type manageOrgUnitOutput struct {
	Action    string         `json:"action"              jsonschema:"The action performed: create, update, or delete."`
	OrgType   string         `json:"orgType"             jsonschema:"The org-unit kind acted on: project or department."`
	DryRun    bool           `json:"dryRun"              jsonschema:"True when this was a preview; no upstream change was made."`
	Applied   bool           `json:"applied"             jsonschema:"True when the change was actually applied upstream. False on dry-run or an ignored-missing delete."`
	OrgUnitID string         `json:"orgUnitId,omitempty" jsonschema:"Resolved org-unit ID. Empty on a create dry-run."`
	Name      string         `json:"name,omitempty"      jsonschema:"Org-unit name."`
	ClusterID string         `json:"clusterId,omitempty" jsonschema:"Cluster the org-unit belongs to."`
	Before    map[string]any `json:"before"              jsonschema:"Org-unit snapshot before the change. Empty object on create."`
	After     map[string]any `json:"after"               jsonschema:"Org-unit snapshot after the change. Empty object on delete."`
	Diff      map[string]any `json:"diff"                jsonschema:"Field-level changes as path -> {before, after}."`
	Warnings  []string       `json:"warnings,omitempty"  jsonschema:"Non-fatal notes, e.g. when a delete target was already missing."`
}

func registerManageOrgUnit(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageOrgUnitInput, manageOrgUnitOutput]{
		Name: "manage_org_unit",
		Description: "Create, update, or delete a Run:ai project or department — its description, quota per node pool, " +
			"default node pools, parent department (projects), and scheduler enforcement. The orgType field selects " +
			"project or department; the action field selects the verb (create | update | delete). For high-impact " +
			"changes, call once with dryRun=true to preview the diff, then again with dryRun=false to apply. A " +
			"non-dry-run delete requires confirm=true; deleting a project also permanently deletes all workloads in it, " +
			"so surface that before confirming. Do NOT call to list org-units or read quota usage (use the " +
			"list_project_resources / list_department_resources / get_org_unit_metrics tools), or to manage node pools " +
			"(use manage_node_pool).",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageOrgUnitInput) (manageOrgUnitOutput, error) {
			return manageOrgUnitRun(ctx, client, in)
		},
	})
}

func manageOrgUnitRun(ctx context.Context, client *runaiclient.Client, in manageOrgUnitInput) (manageOrgUnitOutput, error) {
	action := common.WriteAction(in.Action)
	out := manageOrgUnitOutput{
		Action:    in.Action,
		OrgType:   in.OrgType,
		DryRun:    in.DryRun,
		ClusterID: in.ClusterID,
		Name:      in.Name,
		Before:    map[string]any{},
		After:     map[string]any{},
		Diff:      map[string]any{},
	}

	switch action {
	case common.ActionCreate:
		parentDeptID := ""
		if in.OrgType == common.OrgTypeProject && in.Spec != nil && in.Spec.ParentID != nil {
			parentDeptID = *in.Spec.ParentID
		}
		if err := authorizeOrgUnit(ctx, client, in.OrgType, action, in.ClusterID, "", parentDeptID); err != nil {
			return manageOrgUnitOutput{}, err
		}
		return manageOrgUnitCreate(ctx, client, in, out)
	case common.ActionUpdate, common.ActionDelete:
		id := in.OrgUnitID
		if id == "" {
			resolved, err := common.ResolveOrgUnitID(ctx, client, in.OrgType, in.Name, in.ClusterID)
			if err != nil {
				// A delete-by-name of an already-gone org-unit fails here; honor ignoreMissing.
				if action == common.ActionDelete && in.IgnoreMissing && common.IsNotFound(err) {
					out.Warnings = append(out.Warnings, fmt.Sprintf("%s %q not found; nothing to delete", in.OrgType, in.Name))
					return out, nil
				}
				return manageOrgUnitOutput{}, err
			}
			id = resolved
		}
		out.OrgUnitID = id

		// Update/delete authorize after the fetch (inside these helpers) so the RBAC probe
		// can scope on the org-unit's actual clusterId and parent, which a by-id call doesn't carry.
		if action == common.ActionUpdate {
			return manageOrgUnitUpdate(ctx, client, in, out, id)
		}
		return manageOrgUnitDelete(ctx, client, in, out, id)
	default:
		return manageOrgUnitOutput{}, fmt.Errorf("unsupported action %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

func manageOrgUnitCreate(ctx context.Context, client *runaiclient.Client, in manageOrgUnitInput, out manageOrgUnitOutput) (manageOrgUnitOutput, error) {
	resources, err := buildResources(ctx, client, in.ClusterID, in.Spec.Quota, nil)
	if err != nil {
		return manageOrgUnitOutput{}, err
	}

	if in.DryRun {
		out.After = createPreview(in, resources)
		out.Diff = common.DiffMaps(nil, out.After)
		return out, nil
	}

	snap, newID, err := doCreate(ctx, client, in, resources)
	if err != nil {
		return manageOrgUnitOutput{}, err
	}
	out.After = snap
	out.OrgUnitID = newID
	out.Applied = true
	out.Diff = common.DiffMaps(nil, out.After)
	return out, nil
}

func doCreate(ctx context.Context, client *runaiclient.Client, in manageOrgUnitInput, resources []orgunit.ResourcesNullable) (map[string]any, string, error) {
	if in.OrgType == common.OrgTypeProject {
		req := orgunit.NewProjectCreationRequest(in.Name, in.ClusterID, resources)
		applyProjectCreateSpec(req, in.Spec)
		created, httpResp, err := client.OrgUnit.ProjectsAPI.CreateProject(ctx).ProjectCreationRequest(*req).Execute()
		if err != nil {
			return nil, "", fmt.Errorf("create project %q: %w", in.Name, runaimcp.MapHTTPError(httpResp, err))
		}
		if created == nil {
			return nil, "", fmt.Errorf("create project %q: nil response: %w", in.Name, runaimcp.ErrUpstream)
		}
		return projectSnapshot(created), created.Id, nil
	}
	req := orgunit.NewDepartmentCreationRequest(in.Name, in.ClusterID, resources)
	applyDepartmentCreateSpec(req, in.Spec)
	created, httpResp, err := client.OrgUnit.DepartmentsAPI.CreateDepartment(ctx).DepartmentCreationRequest(*req).Execute()
	if err != nil {
		return nil, "", fmt.Errorf("create department %q: %w", in.Name, runaimcp.MapHTTPError(httpResp, err))
	}
	if created == nil {
		return nil, "", fmt.Errorf("create department %q: nil response: %w", in.Name, runaimcp.ErrUpstream)
	}
	return departmentSnapshot(created), created.Id, nil
}

func manageOrgUnitUpdate(ctx context.Context, client *runaiclient.Client, in manageOrgUnitInput, out manageOrgUnitOutput, id string) (manageOrgUnitOutput, error) {
	if in.OrgType == common.OrgTypeProject {
		return updateProject(ctx, client, in, out, id)
	}
	return updateDepartment(ctx, client, in, out, id)
}

func updateProject(ctx context.Context, client *runaiclient.Client, in manageOrgUnitInput, out manageOrgUnitOutput, id string) (manageOrgUnitOutput, error) {
	current, err := getProject(ctx, client, id)
	if err != nil {
		return manageOrgUnitOutput{}, err
	}
	if err := authorizeOrgUnit(ctx, client, in.OrgType, common.ActionUpdate, current.ClusterId, id, current.ParentId); err != nil {
		return manageOrgUnitOutput{}, err
	}
	out.Before = projectSnapshot(current)
	out.Name = current.Name
	out.ClusterID = current.ClusterId

	resources, err := buildResources(ctx, client, current.ClusterId, in.Spec.Quota, current.Resources)
	if err != nil {
		return manageOrgUnitOutput{}, err
	}

	req := orgunit.NewProjectUpdateRequest(resources)
	applyProjectUpdateSpec(req, in.Spec, current)

	if in.DryRun {
		out.After = updatePreview(out.Before, in.Spec, resources)
		out.Diff = common.DiffMaps(out.Before, out.After)
		return out, nil
	}

	if err := common.ConfirmWrite(ctx, in.Confirm, "update", in.OrgType, current.Name, id,
		common.DiffMaps(out.Before, updatePreview(out.Before, in.Spec, resources))); err != nil {
		return manageOrgUnitOutput{}, err
	}

	updated, httpResp, err := client.OrgUnit.ProjectsAPI.UpdateProject(ctx, id).ProjectUpdateRequest(*req).Execute()
	if err != nil {
		return manageOrgUnitOutput{}, fmt.Errorf("update project %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	if updated == nil {
		return manageOrgUnitOutput{}, fmt.Errorf("update project %q: nil response: %w", id, runaimcp.ErrUpstream)
	}
	out.After = projectSnapshot(updated)
	out.Applied = true
	out.Diff = common.DiffMaps(out.Before, out.After)
	return out, nil
}

func updateDepartment(ctx context.Context, client *runaiclient.Client, in manageOrgUnitInput, out manageOrgUnitOutput, id string) (manageOrgUnitOutput, error) {
	current, err := getDepartment(ctx, client, id)
	if err != nil {
		return manageOrgUnitOutput{}, err
	}
	if err := authorizeOrgUnit(ctx, client, in.OrgType, common.ActionUpdate, current.ClusterId, id, ""); err != nil {
		return manageOrgUnitOutput{}, err
	}
	out.Before = departmentSnapshot(current)
	out.Name = current.Name
	out.ClusterID = current.ClusterId

	resources, err := buildResources(ctx, client, current.ClusterId, in.Spec.Quota, current.Resources)
	if err != nil {
		return manageOrgUnitOutput{}, err
	}

	req := orgunit.NewDepartmentUpdateRequest(resources)
	applyDepartmentUpdateSpec(req, in.Spec, current)

	if in.DryRun {
		out.After = updatePreview(out.Before, in.Spec, resources)
		out.Diff = common.DiffMaps(out.Before, out.After)
		return out, nil
	}

	if err := common.ConfirmWrite(ctx, in.Confirm, "update", in.OrgType, current.Name, id,
		common.DiffMaps(out.Before, updatePreview(out.Before, in.Spec, resources))); err != nil {
		return manageOrgUnitOutput{}, err
	}

	updated, httpResp, err := client.OrgUnit.DepartmentsAPI.UpdateDepartment(ctx, id).DepartmentUpdateRequest(*req).Execute()
	if err != nil {
		return manageOrgUnitOutput{}, fmt.Errorf("update department %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	if updated == nil {
		return manageOrgUnitOutput{}, fmt.Errorf("update department %q: nil response: %w", id, runaimcp.ErrUpstream)
	}
	out.After = departmentSnapshot(updated)
	out.Applied = true
	out.Diff = common.DiffMaps(out.Before, out.After)
	return out, nil
}

func manageOrgUnitDelete(ctx context.Context, client *runaiclient.Client, in manageOrgUnitInput, out manageOrgUnitOutput, id string) (manageOrgUnitOutput, error) {
	snap, childCount, name, clusterID, parentDeptID, err := loadForDelete(ctx, client, in.OrgType, id)
	if err != nil {
		if in.IgnoreMissing && common.IsNotFound(err) {
			out.Warnings = append(out.Warnings, fmt.Sprintf("%s not found; nothing to delete", in.OrgType))
			return out, nil
		}
		return manageOrgUnitOutput{}, err
	}
	if err := authorizeOrgUnit(ctx, client, in.OrgType, common.ActionDelete, clusterID, id, parentDeptID); err != nil {
		return manageOrgUnitOutput{}, err
	}
	out.Before = snap
	out.Name = name
	out.ClusterID = clusterID
	out.Diff = common.DiffMaps(out.Before, nil)

	// A department with subordinate projects can't be deleted; report it as a conflict
	// up front rather than relaying an opaque upstream 4xx.
	if in.OrgType == common.OrgTypeDepartment && childCount > 0 {
		return manageOrgUnitOutput{}, fmt.Errorf("department %q (%s) has %d subordinate project(s); delete them first (this tool cannot move a project to another department — parentId is create-only): %w",
			name, id, childCount, runaimcp.ErrConflict)
	}

	// Deleting a project cascades to its workloads (unlike a department, which is blocked
	// above while it has projects). Surface this at confirm time and on the result — the UI
	// warns about it too.
	if in.OrgType == common.OrgTypeProject {
		out.Warnings = append(out.Warnings, projectDeleteCascadeWarning)
	}

	if in.DryRun {
		return out, nil
	}

	confirmDetails := out.Before
	if in.OrgType == common.OrgTypeProject {
		confirmDetails = common.WithConfirmWarning(out.Before, projectDeleteCascadeWarning)
	}
	if err := common.ConfirmWrite(ctx, in.Confirm, "delete", in.OrgType, name, id, confirmDetails); err != nil {
		return manageOrgUnitOutput{}, err
	}

	if err := doDelete(ctx, client, in.OrgType, id); err != nil {
		return manageOrgUnitOutput{}, err
	}
	out.Applied = true
	return out, nil
}

const projectDeleteCascadeWarning = "deleting a project permanently deletes all workloads in it"

func loadForDelete(ctx context.Context, client *runaiclient.Client, orgType, id string) (snap map[string]any, childCount int, name, clusterID, parentDeptID string, err error) {
	if orgType == common.OrgTypeProject {
		p, gErr := getProject(ctx, client, id)
		if gErr != nil {
			return nil, 0, "", "", "", gErr
		}
		return projectSnapshot(p), 0, p.Name, p.ClusterId, p.ParentId, nil
	}
	d, gErr := getDepartment(ctx, client, id)
	if gErr != nil {
		return nil, 0, "", "", "", gErr
	}
	return departmentSnapshot(d), len(d.Children), d.Name, d.ClusterId, "", nil
}

func doDelete(ctx context.Context, client *runaiclient.Client, orgType, id string) error {
	if orgType == common.OrgTypeProject {
		httpResp, err := client.OrgUnit.ProjectsAPI.DeleteProject(ctx, id).Execute()
		if err != nil {
			return fmt.Errorf("delete project %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	}
	httpResp, err := client.OrgUnit.DepartmentsAPI.DeleteDepartment(ctx, id).Execute()
	if err != nil {
		return fmt.Errorf("delete department %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	return nil
}

func getProject(ctx context.Context, client *runaiclient.Client, id string) (*orgunit.Project, error) {
	p, httpResp, err := client.OrgUnit.ProjectsAPI.GetProject(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("get project %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	if p == nil {
		return nil, fmt.Errorf("project %q: %w", id, runaimcp.ErrNotFound)
	}
	return p, nil
}

func getDepartment(ctx context.Context, client *runaiclient.Client, id string) (*orgunit.Department, error) {
	d, httpResp, err := client.OrgUnit.DepartmentsAPI.GetDepartment(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("get department %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	if d == nil {
		return nil, fmt.Errorf("department %q: %w", id, runaimcp.ErrNotFound)
	}
	return d, nil
}
