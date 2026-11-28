/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dinf_to_lws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GPU mode annotation projection", func() {
	It("targets the GPU container and preserves the memory request", func() {
		const mode = "sm-sharing"
		options := composer.FieldsComposerOptions{GpuComputeModeSupported: true}
		var src workloadsApi.WorkspaceSpecSpecAllOfCompute
		Expect(json.Unmarshal([]byte(`{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`), &src)).To(Succeed())
		c := NewDistributedInferenceToLWSComposer("example", "p1", "runai-p1", "user", nil, nil, options, nil)
		for _, name := range []string{"leader-main", "worker-main"} {
			target := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: name}}}}
			Expect(c.processComputeFields(context.Background(), &src, nil, target)()).To(BeNil())
			Expect(target.Annotations).To(HaveKeyWithValue(fmt.Sprintf("nvidia.com/container.%s.gpu-compute.mode", name), mode))
			Expect(target.Annotations).To(HaveKeyWithValue(composer.GpuFractionAnnotation, "0.25"))
		}
	})
})

var _ = Describe("GPU selection in distributed inference", func() {
	It("keeps separate resource names and models for the leader and worker", func() {
		var spec workloadsApi.DistributedInferenceSpecSpec
		Expect(json.Unmarshal([]byte(`{
			"servingPort":{"port":8080,"exposeExternally":false},"workers":1,
			"leader":{"image":"leader:latest","compute":{"gpuDevicesRequest":1},"gpuType":"H100"},
			"worker":{"image":"worker:latest","compute":{"gpuDevicesRequest":2},"gpuType":"Model-X",
					  "accelerator":{"resourceName":"vendor.example/device","deviceLabelKey":"hardware.example.com/model"}}
		}`), &spec)).To(Succeed())
		options := composer.FieldsComposerOptions{NodeAffinitySupported: true, WorkloadV2NetworkingSupported: true}
		c := NewDistributedInferenceToLWSComposer("example", "team", "runai-team", "user", nil, nil, options, nil).
			WithClusterStorageConfig(legacyTypedTestClusterConfig()).WithLeaderWorkerSetName("example")
		target := NewMinimalLeaderWorkerSetTarget()
		Expect(c.Compose(context.Background(), &spec, target)()).To(BeNil())
		templates := target.LeaderWorkerSet.Spec.LeaderWorkerTemplate
		for _, tc := range []struct {
			pod                  *corev1.PodTemplateSpec
			key, model, resource string
			count                int64
		}{
			{templates.LeaderTemplate, composer.NvidiaProductLabel, "H100", composer.NvidiaResourceName, 1},
			{&templates.WorkerTemplate, "hardware.example.com/model", "Model-X", "vendor.example/device", 2},
		} {
			Expect(tc.pod.Spec.NodeSelector).To(BeEmpty())
			terms := tc.pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms).To(HaveLen(1))
			Expect(terms[0].MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
				Key: tc.key, Operator: corev1.NodeSelectorOpIn, Values: []string{tc.model},
			}))
			limits := tc.pod.Spec.Containers[0].Resources.Limits
			quantity := limits[corev1.ResourceName(tc.resource)]
			Expect(quantity.Value()).To(Equal(tc.count))
		}
	})
})
