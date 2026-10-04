/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	workloads_openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	clusterCommonConstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("AI Application personal credentials", func() {
	newHandler := func(direct client.Client, cached client.Reader) *ApplicationsHandler {
		return &ApplicationsHandler{
			CommonHandler: &CommonHandler{
				runaiClient:  direct,
				cachedClient: cached,
			},
		}
	}

	It("rewrites nested canonical placeholders and deduplicates annotation sources", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		direct := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Secret{
			ObjectMeta: metav1ObjectMeta("token-1234567890", clusterCommonConstants.NamespaceRunAi, map[string]string{
				userCredentialsLabel:                   boolTrueAsString,
				clusterCommonConstants.RunaiAssetLabel: genericSecretKindLabel,
			}),
		}).Build()
		handler := newHandler(direct, direct)
		item := &workloads_openapi.AIApplicationSyncItem{
			Id:        convutils.Ref("app-id"),
			Name:      convutils.Ref("my-app"),
			Source:    convutils.Ref("helm"),
			Type:      convutils.Ref("chart"),
			ProjectId: convutils.Ref("project-id"),
			Spec: &workloads_openapi.AIApplicationSpec{
				ChartRef: workloads_openapi.AIApplicationSpecChartRef{
					Repo: convutils.Ref("https://charts.example.com"),
				},
				Values: map[string]interface{}{
					"secretName": "USERCREDENTIAL:token-1234567890",
					"nested": []interface{}{
						map[string]interface{}{"same": "USERCREDENTIAL:token-1234567890"},
						"prefix-USERCREDENTIAL:leave-embedded-unchanged",
					},
				},
			},
		}

		app, err := handler.convertToApplication(context.Background(), applicationProjectInfo{namespace: "project-ns"}, item, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(app.Spec.Helm.Values["secretName"].Raw).To(MatchJSON(`"pc-my-app-token-1234567890"`))
		Expect(app.Spec.Helm.Values["nested"].Raw).To(MatchJSON(
			`[{"same":"pc-my-app-token-1234567890"},"prefix-USERCREDENTIAL:leave-embedded-unchanged"]`,
		))
		credentials, err := personalcreds.Get(app)
		Expect(err).NotTo(HaveOccurred())
		Expect(credentials).To(Equal([]personalcreds.Cred{{SourceSecret: "token-1234567890"}}))
	})

	DescribeTable("selects the required personal chart credential representation",
		func(repository, chartURL, expectedKind, sourceName string) {
			scheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			direct := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Secret{
				ObjectMeta: metav1ObjectMeta(sourceName, clusterCommonConstants.NamespaceRunAi, map[string]string{
					assetIdLabel:                           "asset-id",
					userCredentialsLabel:                   boolTrueAsString,
					clusterCommonConstants.RunaiAssetLabel: expectedKind,
				}),
			}).Build()
			cachedWithoutSource := fake.NewClientBuilder().WithScheme(scheme).Build()
			handler := newHandler(direct, cachedWithoutSource)
			auth := &workloads_openapi.AIApplicationAuth{AssetId: convutils.Ref("asset-id")}
			auth.SetUserCredential(true)
			spec := &workloads_openapi.AIApplicationSpec{
				ChartRef: workloads_openapi.AIApplicationSpecChartRef{
					Repo:     convutils.Ref(repository),
					ChartURL: convutils.Ref(chartURL),
				},
				Auth: auth,
			}

			helmSpec, credentials, err := handler.convertHelmSpec(
				context.Background(), "my-app", "project-ns", spec, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(helmSpec.Auth.SecretRef.Name).To(Equal(personalcreds.DestName("my-app", sourceName)))
			Expect(credentials).To(ConsistOf(personalcreds.Cred{SourceSecret: sourceName}))
		},
		Entry("generic for HTTP repository", "https://charts.example.com", "", genericSecretKindLabel, "ngcgs-token"),
		Entry("docker registry for OCI repository", "oci://registry.example.com/charts", "", dockerRegistryKindLabel, "ngcdr-token"),
		Entry("docker registry for direct OCI URL", "", "oci://registry.example.com/charts/app", dockerRegistryKindLabel, "ngcdr-token"),
	)

	It("requires a personal source and prefers the repository-compatible representation", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		direct := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&corev1.Secret{ObjectMeta: metav1ObjectMeta("wrong-private-label", clusterCommonConstants.NamespaceRunAi, map[string]string{
				assetIdLabel:                           "asset-id",
				userCredentialsLabel:                   "false",
				clusterCommonConstants.RunaiAssetLabel: genericSecretKindLabel,
			})},
		).Build()
		handler := newHandler(direct, direct)

		_, err := handler.resolvePersonalSecretNameFromAssetId(context.Background(), "asset-id", genericSecretKindLabel)
		Expect(err).To(MatchError("no personal credential source found"))

		Expect(direct.Create(context.Background(), &corev1.Secret{
			ObjectMeta: metav1ObjectMeta("docker-source", clusterCommonConstants.NamespaceRunAi, map[string]string{
				assetIdLabel:                           "asset-id",
				userCredentialsLabel:                   boolTrueAsString,
				clusterCommonConstants.RunaiAssetLabel: dockerRegistryKindLabel,
			}),
		})).To(Succeed())
		sourceName, err := handler.resolvePersonalSecretNameFromAssetId(context.Background(), "asset-id", genericSecretKindLabel)
		Expect(err).NotTo(HaveOccurred())
		Expect(sourceName).To(Equal("docker-source"))

		Expect(direct.Create(context.Background(), &corev1.Secret{
			ObjectMeta: metav1ObjectMeta("generic-source", clusterCommonConstants.NamespaceRunAi, map[string]string{
				assetIdLabel:                           "asset-id",
				userCredentialsLabel:                   boolTrueAsString,
				clusterCommonConstants.RunaiAssetLabel: genericSecretKindLabel,
			}),
		})).To(Succeed())
		sourceName, err = handler.resolvePersonalSecretNameFromAssetId(context.Background(), "asset-id", genericSecretKindLabel)
		Expect(err).NotTo(HaveOccurred())
		Expect(sourceName).To(Equal("generic-source"))
	})

	It("preserves project-scoped shared chart authentication", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		shared := &corev1.Secret{ObjectMeta: metav1ObjectMeta("shared-secret", "project-ns", map[string]string{
			assetIdLabel:                           "shared-id",
			clusterCommonConstants.RunaiAssetLabel: genericSecretKindLabel,
		})}
		cached := fake.NewClientBuilder().WithScheme(scheme).WithObjects(shared).Build()
		handler := newHandler(cached, cached)
		spec := &workloads_openapi.AIApplicationSpec{
			ChartRef: workloads_openapi.AIApplicationSpecChartRef{Repo: convutils.Ref("https://charts.example.com")},
			Auth:     &workloads_openapi.AIApplicationAuth{AssetId: convutils.Ref("shared-id")},
		}

		helmSpec, credentials, err := handler.convertHelmSpec(
			context.Background(), "my-app", "project-ns", spec, false,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(helmSpec.Auth.SecretRef.Name).To(Equal("shared-secret"))
		Expect(credentials).To(BeEmpty())
	})

})

func metav1ObjectMeta(name, namespace string, labels map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels}
}
