/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iw_tw_workloads

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
)

var computeTests = []TableEntry{
	Entry("cpu memory request syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "cpu_memory_req_syntax",
		expectedError: "cpuMemoryRequest does not comply with k8s resource units rules",
	}),

	Entry("cpu request negative", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "cpu_core_req_negative",
		expectedError: "compute.cpuCoreRequest must be no less than zero",
	}),

	Entry("cpu limit negative", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "cpu_core_limit_negative",
		expectedError: "compute.cpuCoreLimit must be no less than zero",
	}),

	Entry("cpu memory limit syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "cpu_memory_limit_syntax",
		expectedError: "cpuMemoryLimit does not comply with k8s resource units rules",
	}),

	Entry("cpu memory limit negative", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "cpu_memory_limit_negative",
		expectedError: "cpuMemoryLimit must be no less than zero",
	}),

	Entry("cpu memory request negative", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "cpu_memory_req_negative",
		expectedError: "compute.cpuMemoryRequest must be no less than zero",
	}),

	Entry("cpu memory request vs limit error", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "cpu_memory_limit_error",
		expectedError: "cpuMemoryLimit (10000000) should be no less than cpuMemoryRequest (11000000)",
	}),

	Entry("cpu core request vs limit error", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "cpu_core_limit_error",
		expectedError: "cpuCoreLimit (0.5) should be no less than cpuCoreRequest (1.0)",
	}),

	Entry("multiple_gpus", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "multiple_gpus",
		expectedFname: "multiple_gpus_tr",
	}),

	Entry("gpu portion req negative", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "gpu_portion_req_negative",
		expectedError: "compute.gpuPortionRequest must be greater than zero",
	}),

	Entry("gpu portion limit negative", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "gpu_portion_limit_negative",
		expectedError: "compute.gpuPortionLimit must be greater than zero",
	}),

	// gpuPortionRequest must be either 0-1 (fraction) or a whole number (1, 2, 3...).
	// Values like 2.5 are invalid - they're neither a valid fraction nor a valid device count.
	Entry("gpu portion invalid fractional", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "gpu_portion_invalid_fractional",
		expectedError: "compute.gpuPortionRequest must be either a fraction (0 to 1) or a whole number",
	}),

	Entry("gpu portion - request and limit", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_portion_req_and_limit",
		expectedFname: "gpu_portion_req_and_limit_ws",
		expectedRequirements: composer.WorkloadRequirements{
			GpuOptimization: true,
		},
	}),

	// Same input as above but pinned to a cluster version below
	// MinVersionGpuMemoryLimitAnnotations (3.0.0). Verifies the composer
	// falls back to the legacy RUNAI_GPU_PORTION_LIMIT env var instead of
	// the kai.scheduler/container.<name>.gpu-memory.portion.limit annotation.
	Entry("gpu portion - request and limit (legacy env, cluster < 3.0)", ComposerTest{
		workloadType:   string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:      "gpu_portion_req_and_limit",
		expectedFname:  "gpu_portion_req_and_limit_ws_legacy_env",
		clusterVersion: "2.99.99",
		expectedRequirements: composer.WorkloadRequirements{
			GpuOptimization: true,
		},
	}),

	Entry("gpu portion - no portion", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_portion_no_portion",
		expectedError: "compute.gpuPortionRequest is mandatory when gpuRequestType is portion",
	}),

	Entry("gpu portion - request vs limit error", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_portion_limit_error",
		expectedError: "gpuPortionLimit should be no less than gpuPortionRequest",
	}),

	Entry("one_gpu", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "one_gpu",
		expectedFname: "one_gpu_ws",
	}),

	Entry("gpu_multiple", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_multiple",
		expectedFname: "gpu_multiple_ws",
	}),

	Entry("gpu memory no memory", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_no_memory",
		expectedError: "compute.gpuMemoryRequest is mandatory when gpuRequestType is memory",
	}),

	Entry("gpu memory negative", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_negative",
		expectedError: "compute.gpuMemoryRequest must be greater than zero",
	}),

	Entry("gpu memory limit negative", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_limit_negative",
		expectedError: "compute.gpuMemoryLimit must be greater than zero",
	}),

	Entry("gpu memory request and limit", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_req_and_limit",
		expectedFname: "gpu_memory_req_and_limit_ws",
		expectedRequirements: composer.WorkloadRequirements{
			GpuOptimization: true,
		},
	}),

	// Same input as above but pinned to a cluster version below
	// MinVersionGpuMemoryLimitAnnotations (3.0.0). Verifies the composer
	// falls back to the legacy RUNAI_GPU_MEMORY_LIMIT env var instead of
	// the nvidia.com/container.<name>.gpu-memory.limit annotation.
	Entry("gpu memory request and limit (legacy env, cluster < 3.0)", ComposerTest{
		workloadType:   string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:      "gpu_memory_req_and_limit",
		expectedFname:  "gpu_memory_req_and_limit_ws_legacy_env",
		clusterVersion: "2.99.99",
		expectedRequirements: composer.WorkloadRequirements{
			GpuOptimization: true,
		},
	}),

	Entry("gpu memory only", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_limit_only",
		expectedFname: "gpu_memory_limit_only_ws",
		expectedRequirements: composer.WorkloadRequirements{
			GpuOptimization: true,
		},
	}),

	Entry("gpu memory request vs limit err", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_limit_error",
		expectedError: "gpuMemoryLimit should be no less than gpuMemoryRequest",
	}),

	Entry("gpu memory request syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_req_syntax",
		expectedError: "gpuMemoryRequest does not comply with k8s resource units rules",
	}),

	Entry("gpu memory limit syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_limit_syntax",
		expectedError: "gpuMemoryLimit does not comply with k8s resource units rules",
	}),

	Entry("submitting a workload with gpuMemoryRequest of 0 or below would fail", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "gpu_memory_request_of_0",
		expectedError: "compute.gpuMemoryRequest must be greater than zero",
	}),

	Entry("memory with 2 devices request", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "compute_memory_2_devices_req",
		expectedFname: "compute_memory_2_devices_req",
	}),

	Entry("portion with 2 devices request", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "compute_portion_2_devices_req",
		expectedFname: "compute_portion_2_devices_req",
	}),

	Entry("conflicting portion memory", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "compute_conflicting_portion_memory",
		expectedError: "compute.gpuPortionRequest and gpuMemoryRequest are mutually exclusive",
	}),

	// gpuPortionRequest: 1 is a whole-GPU allocation, not a fraction. It must emit a
	// whole-GPU resource limit and no gpu-fraction annotation, matching native behavior.
	Entry("gpu portion one with device", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "gpu_portion_one_with_device",
		expectedFname: "gpu_portion_one_with_device_tr",
	}),

	// gpuPortionRequest: 1 without an explicit device count should still infer a single
	// whole GPU, matching the native controller.
	Entry("gpu portion one no device", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "gpu_portion_one_no_device",
		expectedFname: "gpu_portion_one_no_device_tr",
	}),

	Entry("assets_no_cpu", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "assets_no_cpu",
		expectedFname: "assets_no_cpu_ws",

		assetsInfo: &composer.ComplimentaryAssetInfo{
			Environment: workloadsApi.Environment{
				Name: "jupyter-lab",
				Id:   "7299cbff-120d-4bcd-9cb6-0b2769de7b7e",
			},
		},
	}),

	Entry("ext_resources", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "ext_resources",
		expectedFname: "ext_resources_tr",
	}),

	Entry("ext_resources_no_quantity", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "ext_resources_no_quantity",
		expectedError: "compute.extendedResources: quantity is not provided",
	}),

	Entry("ext_resources_syntax", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "ext_resources_syntax",
		expectedError: "malformed quantity: quantities must match the regular expression",
	}),

	Entry("restart_policy", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "restart_policy",
		expectedFname: "restart_policy_tr",
	}),

	Entry("restart_policy_always", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "restart_policy_always",
		expectedFname: "restart_policy_always_tr",
	}),
}
