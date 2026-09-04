/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iw_tw_workloads

import (
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	. "github.com/onsi/ginkgo/v2"
)

var storageClassConfigTests = []TableEntry{
	Entry("stconfig: defaults", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "pvc_combo",
		expectedFname:      "pvc_combo_stconfig_defaults_ws",
		storageClassConfig: "defaults",
	}),
	Entry("stconfig: unit error", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "pvc_combo",
		storageClassConfig: "ephemeral_unit_error",
		expectedError:      "pvcSize: unit suffix 'T' is not supported",
	}),
	Entry("stconfig: access mode", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "pvc_combo",
		storageClassConfig: "persistent_access_mode_error",
		expectedError:      "readWriteOnce is not supported for storage class persistent",
	}),
	Entry("stconfig: volume error", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "pvc_combo",
		storageClassConfig: "persistent_volume_error",
		expectedError:      "Block is not supported for storage class persistent",
	}),
	Entry("stconfig: max error", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "pvc_combo",
		storageClassConfig: "small_max_error",
		expectedError:      "claim size should be no greater than 4M",
	}),
	Entry("stconfig: step error", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "pvc_combo",
		storageClassConfig: "small_step_error",
		expectedError:      "claim size is not aligned with 2M steps from the minimal value 1M",
	}),
	Entry("stconfig: min_error", ComposerTest{
		workloadType:       string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:          "pvc_combo",
		storageClassConfig: "static_min_error",
		expectedError:      "claim size should be no less than 1G",
	}),
}
