package handlers

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/common/getters"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/runaistatus"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	statuscontract "github.com/run-ai/runai/runai-cluster/cluster/workload-common/status/contract"
	monorepocommonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewJobTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *JobTopOwnerHandler {
	return &JobTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *JobTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &batchv1.Job{} })
}

func (h *JobTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	job, ok := object.(*batchv1.Job)
	if !ok {
		return fmt.Errorf("failed converting object to batchv1.Job")
	}

	commonparsers.FillExternalWorkloadDataFromPodTemplateSpec(ctx, job.ObjectMeta, job.GroupVersionKind().GroupKind(), &job.Spec.Template,
		nil, h.cachedClient, h.draParser, externalWorkload)

	externalWorkload.Spec.RequestedPods.Parallelism = job.Spec.Parallelism
	externalWorkload.Spec.RequestedPods.Completions = job.Spec.Completions

	// v2 Run:ai workloads run as a wlr-owned Job; reproduce the lost v1 phase/message.
	if wlrName := workloadRunnerName(job); wlrName != "" {
		h.setRunaiWorkloadObservedStatus(ctx, job, wlrName, externalWorkload)
	} else {
		setJobObservedStatus(&job.Status, &externalWorkload.Status.ObservedStatus)
	}

	return nil
}

func (h *JobTopOwnerHandler) setRunaiWorkloadObservedStatus(ctx context.Context, job *batchv1.Job,
	wlrName string, externalWorkload *v1alpha1.ExternalWorkload) {
	observedStatus := &externalWorkload.Status.ObservedStatus

	// Native Job status first (K8sStatus + conditions); we override only Phase + Message.
	setJobObservedStatus(&job.Status, observedStatus)

	pods, err := getters.GetPodsOfWorkload(ctx,
		topowner.GetRunaiWorkloadId(job.GetUID(), job.GetAnnotations()), job.GetNamespace(), h.cachedClient)
	if err != nil {
		// Don't publish a pod-derived phase from a failed fetch; let cluster-sync apply the wlr phase.
		// Clear any stale authoritative flag from a previous reconcile so cluster-sync isn't blocked.
		log.Warn().Msgf("runaistatus: failed getting pods of Job <%s/%s>: %s",
			job.GetNamespace(), job.GetName(), err.Error())
		delete(externalWorkload.Annotations, monorepocommonconstants.EwiStatusAnnotation)
		return
	}

	wlr := h.getWorkloadRunner(ctx, wlrName, job.GetNamespace())
	if wlr == nil {
		// WLR fetch failed; keep the Job-derived phase and let cluster-sync apply the wlr phase.
		// Clear any stale authoritative flag from a previous reconcile so cluster-sync isn't blocked.
		delete(externalWorkload.Annotations, monorepocommonconstants.EwiStatusAnnotation)
		return
	}

	phase, message := runaistatus.ComputeStatus(job, pods, wlr)
	observedStatus.Phase = phase
	observedStatus.Message = message

	if group, desc, ok := runaistatus.ComputeRuntimeIssue(job, pods, wlr); ok {
		if desc != "" {
			observedStatus.Message = desc
		} else {
			desc = statuscontract.MsgGenericJobFailed
		}
		observedStatus.Conditions = append(observedStatus.Conditions, v1alpha1.ObservedCondition{
			Type:    monorepocommonconstants.WorkloadIssueConditionType,
			Status:  corev1.ConditionTrue,
			Reason:  group,
			Message: desc,
		})
	}

	if externalWorkload.Annotations == nil {
		externalWorkload.Annotations = map[string]string{}
	}
	externalWorkload.Annotations[monorepocommonconstants.EwiStatusAnnotation] = "true"
}

func (h *JobTopOwnerHandler) getWorkloadRunner(ctx context.Context, name, namespace string) *dromosv1alpha1.WorkloadRunner {
	wlr := &dromosv1alpha1.WorkloadRunner{}
	// Set GVK so the cached client resolves the type; the wlr watcher informs the cache.
	wlr.SetGroupVersionKind(dromosv1alpha1.SchemeGroupVersion.WithKind(dromosv1alpha1.WorkloadRunnerKind))
	if err := h.cachedClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, wlr); err != nil {
		log.Warn().Msgf("runaistatus: failed getting WorkloadRunner <%s/%s>: %s", namespace, name, err.Error())
		return nil
	}
	return wlr
}

func workloadRunnerName(job *batchv1.Job) string {
	for _, ownerRef := range job.GetOwnerReferences() {
		if ownerRef.Kind == dromosv1alpha1.WorkloadRunnerKind &&
			ownerRef.APIVersion == dromosv1alpha1.SchemeGroupVersion.String() {
			return ownerRef.Name
		}
	}
	return ""
}

func setJobObservedStatus(jobStatus *batchv1.JobStatus,
	observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	observedStatus.Phase = v1alpha1.ExWUnknownPhase
	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range jobStatus.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})

		if condition.Type == batchv1.JobSuspended && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWStoppedPhase
			observedStatus.K8sStatus = commonparsers.SuspendedStatus
			observedStatus.Message = condition.Message
			continue
		}

		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWCompletedPhase
			observedStatus.K8sStatus = commonparsers.SucceededStatus
			observedStatus.Message = condition.Message
			continue
		}

		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWFailedPhase
			observedStatus.K8sStatus = commonparsers.FailedStatus
			observedStatus.Message = condition.Message
		}
	}

	observedStatus.Conditions = observedConditions

	if observedStatus.Phase != v1alpha1.ExWUnknownPhase {
		return
	}

	if jobStatus.Active > 0 {
		observedStatus.K8sStatus = commonparsers.ActiveStatus
		if jobStatus.Ready != nil {
			if *jobStatus.Ready > 0 {
				observedStatus.Phase = v1alpha1.ExWRunningPhase
			} else {
				observedStatus.Phase = v1alpha1.ExWPendingPhase
			}
		} else {
			observedStatus.Phase = v1alpha1.ExWRunningPhase
		}
		return
	}

	observedStatus.K8sStatus = string(v1alpha1.ExWCreatingPhase)
	observedStatus.Phase = v1alpha1.ExWCreatingPhase
}

func (h *JobTopOwnerHandler) Filter(object client.Object) bool {
	job, ok := object.(*batchv1.Job)
	if !ok {
		return false
	}
	return job.Spec.Template.Spec.SchedulerName == constants.RunaiSchedulerName
}
