/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	swagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("ClusterInfoParser", func() {
	Describe("ParseClusterInfoToServerScheme", func() {
		It("Should parse cluster info", func() {
			// Prepare
			startTime := metav1.Now()
			runaiPublic := runv1.RunaiPublic{
				Version: "1.0.0",
				RunaiConfigStatus: &runv1.RunaiConfigStatus{
					RunaiCustomTypeStatus: runv1.RunaiCustomTypeStatus{
						Conditions: []metav1.Condition{
							{
								Type:    "Ready",
								Status:  "True",
								Reason:  "RunaiConfigReady",
								Message: "RunaiConfig is ready",
							},
						},
						ObservedGeneration: 1,
					},
					OperandStatuses: map[string]runv1.TimedOperandStatus{
						"a": {
							OperandStatus: runv1.OperandStatus{
								Ready: true,
							},
							LastTransitionTime: startTime,
						},
						"b": {
							OperandStatus: runv1.OperandStatus{
								Ready:   false,
								Reasons: []string{"NotReady"},
							},
							LastTransitionTime: startTime,
						},
						"c": {
							OperandStatus: runv1.OperandStatus{
								Ready:   false,
								Reasons: []string{"NotReady"},
							},
							LastTransitionTime: metav1.NewTime(startTime.Add(-2 * unreadyThreshold)),
						},
					},
				},
				Dependencies: &runv1.PublicDependencies{
					Required: map[string]runv1.PublicStatus{
						"gpu-stack": {
							Available: ptr.To[bool](false),
							Reason:    "GPU Stack is not installed",
							Components: map[string]runv1.PublicStatus{
								"nvidia": {
									Available: ptr.To[bool](false),
									Reason:    "Nvidia driver is not installed",
								},
								"amd": {
									Available: ptr.To[bool](false),
									Reason:    "AMD driver is not installed",
								},
							},
						},
					},
				},
			}

			runaiPublicYaml, err := yaml.Marshal(runaiPublic)
			Expect(err).To(BeNil())
			clusterInfoCm := &corev1.ConfigMap{
				Data: map[string]string{
					"runai-public": string(runaiPublicYaml),
				},
			}

			// Act
			parser := NewClusterInfoParser()
			parsedClusterInfo, err := parser.ParseClusterInfoToServerScheme(clusterInfoCm)

			// Assert
			Expect(err).To(BeNil())
			Expect(parsedClusterInfo.Version).To(Equal(runaiPublic.Version))
			statusValue := parsedClusterInfo.Status.Get()
			Expect(statusValue.Dependencies.Required["gpu-stack"].Available).To(Equal(*runaiPublic.Dependencies.Required["gpu-stack"].Available))
			Expect(statusValue.Dependencies.Required["gpu-stack"].Reason).To(Equal(utils.Ref(runaiPublic.Dependencies.Required["gpu-stack"].Reason)))
			Expect(statusValue.Dependencies.Required["gpu-stack"].Components).ToNot(BeNil())
			Expect((*statusValue.Dependencies.Required["gpu-stack"].Components)["nvidia"].Available).To(Equal(*runaiPublic.Dependencies.Required["gpu-stack"].Components["nvidia"].Available))
			Expect((*statusValue.Dependencies.Required["gpu-stack"].Components)["nvidia"].Reason).To(Equal(utils.Ref(runaiPublic.Dependencies.Required["gpu-stack"].Components["nvidia"].Reason)))
			Expect((*statusValue.Dependencies.Required["gpu-stack"].Components)["amd"].Available).To(Equal(*runaiPublic.Dependencies.Required["gpu-stack"].Components["amd"].Available))
			Expect((*statusValue.Dependencies.Required["gpu-stack"].Components)["amd"].Reason).To(Equal(utils.Ref(runaiPublic.Dependencies.Required["gpu-stack"].Components["amd"].Reason)))

			operands := *statusValue.Operands
			Expect(operands["a"].Ready).To(Equal(runaiPublic.RunaiConfigStatus.OperandStatuses["a"].Ready))
			Expect(operands["a"].Reasons).To(BeEmpty())
			Expect(operands["a"].LastTransitionTime.Equal(runaiPublic.RunaiConfigStatus.OperandStatuses["a"].LastTransitionTime.Time)).To(BeTrue())
			Expect(*operands["a"].UnreadyThresholdCrossed).To(BeFalse())
			Expect(operands["b"].Ready).To(Equal(runaiPublic.RunaiConfigStatus.OperandStatuses["b"].Ready))
			Expect(operands["b"].Reasons).To(Equal(runaiPublic.RunaiConfigStatus.OperandStatuses["b"].Reasons))
			Expect(operands["b"].LastTransitionTime.Equal(runaiPublic.RunaiConfigStatus.OperandStatuses["b"].LastTransitionTime.Time)).To(BeTrue())
			Expect(*operands["b"].UnreadyThresholdCrossed).To(BeFalse())
			Expect(operands["c"].Ready).To(Equal(runaiPublic.RunaiConfigStatus.OperandStatuses["c"].Ready))
			Expect(operands["c"].Reasons).To(Equal(runaiPublic.RunaiConfigStatus.OperandStatuses["c"].Reasons))
			Expect(operands["c"].LastTransitionTime.Equal(runaiPublic.RunaiConfigStatus.OperandStatuses["c"].LastTransitionTime.Time)).To(BeTrue())
			Expect(*operands["c"].UnreadyThresholdCrossed).To(BeTrue())
			// Dependencies field is deprecated
			Expect(parsedClusterInfo.Dependencies.IsSet()).To(BeFalse())
		})

		It("rejects invalid workload-injected container resources before reporting them", func() {
			for _, resources := range []*runv1.PublicResourceRequirements{
				{Requests: map[string]string{"cpu": "invalid"}},
				{
					Requests: map[string]string{"memory": "256Mi"},
					Limits:   map[string]string{"memory": "128Mi"},
				},
			} {
				runaiPublic := &runv1.RunaiPublic{
					RunaiConfigStatus: &runv1.RunaiConfigStatus{},
					Configuration: &runv1.PublicConfiguration{
						GitSyncResources: resources,
					},
				}

				_, err := parseClusterInfoStatusToServerScheme(runaiPublic)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("gitSyncResources"))
			}
		})
	})

	Describe("parseClusterInfoConfigToServerScheme", func() {
		It("maps gateway name and namespace when set", func() {
			gwName := "my-gateway"
			gwNs := "gateway-ns"
			ingressEnabled := false
			gatewayAPIEnabled := true
			configuration := &runv1.PublicConfiguration{
				IngressEnabled:    &ingressEnabled,
				GatewayName:       &gwName,
				GatewayNamespace:  &gwNs,
				GatewayAPIEnabled: &gatewayAPIEnabled,
			}

			result := parseClusterInfoConfigToServerScheme(configuration)

			Expect(result).NotTo(BeNil())
			Expect(result.IngressEnabled.Get()).To(Equal(&ingressEnabled))
			Expect(result.GatewayName.Get()).To(Equal(&gwName))
			Expect(result.GatewayNamespace.Get()).To(Equal(&gwNs))
			Expect(result.GatewayAPIEnabled.Get()).To(Equal(&gatewayAPIEnabled))
		})

		It("maps nil gateway fields when not set", func() {
			configuration := &runv1.PublicConfiguration{}

			result := parseClusterInfoConfigToServerScheme(configuration)

			Expect(result).NotTo(BeNil())
			Expect(result.GatewayName.Get()).To(BeNil())
			Expect(result.GatewayNamespace.Get()).To(BeNil())
		})

		It("returns nil when configuration is nil", func() {
			result := parseClusterInfoConfigToServerScheme(nil)
			Expect(result).To(BeNil())
		})

		It("maps init container constructor inputs when set", func() {
			gitSync := "reg.example.com/git-sync:v1"
			s3 := "reg.example.com/goofys:v2"
			oauth := "reg.example.com/oauth2-proxy:v3"
			usernameLoader := "reg.example.com/username-loader:v4"
			redisFQDN := "redis.example.com:6379"
			workloadAuthSecret := "workload-auth-secret" //nolint:gosec // G101: Secret resource name, not a credential value
			sessionStoreType := "cookie"
			oidcIssuer := "https://mock-q7z3k.example.com/auth/realms/runai"
			externalAuthEnabled := false
			openshiftRouteTerminationPolicy := "passthrough"
			distGPUNetworkAcceleration := true
			dinfGPUNetworkAcceleration := false

			configuration := &runv1.PublicConfiguration{
				GitSyncImage:                &gitSync,
				S3FileSystemImage:           &s3,
				S3FileSystemMountCheckImage: &s3,
				RunaiOauth2ProxyImage:       &oauth,
				RunaiOauth2ProxyResources: &runv1.PublicResourceRequirements{
					Requests: map[string]string{"cpu": "100m", "memory": "128Mi"},
					Limits:   map[string]string{"cpu": "100m", "memory": "128Mi"},
				},
				UsernameLoaderResources: &runv1.PublicResourceRequirements{
					Requests: map[string]string{"cpu": "25m"},
				},
				GitSyncResources: &runv1.PublicResourceRequirements{
					Limits: map[string]string{"memory": "64Mi"},
				},
				S3FileSystemResources:                  &runv1.PublicResourceRequirements{},
				RunaiOauth2ProxyUsernameLoaderImage:    &usernameLoader,
				RunaiOauth2ProxyRedisFQDN:              &redisFQDN,
				RunaiOauth2ProxyWorkloadAuthSecretName: &workloadAuthSecret,
				RunaiOauth2ProxySessionStoreType:       &sessionStoreType,
				RunaiOauth2ProxyOidcIssuerUrl:          &oidcIssuer,
				ExternalAuthUrlEnabled:                 &externalAuthEnabled,
				OpenshiftRouteTerminationPolicy:        &openshiftRouteTerminationPolicy,
				DistGPUNetworkAcceleration:             &distGPUNetworkAcceleration,
				DinfGPUNetworkAcceleration:             &dinfGPUNetworkAcceleration,
				CustomCAGit: &runv1.CustomCAConfig{
					Enabled:    ptr.To(true),
					SecretName: ptr.To("ca-git"),
					SecretKey:  ptr.To("ca.crt"),
					Optional:   ptr.To(false),
				},
				CustomCAS3: &runv1.CustomCAConfig{
					Enabled:    ptr.To(true),
					SecretName: ptr.To("ca-s3"),
					SecretKey:  ptr.To("ca.crt"),
					Optional:   ptr.To(true),
				},
				CustomCAOauth2Proxy: &runv1.CustomCAConfig{
					Enabled: ptr.To(false),
				},
			}

			result := parseClusterInfoConfigToServerScheme(configuration)
			Expect(result).NotTo(BeNil())

			Expect(result.GitSyncImage.Get()).To(Equal(&gitSync))
			Expect(result.S3FileSystemImage.Get()).To(Equal(&s3))
			Expect(result.S3FileSystemMountCheckImage.Get()).To(Equal(&s3))
			Expect(result.RunaiOauth2ProxyImage.Get()).To(Equal(&oauth))
			Expect(result.RunaiOauth2ProxyResources.Get()).NotTo(BeNil())
			Expect(result.RunaiOauth2ProxyResources.Get().GetRequests()).To(Equal(map[string]string{"cpu": "100m", "memory": "128Mi"}))
			Expect(result.RunaiOauth2ProxyResources.Get().GetLimits()).To(Equal(map[string]string{"cpu": "100m", "memory": "128Mi"}))
			Expect(result.UsernameLoaderResources.Get().GetRequests()).To(Equal(map[string]string{"cpu": "25m"}))
			Expect(result.GitSyncResources.Get().GetLimits()).To(Equal(map[string]string{"memory": "64Mi"}))
			Expect(result.S3FileSystemResources.Get()).NotTo(BeNil())
			Expect(result.RunaiOauth2ProxyUsernameLoaderImage.Get()).To(Equal(&usernameLoader))
			Expect(result.RunaiOauth2ProxyRedisFQDN.Get()).To(Equal(&redisFQDN))
			Expect(result.RunaiOauth2ProxyWorkloadAuthSecretName.Get()).To(Equal(&workloadAuthSecret))
			Expect(result.RunaiOauth2ProxySessionStoreType.Get()).To(Equal(&sessionStoreType))
			Expect(result.RunaiOauth2ProxyOidcIssuerUrl.Get()).To(Equal(&oidcIssuer))
			Expect(result.ExternalAuthUrlEnabled.Get()).To(Equal(&externalAuthEnabled))
			Expect(result.OpenshiftRouteTerminationPolicy.Get()).To(Equal(swagger.OPENSHIFTROUTETERMINATIONPOLICY_PASSTHROUGH.Ptr()))
			Expect(result.DistGPUNetworkAcceleration.Get()).To(Equal(&distGPUNetworkAcceleration))
			Expect(result.DinfGPUNetworkAcceleration.Get()).To(Equal(&dinfGPUNetworkAcceleration))

			gitCA := result.CustomCAGit.Get()
			Expect(gitCA).NotTo(BeNil())
			Expect(*gitCA.Enabled.Get()).To(BeTrue())
			Expect(*gitCA.SecretName.Get()).To(Equal("ca-git"))
			Expect(*gitCA.SecretKey.Get()).To(Equal("ca.crt"))
			Expect(*gitCA.Optional.Get()).To(BeFalse())

			s3CA := result.CustomCAS3.Get()
			Expect(s3CA).NotTo(BeNil())
			Expect(*s3CA.Enabled.Get()).To(BeTrue())
			Expect(*s3CA.SecretName.Get()).To(Equal("ca-s3"))
			Expect(*s3CA.SecretKey.Get()).To(Equal("ca.crt"))
			Expect(*s3CA.Optional.Get()).To(BeTrue())

			oauthCA := result.CustomCAOauth2Proxy.Get()
			Expect(oauthCA).NotTo(BeNil())
			Expect(*oauthCA.Enabled.Get()).To(BeFalse())
			Expect(oauthCA.SecretName.Get()).To(BeNil())
			Expect(oauthCA.SecretKey.Get()).To(BeNil())
			Expect(oauthCA.Optional.Get()).To(BeNil())
		})

		It("leaves init container constructor inputs unset when configuration is empty", func() {
			result := parseClusterInfoConfigToServerScheme(&runv1.PublicConfiguration{})
			Expect(result).NotTo(BeNil())

			Expect(result.GitSyncImage.Get()).To(BeNil())
			Expect(result.S3FileSystemImage.Get()).To(BeNil())
			Expect(result.S3FileSystemMountCheckImage.Get()).To(BeNil())
			wire, err := result.ToMap()
			Expect(err).NotTo(HaveOccurred())
			Expect(wire).NotTo(HaveKey("s3FileSystemMountCheckImage"))
			Expect(result.RunaiOauth2ProxyImage.Get()).To(BeNil())
			Expect(result.RunaiOauth2ProxyResources.Get()).To(BeNil())
			Expect(result.UsernameLoaderResources.Get()).To(BeNil())
			Expect(result.GitSyncResources.Get()).To(BeNil())
			Expect(result.S3FileSystemResources.Get()).To(BeNil())
			Expect(result.RunaiOauth2ProxyUsernameLoaderImage.Get()).To(BeNil())
			Expect(result.RunaiOauth2ProxyRedisFQDN.Get()).To(BeNil())
			Expect(result.RunaiOauth2ProxyWorkloadAuthSecretName.Get()).To(BeNil())
			Expect(result.RunaiOauth2ProxySessionStoreType.Get()).To(BeNil())
			Expect(result.RunaiOauth2ProxyOidcIssuerUrl.Get()).To(BeNil())
			Expect(result.ExternalAuthUrlEnabled.Get()).To(BeNil())
			Expect(result.OpenshiftRouteTerminationPolicy.Get()).To(BeNil())
			Expect(result.DistGPUNetworkAcceleration.Get()).To(BeNil())
			Expect(result.DinfGPUNetworkAcceleration.Get()).To(BeNil())
			Expect(result.CustomCAGit.IsSet()).To(BeFalse())
			Expect(result.CustomCAS3.IsSet()).To(BeFalse())
			Expect(result.CustomCAOauth2Proxy.IsSet()).To(BeFalse())
		})
	})
})
