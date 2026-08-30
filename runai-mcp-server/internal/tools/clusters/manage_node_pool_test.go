package clusters

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient/runaiclienttest"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

const (
	mnpClusterUUID = "d1405b3f-761a-4d4b-acb6-a00327ca853c"
	permScopesPath = "/api/v1/authorization/permitted-scopes"
	nodePoolsPath  = "/api/v1/node-pools"
)

// permScopesHandler grants the given actions on mnpClusterUUID; everything else is denied.
func permScopesHandler(grant ...string) http.HandlerFunc {
	allowed := map[string]bool{}
	for _, a := range grant {
		allowed[a] = true
	}
	scope := func(action string) map[string]any {
		if allowed[action] {
			return map[string]any{"clusters": []string{mnpClusterUUID}}
		}
		return map[string]any{"clusters": []string{}}
	}
	return func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"create": scope("create"),
			"read":   scope("read"),
			"update": scope("update"),
			"delete": scope("delete"),
		})
	}
}

// permScopesTenantHandler grants all actions tenant-wide — the shape a tenant
// admin's permitted-scopes actually returns (tenants set, clusters empty).
func permScopesTenantHandler() http.HandlerFunc {
	scope := map[string]any{"tenants": []string{"1001"}}
	return func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"create": scope, "read": scope, "update": scope, "delete": scope,
		})
	}
}

func mnpInput(action string) manageNodePoolInput {
	return manageNodePoolInput{Action: action, ClusterID: mnpClusterUUID}
}

func TestManageNodePoolInput_Validate(t *testing.T) {
	createSpec := &nodePoolSpec{LabelKey: "k", LabelValue: "v"}
	cases := []struct {
		name    string
		in      manageNodePoolInput
		wantErr string
	}{
		{name: "create_happy", in: manageNodePoolInput{Action: "create", ClusterID: mnpClusterUUID, Name: "gpu-a", Spec: createSpec}},
		{name: "update_happy", in: manageNodePoolInput{Action: "update", ClusterID: mnpClusterUUID, NodePoolID: "np-1", Spec: &nodePoolSpec{}}},
		{name: "delete_happy", in: manageNodePoolInput{Action: "delete", ClusterID: mnpClusterUUID, Name: "gpu-a"}},
		{name: "bad_action", in: manageNodePoolInput{Action: "frobnicate", ClusterID: mnpClusterUUID}, wantErr: "action must be one of"},
		{name: "bad_cluster", in: manageNodePoolInput{Action: "delete", ClusterID: "not-a-uuid", Name: "x"}, wantErr: "UUID"},
		{name: "id_and_name", in: manageNodePoolInput{Action: "delete", ClusterID: mnpClusterUUID, NodePoolID: "np-1", Name: "x"}, wantErr: "mutually exclusive"},
		{name: "create_no_name", in: manageNodePoolInput{Action: "create", ClusterID: mnpClusterUUID, Spec: createSpec}, wantErr: "name is required"},
		{name: "create_no_labels", in: manageNodePoolInput{Action: "create", ClusterID: mnpClusterUUID, Name: "gpu-a", Spec: &nodePoolSpec{}}, wantErr: "labelKey"},
		{name: "update_no_target", in: manageNodePoolInput{Action: "update", ClusterID: mnpClusterUUID, Spec: &nodePoolSpec{}}, wantErr: "one of nodePoolId or name"},
		{name: "ignore_missing_on_create", in: manageNodePoolInput{Action: "create", ClusterID: mnpClusterUUID, Name: "x", Spec: createSpec, IgnoreMissing: true}, wantErr: "ignoreMissing is only valid"},
		{name: "comma_in_name", in: manageNodePoolInput{Action: "delete", ClusterID: mnpClusterUUID, Name: "a,b"}, wantErr: "name"},
		{name: "bad_placement", in: manageNodePoolInput{Action: "update", ClusterID: mnpClusterUUID, NodePoolID: "np-1", Spec: &nodePoolSpec{Scheduling: &nodePoolScheduling{GPUPlacementStrategy: "sideways"}}}, wantErr: "must be one of"},
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

func TestManageNodePool_CreateHappy(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("create"))
	mux.HandleFunc(nodePoolsPath, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "np-new", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v", "isDefault": false,
		})
	})

	in := mnpInput("create")
	in.Name = "gpu-a"
	in.Spec = &nodePoolSpec{LabelKey: "k", LabelValue: "v"}
	in.Confirm = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "np-new", out.NodePoolID)
	assert.Empty(t, out.Before)
	assert.Equal(t, "gpu-a", out.After["name"])
}

func TestManageNodePool_CreateConflict(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("create"))
	mux.HandleFunc(nodePoolsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusConflict, map[string]any{"message": "exists"})
	})

	in := mnpInput("create")
	in.Name = "gpu-a"
	in.Spec = &nodePoolSpec{LabelKey: "k", LabelValue: "v"}
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	assert.ErrorIs(t, err, runaimcp.ErrConflict)
}

func TestManageNodePool_UpdateHappy(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "old", "isDefault": false,
			})
		case http.MethodPut:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "new", "isDefault": false,
			})
		default:
			http.Error(w, "unexpected method "+r.Method, http.StatusInternalServerError)
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{LabelValue: "new"}
	in.Confirm = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "old", out.Before["labelValue"])
	assert.Equal(t, "new", out.After["labelValue"])
	assert.Contains(t, out.Diff, "labelValue")
}

// A broadly-scoped caller passes the auth check for any clusterId, so the pool is fetched by
// ID alone. The tool must still reject an update when the pool isn't in the requested cluster.
func TestManageNodePool_UpdateWrongClusterRejected(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	var putCalled int32
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": "11111111-1111-1111-1111-111111111111",
				"labelKey": "k", "labelValue": "v", "isDefault": false,
			})
		case http.MethodPut:
			atomic.AddInt32(&putCalled, 1)
			w.WriteHeader(http.StatusOK)
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{LabelValue: "new"}
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
	assert.Contains(t, err.Error(), "not in cluster")
	assert.Equal(t, int32(0), atomic.LoadInt32(&putCalled), "update must not fire for a pool in a different cluster")
}

// The backend requires labelKey/labelValue in every PUT (400 otherwise). When the
// caller's spec omits them, the tool must carry them forward from the current pool.
func TestManageNodePool_UpdateCarriesLabelsForward(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	var putBody map[string]any
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "existing-key", "labelValue": "existing-val",
			})
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "existing-key", "labelValue": "existing-val",
			})
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	// Spec changes only scheduling — omits labelKey/labelValue entirely.
	in.Spec = &nodePoolSpec{Scheduling: &nodePoolScheduling{GPUPlacementStrategy: "spread"}}
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)
	require.NotNil(t, putBody, "PUT must have been sent")
	assert.Equal(t, "existing-key", putBody["labelKey"], "labelKey carried from current pool")
	assert.Equal(t, "existing-val", putBody["labelValue"], "labelValue carried from current pool")
}

// Upstream replaces the whole gpuPodsPlacementStrategy block on update, so a
// device-only change must carry the current node placement forward (and vice
// versa) or the omitted axis silently resets to its default.
func TestManageNodePool_UpdateCarriesGPUPlacementForward(t *testing.T) {
	gpuPlacement := func(body map[string]any) map[string]any {
		sc, _ := body["schedulingConfiguration"].(map[string]any)
		gpu, _ := sc["gpuPodsPlacementStrategy"].(map[string]any)
		return gpu
	}
	cases := []struct {
		name       string
		current    map[string]any
		spec       *nodePoolScheduling
		wantNode   any
		wantDevice any
	}{
		{
			name:       "device_only_preserves_node",
			current:    map[string]any{"node": "spread", "device": "binpack"},
			spec:       &nodePoolScheduling{GPUDevicePlacementStrategy: "spread"},
			wantNode:   "spread",
			wantDevice: "spread",
		},
		{
			name:       "node_only_preserves_device",
			current:    map[string]any{"node": "binpack", "device": "spread"},
			spec:       &nodePoolScheduling{GPUPlacementStrategy: "spread"},
			wantNode:   "spread",
			wantDevice: "spread",
		},
		{
			name:       "both_set_no_seeding",
			current:    map[string]any{"node": "spread", "device": "spread"},
			spec:       &nodePoolScheduling{GPUPlacementStrategy: "bin-pack", GPUDevicePlacementStrategy: "bin-pack"},
			wantNode:   "binpack",
			wantDevice: "binpack",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(permScopesPath, permScopesHandler("update"))
			var putBody map[string]any
			mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
						"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
						"schedulingConfiguration": map[string]any{"gpuPodsPlacementStrategy": tc.current},
					})
				case http.MethodPut:
					_ = json.NewDecoder(r.Body).Decode(&putBody)
					runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
						"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
					})
				}
			})

			in := mnpInput("update")
			in.NodePoolID = "np-1"
			in.Spec = &nodePoolSpec{Scheduling: tc.spec}
			in.Confirm = true
			_, err := manageNodePoolRun(context.Background(), client, in)
			require.NoError(t, err)
			require.NotNil(t, putBody, "PUT must have been sent")
			gpu := gpuPlacement(putBody)
			require.NotNil(t, gpu, "gpuPodsPlacementStrategy must be present in PUT")
			assert.Equal(t, tc.wantNode, gpu["node"], "node placement in PUT")
			assert.Equal(t, tc.wantDevice, gpu["device"], "device placement in PUT")
		})
	}
}

// Upstream resets omitted schedulingConfiguration fields on update — even an unrelated label
// change wipes GPU placement. The whole scheduling block must be re-sent from the current pool.
func TestManageNodePool_UpdatePreservesSchedulingOnUnrelatedChange(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	var putBody map[string]any
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "old",
				"schedulingConfiguration": map[string]any{
					"gpuPodsPlacementStrategy": map[string]any{"node": "spread", "device": "binpack"},
					"cpuPodsPlacementStrategy": map[string]any{"node": "spread"},
					"minGuaranteedRuntime":     "10m",
					"numaAwareScheduling":      true,
				},
			})
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "new",
			})
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{LabelValue: "new"} // scheduling not touched at all
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)

	sc, ok := putBody["schedulingConfiguration"].(map[string]any)
	require.True(t, ok, "full scheduling must be re-sent even on a label-only update")
	gpu, _ := sc["gpuPodsPlacementStrategy"].(map[string]any)
	require.NotNil(t, gpu)
	assert.Equal(t, "spread", gpu["node"], "gpu node placement preserved")
	assert.Equal(t, "binpack", gpu["device"], "gpu device placement preserved")
	assert.Equal(t, "spread", sc["cpuPodsPlacementStrategy"].(map[string]any)["node"], "cpu placement preserved")
	assert.Equal(t, "10m", sc["minGuaranteedRuntime"], "minGuaranteedRuntime preserved")
	assert.Equal(t, true, sc["numaAwareScheduling"], "numaAwareScheduling preserved")
}

// Attaching a network topology is an unrelated update: it must set the topology and preserve
// the pool's placement strategies, not reset them to defaults.
func TestManageNodePool_UpdateAttachTopologyPreservesScheduling(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	mux.HandleFunc("/api/v1/network-topologies", func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("filterBy"), "name==chaoyu-topo")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"networkTopologies": []map[string]any{{"id": "topo-9", "name": "chaoyu-topo", "clusterId": mnpClusterUUID}},
		})
	})
	var putBody map[string]any
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
				"schedulingConfiguration": map[string]any{
					"gpuPodsPlacementStrategy": map[string]any{"node": "spread", "device": "spread"},
					"minGuaranteedRuntime":     "10m",
				},
			})
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
			})
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{NetworkTopologyName: "chaoyu-topo"} // only attaching a topology
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)

	assert.Equal(t, "topo-9", putBody["networkTopologyId"], "topology attached")
	sc, ok := putBody["schedulingConfiguration"].(map[string]any)
	require.True(t, ok, "scheduling must be re-sent, not reset")
	gpu, _ := sc["gpuPodsPlacementStrategy"].(map[string]any)
	require.NotNil(t, gpu)
	assert.Equal(t, "spread", gpu["node"], "gpu placement preserved on topology attach")
	assert.Equal(t, "spread", gpu["device"])
	assert.Equal(t, "10m", sc["minGuaranteedRuntime"], "minGuaranteedRuntime preserved on topology attach")
}

// A scheduling-only update must not detach an already-bound topology: upstream's
// full-replace PUT clears networkTopologyId if it's omitted, so the current binding
// must be carried forward when the caller doesn't change or detach it.
func TestManageNodePool_UpdatePreservesTopologyOnSchedulingChange(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	var putBody map[string]any
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
				"networkTopologyId": "topo-9", "networkTopologyName": "chaoyu-topo",
				"schedulingConfiguration": map[string]any{
					"gpuPodsPlacementStrategy": map[string]any{"node": "binpack", "device": "binpack"},
				},
			})
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
			})
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{Scheduling: &nodePoolScheduling{GPUPlacementStrategy: "spread"}} // change placement only
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)

	require.NotNil(t, putBody, "PUT must have been sent")
	assert.Equal(t, "topo-9", putBody["networkTopologyId"], "existing topology must be carried forward, not detached")
}

// Class-wide guard: an unrelated (label-only) update must carry forward EVERY managed
// sub-object so the full-replace PUT wipes nothing.
func TestManageNodePool_UpdateLabelOnlyPreservesAllManagedFields(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	var putBody map[string]any
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "old",
				"networkTopologyId": "topo-9", "networkTopologyName": "chaoyu-topo",
				"gpuNetworkAccelerationLabelKey":  "nvidia.com/gpu.clique",
				"gpuNetworkAccelerationDetection": "Use",
				"schedulingConfiguration":         map[string]any{"gpuPodsPlacementStrategy": map[string]any{"node": "spread", "device": "spread"}},
				"schedulingCapabilities":          map[string]any{"adjustableWeights": true, "ranks": true, "quota": true, "cpuQuota": false},
			})
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "new"})
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{LabelValue: "new"} // only the label changes
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)

	require.NotNil(t, putBody)
	assert.Equal(t, "new", putBody["labelValue"])
	assert.Equal(t, "topo-9", putBody["networkTopologyId"], "topology preserved")
	assert.Equal(t, "nvidia.com/gpu.clique", putBody["gpuNetworkAccelerationLabelKey"], "gpu-net-accel label preserved")
	assert.Equal(t, "Use", putBody["gpuNetworkAccelerationDetection"], "gpu-net-accel detection preserved")
	sc, ok := putBody["schedulingConfiguration"].(map[string]any)
	require.True(t, ok, "schedulingConfiguration preserved")
	assert.Equal(t, "spread", sc["gpuPodsPlacementStrategy"].(map[string]any)["node"])
	caps, ok := putBody["schedulingCapabilities"].(map[string]any)
	require.True(t, ok, "schedulingCapabilities preserved")
	assert.Equal(t, true, caps["ranks"])
}

func TestManageNodePool_UpdateDetachTopologySendsNull(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	var putBody map[string]any
	sawPut := false
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
				"networkTopologyId": "topo-9", "networkTopologyName": "chaoyu-topo",
				"schedulingConfiguration": map[string]any{
					"gpuPodsPlacementStrategy": map[string]any{"node": "spread", "device": "spread"},
				},
			})
		case http.MethodPut:
			sawPut = true
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
			})
		}
	})

	detach := true
	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{DetachNetworkTopology: &detach}
	in.Confirm = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)

	require.True(t, sawPut, "detach must reach upstream")
	v, present := putBody["networkTopologyId"]
	assert.True(t, present, "networkTopologyId key must be sent")
	assert.Nil(t, v, "detach sends an explicit null to clear the binding")
	sc, ok := putBody["schedulingConfiguration"].(map[string]any)
	require.True(t, ok, "scheduling must still be carried forward on detach")
	gpu, _ := sc["gpuPodsPlacementStrategy"].(map[string]any)
	require.NotNil(t, gpu)
	assert.Equal(t, "spread", gpu["node"], "gpu placement preserved on detach")
	assert.NotContains(t, out.After, keyNetworkTopologyName, "detach clears the topology in the result")
}

func TestManageNodePool_DetachTopologyValidation(t *testing.T) {
	detach := true
	t.Run("rejects detach with attach name", func(t *testing.T) {
		in := mnpInput("update")
		in.NodePoolID = "np-1"
		in.Spec = &nodePoolSpec{DetachNetworkTopology: &detach, NetworkTopologyName: "topo"}
		assert.ErrorContains(t, in.Validate(), "mutually exclusive")
	})
	t.Run("rejects detach on create", func(t *testing.T) {
		in := mnpInput("create")
		in.Name = "gpu-a"
		in.Spec = &nodePoolSpec{LabelKey: "k", LabelValue: "v", DetachNetworkTopology: &detach}
		assert.ErrorContains(t, in.Validate(), "only valid with action=update")
	})
}

// Capabilities/optimization are plain (non-pointer) fields upstream; the snapshot
// must surface plain values, not pointer addresses, so the diff stays meaningful.
func TestManageNodePool_SnapshotNestedFieldsArePlainValues(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{
			"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
			"schedulingCapabilities": map[string]any{"adjustableWeights": true, "ranks": false, "quota": true, "cpuQuota": false},
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, body)
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{LabelValue: "v"}
	in.Confirm = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.Equal(t, true, out.Before["schedulingCapabilities.adjustableWeights"])
	assert.Equal(t, false, out.Before["schedulingCapabilities.ranks"])
	// No spurious diff: identical before/after for unchanged nested fields.
	assert.NotContains(t, out.Diff, "schedulingCapabilities.adjustableWeights")
}

// networkTopologyName is resolved to its ID via the topologies API and sent as
// networkTopologyId in the create body.
func TestManageNodePool_ResolvesNetworkTopologyName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("create"))
	mux.HandleFunc("/api/v1/network-topologies", func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("filterBy"), "name==rail-optimized")
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"networkTopologies": []map[string]any{{"id": "topo-7", "name": "rail-optimized", "clusterId": mnpClusterUUID}},
		})
	})
	var createBody map[string]any
	mux.HandleFunc(nodePoolsPath, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&createBody)
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"id": "np-new", "name": "gpu-a", "clusterId": mnpClusterUUID})
	})

	in := mnpInput("create")
	in.Name = "gpu-a"
	in.Spec = &nodePoolSpec{LabelKey: "k", LabelValue: "v", NetworkTopologyName: "rail-optimized"}
	in.Confirm = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "topo-7", createBody["networkTopologyId"], "resolved topology ID sent in create body")
}

func TestManageNodePool_NetworkTopologyNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("create"))
	mux.HandleFunc("/api/v1/network-topologies", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"networkTopologies": []map[string]any{}})
	})

	in := mnpInput("create")
	in.Name = "gpu-a"
	in.Spec = &nodePoolSpec{LabelKey: "k", LabelValue: "v", NetworkTopologyName: "ghost"}
	_, err := manageNodePoolRun(context.Background(), client, in)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

// A name with the filter operator (==) must be rejected, not smuggled into FilterBy.
func TestManageNodePool_RejectsFilterOperatorInName(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	in := mnpInput("delete")
	in.Name = "gpu==a"
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
}

func TestManageNodePool_DeleteWithConfirm(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	var deleteCalled int32
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "legacy", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v", "isDefault": false,
			})
		case http.MethodDelete:
			atomic.AddInt32(&deleteCalled, 1)
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{})
		default:
			http.Error(w, "unexpected method", http.StatusInternalServerError)
		}
	})

	in := mnpInput("delete")
	in.NodePoolID = "np-1"
	in.Confirm = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, int32(1), atomic.LoadInt32(&deleteCalled))
	assert.Equal(t, "legacy", out.Before["name"])
	assert.Empty(t, out.After)
}

func TestManageNodePool_DeleteWithoutConfirm(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	var deleteCalled int32
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			atomic.AddInt32(&deleteCalled, 1)
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "np-1", "name": "legacy", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v", "isDefault": false,
		})
	})

	in := mnpInput("delete")
	in.NodePoolID = "np-1"
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "legacy")
	assert.Equal(t, int32(0), atomic.LoadInt32(&deleteCalled), "delete must not fire without confirm")
}

// The backend forbids deleting the default pool and rejects it with an
// undecodable error body; the tool must refuse with a clear message and
// never fire the DELETE.
func TestManageNodePool_DeleteDefaultRejected(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	var deleteCalled int32
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			atomic.AddInt32(&deleteCalled, 1)
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"id": "np-1", "name": "default", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v", "isDefault": true,
		})
	})

	in := mnpInput("delete")
	in.NodePoolID = "np-1"
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "default")
	assert.Equal(t, int32(0), atomic.LoadInt32(&deleteCalled), "delete must not fire for the default pool")
}

func TestManageNodePool_DryRunNoMutation(t *testing.T) {
	for _, action := range []string{"create", "update", "delete"} {
		t.Run(action, func(t *testing.T) {
			client, mux := runaiclienttest.NewServer(t)
			mux.HandleFunc(permScopesPath, permScopesHandler("create", "update", "delete"))
			var mutated int32
			mux.HandleFunc(nodePoolsPath, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					atomic.AddInt32(&mutated, 1)
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodepools": []map[string]any{}})
			})
			mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPut || r.Method == http.MethodDelete {
					atomic.AddInt32(&mutated, 1)
				}
				runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
					"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v", "isDefault": false,
				})
			})

			in := mnpInput(action)
			in.DryRun = true
			if action == "create" {
				in.Name = "gpu-a"
				in.Spec = &nodePoolSpec{LabelKey: "k", LabelValue: "v"}
			} else {
				in.NodePoolID = "np-1"
				in.Spec = &nodePoolSpec{LabelValue: "new"}
			}
			out, err := manageNodePoolRun(context.Background(), client, in)
			require.NoError(t, err)
			assert.False(t, out.Applied)
			assert.True(t, out.DryRun)
			assert.Equal(t, int32(0), atomic.LoadInt32(&mutated), "dry-run must not hit any mutation endpoint")
		})
	}
}

func TestManageNodePool_TenantScopedGrantAllowed(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesTenantHandler())
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelValue": "old"})
		case http.MethodPut:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelValue": "new"})
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{LabelValue: "new"}
	in.Confirm = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err, "a tenant-level grant must authorize cluster-resource writes")
	assert.True(t, out.Applied)
}

func TestManageNodePool_ForbiddenViaProbe(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("read")) // no delete grant
	var deleteCalled int32
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			atomic.AddInt32(&deleteCalled, 1)
		}
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"id": "np-1"})
	})

	in := mnpInput("delete")
	in.NodePoolID = "np-1"
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	assert.ErrorIs(t, err, runaimcp.ErrForbidden)
	assert.Equal(t, int32(0), atomic.LoadInt32(&deleteCalled), "mutation must not fire when RBAC probe denies")
}

func TestManageNodePool_DeleteNotFoundIgnoreMissing(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "gone"})
	})

	in := mnpInput("delete")
	in.NodePoolID = "np-1"
	in.Confirm = true
	in.IgnoreMissing = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.False(t, out.Applied)
	assert.NotEmpty(t, out.Warnings)
}

func TestManageNodePool_DeleteNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusNotFound, map[string]any{"message": "gone"})
	})

	in := mnpInput("delete")
	in.NodePoolID = "np-1"
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	assert.ErrorIs(t, err, runaimcp.ErrNotFound)
}

func TestManageNodePool_NameAmbiguous(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(nodePoolsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodepools": []map[string]any{{"id": "np-1", "name": "dup"}, {"id": "np-2", "name": "dup"}},
		})
	})

	in := mnpInput("delete")
	in.Name = "dup"
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.Error(t, err)
	assert.ErrorIs(t, err, runaimcp.ErrValidation)
	assert.Contains(t, err.Error(), "ambiguous_node_pool")
}

func TestManageNodePool_NameResolvesAndDeletes(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(nodePoolsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
			"nodepools": []map[string]any{{"id": "np-7", "name": "gpu-a", "clusterId": mnpClusterUUID}},
		})
	})
	mux.HandleFunc(nodePoolsPath+"/np-7", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"id": "np-7", "name": "gpu-a", "clusterId": mnpClusterUUID})
		case http.MethodDelete:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{})
		}
	})

	in := mnpInput("delete")
	in.Name = "gpu-a"
	in.Confirm = true
	out, err := manageNodePoolRun(context.Background(), client, in)
	require.NoError(t, err)
	assert.True(t, out.Applied)
	assert.Equal(t, "np-7", out.NodePoolID)
}

func TestManageNodePool_NameNotFound(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("delete"))
	mux.HandleFunc(nodePoolsPath, func(w http.ResponseWriter, _ *http.Request) {
		runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{"nodepools": []map[string]any{}})
	})

	in := mnpInput("delete")
	in.Name = "ghost"
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	assert.True(t, errors.Is(err, runaimcp.ErrNotFound))
}

func TestManageNodePoolToolsLint(t *testing.T) {
	errs := runaimcp.Lint(func(r *runaimcp.Registry) {
		Register(r, nil)
	})
	if len(errs) > 0 {
		t.Fatalf("tool lint errors: %v", errs)
	}
}

// The real generated client fails to decode a bare-string error body into its Error model;
// MapHTTPError must surface that raw body instead of the "cannot unmarshal string" noise.
func TestManageNodePool_UpdateSurfacesRawUpstreamErrorBody(t *testing.T) {
	client, mux := runaiclienttest.NewServer(t)
	mux.HandleFunc(permScopesPath, permScopesHandler("update"))
	mux.HandleFunc(nodePoolsPath+"/np-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			runaiclienttest.WriteJSON(w, http.StatusOK, map[string]any{
				"id": "np-1", "name": "gpu-a", "clusterId": mnpClusterUUID, "labelKey": "k", "labelValue": "v",
			})
		case http.MethodPut:
			runaiclienttest.WriteJSON(w, http.StatusBadRequest,
				"can not enable node level scheduling while gpu-over-provisioning flag is disabled")
		}
	})

	in := mnpInput("update")
	in.NodePoolID = "np-1"
	in.Spec = &nodePoolSpec{LabelValue: "new"}
	in.Confirm = true
	_, err := manageNodePoolRun(context.Background(), client, in)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gpu-over-provisioning", "the real upstream reason must be surfaced")
	assert.NotContains(t, err.Error(), "unmarshal", "the decode-failure noise must not leak")
}
