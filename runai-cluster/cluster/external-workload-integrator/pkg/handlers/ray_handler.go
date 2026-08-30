package handlers

import (
	"context"
	"fmt"
	rayv1alpha1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	rayv1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/ray/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NOTE: rayv1 is identical to rayv1alpha1, except for the version string itself
// (for all: rayjob, raycluster, and rayservice).
// instead of using "github.com/ray-project/kuberay/ray-operator/apis/ray/v1" directly,
// for easier conversions, we created "external-workload-integrator/pkg/thirdpartyapi/ray/v1" that point to the
// v1alpha1 Spec and Status.

const rayHeadSpecName = "head"

// rayWorkerSpecName returns the SpecName used for a Ray worker group's nodeAffinity
// entry in the EW spec. Worker groups in a RayCluster are named, so each named
// group gets its own entry; unnamed groups fall back to a generic "worker" name.
func rayWorkerSpecName(groupName string) string {
	if groupName == "" {
		return v1alpha1.DefaultNodeAffinitySpecName
	}
	return "worker-" + groupName
}

//nolint:unused // reserved for future use
var rayJobStatusToExternalWorkloadPhase = map[rayv1alpha1.JobStatus]v1alpha1.ExternalWorkloadPhase{
	rayv1alpha1.JobStatusPending:   v1alpha1.ExWPendingPhase,
	rayv1alpha1.JobStatusRunning:   v1alpha1.ExWRunningPhase,
	rayv1alpha1.JobStatusStopped:   v1alpha1.ExWStoppedPhase,
	rayv1alpha1.JobStatusSucceeded: v1alpha1.ExWCompletedPhase,
	rayv1alpha1.JobStatusFailed:    v1alpha1.ExWFailedPhase,
	rayv1alpha1.JobStatus(""):      v1alpha1.ExWUnknownPhase,
}

var rayClusterStateToExternalWorkloadPhase = map[rayv1alpha1.ClusterState]v1alpha1.ExternalWorkloadPhase{
	rayv1alpha1.Ready:            v1alpha1.ExWRunningPhase,
	rayv1alpha1.Unhealthy:        v1alpha1.ExWDegradedPhase,
	rayv1alpha1.Failed:           v1alpha1.ExWFailedPhase,
	rayv1alpha1.ClusterState(""): v1alpha1.ExWUnknownPhase,
}

type RayJobTopOwnerHandler struct {
	cachedClient   client.Client
	draParser      *dra.DRAParser
	getEmptyObject func() client.Object
}

type RayClusterTopOwnerHandler struct {
	cachedClient   client.Client
	draParser      *dra.DRAParser
	getEmptyObject func() client.Object
}

type RayServiceTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewRayJobTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser, getEmptyObject func() client.Object) *RayJobTopOwnerHandler {
	return &RayJobTopOwnerHandler{
		cachedClient:   cachedClient,
		draParser:      draParser,
		getEmptyObject: getEmptyObject,
	}
}

func NewRayClusterTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser, getEmptyObject func() client.Object) *RayClusterTopOwnerHandler {
	return &RayClusterTopOwnerHandler{
		cachedClient:   cachedClient,
		draParser:      draParser,
		getEmptyObject: getEmptyObject,
	}
}

func NewRayServiceTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *RayServiceTopOwnerHandler {
	return &RayServiceTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *RayClusterTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, h.getEmptyObject)
}

func (h *RayServiceTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &rayv1alpha1.RayService{} })
}

func (h *RayJobTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, h.getEmptyObject)
}

func (h *RayJobTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	rayJob, ok := object.(*rayv1alpha1.RayJob)
	if !ok {
		rayJobV1, ok1 := object.(*rayv1.RayJob)
		if !ok1 {
			return fmt.Errorf("failed converting object to rayv1/rayv1alpha1.RayJob")
		}

		rayJob = &rayv1alpha1.RayJob{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RayJob",
				APIVersion: "ray.io/v1",
			},
			ObjectMeta: rayJobV1.ObjectMeta,
			Spec:       rayJobV1.Spec,
			Status:     rayJobV1.Status,
		}
	}

	if rayJob.Spec.RayClusterSpec != nil {
		parseRayClusterSpec(ctx, rayJob.ObjectMeta, rayJob.GroupVersionKind().GroupKind(), rayJob.Spec.RayClusterSpec,
			h.cachedClient, h.draParser, externalWorkload)
		parseRayClusterStatus(&externalWorkload.Status.ObservedStatus, &rayJob.Status.RayClusterStatus)
	}

	return nil
}

func (h *RayClusterTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	rayCluster, ok := object.(*rayv1alpha1.RayCluster)
	if !ok {
		rayClusterV1, ok1 := object.(*rayv1.RayCluster)
		if !ok1 {
			return fmt.Errorf("failed converting object to rayv1/rayv1alpha1.RayCluster")
		}

		rayCluster = &rayv1alpha1.RayCluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RayCluster",
				APIVersion: "ray.io/v1",
			},
			ObjectMeta: rayClusterV1.ObjectMeta,
			Spec:       rayClusterV1.Spec,
			Status:     rayClusterV1.Status,
		}
	}

	parseRayClusterSpec(ctx, rayCluster.ObjectMeta, rayCluster.GroupVersionKind().GroupKind(), &rayCluster.Spec,
		h.cachedClient, h.draParser, externalWorkload)
	parseRayClusterStatus(&externalWorkload.Status.ObservedStatus, &rayCluster.Status)

	return nil
}

func (h *RayServiceTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	rayService, ok := object.(*rayv1alpha1.RayService)
	if !ok {
		rayServiceV1, ok1 := object.(*rayv1.RayService)
		if !ok1 {
			return fmt.Errorf("failed converting object to rayv1/rayv1alpha1.RayService")
		}

		rayService = &rayv1alpha1.RayService{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RayService",
				APIVersion: "ray.io/v1",
			},
			ObjectMeta: rayServiceV1.ObjectMeta,
			Spec:       rayServiceV1.Spec,
			Status:     rayServiceV1.Status,
		}
	}

	parseRayClusterSpec(ctx, rayService.ObjectMeta, rayService.GroupVersionKind().GroupKind(), &rayService.Spec.RayClusterSpec,
		h.cachedClient, h.draParser, externalWorkload)
	parseRayClusterStatus(&externalWorkload.Status.ObservedStatus, &rayService.Status.ActiveServiceStatus.RayClusterStatus)

	return nil
}

func parseRayClusterSpec(ctx context.Context, objectMeta metav1.ObjectMeta, groupKind schema.GroupKind,
	rayClusterSpec *rayv1alpha1.RayClusterSpec, cachedClient client.Client, draParser *dra.DRAParser,
	externalWorkload *v1alpha1.ExternalWorkload) {
	commonparsers.FillExternalWorkloadDataFromPodTemplateSpec(ctx, objectMeta, groupKind,
		&rayClusterSpec.HeadGroupSpec.Template, ptr.To(rayHeadSpecName), cachedClient, draParser, externalWorkload)

	var totalReplicas int32 = 1 // the head pod
	var totalMinReplicas int32 = 0
	var totalMaxReplicas int32 = 0

	for _, workerGroupSpecs := range rayClusterSpec.WorkerGroupSpecs {
		workerReplicas := utils.DerefInt32(workerGroupSpecs.Replicas, 0)
		workerMinReplicas := utils.DerefInt32(workerGroupSpecs.MinReplicas, 0)

		totalReplicas += workerReplicas
		totalMinReplicas += workerMinReplicas
		totalMaxReplicas += utils.DerefInt32(workerGroupSpecs.MaxReplicas, 0)

		workerNumOfPods := workerReplicas
		if workerNumOfPods == 0 {
			workerNumOfPods = workerMinReplicas
		}

		podRequestedResources := commonparsers.GetPodRequestedResources(&workerGroupSpecs.Template)
		workerRequestedResources := commonparsers.MultiplyResources(workerNumOfPods, podRequestedResources)

		externalWorkload.Spec.RequestedResources = resources.SumResourceRequirements(
			externalWorkload.Spec.RequestedResources, workerRequestedResources)

		// Capture the worker group's nodeAffinity under its group name so each
		// worker group is distinguishable in the EW spec and downstream API.
		commonparsers.UpsertNodeAffinityFromPodTemplate(rayWorkerSpecName(workerGroupSpecs.GroupName),
			&workerGroupSpecs.Template, externalWorkload)

		commonparsers.UpsertStorageRefsFromPodTemplate(ctx, &workerGroupSpecs.Template, cachedClient, externalWorkload)
	}

	externalWorkload.Spec.RequestedPods.Replicas = &totalReplicas
	externalWorkload.Spec.RequestedPods.MinReplicas = utils.Int32PtrOrNil(totalMinReplicas)
	externalWorkload.Spec.RequestedPods.MaxReplicas = utils.Int32PtrOrNil(totalMaxReplicas)
}

func parseRayClusterStatus(observedStatus *v1alpha1.ExternalWorkloadObservedStatus, rayClusterStatus *rayv1alpha1.RayClusterStatus) {
	observedStatus.Phase = convertRayClusterStatusPhase(rayClusterStatus.State) //nolint:staticcheck // State is deprecated but still in use
	observedStatus.K8sStatus = string(rayClusterStatus.State)                   //nolint:staticcheck // State is deprecated but still in use
	observedStatus.Message = rayClusterStatus.Reason
}

//nolint:unused // reserved for future use
func convertRayJobStatusPhase(jobStatus rayv1alpha1.JobStatus) v1alpha1.ExternalWorkloadPhase {
	phase, found := rayJobStatusToExternalWorkloadPhase[jobStatus]
	if found {
		return phase
	}
	return v1alpha1.ExWUnknownPhase
}

func convertRayClusterStatusPhase(clusterState rayv1alpha1.ClusterState) v1alpha1.ExternalWorkloadPhase {
	phase, found := rayClusterStateToExternalWorkloadPhase[clusterState]
	if found {
		return phase
	}
	return v1alpha1.ExWUnknownPhase
}

func (h *RayJobTopOwnerHandler) Filter(object client.Object) bool {
	switch object.(type) {
	case *rayv1alpha1.RayJob:
		return filterRayCluster(object.(*rayv1alpha1.RayJob).Spec.RayClusterSpec)
	case *rayv1.RayJob:
		return filterRayCluster(object.(*rayv1.RayJob).Spec.RayClusterSpec)
	default:
		return false
	}
}

func (h *RayClusterTopOwnerHandler) Filter(object client.Object) bool {
	switch object.(type) {
	case *rayv1alpha1.RayCluster:
		return filterRayCluster(&object.(*rayv1alpha1.RayCluster).Spec)
	case *rayv1.RayCluster:
		return filterRayCluster(&object.(*rayv1.RayCluster).Spec)
	default:
		return false
	}
}

func (h *RayServiceTopOwnerHandler) Filter(object client.Object) bool {
	switch object.(type) {
	case *rayv1alpha1.RayService:
		return filterRayCluster(&object.(*rayv1alpha1.RayService).Spec.RayClusterSpec)
	case *rayv1.RayService:
		return filterRayCluster(&object.(*rayv1.RayService).Spec.RayClusterSpec)
	default:
		return false
	}
}

func filterRayCluster(rayClusterSpec *rayv1alpha1.RayClusterSpec) bool {
	return rayClusterSpec != nil && rayClusterSpec.HeadGroupSpec.Template.Spec.SchedulerName == constants.RunaiSchedulerName
}
