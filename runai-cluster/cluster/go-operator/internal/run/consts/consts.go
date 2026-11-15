/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package consts

const (
	// Namespaces
	MonitoringNamespace                 = "monitoring"
	OpenshiftMonitoringNamespace        = "openshift-monitoring"
	KubeSystemForRKEMonitoringNamespace = "kube-system"
	KubeSystemNamespace                 = "kube-system"
	KnativeNamespace                    = "knative-serving"
	DGXCSystemNamespace                 = "dgxc-system"
	RancherMonitoringSystemNamespace    = "cattle-monitoring-system"
	GPUOperatorNamespace                = "gpu-operator"

	// Deployment Names
	PrometheusOperatorDeploymentName        = "prometheus-operator"
	GPUOperatorDeploymentName               = "gpu-operator"
	RancherMonitoringOperatorDeploymentName = "rancher-monitoring-operator"

	// ConfigMap Names
	KnativeFeaturesConfigMapName   = "config-features"
	KnativeAutoScalerConfigMapName = "config-autoscaler"
	KnativeDomainConfigMapName     = "config-domain"
	KnativeNetworkConfigMapName    = "config-network"
	KnativeGatewayConfigMapName    = "config-gateway"

	// KnativeGatewayAPIIngressClass is the Knative ingress class of the net-gateway-api networking layer.
	KnativeGatewayAPIIngressClass = "gateway-api.ingress.networking.knative.dev"
	// KnativeExternalGatewaysKey is the config-gateway key that lists the Gateways serving external traffic.
	KnativeExternalGatewaysKey = "external-gateways"

	// Service Names
	DCGMExporterServiceName = "dcgm-exporter"
)
