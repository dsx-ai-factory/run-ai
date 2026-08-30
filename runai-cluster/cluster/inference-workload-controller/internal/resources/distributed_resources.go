package resources

import (
	"context"
	"fmt"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

var logv = logger.Log.WithName("distributed-resources")

// DistributedResources holds all resources related to a distributed inference workload
type DistributedResources struct {
	Lws            *lwsv1.LeaderWorkerSet
	Services       []*corev1.Service
	Ingresses      []*netv1.Ingress
	Routes         []*routev1.Route
	HTTPRoutes     []*gatewayv1.HTTPRoute
	Pvcs           []*corev1.PersistentVolumeClaim
	ComputeDomains []*nvapi.ComputeDomain
}

// resourceHandler defines how to handle a specific resource type
type resourceHandler struct {
	createResource func() client.Object
	addToResult    func(*DistributedResources, client.Object)
}

// DistributedResourcesManagerInterface defines the interface for managing distributed resources
type DistributedResourcesManagerInterface interface {
	GetResources(ctx context.Context, resources []client.Object) (DistributedResources, []client.Object, error)
}

// DistributedResourcesManager handles operations on distributed inference workload resources
type DistributedResourcesManager struct {
	k8sClient client.Client
}

// NewDistributedResourcesManager creates a new DistributedResourcesManager instance
func NewDistributedResourcesManager(k8sClient client.Client) *DistributedResourcesManager {
	return &DistributedResourcesManager{
		k8sClient: k8sClient,
	}
}

// GetResources fetches all distributed resources from the k8s client using ObservedResources
func (drm *DistributedResourcesManager) GetResources(ctx context.Context,
	resources []client.Object) (DistributedResources, []client.Object, error) {
	result := DistributedResources{}
	missingResources := []client.Object{}

	if len(resources) == 0 {
		return result, missingResources, nil
	}

	// Define resource handlers for each supported kind
	resourceHandlers := map[string]resourceHandler{
		constants.KindLeaderWorkerSet: {
			createResource: func() client.Object { return &lwsv1.LeaderWorkerSet{} },
			addToResult: func(dr *DistributedResources, obj client.Object) {
				dr.Lws = obj.(*lwsv1.LeaderWorkerSet)
			},
		},
		constants.KindService: {
			createResource: func() client.Object { return &corev1.Service{} },
			addToResult: func(dr *DistributedResources, obj client.Object) {
				dr.Services = append(dr.Services, obj.(*corev1.Service))
			},
		},
		constants.KindIngress: {
			createResource: func() client.Object { return &netv1.Ingress{} },
			addToResult: func(dr *DistributedResources, obj client.Object) {
				dr.Ingresses = append(dr.Ingresses, obj.(*netv1.Ingress))
			},
		},
		constants.KindRoute: {
			createResource: func() client.Object { return &routev1.Route{} },
			addToResult: func(dr *DistributedResources, obj client.Object) {
				dr.Routes = append(dr.Routes, obj.(*routev1.Route))
			},
		},
		constants.KindHTTPRoute: {
			createResource: func() client.Object { return &gatewayv1.HTTPRoute{} },
			addToResult: func(dr *DistributedResources, obj client.Object) {
				dr.HTTPRoutes = append(dr.HTTPRoutes, obj.(*gatewayv1.HTTPRoute))
			},
		},
		constants.KindPvc: {
			createResource: func() client.Object { return &corev1.PersistentVolumeClaim{} },
			addToResult: func(dr *DistributedResources, obj client.Object) {
				dr.Pvcs = append(dr.Pvcs, obj.(*corev1.PersistentVolumeClaim))
			},
		},
		constants.KindComputeDomain: {
			createResource: func() client.Object { return &nvapi.ComputeDomain{} },
			addToResult: func(dr *DistributedResources, obj client.Object) {
				dr.ComputeDomains = append(dr.ComputeDomains, obj.(*nvapi.ComputeDomain))
			},
		},
	}

	// Iterate through resources and fetch each resource by Kind
	for _, resource := range resources {
		kind := resource.GetObjectKind().GroupVersionKind().Kind
		handler, exists := resourceHandlers[kind]
		if !exists {
			logv.Error(fmt.Errorf("unsupported resource kind"), "skipping unsupported resource kind",
				"kind", kind, "name", resource.GetName())
			continue
		}

		objectKey := client.ObjectKey{
			Namespace: resource.GetNamespace(),
			Name:      resource.GetName(),
		}

		resourceFromClient := handler.createResource()
		err := drm.k8sClient.Get(ctx, objectKey, resourceFromClient)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				logv.Error(err, "failed to get resource", "kind", kind, "name", resourceFromClient.GetName())
				return DistributedResources{}, []client.Object{}, err
			}

			// resource not found, add to missing resources
			missingResources = append(missingResources, resource)
			continue
		}

		handler.addToResult(&result, resourceFromClient)
	}

	return result, missingResources, nil
}
