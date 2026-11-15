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
			Expect(c.processComputeFields(context.Background(), &src, target)()).To(BeNil())
			Expect(target.Job.Spec.Template.Annotations).To(HaveKeyWithValue("nvidia.com/container.workload-main.gpu-compute.mode", mode))
			Expect(target.Job.Spec.Template.Annotations).To(HaveKeyWithValue(composer.GpuFractionAnnotation, "0.25"))
		}
	})
})
