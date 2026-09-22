/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed_to_kubeflow

import (
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
)

var storageTests = []TableEntry{
	Entry("git", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "git",
		expectedFname:   "git_pytorch",
	}),

	Entry("git_with_branch", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "git_with_branch",
		expectedFname:   "git_with_branch_xgboost",
	}),

	Entry("invalid_git", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "invalid_git",
		expectedError:   "To point to your Git revision, provide only one of the following: ‘branch’ OR ‘revision’",
	}),

	Entry("git_with_ssh_key", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "git_with_ssh_key",
		expectedFname:   "git_with_ssh_key_mpi",
	}),

	Entry("git_with_ssh_key_old_cluster", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "git_with_ssh_key",
		expectedError:   "git: authenticatingMethod is not supported in this cluster",
		clusterVersion:  "2.22.0",
	}),

	Entry("git_with_password", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "git_with_password",
		expectedFname:   "git_with_password_tf",
	}),

	Entry("s3", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "s3",
		expectedFname:   "s3_jax",
	}),

	Entry("s3_run_as_user", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "s3_run_as_user",
		expectedFname:   "s3_run_as_user_pytorch",
	}),

	Entry("s3_excluded", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "s3_excluded",
		expectedFname:   "s3_excluded_pytorch",
	}),

	Entry("host path - multiple entries with different mount propagation values - invalid 2", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "host_path_mount_propagation_multiple_invalid2",
		expectedError:   "hostPath: all volumes must have the same mountPropagation value",
	}),

	Entry("host_path_and_nfs", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "host_path_and_nfs",
		expectedFname:   "host_path_and_nfs_pytorch",
	}),

	Entry("pvc", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "pvc",
		expectedFname:   "pvc_tf",
	}),

	Entry("pvc_no_claim_name", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "pvc_no_claim_name",
		expectedFname:   "pvc_no_claim_name_xgboost",
	}),

	Entry("pvc_no_claim_name_existing", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "pvc_existing_no_claim_name",
		expectedError:   "claimName is mandatory for existing PVC",
	}),

	Entry("pvc_existing_with_claim_info", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "pvc_existing_with_claim_info",
		expectedFname:   "pvc_existing_with_claim_info_mpi",
	}),

	Entry("pvc_new_with_no_claim_info", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "pvc_new_with_no_claim_info",
		expectedError:   "claimInfo is mandatory for non-existing PVC",
	}),

	Entry("pvc_existing", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "pvc_existing",
		expectedFname:   "pvc_existing_tf",
	}),

	Entry("pvc_existing_not_specified", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "pvc_existing_not_specified",
		expectedFname:   "pvc_existing_not_specified_xgboost",
	}),

	Entry("pvc_not_existing", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		workerSpecFname: "pvc_not_existing",
		expectedFname:   "pvc_not_existing_jax",
	}),

	Entry("pvc_not_existing_ephemeral_false", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "pvc_not_existing_ephemeral_false",
		expectedFname:   "pvc_not_existing_ephemeral_false_mpi",
	}),

	Entry("pvc_not_existing_ephemeral_true", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "pvc_not_existing_ephemeral_true",
		expectedFname:   "pvc_not_existing_ephemeral_true_pytorch",
	}),

	Entry("config_map", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "config_map",
		expectedFname:   "config_map_tf",
	}),

	Entry("config_map_subpath", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "config_map_subpath",
		expectedFname:   "config_map_subpath_xgboost",
	}),

	Entry("config_map_defaultmode", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "config_map_defaultmode",
		expectedFname:   "config_map_defaultmode_jax",
	}),

	Entry("empty_dir", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "empty_dir",
		expectedFname:   "empty_dir_mpi",
	}),

	Entry("empty_dir_with_memory", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "empty_dir_with_memory",
		expectedFname:   "empty_dir_with_memory_pytorch",
	}),

	Entry("empty_dir_no_path", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "empty_dir_no_path",
		expectedError:   "path is mandatory for emptyDir",
	}),

	Entry("empty_dir_invalid_medium", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "empty_dir_invalid_medium",
		expectedError:   "emptyDir medium must be either empty or 'Memory'",
	}),

	Entry("secret", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "secret",
		expectedFname:   "secret_jax",
	}),

	Entry("secret_defaultmode", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "secret_defaultmode",
		expectedFname:   "secret_defaultmode_mpi",
	}),

	Entry("host path with mount propagation", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "host_path_mount_propagation",
		expectedFname:   "host_path_mount_propagation_pytorch",
	}),

	Entry("host path - multiple entries with same mount propagation value", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "host_path_mount_propagation_multiple",
		expectedFname:   "host_path_mount_propagation_multiple_tf",
	}),

	Entry("host path - multiple entries without mount propagation", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST,
		workerSpecFname: "host_path_multiple",
		expectedFname:   "host_path_multiple_xgboost",
	}),

	Entry("host path - multiple entries without mount propagation 2", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_JAX,
		noMaster:        true,
		workerSpecFname: "host_path_mount_propagation_multiple2",
		expectedFname:   "host_path_mount_propagation_multiple2_jax",
	}),

	Entry("host path - multiple entries with different mount propagation values - invalid", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname: "host_path_mount_propagation_multiple_invalid",
		expectedError:   "hostPath: all volumes must have the same mountPropagation value",
	}),

	Entry("data-volume-missing-id", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH,
		workerSpecFname: "data-volume-missing-id",
		expectedError:   "dataVolume: id is mandatory",
	}),

	Entry("data-volume-missing-mountPath", ComposerTest{
		framework:       workloadsApi.DISTRIBUTEDFRAMEWORK_TF,
		workerSpecFname: "data-volume-missing-mountPath",
		expectedError:   "dataVolume: mountPath is mandatory",
	}),
}
