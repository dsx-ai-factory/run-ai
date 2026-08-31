/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package framework

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const defaultMaxMutationRounds = 5

type Config struct {
	MaxMutationRounds int
}

type Framework struct {
	registry                   Registry
	watcherPlugins             []WatcherPlugin
	validatorPlugins           []ValidatorPlugin
	mutatorPlugins             []MutatorPlugin
	resourceMutatePlugins      []ResourceMutatePlugin
	additionalResourcesPlugins []AdditionalResourcePlugin
	statusPlugins              []StatusPlugin
	postPhasePlugins           []PostPhasePlugin
	config                     Config
}

func NewFramework(registry Registry, config *Config) *Framework {
	if config == nil {
		config = &Config{
			MaxMutationRounds: defaultMaxMutationRounds,
		}
	}

	f := &Framework{
		registry:                   registry,
		watcherPlugins:             []WatcherPlugin{},
		validatorPlugins:           []ValidatorPlugin{},
		mutatorPlugins:             []MutatorPlugin{},
		resourceMutatePlugins:      []ResourceMutatePlugin{},
		additionalResourcesPlugins: []AdditionalResourcePlugin{},
		statusPlugins:              []StatusPlugin{},
		postPhasePlugins:           []PostPhasePlugin{},
		config:                     *config,
	}

	// Plugins run in name order, so plugin output is stable across processes.
	for _, name := range slices.Sorted(maps.Keys(registry)) {
		plugin := registry[name]
		if watcherPlugin, ok := plugin.(WatcherPlugin); ok {
			f.watcherPlugins = append(f.watcherPlugins, watcherPlugin)
		}

		if validatorPlugin, ok := plugin.(ValidatorPlugin); ok {
			f.validatorPlugins = append(f.validatorPlugins, validatorPlugin)
		}

		if mutatorPlugin, ok := plugin.(MutatorPlugin); ok {
			f.mutatorPlugins = append(f.mutatorPlugins, mutatorPlugin)
		}

		if resourceMutatePlugin, ok := plugin.(ResourceMutatePlugin); ok {
			f.resourceMutatePlugins = append(f.resourceMutatePlugins, resourceMutatePlugin)
		}

		if additionalResourcesPlugin, ok := plugin.(AdditionalResourcePlugin); ok {
			f.additionalResourcesPlugins = append(f.additionalResourcesPlugins, additionalResourcesPlugin)
		}

		if resourcesStatusPlugin, ok := plugin.(StatusPlugin); ok {
			f.statusPlugins = append(f.statusPlugins, resourcesStatusPlugin)
		}

		if postPhasePlugin, ok := plugin.(PostPhasePlugin); ok {
			f.postPhasePlugins = append(f.postPhasePlugins, postPhasePlugin)
		}
	}

	return f
}

// NewFrameworkFromPluginBuilders invokes each PluginBuilder and assembles the
// resulting plugins into a Framework. Plugins that signal opt-out via
// ErrPluginDisabled are silently skipped. Any other builder error aborts.
//
// The framework checks that each constructed plugin's Name() is unique across
// the registry; collisions abort with ErrPluginNameMismatch.
func NewFrameworkFromPluginBuilders(
	ctx context.Context,
	conf *Config,
	pluginBuilders []PluginBuilder,
	deps PluginDeps,
) (*Framework, error) {
	registry := Registry{}
	for _, build := range pluginBuilders {
		p, err := build(ctx, deps)
		if errors.Is(err, ErrPluginDisabled) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("build plugin: %w", err)
		}
		if _, exists := registry[p.Name()]; exists {
			return nil, fmt.Errorf("plugin %q already registered: %w", p.Name(), ErrPluginNameMismatch)
		}
		registry[p.Name()] = p
	}
	return NewFramework(registry, conf), nil
}

// RunWatchers runs all the watcher plugins on the controller builder.
func (f *Framework) RunWatchers(bldr *builder.Builder) (*builder.Builder, error) {
	var err error

	for _, plugin := range f.watcherPlugins {
		bldr, err = plugin.Watchers(bldr)
		if err != nil {
			return nil, err
		}
	}

	return bldr, nil
}

// RunValidators runs all the validator plugins on the WorkloadRunner.
func (f *Framework) RunValidators(
	ctx context.Context,
	existingRunner *dromosv1alpha1.WorkloadRunner,
	desiredRunner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, field.ErrorList) {
	logger := log.FromContext(ctx)

	var (
		allWarnings admission.Warnings
		allErrors   field.ErrorList
	)

	for _, plugin := range f.validatorPlugins {
		warnings, fieldErrs := plugin.WebhookValidate(ctx, existingRunner, desiredRunner)

		if len(warnings) > 0 {
			allWarnings = append(allWarnings, warnings...)
			logger.Info("Validation warnings", "plugin", plugin.Name(), "warnings", warnings)
		}

		if len(fieldErrs) > 0 {
			allErrors = append(allErrors, fieldErrs...)
			logger.Error(fmt.Errorf("validation failed"), "Validation failed", "plugin", plugin.Name(), "errors", fieldErrs)
		}
	}

	return allWarnings, allErrors
}

// RunMutators runs all the mutator plugins on the WorkloadRunner.
func (f *Framework) RunMutators(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, error) {
	logger := log.FromContext(ctx)

	var allWarnings admission.Warnings

	for _, plugin := range f.mutatorPlugins {
		warnings, err := plugin.WebhookMutate(ctx, runner)

		if len(warnings) > 0 {
			allWarnings = append(allWarnings, warnings...)
			logger.Info("Mutation warnings", "plugin", plugin.Name(), "warnings", warnings)
		}

		if err != nil {
			logger.Error(err, "Mutation failed", "plugin", plugin.Name())
			return nil, fmt.Errorf("mutation failed in plugin %s: %w", plugin.Name(), err)
		}
	}

	return allWarnings, nil
}

// RunMutateResource runs all the resource mutate plugins on the AttachedResource with reinvocation logic.
func (f *Framework) RunMutateResource(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) error {
	logger := log.FromContext(ctx)

	// In order to accomplish overwrite between plugins, we run the plugins in a loop
	// until no more mutations are needed or we reach the max rounds.
	for round := 1; round <= f.config.MaxMutationRounds; round++ {
		anyMutated := false

		for _, plugin := range f.resourceMutatePlugins {
			resourceCopy := resource.DeepCopy()
			if err := plugin.MutateResource(ctx, runner, resource, karta, state); err != nil {
				return fmt.Errorf("MutateResource failed in plugin %s: %w", plugin.Name(), err)
			}
			if !reflect.DeepEqual(resourceCopy.Object, resource.Object) {
				anyMutated = true

				logger.V(3).Info( //nolint:mnd // log verbosity level
					"Plugin mutated resource",
					"plugin", plugin.Name(),
					"round", round,
				)
			}
		}

		if !anyMutated {
			logger.V(3).Info("MutateResource completed", "rounds", round) //nolint:mnd // log verbosity level

			return nil
		}
	}

	err := fmt.Errorf("MutateResource exceeded max rounds (%d) - mutations did not converge", f.config.MaxMutationRounds)
	state.AddControllerIssue(resource, err, true, "", issue.Internal)
	return err
}

// RunHandleAdditionalResources runs all the additional resources plugins to handle additional resources.
func (f *Framework) RunHandleAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	logger := log.FromContext(ctx)

	for _, plugin := range f.additionalResourcesPlugins {
		if err := plugin.HandleAdditionalResources(ctx, runner, resource, karta, k8sClient, state); err != nil {
			return fmt.Errorf("HandleAdditionalResources failed in plugin %s: %w", plugin.Name(), err)
		}

		logger.Info("Plugin handled additional resources", "plugin", plugin.Name())
	}

	return nil
}

// RunResourcesStatus runs all the resources status plugins to handle resources status.
func (f *Framework) RunResourcesStatus(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, k8sClient client.Client, state *reconcile.State) error {
	logger := log.FromContext(ctx)
	var pluginsError error
	if runner.Status.Plugins == nil {
		runner.Status.Plugins = make(map[string]*dromosv1alpha1.PluginStatus)
	}

	for _, plugin := range f.statusPlugins {
		if runner.Status.Plugins[plugin.Name()] == nil {
			runner.Status.Plugins[plugin.Name()] = &dromosv1alpha1.PluginStatus{Conditions: make([]v1.Condition, 0), Phase: dromosv1alpha1.PluginPhaseUnknown}
		}
		if err := plugin.Status(ctx, runner, k8sClient, runner.Status.Plugins[plugin.Name()], state); err != nil {
			err = state.AddPluginIssue(ctx, plugin, runner, fmt.Errorf("status failed in plugin %s: %w", plugin.Name(), err), false, "", issue.Internal)
			pluginsError = errors.Join(pluginsError, err)
		}

		logger.Info("Plugin handled resources status", "plugin", plugin.Name())
	}

	return pluginsError
}

// RunPostPhase runs all the post-phase plugins after the WorkloadRunner status and phase
// have been computed and persisted.
func (f *Framework) RunPostPhase(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	k8sClient client.Client,
) error {
	for _, plugin := range f.postPhasePlugins {
		if err := plugin.HandlePostPhase(ctx, runner, resource, karta, k8sClient); err != nil {
			return fmt.Errorf("HandlePostPhase failed in plugin %s: %w", plugin.Name(), err)
		}
	}
	return nil
}
