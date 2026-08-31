/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/run-ai/runai/runai-cluster/cluster/workload-common/feature-gate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"knative.dev/pkg/apis"
)

var _ = Describe("OAuth2Proxy tests", func() {
	var (
		ctx                    context.Context
		jobUrl                 *apis.URL
		containerPort          int32
		authorizedUsers        *string
		authorizedGroups       *string
		idx                    int
		nick                   string
		isInference            bool
		isInferenceServingPort bool
		cookiePrefix           string
		exposedUrl             flds.ExposedUrlItemAndRules
		noSuffixUrlGenerated   *bool
		name                   string
		containers             []corev1.Container
		initContainers         []corev1.Container
	)

	BeforeEach(func() {
		// Setup common test variables
		ctx = context.Background()
		var err error
		jobUrl, err = apis.ParseURL("https://example.com/path")
		Expect(err).NotTo(HaveOccurred())
		containerPort = 8080
		authorizedUsers = nil
		authorizedGroups = nil
		idx = 0
		nick = "test-workload"
		isInference = false
		isInferenceServingPort = false
		cookiePrefix = "oauth2-proxy"

		exposedUrl = flds.ExposedUrlItemAndRules{
			Value: &flds.ExposedUrlItemValue{
				CustomUrl: ptr.To(false),
				Url:       ptr.To(""),
			},
		}
		noSuffixUrlGenerated = ptr.To(false)
		name = "test-workload"

		containers = []corev1.Container{}
		initContainers = []corev1.Container{}

		// Set environment variables needed for tests
		os.Setenv(ExternalUrlAuthContainerImageEnv, "oauth2-proxy:latest")
		os.Setenv(WorkloadAuthSecretNameEnv, "oauth2-proxy-secret")
		os.Setenv(WorkloadAuthSessionStoreTypeEnv, "redis")
		os.Setenv(ExternalFQDNClusterProxyRedisEnv, "redis.example.com:6379")
	})

	AfterEach(func() {
		// Clean up environment variables
		os.Unsetenv(ExternalUrlAuthContainerImageEnv)
		os.Unsetenv(WorkloadAuthSecretNameEnv)
		os.Unsetenv(WorkloadAuthSessionStoreTypeEnv)
		os.Unsetenv(ExternalFQDNClusterProxyRedisEnv)
		os.Unsetenv(workloadconfig.EnvVarOauth2ProxyResources)
	})

	Context("GetOauth2ProxyContainer", func() {
		It("should create a basic OAuth2 proxy container", func() {
			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container).NotTo(BeNil())
			Expect(container.Name).To(Equal(ExternalUrlAuthContainerName + "-0"))
			Expect(container.Image).To(Equal("oauth2-proxy:latest"))
			Expect(container.Resources).To(Equal(corev1.ResourceRequirements{}))
			Expect(container.Args).To(ContainElement("--provider=keycloak-oidc"))
			Expect(container.Args).To(ContainElement("--upstream=http://localhost:8080"))
			Expect(container.Args).To(ContainElement("--cookie-name=oauth2-proxy"))
			Expect(container.Args).To(ContainElement("--cookie-domain=example.com"))
			Expect(container.Args).To(ContainElement("--pass-authorization-header=false"))

			// Verify volume mounts
			Expect(container.VolumeMounts).To(HaveLen(1))
			Expect(container.VolumeMounts[0].Name).To(Equal(constants.ExternalUrlAuthVolumeName))

			// Verify ports
			Expect(container.Ports).To(HaveLen(1))
			Expect(container.Ports[0].ContainerPort).To(Equal(constants.ExternalUrlAuthContainerDefaultPort))

			// Verify security context
			Expect(*container.SecurityContext.AllowPrivilegeEscalation).To(BeFalse())
			Expect(*container.SecurityContext.RunAsNonRoot).To(BeTrue())

			// Verify Redis session store arguments
			Expect(container.Args).To(ContainElement("--session-store-type=redis"))
			Expect(container.Args).To(ContainElement(fmt.Sprintf("--redis-connection-url=redis://%s", os.Getenv(ExternalFQDNClusterProxyRedisEnv))))
		})

		It("should apply configured resource requirements", func() {
			os.Setenv(workloadconfig.EnvVarOauth2ProxyResources, `{"requests":{"cpu":"100m"},"limits":{"memory":"128Mi"}}`)

			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Resources).To(Equal(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")},
				Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("128Mi")},
			}))
		})

		It("should reject malformed resource requirements", func() {
			os.Setenv(workloadconfig.EnvVarOauth2ProxyResources, `{"requests":{"cpu":"not-a-quantity"}}`)

			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(container).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(workloadconfig.EnvVarOauth2ProxyResources)))
		})

		It("should preserve explicitly empty resource requirements", func() {
			os.Setenv(workloadconfig.EnvVarOauth2ProxyResources, `{}`)

			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Resources).To(Equal(corev1.ResourceRequirements{}))
		})

		It("should handle authorized users correctly", func() {
			users := "user1@example.com,user2@example.com"
			authorizedUsers = &users

			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Args).To(ContainElement("--authenticated-emails-file=/etc/runai/auth/users/" + GetSafeFileNameFromNick(nick)))
		})

		It("should handle authorized groups correctly", func() {
			groups := "group-a,group-b"
			authorizedGroups = &groups

			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Args).To(ContainElement("--allowed-group=group-a"))
			Expect(container.Args).To(ContainElement("--allowed-group=group-b"))
			Expect(container.Args).To(ContainElement("--email-domain=*"))
		})

		It("should fail with invalid authorized groups", func() {
			groups := "invalid**group,group-b"
			authorizedGroups = &groups

			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid group: invalid**group"))
			Expect(container).To(BeNil())
		})

		It("should handle inference mode correctly", func() {
			isInference = true
			isInferenceServingPort = true

			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Ports).To(HaveLen(1))
			// In inference mode with serving port, container should not have protocol or name set
			Expect(container.Ports[0].Protocol).To(Equal(corev1.Protocol("")))
			Expect(container.Ports[0].Name).To(Equal(""))
			Expect(container.Args).To(ContainElement("--force-json-errors=true"))
		})

		It("should handle feature gate sidecar containers", func() {
			// Set the feature gate to true
			featureGateUtil := featuregate.GetFeatureGateUtilInstance()
			originalValue := featureGateUtil.FeatureGateSidecarContainersEnable
			featureGateUtil.FeatureGateSidecarContainersEnable = true

			container, err := GetOauth2ProxyContainer(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.RestartPolicy).NotTo(BeNil())
			Expect(*container.RestartPolicy).To(Equal(corev1.ContainerRestartPolicyAlways))

			// Reset the feature gate
			featureGateUtil.FeatureGateSidecarContainersEnable = originalValue
		})
	})

	Context("SanitizeAuthorizedGroups", func() {
		It("should validate and sanitize authorized groups correctly", func() {
			// Valid groups
			result, err := SanitizeAuthorizedGroups("group-a,group-b,group-c")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("group-a,group-b,group-c"))

			// Valid groups with spaces
			result, err = SanitizeAuthorizedGroups(" group-a , group-b ,group-c ")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("group-a,group-b,group-c"))

			// Valid groups with various characters
			result, err = SanitizeAuthorizedGroups("group-a,group_b,group.c")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("group-a,group_b,group.c"))
		})

		It("should reject invalid group names", func() {
			// Invalid group with special characters
			_, err := SanitizeAuthorizedGroups("group-a,invalid**group,group-c")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid group: invalid**group"))

			// Invalid group starting with non-alphanumeric
			_, err = SanitizeAuthorizedGroups("group-a,-invalid-group,group-c")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid group: -invalid-group"))

			// Invalid group ending with non-alphanumeric
			_, err = SanitizeAuthorizedGroups("group-a,invalid-group-,group-c")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid group: invalid-group-"))

			// Group name too long (>64 chars)
			longGroupName := strings.Repeat("a", 65)
			_, err = SanitizeAuthorizedGroups("group-a," + longGroupName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid group: " + longGroupName))
		})

		It("should handle empty input", func() {
			// Empty string
			_, err := SanitizeAuthorizedGroups("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("authorizedGroups: no valid groups found"))

			// Only commas
			_, err = SanitizeAuthorizedGroups(",,,")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("authorizedGroups: no valid groups found"))

			// Only spaces
			_, err = SanitizeAuthorizedGroups("   ")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("authorizedGroups: no valid groups found"))
		})
	})

	Context("ComposeOauthIngressUrl", func() {
		It("should validate customUrl flag with url value consistency", func() {
			// Test case: customUrl enabled but no url provided
			exposedUrl.Value.CustomUrl = ptr.To(true)
			exposedUrl.Value.Url = ptr.To("")

			url, err := ComposeOauthIngressUrl(ctx, exposedUrl, nick, noSuffixUrlGenerated, name)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("url is mandatory when customUrl is enabled"))
			Expect(url).To(BeNil())

			// Test case: customUrl not enabled but url provided
			exposedUrl.Value.CustomUrl = ptr.To(false)
			exposedUrl.Value.Url = ptr.To("https://example.com")

			url, err = ComposeOauthIngressUrl(ctx, exposedUrl, nick, noSuffixUrlGenerated, name)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("url should only be provided for customUrl"))
			Expect(url).To(BeNil())
		})

		It("should use custom URL when custom URL is enabled", func() {
			// Set up a valid custom URL configuration
			customUrl := "https://custom.example.com/path"
			exposedUrl.Value.CustomUrl = ptr.To(true)
			exposedUrl.Value.Url = ptr.To(customUrl)

			url, err := ComposeOauthIngressUrl(ctx, exposedUrl, nick, noSuffixUrlGenerated, name)

			Expect(err).NotTo(HaveOccurred())
			Expect(url).NotTo(BeNil())
			Expect(url.String()).To(Equal(customUrl))
		})

		// This test will need to be mocked or adapted since ComposeExposedUrl is not directly accessible
		It("should call ComposeExposedUrl when custom URL is not enabled", func() {
			// This test is pending as it requires mocking/stubbing ComposeExposedUrl which is not exported
		})
	})

	Context("IsWorkloadSupportsAuth", func() {
		It("should detect OAuth container in regular containers", func() {
			// Create a workload with an OAuth container
			containers = []corev1.Container{
				{Name: "container1"},
				{Name: ExternalUrlAuthContainerName + "-0"},
				{Name: "container2"},
			}
			initContainers = []corev1.Container{}

			Expect(IsWorkloadSupportsAuth(containers, initContainers)).To(BeTrue())
		})

		It("should detect OAuth container in init containers", func() {
			// Create a workload with an OAuth init container
			containers = []corev1.Container{}
			initContainers = []corev1.Container{
				{Name: "init-container1"},
				{Name: ExternalUrlAuthContainerName + "-0"},
				{Name: "init-container2"},
			}

			Expect(IsWorkloadSupportsAuth(containers, initContainers)).To(BeTrue())
		})

		It("should return false when no OAuth container exists", func() {
			// Create a workload without OAuth containers
			containers = []corev1.Container{
				{Name: "container1"},
				{Name: "container2"},
			}
			initContainers = []corev1.Container{
				{Name: "init-container1"},
				{Name: "init-container2"},
			}

			Expect(IsWorkloadSupportsAuth(containers, initContainers)).To(BeFalse())
		})

		It("should handle empty container lists", func() {
			// Empty container lists
			containers = []corev1.Container{}
			initContainers = []corev1.Container{}

			Expect(IsWorkloadSupportsAuth(containers, initContainers)).To(BeFalse())
		})
	})
})
