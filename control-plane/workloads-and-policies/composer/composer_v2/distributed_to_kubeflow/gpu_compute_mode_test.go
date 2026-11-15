/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed_to_kubeflow

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
		c := &DistributedToKubeflowComposer{fc: &composer.FieldsComposer{Options: options}}
		for _, framework := range []workloadsApi.DistributedFramework{workloadsApi.DISTRIBUTEDFRAMEWORK_MPI, workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH, workloadsApi.DISTRIBUTEDFRAMEWORK_TF, workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST, workloadsApi.DISTRIBUTEDFRAMEWORK_JAX} {
			target := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{}}}}
			Expect(c.composeContainerName(framework, &target.Spec.Containers[0].Name, composer.StringToString, nil)()).To(BeNil())
			var sizing composer.GpuSizingInfo
			Expect(c.processComputeFields(context.Background(), &src, target, &sizing)()).To(BeNil())
			name := target.Spec.Containers[0].Name
			Expect(name).NotTo(BeEmpty())
			Expect(target.Annotations).To(HaveKeyWithValue(fmt.Sprintf("nvidia.com/container.%s.gpu-compute.mode", name), mode))
			Expect(target.Annotations).To(HaveKeyWithValue(composer.GpuFractionAnnotation, "0.25"))
		}
	})
})
