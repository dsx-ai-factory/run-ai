/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

func ServiceForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		name      = operands.RunaiApplicationControllerName
		namespace = runaiConfig.Namespace

		config = runaiConfig.Spec.ApplicationController
	)

	service := &v1.Service{}

	// Get the existing service if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, service)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	service.ObjectMeta.Name = name
	service.ObjectMeta.Namespace = namespace

	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = map[string]string{}
	}

	if service.GetAnnotations() == nil {
		service.SetAnnotations(map[string]string{})
	}

	if runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift {
		service.Annotations["service.beta.openshift.io/serving-cert-secret-name"] = tlsSecretName
	}

	service.ObjectMeta.Labels["app"] = operands.RunaiApplicationControllerName
	service.Spec.Type = config.Service.Type
	service.Spec.Ports = []v1.ServicePort{
		{
			Protocol:   v1.ProtocolTCP,
			Name:       config.Service.Metrics.PortName,
			Port:       int32(*config.Service.Metrics.Port),
			TargetPort: intstr.FromInt32(int32(*config.Service.Metrics.TargetPort)),
		},
		{
			Protocol:   v1.ProtocolTCP,
			Name:       "webhook",
			Port:       443,
			TargetPort: intstr.FromInt32(9443),
		},
	}
	service.Spec.Selector = map[string]string{
		"app": operands.RunaiApplicationControllerName,
	}

	return service, nil
}

func ServiceStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	service, err := ServiceForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if service.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("service [%s] is missing",
			service.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
