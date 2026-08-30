package workload

import (
	"context"

	"github.com/rs/zerolog/log"
	diwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type DistributedWorkloadController struct {
	client client.Client
}

func NewDistributedWorkloadController(client client.Client) *DistributedWorkloadController {
	return &DistributedWorkloadController{
		client: client,
	}
}

func (c *DistributedWorkloadController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&diwv2a1.DistributedWorkload{}).
		WithEventFilter(workloadsEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *DistributedWorkloadController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling distributedworkload %s/%s", req.Name, req.Namespace)

	distributedWorkload := &diwv2a1.DistributedWorkload{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "distributedworkload", distributedWorkload, c.client)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	} else if !isObjectFound || !distributedWorkload.DeletionTimestamp.IsZero() {
		metrics.ResetWorkloadMetrics(req.Name, req.Namespace, metrics.DistributedSource)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetDistributedWorkloadMetrics(ctx, distributedWorkload, c.client)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for distributed workload %s/%s, error: %s",
			distributedWorkload.Namespace, distributedWorkload.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	}
	return ctrl.Result{}, nil
}
