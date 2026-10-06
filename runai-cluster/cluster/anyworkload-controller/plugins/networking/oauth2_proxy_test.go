/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/feature-gate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"knative.dev/pkg/apis"
)

var _ = Describe("OAuth2Proxy tests", func() {
	var (
		jobUrl                 *apis.URL
		containerPort          int32
		authorizedUsers        *string
		authorizedGroups       *string
		idx                    int
		nick                   string
		isInference            bool
		isInferenceServingPort bool
		cookiePrefix           string
	)

	BeforeEach(func() {
		// Setup common test variables
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

		// Restore the networking suite defaults after each moved auth test.
		GinkgoT().Setenv(workloadconfig.EnvVarOauth2ProxyResources, "")
		GinkgoT().Setenv(ExternalUrlAuthContainerImageEnv, "oauth2-proxy:latest")
		GinkgoT().Setenv(WorkloadAuthSecretNameEnv, "oauth2-proxy-secret")
		GinkgoT().Setenv(WorkloadAuthSessionStoreTypeEnv, "redis")
		GinkgoT().Setenv(ExternalFQDNClusterProxyRedisEnv, "redis.example.com:6379")
	})

	Context("legacyOAuth2Container", func() {
		It("should create a basic OAuth2 proxy container", func() {
			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

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
			GinkgoT().Setenv(workloadconfig.EnvVarOauth2ProxyResources, `{"requests":{"cpu":"100m"},"limits":{"memory":"128Mi"}}`)

			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Resources).To(Equal(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")},
				Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("128Mi")},
			}))
		})

		It("should reject malformed resource requirements", func() {
			GinkgoT().Setenv(workloadconfig.EnvVarOauth2ProxyResources, `{"requests":{"cpu":"not-a-quantity"}}`)

			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(container).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(workloadconfig.EnvVarOauth2ProxyResources)))
		})

		It("should preserve explicitly empty resource requirements", func() {
			GinkgoT().Setenv(workloadconfig.EnvVarOauth2ProxyResources, `{}`)

			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Resources).To(Equal(corev1.ResourceRequirements{}))
		})

		It("should handle authorized users correctly", func() {
			users := "user1@example.com,user2@example.com"
			authorizedUsers = &users

			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Args).To(ContainElement("--authenticated-emails-file=/etc/runai/auth/users/" + GetSafeFileNameFromNick(nick)))
		})

		It("should handle authorized groups correctly", func() {
			groups := "group-a,group-b"
			authorizedGroups = &groups

			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Args).To(ContainElement("--allowed-group=group-a"))
			Expect(container.Args).To(ContainElement("--allowed-group=group-b"))
			Expect(container.Args).To(ContainElement("--email-domain=*"))
		})

		It("should fail with invalid authorized groups", func() {
			groups := "invalid**group,group-b"
			authorizedGroups = &groups

			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid group: invalid**group"))
			Expect(container).To(BeNil())
		})

		It("should handle inference mode correctly", func() {
			isInference = true
			isInferenceServingPort = true

			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

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
			DeferCleanup(func() { featureGateUtil.FeatureGateSidecarContainersEnable = originalValue })
			featureGateUtil.FeatureGateSidecarContainersEnable = true

			container, err := legacyOAuth2Container(jobUrl, containerPort, authorizedUsers, authorizedGroups, idx, nick, isInference, isInferenceServingPort, cookiePrefix)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.RestartPolicy).NotTo(BeNil())
			Expect(*container.RestartPolicy).To(Equal(corev1.ContainerRestartPolicyAlways))

			// Reset the feature gate
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

})

// legacyOAuth2Container adapts the retained sidecar test cases to the plugin builder.
func legacyOAuth2Container(
	jobUrl *apis.URL, containerPort int32, authorizedUsers *string, authorizedGroups *string,
	idx int, nick string, omitPortDeclaration bool, forceJsonResponse bool, cookiePrefix string,
	upstreamTimeoutSeconds ...int,
) (*corev1.Container, error) {
	featureGateUtil := featuregate.GetFeatureGateUtilInstance()

	timeout := 0
	if len(upstreamTimeoutSeconds) > 0 {
		timeout = upstreamTimeoutSeconds[0]
	}

	cfg := OAuth2ProxyConfig{
		ContainerName:              ExternalUrlAuthContainerName + "-" + strconv.Itoa(idx),
		UpstreamURL:                "http://localhost:" + strconv.Itoa(int(containerPort)),
		ExternalScheme:             jobUrl.Scheme,
		ExternalHost:               jobUrl.Host,
		ExternalPath:               jobUrl.Path,
		Port:                       constants.ExternalUrlAuthContainerDefaultPort + int32(idx),
		CookiePrefix:               cookiePrefix,
		AuthorizedUsers:            authorizedUsers,
		AuthorizedGroups:           authorizedGroups,
		Nick:                       nick,
		OmitPortDeclaration:        omitPortDeclaration,
		ForceJsonResponse:          forceJsonResponse,
		EnableSidecarRestartPolicy: featureGateUtil.FeatureGateSidecarContainersEnable,
		UpstreamTimeoutSeconds:     timeout,
	}

	return BuildOAuth2ProxyContainer(cfg)
}
