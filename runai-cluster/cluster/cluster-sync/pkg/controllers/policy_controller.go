package controllers

import (
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	dpv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	infpv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	ipv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tpv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

func AddPolicyHandlerToCache(cachedClient cachedclient.CachedClient, policyHandler *handlers.PolicyHandler) {

	registerInteractivePolicy(cachedClient, policyHandler)
	registerTrainingPolicy(cachedClient, policyHandler)
	registerDistributedPolicy(cachedClient, policyHandler)
	registerInferencePolicy(cachedClient, policyHandler)

	zap.S().Debug("Policies handling finished")
}

func registerInteractivePolicy(cachedClient cachedclient.CachedClient, policyHandler *handlers.PolicyHandler) {
	interactivePolicyInformer := cachedClient.GetInformer(cachedclient.InteractivePolicyType)
	if interactivePolicyInformer == nil {
		zap.S().Warn("InteractivePolicy API is not served; skipping its controller")
		return
	}

	ipHandler := handlers.InteractivePolicyHandler{PolicyHandler: policyHandler}
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.InteractivePolicyRL, &ipHandler)

	_, err := interactivePolicyInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			genericControllerErrorHandler(policyHandler.HandleAdd(obj.(*ipv2.InteractivePolicy)),
				constants.AddOperationString, constants.LowInteractivePolicyString, obj.(*ipv2.InteractivePolicy))
		},
		DeleteFunc: func(obj any) {
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj.(*ipv2.InteractivePolicy)))
			genericControllerErrorHandler(policyHandler.HandleDelete(obj.(*ipv2.InteractivePolicy)),
				constants.DeleteOperationString, constants.LowInteractivePolicyString, obj.(*ipv2.InteractivePolicy))
		},
		UpdateFunc: func(oldObj any, newObj any) {
			rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*ipv2.InteractivePolicy)))
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to interactive policy informer, error: %s", err.Error())
	}
}

func registerTrainingPolicy(cachedClient cachedclient.CachedClient, policyHandler *handlers.PolicyHandler) {
	trainingPolicyInformer := cachedClient.GetInformer(cachedclient.TrainingPolicyType)
	if trainingPolicyInformer == nil {
		zap.S().Warn("TrainingPolicy API is not served; skipping its controller")
		return
	}

	tpHandler := handlers.TrainingPolicyHandler{PolicyHandler: policyHandler}
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.TrainingPolicyRL, &tpHandler)

	_, err := trainingPolicyInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			genericControllerErrorHandler(policyHandler.HandleAdd(obj.(*tpv2.TrainingPolicy)),
				constants.AddOperationString, constants.LowTrainingPolicyString, obj.(*tpv2.TrainingPolicy))
		},
		DeleteFunc: func(obj any) {
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(obj.(*tpv2.TrainingPolicy)))
			genericControllerErrorHandler(policyHandler.HandleDelete(obj.(*tpv2.TrainingPolicy)),
				constants.DeleteOperationString, constants.LowTrainingPolicyString, obj.(*tpv2.TrainingPolicy))
		},
		UpdateFunc: func(oldObj any, newObj any) {
			rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*tpv2.TrainingPolicy)))
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to training policy informer, error: %s", err.Error())
	}
}

func registerInferencePolicy(cachedClient cachedclient.CachedClient, policyHandler *handlers.PolicyHandler) {
	inferencePolicyInformer := cachedClient.GetInformer(cachedclient.InferencePolicyType)
	if inferencePolicyInformer == nil {
		zap.S().Warn("InferencePolicy API is not served; skipping its controller")
		return
	}

	tpHandler := handlers.InferencePolicyHandler{PolicyHandler: policyHandler}
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.InferencePolicyRL, &tpHandler)

	_, err := inferencePolicyInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			object := obj.(*infpv2.InferencePolicy)
			genericControllerErrorHandler(policyHandler.HandleAdd(object), constants.AddOperationString, constants.LowInferencePolicyString, object)
		},
		DeleteFunc: func(obj any) {
			object := obj.(*infpv2.InferencePolicy)
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(object))
			genericControllerErrorHandler(policyHandler.HandleDelete(object), constants.DeleteOperationString, constants.LowInferencePolicyString, object)
		},
		UpdateFunc: func(oldObj any, newObj any) {
			rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*infpv2.InferencePolicy)))
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to inference policy informer, error: %s", err.Error())
	}
}

func registerDistributedPolicy(cachedClient cachedclient.CachedClient, policyHandler *handlers.PolicyHandler) {
	distributedPolicyInformer := cachedClient.GetInformer(cachedclient.DistributedPolicyType)
	if distributedPolicyInformer == nil {
		zap.S().Warn("DistributedPolicy API is not served; skipping its controller")
		return
	}

	dpHandler := handlers.DistributedPolicyHandler{PolicyHandler: policyHandler}
	rateLimitedController := cachedClient.NewRateLimitedController(cachedclient.DistributedPolicyRL, &dpHandler)

	_, err := distributedPolicyInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			genericControllerErrorHandler(policyHandler.HandleAdd(obj.(*dpv2.DistributedPolicy)), constants.AddOperationString, constants.LowDistributedPolicyString, obj.(*dpv2.DistributedPolicy))
		},
		DeleteFunc: func(obj any) {
			dp := obj.(*dpv2.DistributedPolicy)
			rateLimitedController.MarkItemDeleted(objectKeyFromObject(dp))
			genericControllerErrorHandler(policyHandler.HandleDelete(dp), constants.DeleteOperationString, constants.LowDistributedPolicyString, dp)
		},
		UpdateFunc: func(oldObj any, newObj any) {
			rateLimitedController.AddRateLimited(objectKeyFromObject(newObj.(*dpv2.DistributedPolicy)))
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to distributed policy informer, error: %s", err.Error())
	}
}
