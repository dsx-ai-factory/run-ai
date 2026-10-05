/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package inferenceendpoint ties a workload's resolved inference routing intent
// to an owned InferenceEndpoint. Its children belong to the endpoint controller.
package inferenceendpoint

import (
	"context"
	"fmt"
	ep "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromos "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const PluginName = ep.InferenceEndpointPluginName
const ReadyCondition = ep.InferenceEndpointReadyCondition

type Plugin struct {
	reader client.Reader
	mapper meta.RESTMapper
}

func NewPlugin(reader client.Reader, mapper meta.RESTMapper) *Plugin {
	return &Plugin{reader: reader, mapper: mapper}
}

func New(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
	c, ok := deps.Client.(client.Client)
	if !ok || c == nil {
		return nil, fmt.Errorf("inference endpoint plugin requires a client with a REST mapper")
	}
	return NewPlugin(c, c.RESTMapper()), nil
}
func (p *Plugin) Name() string { return PluginName }

func (p *Plugin) available() error {
	if p.mapper == nil {
		return fmt.Errorf("InferenceEndpoint REST mapper is unavailable")
	}
	_, err := p.mapper.RESTMapping(ep.InferenceEndpointGVK.GroupKind(), ep.InferenceEndpointGVK.Version)
	if err != nil {
		return fmt.Errorf("InferenceEndpoint API is unavailable: %w", err)
	}
	return nil
}

func (p *Plugin) Watchers(b *builder.Builder) (*builder.Builder, error) {
	if err := p.available(); err != nil {
		if meta.IsNoMatchError(err) {
			return b, nil
		}
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(ep.InferenceEndpointGVK)
	// The root controller watches generation changes, which exclude annotations.
	// Watch this contract separately so add, update and removal reconcile even
	// when neither the template nor an existing endpoint emits an event.
	return b.Owns(obj).Watches(&dromos.WorkloadRunner{}, &handler.EnqueueRequestForObject{},
		builder.WithPredicates(predicate.Funcs{
			CreateFunc:  func(event.CreateEvent) bool { return false },
			DeleteFunc:  func(event.DeleteEvent) bool { return false },
			GenericFunc: func(event.GenericEvent) bool { return false },
			UpdateFunc: func(e event.UpdateEvent) bool {
				if e.ObjectOld == nil || e.ObjectNew == nil {
					return false
				}
				old, oldSet := e.ObjectOld.GetAnnotations()[ep.InferenceEndpointAnnotation]
				next, nextSet := e.ObjectNew.GetAnnotations()[ep.InferenceEndpointAnnotation]
				return oldSet != nextSet || old != next
			},
		})), nil
}

var (
	_ framework.AdditionalResourcePlugin = &Plugin{}
	_ framework.StatusPlugin             = &Plugin{}
	_ framework.ValidatorPlugin          = &Plugin{}
	_ framework.WatcherPlugin            = &Plugin{}
)
