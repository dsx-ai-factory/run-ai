/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iw_tw_workloads

import (
	"context"
	"encoding/json"

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
		for _, kind := range []string{composer.Workspace, composer.Training} {
			c := NewInterTrainJobComposer(kind, "example", "p1", "runai-p1", "user", nil, options, nil)
			target := NewMinimalBatchJobTarget()
			target.Job.Spec.Template.Spec.Containers[0] = corev1.Container{Name: "workload-main"}
			Expect(c.processComputeFields(context.Background(), &src, nil, target)()).To(BeNil())
			Expect(target.Job.Spec.Template.Annotations).To(HaveKeyWithValue("nvidia.com/container.workload-main.gpu-compute.mode", mode))
			Expect(target.Job.Spec.Template.Annotations).To(HaveKeyWithValue(composer.GpuFractionAnnotation, "0.25"))
		}
	})
})

var _ = Describe("GPU selection in submitted jobs", func() {
	It("keeps the selected resource and GPU model in the final Pod template", func() {
		var spec workloadsApi.SupersetSpec
		Expect(json.Unmarshal([]byte(`{
			"image":"ubuntu", "compute":{"gpuDevicesRequest":2},
			"accelerator":{"resourceName":"amd.com/gpu","deviceLabelKey":"hardware.example.com/model"},
			"gpuType":"MI300X"
		}`), &spec)).To(Succeed())
		for _, kind := range []string{composer.Training, composer.Workspace} {
			options := composer.FieldsComposerOptions{NodeAffinitySupported: true}
			c := NewInterTrainJobComposer(kind, "example", "team", "team-namespace", "user", nil, options, nil)
			c.WithClusterStorageConfig(testClusterStorageConfig("https://example.com", false))
			target := NewMinimalBatchJobTarget()
			Expect(c.Compose(context.Background(), &spec, target)()).To(BeNil())
			template := target.Job.Spec.Template
			resources := template.Spec.Containers[0].Resources
			for _, list := range []corev1.ResourceList{resources.Requests, resources.Limits} {
				Expect(list).To(HaveKey(corev1.ResourceName("amd.com/gpu")))
				Expect(list).NotTo(HaveKey(corev1.ResourceName("nvidia.com/gpu")))
				quantity := list[corev1.ResourceName("amd.com/gpu")]
				Expect(quantity.Value()).To(Equal(int64(2)))
			}
			Expect(template.Spec.NodeSelector).To(BeEmpty())
			terms := template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms).To(HaveLen(1))
			Expect(terms[0].MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
				Key: "hardware.example.com/model", Operator: corev1.NodeSelectorOpIn, Values: []string{"MI300X"},
			}))
		}
	})
})
