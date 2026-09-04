/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"fmt"

	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
	reader client.Reader
}

// NewPlugin builds the networking plugin with an explicit client.Reader. Useful
// for tests that want to wire their own reader.
func NewPlugin(reader client.Reader) *Plugin {
	return &Plugin{reader: reader}
}

// New is the framework.PluginBuilder. Pass it to app.RunController in the
// pluginBuilders slice. It loads Config (env vars NETWORKING_*) and returns
// framework.ErrPluginDisabled when the plugin is disabled - the framework
// skips it.
func New(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
	cfg, err := rcpconfig.GetRootR[*Config]()
	if err != nil {
		return nil, fmt.Errorf("load networking config: %w", err)
	}
	if !cfg.Enabled {
		return nil, framework.ErrPluginDisabled
	}
	if deps.Client == nil {
		return nil, fmt.Errorf("networking plugin requires a non-nil Client in PluginDeps")
	}
	return NewPlugin(deps.Client), nil
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	enqueue := handler.EnqueueRequestsFromMapFunc(p.mapToWorkloadRunner)
	bldr = bldr.
		Watches(&dromosv1alpha1.WorkloadRunner{},
			handler.EnqueueRequestsFromMapFunc(mapWorkloadRunnerToSelf),
			builder.WithPredicates(endpointAnnotationChangedPredicate())).
		Watches(&appsv1.Deployment{}, enqueue).
		Watches(&corev1.Service{}, enqueue)

	if workloadconfig.IsOpenshift() {
		bldr = bldr.Watches(&openshiftv1.Route{}, enqueue)
	} else {
		bldr = bldr.Watches(&networkingv1.Ingress{}, enqueue)
	}

	if workloadconfig.GatewayEnabled() {
		bldr = bldr.Watches(&gatewayv1.HTTPRoute{}, enqueue)
	}

	return bldr, nil
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

		owner := &unstructured.Unstructured{}
		owner.SetGroupVersionKind(schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind))
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

var (
	_ framework.ValidatorPlugin          = &Plugin{}
	_ framework.AdditionalResourcePlugin = &Plugin{}
	_ framework.WatcherPlugin            = &Plugin{}
	_ framework.StatusPlugin             = &Plugin{}
)
