package controllers

import (
	runaiclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AddInferenceHandlerToCache(backendClient *runaiclient.BackendClient, cachedClient cachedclient.CachedClient,
	workloadHandler *handlers.WorkloadHandler) *cachedclient.RateLimitedController {
	inferenceInformer := cachedClient.GetInformer(cachedclient.InferenceWorkloadType)
	if inferenceInformer == nil {
		zap.S().Warn("InferenceWorkload API is not served; skipping its controller")
		return nil
	}
	inferenceHandler := handlers.NewInferenceHandler(workloadHandler)
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.InferenceWorkloadRL, inferenceHandler)

	_, err := inferenceInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(object any) bool {
			return handlers.FilterWorkloads(object.(*v2alpha1.InferenceWorkload).Spec.WorkloadFields)
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(object any) {
				obj := object.(*v2alpha1.InferenceWorkload)
				if handlers.FilterWorkloads(obj.Spec.WorkloadFields) {
					genericControllerErrorHandler(workloadHandler.HandleInferenceAdd(obj),
						constants.AddOperation, "inferenceworkload", obj)
				}
			},
			DeleteFunc: func(object any) {
				obj := object.(*v2alpha1.InferenceWorkload)
				if handlers.FilterWorkloads(obj.Spec.WorkloadFields) {
					rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj))
					genericControllerErrorHandler(workloadHandler.HandleRunaiWorkloadDelete(object.(client.Object)),
						constants.DeleteOperation, "inferenceworkload", obj)
				}
			},
			UpdateFunc: func(oldObject any, newObject any) {
				newInferenceWorkload := newObject.(*v2alpha1.InferenceWorkload)
				oldInferenceWorkload := oldObject.(*v2alpha1.InferenceWorkload)
				if handlers.FilterWorkloads(newInferenceWorkload.Spec.WorkloadFields) {
					if newInferenceWorkload.Status.Phase != oldInferenceWorkload.Status.Phase {
						genericControllerErrorHandler(workloadHandler.HandleInferenceUpdate(newInferenceWorkload),
							constants.UpdateOperation, "inferenceworkload", newInferenceWorkload)
					} else {
						rateLimitedController.AddRateLimited(objectKeyFromObject(newInferenceWorkload))
					}
				}
			},
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to inference informer, error: %s", err.Error())
	}

	zap.S().Debug("The inference controller has started")
	return nil
}
