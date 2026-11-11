/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Stamped with or without Snapshot, so installing or removing it never rolls a workload.
func TestMutateResourceStampsAnEligibleWorkload(t *testing.T) {
	for _, installed := range []bool{true, false} {
		t.Run(fmt.Sprintf("snapshot installed=%t", installed), func(t *testing.T) {
			workload := deployment(t, optIn, gpuContainer("main"))
			want := withPodLabel(t, withPodLabel(t, workload.DeepCopy(),
				constants.SnapshotOptInLabel, "true"), commonconstants.WorkloadRunnerNameLabel, "llama")

			// Every reconcile stamps again; the second time must change nothing.
			for range 2 {
				mutate(t, newPlugin(kartaStub{}, installed), workload, deploymentKarta())
			}

			if !reflect.DeepEqual(workload.Object, want.Object) {
				t.Errorf("workload = %v\nwant %v", workload.Object, want.Object)
			}
		})
	}
}

func TestMutateResourceLeavesOtherWorkloadsUnchanged(t *testing.T) {
	gpu := gpuContainer("main")

	tests := []struct {
		name     string
		karta    *kartav1alpha1.Karta
		workload *unstructured.Unstructured
	}{
		{
			name:     "not opted in",
			karta:    deploymentKarta(),
			workload: deployment(t, nil, gpu),
		},
		{
			name:     "ineligible",
			karta:    deploymentKarta(),
			workload: deployment(t, optIn, gpu, gpuContainer("sidecar")),
		},
		{
			name:     "two pod templates",
			karta:    pairKarta(),
			workload: pair(t, optIn, gpu),
		},
		{
			name:     "pods cannot carry labels",
			karta:    bareKarta(),
			workload: bare(t, optIn, gpu),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.workload.DeepCopy()

			mutate(t, newPlugin(kartaStub{}, true), tc.workload, tc.karta)

			if !reflect.DeepEqual(tc.workload.Object, want.Object) {
				t.Errorf("workload = %v\nwant it unchanged: %v", tc.workload.Object, want.Object)
			}
		})
	}
}

func mutate(t *testing.T, plugin *Plugin, workload *unstructured.Unstructured, karta *kartav1alpha1.Karta) {
	t.Helper()
	state := reconcile.New()
	if err := plugin.MutateResource(context.Background(), runnerFor(t, workload), workload, karta, state); err != nil {
		t.Fatalf("MutateResource: %v", err)
	}
	if issues := state.PluginIssues(plugin); len(issues) != 0 {
		t.Fatalf("issues = %v, want none", issues)
	}
}
