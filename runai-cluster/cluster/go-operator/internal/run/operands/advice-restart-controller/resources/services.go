/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	metricPortName = "metrics"
)

type ServiceResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewServiceResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *ServiceResource {
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
	var (
		name             = s.operandName
		serviceNamespace = s.runaiConfig.GetNamespace()
		config           = s.runaiConfig.Spec.AdviceRestartController
	)

	service := &v1.Service{}
	if err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: serviceNamespace,
	}, service); err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.Name = name
	service.Namespace = serviceNamespace

	if service.Labels == nil {
		service.Labels = map[string]string{}
	}
	service.Labels["app"] = name

	service.Spec.Selector = map[string]string{
		"app": name,
	}

	service.Spec.Ports = []v1.ServicePort{
		{
			Protocol:   v1.ProtocolTCP,
			Name:       "webhook",
			Port:       int32(*config.Service.Webhook.Port),                         // #nosec G115 -- port value is bounded by config defaults (443)
			TargetPort: intstr.FromInt32(int32(*config.Service.Webhook.TargetPort)), // #nosec G115 -- port value is bounded by config defaults (8443)
		},
		{
			Protocol:   v1.ProtocolTCP,
			Name:       metricPortName,
			Port:       int32(*config.Metric.Port),                         // #nosec G115 -- port value is bounded by config defaults (8383)
			TargetPort: intstr.FromInt32(int32(*config.Metric.TargetPort)), // #nosec G115 -- port value is bounded by config defaults (8383)
		},
	}

	if s.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runaiv1_internal.KubernetesFlavorOpenShift {
		if service.Annotations == nil {
			service.Annotations = map[string]string{}
		}
		service.Annotations["service.beta.openshift.io/serving-cert-secret-name"] = tlsSecretName
	}

	return service, nil
}

func (s *ServiceResource) Status(ctx context.Context) error {
	resource, err := s.Object(ctx)
	if err != nil {
		return err
	}

	service := resource.(*v1.Service)
	if service.GetUID() == "" {
		return fmt.Errorf("service [%s] is missing", service.GetName())
	}

	return nil
}

func (s *ServiceResource) Enabled() bool {
	return true
}
