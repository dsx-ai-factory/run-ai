package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemRunAiDistributedV1Mapping is the system seed for the fictive
// run.ai/Distributed/v1 GVK.
//
// The distributed workload manifest nests both replica specs under .spec:
//   - .spec.worker (DistributedSpecSpec) — the worker spec, includes the
//     distributed-specific fields (numWorkers, distributedFramework, min/maxReplicas, ...)
//   - .spec.master (MasterSpec) — optional master spec (plain per-pod spec,
//     same shape as the workspace spec, no distributed-specific fields)
//
// This mirrors the workloads service's UnifiedSpec ({worker, master}).
func BuildSystemRunAiDistributedV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("run.ai"), Kind: "Distributed", Version: "v1"},
		buildRunAiDistributedV1SpecSelectors()
}

func buildRunAiDistributedV1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		"workerSpec": {
			Description: convutils.Ref("the distributed worker spec (.spec.worker, DistributedSpecSpec)"),
			Query: `(
			  .spec.worker
			)`,
			FieldGroups: []string{SystemNativeSpecFgName, SystemNativeDistributedSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec.worker",
			},
		},

		"masterSpec": {
			Description: convutils.Ref("the distributed master spec (.spec.master, MasterSpec)"),
			Query: `(
			  .spec.master
			)`,
			FieldGroups: []string{SystemNativeSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec.master",
			},
		},
	}
}
