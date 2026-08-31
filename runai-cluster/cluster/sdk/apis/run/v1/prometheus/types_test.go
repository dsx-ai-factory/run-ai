/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	cpclusterclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions/remote-write-relabel-configs/mocks"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Prometheus types", func() {
	var (
		internalSpec *runv1_internal.Internal
		prom         *Prometheus
	)

	Describe("SetDefaultsWhereNeeded", func() {
		BeforeEach(func() {
			tenantId := int32(1)
			env := "test"
			releaseName := "release.name"
			internalSpec = &runv1_internal.Internal{
				RemoteConfig: &runv1_internal.ClusterFullConfig{
					ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
						SystemConfig: &cpclusterclient.ClusterSystemConfigData{
							Deactivated: false,
						},
					},
				},
				Global: &runv1_internal.Global{
					Platform: &runv1_internal.Platform{
						KubernetesFlavor: runv1_internal.KubernetesFlavorVanilla,
					},
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
								},
								TenantId:    &tenantId,
								Environment: &env,
								ReleaseName: &releaseName,
							},
						},
					},
				},
			}

			prom = &Prometheus{
				Spec: Spec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						RemoteWrite: []monitoringv1.RemoteWriteSpec{
							{
								URL:  "http://example.com",
								Name: ptr.To("runai"),
							},
							{
								URL: "http://example.org",
							},
						},
					},
				},
			}
		})

		When("CustomCA is enabled", func() {
			BeforeEach(func() {
				internalSpec.Global.CustomCA = &common.CustomCA{
					Enabled: ptr.To[bool](true),
					Secret: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "my-secret",
						},
						Key: "ca.crt",
					},
				}
			})

			It("should set TLSConfig CA configuration only for the first remoteWrite spec", func() {
				prom.SetDefaultsWhereNeeded(internalSpec, nil, nil)

				remoteWrite := prom.Spec.RemoteWrite

				Expect(remoteWrite).To(HaveLen(2))

				Expect(remoteWrite[0].TLSConfig).NotTo(BeNil())
				Expect(remoteWrite[0].TLSConfig.CA).NotTo(BeNil())
				Expect(remoteWrite[0].TLSConfig.CA.Secret).NotTo(BeNil())
				Expect(remoteWrite[0].TLSConfig.CA.Secret.Name).To(Equal("my-secret"))
				Expect(remoteWrite[0].TLSConfig.CA.Secret.Key).To(Equal("ca.crt"))

				Expect(remoteWrite[1].TLSConfig).To(BeNil())
			})
		})

		When("CustomCA is disabled", func() {
			It("should not set TLSConfig CA configuration", func() {
				prom.SetDefaultsWhereNeeded(internalSpec, nil, ptr.To(int32(3)))

				remoteWrite := prom.Spec.RemoteWrite

				Expect(remoteWrite).To(HaveLen(2))

				Expect(remoteWrite[1].TLSConfig).To(BeNil())
				Expect(remoteWrite[0].TLSConfig).To(BeNil())
				Expect(prom.Spec.Replicas).To(Equal(ptr.To[int32](3)))
			})
		})

		It("should set BasicAuth only for the first remoteWrite spec", func() {
			prom.SetDefaultsWhereNeeded(internalSpec, nil, nil)

			remoteWrite := prom.Spec.RemoteWrite

			Expect(remoteWrite).To(HaveLen(2))

			Expect(remoteWrite[0].BasicAuth).NotTo(BeNil())
			Expect(remoteWrite[0].BasicAuth.Username).NotTo(BeNil())
			Expect(remoteWrite[0].BasicAuth.Username.Key).To(Equal("remoteWriteUsername-runai"))
			Expect(remoteWrite[0].BasicAuth.Password).NotTo(BeNil())
			Expect(remoteWrite[0].BasicAuth.Password.Key).To(Equal("remoteWritePassword-runai"))

			Expect(remoteWrite[1].BasicAuth).To(BeNil())
		})
	})
	Describe("SetDefaultsWhereNeededMultipleRemoteWrites", func() {
		BeforeEach(func() {
			tenantId := int32(1)
			env := "test"
			releaseName := "release.name"
			internalSpec = &runv1_internal.Internal{
				RemoteConfig: &runv1_internal.ClusterFullConfig{
					ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
						SystemConfig: &cpclusterclient.ClusterSystemConfigData{
							Deactivated: false,
						},
					},
				},
				Global: &runv1_internal.Global{
					ControlPlane: &runv1_internal.ControlPlane{
						Config: &runv1_internal.ClusterControlPlaneConfig{
							ClusterControlPlaneConfig: cpclusterclient.ClusterControlPlaneConfig{
								MetricsRemoteWrites: []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
									{
										Label: ptr.To("runai"),
										Url:   ptr.To("http://runai-remote.com"),
										OrgID: ptr.To("my-org"),
										Credentials: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
											Username: ptr.To("my-username"),
											Password: ptr.To("my-password"),
										},
									},
									{
										Label: ptr.To("applicative_runai"),
										Url:   ptr.To("http://app-remote.com"),
										Credentials: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
											Username: ptr.To("my-app"),
											Password: ptr.To("my-password-app"),
										},
									},
								},
								TenantId:    &tenantId,
								Environment: &env,
								ReleaseName: &releaseName,
							},
						},
					},
				},
			}

			prom = &Prometheus{
				Spec: Spec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						RemoteWrite: []monitoringv1.RemoteWriteSpec{
							{
								URL: "http://example.com",
							},
							{
								URL: "http://example.org",
							},
						},
					},
				},
			}
		})
		It("should add all remote writes", func() {

			regExpHandlerMock := mocks.NewRemoteWriteRelabelConfigsHandler(GinkgoT())
			regExpHandlerMock.On("GetRemoteWriteRelabelConfigs", "runai").Return(
				[]monitoringv1.RelabelConfig{{Regex: ".*"}})
			regExpHandlerMock.On("GetRemoteWriteRelabelConfigs", "applicative_runai").Return(
				[]monitoringv1.RelabelConfig{{Regex: "app.*"}})
			prom.Spec.SetDefaultsWhereNeeded(internalSpec, regExpHandlerMock, nil)
			remoteWrite := prom.Spec.RemoteWrite

			Expect(remoteWrite).To(HaveLen(4))

			Expect(*remoteWrite[2].Name).To(Equal("runai"))
			Expect(remoteWrite[2].URL).To(Equal("http://runai-remote.com"))
			Expect(remoteWrite[1].Name).To(BeNil())
			Expect(*remoteWrite[3].Name).To(Equal("applicative_runai"))
			Expect(remoteWrite[3].URL).To(Equal("http://app-remote.com"))
		})
		It("should not add applicative write since it don't have a regexp", func() {

			regExpHandlerMock := mocks.NewRemoteWriteRelabelConfigsHandler(GinkgoT())
			regExpHandlerMock.On("GetRemoteWriteRelabelConfigs", "runai").Return(
				[]monitoringv1.RelabelConfig{{Regex: ".*"}})
			regExpHandlerMock.On("GetRemoteWriteRelabelConfigs", "applicative_runai").Return(
				nil)
			prom.Spec.SetDefaultsWhereNeeded(internalSpec, regExpHandlerMock, nil)
			remoteWrite := prom.Spec.RemoteWrite
			Expect(remoteWrite).To(HaveLen(3))
			Expect(*remoteWrite[2].Name).To(Equal("runai"))
			Expect(remoteWrite[2].URL).To(Equal("http://runai-remote.com"))
			Expect(remoteWrite[1].Name).To(BeNil())
			Expect(remoteWrite[0].Name).To(BeNil())
		})

	})
	Describe("SetDefaultsWhereNeeded rule groups", func() {
		BeforeEach(func() {
			tenantId := int32(1)
			env := "test"
			releaseName := "release.name"

			internalSpec = &runv1_internal.Internal{
				RemoteConfig: &runv1_internal.ClusterFullConfig{
					ClusterFullConfigData: cpclusterclient.ClusterFullConfigData{
						SystemConfig: &cpclusterclient.ClusterSystemConfigData{
							Deactivated: false,
						},
					},
				},
				Global: &runv1_internal.Global{
					ControlPlane: &runv1_internal.ControlPlane{
						Config: &runv1_internal.ClusterControlPlaneConfig{
							ClusterControlPlaneConfig: cpclusterclient.ClusterControlPlaneConfig{
								MetricsRemoteWrites: []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
									{
										Label: ptr.To("runai"),
										Url:   ptr.To("http://runai-remote.com"),
										OrgID: ptr.To("my-org"),
										Credentials: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
											Username: ptr.To("my-username"),
											Password: ptr.To("my-password"),
										},
									},
									{
										Label: ptr.To("applicative_runai"),
										Url:   ptr.To("http://app-remote.com"),
										Credentials: &cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
											Username: ptr.To("my-app"),
											Password: ptr.To("my-password-app"),
										},
									},
								},
								TenantId:    &tenantId,
								Environment: &env,
								ReleaseName: &releaseName,
							},
						},
					},
				},
			}

			prom = &Prometheus{
				Spec: Spec{
					CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
						RemoteWrite: []monitoringv1.RemoteWriteSpec{
							{
								URL: "http://example.com",
							},
							{
								URL: "http://example.org",
							},
						},
					},
				},
				AdditionalAlertLabels: &map[string]string{
					"label-test1": "label-value1",
					"label-test2": "label-value2",
				},
			}
		})
		It("should combine existing and additional alert labels", func() {
			prom.SetDefaultsWhereNeeded(internalSpec, nil, nil)
			ruleGroups := prom.RuleGroups
			for _, rg := range ruleGroups {
				if rg.Name == "runai-alerts" {
					for _, rule := range rg.Rules {
						Expect(rule.Labels["label-test1"]).To(Equal("label-value1"))
						Expect(rule.Labels["label-test2"]).To(Equal("label-value2"))
						Expect(rule.Labels["severity"]).To(Not(BeNil()))
					}
				}
			}
		})
		It("should give priority to our severity label over customer's additional alert labels", func() {
			(*prom.AdditionalAlertLabels)["severity"] = "additionalAlertLabelSeverity"
			prom.SetDefaultsWhereNeeded(internalSpec, nil, nil)
			ruleGroups := prom.RuleGroups
			for _, rg := range ruleGroups {
				if rg.Name == "runai-alerts" {
					for _, rule := range rg.Rules {
						Expect(rule.Labels["severity"]).To(Not(Equal("additionalAlertLabelSeverity")))
					}
				}
			}
		})
	})
})
