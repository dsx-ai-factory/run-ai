/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewServiceResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceResource {
	return &ServiceResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (s *ServiceResource) OperandName() string {
	return s.operandName
}

func (s *ServiceResource) Object(ctx context.Context) (client.Object, error) {
	namespace := s.runaiConfig.Namespace
	config := s.runaiConfig.Spec.ClusterApi

	service := &v1.Service{}

	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, service)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = name
	service.ObjectMeta.Namespace = namespace
	service.ObjectMeta.Labels = appStringMap
	service.Spec.Selector = appStringMap

	if s.runaiConfig.Spec.Internal.Global.ClusterURL != "" {
		service.Spec.Type = v1.ServiceTypeClusterIP
		service.Spec.Ports = []v1.ServicePort{
			{
				Protocol:   v1.ProtocolTCP,
				Name:       servicePortName,
				Port:       serviceInPort,
				TargetPort: intstr.FromInt32(serviceTargetPort),
			},
		}
	} else {
		service.Spec.Type = v1.ServiceTypeNodePort
		service.Spec.Ports = []v1.ServicePort{
			{
				Protocol:   v1.ProtocolTCP,
				Name:       servicePortName,
				NodePort:   int32(s.runaiConfig.Spec.ClusterApi.NodePortService.NodePort),
				Port:       serviceInPort,
				TargetPort: intstr.FromInt32(serviceTargetPort),
			},
		}
	}

	service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{
		Protocol:   v1.ProtocolTCP,
		Name:       metricPortName,
		Port:       int32(*config.Metrics.Port),
		TargetPort: intstr.FromInt32(int32(*config.Metrics.TargetPort)),
	})

	return service, nil
}

func (s *ServiceResource) Status(ctx context.Context) error {
	resource, err := s.Object(ctx)
	if err != nil {
		return err
	}

	if resource.GetUID() == "" {
		return fmt.Errorf("service [%s] is missing", resource.GetName())
	}

	return nil
}

func (s *ServiceResource) Enabled() bool {
	return true
}
