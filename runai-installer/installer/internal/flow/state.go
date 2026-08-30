package flow

type InstallFlowState int

const (
	InstallFlowStateNotStarted InstallFlowState = iota
	InstallFlowStateInProgress
	InstallFlowStateCompleted
	InstallFlowStateFailed
)
