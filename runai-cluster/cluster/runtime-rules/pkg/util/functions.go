package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	fields "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	gpuvalidator "github.com/run-ai/runai/runai-cluster/cluster/workload-common/gpu-validator"
	dw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	tw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/common/constants"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	UnknownKind    = "unknown"
	UnknownKindMsg = "unknown workload kind '%s'"
)

// GetWorkloadValues returns true if the workload is valid for workloadTask creation
//
//	like job uuid, job name and start time of the workload ...
func GetWorkloadValues(workload *jobs.JobWorkload, supportedJobKinds map[string]client.Object) (types.UID, string, *metav1.Time, error) {
	var (
		jobName         string
		jobUuid         types.UID
		validationError error
	)

	workloadName := workload.Name

	for _, ownedResource := range workload.Status.OwnedResources {
		if _, ok := supportedJobKinds[ownedResource.Kind.Kind]; ok {
			jobUuid = ownedResource.Uid
			jobName = ownedResource.Name
			break
		}
	}

	if jobUuid == "" || jobName == "" {
		validationError = fmt.Errorf("workload %s has no job uuid (%s) or job name (%s) ", workloadName, jobUuid, jobName)
		return "", "", nil, validationError
	}

	if !IsActiveWorkload(workload.Spec.Active) {
		validationError = fmt.Errorf("workload %s is not active ", workloadName)
		return "", "", nil, validationError
	}

	if !workload.Status.CompletionTime.IsZero() {
		validationError = fmt.Errorf("workload %s is completed ", workloadName)
		return "", "", nil, validationError
	}

	startTime := workload.Status.StartTime
	if startTime == nil || startTime.Time.IsZero() {
		validationError = fmt.Errorf("startTime is zero, skipping workload %s ", workloadName)
		return "", "", nil, validationError
	}

	return jobUuid, jobName, startTime, nil
}

func IsActiveWorkload(active *fields.BoolField) bool {
	return active == nil || ptr.Deref(active.Value, false)
}

func GetEnvOrDefaultInt(environmentVariableName string, defaultValue int) int {
	env := os.Getenv(environmentVariableName)
	ans, err := strconv.Atoi(env)
	if env == "" {
		zap.S().Debugf("environment variable %s was not provided, using default value %v instead", environmentVariableName, defaultValue)
		ans = defaultValue
	} else if err != nil {
		zap.S().Infof("environment variable %s has incorrect format, using default value %v instead, %v", environmentVariableName, defaultValue, err)
		ans = defaultValue
	}
	return ans
}

func GetEnvOrDefaultIntArray(environmentVariableName string, defaultValue []int) []int {
	env, found := os.LookupEnv(environmentVariableName)
	if !found {
		zap.S().Infof("environment variable %s is missing, using default value %v instead", environmentVariableName, defaultValue)
		return defaultValue
	}

	var ans []int
	if strings.Compare(env, "") == 0 {
		zap.S().Infof("environment variable %s is empty (%v) no event will be added", environmentVariableName, env)
		return ans
	}
	split := strings.Split(env, ",")
	for i := range split {
		strInt, err := strconv.Atoi(split[i])
		if err != nil {
			zap.S().Infof("environment variable %s failed to format %v, error:(%s) using default instead, %v", environmentVariableName, split[i], err, defaultValue)
			return defaultValue
		}
		ans = append(ans, strInt)
	}
	return ans
}

// ConvertToJobWorkload projects a legacy CRD (Training/Interactive/Distributed)
// onto the unified JobWorkload shape via its CopyTo method. WorkloadRunner has
// its own converter (util.RunnerConverter) because the projection reads
// GPU/auto-delete metadata off the runner's Spec.Template — see runner_converter.go.
func ConvertToJobWorkload(object runtime.Object) (*jobs.JobWorkload, error) {
	var (
		jobWorkload jobs.JobWorkload
		err         error
	)

	switch workloadObject := object.(type) {
	case *tw.TrainingWorkload:
		workloadObject.CopyTo(&jobWorkload)
	case *iw.InteractiveWorkload:
		workloadObject.CopyTo(&jobWorkload)
	case *dw.DistributedWorkload:
		workloadObject.CopyTo(&jobWorkload)
	default:
		unknownKind := UnknownKind
		if object != nil && object.GetObjectKind() != nil {
			unknownKind = object.GetObjectKind().GroupVersionKind().Kind
		}
		err = fmt.Errorf(UnknownKindMsg, unknownKind)
	}

	return &jobWorkload, err
}

// isRunnerActivePhase maps WorkloadRunnerPhase to the active/inactive split the
// runtime-rules' existing pipeline expects. Fail-closed: only explicitly-listed
// active phases return true; any new or unrecognised phase is treated as
// inactive (skip enforcement). Matches the design doc §8 risk-6 mitigation —
// a future phase added to the CRD should not silently enter the enforcement
// path.
func isRunnerActivePhase(phase dromosv1alpha1.WorkloadRunnerPhase) bool {
	switch phase {
	case dromosv1alpha1.WorkloadRunnerPhaseCreating,
		dromosv1alpha1.WorkloadRunnerPhasePending,
		dromosv1alpha1.WorkloadRunnerPhaseInitializing,
		dromosv1alpha1.WorkloadRunnerPhaseRunning,
		dromosv1alpha1.WorkloadRunnerPhaseDegraded:
		return true
	default:
		return false
	}
}

func GetWorkloadID(workload client.Object) (types.UID, error) {
	if workload == nil {
		return "", fmt.Errorf("workload object is Nil")
	}
	if annotation, exists := workload.GetAnnotations()[constants.ControlPlaneWorkloadIdAnnotationForWorkload]; exists {
		return types.UID(annotation), nil
	}
	return workload.GetUID(), nil
}

func HasRequestedGpu(workload *jobs.JobWorkload) bool {
	if workload == nil {
		return false
	}

	if gpuvalidator.HasRequestedGpus(&workload.Spec.ResourceAllocationFields) ||
		(workload.MasterSpec != nil && gpuvalidator.HasRequestedGpus(&workload.MasterSpec.ResourceAllocationFields)) {
		return true
	}

	return false
}

func IsJobPreemptible(jw *jobs.JobWorkload) bool {
	// just for sanitation
	if jw == nil {
		return false
	}

	if jw.Kind != commonconstants.KindInteractiveWorkload {
		return true
	}

	if jw.Spec.Preemptibility.IsValueInitialized() {
		return jw.Spec.Preemptibility.Deref() == field_groups.Preemptible
	}

	// default for interactive preemptible is false
	return jw.Spec.Preemptible.Deref() //nolint:staticcheck
}
