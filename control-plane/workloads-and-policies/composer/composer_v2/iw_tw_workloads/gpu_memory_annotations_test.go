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
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
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
			options := composer.FieldsComposerOptions{GpuMemoryLimitAnnotationsSupported: true}
			c := NewInterTrainJobComposer(kind, "example", "p1", "runai-p1", "user", nil, options, nil)
			target := NewMinimalBatchJobTarget()
			target.Job.Spec.Template.Spec.Containers[0] = corev1.Container{Name: "worker"}
			Expect(c.processComputeFields(context.Background(), &src, target)()).To(BeNil())
			pod := &corev1.Pod{ObjectMeta: target.Job.Spec.Template.ObjectMeta, Spec: target.Job.Spec.Template.Spec}
			Expect(pod.Annotations).NotTo(HaveKey(composer.GpuMemoryAnnotation))
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
