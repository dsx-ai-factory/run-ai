package reconcilers

import (
	"context"
	"log"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/metrics_exporter/pkg/event_handlers"
)

type PodReconciler struct {
	client        client.Client
	eventHandlers []event_handlers.EventHandler
}

func NewPodReconciler(client client.Client) *PodReconciler {
	eventHandlers := []event_handlers.EventHandler{
		event_handlers.NewPodGpuInfoEventHandler(client),
	}

	return &PodReconciler{
		client:        client,
		eventHandlers: eventHandlers,
	}
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Pod{}).
		WithEventFilter(FilterOnlyRunAiPod()).
		Complete(r)
}

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Printf("Reconciling pod %s/%s", req.Name, req.Namespace)
	pod := &v1.Pod{}
	err := r.client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("Pod %s/%s not found, was probably deleted", req.Namespace, req.Name)
			identifier := event_handlers.ObjectIdentifier{Name: req.Name, Namespace: req.Namespace}
			for _, handler := range r.eventHandlers {
				handler.ResetMetrics(identifier)
			}
			return ctrl.Result{}, nil
		}

		log.Printf("Failed to get pod %s/%s, error: %s", req.Namespace, req.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: 1}, err
	}

	for _, handler := range r.eventHandlers {
		err = handler.SetMetrics(ctx, pod)
		if err != nil {
			log.Printf("Failed to set metrics for pod %s/%s, err: %v", req.Namespace, req.Name, err)
			return ctrl.Result{Requeue: true, RequeueAfter: 1}, err
		}
	}

	return ctrl.Result{}, nil
}

func FilterOnlyRunAiPod() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isRunAiPod(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isRunAiPod(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isRunAiPod(e.Object)
		},
	}
}

func isRunAiPod(object client.Object) bool {
	pod, isPod := object.(*v1.Pod)
	if isPod {
		return pod.Spec.SchedulerName == constants.RunaiScheduler
	}
	return false
}
