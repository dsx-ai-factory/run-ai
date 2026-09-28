/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
)

func jobKartaForNodePoolsTest() *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
		StructureDefinition: kartav1alpha1.StructureDefinition{
			RootComponent: kartav1alpha1.ComponentDefinition{
				Name:           "job",
				Kind:           &kartav1alpha1.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"},
				SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: convutils.Ref(".spec.template")},
			},
		},
	}}
}

func jobManifestForNodePoolsTest(podMetadata map[string]interface{}, podSpec map[string]interface{}) map[string]interface{} {
	podSpec["containers"] = []interface{}{map[string]interface{}{"name": "main", "image": "alpine"}}
	return map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata":   map[string]interface{}{"name": "train"},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{"metadata": podMetadata, "spec": podSpec},
		},
	}
}

func requiredNodeAffinityForNodePoolsTest(key string, values ...interface{}) map[string]interface{} {
	return map[string]interface{}{"affinity": map[string]interface{}{"nodeAffinity": map[string]interface{}{
		"requiredDuringSchedulingIgnoredDuringExecution": map[string]interface{}{
			"nodeSelectorTerms": []interface{}{map[string]interface{}{
				"matchExpressions": []interface{}{map[string]interface{}{"key": key, "operator": "In", "values": values}},
			}},
		},
	}}}
}

var _ = Describe("requestedWorkloadNodePools", func() {
	const nodePoolLabelKey = "bug-ng"
	var (
		ctx             context.Context
		karta           *kartav1alpha1.Karta
		actualNodePools []clusterApi.Nodepool
	)

	BeforeEach(func() {
		ctx = context.Background()
		karta = jobKartaForNodePoolsTest()
		actualNodePools = []clusterApi.Nodepool{
			{Name: "default"},
			{Name: "pool-a", LabelKey: nodePoolLabelKey, LabelValue: "poolA"},
			{Name: "pool-b", LabelKey: nodePoolLabelKey, LabelValue: "poolB"},
		}
	})

	Context("pod selects node pools", func() {
		It("should resolve the pool from required node affinity", func() {
			manifest := jobManifestForNodePoolsTest(map[string]interface{}{}, requiredNodeAffinityForNodePoolsTest(nodePoolLabelKey, "poolB"))

			Expect(requestedWorkloadNodePools(ctx, karta, manifest, actualNodePools)).To(Equal([]string{"pool-b"}))
		})

		It("should prefer the node-pool label over affinity", func() {
			manifest := jobManifestForNodePoolsTest(
				map[string]interface{}{"labels": map[string]interface{}{clusterconstants.NodePoolNameLabel: "pool-a"}},
				requiredNodeAffinityForNodePoolsTest(nodePoolLabelKey, "poolB"),
			)

			Expect(requestedWorkloadNodePools(ctx, karta, manifest, actualNodePools)).To(Equal([]string{"pool-a"}))
		})

		It("should prefer the node-pools annotation over the label", func() {
			manifest := jobManifestForNodePoolsTest(map[string]interface{}{
				"labels":      map[string]interface{}{clusterconstants.NodePoolNameLabel: "pool-a"},
				"annotations": map[string]interface{}{commonconstants.AnnotationNodepools: "pool-b default"},
			}, map[string]interface{}{})

			Expect(requestedWorkloadNodePools(ctx, karta, manifest, actualNodePools)).To(Equal([]string{"pool-b", "default"}))
		})
	})

	Context("no node pools can be read", func() {
		It("should ignore unknown pools", func() {
			manifest := jobManifestForNodePoolsTest(
				map[string]interface{}{"labels": map[string]interface{}{clusterconstants.NodePoolNameLabel: unexistingNodePoolLabelValue}},
				requiredNodeAffinityForNodePoolsTest(nodePoolLabelKey, "unknown"),
			)

			Expect(requestedWorkloadNodePools(ctx, karta, manifest, actualNodePools)).To(BeNil())
		})

		It("should return nil when the pod does not select a pool", func() {
			manifest := jobManifestForNodePoolsTest(map[string]interface{}{}, map[string]interface{}{})

			Expect(requestedWorkloadNodePools(ctx, karta, manifest, actualNodePools)).To(BeNil())
		})

		It("should return nil without a Karta", func() {
			manifest := jobManifestForNodePoolsTest(map[string]interface{}{}, requiredNodeAffinityForNodePoolsTest(nodePoolLabelKey, "poolB"))

			Expect(requestedWorkloadNodePools(ctx, nil, manifest, actualNodePools)).To(BeNil())
		})
	})
})
