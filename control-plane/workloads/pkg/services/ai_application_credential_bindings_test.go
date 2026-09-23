/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/base64"
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	assetsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	clusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AI Application runtime credential bindings", func() {
	const assetID = "12345678-abcd-4000-8000-000000000001"
	var request api.AiApplicationCreationRequest
	var credential assetsApi.CredentialsListResponseEntry
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, nil)
	BeforeEach(func() {
		clusterClient.SetMock(clusterClient.NewSimpleMockClient().WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, "3.1.0", nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
		credential = assetsApi.CredentialsListResponseEntry{Meta: assetsApi.AssetMeta{
			Id: assetID, Name: "selected", Kind: assetsApi.ASSETKIND_NGC_API_KEY, Scope: assetsApi.SCOPE_TENANT, CreatedBy: testing.TestUser,
		}}
		request = api.AiApplicationCreationRequest{
			Name: "my-app", ClusterId: testing.ClusterUuid1, ProjectId: "1",
			Spec: api.AiApplicationSpec{CredentialBindings: []api.AiApplicationCredentialBinding{
				{Variable: "global.ngcApiSecret.name", AssetId: assetID, CredentialType: "ngc-api-key"},
			}},
		}
	})
	process := func() (bool, *api.Error) {
		assetsClient.SetMock(assetsClient.NewSimpleMock().WithCredentials([]assetsApi.CredentialsListResponseEntry{credential}))
		return processAIApplicationCredentialBindings(ctx, &request, "department")
	}

	It("keeps IDs separate from values and derives ownership on the server", func() {
		request.Spec.Values = map[string]interface{}{"unrelated": "preserved"}
		request.Spec.CredentialBindings[0].UserCredential = true
		request.Spec.CredentialBindings[0].ExistingSecretName = "untrusted-client-name"
		personal, err := process()
		Expect(err).To(BeNil())
		Expect(personal).To(BeFalse())
		Expect(request.Spec.CredentialBindings[0].UserCredential).To(BeFalse())
		Expect(request.Spec.CredentialBindings[0].ExistingSecretName).To(BeEmpty())
		Expect(request.Spec.Values).To(Equal(map[string]interface{}{"unrelated": "preserved"}))
	})

	It("accepts two uses of the same personal NGC asset and redacts its IDs for another viewer", func() {
		credential.Meta.Private.Set(convutils.Ref(true))
		request.Spec.CredentialBindings = append(request.Spec.CredentialBindings, api.AiApplicationCredentialBinding{
			Variable: "global.imagePullSecret.name", AssetId: assetID, CredentialType: "image-pull-secret",
		})
		personal, err := process()
		Expect(err).To(BeNil())
		Expect(personal).To(BeTrue())
		Expect(request.Spec.CredentialBindings[0].UserCredential).To(BeTrue())
		Expect(request.Spec.CredentialBindings[1].UserCredential).To(BeTrue())
		app := &api.AiApplication{CreatedBy: testing.TestUser, Spec: request.Spec}
		sanitizeAIApplicationResponse(testing.SetContext(testing.TestTenant, testing.AnotherUser, nil), app)
		Expect(app.Spec.CredentialBindings[0].AssetId).To(BeEmpty())
		Expect(app.Spec.CredentialBindings[1].AssetId).To(BeEmpty())
	})

	It("rejects deleted and inaccessible selections with a reselection message", func() {
		assetsClient.SetMock(assetsClient.NewSimpleMock())
		_, err := processAIApplicationCredentialBindings(ctx, &request, "department")
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("no longer exists or is not accessible. Please reselect"))
	})

	It("drops overridden literals and personal placeholders before persistence and legacy resolution", func() {
		request.Spec.Values = map[string]interface{}{"global": map[string]interface{}{"ngcApiSecret": map[string]interface{}{
			"name": "USERCREDENTIAL:deleted-personal-credential", "create": "test-only-sensitive-literal", "keep": true,
		}}}
		_, err := process()
		Expect(err).To(BeNil())
		used, err := processAIApplicationCredentials(ctx, request.Name, &request.Spec)
		Expect(err).To(BeNil())
		Expect(used).To(BeFalse())
		raw, marshalErr := json.Marshal(request.Spec.Values)
		Expect(marshalErr).NotTo(HaveOccurred())
		Expect(string(raw)).To(MatchJSON(`{"global":{"ngcApiSecret":{"name":null,"create":null,"keep":true}}}`))
	})

	It("rejects another user's personal credential and credentials from another project", func() {
		credential.Meta.Private.Set(convutils.Ref(true))
		credential.Meta.CreatedBy = testing.AnotherUser
		_, err := process()
		Expect(err).NotTo(BeNil())
		credential.Meta.Private.Set(convutils.Ref(false))
		credential.Meta.Scope = assetsApi.SCOPE_PROJECT
		credential.Meta.ClusterId.Set(convutils.Ref(request.ClusterId))
		credential.Meta.ProjectId.Set(convutils.Ref(int32(2)))
		_, err = process()
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("target project"))
	})

	DescribeTable("enforces compatible Secret representations",
		func(kind assetsApi.AssetKind, fieldType string, valid bool) {
			credential.Meta.Kind = kind
			request.Spec.CredentialBindings[0].CredentialType = fieldType
			_, err := process()
			Expect(err == nil).To(Equal(valid))
		},
		Entry("NGC API", assetsApi.ASSETKIND_NGC_API_KEY, "ngc-api-key", true),
		Entry("NGC image pull", assetsApi.ASSETKIND_NGC_API_KEY, "image-pull-secret", true),
		Entry("NGC generic", assetsApi.ASSETKIND_NGC_API_KEY, "generic", true),
		Entry("registry image pull", assetsApi.ASSETKIND_DOCKER_REGISTRY, "image-pull-secret", true),
		Entry("registry is not generic", assetsApi.ASSETKIND_DOCKER_REGISTRY, "generic", false),
		Entry("generic is not NGC", assetsApi.ASSETKIND_GENERIC_SECRET, "ngc-api-key", false),
		Entry("password unsupported", assetsApi.ASSETKIND_PASSWORD, "generic", false),
		Entry("old enum rejected", assetsApi.ASSETKIND_DOCKER_REGISTRY, "docker-registry", false),
	)

	It("normalizes base64 values while preserving unrelated overrides", func() {
		request.Spec.ValuesBase64 = base64.StdEncoding.EncodeToString([]byte("pulls:\n  - name: existing\n    extra: preserve\n"))
		_, err := process()
		Expect(err).To(BeNil())
		Expect(request.Spec.ValuesBase64).To(BeEmpty())
		Expect(request.Spec.Values["pulls"]).To(Equal([]interface{}{map[string]interface{}{"name": "existing", "extra": "preserve"}}))
	})

	It("rejects conflicting destinations and mutually exclusive value inputs", func() {
		request.Spec.CredentialBindings = append(request.Spec.CredentialBindings, api.AiApplicationCredentialBinding{
			Variable: "global.ngcApiSecret.create", AssetId: assetID, CredentialType: "generic",
		})
		_, err := process()
		Expect(err).NotTo(BeNil())
		request.Spec.CredentialBindings = request.Spec.CredentialBindings[:1]
		request.Spec.Values = map[string]interface{}{}
		request.Spec.ValuesBase64 = "e30="
		_, err = process()
		Expect(err).NotTo(BeNil())
	})

	It("gates older clusters before their agents could ignore bindings", func() {
		clusterClient.SetMock(clusterClient.NewSimpleMockClient().WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, "3.0.0", nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
		_, err := process()
		Expect(err).NotTo(BeNil())
	})

	It("takes existing Secret names only from authorized asset metadata and never copies data", func() {
		credential.Meta.Kind = assetsApi.ASSETKIND_GENERIC_SECRET
		request.Spec.CredentialBindings[0].CredentialType = "generic"
		credential.Spec.GenericSecret.Set(&assetsApi.GenericSecretSpec{ExistingSecretName: *assetsApi.NewNullableString(convutils.Ref("preexisting-reference"))})
		_, err := process()
		Expect(err).To(BeNil())
		Expect(request.Spec.CredentialBindings[0].ExistingSecretName).To(Equal("preexisting-reference"))
		raw, marshalErr := json.Marshal(request.Spec)
		Expect(marshalErr).NotTo(HaveOccurred())
		Expect(string(raw)).NotTo(ContainSubstring("keyValuePairs"))
	})
})
