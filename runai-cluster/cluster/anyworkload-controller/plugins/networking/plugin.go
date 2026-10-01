/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"fmt"
	"reflect"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	openshiftv1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	PluginName = "networking-plugin"
)

type Plugin struct {
	reader           client.Reader
	ownershipReader  client.Reader
	templatesReader  client.Reader
	enforceOwnership bool
	templatesEnabled bool
	routingDisabled  bool
}

// Integration owns the networking plugin's anyworkload-controller-specific
// runtime setup. Dromos builds plugins before it runs manager setup callbacks,
// and starts the manager only after both stages complete.
type Integration struct {
	plugin   *Plugin
	disabled bool
}

// NewIntegration creates the networking plugin's production lifecycle adapter.
func NewIntegration() *Integration {
	return &Integration{}
}

// NewSetup returns the manager setup for this networking integration.
func (i *Integration) NewSetup() func(context.Context, ctrl.Manager) error {
	return func(ctx context.Context, manager ctrl.Manager) error {
		return i.Setup(ctx, manager.GetAPIReader(), manager.GetFieldIndexer())
	}
}

// NewPlugin builds the networking plugin with an explicit client.Reader. Useful
// for tests that want to wire their own reader.
func NewPlugin(reader client.Reader) *Plugin {
	return newPlugin(reader, reader)
}

func newPlugin(reader, ownershipReader client.Reader) *Plugin {
	return &Plugin{
		reader:           reader,
		ownershipReader:  ownershipReader,
		templatesReader:  ownershipReader,
		enforceOwnership: true,
		templatesEnabled: workloadconfig.OAuth2ProxyTemplatesEnabled(),
	}
}

// Build is the framework.PluginBuilder for this integration. Setup must run
// successfully before the controller manager starts.
func (i *Integration) Build(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
	cfg, err := rcpconfig.GetRootR[*Config]()
	if err != nil {
		return nil, fmt.Errorf("load networking config: %w", err)
	}
	if !cfg.Enabled && !workloadconfig.OAuth2ProxyTemplatesEnabled() {
		i.disabled = true
		return nil, framework.ErrPluginDisabled
	}
	if deps.Client == nil {
		return nil, fmt.Errorf("networking plugin requires a non-nil Client in PluginDeps")
	}
	i.plugin = newPlugin(deps.Client, nil)
	i.plugin.routingDisabled = !cfg.Enabled
	return i.plugin, nil
}

// Setup injects the live reader into networking and registers its cache
// indexes. It must be called from app.WithManagerSetup before mgr.Start.
func (i *Integration) Setup(
	ctx context.Context,
	apiReader client.Reader,
	fieldIndexer client.FieldIndexer,
) error {
	if i.disabled {
		return nil
	}
	if i.plugin == nil {
		return fmt.Errorf("networking integration must be built before setup")
	}
	if apiReader == nil {
		return fmt.Errorf("networking integration requires a non-nil APIReader")
	}
	i.plugin.templatesReader = apiReader
	if i.plugin.routingDisabled {
		return nil
	}
	if fieldIndexer == nil {
		return fmt.Errorf("networking integration requires a non-nil FieldIndexer")
	}
	if err := registerDependencyIndexes(ctx, fieldIndexer); err != nil {
		return err
	}
	i.plugin.ownershipReader = apiReader
	return nil
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	if p.templatesEnabled {
		bldr = bldr.Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(p.mapConfigMapToRunners), builder.OnlyMetadata)
	}
	if p.routingDisabled {
		return bldr, nil
	}
	ownedEnqueue := handler.EnqueueRequestsFromMapFunc(p.mapToWorkloadRunner)
	serviceEnqueue := handler.EnqueueRequestsFromMapFunc(p.mapServiceDependency)
	podEnqueue := handler.EnqueueRequestsFromMapFunc(p.mapPodDependency)
	endpointSliceEnqueue := handler.EnqueueRequestsFromMapFunc(p.mapEndpointSliceDependency)
	exposureEnqueue := handler.EnqueueRequestsFromMapFunc(p.mapExposureDependency)
	bldr = bldr.
		Watches(&dromosv1alpha1.WorkloadRunner{},
			handler.EnqueueRequestsFromMapFunc(mapWorkloadRunnerToSelf),
			builder.WithPredicates(endpointAnnotationChangedPredicate())).
		Watches(&appsv1.Deployment{}, ownedEnqueue).
		Watches(&appsv1.ReplicaSet{}, ownedEnqueue,
			builder.OnlyMetadata, builder.WithPredicates(ownerReferencesChangedPredicate())).
		Watches(&appsv1.StatefulSet{}, ownedEnqueue,
			builder.OnlyMetadata, builder.WithPredicates(ownerReferencesChangedPredicate())).
		Watches(&batchv1.Job{}, ownedEnqueue,
			builder.OnlyMetadata, builder.WithPredicates(ownerReferencesChangedPredicate())).
		Watches(&corev1.Service{}, serviceEnqueue).
		Watches(&corev1.Pod{}, podEnqueue, builder.WithPredicates(podDependencyChangedPredicate())).
		Watches(&discoveryv1.EndpointSlice{}, endpointSliceEnqueue).
		Watches(&networkingv1.Ingress{}, exposureEnqueue)

	if workloadconfig.IsOpenshift() {
		bldr = bldr.Watches(&openshiftv1.Route{}, exposureEnqueue)
	}

	if workloadconfig.GatewayEnabled() {
		bldr = bldr.Watches(&gatewayv1.HTTPRoute{}, exposureEnqueue)
	}

	return bldr, nil
}

func ownerReferencesChangedPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectOld != nil && e.ObjectNew != nil &&
				!reflect.DeepEqual(e.ObjectOld.GetOwnerReferences(), e.ObjectNew.GetOwnerReferences())
		},
		GenericFunc: func(_ event.GenericEvent) bool {
			return false
		},
	}
}

func podDependencyChangedPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldPod, oldOK := e.ObjectOld.(*corev1.Pod)
			newPod, newOK := e.ObjectNew.(*corev1.Pod)
			if !oldOK || !newOK {
				return false
			}
			return !reflect.DeepEqual(oldPod.Labels, newPod.Labels) ||
				!reflect.DeepEqual(oldPod.OwnerReferences, newPod.OwnerReferences) ||
				oldPod.Status.PodIP != newPod.Status.PodIP ||
				!reflect.DeepEqual(oldPod.Status.PodIPs, newPod.Status.PodIPs)
		},
		GenericFunc: func(_ event.GenericEvent) bool {
			return false
		},
	}
}

func endpointAnnotationChangedPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(_ event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(_ event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == nil || e.ObjectNew == nil {
				return false
			}
			return annotationValueChanged(e.ObjectOld, e.ObjectNew, endpoints.Annotation)
		},
		GenericFunc: func(_ event.GenericEvent) bool {
			return false
		},
	}
}

func annotationValueChanged(oldObj, newObj metav1.Object, annotation string) bool {
	oldValue, oldExists := oldObj.GetAnnotations()[annotation]
	newValue, newExists := newObj.GetAnnotations()[annotation]
	return oldExists != newExists || oldValue != newValue
}

func mapWorkloadRunnerToSelf(_ context.Context, obj client.Object) []reconcile.Request {
	if obj == nil {
		return nil
	}

	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()},
	}}
}

func (p *Plugin) mapToWorkloadRunner(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	for _, ref := range obj.GetOwnerReferences() {
		if ref.Kind == "WorkloadRunner" {
			logger.V(2).Info("mapToWorkloadRunner: direct WorkloadRunner owner",
				"resource", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()),
				"workloadRunner", ref.Name)
			return []reconcile.Request{{
				NamespacedName: types.NamespacedName{Name: ref.Name, Namespace: obj.GetNamespace()},
			}}
		}

		owner := dependencyOwnerObject(ref)
		if err := p.reader.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: obj.GetNamespace()}, owner); err != nil {
			logger.V(2).Info("mapToWorkloadRunner: could not fetch intermediate owner",
				"resource", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()),
				"ownerKind", ref.Kind, "ownerName", ref.Name, "error", err)
			continue
		}

		for _, ownerRef := range owner.GetOwnerReferences() {
			if ownerRef.Kind == "WorkloadRunner" {
				logger.V(2).Info("mapToWorkloadRunner: two-hop match",
					"resource", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()),
					"intermediateOwner", fmt.Sprintf("%s/%s", ref.Kind, ref.Name),
					"workloadRunner", ownerRef.Name)
				return []reconcile.Request{{
					NamespacedName: types.NamespacedName{Name: ownerRef.Name, Namespace: obj.GetNamespace()},
				}}
			}
		}
	}
	return nil
}

// dependencyOwnerObject matches the cache projection used by the dependency
// watches. This avoids starting a second, full-object informer merely to read
// owner references from metadata-only watched resources.
func dependencyOwnerObject(ref metav1.OwnerReference) client.Object {
	gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind)
	var owner client.Object
	switch gvk {
	case appsv1.SchemeGroupVersion.WithKind("Deployment"):
		// Deployment status is read elsewhere, so its existing informer must keep
		// the structured object rather than creating a second metadata cache.
		owner = &appsv1.Deployment{}
	case appsv1.SchemeGroupVersion.WithKind("ReplicaSet"),
		appsv1.SchemeGroupVersion.WithKind("StatefulSet"),
		batchv1.SchemeGroupVersion.WithKind("Job"):
		owner = &metav1.PartialObjectMetadata{}
	default:
		owner = &unstructured.Unstructured{}
	}
	owner.GetObjectKind().SetGroupVersionKind(gvk)
	return owner
}

var (
	_ framework.ValidatorPlugin          = &Plugin{}
	_ framework.AdditionalResourcePlugin = &Plugin{}
	_ framework.WatcherPlugin            = &Plugin{}
	_ framework.StatusPlugin             = &Plugin{}
)
