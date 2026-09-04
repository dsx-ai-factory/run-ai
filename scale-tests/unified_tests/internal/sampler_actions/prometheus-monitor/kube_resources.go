/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus_monitor

import (
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
)

func generatePrometheusMonitorResources(clusterConfig *scale.ClusterConfig, metricLabels map[string]string) []client.Object {
	return []client.Object{
		// Dedicated Namespace
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: TestResourcesNamespace,
			},
		},
		// ServiceMonitor
		&monitoringv1.ServiceMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai-prometheus",
				Namespace: TestResourcesNamespace,
				Labels: map[string]string{
					"app.kubernetes.io/name": "runai-prometheus-monitor",
				},
			},
			Spec: monitoringv1.ServiceMonitorSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"operated-prometheus": "true",
					},
				},
				NamespaceSelector: monitoringv1.NamespaceSelector{
					MatchNames: []string{"runai"},
				},
				Endpoints: []monitoringv1.Endpoint{
					{
						Port: "web",
					},
				},
			},
		},
		// ServiceAccount
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai-prometheus-monitor",
				Namespace: TestResourcesNamespace,
			},
		},
		// ClusterRoleBinding
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "runai-prometheus-monitor",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "runai-prometheus-monitor",
					Namespace: TestResourcesNamespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		// Secret
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "prometheus-remote-write-credentials",
				Namespace: TestResourcesNamespace,
			},
			Data: map[string][]byte{
				"username": []byte(clusterConfig.PrometheusRemoteWriteUsername),
				"password": []byte(clusterConfig.PrometheusRemoteWritePassword),
			},
		},
		// Prometheus
		&monitoringv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai-prometheus-monitor",
				Namespace: TestResourcesNamespace,
			},
			Spec: monitoringv1.PrometheusSpec{
				RuleNamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "kubernetes.io/metadata.name",
							Operator: metav1.LabelSelectorOpIn,
							Values: []string{
								TestResourcesNamespace,
								"openshift-monitoring",
								"monitoring",
								"runai",
							},
						},
					},
				},
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					ServiceMonitorNamespaceSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "kubernetes.io/metadata.name",
								Operator: metav1.LabelSelectorOpIn,
								Values: []string{
									TestResourcesNamespace,
									"monitoring",
									"openshift-monitoring",
									"runai",
									"runai-backend",
								},
							},
						},
					},
					ServiceMonitorSelector: &metav1.LabelSelector{},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("500Mi"),
						},
					},

					ScrapeInterval:     "10s",
					ServiceAccountName: "runai-prometheus-monitor",
					PodMetadata: &monitoringv1.EmbeddedObjectMetadata{
						Labels: map[string]string{
							// Set a label with a generated value to restart the Prometheus pod if it exists
							"run.ai/test-monitoring-restart": generateRandomString(5),
						},
					},
					RemoteWrite: func() []monitoringv1.RemoteWriteSpec {
						if !clusterConfig.PrometheusRemoteWriteEnabled {
							return nil
						}

						additionalLabelsRelabelConfigs := []monitoringv1.RelabelConfig{}
						for key, value := range metricLabels {
							additionalLabelsRelabelConfigs = append(additionalLabelsRelabelConfigs, monitoringv1.RelabelConfig{
								TargetLabel: fmt.Sprintf("test_%s", key),
								Replacement: ptr.To(value),
							})
						}

						return []monitoringv1.RemoteWriteSpec{
							{
								URL: clusterConfig.PrometheusRemoteWriteURL,
								WriteRelabelConfigs: append([]monitoringv1.RelabelConfig{
									{
										SourceLabels: []monitoringv1.LabelName{"__name__", "type"},
										Action:       "keep",
										Regex:        "(kube|container)_.*|.*;stats",
									},
									{
										SourceLabels: []monitoringv1.LabelName{"namespace"},
										Action:       "keep",
										Regex:        "(runai|runai-backend)",
									},
								}, additionalLabelsRelabelConfigs...),
								BasicAuth: &monitoringv1.BasicAuth{
									Username: corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "prometheus-remote-write-credentials",
										},
										Key: "username",
									},
									Password: corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "prometheus-remote-write-credentials",
										},
										Key: "password",
									},
								},
							},
						}
					}(),
				},
			},
		},
	}
}
