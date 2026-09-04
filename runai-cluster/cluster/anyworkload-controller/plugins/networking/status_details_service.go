/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"fmt"
	"sort"

	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceReadyReason     = "ServiceReady"
	serviceNotFoundReason  = "ServiceNotFound"
	serviceGetFailedReason = "ServiceGetFailed"
)

func (p *Plugin) buildServiceDetails(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
	k8sClient client.Client,
) (endpoints.ServiceStatus, error) {
	status := endpoints.ServiceStatus{Name: name}
	if name == "" {
		status.Phase = endpoints.PhaseFailed
		status.Reason = serviceGetFailedReason
		status.Message = "service has no resolvable name"
		return status, nil
	}

	svc := &corev1.Service{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: runner.Namespace}, svc)
	switch {
	case apierrors.IsNotFound(err):
		status.Phase = endpoints.PhaseCreating
		status.Reason = serviceNotFoundReason
		status.Message = fmt.Sprintf("service %s not found", name)
		return status, nil
	case err != nil:
		status.Phase = endpoints.PhaseFailed
		status.Reason = serviceGetFailedReason
		status.Message = fmt.Sprintf("failed to get service %s: %v", name, err)
		return status, fmt.Errorf("failed to get service %s: %w", name, err)
	}

	status.Phase = endpoints.PhaseReady
	status.Reason = serviceReadyReason
	status.Message = fmt.Sprintf("service %s exists", name)
	status.Type = string(svc.Spec.Type)
	status.ClusterIP = svc.Spec.ClusterIP
	status.Ports = servicePortDetails(svc)
	status.LoadBalancer = loadBalancerDetails(svc)
	return status, nil
}

func servicePortDetails(svc *corev1.Service) []endpoints.ServicePortStatus {
	if svc == nil || len(svc.Spec.Ports) == 0 {
		return nil
	}
	ports := make([]endpoints.ServicePortStatus, 0, len(svc.Spec.Ports))
	for _, port := range svc.Spec.Ports {
		exposedPort := port.Port
		if port.NodePort != 0 {
			exposedPort = port.NodePort
		}
		ports = append(ports, endpoints.ServicePortStatus{
			Name:       port.Name,
			Port:       exposedPort,
			TargetPort: port.TargetPort,
			Protocol:   string(port.Protocol),
		})
	}
	return ports
}

func loadBalancerDetails(svc *corev1.Service) *endpoints.LoadBalancerStatus {
	if svc == nil || svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return nil
	}
	status := &endpoints.LoadBalancerStatus{}
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.IP == "" && ingress.Hostname == "" {
			continue
		}
		status.Ingress = append(status.Ingress, endpoints.LoadBalancerIngress{
			IP:       ingress.IP,
			Hostname: ingress.Hostname,
		})
	}
	sort.Slice(status.Ingress, func(i, j int) bool {
		return loadBalancerIngressKey(status.Ingress[i]) < loadBalancerIngressKey(status.Ingress[j])
	})
	return status
}

func loadBalancerIngressKey(ingress endpoints.LoadBalancerIngress) string {
	if ingress.IP != "" {
		return ingress.IP
	}
	return ingress.Hostname
}
