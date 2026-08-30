package composer

import (
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func IsMainNode(replicaType *workloadsApi.ReplicaType) bool {
	return replicaType == nil || *replicaType == workloadsApi.REPLICATYPE_WORKER
}
