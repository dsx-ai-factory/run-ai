/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_global "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runv1_prometheus "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	kai_consts "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"

	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const controlPlaneImagePullSecretName = "runai-reg-creds"

var _ = Describe("Prometheuses", func() {
	var (
		ctx       = context.Background()
		k8sReader client.WithWatch
	)

	format.MaxLength = 0

	BeforeEach(func() {
		k8sReader = fake.NewClientBuilder().
			WithScheme(scheme_init.Scheme()).
			WithObjects(
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: consts.MonitoringNamespace,
					},
				},
				// Prometheus Operator Deployment
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      consts.PrometheusOperatorDeploymentName,
						Namespace: consts.MonitoringNamespace,
					},
				},
			).
			Build()
	})

	type TestCase struct {
		runaiConfig  *runv1.RunaiConfig
		expected     *monitoringv1.Prometheus
		extraObjects []client.Object
	}

	cases := map[string]TestCase{
		"basic": func() TestCase {
			promSpec := monitoringv1.PrometheusSpec{
				EvaluationInterval: "1m",
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas: ptr.To(int32(2)),
					PodMonitorNamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"k": "v",
						},
					},
					PodMonitorSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"k": "v",
						},
					},
					ScrapeInterval: "1m",
					RemoteWrite: []monitoringv1.RemoteWriteSpec{
						{
							URL: "http://localhost:1234",
						},
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"memory": resource.MustParse("1Gi"),
						},
						Requests: v1.ResourceList{
							"memory": resource.MustParse("1Gi"),
							"cpu":    resource.MustParse("1"),
						},
					},
					Affinity:    nil,
					Tolerations: []v1.Toleration{},
					Image:       ptr.To("image"),
					Storage: &monitoringv1.StorageSpec{
						EmptyDir: &v1.EmptyDirVolumeSource{
							Medium: v1.StorageMediumMemory,
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "secret",
						},
					},
				},
				Retention:     "1d",
				RetentionSize: "1KB",
			}

			return TestCase{
				runaiConfig: &runv1.RunaiConfig{
					Spec: runv1.RunaiConfigSpec{
						Global: &runv1_global.Global{},
						Internal: &runv1_internal.Internal{
							Global: &runv1_internal.Global{
								Platform: &runv1_internal.Platform{
									KubernetesFlavor: runv1_internal.KubernetesFlavorVanilla,
								},
								ControlPlane: &runv1_internal.ControlPlane{
									ClusterUUID: "uuid",
								},
								// The chart delivers the control-plane pull secret through this list.
								AdditionalImagePullSecrets: []string{controlPlaneImagePullSecretName},
							},
							Development: &runv1_internal.Development{},
						},
						Prometheus: &runv1_prometheus.Prometheus{
							Spec:    runv1_prometheus.Spec(promSpec),
							Enabled: ptr.To(true),
						},
					},
				},
				expected: &monitoringv1.Prometheus{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "runai",
						Namespace: "runai",
					},
					Spec: func() monitoringv1.PrometheusSpec {
						spec := promSpec.DeepCopy()

						spec.RuleSelector = &metav1.LabelSelector{
							MatchLabels: nil,
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      kai_consts.DefaultAccountingLabelKey,
									Operator: metav1.LabelSelectorOpNotIn,
									Values:   []string{kai_consts.DefaultAccountingLabelValue},
								},
							},
						}
						spec.RuleNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemForRKEMonitoringNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}
						spec.SecurityContext = &v1.PodSecurityContext{
							SELinuxOptions:      nil,
							WindowsOptions:      nil,
							RunAsUser:           ptr.To(int64(10000)),
							RunAsGroup:          nil,
							RunAsNonRoot:        ptr.To(true),
							SupplementalGroups:  nil,
							FSGroup:             nil,
							Sysctls:             nil,
							FSGroupChangePolicy: nil,
							SeccompProfile: &v1.SeccompProfile{
								Type:             v1.SeccompProfileTypeRuntimeDefault,
								LocalhostProfile: nil,
							},
						}
						spec.ServiceAccountName = "runai-prometheus"
						spec.ExternalLabels = map[string]string{
							"prometheus":         "",
							"prometheus_replica": "",
							"__replica__":        "$(POD_NAME)",
							"cluster":            "uuid",
						}
						spec.ReplicaExternalLabelName = ptr.To("__replica__")
						spec.ServiceMonitorSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      kai_consts.DefaultAccountingLabelKey,
									Operator: metav1.LabelSelectorOpNotIn,
									Values:   []string{kai_consts.DefaultAccountingLabelValue},
								},
							},
						}
						spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}

						spec.PodMonitorSelector = &metav1.LabelSelector{}
						spec.PodMonitorNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}

						spec.ImagePullSecrets = []v1.LocalObjectReference{
							{
								Name: controlPlaneImagePullSecretName,
							},
						}

						return *spec
					}(),
				},
			}
		}(),
		"setExternalLabels": func() TestCase {
			promSpec := monitoringv1.PrometheusSpec{
				EvaluationInterval: "1m",
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas: ptr.To(int32(2)),
					PodMonitorNamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"k": "v",
						},
					},
					ExternalLabels: map[string]string{
						"prometheus":         "p",
						"prometheus_replica": "pr",
					},
					PodMonitorSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"k": "v",
						},
					},
					ScrapeInterval: "1m",
					RemoteWrite: []monitoringv1.RemoteWriteSpec{
						{
							URL: "http://localhost:1234",
						},
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"memory": resource.MustParse("1Gi"),
						},
						Requests: v1.ResourceList{
							"memory": resource.MustParse("1Gi"),
							"cpu":    resource.MustParse("1"),
						},
					},
					Affinity:    nil,
					Tolerations: []v1.Toleration{},
					Image:       ptr.To("image"),
					Storage: &monitoringv1.StorageSpec{
						EmptyDir: &v1.EmptyDirVolumeSource{
							Medium: v1.StorageMediumMemory,
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "secret",
						},
					},
				},
				Retention:     "1d",
				RetentionSize: "1KB",
			}

			return TestCase{
				runaiConfig: &runv1.RunaiConfig{
					Spec: runv1.RunaiConfigSpec{
						Internal: &runv1_internal.Internal{
							Global: &runv1_internal.Global{
								Platform: &runv1_internal.Platform{
									KubernetesFlavor: runv1_internal.KubernetesFlavorVanilla,
								},
								ControlPlane: &runv1_internal.ControlPlane{
									ClusterUUID: "uuid",
								},
								// The chart delivers the control-plane pull secret through this list.
								AdditionalImagePullSecrets: []string{controlPlaneImagePullSecretName},
							},
							Development: &runv1_internal.Development{},
						},
						Prometheus: &runv1_prometheus.Prometheus{
							Spec:    runv1_prometheus.Spec(promSpec),
							Enabled: ptr.To(true),
						},
						Global: &runv1_global.Global{},
					},
				},
				expected: &monitoringv1.Prometheus{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "runai",
						Namespace: "runai",
					},
					Spec: func() monitoringv1.PrometheusSpec {
						spec := promSpec.DeepCopy()

						spec.RuleSelector = &metav1.LabelSelector{
							MatchLabels: nil,
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      kai_consts.DefaultAccountingLabelKey,
									Operator: metav1.LabelSelectorOpNotIn,
									Values:   []string{kai_consts.DefaultAccountingLabelValue},
								},
							},
						}
						spec.RuleNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}
						spec.SecurityContext = &v1.PodSecurityContext{
							SELinuxOptions:      nil,
							WindowsOptions:      nil,
							RunAsUser:           ptr.To(int64(10000)),
							RunAsGroup:          nil,
							RunAsNonRoot:        ptr.To(true),
							SupplementalGroups:  nil,
							FSGroup:             nil,
							Sysctls:             nil,
							FSGroupChangePolicy: nil,
							SeccompProfile: &v1.SeccompProfile{
								Type:             v1.SeccompProfileTypeRuntimeDefault,
								LocalhostProfile: nil,
							},
						}
						spec.ServiceAccountName = "runai-prometheus"
						spec.ExternalLabels = map[string]string{
							"prometheus":         "p",
							"prometheus_replica": "pr",
							"__replica__":        "$(POD_NAME)",
							"cluster":            "uuid",
						}
						spec.ReplicaExternalLabelName = ptr.To("__replica__")
						spec.ServiceMonitorSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      kai_consts.DefaultAccountingLabelKey,
									Operator: metav1.LabelSelectorOpNotIn,
									Values:   []string{kai_consts.DefaultAccountingLabelValue},
								},
							},
						}
						spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}
						spec.PodMonitorSelector = &metav1.LabelSelector{}
						spec.PodMonitorNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}

						spec.ImagePullSecrets = []v1.LocalObjectReference{
							{
								Name: controlPlaneImagePullSecretName,
							},
						}

						return *spec
					}(),
				},
			}
		}(),
		"restrictedAffinity": func() TestCase {
			promSpec := monitoringv1.PrometheusSpec{
				EvaluationInterval: "1m",
				CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
					Replicas: ptr.To(int32(2)),
					PodMonitorNamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"k": "v",
						},
					},
					PodMonitorSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"k": "v",
						},
					},
					ScrapeInterval: "1m",
					RemoteWrite: []monitoringv1.RemoteWriteSpec{
						{
							URL: "http://localhost:1234",
						},
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"memory": resource.MustParse("1Gi"),
						},
						Requests: v1.ResourceList{
							"memory": resource.MustParse("1Gi"),
							"cpu":    resource.MustParse("1"),
						},
					},
					Affinity: &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
								{
									Weight: 1,
									Preference: v1.NodeSelectorTerm{
										MatchExpressions: []v1.NodeSelectorRequirement{
											{
												Key:      "node-role.kubernetes.io/runai-system",
												Operator: v1.NodeSelectorOpExists,
											},
										},
									},
								},
							},
							RequiredDuringSchedulingIgnoredDuringExecution: nil,
						},
						PodAffinity:     nil,
						PodAntiAffinity: nil,
					},
					Tolerations: []v1.Toleration{},
					Image:       ptr.To("image"),
					Storage: &monitoringv1.StorageSpec{
						EmptyDir: &v1.EmptyDirVolumeSource{
							Medium: v1.StorageMediumMemory,
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "secret",
						},
					},
				},
				Retention:     "1d",
				RetentionSize: "1KB",
			}

			return TestCase{
				runaiConfig: &runv1.RunaiConfig{
					Spec: runv1.RunaiConfigSpec{
						Internal: &runv1_internal.Internal{
							Global: &runv1_internal.Global{
								Platform: &runv1_internal.Platform{
									KubernetesFlavor: runv1_internal.KubernetesFlavorVanilla,
								},
								ControlPlane: &runv1_internal.ControlPlane{
									ClusterUUID: "uuid",
								},
								// The chart delivers the control-plane pull secret through this list.
								AdditionalImagePullSecrets: []string{controlPlaneImagePullSecretName},
							},
							Development: &runv1_internal.Development{},
						},
						Prometheus: &runv1_prometheus.Prometheus{
							Spec:    runv1_prometheus.Spec(promSpec),
							Enabled: ptr.To(true),
						},
						Global: &runv1_global.Global{
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
										{
											Weight: 1,
											Preference: v1.NodeSelectorTerm{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      "node-role.kubernetes.io/runai-system",
														Operator: v1.NodeSelectorOpExists,
													},
												},
											},
										},
									},
									RequiredDuringSchedulingIgnoredDuringExecution: nil,
								},
								PodAffinity:     nil,
								PodAntiAffinity: nil,
							},
						},
					},
				},
				expected: &monitoringv1.Prometheus{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "runai",
						Namespace: "runai",
					},
					Spec: func() monitoringv1.PrometheusSpec {
						spec := promSpec.DeepCopy()

						spec.RuleSelector = &metav1.LabelSelector{
							MatchLabels: nil,
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      kai_consts.DefaultAccountingLabelKey,
									Operator: metav1.LabelSelectorOpNotIn,
									Values:   []string{kai_consts.DefaultAccountingLabelValue},
								},
							},
						}
						spec.RuleNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}
						spec.SecurityContext = &v1.PodSecurityContext{
							SELinuxOptions:      nil,
							WindowsOptions:      nil,
							RunAsUser:           ptr.To(int64(10000)),
							RunAsGroup:          nil,
							RunAsNonRoot:        ptr.To(true),
							SupplementalGroups:  nil,
							FSGroup:             nil,
							Sysctls:             nil,
							FSGroupChangePolicy: nil,
							SeccompProfile: &v1.SeccompProfile{
								Type:             v1.SeccompProfileTypeRuntimeDefault,
								LocalhostProfile: nil,
							},
						}
						spec.ServiceAccountName = "runai-prometheus"
						spec.ExternalLabels = map[string]string{
							"prometheus":         "",
							"prometheus_replica": "",
							"__replica__":        "$(POD_NAME)",
							"cluster":            "uuid",
						}
						spec.ReplicaExternalLabelName = ptr.To("__replica__")
						spec.ServiceMonitorSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      kai_consts.DefaultAccountingLabelKey,
									Operator: metav1.LabelSelectorOpNotIn,
									Values:   []string{kai_consts.DefaultAccountingLabelValue},
								},
							},
						}
						spec.ServiceMonitorNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}
						spec.PodMonitorSelector = &metav1.LabelSelector{}
						spec.PodMonitorNamespaceSelector = &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "kubernetes.io/metadata.name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										runv1.DefaultNamespace,
										consts.OpenshiftMonitoringNamespace,
										consts.MonitoringNamespace,
										consts.KubeSystemNamespace,
										consts.DGXCSystemNamespace,
										consts.MonitoringNamespace,
									},
								},
							},
						}

						spec.ImagePullSecrets = []v1.LocalObjectReference{
							{
								Name: controlPlaneImagePullSecretName,
							},
						}

						spec.Affinity.PodAntiAffinity = &v1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: v1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{
												"prometheus": "runai",
											},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						}

						return *spec
					}(),
				},
			}
		}(),
	}

	for caseName, caseDetails := range cases {
		It(fmt.Sprintf("should propagate prometheus configuration correctly: %s", caseName), func() {
			reader := client.Reader(k8sReader)
			if len(caseDetails.extraObjects) > 0 {
				reader = fake.NewClientBuilder().
					WithScheme(scheme_init.Scheme()).
					WithObjects(
						&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: consts.MonitoringNamespace}},
						&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: consts.PrometheusOperatorDeploymentName, Namespace: consts.MonitoringNamespace}},
					).
					WithObjects(caseDetails.extraObjects...).
					Build()
			}
			resource := NewPrometheusResource("prometheus", reader, caseDetails.runaiConfig)
			Expect(resource).ToNot(BeNil())

			obj, err := resource.Object(ctx)
			Expect(obj).To(BeAssignableToTypeOf(&monitoringv1.Prometheus{}))
			Expect(err).ToNot(HaveOccurred())

			objJSON, err := json.MarshalIndent(obj, "", "  ")
			Expect(err).ToNot(HaveOccurred())

			expectedJSON, err := json.MarshalIndent(caseDetails.expected, "", "  ")
			Expect(err).ToNot(HaveOccurred())

			err = os.WriteFile("expected.json", expectedJSON, 0644)
			Expect(err).ToNot(HaveOccurred())
			err = os.WriteFile("actual.json", objJSON, 0644)
			Expect(err).ToNot(HaveOccurred())

			Expect(cmp.Diff(expectedJSON, objJSON)).To(BeEmpty())

			Expect(string(objJSON)).To(Equal(string(expectedJSON)), caseName)
		})
	}
})

var _ = Describe("shouldEnableOTLPReceiver", func() {
	var ctx = context.Background()

	makeRunaiConfig := func(enabled *bool) *runv1.RunaiConfig {
		return &runv1.RunaiConfig{
			Spec: runv1.RunaiConfigSpec{
				Prometheus: &runv1_prometheus.Prometheus{
					Config: runv1_prometheus.Config{EnableOTLPReceiver: enabled},
				},
			},
		}
	}

	makeReader := func(knativeAvailable bool) client.Reader {
		data := fmt.Sprintf("dependencies:\n  optional:\n    knative:\n      available: %v\n", knativeAvailable)
		return fake.NewClientBuilder().
			WithScheme(scheme_init.Scheme()).
			WithObjects(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: "runai-public"},
				Data:       map[string]string{"runai-public": data},
			}).
			Build()
	}

	It("returns true, nil when flag is enabled and Knative is available", func() {
		enabled, err := shouldEnableOTLPReceiver(ctx, makeRunaiConfig(ptr.To(true)), makeReader(true))
		Expect(err).ToNot(HaveOccurred())
		Expect(enabled).To(BeTrue())
	})

	It("returns false, nil when flag is enabled but Knative is unavailable", func() {
		enabled, err := shouldEnableOTLPReceiver(ctx, makeRunaiConfig(ptr.To(true)), makeReader(false))
		Expect(err).ToNot(HaveOccurred())
		Expect(enabled).To(BeFalse())
	})

	It("returns false, nil when flag is disabled", func() {
		enabled, err := shouldEnableOTLPReceiver(ctx, makeRunaiConfig(ptr.To(false)), makeReader(true))
		Expect(err).ToNot(HaveOccurred())
		Expect(enabled).To(BeFalse())
	})

	It("returns an error when flag is enabled but the runai-public ConfigMap is missing", func() {
		emptyReader := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		_, err := shouldEnableOTLPReceiver(ctx, makeRunaiConfig(ptr.To(true)), emptyReader)
		Expect(err).To(HaveOccurred())
	})
})
