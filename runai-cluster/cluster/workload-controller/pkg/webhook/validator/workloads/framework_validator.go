package workloads

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DistributedAvailableErr  = "distributed job %s is not available, make sure that %s is installed on your cluster"
	FailedToCheckSuspendable = "workload is suspended, but failed to check if framework %s is suspendable. error: %s"
	NotSuspendable           = "workload is suspended, but framework %s is not suspendable"
)

type WorkloadFrameworkValidator struct {
	k8sClient              client.Client
	EnabledDistributedJobs map[string]controller.DistributedJobFrameworksStatus
	workloadType           string
}

func NewWorkloadFrameworkValidator(k8s client.Client, enabledDistributedJobs map[string]controller.DistributedJobFrameworksStatus) *WorkloadFrameworkValidator {
	if enabledDistributedJobs == nil {
		enabledDistributedJobs = make(map[string]controller.DistributedJobFrameworksStatus)
	}
	return &WorkloadFrameworkValidator{k8sClient: k8s, EnabledDistributedJobs: enabledDistributedJobs}
}

func (frame *WorkloadFrameworkValidator) Validate(_ context.Context, workload *jobs.JobWorkload) error {
	if !frame.isDistributedWorkload(workload) {
		return nil
	}

	if isAvailable, err := frame.isDistributedJobAvailable(workload); !isAvailable {
		return err
	}

	if isSuspended(workload) {
		suspendable, err := frame.isDistributedJobSuspendable(workload)
		if err != nil { //should not happen
			return fmt.Errorf(FailedToCheckSuspendable, workload.Spec.JobType, err.Error())
		}

		if !suspendable {
			return fmt.Errorf(NotSuspendable, workload.Spec.JobType)
		}
	}

	return nil
}

func (frame *WorkloadFrameworkValidator) isDistributedJobAvailable(workload *jobs.JobWorkload) (bool, error) {
	if _, ok := frame.EnabledDistributedJobs[workload.Spec.JobType]; !ok {
		return false, fmt.Errorf(DistributedAvailableErr, workload.Spec.JobType, workload.Spec.JobType)
	}
	return true, nil
}

func (frame *WorkloadFrameworkValidator) isDistributedWorkload(workload *jobs.JobWorkload) bool {
	return strings.Compare(workload.Kind, constants.KindDistributedWorkload) == 0
}

func isSuspended(workload *jobs.JobWorkload) bool {
	return !workload.Spec.Active.Deref()
}

func (frame *WorkloadFrameworkValidator) isDistributedJobSuspendable(workload *jobs.JobWorkload) (bool, error) {
	frameworkStatus, isAvailable := frame.EnabledDistributedJobs[workload.Spec.JobType]
	if !isAvailable {
		return false, fmt.Errorf(DistributedAvailableErr, workload.Spec.JobType, workload.Spec.JobType)
	}

	return frameworkStatus.SuspendEnabled, nil
}
