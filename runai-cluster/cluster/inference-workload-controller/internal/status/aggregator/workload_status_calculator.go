/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import (
	"context"
	"net/url"

	"github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/aggregator"
	corev1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	Ready        = "Ready"
	Pending      = "Pending"
	Initializing = "Initializing"
	Creating     = "Creating"
	Deleting     = "Deleting"
	Failed       = "Failed"
	Degraded     = "Degraded"
)

var log = logf.Log.WithName("status-calculator")

type StatusCalculator struct {
	workload                  *job.JobWorkload
	knativeService            *knv1.Service
	services                  map[string]*corev1.Service
	ingresses                 map[string]*network.Ingress
	routes                    map[string]*openshiftV1.Route
	httpRoutes                map[string]*gatewayv1.HTTPRoute
	secrets                   map[string]*corev1.Secret
	latestCreatedRevisionPods []corev1.Pod
	buildError                *string
}

type InferenceWorkloadStatusCalculater interface {
	aggregator.WorkloadStatusCalculater
	WithKnativeService(services *knv1.Service)
	WithLatestCreatedRevisionPods(revisionPods []corev1.Pod)
	WithBuildError(buildError *string)
}

func NewStatusCalculator() InferenceWorkloadStatusCalculater {
	return &StatusCalculator{
		services:       make(map[string]*corev1.Service),
		knativeService: nil,
		routes:         make(map[string]*openshiftV1.Route),
		ingresses:      make(map[string]*network.Ingress),
		httpRoutes:     make(map[string]*gatewayv1.HTTPRoute),
		secrets:        make(map[string]*corev1.Secret),
	}
}

//
// setters for the resources
//

func (sc *StatusCalculator) WithServices(services []*corev1.Service) {
	for _, service := range services {
		sc.services[service.Name] = service
	}
}

func (sc *StatusCalculator) WithSecrets(secrets []*corev1.Secret) {
	for _, secret := range secrets {
		sc.secrets[secret.Name] = secret
	}
}

func (sc *StatusCalculator) WithKnativeService(knativeService *knv1.Service) {
	sc.knativeService = knativeService
}

func (sc *StatusCalculator) WithBuildError(buildError *string) {
	sc.buildError = buildError
}

func (sc *StatusCalculator) WithLatestCreatedRevisionPods(latestCreatedRevisionPods []corev1.Pod) {
	sc.latestCreatedRevisionPods = latestCreatedRevisionPods
}

func (sc *StatusCalculator) WithWorkload(workload *job.JobWorkload) {
	sc.workload = workload
}

func (sc *StatusCalculator) WithJob(_ jobs.JobAdapter) {
	// not implemented
	// not needed for job when running with inference workload
}

func (sc *StatusCalculator) WithComputeDomain(_ *v1beta1.ComputeDomain) {
	// not implemented
	// not needed for inference workload
}

func (sc *StatusCalculator) WithResources(objects []client.Object) {
	for _, resource := range objects {
		switch resourceType := resource.(type) {
		case *corev1.Service:
			sc.services[resourceType.Name] = resourceType
		case *openshiftV1.Route:
			sc.routes[resourceType.Name] = resourceType
		case *network.Ingress:
			sc.ingresses[resourceType.Name] = resourceType
		case *gatewayv1.HTTPRoute:
			sc.httpRoutes[resourceType.Name] = resourceType
		default:
			// current log framework does not support logs level in convent way
			// for now it am adding it direct to string // remove this wheN moving to `zap`
			log.Info("Warning supported resource type", "resource", resource)
			continue
		}
	}
}

// Clear clears the status, and the resources added to struct before,
func (sc *StatusCalculator) Clear() {
	sc.knativeService = nil
	sc.services = make(map[string]*corev1.Service)
	sc.routes = make(map[string]*openshiftV1.Route)
	sc.ingresses = make(map[string]*network.Ingress)
	sc.httpRoutes = make(map[string]*gatewayv1.HTTPRoute)
	sc.workload = nil
	sc.latestCreatedRevisionPods = nil
	sc.buildError = nil
}

//
// Calculate calculates the status of the workload
//

func (sc *StatusCalculator) CalculateAll(ctx context.Context, aggr *resource.Aggregator) job.JobStatus {

	if sc.workload == nil {
		sc.workload = &job.JobWorkload{}
	}

	sc.handleExportedPorts()

	sc.handleJobUrls()

	sc.handleLatestRevisionNames(ctx)

	return sc.CalculatePhaseAndConditions(ctx, aggr)
}

func (sc *StatusCalculator) CalculatePhaseAndConditions(ctx context.Context, aggr *resource.Aggregator) job.JobStatus {
	sc.handlePhaseAndPhaseMessage(ctx)

	return sc.workload.Status
}

//
// helper functions
//

func (sc *StatusCalculator) handleExportedPorts() {
	if sc.services == nil {
		return
	}

	serviceStatuses := sc.workload.Status.NetworkStatus.ServiceStatuses

	for _, service := range sc.services {
		serviceStatus := common.ServiceStatus{
			Name: service.Name,
			Type: service.Spec.Type,
		}
		for _, port := range service.Spec.Ports {
			externalPort := port.Port
			if port.NodePort != 0 {
				externalPort = port.NodePort
			}
			exposedServicePort := common.ExposePort{
				Name:       port.Name,
				Protocol:   port.Protocol,
				Port:       externalPort,
				TargetPort: port.TargetPort,
			}
			serviceStatus.Ports = append(serviceStatus.Ports, exposedServicePort)
		}
		if serviceStatuses == nil {
			serviceStatuses = make([]common.ServiceStatus, 0)
		}
		serviceStatuses = append(serviceStatuses, serviceStatus)
	}
}

func (sc *StatusCalculator) handleJobUrls() {
	urlSet := make(map[string]struct{})
	var externalUrls []string

	addURL := func(u *url.URL) {
		s := u.String()
		if _, exists := urlSet[s]; !exists {
			urlSet[s] = struct{}{}
			externalUrls = append(externalUrls, s)
		}
	}

	if sc.ingresses != nil {
		for _, ingress := range sortIngressesByName(sc.ingresses) {
			if len(ingress.Spec.Rules) > 0 {
				ingressUrl, err := BuildURLFromIngress(ingress)
				if err != nil {
					log.Error(err, "Failed to build URL from ingress", "ingress", ingress.Name)
					ingressUrl = &url.URL{Host: "http://"}
				}
				addURL(ingressUrl)
			}
		}
	}

	if sc.routes != nil {
		for _, route := range sortRoutesByName(sc.routes) {
			routeUrl, err := BuildURLFromRoute(route)
			if err != nil {
				log.Error(err, "Failed to build URL from route", "route", route.Name)
				routeUrl = &url.URL{Host: "http://"}
			}
			addURL(routeUrl)
		}
	}

	if sc.httpRoutes != nil {
		for _, httpRoute := range sortHTTPRoutesByName(sc.httpRoutes) {
			if httpRoute == nil {
				continue
			}
			httpRouteUrl, err := BuildURLFromHTTPRoute(httpRoute)
			if err != nil {
				log.Error(err, "Failed to build URL from httpRoute", "httpRoute", httpRoute.Name)
				httpRouteUrl = &url.URL{Host: "http://"}
			}
			addURL(httpRouteUrl)
		}
	}

	if sc.knativeService != nil && sc.knativeService.Status.URL != nil {
		addURL(sc.knativeService.Status.URL.URL())
	}

	sc.workload.Status.InferenceStatus.ExternalUrls = externalUrls
}

func (sc *StatusCalculator) WithExpectedResources(_ []common.Resource) {
	// not implemented
	// no use for expected resources in inference workload
}

func (sc *StatusCalculator) handleLatestRevisionNames(_ context.Context) {
	if sc.knativeService == nil {
		return
	}

	latestCreatedRevisionName := sc.knativeService.Status.LatestCreatedRevisionName
	if latestCreatedRevisionName != "" {
		sc.workload.Status.LatestCreatedRevisionName = &latestCreatedRevisionName
	}

	latestReadyRevisionName := sc.knativeService.Status.LatestReadyRevisionName
	if latestReadyRevisionName != "" {
		sc.workload.Status.LatestReadyRevisionName = &latestReadyRevisionName
	}
}
