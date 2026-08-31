/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed_to_kubeflow

import (
	. "github.com/onsi/ginkgo/v2"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var computeTests = []TableEntry{
	Entry("cpu memory request syntax", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "cpu_memory_req_syntax",
		expectedError:   "cpuMemoryRequest does not comply with k8s resource units rules",
	}),

	Entry("cpu request negative", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "cpu_core_req_negative",
		expectedError:   "compute.cpuCoreRequest must be no less than zero",
	}),

	Entry("cpu limit negative", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "cpu_core_limit_negative",
		expectedError:   "compute.cpuCoreLimit must be no less than zero",
	}),

	Entry("cpu memory limit syntax", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "cpu_memory_limit_syntax",
		expectedError:   "cpuMemoryLimit does not comply with k8s resource units rules",
	}),

	Entry("cpu memory limit negative", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "cpu_memory_limit_negative",
		expectedError:   "cpuMemoryLimit must be no less than zero",
	}),

	Entry("cpu memory request negative", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "cpu_memory_req_negative",
		expectedError:   "compute.cpuMemoryRequest must be no less than zero",
	}),

	Entry("cpu memory request vs limit error", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "cpu_memory_limit_error",
		expectedError:   "cpuMemoryLimit (10000000) should be no less than cpuMemoryRequest (11000000)",
	}),

	Entry("cpu core request vs limit error", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "cpu_core_limit_error",
		expectedError:   "cpuCoreLimit (0.5) should be no less than cpuCoreRequest (1.0)",
	}),

	Entry("multiple_gpus", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "multiple_gpus",
		expectedFname:   "multiple_gpus_pytorch",
	}),

	Entry("gpu portion req negative", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "gpu_portion_req_negative",
		expectedError:   "compute.gpuPortionRequest must be greater than zero",
	}),

	Entry("gpu portion limit negative", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "gpu_portion_limit_negative",
		expectedError:   "compute.gpuPortionLimit must be greater than zero",
	}),

	Entry("gpu portion - no portion", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "gpu_portion_no_portion",
		expectedError:   "compute.gpuPortionRequest is mandatory when gpuRequestType is portion",
	}),

	Entry("gpu portion - request vs limit error", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "gpu_portion_limit_error",
		expectedError:   "gpuPortionLimit should be no less than gpuPortionRequest",
	}),

	Entry("one_gpu", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "one_gpu",
		expectedFname:   "one_gpu_pytorch",
	}),

	Entry("gpu_multiple", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "gpu_multiple",
		expectedFname:   "gpu_multiple_pytorch",
	}),

	Entry("gpu memory no memory", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "gpu_memory_no_memory",
		expectedError:   "compute.gpuMemoryRequest is mandatory when gpuRequestType is memory",
	}),

	Entry("gpu memory negative", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "gpu_memory_negative",
		expectedError:   "compute.gpuMemoryRequest must be greater than zero",
	}),

	Entry("gpu memory limit negative", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "gpu_memory_limit_negative",
		expectedError:   "compute.gpuMemoryLimit must be greater than zero",
	}),

	Entry("gpu memory request vs limit err", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "gpu_memory_limit_error",
		expectedError:   "gpuMemoryLimit should be no less than gpuMemoryRequest",
	}),

	Entry("gpu memory request syntax", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "gpu_memory_req_syntax",
		expectedError:   "gpuMemoryRequest does not comply with k8s resource units rules",
	}),

	Entry("gpu memory limit syntax", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "gpu_memory_limit_syntax",
		expectedError:   "gpuMemoryLimit does not comply with k8s resource units rules",
	}),

	Entry("submitting a workload with gpuMemoryRequest of 0 or below would fail", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "gpu_memory_request_of_0",
		expectedError:   "compute.gpuMemoryRequest must be greater than zero",
	}),

	Entry("memory with 2 devices request", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "compute_memory_2_devices_req",
		expectedFname:   "compute_memory_2_devices_req_jax",
	}),

	Entry("portion with 2 devices request", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "compute_portion_2_devices_req",
		masterSpecFname: "master_no_gpu",
		expectedFname:   "compute_portion_2_devices_req_mpi",
	}),

	Entry("conflicting portion memory", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "compute_conflicting_portion_memory",
		expectedError:   "compute.gpuPortionRequest and gpuMemoryRequest are mutually exclusive",
	}),

	Entry("ext_resources", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "ext_resources",
		expectedFname:   "ext_resources_tf",
	}),

	Entry("ext_resources_no_quantity", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "ext_resources_no_quantity",
		expectedError:   "compute.extendedResources: quantity is not provided",
	}),

	Entry("ext_resources_syntax", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "ext_resources_syntax",
		expectedError:   "malformed quantity: quantities must match the regular expression",
	}),
}
