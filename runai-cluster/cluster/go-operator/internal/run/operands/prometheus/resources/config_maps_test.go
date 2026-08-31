/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources_test

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"

	cpclusterclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions/remote-write-relabel-configs/mocks"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/prometheus/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ConfigMaps", func() {
	Describe("ConfigMapForRunaiConfig", func() {
		var (
			operandName    string
			fakeKubeClient client.Client
			runaiConfig    *runaiv1.RunaiConfig
			internalSpec   *runv1_internal.Internal
		)

		BeforeEach(func() {
			operandName = "prometheus"
			fakeKubeClient = fake.NewClientBuilder().WithObjects(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster-monitoring-config",
					Namespace: consts.OpenshiftMonitoringNamespace,
				},
			}).Build()

			tenantId := int32(1)
			environment := "test"
			releaseName := "release.name"
			runaiConfig = &runaiv1.RunaiConfig{}
			internalSpec = &runv1_internal.Internal{
				Global: &runv1_internal.Global{
					ControlPlane: &runv1_internal.ControlPlane{
						Config: &runv1_internal.ClusterControlPlaneConfig{
							ClusterControlPlaneConfig: cpclusterclient.ClusterControlPlaneConfig{
								MetricsRemoteWrites: []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
									{
										Label: ptr.To("runai"),
										Url:   ptr.To("http://example.com"),
										OrgID: ptr.To("my-org"),
										Credentials: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
											Username: ptr.To("my-username"),
											Password: ptr.To("my-password"),
										},
									},
									{
										Label: ptr.To("applicative_runai"),
										Url:   ptr.To("http://app.com"),
										Credentials: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
											Username: ptr.To("my-app"),
											Password: ptr.To("my-password-app"),
										},
									},
								},
								TenantId:    &tenantId,
								Environment: &environment,
								ReleaseName: &releaseName,
							},
						},
					},
				},
			}
			runaiConfig.Spec.Internal = internalSpec
			runaiConfig.SetDefaultsWhereNeeded()
		})

		When("Openshift is enabled", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor = runv1_internal.KubernetesFlavorOpenShift
			})
			When("labels are set", func() {
				It("check runai_replica", func() {
					runaiConfig.SetDefaultsWhereNeeded()
					result, err := resources.NewConfigMap(operandName, fakeKubeClient, runaiConfig).Object(context.Background())
					Expect(err).NotTo(HaveOccurred())
					Expect(result).NotTo(BeNil())

					configMap, ok := result.(*v1.ConfigMap)
					Expect(ok).To(BeTrue())

					clusterMonitoringConfig := getClusterMonitoringConfig(configMap)

					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite).To(HaveLen(2))
					for _, rw := range clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite {
						found := false
						for _, rls := range rw.WriteRelabelConfigs {
							if rls.TargetLabel == "runai_replica" && utils.Deref(rls.Replacement) == "A" {
								found = true
								break
							}
						}
						Expect(found).To(BeTrue())
					}
				})
			})
			When("CustomCA is enabled", func() {
				BeforeEach(func() {
					runaiConfig.Spec.Internal.Global.CustomCA = &common.CustomCA{
						Enabled: ptr.To(true),
						Secret: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "my-secret",
							},
							Key: "ca.crt",
						},
					}
				})

				It("should set TLSConfig CA configuration for the first remoteWrite spec", func() {
					runaiConfig.SetDefaultsWhereNeeded()

					result, err := resources.NewConfigMap(operandName, fakeKubeClient, runaiConfig).Object(context.Background())
					Expect(err).NotTo(HaveOccurred())
					Expect(result).NotTo(BeNil())

					configMap, ok := result.(*v1.ConfigMap)
					Expect(ok).To(BeTrue())

					clusterMonitoringConfig := getClusterMonitoringConfig(configMap)

					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite).To(HaveLen(2))

					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].TLSConfig).NotTo(BeNil())
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].TLSConfig.CA).NotTo(BeNil())
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].TLSConfig.CA.Secret).NotTo(BeNil())
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].TLSConfig.CA.Secret.Name).To(Equal("my-secret"))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].TLSConfig.CA.Secret.Key).To(Equal("ca.crt"))
				})
			})
			When("CustomCA is disabled", func() {
				It("should not set TLSConfig CA configuration", func() {
					runaiConfig.SetDefaultsWhereNeeded()

					result, err := resources.NewConfigMap(operandName, fakeKubeClient, runaiConfig).Object(context.Background())
					Expect(err).NotTo(HaveOccurred())
					Expect(result).NotTo(BeNil())

					configMap, ok := result.(*v1.ConfigMap)
					Expect(ok).To(BeTrue())

					clusterMonitoringConfig := getClusterMonitoringConfig(configMap)

					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite).To(HaveLen(2))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].TLSConfig).To(Equal(&monitoringv1.SafeTLSConfig{}))
				})
				It("should not set TLSConfig CA configuration with multiple remote writes", func() {
					runaiConfig.SetDefaultsWhereNeeded()
					regExpHandlerMock := mocks.NewRemoteWriteRelabelConfigsHandler(GinkgoT())
					regExpHandlerMock.On("GetRemoteWriteRelabelConfigs", "runai").Return(
						[]monitoringv1.RelabelConfig{{Regex: ".*"}})
					regExpHandlerMock.On("GetRemoteWriteRelabelConfigs", "applicative_runai").Return(
						[]monitoringv1.RelabelConfig{{Regex: "app.*"}})
					runaiConfig.Spec.Prometheus.Spec.SetDefaultsWhereNeeded(internalSpec, regExpHandlerMock, nil)
					result, err := resources.NewConfigMap(operandName, fakeKubeClient, runaiConfig).Object(context.Background())
					Expect(err).NotTo(HaveOccurred())
					Expect(result).NotTo(BeNil())

					configMap, ok := result.(*v1.ConfigMap)
					Expect(ok).To(BeTrue())

					clusterMonitoringConfig := getClusterMonitoringConfig(configMap)

					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite).To(HaveLen(2))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].TLSConfig).To(Equal(&monitoringv1.SafeTLSConfig{}))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].Name).To(Equal("runai"))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].BasicAuth.Username.Key).To(Equal("remoteWriteUsername-runai"))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[0].BasicAuth.Password.Key).To(Equal("remoteWritePassword-runai"))

					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[1].TLSConfig).To(Equal(&monitoringv1.SafeTLSConfig{}))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[1].Name).To(Equal("applicative_runai"))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[1].BasicAuth.Username.Key).To(Equal("remoteWriteUsername-applicative_runai"))
					Expect(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[1].BasicAuth.Password.Key).To(Equal("remoteWritePassword-applicative_runai"))
				})

			})
		})
	})
})

func getClusterMonitoringConfig(configMap *v1.ConfigMap) *manifests.ClusterMonitoringConfiguration {
	ginkgo.GinkgoHelper()

	Expect(configMap.Data).NotTo(BeEmpty())
	Expect(configMap.Data["config.yaml"]).NotTo(BeEmpty())

	clusterMonitoringConfig := &manifests.ClusterMonitoringConfiguration{}
	err := yaml.Unmarshal([]byte(configMap.Data["config.yaml"]), clusterMonitoringConfig)
	Expect(err).NotTo(HaveOccurred())

	return clusterMonitoringConfig
}
