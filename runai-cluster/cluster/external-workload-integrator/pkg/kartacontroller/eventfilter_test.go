/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kartacontroller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("Event filter", Ordered, func() {
	var (
		scheme *runtime.Scheme
		ctx    context.Context
		rc     *KartaController
	)

	BeforeAll(func() {
		scheme = buildTestScheme()
		ctx = context.Background()
	})

	BeforeEach(func() {
		k8s := fake.NewClientBuilder().WithScheme(scheme).Build()
		rc = NewKartaController(k8s, &KartaControlParams{})
	})

	Context("MapClusterRoleToKartaEvent", func() {
		It("returns one reconcile request per existing Karta", func() {
			k8s := fake.NewClientBuilder().WithScheme(scheme).Build()
			rc = NewKartaController(k8s, &KartaControlParams{})

			// create a few Kartas
			expectCreate(ctx, rc.Client, buildKarta("karta-one", nil))
			expectCreate(ctx, rc.Client, buildKarta("karta-two", nil))

			// trigger mapping on any ClusterRole event
			cr := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "any-cr"}}
			reqs := rc.MapClusterRoleToKartaEvent(ctx, cr)

			Expect(reqs).To(HaveLen(2))
			names := []string{reqs[0].Name, reqs[1].Name}
			Expect(names).To(ContainElements("karta-one", "karta-two"))
		})
	})

	Context("ControllerPredicateFuncs", func() {
		It("filters ClusterRole events by configured names and allows others", func() {
			// configure rc with specific ClusterRole names
			k8s := fake.NewClientBuilder().WithScheme(scheme).Build()
			rc = NewKartaController(k8s, &KartaControlParams{ClusterRoleReadName: "karta-read", ClusterRoleCrudName: "karta-crud"})

			preds := rc.ControllerPredicateFuncs()

			crRead := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "karta-read"}}
			crCrud := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "karta-crud"}}
			crOther := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "something-else"}}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p"}}

			// Create events
			Expect(preds.CreateFunc(event.CreateEvent{Object: crRead})).To(BeTrue())
			Expect(preds.CreateFunc(event.CreateEvent{Object: crCrud})).To(BeTrue())
			Expect(preds.CreateFunc(event.CreateEvent{Object: crOther})).To(BeFalse())
			Expect(preds.CreateFunc(event.CreateEvent{Object: pod})).To(BeTrue())

			// Update events
			Expect(preds.UpdateFunc(event.UpdateEvent{ObjectNew: crRead})).To(BeTrue())
			Expect(preds.UpdateFunc(event.UpdateEvent{ObjectNew: crCrud})).To(BeTrue())
			Expect(preds.UpdateFunc(event.UpdateEvent{ObjectNew: crOther})).To(BeFalse())
			Expect(preds.UpdateFunc(event.UpdateEvent{ObjectNew: pod})).To(BeTrue())

			// Delete events
			Expect(preds.DeleteFunc(event.DeleteEvent{Object: crRead})).To(BeTrue())
			Expect(preds.DeleteFunc(event.DeleteEvent{Object: crCrud})).To(BeTrue())
			Expect(preds.DeleteFunc(event.DeleteEvent{Object: crOther})).To(BeFalse())
			Expect(preds.DeleteFunc(event.DeleteEvent{Object: pod})).To(BeTrue())

			// Generic events
			Expect(preds.GenericFunc(event.GenericEvent{Object: crRead})).To(BeTrue())
			Expect(preds.GenericFunc(event.GenericEvent{Object: crCrud})).To(BeTrue())
			Expect(preds.GenericFunc(event.GenericEvent{Object: crOther})).To(BeFalse())
			Expect(preds.GenericFunc(event.GenericEvent{Object: pod})).To(BeTrue())
		})
	})
})
