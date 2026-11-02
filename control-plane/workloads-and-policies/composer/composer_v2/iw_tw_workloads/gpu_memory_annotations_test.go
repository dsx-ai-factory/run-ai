/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iw_tw_workloads

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	kairesources "github.com/kai-scheduler/KAI-scheduler/pkg/common/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("GPU memory annotation compatibility", func() {
	DescribeTable("preserves quantities accepted by KAI", func(request, limit string) {
		compute := map[string]string{"gpuMemoryRequest": request}
		if limit != "" {
			compute["gpuMemoryLimit"] = limit
		}
		input, err := json.Marshal(compute)
		Expect(err).NotTo(HaveOccurred())
		var src workloadsApi.WorkspaceSpecSpecAllOfCompute
		Expect(json.Unmarshal(input, &src)).To(Succeed())

		for _, kind := range []string{composer.Workspace, composer.Training} {
			options := composer.FieldsComposerOptions{GpuMemoryLimitAnnotationsSupported: true, GpuMemoryRequestAnnotationsSupported: true}
			c := NewInterTrainJobComposer(kind, "example", "p1", "runai-p1", "user", nil, options, nil)
			target := NewMinimalBatchJobTarget()
			target.Job.Spec.Template.Spec.Containers[0] = corev1.Container{Name: "worker"}
			target.Job.Spec.Template.Annotations = map[string]string{
				composer.GpuMemoryAnnotation:   "524",
				composer.GpuFractionAnnotation: "0.75",
				"example.com/preserved":        "value",
			}
			Expect(c.processComputeFields(context.Background(), &src, target)()).To(BeNil())
			pod := &corev1.Pod{ObjectMeta: target.Job.Spec.Template.ObjectMeta, Spec: target.Job.Spec.Template.Spec}
			Expect(pod.Annotations).NotTo(HaveKey(composer.GpuMemoryAnnotation))
			Expect(pod.Annotations).NotTo(HaveKey(composer.GpuFractionAnnotation))
			Expect(pod.Annotations).To(HaveKeyWithValue("example.com/preserved", "value"))
			Expect(pod.Annotations).To(HaveKeyWithValue("nvidia.com/container.worker.gpu-memory.request", request))
			if limit != "" {
				Expect(pod.Annotations).To(HaveKeyWithValue("nvidia.com/container.worker.gpu-memory.limit", limit))
			} else {
				Expect(pod.Annotations).NotTo(HaveKey("nvidia.com/container.worker.gpu-memory.limit"))
			}
			Expect(kairesources.ValidateGPUFractionRequest(pod)).To(Succeed())
			parsed, err := kairesources.ParsePodGPUFractionRequest(pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Memory.Cmp(resource.MustParse(request))).To(BeZero())
		}
	},
		Entry("equal decimal quantities", "500M", "500M"),
		Entry("equal binary quantities", "500Mi", "500Mi"),
		Entry("a larger decimal limit", "500M", "525M"),
		Entry("equivalent quantities with different units", "500Mi", "524288k"),
		Entry("a decimal request without a limit", "500M", ""),
		Entry("a binary request without a limit", "500Mi", ""),
	)

	DescribeTable("selects request and limit formats independently by cluster version", func(version string, perContainerRequest, perContainerLimit bool) {
		capabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
		Expect(err).NotTo(HaveOccurred())
		options := composer.FieldsComposerOptions{}
		Expect(composer_v2.SetClusterDependentOptions(context.Background(), &options, capabilities)).To(Succeed())
		var src workloadsApi.WorkspaceSpecSpecAllOfCompute
		Expect(json.Unmarshal([]byte(`{"gpuMemoryRequest":"500Mi","gpuMemoryLimit":"600Mi"}`), &src)).To(Succeed())
		c := NewInterTrainJobComposer(composer.Workspace, "example", "p1", "runai-p1", "user", nil, options, nil)
		target := NewMinimalBatchJobTarget()
		target.Job.Spec.Template.Spec.Containers[0] = corev1.Container{Name: "worker"}
		Expect(c.processComputeFields(context.Background(), &src, target)()).To(BeNil())
		annotations := target.Job.Spec.Template.Annotations
		requestKey := fmt.Sprintf(composer.CudaGpuMemoryRequestAnnotationKeyFormat, "worker")
		limitKey := fmt.Sprintf(composer.CudaGpuMemoryLimitAnnotationKeyFormat, "worker")
		if perContainerRequest {
			Expect(annotations).To(HaveKeyWithValue(requestKey, "500Mi"))
			Expect(annotations).NotTo(HaveKey(composer.GpuMemoryAnnotation))
			pod := &corev1.Pod{ObjectMeta: target.Job.Spec.Template.ObjectMeta, Spec: target.Job.Spec.Template.Spec}
			Expect(kairesources.ValidateGPUFractionRequest(pod)).To(Succeed())
		} else {
			Expect(annotations).To(HaveKeyWithValue(composer.GpuMemoryAnnotation, "524"))
			Expect(annotations).NotTo(HaveKey(requestKey))
		}
		if perContainerLimit {
			Expect(annotations).To(HaveKeyWithValue(limitKey, "600Mi"))
			Expect(target.Job.Spec.Template.Spec.Containers[0].Env).NotTo(ContainElement(HaveField("Name", composer.GpuMemoryLimitEnvVar)))
		} else {
			Expect(annotations).NotTo(HaveKey(limitKey))
			Expect(target.Job.Spec.Template.Spec.Containers[0].Env).To(ContainElement(corev1.EnvVar{Name: composer.GpuMemoryLimitEnvVar, Value: "600Mi"}))
		}
	},
		Entry("older than per-container limits", "2.25.0", false, false),
		Entry("supports limits but not the new readers", "3.0.22", false, true),
		Entry("supports the new readers", "3.1.0", true, true),
		Entry("development cluster", "0.0.0-dev", true, true),
	)

	It("retains the legacy request and environment limit for older clusters", func() {
		var src workloadsApi.WorkspaceSpecSpecAllOfCompute
		Expect(json.Unmarshal([]byte(`{"gpuMemoryRequest":"500M","gpuMemoryLimit":"500M"}`), &src)).To(Succeed())
		c := NewInterTrainJobComposer(composer.Workspace, "example", "p1", "runai-p1", "user", nil, composer.FieldsComposerOptions{}, nil)
		target := NewMinimalBatchJobTarget()
		target.Job.Spec.Template.Spec.Containers[0] = corev1.Container{Name: "worker"}
		Expect(c.processComputeFields(context.Background(), &src, target)()).To(BeNil())
		Expect(target.Job.Spec.Template.Annotations).To(HaveKeyWithValue(composer.GpuMemoryAnnotation, "500"))
		Expect(target.Job.Spec.Template.Annotations).NotTo(HaveKey(fmt.Sprintf(composer.CudaGpuMemoryRequestAnnotationKeyFormat, "worker")))
		Expect(target.Job.Spec.Template.Spec.Containers[0].Env).To(ContainElement(corev1.EnvVar{Name: composer.GpuMemoryLimitEnvVar, Value: "500M"}))
	})
})
