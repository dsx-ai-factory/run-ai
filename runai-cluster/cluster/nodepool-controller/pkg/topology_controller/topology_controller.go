package topology_controller

import (
	"context"
	"encoding/json"
	"time"

	grovev1alpha1 "github.com/ai-dynamo/grove/operator/api/core/v1alpha1"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	controllerName       = "kai-topology-controller"
	rateLimiterBaseDelay = 500 * time.Millisecond
	rateLimiterMaxDelay  = 30 * time.Second
)

type KaiTopologyController struct {
	client client.Client
}

func NewKaiTopologyController(client client.Client) *KaiTopologyController {
	return &KaiTopologyController{
		client: client,
	}
}

// Reconcile handles KAI Topology reconciliation.
// For each KAI Topology, it looks up the corresponding Grove ClusterTopologyBinding
// using the naming convention mapping. If a matching Grove topology exists,
// it annotates the KAI Topology with the Grove topology name and resource version;
// otherwise, it removes any stale Grove annotations.
func (tc *KaiTopologyController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	kaiTopologyName := req.Name

	kaiTopology := &kaiv1alpha1.Topology{}
	if err := tc.client.Get(ctx, types.NamespacedName{Name: kaiTopologyName}, kaiTopology); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Find the grove topology that maps to this KAI topology
	groveTopology, err := tc.findGroveTopologyForKai(ctx, kaiTopologyName)
	if err != nil {
		return ctrl.Result{}, err
	}

	if groveTopology != nil {
		return ctrl.Result{}, tc.setAnnotations(ctx, kaiTopology, groveTopology)
	}

	return ctrl.Result{}, tc.removeAnnotations(ctx, kaiTopology)
}

// SetupWithManager registers the controller with the manager.
// It watches KAI Topology as the primary resource and
// grove ClusterTopologyBinding as a secondary resource (mapping grove events to KAI Topology reconcile requests).
func (tc *KaiTopologyController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&kaiv1alpha1.Topology{}).
		// Watch Grove ClusterTopologyBinding as a secondary resource so that changes to Grove topologies
		// trigger reconciliation of the corresponding KAI Topology, keeping annotations in sync.
		Watches(&grovev1alpha1.ClusterTopologyBinding{},
			handler.EnqueueRequestsFromMapFunc(tc.mapGroveToKaiReconcileRequest),
		).
		WithOptions(controller.Options{
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				rateLimiterBaseDelay, rateLimiterMaxDelay)}).
		Complete(tc)
}

// mapGroveToKaiReconcileRequest maps a grove ClusterTopologyBinding event to a KAI Topology
// reconcile request using the magic mapping function.
func (tc *KaiTopologyController) mapGroveToKaiReconcileRequest(_ context.Context, obj client.Object) []reconcile.Request {
	kaiTopologyName := groveToKaiTopologyName(obj)
	if kaiTopologyName == "" {
		log.Debug().Msgf("Grove topology <%s> does not map to any KAI topology", obj.GetName())
		return nil
	}

	log.Debug().Msgf("Grove topology <%s> maps to KAI topology <%s>, enqueueing reconcile",
		obj.GetName(), kaiTopologyName)
	return []reconcile.Request{
		{NamespacedName: types.NamespacedName{Name: kaiTopologyName}},
	}
}

// findGroveTopologyForKai lists all grove topologies and returns the one
// that maps to the given KAI topology, or nil if none found.
func (tc *KaiTopologyController) findGroveTopologyForKai(ctx context.Context, kaiTopologyName string) (*grovev1alpha1.ClusterTopologyBinding, error) {
	groveTopologies := &grovev1alpha1.ClusterTopologyBindingList{}
	if err := tc.client.List(ctx, groveTopologies); err != nil {
		log.Error().Msgf("Failed to list grove topologies: %v", err)
		return nil, err
	}

	for i := range groveTopologies.Items {
		grove := &groveTopologies.Items[i]
		if groveToKaiTopologyName(grove) == kaiTopologyName {
			return grove, nil
		}
	}

	return nil, nil
}

// setAnnotations sets the grove topology annotations on a KAI Topology.
func (tc *KaiTopologyController) setAnnotations(ctx context.Context, kaiTopology *kaiv1alpha1.Topology, groveTopology *grovev1alpha1.ClusterTopologyBinding) error {
	groveName := groveTopology.GetName()
	resourceVersion := groveTopology.GetResourceVersion()

	currentName := kaiTopology.Annotations[config.Get().GroveTopologyAnnotation]
	currentRV := kaiTopology.Annotations[config.Get().GroveTopologyResourceVersionAnnotation]
	if currentName == groveName && currentRV == resourceVersion {
		return nil
	}

	log.Info().Msgf("Setting grove topology annotations on KAI topology <%s>: name=%s, resourceVersion=%s",
		kaiTopology.Name, groveName, resourceVersion)
	return tc.patchAnnotation(ctx, kaiTopology, map[string]any{
		config.Get().GroveTopologyAnnotation:                groveName,
		config.Get().GroveTopologyResourceVersionAnnotation: resourceVersion,
	})
}

// removeAnnotations removes the grove topology annotations from a KAI Topology
// if they are present.
func (tc *KaiTopologyController) removeAnnotations(ctx context.Context, kaiTopology *kaiv1alpha1.Topology) error {
	_, hasName := kaiTopology.Annotations[config.Get().GroveTopologyAnnotation]
	_, hasRV := kaiTopology.Annotations[config.Get().GroveTopologyResourceVersionAnnotation]
	if !hasName && !hasRV {
		return nil
	}

	log.Info().Msgf("Removing grove topology annotations from KAI topology <%s>", kaiTopology.Name)
	return tc.patchAnnotation(ctx, kaiTopology, map[string]any{
		config.Get().GroveTopologyAnnotation:                nil,
		config.Get().GroveTopologyResourceVersionAnnotation: nil,
	})
}

func (tc *KaiTopologyController) patchAnnotation(ctx context.Context, obj client.Object, annotations map[string]any) error {
	patchBytes, err := json.Marshal(map[string]any{
		"metadata": map[string]any{
			"annotations": annotations,
		},
	})
	if err != nil {
		log.Error().Msgf("Failed to marshal patch for <%s>: %v", obj.GetName(), err)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	if err := tc.client.Patch(ctx, obj, patch); err != nil {
		log.Error().Msgf("Failed to patch annotations on <%s>: %v", obj.GetName(), err)
		return err
	}
	return nil
}
