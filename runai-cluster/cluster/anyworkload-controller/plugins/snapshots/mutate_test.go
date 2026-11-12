/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("MutateResource", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	gpu := gpuContainer("main")

	// Stamped with or without Snapshot, so installing or removing it never rolls a workload.
	DescribeTable("stamps an eligible workload",
		func(installed bool) {
			workload := deployment(optIn, gpu)
			want := withPodLabel(withPodLabel(withPodLabel(workload.DeepCopy(),
				constants.SnapshotOptInLabel, "true"), commonconstants.WorkloadRunnerNameLabel, "llama"),
				commonconstants.WorkloadRunnerUidLabel, "runner-uid")

			// Every reconcile stamps again; the second time must change nothing.
			for range 2 {
				mutate(ctx, newPlugin(kartaStub{}, installed), workload, deploymentKarta())
			}

			Expect(workload.Object).To(Equal(want.Object))
		},
		Entry("with Snapshot installed", true),
		Entry("without Snapshot", false),
	)

	DescribeTable("leaves other workloads unchanged",
		func(karta *kartav1alpha1.Karta, workload func() *unstructured.Unstructured) {
			got := workload()
			want := got.DeepCopy()

			mutate(ctx, newPlugin(kartaStub{}, true), got, karta)

			Expect(got.Object).To(Equal(want.Object))
		},
		Entry("not opted in", deploymentKarta(),
			func() *unstructured.Unstructured { return deployment(nil, gpu) }),
		Entry("ineligible", deploymentKarta(),
			func() *unstructured.Unstructured { return deployment(optIn, gpu, gpuContainer("sidecar")) }),
		Entry("two pod templates", pairKarta(),
			func() *unstructured.Unstructured { return pair(optIn, gpu) }),
		Entry("pods cannot carry labels", bareKarta(),
			func() *unstructured.Unstructured { return bare(optIn, gpu) }),
	)
})

func mutate(ctx context.Context, plugin *Plugin, workload *unstructured.Unstructured, karta *kartav1alpha1.Karta) {
	state := reconcile.New()
	Expect(plugin.MutateResource(ctx, runnerFor(workload), workload, karta, state)).To(Succeed())
	Expect(state.PluginIssues(plugin)).To(BeEmpty())
}
