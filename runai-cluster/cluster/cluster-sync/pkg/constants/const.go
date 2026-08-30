package constants

import (
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

// Status contains the same statuses that appear in the scheduler - update both if needed.
var Status = struct {
	Running   string
	Pending   string
	Succeeded string
	Deleted   string
	Failed    string
	TimedOut  string
	Timeout   string
	Preempted string
	Unknown   string
}{
	Running:   "Running",
	Pending:   "Pending",
	Succeeded: "Succeeded",
	Deleted:   "Deleted",
	Failed:    "Failed",
	TimedOut:  "TimedOut",
	Timeout:   "Timeout",
	Preempted: "Preempted",
	Unknown:   "Unknown",
}

const BEARER = "BEARER "

const RunaiNamespacePrefix = "runai-"

const ParallelismDefault = 1

const CompletionsDefault = 1

const PodByPodGroupIndexerName = "podByPodGroupIndexer"
const NonRunaiPodByNodeNameIndexer = "nonRunaiPodByNodeNameIndexer"
const PodBySchedulerNameIndexer = "spec.schedulerName"

const RunaiWorkloadByUsageIndexer = "runaiWorkloadByUsageIndexer"
const RunaiWorkloadRelevantUsage = "Submit"

// GpuResourceName is the resource name for NVIDIA GPUs.
const GpuResourceName = clustercommonconstants.NvidiaGpuResourceName
const AmdGpuResourceName = clustercommonconstants.AmdGpuResourceName

const GpuMemoryResourceNameDeprecated = PodGpuMemoryRequest

const GpuMemoryResourceName = "run.ai/gpu.memory"
const MigGpuResourceName = "nvidia.com/mig"

var AssignedToNodeStatuses = []string{Status.Running, Status.Failed, Status.Succeeded}

// DeleteOperationString is the operation string for policy delete.
const DeleteOperationString = "delete"
const UpdateOperationString = "update"
const AddOperationString = "add"
const ApplyAction = "apply"
const LowTrainingPolicyString = "trainingpolicy"
const LowInteractivePolicyString = "interactivepolicy"
const LowInferencePolicyString = "inferencepolicy"
const LowDistributedPolicyString = "distributedpolicy"

const TrainingWorkload = "TrainingWorkload"
const InteractiveWorkload = "InteractiveWorkload"
const DistributedWorkload = "DistributedWorkload"
const InferenceWorkload = "InferenceWorkload"
const ExternalWorkload = "ExternalWorkload"
const DistributedInferenceWorkload = "DistributedInferenceWorkload"

// NodeTopologyMismatchCondition Node Pool condition types
const NodeTopologyMismatchCondition = "NodeTopologyMismatch"

const ClusterInfoConfigMapName = "runai-public"
