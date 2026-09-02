/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed_to_kubeflow

import (
	. "github.com/onsi/ginkgo/v2"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var storageClassConfigTests = []TableEntry{
	Entry("stconfig: defaults", ComposerTest{
		framework:          workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname:    "pvc_combo",
		expectedFname:      "pvc_combo_stconfig_defaults_xgboost",
		storageClassConfig: "defaults",
	}),

	Entry("stconfig: unit error", ComposerTest{
		framework:          workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname:    "pvc_combo",
		expectedError:      "pvcSize: unit suffix 'T' is not supported",
		storageClassConfig: "ephemeral_unit_error",
	}),

	Entry("stconfig: access mode", ComposerTest{
		framework:          workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname:    "pvc_combo",
		expectedError:      "readWriteOnce is not supported for storage class persistent",
		storageClassConfig: "persistent_access_mode_error",
	}),

	Entry("stconfig: volume error", ComposerTest{
		framework:          workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname:    "pvc_combo",
		expectedError:      "Block is not supported for storage class persistent",
		storageClassConfig: "persistent_volume_error",
	}),

	Entry("stconfig: max error", ComposerTest{
		framework:          workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname:    "pvc_combo",
		expectedError:      "claim size should be no greater than 4M",
		storageClassConfig: "small_max_error",
	}),

	Entry("stconfig: step error", ComposerTest{
		framework:          workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname:    "pvc_combo",
		expectedError:      "claim size is not aligned with 2M steps from the minimal value 1M",
		storageClassConfig: "small_step_error",
	}),

	Entry("stconfig: min_error", ComposerTest{
		framework:          workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname:    "pvc_combo",
		expectedError:      "claim size should be no less than 1G",
		storageClassConfig: "static_min_error",
	}),
}
