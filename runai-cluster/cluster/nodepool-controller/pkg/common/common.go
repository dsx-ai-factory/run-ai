package common

const (
	UnschedulableNodesMessage = "Unschedulable Nodes"
	SchedulerNotReadyMessage  = "Scheduler is not ready"
	NodesDrainedMessage       = "Nodes Being Drained"
	ErrorDeletingMessage      = "Error Deleting NodePool"

	ProjectsReferencingNodePoolMessage = "NodePool is referenced by the following project(s): %s"

	NodesAssignedToThisNodePoolWaitingForDrainMessage      = "The following node(s) are assigned to this node pool and will be added once they finish draining: %s"
	NodesAssignedToDifferentNodePoolWaitingForDrainMessage = "The following node(s) within the node pool aren't ready because they are assigned to a different node pool but haven't been drained yet: %s"
	NodesNotReadyMessage                                   = "The following node(s) within the node pool aren't ready: %s"

	IsDeletingPhaseField                      = "status.phase.isDeleting"
	PodRunningWithRunaiSchedulerNodeNameField = "spec.runningNodeName"
	NameField                                 = "metadata.name"
)
