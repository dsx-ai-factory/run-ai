package revision

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/getters"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerUtil "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	logRevision = logger.Log.WithName("reconcile.revision")
)

// Reconciler reconciles a servingv1.Revision object
type Reconciler struct {
	client.Client
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconcile int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&servingv1.Revision{},
			builder.WithPredicates(
				revisionLabelPredicate(),
				revisionConditionChangedPredicate(),
			),
		).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(onlyRevisionPods()),
		).
		WithOptions(controllerUtil.Options{MaxConcurrentReconciles: maxConcurrentReconcile,
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				constants.RateLimiterBaseDelay, constants.RateLimiterMaxDelay),
		}).
		Complete(r)
}

func onlyRevisionPods() func(ctx context.Context, obj client.Object) []reconcile.Request {
	// returns a function that filters pods and returns the revision that the pod belongs to
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		// filter pods that are not part of the revision
		if obj.GetLabels()[commonconstants.KnativeRevisionLabel] != "" {
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Name:      obj.GetLabels()[commonconstants.KnativeRevisionLabel],
					Namespace: obj.GetNamespace(),
				}},
			}
		}
		return []reconcile.Request{}
	}
}

func revisionConditionChangedPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldRevision, okOld := e.ObjectOld.(*servingv1.Revision)
			newRevision, okNew := e.ObjectNew.(*servingv1.Revision)
			if !okOld || !okNew {
				return false
			}
			return !reflect.DeepEqual(oldRevision.Status.Conditions, newRevision.Status.Conditions)
		},
	}
}

func revisionLabelPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		revision, ok := obj.(*servingv1.Revision)
		if !ok {
			return false
		}
		return commonconstants.IsManagedInferenceRevision(revision.Labels)
	})
}

func NewSimpleReconcile(client client.Client) *Reconciler {
	return &Reconciler{
		Client: client,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger.IntoContext(ctx, logRevision)

	resourceId := fmt.Sprintf("%s:%s", req.Namespace, req.Name)
	logRevision.Info("reconcile started", "id", resourceId)

	//
	//    obtain the resource being reconciled
	//
	revision := &servingv1.Revision{}

	err := r.Client.Get(ctx, req.NamespacedName, revision)
	if err != nil {
		//
		//   there was an error to get the resource from k8s
		//
		statusError := err.(*k8sErrors.StatusError)
		if statusError == nil || statusError.Status().Code != 404 {
			logRevision.Error(err, "failed to load the revision", "id", resourceId)
		} else {
			logRevision.Info("revision has been deleted", "id", resourceId)
		}

		//
		//    assume that the resource has been deleted with so just return normally
		//
		return ctrl.Result{}, nil
	}

	err = r.SetRevisionPhase(ctx, revision, resourceId)
	if err != nil {
		logRevision.Error(err, "failed to set revision phase data", "id", resourceId)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetRevisionPhase(ctx context.Context, rev *servingv1.Revision, resourceId string) error {
	phase, phaseMessage, err := r.GetRevisionPhase(ctx, rev)
	if err != nil {
		logRevision.Error(err, "failed to calculate revision phase", "id", resourceId)
		return err
	}

	if rev.Status.Annotations == nil {
		rev.Status.Annotations = map[string]string{}
	}

	if rev.Status.Annotations[constants.RunaiPhaseAnnotationKey] == phase.String() &&
		rev.Status.Annotations[constants.RunaiPhaseMessageAnnotationKey] == phaseMessage {
		logRevision.Info("No change in calculated phase, nothing to do", "id", resourceId)
		return nil
	}

	rev.Status.Annotations[constants.RunaiPhaseAnnotationKey] = phase.String()
	rev.Status.Annotations[constants.RunaiPhaseMessageAnnotationKey] = phaseMessage

	err = r.UpdateRevisionPhase(ctx, rev)
	if err != nil {
		logRevision.Error(err, "failed to update revision", "id", resourceId)
		return err
	}

	return nil
}

const (
	PendingPhaseMessage                   = "Revision is waiting to be scheduled"
	WaitingForResourcesPhaseMessage       = "Waiting for resources to be allocated"
	ResourcesAvailablePhaseMessage        = "Resources are available"
	ResourcesAllocationFailedPhaseMessage = "Failed to allocate resources"
	WaitForHealthPhaseMessage             = "Waiting for the container to become healthy"
	ContainerIsHealthyPhaseMessage        = "Container is healthy"
	ContainerIsNotHealthyPhaseMessage     = "Container is not healthy"
)

func (r *Reconciler) GetRevisionPhase(ctx context.Context, rev *servingv1.Revision) (workloadcommon.Phase, string, error) {
	// If the revision is deleted, mark it as Terminating.
	if !rev.DeletionTimestamp.IsZero() {
		return workloadcommon.PhaseTerminating, "", nil
	}

	// Fetch pods associated with the revision
	pods, err := getters.GetPodsOfRevision(ctx, rev.Name, rev.Namespace, r.Client)
	if err != nil {
		return workloadcommon.PhaseFailed, "", err
	}

	somePodRunning := false
	// Check if there is at least 1 pending pod
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodPending {
			return workloadcommon.PhasePending, PendingPhaseMessage, nil
		}

		if pod.Status.Phase == corev1.PodRunning {
			somePodRunning = true
		}
	}

	readyCondition := rev.Status.GetCondition(apis.ConditionReady)
	resourcesAvailable := rev.Status.GetCondition(servingv1.RevisionConditionResourcesAvailable)
	containerHealthy := rev.Status.GetCondition(servingv1.RevisionConditionContainerHealthy)

	if rev.IsFailed() {
		if readyCondition != nil {
			return workloadcommon.PhaseFailed, "Revision failed: " + readyCondition.Reason, nil
		}
		return workloadcommon.PhaseFailed, "", nil
	}

	if rev.IsReady() && somePodRunning {
		return workloadcommon.PhaseRunning, "", nil
	}

	// Default phase is Pending
	phase := workloadcommon.PhasePending
	phaseMessage := PendingPhaseMessage

	// Determine phase based on conditions
	if readyCondition == nil {
		phase = workloadcommon.PhaseCreating
		phaseMessage = "Revision is being created"
	} else {
		switch readyCondition.Status {
		case corev1.ConditionUnknown:
			if containerHealthy != nil {
				switch containerHealthy.Status {
				case corev1.ConditionUnknown:
					phase = workloadcommon.PhaseInitializing
					phaseMessage = WaitForHealthPhaseMessage
				case corev1.ConditionTrue:
					phase = workloadcommon.PhaseInitializing
					phaseMessage = ContainerIsHealthyPhaseMessage
				case corev1.ConditionFalse:
					phase = workloadcommon.PhaseInitializing
					phaseMessage = ContainerIsNotHealthyPhaseMessage
				}
			} else if resourcesAvailable != nil {
				switch resourcesAvailable.Status {
				case corev1.ConditionUnknown:
					phase = workloadcommon.PhaseInitializing
					phaseMessage = WaitingForResourcesPhaseMessage
				case corev1.ConditionTrue:
					phase = workloadcommon.PhaseInitializing
					phaseMessage = ResourcesAvailablePhaseMessage
				case corev1.ConditionFalse:
					phase = workloadcommon.PhaseInitializing
					phaseMessage = ResourcesAllocationFailedPhaseMessage
				}
			}

		case corev1.ConditionTrue:
			if somePodRunning {
				phase = workloadcommon.PhaseRunning
				phaseMessage = ""
			} else {
				phase = workloadcommon.PhaseStopped
				phaseMessage = ""
			}
		case corev1.ConditionFalse:
			phase = workloadcommon.PhaseFailed
			phaseMessage = "Revision failed: " + readyCondition.Reason
		}
	}

	return phase, phaseMessage, nil
}
func (r *Reconciler) UpdateRevisionPhase(ctx context.Context, rev *servingv1.Revision) error {
	rev.Status.Annotations[constants.RunaiPhaseVersionAnnotationKey] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	patchBytes, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"annotations": rev.Status.Annotations},
	})
	if err != nil {
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)

	return r.Client.Status().Patch(ctx, rev, patch)
}
