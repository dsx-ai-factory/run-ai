package controller

import (
	"context"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/controller/datavolume"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/controller/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

type DataVolumeController struct {
	client               client.Client
	datavolumeReconciler *datavolume.DatavolumeReconciler
}

func NewDataVolumeController(client client.Client) *DataVolumeController {
	return &DataVolumeController{
		client:               client,
		datavolumeReconciler: datavolume.NewDatavolumeReconciler(client),
	}
}

func (c *DataVolumeController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := c.indexFields(ctx, mgr)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DataVolume{}).
		Watches(&v1.PersistentVolumeClaim{}, handler.EnqueueRequestsFromMapFunc(c.datavolumeReconciler.MapPvcToDatavolumeEvent)).
		Watches(&v1.PersistentVolume{}, handler.EnqueueRequestsFromMapFunc(c.datavolumeReconciler.MapPvToDatavolumeEvent))

	builder = builder.Watches(&kaiorg.Project{}, handler.EnqueueRequestsFromMapFunc(c.datavolumeReconciler.MapProjectToDatavolumeEvent))

	return builder.
		WithEventFilter(c.datavolumeReconciler.DatavolumePredicateFuncs()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			config.Config().BaseControllerDelay, config.Config().MaxControllerDelay)}).
		Complete(c)
}

func (c *DataVolumeController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Info().Msgf("Reconciling datavolume %s/%s", req.Name, req.Namespace)

	dataVolume := &v1alpha1.DataVolume{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "datavolume", dataVolume, c.client)
	if err != nil {
		return ctrl.Result{Requeue: true}, nil
	} else if !isObjectFound {
		return ctrl.Result{Requeue: false}, nil
	}

	if dataVolume.DeletionTimestamp.IsZero() {
		if err = c.datavolumeReconciler.AddControllerAsFinalizerIfNeeded(ctx, dataVolume); err != nil {
			return ctrl.Result{}, err
		}

		err = c.datavolumeReconciler.Reconcile(ctx, dataVolume)
	} else {
		err = c.datavolumeReconciler.Finalize(ctx, dataVolume)
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (c *DataVolumeController) indexFields(ctx context.Context, mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		ctx, &kaiorg.Project{},
		datavolume.KaiProjectsParentFieldIndexName, ProjectsParentIndexer)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed indexing project field: %s, error: %s",
			datavolume.KaiProjectsParentFieldIndexName, err.Error())
		return err
	}

	return nil
}

func ProjectsParentIndexer(object client.Object) (indexedKeys []string) {
	project, ok := object.(*kaiorg.Project)
	if !ok {
		log.Error().Msgf("ProjectsParentIndexer: Cannot convert object to *kaiorg.Project: %v", object)
		return indexedKeys
	}

	return []string{project.Spec.Parent}
}
