/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package init_containers

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clustersApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/run-ai/runai/runai-cluster/cluster/workload-common/feature-gate"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func oauth2ParityConfig(domain string, subdomain bool) composer.ClusterStorageConfig {
	config := clustersApi.ClusterReportedConfig{}
	config.SetRunaiOauth2ProxyImage(oauth2ParityImage)
	config.SetSubdomainEnabled(subdomain)
	cluster := composer.ClusterWithConfig(config)
	cluster.SetDomain(domain)
	return composer.NewClusterStorageConfigFromCluster(cluster)
}

func oauth2DefaultConfigWithImage() composer.ClusterStorageConfig {
	cfg := composer.DefaultClusterStorageConfig().(composer.StaticClusterStorageConfig)
	cfg.SetOauth2ProxyImage(oauth2ParityImage)
	return cfg
}

var oauth2ParityEnvVars = []string{
	"RUNAI_OAUTH2_PROXY_IMAGE",
	"RUNAI_OAUTH2_PROXY_WORKLOAD_AUTH_SECRET_NAME",
	"RUNAI_OAUTH2_PROXY_SESSION_STORE_TYPE",
	"RUNAI_OAUTH2_PROXY_REDIS_FQDN",
	"RUNAI_OAUTH2_PROXY_USERNAME_LOADER_IMAGE",
	"OAUTH2_PROXY_OIDC_ISSUER_URL",
	"ENDPOINT",
	"IS_OPENSHIFT",
	"SUBDOMAIN_SUPPORT",
	"CUSTOM_CA_ENABLED",
	"EXTERNAL_AUTH_URL_ENABLED",
}

const (
	oauth2ParityProject       = "team-a"
	oauth2ParityWorkload      = "my-workload"
	oauth2ParityClusterDomain = "https://test.run.ai"
	oauth2ParityImage         = "test.registry/oauth2-proxy:parity"
)

func oauth2InfoFromExposedUrls(exposedUrls map[string]flds.ExposedUrlItemAndRules, subdomain, isJupyterOrTensorboard bool) *composer.ExposedUrlInfo {
	info := &composer.ExposedUrlInfo{
		Config:                 oauth2ParityConfig(oauth2ParityClusterDomain, subdomain),
		WorkloadName:           oauth2ParityWorkload,
		ProjectName:            oauth2ParityProject,
		IsJupyterOrTensorboard: isJupyterOrTensorboard,
		Items:                  make(map[string]*composer.ExposedUrlInfoItem, len(exposedUrls)),
	}
	for idx, nick := range slices.Sorted(maps.Keys(exposedUrls)) {
		item := exposedUrls[nick]
		if item.Value == nil {
			info.Items[nick] = nil
			continue
		}
		info.Items[nick] = &composer.ExposedUrlInfoItem{
			Nick:             nick,
			ContainerPort:    item.Value.ContainerPort,
			Url:              convutils.Deref(item.Value.Url, ""),
			AuthorizedUsers:  item.Value.AuthorizedUsers,
			AuthorizedGroups: item.Value.AuthorizedGroups,
			OAuthPort:        convutils.Ref(composer.Oauth2ContainerDefaultPort + int32(idx)),
		}
	}
	return info
}

func clusterOauth2Containers(ctx context.Context, exposedUrls map[string]flds.ExposedUrlItemAndRules, isJupyterOrTensorboard bool) []corev1.Container {
	GinkgoHelper()

	result := make([]corev1.Container, 0)
	noSuffixUrlGenerated := isJupyterOrTensorboard
	for idx, nick := range slices.Sorted(maps.Keys(exposedUrls)) {
		exposedUrl := exposedUrls[nick]
		if exposedUrl.Value == nil {
			continue
		}

		jobUrl, err := builders.ComposeOauthIngressUrl(ctx, exposedUrl, nick, &noSuffixUrlGenerated, oauth2ParityWorkload)
		Expect(err).NotTo(HaveOccurred())

		cookiePrefix := fmt.Sprintf("%s-%s-auth-proxy", oauth2ParityWorkload, nick)
		container, err := builders.GetOauth2ProxyContainer(jobUrl, *exposedUrl.Value.ContainerPort,
			exposedUrl.Value.AuthorizedUsers, exposedUrl.Value.AuthorizedGroups, idx, nick, false, false, cookiePrefix)
		Expect(err).NotTo(HaveOccurred())
		result = append(result, *container)
	}
	return result
}

var _ = Describe("BuildOAuth2ProxyContainers", func() {
	BeforeEach(func() {
		for _, name := range oauth2ParityEnvVars {
			GinkgoT().Setenv(name, "")
		}
	})

	newInfo := func(item *composer.ExposedUrlInfoItem) *composer.ExposedUrlInfo {
		return &composer.ExposedUrlInfo{
			Config:       oauth2DefaultConfigWithImage(),
			WorkloadName: "my-workload",
			Items:        map[string]*composer.ExposedUrlInfoItem{item.Nick: item},
		}
	}

	It("should fail with ServiceUnavailable when a proxy is needed but no image is reported", func() {
		info := &composer.ExposedUrlInfo{
			Config:       composer.NewClusterStorageConfigFromCluster(composer.ClusterWithConfig(clustersApi.ClusterReportedConfig{})),
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items:        map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://app.example.com/app"}},
		}

		_, err := BuildOAuth2ProxyContainers(info)
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusServiceUnavailable)))
	})

	It("does not fail on a missing image when no proxy is injected", func() {
		info := &composer.ExposedUrlInfo{
			Config:       composer.NewClusterStorageConfigFromCluster(composer.ClusterWithConfig(clustersApi.ClusterReportedConfig{})),
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items:        map[string]*composer.ExposedUrlInfoItem{"url-0": nil},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(BeEmpty())
	})

	It("should fail with ServiceUnavailable when reported resources are invalid", func() {
		reported := clustersApi.ClusterReportedConfig{}
		reported.SetRunaiOauth2ProxyImage(oauth2ParityImage)
		resources := clustersApi.ClusterReportedResourceRequirements{}
		resources.SetLimits(map[string]string{"memory": "not-a-quantity"})
		reported.SetRunaiOauth2ProxyResources(resources)
		info := &composer.ExposedUrlInfo{
			Config:       composer.NewClusterStorageConfigFromCluster(composer.ClusterWithConfig(reported)),
			WorkloadName: "my-workload",
			Items:        map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://app.example.com/app"}},
		}

		_, err := BuildOAuth2ProxyContainers(info)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(http.StatusServiceUnavailable)))
		Expect(err.Message).To(ContainSubstring("runaiOauth2ProxyResources"))
		Expect(err.Message).To(ContainSubstring(`limits.memory has invalid quantity "not-a-quantity"`))
	})

	It("should return empty when external auth is disabled", func() {
		info := newInfo(&composer.ExposedUrlInfoItem{Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://app.example.com/app", OAuthPort: convutils.Ref(int32(4180))})
		config := composer.StaticClusterStorageConfig{}
		config.SetExternalAuthUrlEnabled(false)
		info.Config = config

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(BeEmpty())
	})

	It("should emit one sidecar with port, cookie path env, auth secret and restart policy", func() {
		info := newInfo(&composer.ExposedUrlInfoItem{Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://app.example.com/app", OAuthPort: convutils.Ref(int32(4180))})

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))

		c := containers[0]
		Expect(c.Name).To(Equal("runai-oauth2-proxy-0"))
		Expect(c.ImagePullPolicy).To(Equal(corev1.PullIfNotPresent))
		Expect(c.Ports).To(HaveLen(1))
		Expect(c.Ports[0].ContainerPort).To(Equal(int32(4180)))
		Expect(c.Ports[0].Name).To(Equal("port-4180"))
		Expect(c.Env[0].Name).To(Equal("OAUTH2_PROXY_COOKIE_PATH"))
		Expect(c.Env[0].Value).To(Equal("/app"))
		Expect(c.RestartPolicy).NotTo(BeNil())
		Expect(*c.RestartPolicy).To(Equal(corev1.ContainerRestartPolicyAlways))
		Expect(*c.SecurityContext.RunAsNonRoot).To(BeTrue())
		Expect(c.Resources).To(Equal(corev1.ResourceRequirements{}))
		Expect(c.Args).To(ContainElements(
			"--provider=keycloak-oidc",
			"--upstream=http://localhost:8080",
			"--redirect-url=https://app.example.com/app/oauth2/callback",
			"--proxy-prefix=/app/oauth2",
			"--whitelist-domain=app.example.com",
			"--cookie-domain=app.example.com",
			"--cookie-name=my-workload-url-0-auth-proxy",
		))
	})

	It("should apply configured resources to every oauth2-proxy container", func() {
		config := oauth2DefaultConfigWithImage().(composer.StaticClusterStorageConfig)
		expected := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
		}
		config.SetOauth2ProxyResources(expected)
		info := &composer.ExposedUrlInfo{
			Config:       config,
			WorkloadName: "my-workload",
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://one.example.com", OAuthPort: convutils.Ref(int32(4180))},
				"url-1": {Nick: "url-1", ContainerPort: convutils.Ref(int32(8081)), Url: "https://two.example.com", OAuthPort: convutils.Ref(int32(4181))},
			},
		}

		containers, err := BuildOAuth2ProxyContainers(info)

		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(2))
		Expect(containers[0].Resources).To(Equal(expected))
		Expect(containers[1].Resources).To(Equal(expected))
	})

	It("should use empty resources when unset or explicitly empty", func() {
		item := &composer.ExposedUrlInfoItem{
			Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)),
			Url: "https://app.example.com", OAuthPort: convutils.Ref(int32(4180)),
		}
		unset, unsetErr := BuildOAuth2ProxyContainers(&composer.ExposedUrlInfo{
			Config: oauth2DefaultConfigWithImage(),
			Items:  map[string]*composer.ExposedUrlInfoItem{"url-0": item},
		})
		explicitConfig := oauth2DefaultConfigWithImage().(composer.StaticClusterStorageConfig)
		explicitConfig.SetOauth2ProxyResources(corev1.ResourceRequirements{})
		explicit, explicitErr := BuildOAuth2ProxyContainers(&composer.ExposedUrlInfo{
			Config: explicitConfig,
			Items:  map[string]*composer.ExposedUrlInfoItem{"url-0": item},
		})

		Expect(unsetErr).To(BeNil())
		Expect(explicitErr).To(BeNil())
		Expect(unset[0].Resources).To(Equal(corev1.ResourceRequirements{}))
		Expect(explicit[0].Resources).To(Equal(corev1.ResourceRequirements{}))
	})

	It("should add an authenticated-emails-file and no email-domain when users are set", func() {
		info := newInfo(&composer.ExposedUrlInfoItem{
			Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)),
			Url: "https://app.example.com/app", AuthorizedUsers: convutils.Ref("alice@example.com"),
		})

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers[0].Args).To(ContainElement("--authenticated-emails-file=/etc/runai/auth/users/dXJsLTA="))
		Expect(containers[0].Args).NotTo(ContainElement("--email-domain=*"))
	})

	It("should add one allowed-group per group plus email-domain when groups are set", func() {
		info := newInfo(&composer.ExposedUrlInfoItem{
			Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)),
			Url: "https://app.example.com/app", AuthorizedGroups: convutils.Ref("group-a,group-b"),
		})

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers[0].Args).To(ContainElements("--allowed-group=group-a", "--allowed-group=group-b", "--email-domain=*"))
	})

	It("should add email-domain when neither users nor groups are set", func() {
		info := newInfo(&composer.ExposedUrlInfoItem{Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://app.example.com/app"})

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers[0].Args).To(ContainElement("--email-domain=*"))
	})

	It("should reject an invalid group name", func() {
		info := newInfo(&composer.ExposedUrlInfoItem{
			Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)),
			Url: "https://app.example.com/app", AuthorizedGroups: convutils.Ref("bad group!"),
		})

		_, err := BuildOAuth2ProxyContainers(info)
		Expect(err).NotTo(BeNil())
	})

	It("should derive a path-based redirect url when the item has no url", func() {
		info := &composer.ExposedUrlInfo{
			Config:       oauth2ParityConfig("https://test.run.ai", false),
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items:        map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), OAuthPort: convutils.Ref(int32(4180))}},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))
		Expect(containers[0].Env[0].Value).To(Equal("/team-a/my-workload"))
		Expect(containers[0].Args).To(ContainElements(
			"--redirect-url=https://test.run.ai/team-a/my-workload/oauth2/callback",
			"--proxy-prefix=/team-a/my-workload/oauth2",
			"--whitelist-domain=test.run.ai",
		))
	})

	It("should derive a subdomain redirect url when the item has no url and subdomain routing is on", func() {
		info := &composer.ExposedUrlInfo{
			Config:       oauth2ParityConfig("https://test.run.ai", true),
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items:        map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), OAuthPort: convutils.Ref(int32(4180))}},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))
		Expect(containers[0].Env[0].Value).To(Equal("/"))
		Expect(containers[0].Args).To(ContainElements(
			"--redirect-url=https://team-a-my-workload.test.run.ai/oauth2/callback",
			"--proxy-prefix=/oauth2",
			"--whitelist-domain=team-a-my-workload.test.run.ai",
		))
	})

	It("should strip the runai. prefix from redirect and whitelist on openshift with subdomain routing", func() {
		config := composer.StaticClusterStorageConfig{}
		config.SetOauth2ProxyImage(oauth2ParityImage)
		config.SetExternalAuthUrlEnabled(true)
		config.SetIsOpenshift(true)
		config.SetSubdomainSupportEnabled(true)

		info := &composer.ExposedUrlInfo{
			Config:       config,
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items:        map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", Url: "https://runai.apps.example.com/team"}},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))
		Expect(containers[0].Args).To(ContainElements(
			"--redirect-url=https://apps.example.com/team/oauth2/callback",
			"--whitelist-domain=apps.example.com",
			"--cookie-domain=apps.example.com",
		))
	})

	It("should keep the runai. prefix when the cluster is not openshift", func() {
		config := composer.StaticClusterStorageConfig{}
		config.SetOauth2ProxyImage(oauth2ParityImage)
		config.SetExternalAuthUrlEnabled(true)
		config.SetIsOpenshift(false)
		config.SetSubdomainSupportEnabled(true)

		info := &composer.ExposedUrlInfo{
			Config:       config,
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items:        map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", Url: "https://runai.apps.example.com/team"}},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))
		Expect(containers[0].Args).To(ContainElement("--whitelist-domain=runai.apps.example.com"))
	})

	It("should pass the configured oidc issuer into the proxy args", func() {
		config := composer.StaticClusterStorageConfig{}
		config.SetOauth2ProxyImage(oauth2ParityImage)
		config.SetExternalAuthUrlEnabled(true)
		config.SetOauth2ProxyOidcIssuerUrl("https://cp.example.com/auth/realms/acme")

		info := &composer.ExposedUrlInfo{
			Config:       config,
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items:        map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", Url: "https://a.example.com/x"}},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(1))
		Expect(containers[0].Args).To(ContainElements(
			"--oidc-issuer-url=https://cp.example.com/auth/realms/acme",
			"--extra-jwt-issuers=https://cp.example.com/auth/realms/acme=runai",
		))
	})

	It("should suffix only the second of two auto-generated urls", func() {
		info := &composer.ExposedUrlInfo{
			Config:       oauth2ParityConfig("https://test.run.ai", false),
			WorkloadName: "my-workload",
			ProjectName:  "team-a",
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), OAuthPort: convutils.Ref(int32(4180))},
				"url-1": {Nick: "url-1", ContainerPort: convutils.Ref(int32(9090)), OAuthPort: convutils.Ref(int32(4181))},
			},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(2))
		Expect(containers[0].Args).To(ContainElement("--proxy-prefix=/team-a/my-workload/oauth2"))
		Expect(containers[1].Args).To(ContainElement("--proxy-prefix=/team-a/my-workload/url-1/oauth2"))
	})

	It("should not suffix the first auto-generated url even for jupyter or tensorboard", func() {
		// We mirror the notification's connections-pre-processor logic, which
		// always lets the first auto URL be suffix-less, regardless of whether
		// the workload is jupyter/tensorboard. This keeps the OAuth2 proxy-prefix
		// aligned with the Ingress path that the user actually hits.
		info := &composer.ExposedUrlInfo{
			Config:                 oauth2ParityConfig("https://test.run.ai", false),
			WorkloadName:           "my-workload",
			ProjectName:            "team-a",
			IsJupyterOrTensorboard: true,
			Items:                  map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), OAuthPort: convutils.Ref(int32(4180))}},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers[0].Args).To(ContainElement("--proxy-prefix=/team-a/my-workload/oauth2"))
	})

	It("should index ports and names per item in sorted order", func() {
		info := &composer.ExposedUrlInfo{
			Config:       oauth2DefaultConfigWithImage(),
			WorkloadName: "my-workload",
			Items: map[string]*composer.ExposedUrlInfoItem{
				"url-0": {Nick: "url-0", ContainerPort: convutils.Ref(int32(8080)), Url: "https://a.example.com/x", OAuthPort: convutils.Ref(int32(4180))},
				"url-1": {Nick: "url-1", ContainerPort: convutils.Ref(int32(9090)), Url: "https://b.example.com/y", OAuthPort: convutils.Ref(int32(4181))},
			},
		}

		containers, err := BuildOAuth2ProxyContainers(info)
		Expect(err).To(BeNil())
		Expect(containers).To(HaveLen(2))
		Expect(containers[0].Ports[0].ContainerPort).To(Equal(int32(4180)))
		Expect(containers[1].Ports[0].ContainerPort).To(Equal(int32(4181)))
	})
})

var _ = Describe("BuildOAuth2ProxyVolumes", func() {
	BeforeEach(func() {
		for _, name := range oauth2ParityEnvVars {
			GinkgoT().Setenv(name, "")
		}
	})

	It("should return empty when there are no url entries", func() {
		Expect(BuildOAuth2ProxyVolumes(&composer.ExposedUrlInfo{Config: composer.DefaultClusterStorageConfig()})).To(BeEmpty())
	})

	It("should emit a single auth emptyDir volume", func() {
		info := &composer.ExposedUrlInfo{
			Config: composer.DefaultClusterStorageConfig(),
			Items:  map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", Url: "https://a.example.com/x", OAuthPort: convutils.Ref(int32(4180))}},
		}

		volumes := BuildOAuth2ProxyVolumes(info)
		Expect(volumes).To(HaveLen(1))
		Expect(volumes[0].Name).To(Equal("runai-auth"))
		Expect(volumes[0].EmptyDir).NotTo(BeNil())
	})

	It("should append the custom-CA secret volume when auth custom CA is enabled", func() {
		config := composer.StaticClusterStorageConfig{}
		config.SetExternalAuthUrlEnabled(true)
		config.SetIsCustomCAAuthEnabled(true)
		config.SetCustomCAAuthSecret("runai-ca-cert")
		config.SetCustomCAAuthKey("runai-ca.pem")

		info := &composer.ExposedUrlInfo{
			Config: config,
			Items:  map[string]*composer.ExposedUrlInfoItem{"url-0": {Nick: "url-0", Url: "https://a.example.com/x", OAuthPort: convutils.Ref(int32(4180))}},
		}

		volumes := BuildOAuth2ProxyVolumes(info)
		Expect(volumes).To(HaveLen(2))
		Expect(volumes[0].Name).To(Equal("runai-auth"))
		Expect(volumes[1].Name).To(Equal("runai-ca-cert"))
		Expect(volumes[1].Secret).NotTo(BeNil())
		Expect(volumes[1].Secret.Items[0].Key).To(Equal("runai-ca.pem"))
	})
})

var _ = Describe("BuildOAuth2ProxyContainers parity with the cluster builder", func() {
	var (
		exposedUrls map[string]flds.ExposedUrlItemAndRules
		fgPrev      bool
	)

	BeforeEach(func() {
		for _, name := range oauth2ParityEnvVars {
			GinkgoT().Setenv(name, "")
		}
		GinkgoT().Setenv(builders.ExternalUrlAuthContainerImageEnv, oauth2ParityImage)
		fg := featuregate.GetFeatureGateUtilInstance()
		fgPrev = fg.FeatureGateSidecarContainersEnable
		fg.FeatureGateSidecarContainersEnable = true
		GinkgoT().Setenv(config.EnvVarClusterDomain, oauth2ParityClusterDomain)

		exposedUrls = map[string]flds.ExposedUrlItemAndRules{}
	})

	AfterEach(func() {
		featuregate.GetFeatureGateUtilInstance().FeatureGateSidecarContainersEnable = fgPrev
	})

	addCustomUrl := func(nick string, port int32, url string, users, groups *string) {
		exposedUrls[nick] = flds.ExposedUrlItemAndRules{
			Value: &flds.ExposedUrlItemValue{
				ContainerPort:    convutils.Ref(port),
				CustomUrl:        convutils.Ref(true),
				Url:              convutils.Ref(url),
				AuthorizedUsers:  users,
				AuthorizedGroups: groups,
			},
		}
	}

	addAutoUrl := func(nick string, port int32) {
		exposedUrls[nick] = flds.ExposedUrlItemAndRules{
			Value: &flds.ExposedUrlItemValue{ContainerPort: convutils.Ref(port)},
		}
	}

	expectParity := func(subdomain, isJupyterOrTensorboard bool) {
		GinkgoHelper()
		ctx := project.SetProjectInContext(context.Background(), oauth2ParityProject)
		if subdomain {
			GinkgoT().Setenv(config.EnvVarSubdomainSupport, "true")
		}
		clusterOut := clusterOauth2Containers(ctx, exposedUrls, isJupyterOrTensorboard)
		cpOut, err := BuildOAuth2ProxyContainers(oauth2InfoFromExposedUrls(exposedUrls, subdomain, isJupyterOrTensorboard))
		Expect(err).To(BeNil())
		Expect(cpOut).To(Equal(clusterOut))
	}

	It("mirrors the cluster for a url with no specific authorization", func() {
		addCustomUrl("url-0", 8080, "https://app.example.com/app", nil, nil)
		expectParity(false, false)
	})

	It("mirrors the cluster for a url authorized by users", func() {
		addCustomUrl("url-0", 8080, "https://app.example.com/app", convutils.Ref("alice@example.com,bob@example.com"), nil)
		expectParity(false, false)
	})

	It("mirrors the cluster for a url authorized by groups", func() {
		addCustomUrl("url-0", 8080, "https://app.example.com/app", nil, convutils.Ref("group-a,group-b"))
		expectParity(false, false)
	})

	It("mirrors the cluster for a root-path url", func() {
		addCustomUrl("url-0", 8080, "https://app.example.com/", nil, nil)
		expectParity(false, false)
	})

	It("mirrors the cluster across multiple urls in sorted order", func() {
		addCustomUrl("url-0", 8080, "https://a.example.com/x", nil, nil)
		addCustomUrl("url-1", 9090, "https://b.example.com/y", convutils.Ref("alice@example.com"), nil)
		expectParity(false, false)
	})

	It("mirrors the cluster when a nil item sorts before a real one", func() {
		exposedUrls["url-0"] = flds.ExposedUrlItemAndRules{Value: nil}
		addCustomUrl("url-1", 8080, "https://app.example.com/app", nil, nil)
		expectParity(false, false)
	})

	It("mirrors the cluster for an auto-generated path-based url", func() {
		addAutoUrl("url-0", 8080)
		expectParity(false, false)
	})

	It("mirrors the cluster for an auto-generated subdomain url", func() {
		addAutoUrl("url-0", 8080)
		expectParity(true, false)
	})

	It("mirrors the cluster for multiple auto-generated urls where only the first is suffixless", func() {
		addAutoUrl("url-0", 8080)
		addAutoUrl("url-1", 9090)
		expectParity(false, false)
	})

	// V2 intentionally diverges from the V1 cluster builder for jupyter/tensorboard:
	// the cluster always suffixes every auto URL when isJupyter/isTensorboard is true,
	// but V2 keeps the first auto URL suffix-less to stay aligned with the
	// notification's connections-pre-processor and the Ingress path generated by
	// UrlsToWorkloadRunner. This is intentional and is covered elsewhere by tests
	// that assert the V2 (no-suffix) behavior directly.

	It("mirrors the cluster for a mix of custom and auto-generated urls", func() {
		addAutoUrl("url-0", 8080)
		addCustomUrl("url-1", 9090, "https://app.example.com/app", nil, nil)
		expectParity(false, false)
	})
})
