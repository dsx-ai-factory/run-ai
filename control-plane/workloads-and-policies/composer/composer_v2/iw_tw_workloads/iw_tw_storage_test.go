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

var storageTests = []TableEntry{
	Entry("git", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "git",
		expectedFname: "git_ws",
	}),

	Entry("git_with_branch", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "git_with_branch",
		expectedFname: "git_with_branch_ws",
	}),

	Entry("invalid_git", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "invalid_git",
		expectedError: "To point to your Git revision, provide only one of the following: ‘branch’ OR ‘revision’",
	}),

	Entry("git_password_and_secretref", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "git_password_and_secretref",
		expectedError: "git: passwordSecret is mutually exclusive with secretRef",
	}),

	Entry("git_secretref_no_name", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "git_secretref_no_name",
		expectedError: "git: secret's name required",
	}),

	Entry("git_secretref_no_auth_method", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "git_secretref_no_auth_method",
		expectedError: "git: authenticatingMethod is required",
	}),

	Entry("git_with_ssh_key", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "git_with_ssh_key",
		expectedFname: "git_with_ssh_key_ws",
	}),

	Entry("git_with_ssh_key_old_cluster", ComposerTest{
		workloadType:   string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:      "git_with_ssh_key",
		clusterVersion: "2.22.0",
		expectedError:  "git: authenticatingMethod is not supported in this cluster",
	}),

	Entry("git_with_password", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "git_with_password",
		expectedFname: "git_with_password_ws",
	}),

	Entry("s3", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "s3",
		expectedFname: "s3_ws",
	}),

	Entry("s3_training", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "s3",
		expectedFname: "s3_tr",
	}),

	Entry("s3_run_as_user", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "s3_run_as_user",
		expectedFname: "s3_run_as_user_ws",
	}),

	Entry("s3_run_as_user_training", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "s3_run_as_user",
		expectedFname: "s3_run_as_user_tr",
	}),

	// Regression: runAsUid=0 / runAsGid=0 (root) with custom uidGidSource must
	// be preserved on the container securityContext instead of being silently
	// dropped as if the field were unset.
	Entry("s3_run_as_user_root", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "s3_run_as_user_root",
		expectedFname: "s3_run_as_user_root_ws",
	}),

	Entry("host path - multiple entries with different mount propagation values - invalid 2", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "host_path_mount_propagation_multiple_invalid2",
		expectedError: "hostPath: all volumes must have the same mountPropagation value",
	}),

	Entry("host_path_and_nfs", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "host_path_and_nfs",
		expectedFname: "host_path_and_nfs_ws",
	}),

	Entry("pvc", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc",
		expectedFname: "pvc_ws",
	}),

	Entry("pvc_no_claim_name", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_no_claim_name",
		expectedFname: "pvc_no_claim_name_ws",
	}),

	Entry("pvc_no_claim_name_existing", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_existing_no_claim_name",
		expectedError: "claimName is mandatory for existing PVC",
	}),

	Entry("pvc_existing_with_claim_info", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_existing_with_claim_info",
		expectedFname: "pvc_existing_with_claim_info_ws",
	}),

	Entry("pvc_new_with_no_claim_info", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_new_with_no_claim_info",
		expectedError: "claimInfo is mandatory for non-existing PVC",
	}),

	Entry("pvc_existing", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_existing",
		expectedFname: "pvc_existing_ws",
	}),

	Entry("pvc_existing_not_specified", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_existing_not_specified",
		expectedFname: "pvc_existing_not_specified_ws",
	}),

	Entry("pvc_not_existing", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_not_existing",
		expectedFname: "pvc_not_existing_ws",
	}),

	Entry("pvc_not_existing_ephemeral_false", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_not_existing_ephemeral_false",
		expectedFname: "pvc_not_existing_ephemeral_false_ws",
	}),

	Entry("pvc_not_existing_ephemeral_true", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "pvc_not_existing_ephemeral_true",
		expectedFname: "pvc_not_existing_ephemeral_true_ws",
	}),

	Entry("config_map", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "config_map",
		expectedFname: "config_map_ws",
	}),

	Entry("config_map_subpath", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "config_map_subpath",
		expectedFname: "config_map_subpath_ws",
	}),

	Entry("config_map_defaultmode", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "config_map_defaultmode",
		expectedFname: "config_map_defaultmode_ws",
	}),

	Entry("empty_dir", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "empty_dir",
		expectedFname: "empty_dir_ws",
	}),

	Entry("empty_dir_with_memory", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "empty_dir_with_memory",
		expectedFname: "empty_dir_with_memory_ws",
	}),

	Entry("empty_dir_no_path", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "empty_dir_no_path",
		expectedError: "path is mandatory for emptyDir",
	}),

	Entry("empty_dir_invalid_medium", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "empty_dir_invalid_medium",
		expectedError: "emptyDir medium must be either empty or 'Memory'",
	}),

	Entry("secret", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "secret",
		expectedFname: "secret_ws",
	}),

	Entry("secret_defaultmode", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "secret_defaultmode",
		expectedFname: "secret_defaultmode_ws",
	}),

	Entry("host path with mount propagation", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "host_path_mount_propagation",
		expectedFname: "host_path_mount_propagation_ws",
	}),

	Entry("host path - multiple entries with same mount propagation value", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "host_path_mount_propagation_multiple",
		expectedFname: "host_path_mount_propagation_multiple_ws",
	}),

	Entry("host path - multiple entries without mount propagation", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "host_path_multiple",
		expectedFname: "host_path_multiple_ws",
	}),

	Entry("host path - multiple entries without mount propagation 2", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "host_path_mount_propagation_multiple2",
		expectedFname: "host_path_mount_propagation_multiple2_ws",
	}),

	Entry("host path - multiple entries with different mount propagation values - invalid", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "host_path_mount_propagation_multiple_invalid",
		expectedError: "hostPath: all volumes must have the same mountPropagation value",
	}),

	// data-volume tests

	Entry("data-volume", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "data-volume",
		expectedFname: "data-volume",
		dataVolumeClaimNames: map[string]string{
			"1234-5678": "my-data-volume-pvc",
		},
	}),

	Entry("data-volume-multi", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "data-volume-multi",
		expectedFname: "data-volume-multi",
		dataVolumeClaimNames: map[string]string{
			"1234-5678": "data-volume-pvc-1",
			"8765-4321": "data-volume-pvc-2",
		},
	}),

	Entry("data-volume-missing-id", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "data-volume-missing-id",
		expectedError: "dataVolume: id is mandatory",
	}),

	Entry("data-volume-missing-mountPath", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
		specFname:     "data-volume-missing-mountPath",
		expectedError: "dataVolume: mountPath is mandatory",
	}),

	Entry("data-volume-id-not-found", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "data-volume-id-not-found",
		expectedError: "dataVolume: claim name not found for data volume ID unknown-id-12345",
		dataVolumeClaimNames: map[string]string{
			"different-id": "some-pvc",
		},
	}),

	// data-volume tests end

	Entry("volume_pvc", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "volume_pvc",
		expectedFname: "volume_pvc_ws",

		assetsInfo: &composer.ComplimentaryAssetInfo{
			Environment: workloadsApi.Environment{
				Name: "volume-pvc-env-name",
				Id:   "volume-pvc-env-id",
			},
		},
	}),

	Entry("existing_pvc", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		specFname:     "existing_pvc",
		expectedFname: "existing_pvc_ws",

		assetsInfo: &composer.ComplimentaryAssetInfo{
			Environment: workloadsApi.Environment{
				Name: "existing-pvc-env-name",
				Id:   "existing-pvc-env-id",
			},
		},
	}),
}
