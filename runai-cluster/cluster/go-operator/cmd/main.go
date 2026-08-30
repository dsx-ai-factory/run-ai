/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"os"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/clients"
	legacycleanup "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/legacy-workload-cleanup"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	v1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	controllers "github.com/run-ai/runai/runai-cluster/cluster/go-operator/controllers/run"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	//+kubebuilder:scaffold:imports

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/installation/util"

	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/common/profiling"

	_ "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/metrics"
	appsv1 "k8s.io/api/apps/v1"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	opts := util.SetOptions()

	if opts.EnableProfiler {
		go profiling.RegisterProfiler(opts.ProfilerApiPort)
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts.ZapOptions)))

	signalHandlerContext := ctrl.SetupSignalHandler()
	if opts.LegacyWorkloadCleanup {
		if err := legacycleanup.RunCommand(signalHandlerContext); err != nil {
			setupLog.Error(err, "legacy workload cleanup failed")
			os.Exit(1)
		}
		return
	}

	localClusterConfig := ctrl.GetConfigOrDie()

	localClusterManager, err := ctrl.NewManager(localClusterConfig, ctrl.Options{
		Scheme: scheme_init.Scheme(),
		Metrics: metricsserver.Options{
			BindAddress: opts.MetricsAddr,
		},
		HealthProbeBindAddress: opts.ProbeAddr,
		LeaderElection:         opts.EnableLeaderElection,
		LeaderElectionID:       "8b40c1bd.run.ai",
		Cache:                  cacheOptions(),
	})
	if err != nil {
		setupLog.Error(err, "unable to create local manager")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder
	if err := localClusterManager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := localClusterManager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	go func() {
		if err := localClusterManager.Start(signalHandlerContext); err != nil {
			setupLog.Error(err, "unable to run localmanager")
			os.Exit(1)
		}
	}()

	localClusterClient := kc.New(localClusterManager.GetClient(), localClusterManager.GetAPIReader(), localClusterConfig)
	localClusterDiscoveryClient, err := localClusterClient.NewDiscoveryClient()
	if err != nil {
		setupLog.Error(err, "unable to create discovery client")
		os.Exit(1)
	}

	if opts.NMCRemoteClusterAPIServerURL != "" {
		remoteClusterManager, err := initRemoteClusterManager(opts, signalHandlerContext)
		if err != nil {
			setupLog.Error(err, "unable to create remote cluster manager")
			os.Exit(1)
		}

		remoteClusterHybridClient := kc.New(remoteClusterManager.GetClient(), remoteClusterManager.GetAPIReader(), remoteClusterManager.GetConfig())
		remoteClusterDiscoveryClient, err := remoteClusterHybridClient.NewDiscoveryClient()
		if err != nil {
			setupLog.Error(err, "unable to create discovery client")
			os.Exit(1)
		}

		err = controllers.NewRemoteClusterReconciler(localClusterClient, remoteClusterDiscoveryClient, opts.Verbosity).
			SetupWithManager(remoteClusterManager)
		if err != nil {
			setupLog.Error(err, "unable to create remote cluster controller")
			os.Exit(1)
		}
		err = controllers.NewRunAIConfigReconciler(localClusterClient,
			localClusterDiscoveryClient,
			remoteClusterHybridClient,
			remoteClusterDiscoveryClient,
			opts.Verbosity,
			opts.DisableRunaiConfigFinalizer).SetupWithManager(localClusterManager)
		if err != nil {
			setupLog.Error(err, "unable to create runai config controller")
			os.Exit(1)
		}
	} else {
		err = controllers.NewRunAIConfigReconciler(localClusterClient,
			localClusterDiscoveryClient,
			localClusterClient,
			localClusterDiscoveryClient,
			opts.Verbosity,
			opts.DisableRunaiConfigFinalizer).SetupWithManager(localClusterManager)
		if err != nil {
			setupLog.Error(err, "unable to create runai config controller")
			os.Exit(1)
		}
	}

	err = controllers.NewRemoteConfigReconciler(localClusterClient, opts.Verbosity).SetupWithManager(localClusterManager)
	if err != nil {
		setupLog.Error(err, "unable to create remote config controller")
		os.Exit(1)
	}

	clusterServiceClient := clients.NewClusterServiceClient()
	err = controllers.NewRemoteConfigSyncReconciler(localClusterClient, opts.Verbosity, clusterServiceClient).SetupWithManager(localClusterManager)
	if err != nil {
		setupLog.Error(err, "unable to create remote config sync controller")
		os.Exit(1)
	}

	select {}
}

func initRemoteClusterManager(opts *util.Options, ctx context.Context) (ctrl.Manager, error) {
	remoteClusterConfig := ctrl.GetConfigOrDie()
	remoteClusterConfig.Host = opts.NMCRemoteClusterAPIServerURL

	remoteClusterManager, err := ctrl.NewManager(remoteClusterConfig, ctrl.Options{
		Scheme: scheme_init.Scheme(),
		Metrics: metricsserver.Options{
			BindAddress: opts.RemoteClusterManagerMetricsAddr,
		},
		HealthProbeBindAddress: opts.RemoteClusterManagerProbeAddr,
		LeaderElection:         opts.RemoteClusterManagerEnableLeaderElection,
		LeaderElectionID:       "8b40c1bd-remote-cluster.run.ai",
		Cache:                  cacheOptions(),
	})
	if err != nil {
		setupLog.Error(err, "unable to define remote manager")
		return nil, err
	}

	if err := remoteClusterManager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return nil, err
	}

	if err := remoteClusterManager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	go func() {
		if err := remoteClusterManager.Start(ctx); err != nil {
			setupLog.Error(err, "problem running remote manager")
			os.Exit(1)
		}
	}()

	return remoteClusterManager, nil
}

func cacheOptions() cache.Options {
	return cache.Options{
		Scheme: scheme_init.Scheme(),
		// Namespaces specified in this map will be cached by default.
		// This means that if a resource is requested within any of those namespaces,
		// an informer will be created for that namespace+resource pair.
		DefaultNamespaces: map[string]cache.Config{
			runaiv1.DefaultNamespace:                   {},
			consts.MonitoringNamespace:                 {},
			consts.OpenshiftMonitoringNamespace:        {},
			consts.KubeSystemForRKEMonitoringNamespace: {},
			consts.KnativeNamespace:                    {},
			consts.DGXCSystemNamespace:                 {},
			consts.GPUOperatorNamespace:                {},
			consts.RancherMonitoringSystemNamespace:    {},
			consts.DefaultLWSNamespace:                 {},
		},
		// For resources that we don't know the namespace of, we specify them in the following map.
		// Cluster-scoped resources should have an empty configuration (see &v1.Namespace{})
		// For namespaced resources that we have no knowledge about the namespace they reside in,
		// We specify that we want to cache them in all namespaces (still an informer per-namespace will be created).
		ByObject: map[client.Object]cache.ByObject{
			&v1.Namespace{}:    {},
			&runaiv1.Cluster{}: {},
			&v1.Service{}: {
				Namespaces: map[string]cache.Config{
					cache.AllNamespaces: {},
				},
			},
			&appsv1.Deployment{}: {
				Namespaces: map[string]cache.Config{
					cache.AllNamespaces: {},
				},
			},
			&appsv1.DaemonSet{}: {
				Namespaces: map[string]cache.Config{
					cache.AllNamespaces: {},
				},
			},
		},
	}
}
