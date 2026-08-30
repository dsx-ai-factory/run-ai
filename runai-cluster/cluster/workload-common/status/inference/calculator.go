// Package inference calculates inference workload phases without performing Kubernetes I/O.
package inference

import (
	"fmt"
	"strings"

	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	corev1 "k8s.io/api/core/v1"
)

type ServiceCondition struct {
	Status  corev1.ConditionStatus
	Message string
}

type ServiceState struct {
	LatestCreatedRevision string
	LatestReadyRevision   string
	Conditions            []ServiceCondition
}

type Input struct {
	Service           *ServiceState
	Pods              []corev1.Pod
	BuildError        *string
	MarkedForDeletion bool
}

type Result struct {
	Phase   statuscontract.Phase
	Message string
}

type podPhase string

const (
	podPhaseNone              podPhase = ""
	podPhaseContainerCreating podPhase = "ContainerCreating"
	podPhaseImagePull         podPhase = "ErrImagePull"
	podPhaseRunning           podPhase = "Running"
	podPhasePending           podPhase = "Pending"
)

type serviceIssue string

const (
	serviceIssueImagePull serviceIssue = "image-pull"
	serviceIssueNameLong  serviceIssue = "name-too-long"
	serviceIssueTimeout   serviceIssue = "timeout"
)

func Calculate(input Input) Result {
	if input.MarkedForDeletion {
		return Result{Phase: statuscontract.Deleting}
	}
	if input.Service == nil {
		return Result{Phase: statuscontract.Creating}
	}

	created := input.Service.LatestCreatedRevision
	ready := input.Service.LatestReadyRevision
	podsPhase := determinePodsPhase(input.Pods)
	allReady := allPodsReady(input.Pods)
	issue, failed := analyzeConditions(input.Service.Conditions)
	updating := ready != "" && (ready != created || input.BuildError != nil)
	failureMessage := failureMessage(input.BuildError, issue, failed, updating)

	phase := statuscontract.Pending
	switch {
	case failureMessage != "" && (ready == "" || ready == created):
		if ready == "" {
			phase = statuscontract.Failed
		} else {
			phase = statuscontract.Degraded
		}
	case created == "":
		phase = statuscontract.Creating
	case ready != "" && (updating || allReady):
		phase = statuscontract.Ready
	case podsPhase == podPhaseRunning ||
		podsPhase == podPhaseContainerCreating ||
		podsPhase == podPhaseImagePull:
		phase = statuscontract.Initializing
	}

	message := ""
	switch phase {
	case statuscontract.Failed, statuscontract.Degraded:
		message = failureMessage
	case statuscontract.Ready:
		if updating {
			if failureMessage != "" {
				message = failureMessage
			} else {
				message = podMessage(podsPhase, true)
			}
		}
	case statuscontract.Initializing:
		message = podMessage(podsPhase, false)
	}

	return Result{Phase: phase, Message: message}
}

func analyzeConditions(conditions []ServiceCondition) (serviceIssue, bool) {
	var issue serviceIssue
	failed := false
	for _, condition := range conditions {
		if condition.Status != corev1.ConditionFalse {
			continue
		}
		failed = true
		message := strings.ToLower(condition.Message)
		switch {
		case strings.Contains(message, "unable to fetch image"):
			issue = serviceIssueImagePull
		case strings.Contains(message, "must be no more than 63 characters"),
			strings.Contains(message, "name part must be no more than"):
			issue = serviceIssueNameLong
		case issue == "":
			issue = serviceIssueTimeout
		}
	}
	return issue, failed
}

func failureMessage(buildError *string, issue serviceIssue, failed, updating bool) string {
	if buildError != nil {
		message := normalizePrerequisiteError(*buildError)
		if updating {
			return fmt.Sprintf("Workload is ready. Update failed: %s", message)
		}
		return fmt.Sprintf("Failed to create workload: %s", message)
	}
	if !failed {
		return ""
	}

	message := "Timed out waiting for the workload to become ready."
	switch issue {
	case serviceIssueImagePull:
		message = "Failed to pull container image."
	case serviceIssueNameLong:
		message = "The workload name is too long and causes derived resource names to exceed Kubernetes 63 character limit. Please use a shorter workload name."
	}
	if updating {
		return "Workload is ready. Update failed: " + message
	}
	return message
}

func normalizePrerequisiteError(message string) string {
	if !strings.Contains(message, "validation.webhook.serving.knative.dev") {
		return message
	}
	for _, fragment := range []string{
		"must not set the field(s)",
		"Persistent volume claim support is disabled",
		"Persistent volume write support is disabled",
		"EmptyDir volume support is disabled",
		"multi-container is off",
	} {
		if strings.Contains(message, fragment) {
			return message + ". Check the inference prerequisites in the Run:ai documentation"
		}
	}
	return message
}

func determinePodsPhase(pods []corev1.Pod) podPhase {
	if len(pods) == 0 {
		return podPhaseNone
	}

	allRunning := true
	hasContainerCreating := false
	hasPending := false
	hasImagePull := false
	for i := range pods {
		pod := &pods[i]
		if pod.Status.Phase != corev1.PodRunning {
			allRunning = false
		}

		podHasImagePull := false
		statuses := append([]corev1.ContainerStatus{}, pod.Status.InitContainerStatuses...)
		statuses = append(statuses, pod.Status.ContainerStatuses...)
		for _, status := range statuses {
			if status.State.Waiting == nil {
				continue
			}
			switch status.State.Waiting.Reason {
			case "ImagePullBackOff", "ErrImagePull", "ErrImageNeverPull":
				hasImagePull = true
				podHasImagePull = true
			}
		}

		if pod.Annotations[commonconstants.PodStatusAnnotation] == string(podPhaseContainerCreating) {
			hasContainerCreating = true
		} else if pod.Status.Phase == corev1.PodPending && !podHasImagePull {
			hasPending = true
		}
	}

	switch {
	case hasPending:
		return podPhasePending
	case hasImagePull:
		return podPhaseImagePull
	case hasContainerCreating:
		return podPhaseContainerCreating
	case allRunning:
		return podPhaseRunning
	default:
		return podPhaseNone
	}
}

func allPodsReady(pods []corev1.Pod) bool {
	for i := range pods {
		ready := false
		for _, condition := range pods[i].Status.Conditions {
			if condition.Type == corev1.PodReady {
				ready = condition.Status == corev1.ConditionTrue
				break
			}
		}
		if !ready {
			return false
		}
	}
	return true
}

func podMessage(phase podPhase, updating bool) string {
	messages := map[podPhase][2]string{
		podPhasePending: {
			"Workload is waiting for resources.",
			"Workload is ready. Update in progress: waiting for resources.",
		},
		podPhaseContainerCreating: {
			"Workload is pulling container images.",
			"Workload is ready. Update in progress: pulling container images.",
		},
		podPhaseImagePull: {
			"Failed to pull container image. Retrying...",
			"Workload is ready. Update in progress: failed to pull container image, retrying...",
		},
		podPhaseRunning: {
			"Container images pulled successfully. Workload is initializing.",
			"Workload is ready. Update in progress: container images pulled successfully, new revision is initializing.",
		},
	}
	pair, ok := messages[phase]
	if !ok {
		return ""
	}
	if updating {
		return pair[1]
	}
	return pair[0]
}
