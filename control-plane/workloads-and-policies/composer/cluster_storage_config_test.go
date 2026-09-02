/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clustersApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	sdkcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func reportedCustomCA(enabled bool, secret, key string) clustersApi.ClusterReportedCustomCA {
	ca := clustersApi.ClusterReportedCustomCA{}
	ca.SetEnabled(enabled)
	ca.SetSecretName(secret)
	ca.SetSecretKey(key)
	return ca
}

func reportedResources(requests, limits map[string]string) clustersApi.ClusterReportedResourceRequirements {
	resources := clustersApi.ClusterReportedResourceRequirements{}
	resources.SetRequests(requests)
	resources.SetLimits(limits)
	return resources
}

func clusterWithPlatformType(platformType string) *clustersApi.DisplayedCluster {
	platform := clustersApi.ClusterReportedStatusPlatform{}
	platform.SetType(platformType)

	status := clustersApi.ClusterDisplayedStatus{}
	status.Platform.Set(&platform)

	cluster := clustersApi.DisplayedCluster{}
	cluster.Status.Set(&status)
	return &cluster
}

var _ = Describe("NewClusterStorageConfigFromCluster", func() {
	defaults := DefaultClusterStorageConfig()

	Context("images", func() {
		It("maps the reported git-sync and S3 images through", func() {
			config := clustersApi.ClusterReportedConfig{}
			config.SetGitSyncImage("my.registry/git-sync:custom")
			config.SetS3FileSystemImage("my.registry/goofys:custom")

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			Expect(cfg.GitSyncImage()).To(Equal("my.registry/git-sync:custom"))
			Expect(cfg.S3FsImage()).To(Equal("my.registry/goofys:custom"))
		})

		It("returns empty image strings when the cluster has not reported them", func() {
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))

			Expect(cfg.GitSyncImage()).To(Equal(""))
			Expect(cfg.S3FsImage()).To(Equal(""))
		})

		It("returns empty image strings when the cluster has no reported status", func() {
			cfg := NewClusterStorageConfigFromCluster(&clustersApi.DisplayedCluster{})

			Expect(cfg.GitSyncImage()).To(Equal(""))
			Expect(cfg.S3FsImage()).To(Equal(""))
		})
	})

	Context("oauth2-proxy and username-loader", func() {
		It("uses the pinned public username-loader image by default", func() {
			Expect(defaults.OAuth2ProxyUserLoaderImage()).To(
				Equal(sdkcommon.WorkloadRunaiOauth2ProxyUsernameLoaderImageRef),
			)
		})

		It("maps the reported oauth2 values through", func() {
			config := clustersApi.ClusterReportedConfig{}
			config.SetRunaiOauth2ProxyImage("my.registry/oauth2-proxy:custom")
			resources := clustersApi.ClusterReportedResourceRequirements{}
			resources.SetRequests(map[string]string{"cpu": "100m", "memory": "128Mi"})
			resources.SetLimits(map[string]string{"cpu": "100m", "memory": "128Mi"})
			config.SetRunaiOauth2ProxyResources(resources)
			config.SetRunaiOauth2ProxyUsernameLoaderImage("my.registry/loader:custom")
			config.SetRunaiOauth2ProxyWorkloadAuthSecretName("my-auth-secret")
			config.SetRunaiOauth2ProxySessionStoreType("redis")
			config.SetRunaiOauth2ProxyRedisFQDN("redis.my-ns.svc")
			config.SetSubdomainEnabled(true)
			config.SetExternalAuthUrlEnabled(false)

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			Expect(cfg.OAuth2ProxyImage()).To(Equal("my.registry/oauth2-proxy:custom"))
			actualResources, resourcesErr := cfg.OAuth2ProxyResources()
			Expect(resourcesErr).NotTo(HaveOccurred())
			Expect(actualResources).To(Equal(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			}))
			Expect(cfg.OAuth2ProxyUserLoaderImage()).To(Equal("my.registry/loader:custom"))
			Expect(cfg.OAuth2ProxyAuthSecretName()).To(Equal("my-auth-secret"))
			Expect(cfg.OAuth2ProxySessionStoreType()).To(Equal("redis"))
			Expect(cfg.OAuth2ProxyRedisFQDN()).To(Equal("redis.my-ns.svc"))
			Expect(cfg.SubdomainSupportEnabled()).To(BeTrue())
			Expect(cfg.ExternalAuthUrlEnabled()).To(BeFalse())
		})

		It("leaves the oauth2-proxy image empty but keeps the loader default when unreported", func() {
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))

			Expect(cfg.OAuth2ProxyImage()).To(Equal(""))
			Expect(cfg.OAuth2ProxyUserLoaderImage()).To(Equal(defaults.OAuth2ProxyUserLoaderImage()))
			resources, err := cfg.OAuth2ProxyResources()
			Expect(err).NotTo(HaveOccurred())
			Expect(resources).To(Equal(corev1.ResourceRequirements{}))
		})

		It("surfaces invalid reported oauth2 resource quantities", func() {
			config := clustersApi.ClusterReportedConfig{}
			resources := clustersApi.ClusterReportedResourceRequirements{}
			resources.SetLimits(map[string]string{"memory": "not-a-quantity"})
			config.SetRunaiOauth2ProxyResources(resources)

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))
			_, err := cfg.OAuth2ProxyResources()

			Expect(err).To(MatchError(ContainSubstring(`limits.memory has invalid quantity "not-a-quantity"`)))
		})

		It("keeps the default external-auth and oidc-issuer when unreported", func() {
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))

			Expect(cfg.ExternalAuthUrlEnabled()).To(Equal(defaults.ExternalAuthUrlEnabled()))
			Expect(cfg.OAuth2ProxyOidcIssuerUrl()).To(Equal(defaults.OAuth2ProxyOidcIssuerUrl()))
		})

		It("sources the cluster domain from the displayed cluster", func() {
			cluster := ClusterWithConfig(clustersApi.ClusterReportedConfig{})
			cluster.SetDomain("cluster.example.com")

			cfg := NewClusterStorageConfigFromCluster(cluster)

			Expect(cfg.ClusterDomain()).To(Equal("cluster.example.com"))
		})

		It("uses the reported oauth2 custom CA when present and falls back otherwise", func() {
			config := clustersApi.ClusterReportedConfig{}
			config.SetCustomCAOauth2Proxy(reportedCustomCA(true, "auth-ca", "auth-ca.pem"))

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))
			Expect(cfg.IsCustomCAAuthEnabled()).To(BeTrue())
			Expect(cfg.CustomCAAuthSecret()).To(Equal("auth-ca"))
			Expect(cfg.CustomCAAuthKey()).To(Equal("auth-ca.pem"))

			fallback := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))
			Expect(fallback.CustomCAAuthSecret()).To(Equal(defaults.CustomCAAuthSecret()))
			Expect(fallback.CustomCAAuthKey()).To(Equal(defaults.CustomCAAuthKey()))
		})
	})

	Context("reported container resources", func() {
		It("maps configured resources for every composed container", func() {
			config := clustersApi.ClusterReportedConfig{}
			config.SetRunaiOauth2ProxyResources(reportedResources(map[string]string{"cpu": "10m"}, nil))
			config.SetUsernameLoaderResources(reportedResources(map[string]string{"cpu": "20m"}, nil))
			config.SetGitSyncResources(reportedResources(map[string]string{"cpu": "30m"}, nil))
			config.SetS3FileSystemResources(reportedResources(map[string]string{"cpu": "40m"}, nil))

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			oauth, oauthErr := cfg.OAuth2ProxyResources()
			usernameLoader, usernameLoaderErr := cfg.UsernameLoaderResources()
			gitSync, gitSyncErr := cfg.GitSyncResources()
			s3, s3Err := cfg.S3FileSystemResources()
			Expect(oauthErr).NotTo(HaveOccurred())
			Expect(usernameLoaderErr).NotTo(HaveOccurred())
			Expect(gitSyncErr).NotTo(HaveOccurred())
			Expect(s3Err).NotTo(HaveOccurred())
			Expect(oauth.Requests.Cpu().Equal(resource.MustParse("10m"))).To(BeTrue())
			Expect(usernameLoader.Requests.Cpu().Equal(resource.MustParse("20m"))).To(BeTrue())
			Expect(gitSync.Requests.Cpu().Equal(resource.MustParse("30m"))).To(BeTrue())
			Expect(s3.Requests.Cpu().Equal(resource.MustParse("40m"))).To(BeTrue())
		})

		It("uses only the username-loader fallback when resources are unreported", func() {
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))

			oauth, oauthErr := cfg.OAuth2ProxyResources()
			usernameLoader, usernameLoaderErr := cfg.UsernameLoaderResources()
			gitSync, gitSyncErr := cfg.GitSyncResources()
			s3, s3Err := cfg.S3FileSystemResources()
			Expect(oauthErr).NotTo(HaveOccurred())
			Expect(usernameLoaderErr).NotTo(HaveOccurred())
			Expect(gitSyncErr).NotTo(HaveOccurred())
			Expect(s3Err).NotTo(HaveOccurred())
			Expect(oauth).To(Equal(corev1.ResourceRequirements{}))
			Expect(gitSync).To(Equal(corev1.ResourceRequirements{}))
			Expect(s3).To(Equal(corev1.ResourceRequirements{}))
			Expect(usernameLoader.Requests.Cpu().Equal(resource.MustParse("50m"))).To(BeTrue())
			Expect(usernameLoader.Requests.Memory().Equal(resource.MustParse("50Mi"))).To(BeTrue())
			Expect(usernameLoader.Limits.Cpu().Equal(resource.MustParse("100m"))).To(BeTrue())
			Expect(usernameLoader.Limits.Memory().Equal(resource.MustParse("100Mi"))).To(BeTrue())
		})

		It("distinguishes explicit empty objects from unreported resources", func() {
			config := clustersApi.ClusterReportedConfig{}
			empty := reportedResources(nil, nil)
			config.SetRunaiOauth2ProxyResources(empty)
			config.SetUsernameLoaderResources(empty)
			config.SetGitSyncResources(empty)
			config.SetS3FileSystemResources(empty)
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			oauth, _ := cfg.OAuth2ProxyResources()
			usernameLoader, _ := cfg.UsernameLoaderResources()
			gitSync, _ := cfg.GitSyncResources()
			s3, _ := cfg.S3FileSystemResources()
			Expect(oauth).To(Equal(corev1.ResourceRequirements{}))
			Expect(usernameLoader).To(Equal(corev1.ResourceRequirements{}))
			Expect(gitSync).To(Equal(corev1.ResourceRequirements{}))
			Expect(s3).To(Equal(corev1.ResourceRequirements{}))
		})

		It("retains parse errors for each configured resource field", func() {
			config := clustersApi.ClusterReportedConfig{}
			invalid := reportedResources(nil, map[string]string{"memory": "invalid"})
			config.SetRunaiOauth2ProxyResources(invalid)
			config.SetUsernameLoaderResources(invalid)
			config.SetGitSyncResources(invalid)
			config.SetS3FileSystemResources(invalid)
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			_, oauthErr := cfg.OAuth2ProxyResources()
			_, usernameLoaderErr := cfg.UsernameLoaderResources()
			_, gitSyncErr := cfg.GitSyncResources()
			_, s3Err := cfg.S3FileSystemResources()
			for _, err := range []error{oauthErr, usernameLoaderErr, gitSyncErr, s3Err} {
				Expect(err).To(MatchError(ContainSubstring(`limits.memory has invalid quantity "invalid"`)))
			}
		})

		It("rejects semantically invalid reported resources", func() {
			for _, invalid := range []clustersApi.ClusterReportedResourceRequirements{
				reportedResources(nil, map[string]string{"invalid resource": "1"}),
				reportedResources(nil, map[string]string{"cpu": "-1m"}),
				reportedResources(map[string]string{"memory": "256Mi"}, map[string]string{"memory": "128Mi"}),
			} {
				_, err := parseReportedResourceRequirements(&invalid)
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Context("openshift", func() {
		It("reports openshift when the cluster platform type is openshift", func() {
			cfg := NewClusterStorageConfigFromCluster(clusterWithPlatformType("openshift"))

			Expect(cfg.IsOpenshift()).To(BeTrue())
		})

		It("reports not-openshift for any other platform type", func() {
			cfg := NewClusterStorageConfigFromCluster(clusterWithPlatformType("vanilla"))

			Expect(cfg.IsOpenshift()).To(BeFalse())
		})

		It("reports not-openshift when no platform is reported", func() {
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))

			Expect(cfg.IsOpenshift()).To(BeFalse())
		})
	})

	Context("external routing flags", func() {
		It("defaults ingress to enabled and Gateway API to disabled", func() {
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))

			Expect(cfg.IngressEnabled()).To(BeTrue())
			Expect(cfg.GatewayAPIEnabled()).To(BeFalse())
		})

		It("uses explicitly reported ingress and Gateway API flags", func() {
			config := clustersApi.ClusterReportedConfig{}
			config.SetIngressEnabled(false)
			config.SetGatewayAPIEnabled(true)
			config.SetGatewayName("runai-gateway")
			config.SetGatewayNamespace("gateway-system")

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			Expect(cfg.IngressEnabled()).To(BeFalse())
			Expect(cfg.GatewayAPIEnabled()).To(BeTrue())
			Expect(cfg.GatewayName()).To(Equal("runai-gateway"))
			Expect(cfg.GatewayNamespace()).To(Equal("gateway-system"))
		})

		It("keeps the previous Gateway API inference for clusters that do not report the flag yet", func() {
			config := clustersApi.ClusterReportedConfig{}
			config.SetGatewayName("runai-gateway")

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			Expect(cfg.GatewayAPIEnabled()).To(BeTrue())
		})
	})

	Context("oidc issuer", func() {
		It("uses the cluster-reported issuer when set", func() {
			issuer := "https://cp.example.com/auth/realms/acme"
			reported := clustersApi.ClusterReportedConfig{
				RunaiOauth2ProxyOidcIssuerUrl: *clustersApi.NewNullableString(&issuer),
			}

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(reported))

			Expect(cfg.OAuth2ProxyOidcIssuerUrl()).To(Equal(issuer))
		})

		It("keeps the default issuer when the cluster has not reported one", func() {
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))

			Expect(cfg.OAuth2ProxyOidcIssuerUrl()).To(Equal(defaults.OAuth2ProxyOidcIssuerUrl()))
		})
	})

	Context("custom CA", func() {
		It("falls back to the default CA wiring when the cluster reported none", func() {
			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))

			Expect(cfg.IsCustomCAGitEnabled()).To(Equal(defaults.IsCustomCAGitEnabled()))
			Expect(cfg.CustomCAGitSecret()).To(Equal(defaults.CustomCAGitSecret()))
			Expect(cfg.CustomCAGitKey()).To(Equal(defaults.CustomCAGitKey()))
			Expect(cfg.IsCustomCAS3Enabled()).To(Equal(defaults.IsCustomCAS3Enabled()))
			Expect(cfg.CustomCAS3Secret()).To(Equal(defaults.CustomCAS3Secret()))
			Expect(cfg.CustomCAS3Key()).To(Equal(defaults.CustomCAS3Key()))
		})

		It("uses the reported CA values when present", func() {
			config := clustersApi.ClusterReportedConfig{}
			config.SetCustomCAGit(reportedCustomCA(true, "git-ca", "git-ca.pem"))
			config.SetCustomCAS3(reportedCustomCA(true, "s3-ca", "s3-ca.pem"))

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			Expect(cfg.IsCustomCAGitEnabled()).To(BeTrue())
			Expect(cfg.CustomCAGitSecret()).To(Equal("git-ca"))
			Expect(cfg.CustomCAGitKey()).To(Equal("git-ca.pem"))
			Expect(cfg.IsCustomCAS3Enabled()).To(BeTrue())
			Expect(cfg.CustomCAS3Secret()).To(Equal("s3-ca"))
			Expect(cfg.CustomCAS3Key()).To(Equal("s3-ca.pem"))
		})

		It("falls back per-field when the cluster reported only some CA values", func() {
			gitCA := clustersApi.ClusterReportedCustomCA{}
			gitCA.SetEnabled(true)
			config := clustersApi.ClusterReportedConfig{}
			config.SetCustomCAGit(gitCA)

			cfg := NewClusterStorageConfigFromCluster(ClusterWithConfig(config))

			Expect(cfg.IsCustomCAGitEnabled()).To(BeTrue())
			Expect(cfg.CustomCAGitSecret()).To(Equal(defaults.CustomCAGitSecret()))
			Expect(cfg.CustomCAGitKey()).To(Equal(defaults.CustomCAGitKey()))
		})
	})

	Context("flowing into the built init containers", func() {
		It("ships the reported git-sync image on the init container", func() {
			config := clustersApi.ClusterReportedConfig{}
			config.SetGitSyncImage("my.registry/git-sync:custom")
			info := &GitSyncInfo{
				Config: NewClusterStorageConfigFromCluster(ClusterWithConfig(config)),
				Items:  map[string]*GitSyncItem{"git-0": {Nick: "git-0", Repository: "https://example.com/r.git"}},
			}

			containers, err := BuildGitSyncInitContainers(info)
			Expect(err).To(BeNil())
			Expect(containers).To(HaveLen(1))
			Expect(containers[0].Image).To(Equal("my.registry/git-sync:custom"))
		})
	})
})

var _ = Describe("storage compose with a missing cluster-reported image", func() {
	var noopGitSetter FnGitSyncSetter = func(*GitSyncInfo, any, any) *workloadsApi.Error { return nil }
	var noopS3Setter FnS3Setter = func(*S3Info, any, any) *workloadsApi.Error { return nil }

	emptyReportedConfig := func() ClusterStorageConfig {
		return NewClusterStorageConfigFromCluster(ClusterWithConfig(clustersApi.ClusterReportedConfig{}))
	}

	newComposer := func(cfg ClusterStorageConfig) *FieldsComposer {
		return &FieldsComposer{
			ItemizedNumbering:    make(itemized.ItemizedNumbering),
			ClusterStorageConfig: cfg,
		}
	}

	gitInstance := func(name string, exclude bool) workloadsApi.GitInstance {
		return workloadsApi.GitInstance{
			Name:       *workloadsApi.NewNullableString(convutils.Ref(name)),
			Repository: *workloadsApi.NewNullableString(convutils.Ref("https://example.com/r.git")),
			Exclude:    *workloadsApi.NewNullableBool(convutils.Ref(exclude)),
		}
	}

	s3Instance := func(name string, exclude bool) workloadsApi.S3Instance {
		return workloadsApi.S3Instance{
			Name:    *workloadsApi.NewNullableString(convutils.Ref(name)),
			Bucket:  *workloadsApi.NewNullableString(convutils.Ref("bucket")),
			Path:    *workloadsApi.NewNullableString(convutils.Ref("/data")),
			Exclude: *workloadsApi.NewNullableBool(convutils.Ref(exclude)),
		}
	}

	Context("git-sync", func() {
		It("fails with ServiceUnavailable when an item is injected but no image is reported", func() {
			fc := newComposer(emptyReportedConfig())
			err := fc.ComposeGit("git", []workloadsApi.GitInstance{gitInstance("repo1", false)}, nil, noopGitSetter, nil)()

			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusServiceUnavailable)))
		})

		It("does not fail when every item is excluded", func() {
			fc := newComposer(emptyReportedConfig())
			err := fc.ComposeGit("git", []workloadsApi.GitInstance{gitInstance("repo1", true)}, nil, noopGitSetter, nil)()

			Expect(err).To(BeNil())
		})

		It("does not fail when the default config supplies the image", func() {
			fc := newComposer(DefaultClusterStorageConfig())
			err := fc.ComposeGit("git", []workloadsApi.GitInstance{gitInstance("repo1", false)}, nil, noopGitSetter, nil)()

			Expect(err).To(BeNil())
		})
	})

	Context("S3", func() {
		It("fails with ServiceUnavailable when an item is injected but no image is reported", func() {
			fc := newComposer(emptyReportedConfig())
			err := fc.ComposeS3("s3", []workloadsApi.S3Instance{s3Instance("bucket1", false)}, UidGid{}, nil, noopS3Setter, nil)()

			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusServiceUnavailable)))
		})

		It("does not fail when every item is excluded", func() {
			fc := newComposer(emptyReportedConfig())
			err := fc.ComposeS3("s3", []workloadsApi.S3Instance{s3Instance("bucket1", true)}, UidGid{}, nil, noopS3Setter, nil)()

			Expect(err).To(BeNil())
		})

		It("does not fail when the default config supplies the image", func() {
			fc := newComposer(DefaultClusterStorageConfig())
			err := fc.ComposeS3("s3", []workloadsApi.S3Instance{s3Instance("bucket1", false)}, UidGid{}, nil, noopS3Setter, nil)()

			Expect(err).To(BeNil())
		})
	})
})
