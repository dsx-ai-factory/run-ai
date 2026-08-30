package services

import (
	"fmt"
)

type ClusterDoesNotSupportProjectDeletionErr struct{}

func (e *ClusterDoesNotSupportProjectDeletionErr) Error() string {
	return fmt.Sprintf("cluster does not support cascade project deletion. Cascade deletion is supported for clusters of version >= %s", ProjectDeletionMinVersion)
}

type ProjectIsAlreadyDeletingErr struct{}

func (e *ProjectIsAlreadyDeletingErr) Error() string {
	return "project is already in deleting phase, to delete it anyway please set force to true"
}

type ProjectIsNotEmptyErr struct {
	Workloads   []string
	Datavolumes []string
}

func (e *ProjectIsNotEmptyErr) Error() string {
	messages := "The project couldn't be deleted because the following needs to be deleted first:\n"
	if len(e.Workloads) > 0 {
		if len(e.Workloads) == 1 {
			messages += fmt.Sprintf("Workload: %s\n", e.Workloads[0])
		} else {
			messages += fmt.Sprintf("Workload: %s +%d more\n", e.Workloads[0], len(e.Workloads)-1)
		}
	}

	if len(e.Datavolumes) > 0 {
		if len(e.Datavolumes) == 1 {
			messages += fmt.Sprintf("Datavolume: %s\n", e.Datavolumes[0])
		} else {
			messages += fmt.Sprintf("Datavolume: %s +%d more\n", e.Datavolumes[0], len(e.Datavolumes)-1)
		}
	}

	return messages
}

func (e *ProjectIsNotEmptyErr) IsEmpty() bool {
	if len(e.Workloads) > 0 {
		return false
	}
	if len(e.Datavolumes) > 0 {
		return false
	}

	return true
}
