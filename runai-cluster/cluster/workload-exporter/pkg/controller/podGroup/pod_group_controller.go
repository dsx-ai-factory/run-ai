package podgroup

import (
	"context"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/rs/zerolog/log"

	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type PodGroupController struct {
	client client.Client
}

func NewPodGroupController(client client.Client) *PodGroupController {
	return &PodGroupController{
		client: client,
	}
}

func (c *PodGroupController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kaiv2alpha2.PodGroup{}).
		WithEventFilter(podGroupsMigEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *PodGroupController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling PodGroup %s/%s", req.Namespace, req.Name)

	podGroup := &kaiv2alpha2.PodGroup{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "PodGroup", podGroup, c.client)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	} else if !isObjectFound {
		metrics.ResetPodGroupMetrics(req.Name, req.Namespace)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetPodGroupMetrics(ctx, podGroup, c.client)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for PodGroup %s/%s, error: %s", podGroup.Namespace, podGroup.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	}
	return ctrl.Result{}, nil
}

func podGroupsMigEventFilter() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return hasMigAnnotation(e.Object.(*kaiv2alpha2.PodGroup))
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldPg := e.ObjectOld.(*kaiv2alpha2.PodGroup)
			newPg := e.ObjectNew.(*kaiv2alpha2.PodGroup)
			// Reconcile if MIG annotation was added, removed, or changed
			return hasMigAnnotation(oldPg) || hasMigAnnotation(newPg)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return hasMigAnnotation(e.Object.(*kaiv2alpha2.PodGroup))
		},
	}
}

func hasMigAnnotation(podGroup *kaiv2alpha2.PodGroup) bool {
	_, found := podGroup.Annotations[commonconstants.PodGroupRequestedMigDevicesAnnotation]
	return found
}
