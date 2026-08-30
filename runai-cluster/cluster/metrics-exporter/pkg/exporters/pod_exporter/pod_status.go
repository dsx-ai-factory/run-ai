package pod_exporter

import v1 "k8s.io/api/core/v1"

type PodStatus int

const (
	// pending means the task is pending in the apiserver.
	pending PodStatus = 1 << iota

	// allocated means the scheduler assigns a host to it.
	allocated

	// pipelined means the scheduler assigns a host to wait for releasing resource.
	pipelined

	// binding means the scheduler send Bind request to apiserver.
	binding

	// bound means the task/Pod bounds to a host.
	bound

	// running means a task is running on the host.
	running

	// releasing means a task/pod is deleted.
	releasing

	// succeeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	succeeded

	// Failed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	failed

	// Unknown means the status of task/pod is unknown to the scheduler.
	unknown
)

var activeUsedStatuses = []PodStatus{allocated, pipelined, binding, bound, running, releasing}
var boundStatuses = []PodStatus{allocated, bound, running, releasing}

func getPodStatus(pod *v1.Pod) PodStatus {
	switch pod.Status.Phase {
	case v1.PodRunning:
		if pod.DeletionTimestamp != nil {
			return releasing
		}

		return running
	case v1.PodPending:
		if pod.DeletionTimestamp != nil {
			return releasing
		}

		if len(pod.Spec.NodeName) == 0 {
			return pending
		}
		return bound
	case v1.PodUnknown:
		return unknown
	case v1.PodSucceeded:
		return succeeded
	case v1.PodFailed:
		return failed
	}

	return unknown
}
