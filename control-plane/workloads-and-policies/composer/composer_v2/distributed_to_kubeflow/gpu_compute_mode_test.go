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
			Expect(c.processComputeFields(context.Background(), &src, nil, target, &sizing)()).To(BeNil())
			name := target.Spec.Containers[0].Name
			Expect(name).NotTo(BeEmpty())
			Expect(target.Annotations).To(HaveKeyWithValue(fmt.Sprintf("nvidia.com/container.%s.gpu-compute.mode", name), mode))
			Expect(target.Annotations).To(HaveKeyWithValue(composer.GpuFractionAnnotation, "0.25"))
		}
	})
})

var _ = Describe("GPU selection in distributed training", func() {
	It("keeps the worker model off the inherited CPU-only MPI launcher", func() {
		var workload workloadsApi.DistributedWorkload
		Expect(json.Unmarshal([]byte(`{
		    "spec":{"image":"ubuntu", "numWorkers":2, "distributedFramework":"MPI",
		            "compute":{"gpuDevicesRequest":2},"gpuType":"H100"}
		}`), &workload)).To(Succeed())
		workload.Spec.SetDistributedFramework(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)
		options := composer.FieldsComposerOptions{NodeAffinitySupported: true}
		c := NewDistributedToKubeflowComposer(composer.Distributed, "mpi-example", "team", "runai-team", "tester", nil, nil, options, nil)
		c.WithClusterStorageConfig(testClusterStorageConfig("https://example.com", false))
		target := NewMinimalDistributedTarget(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI, true)
		Expect(c.Compose(context.Background(), &workload, target)()).To(BeNil())
		Expect(target.WorkerSpec.Template.Spec.NodeSelector).To(HaveKeyWithValue(composer.NvidiaProductLabel, "H100"))
		Expect(target.MasterSpec.Template.Spec.NodeSelector).NotTo(HaveKey(composer.NvidiaProductLabel))
		Expect(target.MasterSpec.Template.Annotations).NotTo(HaveKey("run.ai/gpu-type-label"))
	})
})
