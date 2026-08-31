/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/deployers/helm_deployer"
	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/providers"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApplicationController struct {
	client    client.Client
	factories map[string]common.ApplicationFactory
	providers map[string]common.ApplicationSourceProvider

	helmWorkers HelmWorkers

	// allows us to mock the release client functionality with a memory-based mock
	releaseClientFactory helm_deployer.ReleaseClientFactory

	// allows us to mock the chart downloader functionality with a memory-based mock
	downloaderFactory helm_deployer.DownloaderFactory
}

func NewApplicationController(client client.Client) *ApplicationController {
	helmSource := providers.NewHelmSourceProvider()
	return &ApplicationController{
		client: client,

		factories: map[string]common.ApplicationFactory{
			helmSource.GetApplicationSource(): helmSource,
		},
		providers: map[string]common.ApplicationSourceProvider{
			helmSource.GetApplicationSource(): helmSource,
		},
		releaseClientFactory: helm_deployer.NewReleaseClient,
		downloaderFactory:    helm_deployer.NewDownloader,
	}
}

// NewApplicationControllerWithChartAuth is like NewApplicationController but downloads
// control-plane-hosted charts (spec.helm.requiresControlPlaneAuth) using chartAuthClient, an HTTP
// client that attaches the cluster's control-plane bearer token. All other charts are unaffected.
func NewApplicationControllerWithChartAuth(client client.Client, chartAuthClient *http.Client) *ApplicationController {
	c := NewApplicationController(client)
	c.downloaderFactory = helm_deployer.NewAuthenticatedDownloaderFactory(chartAuthClient)
	return c
}

func (c *ApplicationController) SetupWithManager(mgr ctrl.Manager) error {
	applicationPredicate := c.CreateApplicationPredicate()

	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Application{})

	// trigger the Start function
	if err := mgr.Add(c); err != nil {
		return err
	}

	// Dynamically add watches for all source resource types
	// Only add watches for resource types whose CRDs are installed in the cluster
	for _, factory := range c.factories {
		sourceResourceType := factory.GetSourceResource()

		// Get the GVK from the scheme for this object type
		gvks, _, err := mgr.GetScheme().ObjectKinds(sourceResourceType)
		if err != nil || len(gvks) == 0 {
			log.Warn().Msgf("Skipping watch for source %s (cannot determine GVK from scheme): %v", factory.GetApplicationSource(), err)
			continue
		}

		gvk := gvks[0]

		// Check if the CRD is installed by trying to get REST mapping
		_, err = mgr.GetRESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			log.Warn().Msgf("Skipping watch for source %s (CRD not installed in cluster): %s", factory.GetApplicationSource(), err.Error())
			continue
		}

		log.Info().Msgf("Adding watch for source %s", factory.GetApplicationSource())
		controllerBuilder = controllerBuilder.Watches(
			sourceResourceType,
			handler.EnqueueRequestsFromMapFunc(c.mapSourceResourceToApplication),
			builder.WithPredicates(c.createSourceResourcePredicate(factory)),
		)
	}

	// Watch application resources for status updates
	controllerBuilder = controllerBuilder.
		Watches(
			&v1alpha1.ExternalWorkload{},
			handler.EnqueueRequestsFromMapFunc(c.mapResourceToApplication),
			builder.WithPredicates(applicationPredicate),
		).
		Watches(
			&corev1.PersistentVolumeClaim{},
			handler.EnqueueRequestsFromMapFunc(c.mapResourceToApplication),
			builder.WithPredicates(applicationPredicate),
		).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(c.mapResourceToApplication),
			builder.WithPredicates(applicationPredicate),
		).
		Watches(
			&corev1.Service{},
			handler.EnqueueRequestsFromMapFunc(c.mapResourceToApplication),
			builder.WithPredicates(applicationPredicate),
		).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(c.mapResourceToApplication),
			builder.WithPredicates(applicationPredicate),
		)

	return controllerBuilder.Complete(c)
}

func (c *ApplicationController) Start(ctx context.Context) error {
	c.startHelmWorkers(ctx)
	return nil
}

// startHelmWorkers starts the Helm task worker pool.
func (c *ApplicationController) startHelmWorkers(ctx context.Context) {
	c.helmWorkers.jobsChannel = make(chan HelmJob, helmJobBufSize)

	for i := 0; i < helmWorkerCount; i++ {
		c.helmWorkers.waitGroup.Add(1)
		go c.helmWorker(ctx, c.helmWorkers.jobsChannel)
	}

	// (Todo) this section can be robusted to avoid race condition, it's a corner case so not done yet
	go func() {
		<-ctx.Done()
		close(c.helmWorkers.jobsChannel)
		c.helmWorkers.waitGroup.Wait()
		log.Info().Msg("helm worker pool stopped")
	}()
}

func (c *ApplicationController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Debug().Msgf("Reconciling application %s/%s", req.Name, req.Namespace)

	application := &v1alpha1.Application{}
	err := c.getApplicationForRequest(ctx, req, application)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error().Msgf("Could not find application <%v> from incoming request: %v, err: %v", req.Name, req, err.Error())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Check if this is the run.ai/application annotation is provided, which indicates controller install mode
	if application.Annotations != nil && application.Annotations[AnnotationRunaiApp] == "true" {
		if application.Spec.Helm != nil {
			result, helmErr := c.reconcileHelmDeployment(ctx, application)

			if application.DeletionTimestamp.IsZero() {
				if provider, ok := c.providers[providers.ApplicationHelmSource]; ok {
					if statusErr := common.UpdateApplicationStatus(ctx, c.client, application, nil, provider); statusErr != nil {
						log.Error().Msgf("Failed to update workload labels for application %s: %v", application.Name, statusErr)
					}
				}
			}

			return result, helmErr
		}

		log.Error().Msgf("spec.helm is missing for %s", application.Name)
		return ctrl.Result{}, fmt.Errorf("spec.helm is required for %s", AnnotationRunaiApp)
	}

	if application.DeletionTimestamp.IsZero() {
		// Find the provider matching this application's source
		provider := c.getProviderForApplication(application)
		if provider == nil {
			log.Error().Msgf("No provider found for application %s with source: %s", application.Name, application.Spec.Source)
			return ctrl.Result{}, fmt.Errorf("no provider found for application source: %s. It usually indicate that specific provider was not implemented", application.Spec.Source)
		}
		err = common.UpdateApplicationStatus(ctx, c.client, application, nil, provider)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (c *ApplicationController) getApplicationForRequest(ctx context.Context, req ctrl.Request, application *v1alpha1.Application) (err error) {
	err = c.client.Get(ctx, req.NamespacedName, application)
	if err != nil {
		return err
	}
	return nil
}

// getProviderForApplication finds the provider matching the application's source
func (c *ApplicationController) getProviderForApplication(application *v1alpha1.Application) common.ApplicationSourceProvider {
	provider, ok := c.providers[application.Spec.Source]
	if !ok {
		return nil
	}
	return provider
}

// mapSourceResourceToApplication maps source resources (Secrets, DGDs) to Application reconcile requests
func (c *ApplicationController) mapSourceResourceToApplication(ctx context.Context, sourceResource client.Object) []reconcile.Request {
	namespace := sourceResource.GetNamespace()

	// Try each factory to extract application name
	for _, factory := range c.factories {
		shouldCreate, applicationName, err := factory.ShouldCreateApplication(ctx, c.client, sourceResource)
		if err != nil {
			log.Debug().Msgf("Error checking if source resource %s should create application: %v", sourceResource.GetName(), err)
			continue
		}
		if shouldCreate && applicationName != "" {
			return []reconcile.Request{{
				NamespacedName: types.NamespacedName{
					Name:      applicationName,
					Namespace: namespace,
				},
			}}
		}
	}

	log.Debug().Msgf("Source resource %s does not map to an application", sourceResource.GetName())
	return []reconcile.Request{}
}

// mapResourceToApplication maps application resources (EWs, PVCs, etc.) to Application reconcile requests
func (c *ApplicationController) mapResourceToApplication(ctx context.Context, resource client.Object) []reconcile.Request {
	namespace := resource.GetNamespace()
	if !common.IsApplicationNamespace(ctx, c.client, resource.GetNamespace()) {
		return []reconcile.Request{}
	}

	// Try each provider to extract application name.
	// We assume that only one provider is relevant for each resource
	var applicationName string
	for _, provider := range c.providers {
		applicationName = provider.GetApplicationName(resource)
		if applicationName != "" {
			break
		}
	}

	if applicationName == "" {
		log.Warn().Msgf("Cannot get application name from resource kind %s name %s", resource.GetObjectKind(), resource.GetName())
		return []reconcile.Request{}
	}

	application := &v1alpha1.Application{}
	err := c.client.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      applicationName,
	}, application)
	if err != nil {
		log.Warn().Msgf("Cannot get application %s", applicationName)
		return []reconcile.Request{}
	}

	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{
			Name:      applicationName,
			Namespace: namespace,
		},
	}}
}

// CreateApplicationPredicate creates a predicate that validates resources against all source providers
func (c *ApplicationController) CreateApplicationPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return c.validateResourceFromAnySource(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return c.validateResourceFromAnySource(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return c.validateResourceFromAnySource(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return c.validateResourceFromAnySource(e.Object)
		},
	}
}

// validateResourceFromAnySource checks if resource matches any of the configured source providers
func (c *ApplicationController) validateResourceFromAnySource(resource client.Object) bool {
	for _, provider := range c.providers {
		if provider.MatchesSource(resource) {
			return true
		}
	}
	return false
}

// createSourceResourcePredicate creates a predicate for a specific source resource type
func (c *ApplicationController) createSourceResourcePredicate(factory common.ApplicationFactory) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return factory.FilterSourceResource(context.Background(), c.client, e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return factory.FilterSourceResource(context.Background(), c.client, e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false // Don't trigger on delete
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return factory.FilterSourceResource(context.Background(), c.client, e.Object)
		},
	}
}

// NewApplicationControllerForTest creates a controller with an injectable release client
// factory and an initialized (but workerless) helm job channel, for unit tests.
func NewApplicationControllerForTest(
	k8sClient client.Client,
	releaseClientFactory helm_deployer.ReleaseClientFactory,
	downloaderFactory helm_deployer.DownloaderFactory,
) *ApplicationController {
	c := NewApplicationController(k8sClient)
	c.releaseClientFactory = releaseClientFactory
	c.downloaderFactory = downloaderFactory
	c.helmWorkers.jobsChannel = make(chan HelmJob, helmJobBufSize)
	return c
}

// HelmJobsChannel returns the helm job channel, for tests only.
func (c *ApplicationController) HelmJobsChannel() chan HelmJob {
	return c.helmWorkers.jobsChannel
}

// ClearHelmTask removes the running-task entry for an app, simulating worker completion.
// For tests only.
func (c *ApplicationController) ClearHelmTask(appKey types.NamespacedName) {
	c.helmWorkers.runningTasks.Delete(appKey)
}
