/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package framework

import (
	"context"
	"errors"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	ErrPluginDisabled     = errors.New("plugin disabled")
	ErrPluginNameMismatch = errors.New("plugin name mismatch")
)

type PluginBuilder func(ctx context.Context, deps PluginDeps) (Plugin, error)

type PluginStatusSetter interface {
	SetCondition(condition metav1.Condition) bool
	SetPhase(phase dromosv1alpha1.PluginPhase)
	SetAdditionalInfo(additionalInfo *runtime.RawExtension)
}

type Plugin interface {
	Name() string
}

// WatcherPlugin is a plugin interface that let plugins add watchers to the controller.
type WatcherPlugin interface {
	Plugin
	Watchers(builder *builder.Builder) (*builder.Builder, error)
}

// ValidatorPlugin is a plugin interface that lets plugins validate WorkloadRunners through Webhook.
type ValidatorPlugin interface {
	Plugin
	WebhookValidate(ctx context.Context, existingRunner, desiredRunner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, field.ErrorList)
}

// MutatorPlugin is a plugin interface that lets plugins mutate WorkloadRunners through Webhook.
type MutatorPlugin interface {
	Plugin
	WebhookMutate(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, error)
}

// ResourceMutatePlugin is a plugin interface that lets plugins mutate the attached resource during reconciliation.
type ResourceMutatePlugin interface {
	Plugin
	MutateResource(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, resource *unstructured.Unstructured, karta *kartav1alpha1.Karta, state *reconcile.State) error
}

// AdditionalResourcePlugin is a plugin interface that lets plugins create additional resources during reconciliation.
type AdditionalResourcePlugin interface {
	Plugin
	HandleAdditionalResources(
		ctx context.Context,
		runner *dromosv1alpha1.WorkloadRunner,
		resource *unstructured.Unstructured,
		karta *kartav1alpha1.Karta,
		k8sClient client.Client,
		state *reconcile.State,
	) error
}

// StatusPlugin is a plugin interface that lets plugins handle resources status by adding plugins dedicated conditions and phase.
type StatusPlugin interface {
	Plugin
	Status(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, k8sClient client.Client, status PluginStatusSetter, state *reconcile.State) error
}

// PostPhasePlugin is a plugin interface that lets plugins react after the WorkloadRunner
// status and phase have been computed and persisted. Plugins inspect the new state
// (runner.Status.ResourcePhase) and act on it, e.g. delete resources they created once
// the workload reaches a terminal phase.
type PostPhasePlugin interface {
	Plugin
	HandlePostPhase(
		ctx context.Context,
		runner *dromosv1alpha1.WorkloadRunner,
		resource *unstructured.Unstructured,
		karta *kartav1alpha1.Karta,
		k8sClient client.Client,
	) error
}
