package controllers

import (
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AddPodGroupHandlerToCache(cachedClient cachedclient.CachedClient,
	trainingRLC, workspaceRLC, distributedRLC, inferenceRLC, externalRLC, distributedInferenceRLC *cachedclient.RateLimitedController) {
	podGroupInformer := cachedClient.GetInformer(cachedclient.PodGroupType)
	_, err := podGroupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj any, newObj any) {
			newPg := newObj.(*kaiv2alpha2.PodGroup)
			handleWorkloadIfNeeded(newPg, trainingRLC, workspaceRLC, distributedRLC, inferenceRLC, externalRLC, distributedInferenceRLC)
		},
	})

	if err != nil {
		zap.S().Fatalf("Failed to add event handler to pod group informer, error: %s", err.Error())
	}

	zap.S().Debug("PodGroup handling has started")
}

// handleWorkloadIfNeeded
// used to send SchedulingMessages for workloads from cluster-sync
func handleWorkloadIfNeeded(pg *kaiv2alpha2.PodGroup,
	trainingRLC, workspaceRLC, distributedRLC, inferenceRLC, externalRLC, distributedInferenceRLC *cachedclient.RateLimitedController) {
	workloadKind, workloadName := getWorkloadOfPodGroup(pg)
	if workloadKind == "" || workloadName == "" {
		zap.S().Debugf("PodGroup <%s/%s> does not have a workload kind or name", pg.Namespace, pg.Name)
		return
	}

	// handle workload of pod group only if it is unscheduleable
	if !parsers.IsPodGroupUnscheduleable(pg) {
		return
	}

	var rateLimitedControllerToUse *cachedclient.RateLimitedController = nil

	switch workloadKind {
	case constants.TrainingWorkload:
		rateLimitedControllerToUse = trainingRLC
	case constants.InteractiveWorkload:
		rateLimitedControllerToUse = workspaceRLC
	case constants.DistributedWorkload:
		rateLimitedControllerToUse = distributedRLC
	case constants.InferenceWorkload:
		rateLimitedControllerToUse = inferenceRLC
	case constants.DistributedInferenceWorkload:
		rateLimitedControllerToUse = distributedInferenceRLC
	case constants.ExternalWorkload:
		rateLimitedControllerToUse = externalRLC
	default:
		zap.S().Errorf("Workload kind %s is not supported", workloadKind)
	}

	if rateLimitedControllerToUse != nil {
		zap.S().Infof("PodGroup <%s/%s> is in Pending phase, updating its %s <%s>",
			pg.Namespace, pg.Name, workloadKind, workloadName)
		rateLimitedControllerToUse.AddRateLimited(client.ObjectKey{Name: workloadName, Namespace: pg.Namespace})
	}
}

// getWorkloadOfPodGroup - return workloadKind, workloadName
func getWorkloadOfPodGroup(pg *kaiv2alpha2.PodGroup) (string, string) {
	kind, name := getRunaiWorkloadOfPodGroup(pg)
	if kind != "" && name != "" {
		return kind, name
	}
	return getExternalWorkloadOfPodGroup(pg)
}

// getRunaiWorkloadOfPodGroup - return workloadKind, workloadName
func getRunaiWorkloadOfPodGroup(pg *kaiv2alpha2.PodGroup) (string, string) {
	workloadKind, ok := pg.Labels[constants.LabelWorkloadKind]
	if !ok {
		return "", ""
	}
	workloadName, ok := pg.Labels[constants.LabelWorkloadName]
	if !ok {
		return "", ""
	}
	return workloadKind, workloadName
}

// getExternalWorkloadOfPodGroup - return workloadKind, workloadName
func getExternalWorkloadOfPodGroup(pg *kaiv2alpha2.PodGroup) (string, string) {
	workloadName, ok := pg.Annotations[commonconstants.ExternalWorkloadName]
	if !ok {
		return "", ""
	}

	return constants.ExternalWorkload, workloadName
}
