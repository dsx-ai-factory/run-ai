/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"bytes"
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/ghodss/yaml"
	openshiftV1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var logCB = logger.Log.WithName("common.builder")

type SharedResourcesBuilder struct {
	serviceBuilder      ResourceBuilder
	ExternalUrlsBuilder ExternalBuilder
	PVCBuilder          ResourceBuilder
}

func NewShardResourcesBuilder(service ResourceBuilder, urlsBuilder ExternalBuilder, pvcBuilder ResourceBuilder) *SharedResourcesBuilder {
	return &SharedResourcesBuilder{
		serviceBuilder:      service,
		ExternalUrlsBuilder: urlsBuilder,
		PVCBuilder:          pvcBuilder,
	}
}

// GetPreview returns a list of resource name and kind that will be created on build
func (self *SharedResourcesBuilder) GetPreview(ctx context.Context, workload *job.JobWorkload) ([]common.Resource, error) {
	var resultResources []common.Resource
	pvcList, err := self.PVCBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("unable to get pvc preview resources, %v", err)
	}
	services, err := self.serviceBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("unable to get services preview resources, %v", err)
	}
	externalObj, err := self.ExternalUrlsBuilder.GetPreview(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("unable to get external urls preview resources, %v", err)
	}

	resultResources = append(resultResources, pvcList...)
	resultResources = append(resultResources, services...)
	resultResources = append(resultResources, externalObj...)
	return resultResources, nil
}

func (self *SharedResourcesBuilder) BuildServices(ctx context.Context, workload *job.JobWorkload, options ResourceBuilderOptions) ([]*v1.Service, error) {
	logCB.Info("Generating job services from Workload")
	var services []*v1.Service
	objets, err := self.serviceBuilder.Build(ctx, workload, options)
	if err != nil {
		return nil, fmt.Errorf("unable to build services, %v", err)
	}
	if len(objets) == 0 {
		logCB.Info("No services were generated from Workload")
	}

	for _, obj := range objets {
		if service, ok := obj.(*v1.Service); ok {
			services = append(services, service)
		} else {
			return nil, fmt.Errorf("unable to cast obj to service, got %v", obj)
		}

	}
	logCB.Info("Completed generating job services from Workload")
	return services, nil
}

func (self *SharedResourcesBuilder) BuildExternalUrls(ctx context.Context, workload *job.JobWorkload, options *ExternalBuilderOptions) ([]client.Object, error) {
	logCB.Info("Generating job external urls from Workload")
	objets, err := self.ExternalUrlsBuilder.Build(ctx, workload, options)
	if err != nil {
		return nil, fmt.Errorf("unable to build external urls, %v", err)
	}
	if len(objets) == 0 {
		logCB.Info("No external urls were generated from Workload")
	}
	logCB.Info("Completed generating job external urls from Workload")
	return objets, nil
}

func (self *SharedResourcesBuilder) BuildPVC(ctx context.Context, workload *job.JobWorkload, options ResourceBuilderOptions) ([]client.Object, error) {
	logCB.Info("Generating job pvc's from Workload")
	objets, err := self.PVCBuilder.Build(ctx, workload, options)
	if err != nil {
		return nil, fmt.Errorf("unable to build pvc's, %v", err)
	}
	if len(objets) == 0 {
		logCB.Info("No pvc's were generated from Workload")
	}
	return objets, nil
}

func (self *SharedResourcesBuilder) ToYaml(objects []client.Object, result *bytes.Buffer) error {
	commonResources := make(map[string][]client.Object)
	for _, obj := range objects {
		switch obj.(type) {
		case *v1.Service:
			commonResources[constants.KindService] = append(commonResources[constants.KindService], obj)
		case *v1.PersistentVolumeClaim:
			commonResources[constants.KindPvc] = append(commonResources[constants.KindPvc], obj)
		case *networkingv1.Ingress:
			commonResources[constants.KindIngress] = append(commonResources[constants.KindIngress], obj)
		case *openshiftV1.Route:
			commonResources[constants.KindRoute] = append(commonResources[constants.KindRoute], obj)
		case *v1.Secret:
			commonResources[constants.KindSecret] = append(commonResources[constants.KindSecret], obj)
		case *gatewayv1.HTTPRoute:
			commonResources[constants.KindHTTPRoute] = append(commonResources[constants.KindHTTPRoute], obj)
		default:
			return fmt.Errorf("unknown resource type %v", obj)
		}
	}

	for _, pvcObj := range commonResources[constants.KindPvc] {
		pvc := pvcObj.(*v1.PersistentVolumeClaim)
		pvc.Kind = constants.KindPvc
		pvc.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(pvc)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, serviceObj := range commonResources[constants.KindService] {
		service := serviceObj.(*v1.Service)
		service.Kind = constants.KindService
		service.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(service)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, ingressObj := range commonResources[constants.KindIngress] {
		ingress := ingressObj.(*networkingv1.Ingress)
		ingress.Kind = constants.KindIngress
		ingress.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(ingress)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, routeObj := range commonResources[constants.KindRoute] {
		route := routeObj.(*openshiftV1.Route)
		route.Kind = constants.KindRoute
		route.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(route)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, httpRouteObj := range commonResources[constants.KindHTTPRoute] {
		httpRoute := httpRouteObj.(*gatewayv1.HTTPRoute)
		httpRoute.Kind = constants.KindHTTPRoute
		httpRoute.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(httpRoute)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, secretObj := range commonResources[constants.KindSecret] {
		secret := secretObj.(*v1.Secret)
		secret.Kind = constants.KindSecret
		secret.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(secret)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	return nil
}
