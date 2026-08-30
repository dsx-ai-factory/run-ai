package reconcilers

import (
	"context"
	"log"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/metrics_exporter/pkg/event_handlers"
)

type PodGroupReconciler struct {
	client              client.Client
	eventHandlers       []event_handlers.EventHandler
	deleteEventHandlers []event_handlers.EventHandler
}

func NewPodGroupReconciler(client client.Client) *PodGroupReconciler {
	eventHandlers := []event_handlers.EventHandler{
		event_handlers.NewMigPodGroupEventHandler(client),
	}

	deleteEventHandlers := []event_handlers.EventHandler{}
	deleteEventHandlers = append(deleteEventHandlers, eventHandlers...)

	return &PodGroupReconciler{
		client:              client,
		eventHandlers:       eventHandlers,
		deleteEventHandlers: deleteEventHandlers,
	}
}

func (r *PodGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kaiv2alpha2.PodGroup{}).
		Complete(r)
}

func (r *PodGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Printf("Reconciling pod group %s/%s", req.Namespace, req.Name)
	podGroup := &kaiv2alpha2.PodGroup{}
	err := r.client.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, podGroup)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("Pod group %s/%s was not found, was probably deleted", req.Namespace, req.Name)
			for _, handler := range r.deleteEventHandlers {
				handler.ResetMetrics(event_handlers.ObjectIdentifier{Name: req.Name, Namespace: req.Namespace})
			}
			return ctrl.Result{}, nil
		}

		log.Printf("Failed to get pod group %s/%s", req.Namespace, req.Name)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 1,
		}, err
	}

	for _, handler := range r.eventHandlers {
		err = handler.SetMetrics(ctx, podGroup)
		if err != nil {
			log.Printf("Failed to set metrics for pod group %s/%s, err: %v", req.Namespace, req.Name, err)
			return ctrl.Result{Requeue: true, RequeueAfter: 1}, err
		}
	}
	return ctrl.Result{}, nil
}
