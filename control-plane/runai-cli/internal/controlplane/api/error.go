package api

import (
	"fmt"
)

type MultiWorkloadsError struct {
	Workloads []MinimalWorkloadInfo
}

func NewMultiWorkloadsError(workloads []MinimalWorkloadInfo) error {
	return MultiWorkloadsError{Workloads: workloads}
}

func (e MultiWorkloadsError) Error() string {
	return fmt.Sprintf("Found multiple workloads. Found: %v", e.Workloads)
}

type MinimalWorkloadInfo struct {
	Name    string
	Project string
	Kind    string
	ID      string
}
