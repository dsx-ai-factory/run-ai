/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package app

import (
	"context"
	"errors"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/webhook/mutator"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/webhook/validator"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const controllerName = "anyworkload-controller"

// SetupFunc receives the controller manager after dromos's built-in
// controller and webhooks are set up, before the manager starts. Consumers
// register additional reconcilers and webhook paths here.
type SetupFunc func(ctx context.Context, mgr ctrl.Manager) error

// opts is the internal options bag populated by Option values.
type opts struct {
	plugins        []framework.PluginBuilder
	schemeBuilders []func(*runtime.Scheme) error
	scheme         *runtime.Scheme
	config         *config.DromosConfig
	restConfig     *rest.Config
	managerSetup   []SetupFunc
}

// Option configures RunController.
type Option func(*opts)

// WithPlugins registers one or more plugin builders.
func WithPlugins(builders ...framework.PluginBuilder) Option {
	return func(o *opts) { o.plugins = append(o.plugins, builders...) }
}

// WithSchemeBuilders registers additional AddToScheme functions on top of
// dromos's defaults and any base scheme provided via WithScheme.
func WithSchemeBuilders(sb ...func(*runtime.Scheme) error) Option {
	return func(o *opts) { o.schemeBuilders = append(o.schemeBuilders, sb...) }
}

// WithConfig overrides dromos's default controller-wide config. If unset,
// config.Default() is used.
func WithConfig(c *config.DromosConfig) Option {
	return func(o *opts) { o.config = c }
}

// WithScheme sets a base scheme. Dromos's defaults are layered on top via
// AddToScheme (idempotent).
func WithScheme(s *runtime.Scheme) Option {
	return func(o *opts) { o.scheme = s }
}

// WithManagerSetup registers setup functions run in registration order right
// before the manager starts; the first error aborts startup.
func WithManagerSetup(fns ...SetupFunc) Option {
	return func(o *opts) { o.managerSetup = append(o.managerSetup, fns...) }
}

// WithRestConfig overrides the manager's rest.Config (default
// ctrl.GetConfigOrDie()).
func WithRestConfig(cfg *rest.Config) Option {
	return func(o *opts) { o.restConfig = cfg }
}

// RunController boots the dromos controller. It blocks until the manager
// exits or returns an error during startup.
//
// Plugin enablement is start-time only: a builder may return
// framework.ErrPluginDisabled to opt out. Toggling a plugin's configuration
// requires restarting the controller pod.
//
// The context controls the lifetime of the controller manager. Callers
// typically pass ctrl.SetupSignalHandler() so the manager shuts down cleanly
// on SIGINT/SIGTERM.
func RunController(ctx context.Context, options ...Option) error {
	o := &opts{config: config.Default()}
	for _, fn := range options {
		fn(o)
	}

	initLogging(o.config)
	log.Info().Msg("Starting dromos controller")

	scheme, err := initScheme(o.scheme, o.schemeBuilders)
	if err != nil {
		return fmt.Errorf("init scheme: %w", err)
	}

	restConfig := o.restConfig
	if restConfig == nil {
		restConfig = ctrl.GetConfigOrDie()
	}

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme: scheme,
		WebhookServer: &webhook.DefaultServer{
			Options: webhook.Options{
				Port:    o.config.WebhookPort,
				CertDir: o.config.CertDir,
			},
		},
		Metrics: metricsserver.Options{
			BindAddress: fmt.Sprintf(":%d", o.config.MetricsPort),
		},
	})
	if err != nil {
		return fmt.Errorf("create manager: %w", err)
	}

	deps := framework.PluginDeps{
		Client: mgr.GetClient(),
		Logger: ctrl.Log.WithName("plugin"),
	}
	fw, err := framework.NewFrameworkFromPluginBuilders(ctx, &framework.Config{
		MaxMutationRounds: o.config.MaxMutationRounds,
	}, o.plugins, deps)
	if err != nil {
		return fmt.Errorf("build framework: %w", err)
	}

	recorder := mgr.GetEventRecorderFor(controllerName)
	workloadRunnerController := controller.NewWorkloadRunnerController(mgr.GetClient(), fw, recorder)
	if err = workloadRunnerController.SetupWithManager(ctx, mgr); err != nil {
		return fmt.Errorf("setup controller: %w", err)
	}

	if err = setupWebhooks(mgr, fw); err != nil {
		return fmt.Errorf("setup webhooks: %w", err)
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("setup health check: %w", err)
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("setup ready check: %w", err)
	}

	for i, setup := range o.managerSetup {
		if err = setup(ctx, mgr); err != nil {
			return fmt.Errorf("run manager setup %d: %w", i, err)
		}
	}

	log.Info().Msg("Starting WorkloadRunnerController")
	if err = mgr.Start(ctx); err != nil {
		return fmt.Errorf("manager exited with error: %w", err)
	}
	return nil
}

func initLogging(c *config.DromosConfig) {
	if !c.UseDebugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: c.UseDebugLogLevel})))
}

// defaultSchemeBuilders is dromos's minimum scheme set: core K8s, CRD
// discovery, the WorkloadRunner CRD, Karta, and NVIDIA DRA ComputeDomain
// (used by the in-tree mnnvl plugin).
var defaultSchemeBuilders = []func(*runtime.Scheme) error{
	clientgoscheme.AddToScheme,
	apiextensionsv1.AddToScheme,
	dromosv1alpha1.AddToScheme,
	kartav1alpha1.AddToScheme,
	nvapi.AddToScheme,
}

func initScheme(base *runtime.Scheme, extra []func(*runtime.Scheme) error) (*runtime.Scheme, error) {
	scheme := base
	if scheme == nil {
		scheme = runtime.NewScheme()
	}
	var errs []error
	for _, add := range defaultSchemeBuilders {
		if err := add(scheme); err != nil {
			errs = append(errs, err)
		}
	}
	for _, add := range extra {
		if err := add(scheme); err != nil {
			errs = append(errs, err)
		}
	}
	if joined := errors.Join(errs...); joined != nil {
		return nil, joined
	}
	return scheme, nil
}

func setupWebhooks(mgr ctrl.Manager, fw *framework.Framework) error {
	webhookServer := mgr.GetWebhookServer()

	validatorHandler := validator.NewWorkloadRunnerValidatorHandler(fw)
	webhookServer.Register("/validate-workloadrunner", &webhook.Admission{Handler: validatorHandler})

	mutatorHandler := mutator.NewWorkloadRunnerMutatorHandler(fw)
	webhookServer.Register("/mutate-workloadrunner", &webhook.Admission{Handler: mutatorHandler})

	log.Info().Msg("Webhooks registered successfully")
	return nil
}
