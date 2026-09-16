/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/base64"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	assetsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AI Application personal credentials", func() {
	const (
		genericID = "12345678-abcd-4000-8000-000000000001"
		dockerID  = "87654321-abcd-4000-8000-000000000002"
		ngcID     = "abcdef12-abcd-4000-8000-000000000003"
	)

	credential := func(name, id string, kind assetsApi.AssetKind) assetsApi.CredentialsListResponseEntry {
		return assetsApi.CredentialsListResponseEntry{
			Meta: assetsApi.AssetMeta{Name: name, Id: id, Kind: kind},
		}
	}

	credentials := []assetsApi.CredentialsListResponseEntry{
		credential("generic", genericID, assetsApi.ASSETKIND_GENERIC_SECRET),
		credential("registry", dockerID, assetsApi.ASSETKIND_DOCKER_REGISTRY),
		credential("ngc", ngcID, assetsApi.ASSETKIND_NGC_API_KEY),
		credential("password", "11111111-abcd-4000-8000-000000000004", assetsApi.ASSETKIND_PASSWORD),
	}

	BeforeEach(func() {
		assetsClient.SetMock(assetsClient.NewSimpleMock().WithCredentials(credentials))
	})

	It("canonicalizes exact placeholders recursively", func() {
		spec := &api.AiApplicationSpec{
			Values: map[string]interface{}{
				"generic": "USERCREDENTIAL:generic",
				"nested": []interface{}{
					map[string]interface{}{"docker": "USERCREDENTIAL:registry"},
					"ordinary",
				},
				"ngcGeneric": "USERCREDENTIAL:ngc,generic-secret",
				"ngcDocker":  "USERCREDENTIAL:ngc,docker-registry",
			},
		}

		used, apiErr := processAIApplicationCredentials(testing.SetContext(testing.TestTenant, testing.TestUser, nil), "my-app", spec)
		Expect(apiErr).To(BeNil())
		Expect(used).To(BeTrue())
		Expect(spec.Values["generic"]).To(Equal("USERCREDENTIAL:generic-12345678-a"))
		Expect(spec.Values["ngcGeneric"]).To(Equal("USERCREDENTIAL:ngcgs-ngc-abcdef12-a"))
		Expect(spec.Values["ngcDocker"]).To(Equal("USERCREDENTIAL:ngcdr-ngc-abcdef12-a"))
		nested := spec.Values["nested"].([]interface{})
		Expect(nested[0].(map[string]interface{})["docker"]).To(Equal("USERCREDENTIAL:registry-87654321-a"))
	})

	It("decodes base64 YAML into structured canonical values", func() {
		encoded := base64.StdEncoding.EncodeToString([]byte("config:\n  secret: USERCREDENTIAL:generic\n"))
		spec := &api.AiApplicationSpec{ValuesBase64: encoded}

		used, apiErr := processAIApplicationCredentials(testing.SetContext(testing.TestTenant, testing.TestUser, nil), "my-app", spec)
		Expect(apiErr).To(BeNil())
		Expect(used).To(BeTrue())
		Expect(spec.ValuesBase64).To(BeEmpty())
		Expect(spec.Values["config"].(map[string]interface{})["secret"]).To(Equal("USERCREDENTIAL:generic-12345678-a"))
	})

	It("preserves placeholder-free base64 values used with personal chart authentication", func() {
		encoded := base64.StdEncoding.EncodeToString([]byte("replicaCount: 2\n"))
		spec := &api.AiApplicationSpec{
			ChartRef:     api.AiApplicationSpecChartRef{Repo: "https://charts.example.com"},
			ValuesBase64: encoded,
			Auth: api.AiApplicationAuth{
				UserCredential: convutils.Ref(true),
				AssetId:        genericID,
			},
		}

		used, apiErr := processAIApplicationCredentials(testing.SetContext(testing.TestTenant, testing.TestUser, nil), "my-app", spec)
		Expect(apiErr).To(BeNil())
		Expect(used).To(BeTrue())
		Expect(spec.ValuesBase64).To(Equal(encoded))
		Expect(spec.Values).To(BeNil())
	})

	DescribeTable("rejects invalid value inputs",
		func(values map[string]interface{}, valuesBase64 string, message string) {
			spec := &api.AiApplicationSpec{Values: values, ValuesBase64: valuesBase64}
			_, apiErr := processAIApplicationCredentials(testing.SetContext(testing.TestTenant, testing.TestUser, nil), "my-app", spec)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Message).To(ContainSubstring(message))
		},
		Entry("base64 placeholder outside an object", nil, base64.StdEncoding.EncodeToString([]byte("USERCREDENTIAL:generic\n")), "valid YAML object"),
		Entry("embedded placeholder", map[string]interface{}{"a": "prefix-USERCREDENTIAL:generic"}, "", "complete value"),
		Entry("unknown credential", map[string]interface{}{"a": "USERCREDENTIAL:other"}, "", "reference is invalid"),
		Entry("unsupported kind", map[string]interface{}{"a": "USERCREDENTIAL:password"}, "", "kind is not supported"),
		Entry("NGC without representation", map[string]interface{}{"a": "USERCREDENTIAL:ngc"}, "", "must specify"),
		Entry("representation on generic", map[string]interface{}{"a": "USERCREDENTIAL:generic,generic-secret"}, "", "only for NGC"),
	)

	It("preserves invalid base64 when it does not contain a personal credential", func() {
		spec := &api.AiApplicationSpec{ValuesBase64: "existing-invalid-value"}

		used, apiErr := processAIApplicationCredentials(testing.SetContext(testing.TestTenant, testing.TestUser, nil), "my-app", spec)
		Expect(apiErr).To(BeNil())
		Expect(used).To(BeFalse())
		Expect(spec.ValuesBase64).To(Equal("existing-invalid-value"))
		Expect(spec.Values).To(BeNil())
	})

	It("rejects an invalid computed destination name", func() {
		longCredentialName := strings.Repeat("a", 250)
		assetsClient.SetMock(assetsClient.NewSimpleMock().WithCredentials([]assetsApi.CredentialsListResponseEntry{
			credential(longCredentialName, genericID, assetsApi.ASSETKIND_GENERIC_SECRET),
		}))
		spec := &api.AiApplicationSpec{
			Values: map[string]interface{}{"secret": "USERCREDENTIAL:" + longCredentialName},
		}
		_, apiErr := processAIApplicationCredentials(
			testing.SetContext(testing.TestTenant, testing.TestUser, nil),
			"my-app",
			spec,
		)
		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Message).To(ContainSubstring("destination name is invalid"))
	})

	It("validates personal chart authentication against caller-owned credentials", func() {
		spec := &api.AiApplicationSpec{
			ChartRef: api.AiApplicationSpecChartRef{Repo: "https://charts.example.com"},
			Auth: api.AiApplicationAuth{
				UserCredential: convutils.Ref(true),
				AssetId:        genericID,
			},
		}
		used, apiErr := processAIApplicationCredentials(
			testing.SetContext(testing.TestTenant, testing.TestUser, nil),
			"my-app",
			spec,
		)
		Expect(apiErr).To(BeNil())
		Expect(used).To(BeTrue())

		spec.Auth.AssetId = "00000000-abcd-4000-8000-000000000000"
		_, apiErr = processAIApplicationCredentials(
			testing.SetContext(testing.TestTenant, testing.TestUser, nil),
			"my-app",
			spec,
		)
		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Message).To(ContainSubstring("invalid"))
	})

	It("sanitizes canonical references only for non-creators without mutating the source map", func() {
		values := map[string]interface{}{
			"secret": "USERCREDENTIAL:generic-12345678-a",
			"plain":  "ordinary",
		}
		application := api.AiApplication{
			CreatedBy: testing.TestUser,
			Spec: api.AiApplicationSpec{
				Values: values,
				Auth: api.AiApplicationAuth{
					UserCredential: convutils.Ref(true),
					AssetId:        genericID,
				},
			},
		}

		otherUserContext := testing.SetContext(testing.TestTenant, "another-user", nil)
		sanitizeAIApplicationResponse(otherUserContext, &application)
		Expect(application.Spec.Values["secret"]).To(Equal(redactedPersonalCredentialReference))
		Expect(application.Spec.Values["plain"]).To(Equal("ordinary"))
		Expect(application.Spec.Auth.AssetId).To(BeEmpty())
		Expect(values["secret"]).To(Equal("USERCREDENTIAL:generic-12345678-a"))
	})

	It("preserves canonical references for the creator", func() {
		application := api.AiApplication{
			CreatedBy: testing.TestUser,
			Spec: api.AiApplicationSpec{
				Values: map[string]interface{}{"secret": "USERCREDENTIAL:generic-12345678-a"},
				Auth: api.AiApplicationAuth{
					UserCredential: convutils.Ref(true),
					AssetId:        genericID,
				},
			},
		}
		sanitizeAIApplicationResponse(
			testing.SetContext(testing.TestTenant, testing.TestUser, nil),
			&application,
		)
		Expect(application.Spec.Values["secret"]).To(Equal("USERCREDENTIAL:generic-12345678-a"))
		Expect(application.Spec.Auth.AssetId).To(Equal(genericID))
	})

	DescribeTable("enforces the personal credentials capability boundary",
		func(version string, shouldSucceed bool) {
			clusterID := "00000000-0000-4000-8000-000000000031"
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(clusterID, testing.BuildDisplayedCluster(clusterID, version, nil)))
			apiErr := validateApplicationPersonalCredentialsCapability(
				testing.SetContext(testing.TestTenant, testing.TestUser, nil),
				clusterID,
			)
			if shouldSucceed {
				Expect(apiErr).To(BeNil())
			} else {
				Expect(apiErr).NotTo(BeNil())
				Expect(apiErr.Code).To(Equal(int32(400)))
			}
		},
		Entry("below 3.1", "3.0.9", false),
		Entry("at 3.1", "3.1.0", true),
	)
})
