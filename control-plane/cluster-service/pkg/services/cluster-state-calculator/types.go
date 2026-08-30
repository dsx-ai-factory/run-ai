package clusterstatecalculator

type ClusterState = string

const (
	ClusterStateWaitingToConnect     ClusterState = "WaitingToConnect"
	ClusterStateConnected            ClusterState = "Connected"
	ClusterStateDisconnected         ClusterState = "Disconnected"
	ClusterStateMissingPrerequisites ClusterState = "MissingPrerequisites"
	ClusterStateServiceIssues        ClusterState = "ServiceIssues"
	ClusterStateUnknown              ClusterState = "Unknown"
)
