/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_submitter

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	nodepools "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/node-pools"
	workloadclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/utils/ptr"
)

var kwokStageGVR = schema.GroupVersionResource{Group: "kwok.x-k8s.io", Version: "v1alpha1", Resource: "stages"}

// GetFirstClusterID returns the UUID of the first cluster visible to the authenticated user.
func GetFirstClusterID(ctx context.Context, clients *Clients) (string, error) {
	clusters, _, err := clients.ClusterServiceClient.ClustersAPI.GetClusters(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get clusters: %w", err)
	}
	if len(clusters) == 0 {
		return "", fmt.Errorf("no clusters found")
	}
	return clusters[0].Uuid, nil
}

// ResolveProjectIDs returns the IDs of all projects whose names start with projectPrefix in clusterID.
func ResolveProjectIDs(ctx context.Context, clients *Clients, clusterID, projectPrefix string) ([]string, error) {
	clusterFilter := fmt.Sprintf("clusterId==%s", clusterID)
	nameFilter := fmt.Sprintf("name=@%s", projectPrefix)

	projects, _, err := clients.OrgUnitServiceClient.ProjectsAPI.
		GetProjects(ctx).
		FilterBy([]string{clusterFilter, nameFilter}).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	if len(projects.Projects) == 0 {
		return nil, fmt.Errorf("no projects found with prefix %q in cluster %s", projectPrefix, clusterID)
	}

	ids := make([]string, len(projects.Projects))
	for i, p := range projects.Projects {
		ids[i] = p.Id
	}
	return ids, nil
}

// ResolveNodepoolNames fetches nodepool names matching prefix, sorted for determinism.
func ResolveNodepoolNames(ctx context.Context, clients *Clients, prefix string) ([]string, error) {
	nps, err := nodepools.FetchNodepools(ctx, clients.ClusterServiceClient.NodePoolsAPI)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, np := range nps {
		if strings.HasPrefix(np.Name, prefix) {
			names = append(names, np.Name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// deleteCompletionStages removes any Stage that would complete load-test pods before
// the intended D-second delay:
//   - "pod-complete"        — installed by kwok/stage-fast; completes Job pods immediately
//   - "dynamic-pod-sleep*"  — any stage from a previous submitter run with a different D
func deleteCompletionStages(ctx context.Context, dc dynamic.Interface) error {
	toDelete := []string{"pod-complete"}

	existing, err := dc.Resource(kwokStageGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list stages: %w", err)
	}
	for _, s := range existing.Items {
		name := s.GetName()
		if strings.HasPrefix(name, "dynamic-pod-sleep") {
			toDelete = append(toDelete, name)
		}
	}

	for _, name := range toDelete {
		err := dc.Resource(kwokStageGVR).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete stage %q: %w", name, err)
		}
	}
	return nil
}

// ApplyKwokStage sets up the KWOK pod lifecycle for load testing:
//
//  1. Deletes any Stage that would complete load-test pods — specifically "pod-complete"
//     (installed by the kwok/stage-fast chart) and any previously created "dynamic-pod-sleepN"
//     stage from a prior run. This makes the function safe to call on env reuse with a
//     different D value.
//
//  2. Creates a cluster-scoped Stage named "dynamic-pod-sleepN" that fires on any Running pod
//     with a "workloadKind" label (set by RunAI on Training/Distributed pods). After dSecs
//     seconds the stage patches the pod status to Succeeded, simulating workload completion
//     after D = W_cap/R seconds — the cycle duration for the load test.
func ApplyKwokStage(ctx context.Context, clients *Clients, dSecs int) error {
	if err := deleteCompletionStages(ctx, clients.DynamicClient); err != nil {
		return err
	}
	name := fmt.Sprintf("dynamic-pod-sleep%d", dSecs)

	stage := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kwok.x-k8s.io/v1alpha1",
			"kind":       "Stage",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"resourceRef": map[string]interface{}{
					"apiGroup": "v1",
					"kind":     "Pod",
				},
				"delay": map[string]interface{}{
					"durationMilliseconds": dSecs * 1000,
				},
				"selector": map[string]interface{}{
					"matchExpressions": []map[string]interface{}{
						{
							"key":      ".status.phase",
							"operator": "In",
							"values":   []string{"Running"},
						},
						{
							"key":      ".metadata.labels.workloadKind",
							"operator": "In",
							"values":   []string{"Training", "Distributed"},
						},
					},
				},
				// "next" defines the pod status patch KWOK applies when the stage fires.
				// The Go template is evaluated by KWOK at fire time.
				// $root captures the full pod object so we can read the existing container
				// statuses (e.g. the actual startedAt time from $origin.state.running.startedAt).
				// Each container is written as terminated with exitCode=0 and the pod phase
				// is set to Succeeded, simulating a clean workload completion.
				"next": map[string]interface{}{
					"statusTemplate": `{{ $now := Now }}
{{ $root := . }}
containerStatuses:
{{ range $index, $item := .spec.containers }}
{{ $origin := index $root.status.containerStatuses $index }}
- image: {{ $item.image | Quote }}
  name: {{ $item.name | Quote }}
  ready: false
  restartCount: 0
  started: false
  state:
    terminated:
      exitCode: 0
      finishedAt: {{ $now | Quote }}
      reason: Completed
      startedAt: {{ $origin.state.running.startedAt | Quote }}
{{ end }}
phase: Succeeded
`,
				},
			},
		},
	}
	return applyUnstructured(ctx, clients.DynamicClient, kwokStageGVR, stage)
}

// applyUnstructured issues a server-side apply for obj. Force=true so this process takes
// ownership of any conflicting managed fields (safe for test infra that may re-apply).
// Stage is cluster-scoped so obj.GetNamespace() returns "" and the dynamic client builds
// a cluster-scoped URL path automatically.
func applyUnstructured(ctx context.Context, dc dynamic.Interface, gvr schema.GroupVersionResource, obj *unstructured.Unstructured) error {
	_, err := dc.Resource(gvr).Namespace(obj.GetNamespace()).Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{FieldManager: "workload-submitter", Force: true})
	return err
}

type submitParams struct {
	Name            string
	ProjectID       string
	ClusterID       string
	NodepoolName    string // "" = no nodepool preference; KWOK toleration still applied
	DSecs           int
	PodsPerWorkload int
	AutoDeleteSecs  int // seconds after completion before the workload is deleted; 0 = never
}

// SubmitOne creates one workload. PodsPerWorkload==1 produces a Training workload;
// PodsPerWorkload>1 produces a Distributed (PyTorch) workload with that many workers.
// Returns the workload ID, the raw HTTP response, and any error.
func SubmitOne(ctx context.Context, clients *Clients, p submitParams) (string, *http.Response, error) {
	if p.PodsPerWorkload > 1 {
		return submitDistributed(ctx, clients, p)
	}
	return submitTraining(ctx, clients, p)
}

func submitTraining(ctx context.Context, clients *Clients, p submitParams) (string, *http.Response, error) {
	command := fmt.Sprintf("sleep %d", p.DSecs)

	req := workloadclient.TrainingCreationRequest{
		Name:      p.Name,
		ProjectId: p.ProjectID,
		ClusterId: p.ClusterID,
		Spec: &workloadclient.TrainingSpecSpec{
			Command: *workloadclient.NewNullableString(ptr.To(command)),
			// Image is never pulled — KWOK fakes the entire pod lifecycle.
			Image:           *workloadclient.NewNullableString(ptr.To("ubuntu:22.04")),
			ImagePullPolicy: *workloadclient.NewNullableImagePullPolicy(ptr.To(workloadclient.IMAGEPULLPOLICY_IF_NOT_PRESENT)),
			// GpuDevicesRequest is what the run:ai scheduler reads to count GPU consumption.
			// Without it the workload requests 0 GPUs and never counts against W_cap.
			// CPU/memory are set to minimal values to pass validation; KWOK nodes have no
			// real resource constraints so these values do not affect scheduling.
			Compute: *workloadclient.NewNullableWorkspaceSpecSpecAllOfCompute(&workloadclient.WorkspaceSpecSpecAllOfCompute{
				CpuMemoryRequest:  *workloadclient.NewNullableString(ptr.To("1M")),
				GpuDevicesRequest: *workloadclient.NewNullableInt32(ptr.To(int32(1))),
				CpuCoreRequest:    *workloadclient.NewNullableFloat64(ptr.To(float64(0.05))),
			}),
		},
	}
	req.Spec.NodePools, req.Spec.Tolerations = nodepoolConfig(p.NodepoolName)
	if p.AutoDeleteSecs > 0 {
		req.Spec.AutoDeletionTimeAfterCompletionSeconds = *workloadclient.NewNullableInt64(ptr.To(int64(p.AutoDeleteSecs)))
	}

	tw, resp, err := clients.WorkloadServiceClient.TrainingsAPI.CreateTraining(ctx).TrainingCreationRequest(req).Execute()
	if err != nil {
		return "", resp, err
	}
	return tw.WorkloadId, resp, nil
}

func submitDistributed(ctx context.Context, clients *Clients, p submitParams) (string, *http.Response, error) {
	command := fmt.Sprintf("sleep %d", p.DSecs)

	req := workloadclient.DistributedCreationRequest{
		Name:      p.Name,
		ProjectId: p.ProjectID,
		ClusterId: p.ClusterID,
		Spec: &workloadclient.DistributedSpecSpec{
			NumWorkers:           *workloadclient.NewNullableInt32(ptr.To(int32(p.PodsPerWorkload))),
			DistributedFramework: *workloadclient.NewNullableDistributedFramework(ptr.To(workloadclient.DISTRIBUTEDFRAMEWORK_PY_TORCH)),
			Command:              *workloadclient.NewNullableString(ptr.To(command)),
			// Image is never pulled — KWOK fakes the entire pod lifecycle.
			Image:           *workloadclient.NewNullableString(ptr.To("ubuntu:22.04")),
			ImagePullPolicy: *workloadclient.NewNullableImagePullPolicy(ptr.To(workloadclient.IMAGEPULLPOLICY_IF_NOT_PRESENT)),
			// GpuDevicesRequest=1 per worker; total GPU consumption = PodsPerWorkload GPUs.
			// CPU/memory are minimal placeholders; KWOK nodes have no real resource constraints.
			Compute: *workloadclient.NewNullableWorkspaceSpecSpecAllOfCompute(&workloadclient.WorkspaceSpecSpecAllOfCompute{
				CpuMemoryRequest:  *workloadclient.NewNullableString(ptr.To("1M")),
				GpuDevicesRequest: *workloadclient.NewNullableInt32(ptr.To(int32(1))),
				CpuCoreRequest:    *workloadclient.NewNullableFloat64(ptr.To(float64(0.05))),
			}),
		},
	}
	req.Spec.NodePools, req.Spec.Tolerations = nodepoolConfig(p.NodepoolName)
	if p.AutoDeleteSecs > 0 {
		req.Spec.AutoDeletionTimeAfterCompletionSeconds = *workloadclient.NewNullableInt64(ptr.To(int64(p.AutoDeleteSecs)))
	}

	dw, resp, err := clients.WorkloadServiceClient.DistributedAPI.CreateDistributed(ctx).DistributedCreationRequest(req).Execute()
	if err != nil {
		return "", resp, err
	}
	return dw.WorkloadId, resp, nil
}

// nodepoolConfig returns the nodepool list and KWOK toleration for a workload.
// The KWOK toleration (kwok.x-k8s.io/node=fake:NoSchedule) is always included so pods
// are accepted by KWOK virtual nodes. nodepoolName="" means no nodepool preference.
func nodepoolConfig(nodepoolName string) (nodepools []string, tolerations []workloadclient.Toleration) {
	if nodepoolName != "" {
		nodepools = []string{nodepoolName}
	}
	tolerations = []workloadclient.Toleration{
		{
			Key:      *workloadclient.NewNullableString(ptr.To("kwok.x-k8s.io/node")),
			Value:    *workloadclient.NewNullableString(ptr.To("fake")),
			Effect:   *workloadclient.NewNullableTolerationEffect(ptr.To(workloadclient.TOLERATIONEFFECT_NO_SCHEDULE)),
			Operator: *workloadclient.NewNullableTolerationOperator(ptr.To(workloadclient.TOLERATIONOPERATOR_EQUAL)),
		},
	}
	return
}
