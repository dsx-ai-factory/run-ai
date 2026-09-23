/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("AI Application runtime credential bindings", func() {
	secret := func(name, namespace, kind string, personal bool) *corev1.Secret {
		item := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{
			assetIdLabel: "asset-id", constants.RunaiAssetLabel: kind,
		}}, Data: map[string][]byte{"NGC_API_KEY": []byte("test-only-secret-value")}}
		if kind == dockerRegistryKindLabel {
			item.Data = map[string][]byte{corev1.DockerConfigJsonKey: []byte("test-only-secret-value")}
			if !personal {
				item.Type = corev1.SecretTypeDockerConfigJson
			}
		}
		if personal {
			item.Labels[userCredentialsLabel] = boolTrueAsString
		}
		return item
	}
	newHandler := func(items ...client.Object) *ApplicationsHandler {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(items...).Build()
		return &ApplicationsHandler{CommonHandler: &CommonHandler{runaiClient: c, cachedClient: c}}
	}
	ngcBindings := func(personal bool) *workloadsapi.AIApplicationSpec {
		return &workloadsapi.AIApplicationSpec{CredentialBindings: []workloadsapi.AIApplicationCredentialBinding{
			{Variable: "global.ngcApiSecret.name", AssetId: "asset-id", CredentialType: "ngc-api-key", UserCredential: convutils.Ref(personal)},
			{Variable: "global.imagePullSecret.name", AssetId: "asset-id", CredentialType: "image-pull-secret", UserCredential: convutils.Ref(personal)},
		}}
	}

	It("resolves the two NGC representations without including secret contents", func() {
		handler := newHandler(secret("api-reference", "project", genericSecretKindLabel, false), secret("pull-reference", "project", dockerRegistryKindLabel, false))
		helm, sources, err := handler.convertHelmSpec(context.Background(), "app", "project", ngcBindings(false), false)
		Expect(err).NotTo(HaveOccurred())
		Expect(sources).To(BeEmpty())
		Expect(helm.CredentialBindings[0].SecretName).To(Equal("api-reference"))
		Expect(helm.CredentialBindings[1].SecretName).To(Equal("pull-reference"))
		raw, err := json.Marshal(helm)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).NotTo(ContainSubstring("test-only-secret-value"))
		Expect(helm.Values).To(BeEmpty())
	})

	It("reuses personal materialization for encrypted generic and registry sources", func() {
		handler := newHandler(secret("api-source", constants.NamespaceRunAi, genericSecretKindLabel, true), secret("pull-source", constants.NamespaceRunAi, dockerRegistryKindLabel, true))
		helm, sources, err := handler.convertHelmSpec(context.Background(), "app", "project", ngcBindings(true), false)
		Expect(err).NotTo(HaveOccurred())
		Expect(helm.CredentialBindings[0].SecretName).To(Equal(personalcreds.DestName("app", "api-source")))
		Expect(helm.CredentialBindings[1].SecretName).To(Equal(personalcreds.DestName("app", "pull-source")))
		Expect(sources).To(ConsistOf(personalcreds.Cred{SourceSecret: "api-source"}, personalcreds.Cred{SourceSecret: "pull-source"}))
	})

	It("rejects a missing representation without falling back to another NGC Secret", func() {
		handler := newHandler(secret("api-reference", "project", genericSecretKindLabel, false))
		_, _, err := handler.convertHelmSpec(context.Background(), "app", "project", ngcBindings(false), false)
		Expect(err).To(MatchError(ContainSubstring("check credential synchronization")))
	})

	It("rejects another namespace, duplicate matches, and revoked personal sources", func() {
		for _, items := range [][]client.Object{
			{secret("api-reference", "other-project", genericSecretKindLabel, false)},
			{secret("first", "project", genericSecretKindLabel, false), secret("second", "project", genericSecretKindLabel, false)},
		} {
			spec := ngcBindings(false)
			spec.CredentialBindings = spec.CredentialBindings[:1]
			_, _, err := newHandler(items...).convertHelmSpec(context.Background(), "app", "project", spec, false)
			Expect(err).To(HaveOccurred())
		}
		revoked := secret("api-source", constants.NamespaceRunAi, genericSecretKindLabel, true)
		delete(revoked.Labels, userCredentialsLabel)
		_, _, err := newHandler(revoked).convertHelmSpec(context.Background(), "app", "project", ngcBindings(true), false)
		Expect(err).To(HaveOccurred())
	})

	It("supports an authorized existing Secret and checks its representation", func() {
		existing := secret("external-reference", "project", genericSecretKindLabel, false)
		existing.Labels = nil
		spec := &workloadsapi.AIApplicationSpec{CredentialBindings: []workloadsapi.AIApplicationCredentialBinding{{
			Variable: "custom.name", AssetId: "asset-id", CredentialType: "generic", ExistingSecretName: convutils.Ref(existing.Name),
		}}}
		helm, _, err := newHandler(existing).convertHelmSpec(context.Background(), "app", "project", spec, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(helm.CredentialBindings[0].SecretName).To(Equal(existing.Name))
		existing.Type = corev1.SecretTypeDockerConfigJson
		_, _, err = newHandler(existing).convertHelmSpec(context.Background(), "app", "project", spec, false)
		Expect(err).To(HaveOccurred())
	})
})
