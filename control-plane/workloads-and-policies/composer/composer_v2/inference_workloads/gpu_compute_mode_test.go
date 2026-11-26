/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_workloads

import (
	"context"
	"encoding/json"

	clusterapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
	"knative.dev/pkg/apis"
	"knative.dev/serving/pkg/apis/config"
	"knative.dev/serving/pkg/apis/serving"
)

var _ = Describe("GPU mode annotation projection", func() {
	It("targets the GPU container and preserves the memory request", func() {
		const mode = "sm-sharing"
		options := composer.FieldsComposerOptions{GpuComputeModeSupported: true}
		var src workloadsApi.WorkspaceSpecSpecAllOfCompute
		Expect(json.Unmarshal([]byte(`{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`), &src)).To(Succeed())
		c := NewInferenceToKnativeComposer("example", "p1", "runai-p1", "user", nil, options)
		target := NewMinimalKnativeServiceTarget()
		target.Service.Spec.Template.Spec.Containers = []corev1.Container{{Name: "inference-main"}, {Name: "auth-proxy"}}
		Expect(c.processComputeFields(context.Background(), &src, nil, target)()).To(BeNil())
		Expect(target.Service.Spec.Template.Annotations).To(HaveKeyWithValue("nvidia.com/container.inference-main.gpu-compute.mode", mode))
		Expect(target.Service.Spec.Template.Annotations).NotTo(HaveKey("nvidia.com/container.auth-proxy.gpu-compute.mode"))
		Expect(target.Service.Spec.Template.Annotations).To(HaveKeyWithValue(composer.GpuFractionAnnotation, "0.25"))
	})
})

var _ = Describe("GPU selection in Knative inference", func() {
	It("rejects a duplicate GPU count in extended resources for a model-only request", func() {
		var spec workloadsApi.InferenceSpecSpec
		Expect(json.Unmarshal([]byte(`{"gpuType":"H100","compute":{"gpuDevicesRequest":2,
			"extendedResources":[{"resource":"nvidia.com/gpu","quantity":"1"}]}}`), &spec)).To(Succeed())
		c := NewInferenceToKnativeComposer("infer", "p1", "runai-p1", "user", nil, composer.FieldsComposerOptions{})
		target := NewMinimalKnativeServiceTarget()
		err := c.processComputeFields(context.Background(), spec.Compute.Get(),
			composer.RequestedAccelerator(spec.GpuType.Get(), spec.Accelerator.Get()), target)()
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(400)))
		Expect(err.Message).To(ContainSubstring("duplicates"))
	})

	It("admits the Pod with nodeSelector disabled and requires the resource, pool and model", func() {
		options := composer.FieldsComposerOptions{NodeAffinitySupported: true}
		c := NewInferenceToKnativeComposer("infer", "p1", "runai-p1", "user", nil, options)
		var spec workloadsApi.InferenceSpecSpec
		Expect(json.Unmarshal([]byte(`{
			"gpuType":"MI300X","accelerator":{"resourceName":"amd.com/gpu","deviceLabelKey":"hardware.example.com/model"},
			"nodePools":["default","research"]
		}`), &spec)).To(Succeed())
		c.ActualNodePools = []clusterapi.Nodepool{{Name: "research", LabelKey: composer.NodePoolNameLabel, LabelValue: "research"}}
		target := NewMinimalKnativeServiceTarget()
		template := &target.Service.Spec.Template
		template.Spec.Containers = []corev1.Container{{Name: "main", Image: "example.com/test:latest"}}
		Expect(c.processPlacementFields(context.Background(), &spec, target)()).To(BeNil())
		cfg := config.FromContextOrDefaults(context.Background())
		cfg.Features.PodSpecAffinity = config.Enabled
		Expect(cfg.Features.PodSpecNodeSelector).To(Equal(config.Disabled))
		ctx := config.ToContext(context.Background(), cfg)
		Expect(serving.ValidatePodSpec(ctx, template.Spec.PodSpec).Filter(apis.ErrorLevel)).To(BeNil())
		Expect(template.Spec.NodeSelector).To(BeEmpty())
		Expect(template.Annotations).To(HaveKeyWithValue("runai-nodepools", "default research"))
		pod := &corev1.Pod{Spec: template.Spec.PodSpec}
		terms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		Expect(terms).To(HaveLen(2))
		for _, term := range terms {
			Expect(term.MatchFields).To(BeEmpty())
		}
		resourceKey := commonconstants.AcceleratorResourceNodeLabel("amd.com/gpu")
		for _, tc := range []struct {
			labels  map[string]string
			matches bool
		}{
			{map[string]string{resourceKey: "true", "hardware.example.com/model": "MI300X", composer.NodePoolNameLabel: "research"}, true},
			{map[string]string{resourceKey: "true", "hardware.example.com/model": "MI250", composer.NodePoolNameLabel: "research"}, false},
			{map[string]string{"hardware.example.com/model": "MI300X", composer.NodePoolNameLabel: "research"}, false},
			{map[string]string{resourceKey: "true", "hardware.example.com/model": "MI300X"}, true},
			{map[string]string{"hardware.example.com/model": "MI300X"}, false},
			{map[string]string{resourceKey: "true", "hardware.example.com/model": "MI250"}, false},
			{map[string]string{resourceKey: "true", "hardware.example.com/model": "MI300X", composer.NodePoolNameLabel: "other"}, false},
		} {
			node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "gpu-node", Labels: tc.labels}}
			matches, err := nodeaffinity.GetRequiredNodeAffinity(pod).Match(node)
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(Equal(tc.matches))
		}
	})
})
