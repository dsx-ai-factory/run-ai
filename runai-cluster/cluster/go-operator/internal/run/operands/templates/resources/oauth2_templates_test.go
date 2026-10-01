/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"errors"
	"fmt"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/templates"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/templates/resources/configmaps"
	cu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	kcfake "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client/fake"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
		runaiConfig = runv1.RunaiConfig{
			TypeMeta:   metav1.TypeMeta{APIVersion: runv1.GroupVersion.String(), Kind: runv1.RunaiConfigKind},
			ObjectMeta: metav1.ObjectMeta{Name: "runai", Namespace: "operator-namespace", UID: "owner"},
		}
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(runv1.AddToScheme(scheme)).To(Succeed())
		k8s = fake.NewClientBuilder().WithScheme(scheme).Build()
		resource = configmaps.NewOAuth2Templates("templates", k8s, runaiConfig)
	})

	controllerReference := func() metav1.OwnerReference {
		return *metav1.NewControllerRef(&runaiConfig, runv1.GroupVersion.WithKind(runv1.RunaiConfigKind))
	}

	reconcile := func(ownResources bool) error {
		cluster := &runv1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: runv1.ClusterName, UID: "cluster-owner"}}
		return cu.CreateOrUpdateResources(k8s, ctx, []operands.Resource{resource}, &runaiConfig, cluster, ownResources)
	}

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
		existing.OwnerReferences = []metav1.OwnerReference{controllerReference()}
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

	for _, ownResources := range []bool{true, false} {
		Context(fmt.Sprintf("with ownResources=%t", ownResources), func() {
			DescribeTable("leaves a foreign ConfigMap unchanged", func(change func(*corev1.ConfigMap)) {
				existing := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: oauth2proxy.ConfigMapName, Namespace: runaiConfig.Namespace,
						Labels: map[string]string{"custom-label": "preserved"},
					},
					Data:       map[string]string{"legacy": "preserved"},
					BinaryData: map[string][]byte{oauth2proxy.TemplateKey: []byte("existing page")},
				}
				change(existing)
				Expect(k8s.Create(ctx, existing)).To(Succeed())
				before := existing.DeepCopy()

				Expect(reconcile(ownResources)).To(HaveOccurred())

				actual := &corev1.ConfigMap{}
				Expect(k8s.Get(ctx, client.ObjectKeyFromObject(existing), actual)).To(Succeed())
				Expect(actual).To(Equal(before))
			},
				Entry("foreign controller", func(cm *corev1.ConfigMap) {
					owner := controllerReference()
					owner.APIVersion = "apps/v1"
					owner.Kind = "Deployment"
					owner.Name = "other-controller"
					owner.UID = "other-owner"
					cm.OwnerReferences = []metav1.OwnerReference{owner}
				}),
				Entry("different controller UID", func(cm *corev1.ConfigMap) {
					owner := controllerReference()
					owner.UID = "previous-owner"
					cm.OwnerReferences = []metav1.OwnerReference{owner}
				}),
				Entry("different annotated owner", func(cm *corev1.ConfigMap) {
					cm.Annotations = map[string]string{runv1.NamespacedNameAnnotation: "other-namespace/runai"}
				}),
				Entry("different annotated owner type", func(cm *corev1.ConfigMap) {
					cm.Annotations = map[string]string{runv1.TypeAnnotation: "OtherController.run.ai"}
				}),
			)

			DescribeTable("repairs an eligible ConfigMap", func(change func(*corev1.ConfigMap)) {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())
				existing := obj.(*corev1.ConfigMap)
				expectedData := existing.Data
				existing.Data = nil
				existing.BinaryData = map[string][]byte{oauth2proxy.TemplateKey: []byte("outdated page")}
				existing.Labels = map[string]string{commonconstants.ClusterWideResourceLabel: "false"}
				change(existing)
				Expect(k8s.Create(ctx, existing)).To(Succeed())

				Expect(reconcile(ownResources)).To(Succeed())

				actual := &corev1.ConfigMap{}
				Expect(k8s.Get(ctx, client.ObjectKeyFromObject(existing), actual)).To(Succeed())
				Expect(actual.Data).To(Equal(expectedData))
				Expect(actual.BinaryData).To(BeNil())
				Expect(actual.Labels).To(HaveKeyWithValue(commonconstants.ClusterWideResourceLabel, "true"))
				Expect(resource.Status(ctx)).To(Succeed())
			},
				Entry("unowned legacy source", func(*corev1.ConfigMap) {}),
				Entry("current RunaiConfig controller and non-controller Cluster owner", func(cm *corev1.ConfigMap) {
					cm.OwnerReferences = []metav1.OwnerReference{
						controllerReference(),
						{APIVersion: runv1.GroupVersion.String(), Kind: runv1.ClusterKind, Name: runv1.ClusterName, UID: "cluster-owner"},
					}
				}),
				Entry("matching remote ownership annotations", func(cm *corev1.ConfigMap) {
					cm.Annotations = map[string]string{
						runv1.NamespacedNameAnnotation: client.ObjectKeyFromObject(&runaiConfig).String(),
						runv1.TypeAnnotation:           runv1.GroupVersion.WithKind(runv1.RunaiConfigKind).GroupKind().String(),
					}
				}),
			)
		})
	}

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
