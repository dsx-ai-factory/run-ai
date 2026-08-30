package queuesManager

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/kubeConfig"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	dw2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iw2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	tw2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"time"
)

// HandleWorkloadTask calculates the time to accomplish at for the workload
// and then
// * in case "time to accomplish at" in the past, will set the workload to inactive
// * else will push the workload to the queue
func HandleWorkloadTask(ctx context.Context, workloadCache *util.WorkloadCache, k8sClient client.Client, promQuerier prometheus.Querier, workloadTask *WorkloadTask, projectQueue *ProjectQueue, skipCache bool) error {
	zap.S().Debugf("Start handling workload task, workload name (%s) workload type (%s)", workloadTask.WorkloadName, workloadTask.WorkloadType)
	var nearestEventAt, err = calculateNearestEvent(ctx, k8sClient, promQuerier, projectQueue, workloadTask)
	if err != nil {
		return fmt.Errorf( //nolint:staticcheck
			"Failed to calculate accomplish time - workloadTask (%s), workload (%s), type (%s), error (%s)",
			workloadTask.JobName, workloadTask.WorkloadName, workloadTask.WorkloadType, err.Error())
	}

	if nearestEventAt.NextEvent.IsZero() {
		// time to accomplish could be zero if the project values are zero (not set yet or unset)
		// then no need to push to queue
		zap.S().Debugf("Workload (%s) next event is zero, skip workload task ", workloadTask.WorkloadName)
		return nil
	}

	workloadTask.NextEventAt = nearestEventAt.NextEvent
	taskAction := &workloadTask.TaskAction

	// if the time to accomplish at is in the past, we will set the workload to inactive
	timeNow := time.Now()
	if workloadTask.NextEventAt.Before(timeNow) {
		jw, err := fetchWorkload(ctx, k8sClient, workloadTask)

		if err != nil {
			return fmt.Errorf("failed to get workload (%s) (%s) (%s)", workloadTask.WorkloadName, workloadTask.WorkloadType, err)
		}

		// verify that the workload is active before setting it to inactive
		if util.IsActiveWorkload(jw.Spec.Active) {
			if taskAction.GetListIndex() == DeactivateEvent {
				err = projectQueue.WorkloadValidator.Validate(jw)
				if err != nil {
					createErrorNotificationEvent(projectQueue.Recorder, jw, workloadTask.JobName, nearestEventAt, err)
					zap.S().Infof("Workload '%s' validation failed, skipping deactivation. validation error: (%s)", workloadTask.WorkloadName, err)
					return nil
				}

				jw.Spec.Active = field_types.NewOptionalBoolField(ptr.To(false))

				err = updateWorkload(ctx, k8sClient, workloadTask, jw)
				if err != nil {
					zap.S().Warnf("Failed to update workload (set active to false) (%s) (%s)", workloadTask.WorkloadName, err)
					return err
				}
				createStoppedEvent(projectQueue.Recorder, jw, workloadTask, nearestEventAt)
				zap.S().Infof("Success to update workload (%s) (set active to false), nearestEventAt (%s)", workloadTask.WorkloadName, workloadTask.NextEventAt.String())
			} else {
				err = projectQueue.WorkloadValidator.Validate(jw)
				if err != nil {
					zap.S().Infof(
						"Workload '%s' validation failed, skipping notification (set task percentage %d), nearestEventAt (%s). validation error: (%s)",
						taskAction.GetListIndex(), workloadTask.NextEventAt.String(), workloadTask.WorkloadName, err)
				} else {
					// only notify about termination if the workload can be terminated
					createNotificationEvent(projectQueue.Recorder, jw, workloadTask.JobName, nearestEventAt)
					zap.S().Infof("Success to notify workload (%s) (set task percentage %d), nearestEventAt (%s)", workloadTask.WorkloadName, taskAction.GetListIndex(), workloadTask.NextEventAt.String())
				}

				taskAction.IncrementIndex()
				zap.S().Infof("Workload (%s) added to queue with event percentile: %d", workloadTask.WorkloadName, taskAction.GetListIndex())
				nearestEventAt, _ = calculateNearestEvent(ctx, k8sClient, promQuerier, projectQueue, workloadTask)
				workloadTask.NextEventAt = nearestEventAt.NextEvent
				projectQueue.PushToQueue(workloadTask)
			}
		} else {
			zap.S().Debugf("Workload (%s) is not active, skip workload task ", workloadTask.WorkloadName)
		}

		// we either suspended the workload or it's already suspended -> no need for further handling
		return nil
	} else {
		zap.S().Debugf("Workload (%s) nearestEventAt in (%v) minutes, pushing workloadTask again to Queue", workloadTask.WorkloadName, workloadTask.NextEventAt.Sub(timeNow).Minutes())
	}

	if !skipCache {
		if workloadCache == nil {
			// this case should not happen - just in case
			zap.S().Warnf("WorkloadCache is nil, while handling workload task, worklod (%s) will not be added to cache ", workloadTask.WorkloadName)
		} else {
			// add to cache
			// if we failed to add to cache, we will not push to queue (already in cache)
			err = (*workloadCache).Add(projectQueue.ProjectName, workloadTask.JobUuid, workloadTask.StartTime)
			if err != nil {
				zap.S().Debugf("Workload (%s) failed to add to cache, skip workload task, error (%s) ", workloadTask.WorkloadName, err)
				return nil
			}
			zap.S().Debugf("Workload (%s) added to cache", workloadTask.WorkloadName)
		}
	}

	zap.S().Infof("Workload (%s) returned to queue", workloadTask.WorkloadName)
	projectQueue.PushToQueue(workloadTask)
	return nil
}

func fetchWorkload(ctx context.Context, k8sClient client.Client, workloadTask *WorkloadTask) (*jobs.JobWorkload, error) {
	var err error
	jw := &jobs.JobWorkload{}

	// A runner is projected onto JobWorkload so the active-phase guard reads a
	// fresh Spec.Active; a vanished runner yields an empty (inactive) JobWorkload.
	if workloadTask.WorkloadType == constants.KindWorkloadRunner {
		runner, getErr := kubeConfig.GetWLR(ctx, k8sClient, workloadTask.Namespace, workloadTask.WorkloadName)
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				// Runner gone: mark inactive so the caller skips it instead of
				// treating an empty (Active==nil) JobWorkload as active.
				jw.Spec.Active = field_types.NewOptionalBoolField(ptr.To(false))
				return jw, nil
			}
			return jw, getErr
		}
		if convErr := util.NewRunnerConverter(k8sClient).Convert(&runner, jw); convErr != nil {
			return jw, convErr
		}
		return jw, nil
	}

	switch workloadTask.WorkloadType {
	case constants.KindInteractiveWorkload:
		var iw iw2.InteractiveWorkload
		iw, err = kubeConfig.GetIW(ctx, k8sClient, workloadTask.Namespace, workloadTask.WorkloadName)
		if err == nil {
			iw.CopyTo(jw)
		}
	case constants.KindTrainingWorkload:
		var tw tw2.TrainingWorkload
		tw, err = kubeConfig.GetTW(ctx, k8sClient, workloadTask.Namespace, workloadTask.WorkloadName)
		if err == nil {
			tw.CopyTo(jw)
		}
	case constants.KindDistributedWorkload:
		var dw dw2.DistributedWorkload
		dw, err = kubeConfig.GetDW(ctx, k8sClient, workloadTask.Namespace, workloadTask.WorkloadName)
		if err == nil {
			dw.CopyTo(jw)
		}
	default:
		err = fmt.Errorf("unknown workload type %s", workloadTask.WorkloadType)
	}

	return jw, err
}

func updateWorkload(ctx context.Context, k8sClient client.Client, workloadTask *WorkloadTask, jw *jobs.JobWorkload) error {
	var err error

	// SuspendRunner acts on the runner's template, not on jw; the caller's
	// jw.Spec.Active = false is irrelevant for runners.
	if workloadTask.WorkloadType == constants.KindWorkloadRunner {
		return util.SuspendRunner(ctx, k8sClient, workloadTask.Namespace, workloadTask.WorkloadName)
	}

	switch workloadTask.WorkloadType {
	case constants.KindInteractiveWorkload:
		iw := iw2.NewFromJobWorkload(jw)
		err = kubeConfig.UpdateIW(ctx, k8sClient, *iw)
	case constants.KindTrainingWorkload:
		tw := tw2.NewFromJobWorkload(jw)
		err = kubeConfig.UpdateTW(ctx, k8sClient, *tw)
	case constants.KindDistributedWorkload:
		dw := dw2.NewFromJobWorkload(jw)
		err = kubeConfig.UpdateDW(ctx, k8sClient, *dw)
	default:
		err = fmt.Errorf("unknown workload type %s", workloadTask.WorkloadType)
	}

	return err
}

// PushProjectWorkloadsToProjectQueue pushes workload tasks to the queue if
// * the workload is not in the cache
// * the workload is still active
// PushProjectWorkloadsToProjectQueue takes the project name and its resolved namespace as
// primitives so it serves both the kai and legacy paths (both expose Name/Status.Namespace
// identically).
func PushProjectWorkloadsToProjectQueue(ctx context.Context, workloadCache *util.WorkloadCache, k8sClient client.Client,
	prometheusQuerier prometheus.Querier, projectName string, namespace string, projectQueue *ProjectQueue, supportedJobKinds map[string]client.Object, legacyWorkloadsEnabled bool) error {
	if namespace == "" {
		return fmt.Errorf("namespace is empty from project status (%s)", projectName)
	}

	if legacyWorkloadsEnabled {
		// List and enqueue legacy workload CRs only before the 3.0 cutover.
		twList, err := kubeConfig.ListTW(ctx, k8sClient, namespace)
		if err != nil {
			zap.S().Errorf("Err %+v ", err)
			return err
		}
		for _, tw := range twList.Items {
			pushWorkloadToQueue(&tw, ctx, workloadCache, k8sClient, prometheusQuerier, projectQueue, supportedJobKinds)
		}

		iwList, err := kubeConfig.ListIW(ctx, k8sClient, namespace)
		if err != nil {
			zap.S().Errorf("err %+v ", err)
			return err
		}
		for _, iw := range iwList.Items {
			pushWorkloadToQueue(&iw, ctx, workloadCache, k8sClient, prometheusQuerier, projectQueue, supportedJobKinds)
		}

		dwList, err := kubeConfig.ListDW(ctx, k8sClient, namespace)
		if err != nil {
			zap.S().Errorf("err %+v ", err)
			return err
		}
		for _, dw := range dwList.Items {
			pushWorkloadToQueue(&dw, ctx, workloadCache, k8sClient, prometheusQuerier, projectQueue, supportedJobKinds)
		}
	}

	// WorkloadRunner enumeration: the kind is always present at the 3.0+ legacy
	// workload cutover. Before 3.0, its presence retains the existing disable
	// flag behavior from NewRuntimeRules.
	if _, ok := supportedJobKinds[constants.KindWorkloadRunner]; ok {
		wlrList, listErr := kubeConfig.ListWLR(ctx, k8sClient, namespace)
		if listErr != nil {
			zap.S().Errorf("err %+v ", listErr)
			return listErr
		}
		for i := range wlrList.Items {
			pushWorkloadToQueue(&wlrList.Items[i], ctx, workloadCache, k8sClient, prometheusQuerier, projectQueue, supportedJobKinds)
		}
	}

	return nil
}

// convertWorkloadForEnumeration projects an enumerated workload onto a
// JobWorkload. WorkloadRunners use the runner converter (reads GPU/auto-delete
// metadata off Spec.Template); legacy CRDs use the kind-agnostic free function.
// Unlike the watch-driven path, the lifecycle annotation helper is NOT invoked
// here — it runs when the watch fires its first event for each runner.
func convertWorkloadForEnumeration(k8sClient client.Client, w client.Object) (*jobs.JobWorkload, error) {
	if runner, ok := w.(*dromosv1alpha1.WorkloadRunner); ok {
		var jw jobs.JobWorkload
		if err := util.NewRunnerConverter(k8sClient).Convert(runner, &jw); err != nil {
			return nil, err
		}
		return &jw, nil
	}
	return util.ConvertToJobWorkload(w)
}

func pushWorkloadToQueue(w client.Object, ctx context.Context, workloadCache *util.WorkloadCache, k8sClient client.Client,
	prometheusQuerier prometheus.Querier, projectQueue *ProjectQueue, supportedJobKinds map[string]client.Object) {
	jw, err := convertWorkloadForEnumeration(k8sClient, w)
	if err != nil {
		zap.S().Errorf("err while handling WorkloadTask %+v ", err)
		return
	}

	jobUuid, jobName, startTime, validationError := util.GetWorkloadValues(jw, supportedJobKinds)
	if validationError != nil {
		zap.S().Debugf("Workload (%s) is not valid to be added to queue, validation error (%s)", jw.Name, validationError.Error())
		return
	}

	// Derive the limit-governing category for runners (and drop out-of-scope
	// ones), mirroring the watch path's scope gate.
	category := ""
	if runner, ok := w.(*dromosv1alpha1.WorkloadRunner); ok {
		derivedCategory, skipReason, inScope := util.ClassifyRunner(runner)
		if !inScope {
			zap.S().Debugf("WorkloadRunner %s/%s skipped during enumeration: reason=%s", runner.Namespace, runner.Name, skipReason)
			return
		}
		category = derivedCategory
	}

	action := NewActionList(projectQueue.EventList, 0, 0, nil)
	workloadId, err := util.GetWorkloadID(w)
	if err != nil {
		zap.S().Errorf("err while handling WorkloadTask %+v ", err)
		return
	}
	workloadTask := &WorkloadTask{ // just a dummy WorkloadTask
		StartTime:            startTime.Time,
		WorkloadName:         jobName,
		WorkloadID:           workloadId,
		ProjectName:          projectQueue.ProjectName,
		Namespace:            jw.Namespace,
		JobUuid:              jobUuid,
		JobName:              jobName,
		WorkloadType:         jw.Kind,
		Category:             category,
		TaskAction:           *action,
		WorkloadRequestedGpu: util.HasRequestedGpu(jw),
		IsPreemptible:        util.IsJobPreemptible(jw),
	}
	err = HandleWorkloadTask(ctx, workloadCache, k8sClient, prometheusQuerier, workloadTask, projectQueue, true)
	if err != nil {
		zap.S().Errorf("err while handling WorkloadTask %+v ", err)
	}
}

// calculateNearestEvent - calculate the time to accomplish at for the workload
// the time to accomplish at is the earlier of the two:
// * the time to accomplish at by the time limit
// * the time to accomplish at by the max idle time

func calculateNearestEvent(ctx context.Context, _ client.Client, prometheusQuerier prometheus.Querier, projectQueue *ProjectQueue, workloadTask *WorkloadTask) (EventClock, error) {
	workloadTimeoutNearestEvent := EventClock{NextEventType: TaskTimerTimeout}
	workloadIdleNearestEvent := EventClock{NextEventType: TaskTimerIdle}

	timeLimitSecs, idleLimitSecs, err := getLimitsByWorkloadType(projectQueue, workloadTask)
	if err != nil {
		zeroTimerArray := EventClock{}
		return zeroTimerArray, err
	}

	if timeLimitSecs > 0 {
		now := metav1.Now()

		allowedDuration := time.Duration(timeLimitSecs) * time.Second
		durationSinceStart := now.Sub(workloadTask.StartTime)
		remainingDuration := allowedDuration - durationSinceStart
		zap.S().Infof(
			"%s timelimit (minutes) name: %s, allowedDuration: %f, durationSinceStart: %f, remainingDuration: %f ",
			workloadTask.WorkloadType, workloadTask.JobName, allowedDuration.Minutes(),
			durationSinceStart.Minutes(), remainingDuration.Minutes())

		workloadTimeoutNearestEvent.DeactivateEvent = time.Now().Add(remainingDuration)
		workloadTask.TaskAction.SetEventType(workloadTimeoutNearestEvent.NextEventType)
		// calculate the percentile value
		workloadTimeoutNearestEvent.SetNextEventTime(workloadTask)
	}
	if idleLimitSecs > 0 && workloadTask.WorkloadRequestedGpu {
		workloadIdleNearestEvent.DeactivateEvent = time.Time{}
		idleDurationSinceStart, err := prometheusQuerier.GetWorkloadIdleDuration(ctx, workloadTask.WorkloadName, string(workloadTask.WorkloadID))
		if err == nil {
			if idleDurationSinceStart == nil {
				idleDurationSinceStart = ptr.To(time.Duration(0))
			}

			allowedDuration := time.Duration(idleLimitSecs) * time.Second
			remainingDuration := time.Duration(allowedDuration.Seconds()-idleDurationSinceStart.Seconds()) * time.Second

			zap.S().Infof(
				"%s idlelimit (minutes) name: %s, allowedDuration: %f, idledurationSinceStart: %f, remainingDuration: %f ",
				workloadTask.WorkloadType, workloadTask.JobName, allowedDuration.Minutes(),
				idleDurationSinceStart.Minutes(), remainingDuration.Minutes())
			workloadIdleNearestEvent.DeactivateEvent = time.Now().Add(remainingDuration)
			workloadTask.TaskAction.SetEventType(workloadIdleNearestEvent.NextEventType)

			resetIdleNotificationIfNeeded(workloadTask, workloadIdleNearestEvent.DeactivateEvent)
			// calculate the percentile value
			workloadIdleNearestEvent.SetNextEventTime(workloadTask)
		} else {
			zap.S().Debugf("Failed to get idle duration for interactive workloadTask (%s) (%s)", workloadTask.JobName, err)
		}
	}

	earlierTime := GetTheEarlier(workloadTimeoutNearestEvent, workloadIdleNearestEvent)
	workloadTask.TaskAction.SetEventType(earlierTime.NextEventType)
	return earlierTime, nil
}

func getLimitsByWorkloadType(projectQueue *ProjectQueue, workloadTask *WorkloadTask) (timeLimit int64, idleLimit int64, err error) {
	switch workloadTask.WorkloadType {
	case constants.KindInteractiveWorkload, constants.KindTrainingWorkload, constants.KindDistributedWorkload:
		timeLimitSecs, idleLimitSecs := limitsForCategory(projectQueue, workloadTask.WorkloadType, workloadTask.IsPreemptible)
		return timeLimitSecs, idleLimitSecs, nil
	case constants.KindWorkloadRunner:
		// A runner is governed by the category derived from its wrapped Kind.
		// An empty Category yields zero limits (observed but not enforced).
		switch workloadTask.Category {
		case constants.KindInteractiveWorkload, constants.KindTrainingWorkload, constants.KindDistributedWorkload:
			timeLimitSecs, idleLimitSecs := limitsForCategory(projectQueue, workloadTask.Category, workloadTask.IsPreemptible)
			return timeLimitSecs, idleLimitSecs, nil
		default:
			return 0, 0, nil
		}
	default:
		return 0, 0, fmt.Errorf("unknown workload type %s", workloadTask.WorkloadType)
	}
}

// limitsForCategory returns the project time-limit and idle-limit (in seconds)
// for one of the three legacy workload categories.
func limitsForCategory(projectQueue *ProjectQueue, category string, isPreemptible bool) (timeLimitSecs int64, idleLimitSecs int64) {
	switch category {
	case constants.KindInteractiveWorkload:
		timeLimitSecs = projectQueue.InteractiveJobTimeLimitSecs
		if isPreemptible {
			idleLimitSecs = projectQueue.InteractivePreemptibleJobMaxIdleDurationSecs
		} else {
			idleLimitSecs = projectQueue.InteractiveJobMaxIdleDurationSecs
		}
	case constants.KindTrainingWorkload, constants.KindDistributedWorkload:
		timeLimitSecs = projectQueue.TrainingJobTimeLimitSecs
		idleLimitSecs = projectQueue.TrainingJobMaxIdleDurationSecs
	}
	return timeLimitSecs, idleLimitSecs
}
