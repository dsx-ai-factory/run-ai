/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
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
	config := s.runaiConfig.Spec.Agent
	projectControllerConfig := s.runaiConfig.Spec.ProjectController
	namespace := s.runaiConfig.Namespace

	service := &v1.Service{}
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

	// On OpenShift the serving certificate for the project/department validating
	// webhook is minted by the platform into tlsSecretName and the matching CA is
	// injected into the webhook configuration via the inject-cabundle annotation.
	if s.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1internal.KubernetesFlavorOpenShift {
		if service.Annotations == nil {
			service.SetAnnotations(map[string]string{})
		}
		service.Annotations["service.beta.openshift.io/serving-cert-secret-name"] = tlsSecretName
	}

	service.Spec.Type = v1.ServiceTypeClusterIP
	service.Spec.Ports = []v1.ServicePort{
		{Protocol: v1.ProtocolTCP, Name: config.Metrics.PortName, Port: config.Metrics.Port, TargetPort: intstr.FromInt32(config.Metrics.Port)},
		{Protocol: v1.ProtocolTCP, Name: "webhook",
			Port:       int32(*projectControllerConfig.Service.Webhook.Port),
			TargetPort: intstr.FromInt32(int32(*projectControllerConfig.Service.Webhook.TargetPort)),
		},
	}
	service.Spec.Selector = map[string]string{"app": DeploymentName}
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
