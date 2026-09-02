/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/providers"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ctrlr "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/controller"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("ApplicationController", func() {
	var (
		ctx        context.Context
		scheme     *runtime.Scheme
		k8sCli     client.Client
		ctrlrC     *ctrlr.ApplicationController
		originalId types.UID
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		uid, _ := uuid.NewUUID()
		originalId = types.UID(uid.String())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
		k8sCli = fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&runv1alpha1.Application{}).Build()
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "ns1",
				Labels:      map[string]string{constants.RunaiQueueLabel: "true"},
				Annotations: map[string]string{constants.EnforceSchedulerAnnotationName: "true"},
			},
		}
		Expect(k8sCli.Create(ctx, ns)).To(Succeed())
		ctrlrC = ctrlr.NewApplicationController(k8sCli)
	})

	Describe("Reconcile", func() {
		It("updates Application status and resources when Application exists", func() {
			// Seed Application
			app := &runv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Name: "my-app", Namespace: "ns1"},
				Spec: runv1alpha1.ApplicationSpec{
					Source: providers.ApplicationHelmSource,
				},
			}
			app.Status.K8sPhase = "deployed"
			Expect(k8sCli.Create(ctx, app)).To(Succeed())

			// Seed managed resources with release annotation
			labels := map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue}
			annotations := map[string]string{constants.ResourceReleaseAnnotationName: "my-app"}

			Expect(k8sCli.Create(ctx, &runv1alpha1.ExternalWorkload{
				ObjectMeta: metav1.ObjectMeta{Name: "ew-1", Namespace: "ns1", Labels: labels, Annotations: annotations},
				Spec:       runv1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: originalId},
			})).To(Succeed())
			Expect(k8sCli.Create(ctx, &runv1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{Name: "ew-2", Namespace: "ns1", Labels: labels, Annotations: annotations}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "pvc-1", Namespace: "ns1", Labels: labels, Annotations: annotations}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cm-1", Namespace: "ns1", Labels: labels, Annotations: annotations}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cm-2", Namespace: "ns1", Labels: labels, Annotations: annotations}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "sec-1", Namespace: "ns1", Labels: labels, Annotations: annotations}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc-1", Namespace: "ns1", Labels: labels, Annotations: annotations}})).To(Succeed())
			// Seed managed resources without release annotation/ other namespace
			Expect(k8sCli.Create(ctx, &runv1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{Name: "ew-3", Namespace: "ns1", Labels: map[string]string{}, Annotations: map[string]string{}}})).To(Succeed())
			Expect(k8sCli.Create(ctx, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc-2", Namespace: "ns2", Labels: labels, Annotations: annotations}})).To(Succeed())

			// Reconcile
			_, err := ctrlrC.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: "my-app", Namespace: "ns1"}})
			Expect(err).NotTo(HaveOccurred())

			// Verify status updated
			updated := &runv1alpha1.Application{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "my-app", Namespace: "ns1"}, updated)).To(Succeed())
			Expect(updated.Status.K8sPhase).To(Equal("deployed"))

			names := map[string]bool{}
			var ew runv1alpha1.ApplicationResource
			for _, r := range updated.Status.Resources {
				names[r.Name] = true
				if r.Name == "ew-1" {
					ew = r
				}
			}
			Expect(names["ew-1"]).To(BeTrue())
			Expect(names["ew-2"]).To(BeTrue())
			Expect(names["pvc-1"]).To(BeTrue())
			Expect(names["cm-1"]).To(BeTrue())
			Expect(names["cm-2"]).To(BeTrue())
			Expect(names["sec-1"]).To(BeTrue())
			Expect(names["svc-1"]).To(BeTrue())
			Expect(ew.ControlPlaneId).To(Equal(string(originalId)))

			Expect(names["ew-3"]).To(BeFalse())
			Expect(names["svc-2"]).To(BeFalse())

			// Validate ApplicationIdLabel and ApplicationNameLabel are set on the ExternalWorkloads
			updatedEW1 := &runv1alpha1.ExternalWorkload{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "ew-1", Namespace: "ns1"}, updatedEW1)).To(Succeed())
			Expect(updatedEW1.Labels).To(HaveKey(constants.ApplicationIdLabel), "ExternalWorkload ew-1 should have ApplicationIdLabel")
			Expect(updatedEW1.Labels[constants.ApplicationIdLabel]).To(Equal(string(updated.UID)), "ApplicationIdLabel should match Application UID")
			Expect(updatedEW1.Labels).To(HaveKey(constants.ApplicationNameLabel), "ExternalWorkload ew-1 should have ApplicationNameLabel")
			Expect(updatedEW1.Labels[constants.ApplicationNameLabel]).To(Equal(updated.Name), "ApplicationNameLabel should match Application Name")

			updatedEW2 := &runv1alpha1.ExternalWorkload{}
			Expect(k8sCli.Get(ctx, client.ObjectKey{Name: "ew-2", Namespace: "ns1"}, updatedEW2)).To(Succeed())
			Expect(updatedEW2.Labels).To(HaveKey(constants.ApplicationIdLabel), "ExternalWorkload ew-2 should have ApplicationIdLabel")
			Expect(updatedEW2.Labels[constants.ApplicationIdLabel]).To(Equal(string(updated.UID)), "ApplicationIdLabel should match Application UID")
			Expect(updatedEW2.Labels).To(HaveKey(constants.ApplicationNameLabel), "ExternalWorkload ew-2 should have ApplicationNameLabel")
			Expect(updatedEW2.Labels[constants.ApplicationNameLabel]).To(Equal(updated.Name), "ApplicationNameLabel should match Application Name")
		})

		It("does not return error when Application does not exist", func() {
			_, err := ctrlrC.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: "missing", Namespace: "ns1"}})
			Expect(err).To(BeNil())
		})
	})

	Describe("ApplicationPredicate", func() {
		It("returns true when managed-by=Helm and release-name annotation present", func() {
			obj := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "cm",
				Namespace: "ns1",
				Labels:    map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{
					constants.ResourceReleaseAnnotationName: "my-app",
				},
			}}
			predicate := ctrlrC.CreateApplicationPredicate()
			Expect(predicate.Create(event.CreateEvent{Object: obj})).To(BeTrue())
			Expect(predicate.Update(event.UpdateEvent{ObjectNew: obj})).To(BeTrue())
			Expect(predicate.Delete(event.DeleteEvent{Object: obj})).To(BeTrue())
			Expect(predicate.Generic(event.GenericEvent{Object: obj})).To(BeTrue())
		})

		It("returns false when label missing", func() {
			obj := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "cm",
				Namespace: "ns1",
				Labels:    map[string]string{providers.ResourceManagedByLabelName: "lala"},
				Annotations: map[string]string{
					constants.ResourceReleaseAnnotationName: "my-app",
				},
			}}
			predicate := ctrlrC.CreateApplicationPredicate()
			Expect(predicate.Create(event.CreateEvent{Object: obj})).To(BeFalse())
		})

		It("returns false when annotation missing", func() {
			obj := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:        "cm1",
				Namespace:   "ns1",
				Labels:      map[string]string{providers.ResourceManagedByLabelName: providers.ResourceManagedByLabelValue},
				Annotations: map[string]string{},
			}}
			predicate := ctrlrC.CreateApplicationPredicate()
			Expect(predicate.Create(event.CreateEvent{Object: obj})).To(BeFalse())
		})
	})
})
