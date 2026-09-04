/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_workloads

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var distributedTests = []TableEntry{
	Entry("basic_mpi", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "basic_mpi",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)),
			},
		},
	}),

	Entry("mpi_slots", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "mpi_slots",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)),
				SlotsPerWorker:       *workloadsApi.NewNullableInt32(convutils.Ref(int32(12))),
			},
		},
	}),

	Entry("mpi_ssh", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "mpi_ssh",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)),
				SshAuthMountPath:     *workloadsApi.NewNullableString(convutils.Ref("/home/tmp")),
			},
		},
	}),

	Entry("mpi_x_pytorch", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "mpi_x_pytorch",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)),
				MaxReplicas:          *workloadsApi.NewNullableInt32(convutils.Ref(int32(32))),
			},
		},
		expectedError: "maxReplicas is not supported by MPI framework",
	}),

	Entry("basic_pytorch", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "basic_pytorch",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)),
			},
		},
	}),

	Entry("pytorch_replicas_range", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "pytorch_replicas_range",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)),
				MinReplicas:          *workloadsApi.NewNullableInt32(convutils.Ref(int32(1))),
				MaxReplicas:          *workloadsApi.NewNullableInt32(convutils.Ref(int32(32))),
			},
		},
	}),

	Entry("pytorch_x_mpi", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "pytorch_x_mpi",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)),
				SlotsPerWorker:       *workloadsApi.NewNullableInt32(convutils.Ref(int32(12))),
			},
		},
		expectedError: "slotsPerWorker is not supported by PyTorch framework",
	}),

	Entry("pytorch_x_mpi_2", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "pytorch_x_mpi_2",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)),
				SshAuthMountPath:     *workloadsApi.NewNullableString(convutils.Ref("/home/tmp")),
			},
		},
		expectedError: "sshAuthMountPath is not supported by PyTorch framework",
	}),

	Entry("pytorch_invalid_range", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "pytorch_x_mpi",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)),
				MinReplicas:          *workloadsApi.NewNullableInt32(convutils.Ref(int32(10))),
				MaxReplicas:          *workloadsApi.NewNullableInt32(convutils.Ref(int32(4))),
			},
		},
		expectedError: "maxReplicas should be no less than minReplicas",
	}),

	Entry("pytorch_no_max_replicas", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "basic",
		expectedFname: "pytorch_no_max_replicas",

		distributed: &DistributedTest{
			numWorkers:      2,
			masterSpecFname: "basic_master",
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)),
				MinReplicas:          *workloadsApi.NewNullableInt32(convutils.Ref(int32(10))),
			},
		},
	}),

	Entry("distributed workload with backoff limit", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "backoffLimit",
		expectedFname: "backoffLimit_dist",

		distributed: &DistributedTest{
			numWorkers: 2,
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)),
			},
			masterSpecFname: "backoffLimit_master",
		},
	}),

	Entry("distributed workload with clean pod policy", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "cleanPodPolicy",
		expectedFname: "cleanPodPolicy_dist",

		distributed: &DistributedTest{
			numWorkers: 2,
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)),
				CleanPodPolicy:       *workloadsApi.NewNullableCleanPodPolicy(convutils.Ref(workloadsApi.CLEANPODPOLICY_NONE)),
			},
			masterSpecFname: "cleanPodPolicy_master",
		},
	}),

	Entry("distributed mpi with launcherCreationPolicy", ComposerTest{
		workloadType:   string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:      "launcherCreationPolicy",
		expectedFname:  "launcherCreationPolicy_dist",
		clusterVersion: "2.22.0",

		distributed: &DistributedTest{
			numWorkers: 2,
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework:      *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)),
				MpiLauncherCreationPolicy: *workloadsApi.NewNullableMpiLauncherCreationPolicy(convutils.Ref(workloadsApi.MPILAUNCHERCREATIONPOLICY_AT_STARTUP)),
			},
			masterSpecFname: "basic_master",
		},
	}),

	Entry("assets_dist", ComposerTest{
		workloadType:  string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		specFname:     "assets",
		expectedFname: "assets_tf",

		assetsInfo: &composer.ComplimentaryAssetInfo{
			Environment: workloadsApi.Environment{
				Name: "assets-worker-name",
				Id:   "assets-worker-id",
			},
			Datasources: composer.MapAssetKindToResources{
				composer_v2.AssetKind_HOST_PATH: []workloadsApi.Datasource{
					{
						Type: convutils.Ref(composer_v2.AssetKind_HOST_PATH),
						Id:   "hp1-id",
						Name: "hp1-name",
					},
					{
						Type: convutils.Ref(composer_v2.AssetKind_HOST_PATH),
						Id:   "hp2-id",
						Name: "hp2-name",
					},
				},
			},
		},

		distributed: &DistributedTest{
			numWorkers: 2,
			spec: workloadsApi.DistributedSpecSpec{
				DistributedFramework: *workloadsApi.NewNullableDistributedFramework(convutils.Ref(workloadsApi.DISTRIBUTEDFRAMEWORK_TF)),
			},
			masterSpecFname: "assets_master",

			masterAssetsInfo: &composer.ComplimentaryAssetInfo{
				Environment: workloadsApi.Environment{
					Name: "assets-master-name",
					Id:   "assets-master-id",
				},
				Datasources: composer.MapAssetKindToResources{
					composer_v2.AssetKind_HOST_PATH: []workloadsApi.Datasource{
						{
							Type: convutils.Ref(composer_v2.AssetKind_HOST_PATH),
							Id:   "hp-master-id",
							Name: "hp-master-name",
						},
					},
				},
			},
		},
	}),
}
