/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"fmt"
	"strings"

	authz "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	clusterservice "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const placementBinPack = "bin-pack"
const placementSpread = "spread"

const (
	keyNetworkTopologyName  = "networkTopologyName"
	keyGPUNetAccelLabelKey  = "gpuNetworkAccelerationLabelKey"
	keyGPUNetAccelDetection = "gpuNetworkAccelerationDetection"
	keyCapAdjustableWeights = "schedulingCapabilities.adjustableWeights"
	keyCapRanks             = "schedulingCapabilities.ranks"
	keyCapQuota             = "schedulingCapabilities.quota"
	keyCapCPUQuota          = "schedulingCapabilities.cpuQuota"
	keyGPUPlacementNode     = "schedulingConfiguration.gpuPodsPlacementStrategy.node"
	keyGPUPlacementDevice   = "schedulingConfiguration.gpuPodsPlacementStrategy.device"
	keyCPUPlacementNode     = "schedulingConfiguration.cpuPodsPlacementStrategy.node"
	keyMinGuaranteedRuntime = "schedulingConfiguration.minGuaranteedRuntime"
	keyNumaAwareScheduling  = "schedulingConfiguration.numaAwareScheduling"
)

type nodePoolScheduling struct {
	CPUPlacementStrategy       string `json:"cpuPlacementStrategy,omitempty"       jsonschema:"Placement of pods across CPU nodes. One of: bin-pack, spread."`
	GPUPlacementStrategy       string `json:"gpuPlacementStrategy,omitempty"       jsonschema:"Placement of GPU pods across GPU nodes (node level). One of: bin-pack, spread."`
	GPUDevicePlacementStrategy string `json:"gpuDevicePlacementStrategy,omitempty" jsonschema:"Placement of GPU pods across GPU devices within a node (device level). One of: bin-pack, spread."`
	MinGuaranteedRuntime       string `json:"minGuaranteedRuntime,omitempty"       jsonschema:"Minimum guaranteed runtime before preemption, as a duration (e.g. 5m, 1h)."`
	NumaAwareScheduling        *bool  `json:"numaAwareScheduling,omitempty"        jsonschema:"Enable NUMA-aware scheduling on the pool."`
}

type nodePoolCapabilities struct {
	AdjustableWeights *bool `json:"adjustableWeights,omitempty" jsonschema:"Pool supports adjustable over-quota weights."`
	Ranks             *bool `json:"ranks,omitempty"             jsonschema:"Pool supports scheduling ranks."`
	Quota             *bool `json:"quota,omitempty"             jsonschema:"Pool supports GPU quota."`
	CPUQuota          *bool `json:"cpuQuota,omitempty"          jsonschema:"Pool supports CPU quota."`
}

type nodePoolSpec struct {
	LabelKey                        string                `json:"labelKey,omitempty"                       jsonschema:"K8s node label key that maps nodes into this pool. Required for create."`
	LabelValue                      string                `json:"labelValue,omitempty"                     jsonschema:"K8s node label value that maps nodes into this pool. Required for create."`
	SchedulingCapabilities          *nodePoolCapabilities `json:"schedulingCapabilities,omitempty"         jsonschema:"Per-pool feature flags controlling what the pool supports."`
	Scheduling                      *nodePoolScheduling   `json:"scheduling,omitempty"                     jsonschema:"Scheduler configuration for the pool."`
	GPUNetworkAccelerationLabelKey  string                `json:"gpuNetworkAccelerationLabelKey,omitempty"  jsonschema:"Node label key used to detect GPU network acceleration."`
	GPUNetworkAccelerationDetection string                `json:"gpuNetworkAccelerationDetection,omitempty" jsonschema:"GPU network acceleration detection mode. One of: Use, DontUse, Auto."`
	NetworkTopologyName             string                `json:"networkTopologyName,omitempty"             jsonschema:"Network topology to bind, by name; resolved to its ID."`
	DetachNetworkTopology           *bool                 `json:"detachNetworkTopology,omitempty"           jsonschema:"Update-only: true clears the network-topology binding. Excludes networkTopologyName."`
}

type manageNodePoolInput struct {
	Action        string        `json:"action"                  jsonschema:"What to do. One of: create, update, delete."`
	ClusterID     string        `json:"clusterId"               jsonschema:"Cluster the node pool lives in (UUID)."`
	NodePoolID    string        `json:"nodePoolId,omitempty"    jsonschema:"Node pool ID. Preferred for update/delete. Mutually exclusive with name."`
	Name          string        `json:"name,omitempty"          jsonschema:"Node pool name. Required for create; resolved to an ID otherwise. Mutually exclusive with nodePoolId. Returns ambiguous_node_pool on a collision."`
	Spec          *nodePoolSpec `json:"spec,omitempty"          jsonschema:"Desired-state spec. Required for create. On update only the fields you set change; omitted ones are preserved."`
	DryRun        bool          `json:"dryRun,omitempty"        jsonschema:"When true, preview the change (before/after/diff) without mutating anything upstream."`
	IgnoreMissing bool          `json:"ignoreMissing,omitempty" jsonschema:"Delete-only. When true, deleting a node pool that no longer exists is a successful no-op instead of not_found."`
	Confirm       bool          `json:"confirm,omitempty"       jsonschema:"Required for a non-dry-run update or delete when the client can't prompt; otherwise returns confirmation_required with the snapshot."`
}

func (in manageNodePoolInput) Validate() error {
	switch common.WriteAction(in.Action) {
	case common.ActionCreate, common.ActionUpdate, common.ActionDelete:
	default:
		return fmt.Errorf("action must be one of: create, update, delete")
	}
	if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
		return err
	}
	if in.NodePoolID != "" && in.Name != "" {
		return fmt.Errorf("nodePoolId and name are mutually exclusive")
	}
	if err := common.ValidateNoCommas("name", in.Name); err != nil {
		return err
	}
	if err := common.ValidateNoCommas("nodePoolId", in.NodePoolID); err != nil {
		return err
	}

	switch common.WriteAction(in.Action) {
	case common.ActionCreate:
		if in.Name == "" {
			return fmt.Errorf("name is required for create")
		}
		if in.NodePoolID != "" {
			return fmt.Errorf("nodePoolId must not be set for create")
		}
		if in.Spec == nil || in.Spec.LabelKey == "" || in.Spec.LabelValue == "" {
			return fmt.Errorf("spec.labelKey and spec.labelValue are required for create")
		}
	case common.ActionUpdate:
		if in.NodePoolID == "" && in.Name == "" {
			return fmt.Errorf("one of nodePoolId or name is required for update")
		}
		if in.Spec == nil {
			return fmt.Errorf("spec is required for update")
		}
	case common.ActionDelete:
		if in.NodePoolID == "" && in.Name == "" {
			return fmt.Errorf("one of nodePoolId or name is required for delete")
		}
	}

	if in.IgnoreMissing && common.WriteAction(in.Action) != common.ActionDelete {
		return fmt.Errorf("ignoreMissing is only valid with action=delete")
	}
	if in.Spec != nil {
		if err := common.ValidateNoCommas("spec.networkTopologyName", in.Spec.NetworkTopologyName); err != nil {
			return err
		}
		if in.Spec.DetachNetworkTopology != nil && *in.Spec.DetachNetworkTopology {
			if in.Spec.NetworkTopologyName != "" {
				return fmt.Errorf("spec.detachNetworkTopology and spec.networkTopologyName are mutually exclusive")
			}
			if common.WriteAction(in.Action) != common.ActionUpdate {
				return fmt.Errorf("spec.detachNetworkTopology is only valid with action=update")
			}
		}
		if s := in.Spec.Scheduling; s != nil {
			for field, v := range map[string]string{
				"spec.scheduling.cpuPlacementStrategy":       s.CPUPlacementStrategy,
				"spec.scheduling.gpuPlacementStrategy":       s.GPUPlacementStrategy,
				"spec.scheduling.gpuDevicePlacementStrategy": s.GPUDevicePlacementStrategy,
			} {
				if v != "" && v != placementBinPack && v != placementSpread {
					return fmt.Errorf("%s must be one of: %s, %s", field, placementBinPack, placementSpread)
				}
			}
		}
	}
	return nil
}

type manageNodePoolOutput struct {
	Action     string         `json:"action"               jsonschema:"The action performed: create, update, or delete."`
	DryRun     bool           `json:"dryRun"               jsonschema:"True when this was a preview; no upstream change was made."`
	Applied    bool           `json:"applied"              jsonschema:"True when the change was actually applied upstream. False on dry-run or an ignored-missing delete."`
	ClusterID  string         `json:"clusterId"            jsonschema:"Cluster the pool belongs to."`
	NodePoolID string         `json:"nodePoolId,omitempty" jsonschema:"Resolved node pool ID. Empty on a create dry-run."`
	Name       string         `json:"name,omitempty"       jsonschema:"Node pool name."`
	Before     map[string]any `json:"before"               jsonschema:"Pool snapshot before the change. Empty object on create."`
	After      map[string]any `json:"after"                jsonschema:"Pool snapshot after the change. Empty object on delete."`
	Diff       map[string]any `json:"diff"                 jsonschema:"Field-level changes as path -> {before, after}."`
	Warnings   []string       `json:"warnings,omitempty"   jsonschema:"Non-fatal notes, e.g. when a delete target was already missing."`
}

func registerManageNodePool(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageNodePoolInput, manageNodePoolOutput]{
		Name: "manage_node_pool",
		Description: "Create, update, or delete a node pool in a cluster — its scheduler, scheduling capabilities, " +
			"GPU network acceleration, GPU resource optimization, and network-topology binding. The action field " +
			"selects the verb (create | update | delete). For high-impact changes, call once with dryRun=true to " +
			"preview the diff, then again with dryRun=false to apply. A non-dry-run delete requires confirm=true. " +
			"Do NOT call to list pools (use list_node_pools), for project/department quota (use list_project_resources " +
			"/ list_department_resources), or for tenant-level settings.",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageNodePoolInput) (manageNodePoolOutput, error) {
			return manageNodePoolRun(ctx, client, in)
		},
	})
}

func manageNodePoolRun(ctx context.Context, client *runaiclient.Client, in manageNodePoolInput) (manageNodePoolOutput, error) {
	action := common.WriteAction(in.Action)
	out := manageNodePoolOutput{
		Action:    in.Action,
		DryRun:    in.DryRun,
		ClusterID: in.ClusterID,
		Name:      in.Name,
		// before/after are schema-typed objects; absent state is an empty object,
		// not null, so the framework's output validation accepts it.
		Before: map[string]any{},
		After:  map[string]any{},
		Diff:   map[string]any{},
	}

	if err := authorizeNodePool(ctx, client, action, in.ClusterID); err != nil {
		return manageNodePoolOutput{}, err
	}

	id := in.NodePoolID
	if action != common.ActionCreate && id == "" {
		resolved, err := common.ResolveNodePoolID(ctx, client, in.ClusterID, in.Name)
		if err != nil {
			// A delete-by-name of an already-gone pool fails here; honor ignoreMissing.
			if action == common.ActionDelete && in.IgnoreMissing && common.IsNotFound(err) {
				out.Warnings = append(out.Warnings, fmt.Sprintf("node pool %q not found; nothing to delete", in.Name))
				return out, nil
			}
			return manageNodePoolOutput{}, err
		}
		id = resolved
	}
	out.NodePoolID = id

	switch action {
	case common.ActionCreate:
		return manageNodePoolCreate(ctx, client, in, out)
	case common.ActionUpdate:
		return manageNodePoolUpdate(ctx, client, in, out, id)
	case common.ActionDelete:
		return manageNodePoolDelete(ctx, client, in, out, id)
	default:
		return manageNodePoolOutput{}, fmt.Errorf("unsupported action %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

func manageNodePoolCreate(ctx context.Context, client *runaiclient.Client, in manageNodePoolInput, out manageNodePoolOutput) (manageNodePoolOutput, error) {
	topologyID, err := resolveTopology(ctx, client, in.ClusterID, in.Spec.NetworkTopologyName)
	if err != nil {
		return manageNodePoolOutput{}, err
	}

	fields := clusterservice.NewNodepoolCreateFields(in.Name, in.Spec.LabelKey, in.Spec.LabelValue, in.ClusterID)
	// The constructor injects a default gpuNetworkAccelerationDetection=Auto; clear it
	// so we only send GPU-network-acceleration settings the caller actually requested
	// (Auto is rejected on clusters/hardware that don't support it).
	fields.UnsetGpuNetworkAccelerationDetection()
	applyCreateSpec(fields, in.Spec, topologyID)

	if in.DryRun {
		out.After = specPreview(in.Name, in.ClusterID, in.Spec)
		out.Diff = common.DiffMaps(nil, out.After)
		return out, nil
	}

	created, httpResp, err := client.Cluster.NodePoolsAPI.CreateNodepool(ctx).NodepoolCreateFields(*fields).Execute()
	if err != nil {
		return manageNodePoolOutput{}, fmt.Errorf("create node pool %q: %w", in.Name, runaimcp.MapHTTPError(httpResp, err))
	}
	if created == nil {
		return manageNodePoolOutput{}, fmt.Errorf("create node pool %q: nil response: %w", in.Name, runaimcp.ErrUpstream)
	}
	out.After = nodePoolSnapshot(created)
	out.NodePoolID = created.Id
	out.Applied = true
	out.Diff = common.DiffMaps(nil, out.After)
	return out, nil
}

func manageNodePoolUpdate(ctx context.Context, client *runaiclient.Client, in manageNodePoolInput, out manageNodePoolOutput, id string) (manageNodePoolOutput, error) {
	current, err := getNodePool(ctx, client, id, in.ClusterID)
	if err != nil {
		return manageNodePoolOutput{}, err
	}
	out.Before = nodePoolSnapshot(current)
	out.Name = current.Name

	spec := *in.Spec
	if spec.LabelKey == "" {
		spec.LabelKey = current.LabelKey
	}
	if spec.LabelValue == "" {
		spec.LabelValue = current.LabelValue
	}
	// Upstream resets omitted schedulingConfiguration fields on update — even an unrelated
	// label change wipes GPU placement. Always re-send the full block: seed every scheduling
	// field from the current pool, then let the caller's spec override what it sets.
	sched := nodePoolScheduling{}
	if spec.Scheduling != nil {
		sched = *spec.Scheduling
	}
	carryForwardScheduling(&sched, current)
	spec.Scheduling = &sched

	carryForwardCapabilities(&spec, current)
	carryForwardGPUNetworkAcceleration(&spec, current)

	topologyID, err := resolveTopology(ctx, client, in.ClusterID, spec.NetworkTopologyName)
	if err != nil {
		return manageNodePoolOutput{}, err
	}

	detaching := spec.DetachNetworkTopology != nil && *spec.DetachNetworkTopology
	if topologyID == "" && !detaching {
		if v := current.NetworkTopologyId.Get(); v != nil {
			topologyID = *v
		}
	}

	fields := clusterservice.NewNodepoolUpdateFields()
	// The constructor injects gpuNetworkAccelerationDetection=Auto; clear it so an unrelated
	// update doesn't send Auto (rejected on clusters/hardware that don't support it). It is
	// re-set below only when the caller asks for it.
	fields.UnsetGpuNetworkAccelerationDetection()
	applyUpdateSpec(fields, &spec, topologyID)
	if spec.DetachNetworkTopology != nil && *spec.DetachNetworkTopology {
		fields.SetNetworkTopologyIdNil()
	}

	if in.DryRun {
		out.After = mergePreview(out.Before, &spec)
		out.Diff = common.DiffMaps(out.Before, out.After)
		return out, nil
	}

	if err := common.ConfirmWrite(ctx, in.Confirm, "update", "node pool", current.Name, id,
		common.DiffMaps(out.Before, mergePreview(out.Before, in.Spec))); err != nil {
		return manageNodePoolOutput{}, err
	}

	updated, httpResp, err := client.Cluster.NodePoolsAPI.UpdateNodepool(ctx, id).NodepoolUpdateFields(*fields).Execute()
	if err != nil {
		return manageNodePoolOutput{}, fmt.Errorf("update node pool %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	if updated == nil {
		return manageNodePoolOutput{}, fmt.Errorf("update node pool %q: nil response: %w", id, runaimcp.ErrUpstream)
	}
	out.After = nodePoolSnapshot(updated)
	out.Applied = true
	out.Diff = common.DiffMaps(out.Before, out.After)
	return out, nil
}

func manageNodePoolDelete(ctx context.Context, client *runaiclient.Client, in manageNodePoolInput, out manageNodePoolOutput, id string) (manageNodePoolOutput, error) {
	current, err := getNodePool(ctx, client, id, in.ClusterID)
	if err != nil {
		if in.IgnoreMissing && common.IsNotFound(err) {
			out.Warnings = append(out.Warnings, "node pool not found; nothing to delete")
			return out, nil
		}
		return manageNodePoolOutput{}, err
	}
	out.Before = nodePoolSnapshot(current)
	out.Name = current.Name
	out.Diff = common.DiffMaps(out.Before, nil)

	// The backend forbids deleting the default pool and rejects it with an error
	// body the generated client can't decode; catch it here with a clear message.
	if current.IsDefault {
		return manageNodePoolOutput{}, fmt.Errorf("cannot delete the default node pool %q: %w", current.Name, runaimcp.ErrValidation)
	}

	if in.DryRun {
		return out, nil
	}

	if err := common.ConfirmWrite(ctx, in.Confirm, "delete", "node pool", current.Name, id, out.Before); err != nil {
		return manageNodePoolOutput{}, err
	}

	_, httpResp, err := client.Cluster.NodePoolsAPI.DeleteNodepool(ctx, id).Execute()
	if err != nil {
		return manageNodePoolOutput{}, fmt.Errorf("delete node pool %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	out.Applied = true
	return out, nil
}

func authorizeNodePool(ctx context.Context, client *runaiclient.Client, action common.WriteAction, clusterID string) error {
	scope, err := common.PermittedScopeForAction(ctx, client, authz.RESOURCETYPE_NODEPOOLS, action.Authz())
	if err != nil {
		return err
	}
	if common.ScopeAllowsCluster(scope, clusterID) {
		return nil
	}
	return fmt.Errorf("not permitted to %s node pools on cluster %q: %w", action, clusterID, runaimcp.ErrForbidden)
}

func resolveTopology(ctx context.Context, client *runaiclient.Client, clusterID, name string) (string, error) {
	if name == "" {
		return "", nil
	}
	return resolveNetworkTopologyID(ctx, client, clusterID, name)
}

func resolveNetworkTopologyID(ctx context.Context, client *runaiclient.Client, clusterID, name string) (string, error) {
	if err := common.ValidateFilterValue("networkTopologyName", name); err != nil {
		return "", fmt.Errorf("%v: %w", err, runaimcp.ErrValidation)
	}
	filters := []string{common.FilterEq("clusterId", clusterID), common.FilterEq("name", name)}
	resp, httpResp, err := client.Cluster.NetworkTopologiesAPI.GetNetworkTopologies(ctx).FilterBy(filters).Execute()
	if err != nil {
		return "", fmt.Errorf("resolve network topology %q: %w", name, runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil || len(resp.NetworkTopologies) == 0 {
		return "", fmt.Errorf("network topology %q on cluster %q: %w", name, clusterID, runaimcp.ErrNotFound)
	}
	if len(resp.NetworkTopologies) > 1 {
		ids := make([]string, 0, len(resp.NetworkTopologies))
		for _, t := range resp.NetworkTopologies {
			ids = append(ids, t.Id)
		}
		return "", fmt.Errorf("ambiguous_network_topology: name %q matches multiple topologies (%s): %w",
			name, strings.Join(ids, ", "), runaimcp.ErrValidation)
	}
	return resp.NetworkTopologies[0].Id, nil
}

func getNodePool(ctx context.Context, client *runaiclient.Client, id, clusterID string) (*clusterservice.Nodepool, error) {
	np, httpResp, err := client.Cluster.NodePoolsAPI.GetNodepool(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("get node pool %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	if np == nil {
		return nil, fmt.Errorf("node pool %q: %w", id, runaimcp.ErrNotFound)
	}
	// The pool is fetched by ID alone and clusterId otherwise only gates the auth check, which a
	// system/tenant-scoped caller passes for any cluster. Verify the pool really lives in the
	// requested cluster so a wrong clusterId is rejected rather than silently acted on.
	if np.ClusterId != clusterID {
		return nil, fmt.Errorf("node pool %q is not in cluster %q: %w", id, clusterID, runaimcp.ErrNotFound)
	}
	return np, nil
}

type nodePoolFieldSetter interface {
	SetGpuNetworkAccelerationLabelKey(string)
	SetGpuNetworkAccelerationDetection(clusterservice.NodepoolGPUNetworkAccelerationDetection)
	SetNetworkTopologyId(string)
	SetSchedulingCapabilities(clusterservice.SchedulingCapabilitiesNullable)
	SetSchedulingConfiguration(clusterservice.SchedulingConfiguration)
}

func applySharedSpec(fields nodePoolFieldSetter, spec *nodePoolSpec, topologyID string) {
	if spec.GPUNetworkAccelerationLabelKey != "" {
		fields.SetGpuNetworkAccelerationLabelKey(spec.GPUNetworkAccelerationLabelKey)
	}
	if spec.GPUNetworkAccelerationDetection != "" {
		fields.SetGpuNetworkAccelerationDetection(clusterservice.NodepoolGPUNetworkAccelerationDetection(spec.GPUNetworkAccelerationDetection))
	}
	if topologyID != "" {
		fields.SetNetworkTopologyId(topologyID)
	}
	if caps := buildCapabilities(spec.SchedulingCapabilities); caps != nil {
		fields.SetSchedulingCapabilities(*caps)
	}
	if sched := buildScheduling(spec.Scheduling); sched != nil {
		fields.SetSchedulingConfiguration(*sched)
	}
}

func applyCreateSpec(fields *clusterservice.NodepoolCreateFields, spec *nodePoolSpec, topologyID string) {
	applySharedSpec(fields, spec, topologyID)
}

func applyUpdateSpec(fields *clusterservice.NodepoolUpdateFields, spec *nodePoolSpec, topologyID string) {
	if spec.LabelKey != "" {
		fields.SetLabelKey(spec.LabelKey)
	}
	if spec.LabelValue != "" {
		fields.SetLabelValue(spec.LabelValue)
	}
	applySharedSpec(fields, spec, topologyID)
}

func buildCapabilities(in *nodePoolCapabilities) *clusterservice.SchedulingCapabilitiesNullable {
	if in == nil {
		return nil
	}
	caps := clusterservice.NewSchedulingCapabilitiesNullable()
	if in.AdjustableWeights != nil {
		caps.SetAdjustableWeights(*in.AdjustableWeights)
	}
	if in.Ranks != nil {
		caps.SetRanks(*in.Ranks)
	}
	if in.Quota != nil {
		caps.SetQuota(*in.Quota)
	}
	if in.CPUQuota != nil {
		caps.SetCpuQuota(*in.CPUQuota)
	}
	return caps
}

func buildScheduling(in *nodePoolScheduling) *clusterservice.SchedulingConfiguration {
	if in == nil {
		return nil
	}
	sched := clusterservice.NewSchedulingConfiguration()
	if in.MinGuaranteedRuntime != "" {
		sched.SetMinGuaranteedRuntime(in.MinGuaranteedRuntime)
	}
	if in.NumaAwareScheduling != nil {
		sched.SetNumaAwareScheduling(*in.NumaAwareScheduling)
	}

	node := toPlacementStrategy(in.GPUPlacementStrategy)
	device := toPlacementStrategy(in.GPUDevicePlacementStrategy)
	if node != "" || device != "" {
		gpu := clusterservice.NewGpuPodsPlacementStrategy()
		if node != "" {
			gpu.SetNode(node)
		}
		if device != "" {
			gpu.SetDevice(device)
		}
		sched.SetGpuPodsPlacementStrategy(*gpu)
	}
	if s := toPlacementStrategy(in.CPUPlacementStrategy); s != "" {
		cpu := clusterservice.NewCpuPodsPlacementStrategy()
		cpu.SetNode(s)
		sched.SetCpuPodsPlacementStrategy(*cpu)
	}
	return sched
}

func toPlacementStrategy(v string) clusterservice.PlacementStrategy {
	switch v {
	case placementBinPack:
		return clusterservice.PLACEMENTSTRATEGY_BINPACK
	case placementSpread:
		return clusterservice.PLACEMENTSTRATEGY_SPREAD
	default:
		return ""
	}
}

func fromPlacementStrategy(v clusterservice.PlacementStrategy) string {
	switch v {
	case clusterservice.PLACEMENTSTRATEGY_BINPACK:
		return placementBinPack
	case clusterservice.PLACEMENTSTRATEGY_SPREAD:
		return placementSpread
	default:
		return ""
	}
}

// carryForwardScheduling seeds every scheduling field the caller left unset from the current pool.
func carryForwardScheduling(sched *nodePoolScheduling, current *clusterservice.Nodepool) {
	sc := current.SchedulingConfiguration.Get()
	if sc == nil {
		return
	}
	if gpu := sc.GpuPodsPlacementStrategy.Get(); gpu != nil {
		if sched.GPUPlacementStrategy == "" {
			if v := gpu.Node.Get(); v != nil {
				sched.GPUPlacementStrategy = fromPlacementStrategy(*v)
			}
		}
		if sched.GPUDevicePlacementStrategy == "" {
			if v := gpu.Device.Get(); v != nil {
				sched.GPUDevicePlacementStrategy = fromPlacementStrategy(*v)
			}
		}
	}
	if cpu := sc.CpuPodsPlacementStrategy.Get(); cpu != nil && sched.CPUPlacementStrategy == "" {
		if v := cpu.Node.Get(); v != nil {
			sched.CPUPlacementStrategy = fromPlacementStrategy(*v)
		}
	}
	if sched.MinGuaranteedRuntime == "" {
		if v := sc.MinGuaranteedRuntime.Get(); v != nil {
			sched.MinGuaranteedRuntime = *v
		}
	}
	if sched.NumaAwareScheduling == nil {
		if v := sc.NumaAwareScheduling.Get(); v != nil {
			sched.NumaAwareScheduling = v
		}
	}
}

// carryForwardCapabilities seeds scheduling-capability flags the caller left unset
// from the current pool.
func carryForwardCapabilities(spec *nodePoolSpec, current *clusterservice.Nodepool) {
	c := current.SchedulingCapabilities
	if c == nil {
		return
	}
	caps := nodePoolCapabilities{}
	if spec.SchedulingCapabilities != nil {
		caps = *spec.SchedulingCapabilities
	}
	if caps.AdjustableWeights == nil {
		caps.AdjustableWeights = boolPtr(c.AdjustableWeights)
	}
	if caps.Ranks == nil {
		caps.Ranks = boolPtr(c.Ranks)
	}
	if caps.Quota == nil {
		caps.Quota = boolPtr(c.Quota)
	}
	if caps.CPUQuota == nil {
		caps.CPUQuota = boolPtr(c.CpuQuota)
	}
	spec.SchedulingCapabilities = &caps
}

// carryForwardGPUNetworkAcceleration carries the label key and detection mode forward.
// Detection is only carried when the pool already has one, so we never resurrect the
// generated-constructor default the backend rejects on unsupported hardware.
func carryForwardGPUNetworkAcceleration(spec *nodePoolSpec, current *clusterservice.Nodepool) {
	if spec.GPUNetworkAccelerationLabelKey == "" && current.GpuNetworkAccelerationLabelKey != nil {
		spec.GPUNetworkAccelerationLabelKey = *current.GpuNetworkAccelerationLabelKey
	}
	if spec.GPUNetworkAccelerationDetection == "" {
		if v := current.GpuNetworkAccelerationDetection.Get(); v != nil {
			spec.GPUNetworkAccelerationDetection = string(*v)
		}
	}
}

func boolPtr(b bool) *bool { return &b }

// nodePoolSnapshot maps the upstream model to a compact view, deliberately excluding
// the deprecated top-level placementStrategy so callers don't reason from it.
func nodePoolSnapshot(np *clusterservice.Nodepool) map[string]any {
	if np == nil {
		return nil
	}
	snap := map[string]any{
		"id":         np.Id,
		"name":       np.Name,
		"clusterId":  np.ClusterId,
		"labelKey":   np.LabelKey,
		"labelValue": np.LabelValue,
		"isDefault":  np.IsDefault,
	}
	if np.Phase != nil {
		snap["phase"] = string(*np.Phase)
	}
	if v := np.NetworkTopologyName.Get(); v != nil {
		snap[keyNetworkTopologyName] = *v
	}
	if v := np.NetworkTopologyId.Get(); v != nil {
		snap["networkTopologyId"] = *v
	}
	if np.GpuNetworkAccelerationLabelKey != nil {
		snap[keyGPUNetAccelLabelKey] = *np.GpuNetworkAccelerationLabelKey
	}
	if v := np.GpuNetworkAccelerationDetection.Get(); v != nil {
		snap[keyGPUNetAccelDetection] = string(*v)
	}
	if c := np.SchedulingCapabilities; c != nil {
		snap[keyCapAdjustableWeights] = c.AdjustableWeights
		snap[keyCapRanks] = c.Ranks
		snap[keyCapQuota] = c.Quota
		snap[keyCapCPUQuota] = c.CpuQuota
	}
	if sc := np.SchedulingConfiguration.Get(); sc != nil {
		if gpu := sc.GpuPodsPlacementStrategy.Get(); gpu != nil {
			if v := gpu.Node.Get(); v != nil {
				snap[keyGPUPlacementNode] = string(*v)
			}
			if v := gpu.Device.Get(); v != nil {
				snap[keyGPUPlacementDevice] = string(*v)
			}
		}
		if cpu := sc.CpuPodsPlacementStrategy.Get(); cpu != nil {
			if v := cpu.Node.Get(); v != nil {
				snap[keyCPUPlacementNode] = string(*v)
			}
		}
		if v := sc.MinGuaranteedRuntime.Get(); v != nil {
			snap[keyMinGuaranteedRuntime] = *v
		}
		if v := sc.NumaAwareScheduling.Get(); v != nil {
			snap[keyNumaAwareScheduling] = *v
		}
	}
	return snap
}

func specPreview(name, clusterID string, spec *nodePoolSpec) map[string]any {
	snap := map[string]any{
		"name":       name,
		"clusterId":  clusterID,
		"labelKey":   spec.LabelKey,
		"labelValue": spec.LabelValue,
	}
	addSpecPreviewFields(snap, spec)
	return snap
}

func mergePreview(before map[string]any, spec *nodePoolSpec) map[string]any {
	after := make(map[string]any, len(before))
	for k, v := range before {
		after[k] = v
	}
	if spec.LabelKey != "" {
		after["labelKey"] = spec.LabelKey
	}
	if spec.LabelValue != "" {
		after["labelValue"] = spec.LabelValue
	}
	if spec.DetachNetworkTopology != nil && *spec.DetachNetworkTopology {
		delete(after, keyNetworkTopologyName)
		delete(after, "networkTopologyId")
	}
	addSpecPreviewFields(after, spec)
	return after
}

func addSpecPreviewFields(snap map[string]any, spec *nodePoolSpec) {
	if spec.NetworkTopologyName != "" {
		snap[keyNetworkTopologyName] = spec.NetworkTopologyName
	}
	if spec.GPUNetworkAccelerationLabelKey != "" {
		snap[keyGPUNetAccelLabelKey] = spec.GPUNetworkAccelerationLabelKey
	}
	if spec.GPUNetworkAccelerationDetection != "" {
		snap[keyGPUNetAccelDetection] = spec.GPUNetworkAccelerationDetection
	}
	if c := spec.SchedulingCapabilities; c != nil {
		setIfNonNil(snap, keyCapAdjustableWeights, c.AdjustableWeights)
		setIfNonNil(snap, keyCapRanks, c.Ranks)
		setIfNonNil(snap, keyCapQuota, c.Quota)
		setIfNonNil(snap, keyCapCPUQuota, c.CPUQuota)
	}
	if s := spec.Scheduling; s != nil {
		if v := toPlacementStrategy(s.GPUPlacementStrategy); v != "" {
			snap[keyGPUPlacementNode] = string(v)
		}
		if v := toPlacementStrategy(s.GPUDevicePlacementStrategy); v != "" {
			snap[keyGPUPlacementDevice] = string(v)
		}
		if v := toPlacementStrategy(s.CPUPlacementStrategy); v != "" {
			snap[keyCPUPlacementNode] = string(v)
		}
		if s.MinGuaranteedRuntime != "" {
			snap[keyMinGuaranteedRuntime] = s.MinGuaranteedRuntime
		}
		setIfNonNil(snap, keyNumaAwareScheduling, s.NumaAwareScheduling)
	}
}

func setIfNonNil(snap map[string]any, key string, v *bool) {
	if v != nil {
		snap[key] = *v
	}
}
