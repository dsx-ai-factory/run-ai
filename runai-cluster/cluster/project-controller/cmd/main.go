/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"os"
	"strings"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/cmd/profiling"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/webhooks/validation"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"go.uber.org/zap/zapcore"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/cmd/version"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/reconcilers"

	// TODO: remove once switch to new repo
	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(kaiv2.AddToScheme(scheme))
	utilruntime.Must(kaiv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	options, projectReconcilerConfig, _ := config.SetOptions()
	logLevel := zapcore.InfoLevel
	if options.Debug {
		logLevel = zapcore.DebugLevel
	}
	ctrl.SetLogger(zap.New(zap.Level(logLevel), zap.UseDevMode(true)))

	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(projectReconcilerConfig.K8sClientConfigQPS)
	clientConfig.Burst = projectReconcilerConfig.K8sClientConfigBurst
	configMapCacheNamespaces := map[string]cache.Config{}
	for _, namespace := range []string{projectReconcilerConfig.InstallNamespace, projectReconcilerConfig.RoleBindingsCmNamespace} {
		if namespace != "" {
			configMapCacheNamespaces[namespace] = cache.Config{}
		}
	}

	mgr, err := ctrl.NewManager(clientConfig, ctrl.Options{
		Scheme: scheme,
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&corev1.ConfigMap{}: {
					Namespaces: configMapCacheNamespaces,
				},
			},
		},
		Metrics: metricsserver.Options{
			BindAddress: fmt.Sprintf(":%s", options.MetricsPort),
		},
		Client: client.Options{
			Cache: &client.CacheOptions{Unstructured: true},
		},
		LeaderElection:   options.EnableLeaderElection,
		LeaderElectionID: "project-controller.run.ai",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// projectEvents is a channel held locally by the controller to pass reconcile events between the reconcilers, instead of going out via the api.
	projectEvents := make(chan event.GenericEvent)
	defer close(projectEvents)

	createMainReconciler(mgr, projectEvents, projectReconcilerConfig)
	createReconcileEventTriggers(mgr, projectEvents)
	createDepartmentReconciler(mgr)
	registerValidationWebhooks(mgr, projectReconcilerConfig)

	// +kubebuilder:scaffold:builder

	if projectReconcilerConfig.EnableProfiler {
		go profiling.RegisterProfiler(projectReconcilerConfig.ProfilerApiPort)
	}

	printVersion()
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Error running manager")
		os.Exit(1)
	}
}

// The 'main' reconciler is the Reconciler that listens to events on Project and owned resources and reconciles them based on data in the Project resource
func createMainReconciler(mgr manager.Manager, projectEvents chan event.GenericEvent, config *config.ProjectReconcilerConfig) {
	if err := (reconcilers.NewProjectReconciler(mgr.GetClient(), mgr.GetAPIReader(), mgr.GetScheme(), projectEvents, config)).SetupWithManager(mgr, config); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", common.LogProjectTag)
		os.Exit(1)
	}
}

// These reconcilers are more like listeners that watch certain events and trigger reconciliation events for all projects
// They are separated mainly for ease of development and code navigation but also because setting several watches and event filters within one 'fat' reconciler is a bit confusing
func createReconcileEventTriggers(mgr manager.Manager, projectEvents chan event.GenericEvent) {
	err := (reconcilers.NewLimitRangeReconciler(mgr.GetClient(), mgr.GetScheme(), projectEvents)).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", common.LogLimitRangeTag)
		os.Exit(1)
	}
}

func createDepartmentReconciler(mgr manager.Manager) {
	err := reconcilers.NewDepartmentReconciler(mgr.GetClient()).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", common.LogDepartmentTag)
		os.Exit(1)
	}
}

// registerValidationWebhooks registers the validating admission handler on a
// webhook server for the resources that are enabled via configuration. The
// handler is only the serving side; the webhook configuration that routes
// admission requests to it is provisioned separately by the deployment's helm
// chart. When no resource is enabled, no webhook server is added so deployments
// without serving certificates are unaffected.
func registerValidationWebhooks(mgr manager.Manager, cfg *config.ProjectReconcilerConfig) {
	if !cfg.EnableProjectValidationWebhook && !cfg.EnableDepartmentValidationWebhook {
		setupLog.Info("validation webhooks are disabled, skipping webhook server registration")
		return
	}

	if cfg.WebhookPort <= 0 || cfg.WebhookPort > 65535 {
		setupLog.Error(fmt.Errorf("invalid webhook port %d", cfg.WebhookPort),
			"a validation webhook is enabled but the webhook port is not configured correctly")
		os.Exit(1)
	}
	if cfg.WebhookCertDir == "" {
		setupLog.Error(fmt.Errorf("webhook cert directory is empty"),
			"a validation webhook is enabled but the webhook cert directory is not configured")
		os.Exit(1)
	}

	validator := validation.NewValidator(mgr.GetClient())
	webhookServer := webhook.NewServer(webhook.Options{
		Port:    cfg.WebhookPort,
		CertDir: cfg.WebhookCertDir,
	})

	if cfg.EnableProjectValidationWebhook {
		webhookServer.Register(validation.ProjectWebhookPath, &webhook.Admission{Handler: validator})
	}
	if cfg.EnableDepartmentValidationWebhook {
		webhookServer.Register(validation.DepartmentWebhookPath, &webhook.Admission{Handler: validator})
	}

	if err := mgr.Add(webhookServer); err != nil {
		setupLog.Error(err, "unable to add validation webhook server to manager")
		os.Exit(1)
	}

	setupLog.Info("successfully registered validation webhook server",
		"port", cfg.WebhookPort,
		"certDir", cfg.WebhookCertDir,
		"projectWebhookEnabled", cfg.EnableProjectValidationWebhook,
		"departmentWebhookEnabled", cfg.EnableDepartmentValidationWebhook)
}

func printVersion() {
	setupLog.Info(`
	
            _______  ______  _                                                                                       
           / ___/ / / / __ \(_)                                                                                      
          / /  / /_/ / / / /                                                                                         
         /_/   \__,_/_/ /__(_)                                                                                        
                     ____ _(_)                                                                                        
                    / __  / /
                   / /_/ / /
                   \__,_/_/
    ____  ____  ____      ____________________   __________  _   ____________  ____  __    __    __________
   / __ \/ __ \/ __ \    / / ____/ ____/_  __/  / ____/ __ \/ | / /_  __/ __ \/ __ \/ /   / /   / ____/ __ \
  / /_/ / /_/ / / / /_  / / __/ / /     / /    / /   / / / /  |/ / / / / /_/ / / / / /   / /   / __/ / /_/ /
 / ____/ _, _/ /_/ / /_/ / /___/ /___  / /    / /___/ /_/ / /|  / / / / _, _/ /_/ / /___/ /___/ /___/ _, _/
/_/   /_/ |_|\____/\____/_____/\____/ /_/     \____/\____/_/ |_/ /_/ /_/ |_|\____/_____/_____/_____/_/ |_|`)

	versionInfo := version.GetVersion()
	setupLog.Info(fmt.Sprintf("Version: %s", versionInfo.Version))
	if strings.HasSuffix(versionInfo.Version, "-DEVELOPMENT") {
		return
	}
	if versionInfo.BuildDate != "" {
		setupLog.Info(fmt.Sprintf("BuildDate: %s", versionInfo.BuildDate))
	}
	if versionInfo.GitCommit != "" {
		setupLog.Info(fmt.Sprintf("Commit Hash: %s", versionInfo.GitCommit))
	}
	if versionInfo.GitTag != "" {
		setupLog.Info(fmt.Sprintf("Tag: %s", versionInfo.GitTag))
	}
	setupLog.Info(fmt.Sprintf("Go Version: %s", versionInfo.GoVersion))
	setupLog.Info(fmt.Sprintf("Platform: %s", versionInfo.Platform))
}
