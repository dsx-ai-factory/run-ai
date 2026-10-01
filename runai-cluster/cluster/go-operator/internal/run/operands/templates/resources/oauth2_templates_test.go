/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"errors"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/templates"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/templates/resources/configmaps"
	kcfake "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client/fake"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("OAuth2 templates ConfigMap", func() {
	var (
		ctx         context.Context
		runaiConfig runv1.RunaiConfig
		k8s         client.Client
		resource    *configmaps.OAuth2Templates
	)

	BeforeEach(func() {
		ctx = context.Background()
		runaiConfig = runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "operator-namespace"}}
		k8s = fake.NewClientBuilder().Build()
		resource = configmaps.NewOAuth2Templates("templates", k8s, runaiConfig)
	})

	It("registers a cluster-wide source with the Run:ai error page in the operator namespace", func() {
		operand := templates.Operand(kcfake.New(k8s, k8s, nil), runaiConfig)
		var registered *configmaps.OAuth2Templates
		for _, candidate := range operand.Resources() {
			if templateResource, ok := candidate.(*configmaps.OAuth2Templates); ok {
				registered = templateResource
			}
		}
		Expect(registered).NotTo(BeNil())

		obj, err := registered.Object(ctx)
		Expect(err).NotTo(HaveOccurred())
		configMap := obj.(*corev1.ConfigMap)
		Expect(configMap.Name).To(Equal(oauth2proxy.ConfigMapName))
		Expect(configMap.Namespace).To(Equal(runaiConfig.Namespace))
		Expect(configMap.Labels).To(HaveKeyWithValue(commonconstants.ClusterWideResourceLabel, "true"))
		Expect(configMap.Data).To(HaveLen(1))
		Expect(configMap.Data[oauth2proxy.TemplateKey]).To(ContainSubstring("Sorry, we can't let you see this page."))
	})

	It("repairs template content and the replication label while preserving metadata", func() {
		obj, err := resource.Object(ctx)
		Expect(err).NotTo(HaveOccurred())
		expectedData := obj.(*corev1.ConfigMap).Data
		existing := obj.(*corev1.ConfigMap).DeepCopy()
		existing.UID = "existing-configmap"
		existing.Labels = map[string]string{"custom-label": "preserved", commonconstants.ClusterWideResourceLabel: "false"}
		existing.Annotations = map[string]string{"custom-annotation": "preserved"}
		existing.OwnerReferences = []metav1.OwnerReference{{APIVersion: "run.ai/v1", Kind: "RunaiConfig", Name: "runai", UID: "owner"}}
		existing.BinaryData = map[string][]byte{oauth2proxy.TemplateKey: []byte("outdated page")}
		existing.Data = nil
		Expect(k8s.Create(ctx, existing)).To(Succeed())

		obj, err = resource.Object(ctx)
		Expect(err).NotTo(HaveOccurred())
		actual := obj.(*corev1.ConfigMap)
		Expect(actual.Data).To(Equal(expectedData))
		Expect(actual.BinaryData).To(BeNil())
		Expect(actual.Labels).To(HaveKeyWithValue(commonconstants.ClusterWideResourceLabel, "true"))
		Expect(actual.Labels).To(HaveKeyWithValue("custom-label", "preserved"))
		Expect(actual.Annotations).To(Equal(existing.Annotations))
		Expect(actual.OwnerReferences).To(Equal(existing.OwnerReferences))
		Expect(actual.UID).To(Equal(existing.UID))
		Expect(actual.ResourceVersion).To(Equal(existing.ResourceVersion))
	})

	It("reports a missing source as not ready", func() {
		Expect(kerrors.IsNotFound(resource.Status(ctx))).To(BeTrue())
	})

	DescribeTable("checks the live source before reporting ready", func(change func(*corev1.ConfigMap), ready bool) {
		obj, err := resource.Object(ctx)
		Expect(err).NotTo(HaveOccurred())
		configMap := obj.(*corev1.ConfigMap)
		change(configMap)
		Expect(k8s.Create(ctx, configMap)).To(Succeed())
		if ready {
			Expect(resource.Status(ctx)).To(Succeed())
		} else {
			Expect(resource.Status(ctx)).To(HaveOccurred())
		}
	},
		Entry("current template and replication label", func(*corev1.ConfigMap) {}, true),
		Entry("outdated template", func(cm *corev1.ConfigMap) { cm.Data[oauth2proxy.TemplateKey] = "old page" }, false),
		Entry("missing template", func(cm *corev1.ConfigMap) { cm.Data = nil }, false),
		Entry("missing replication label", func(cm *corev1.ConfigMap) { cm.Labels = nil }, false),
	)

	It("preserves API errors from object construction and status", func() {
		readErr := errors.New("API unavailable")
		k8s = fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
			Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
				return readErr
			},
		}).Build()
		resource = configmaps.NewOAuth2Templates("templates", k8s, runaiConfig)
		_, err := resource.Object(ctx)
		Expect(errors.Is(err, readErr)).To(BeTrue())
		Expect(errors.Is(resource.Status(ctx), readErr)).To(BeTrue())
	})
})
