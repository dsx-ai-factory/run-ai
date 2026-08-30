package distributed_to_kubeflow

import (
	. "github.com/onsi/ginkgo/v2"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// mnnvlTests exercise the MNNVL plugin flag for distributed-training. The flag is
// only emitted when it carries a meaningful value:
//   - cluster distGPUNetworkAcceleration on + no label -> mnnvl: Required
//   - a concrete run.ai/MNNVL: None label is authoritative -> mnnvl: None
//
// The default-off, no-label case (mnnvl left unset) is covered by the unit tests
// in mnnvl_test.go and by every other golden, which carry no mnnvl field.
var mnnvlTests = []TableEntry{
	Entry("mnnvl: cluster acceleration on defaults to Required", ComposerTest{
		framework:                     workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname:               "basic",
		masterSpecFname:               "master_no_gpu",
		expectedFname:                 "mnnvl_cluster_on_default_required",
		clusterGPUNetworkAcceleration: true,
	}),

	Entry("mnnvl: None label is honored", ComposerTest{
		framework:                     workloadsApi.DISTRIBUTEDFRAMEWORK_MPI,
		workerSpecFname:               "basic",
		masterSpecFname:               "master_no_gpu",
		expectedFname:                 "mnnvl_label_none",
		mnnvlLabel:                    "None",
		clusterGPUNetworkAcceleration: true,
	}),
}
