/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
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
	return &ServiceResource{operandName: operandName, k8sReader: k8sReader, runaiConfig: runaiConfig}
}

func (s *ServiceResource) OperandName() string { return s.operandName }

func (s *ServiceResource) Object(ctx context.Context) (client.Object, error) {
	namespace := s.runaiConfig.Namespace
	config := s.runaiConfig.Spec.ClusterRedis

	service := &v1.Service{}
	service.Spec.Type = v1.ServiceTypeClusterIP
	err := s.k8sReader.Get(ctx, types.NamespacedName{Name: DeploymentName, Namespace: namespace}, service)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = DeploymentName
	service.ObjectMeta.Namespace = namespace

	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = map[string]string{}
	}
	service.ObjectMeta.Labels["app"] = DeploymentName

	service.Spec.Selector = map[string]string{"app": DeploymentName}
	service.Spec.Ports = []v1.ServicePort{
		{
			Name:       "redis",
			Port:       int32(*config.RedisService.Port),
			TargetPort: intstr.FromInt32(int32(*config.RedisService.TargetPort)),
			Protocol:   v1.ProtocolTCP,
		},
	}
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

func (s *ServiceResource) Enabled() bool { return true }
