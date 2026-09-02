/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcilers_test

import (
	"context"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/reconcilers"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var (
	limitRangeReconcileRequest = ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: DefaultLimitRangeConfigMap.Namespace,
			Name:      DefaultLimitRangeConfigMap.Name,
		},
	}
)

var _ = Describe("LimitRange Reconciler Tests", func() {

	var (
		reconciler         *LimitRangeReconciler
		projectEvents      chan event.GenericEvent
		k8sClient          client.Client
		project1, project2 *kaiv1alpha1.Project
	)

	BeforeEach(func() {
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		projectEvents = make(chan event.GenericEvent)
		reconciler = NewLimitRangeReconciler(k8sClient, scheme, projectEvents)
		project1 = TestProject.DeepCopy()
		project2 = KaiAnotherTestProject.DeepCopy()
	})

	It("Reconciles a cluster wide secret", func() {

		_ = k8sClient.Create(context.TODO(), project1)
		_ = k8sClient.Create(context.TODO(), project2)
		By("triggering events for all Projects", func() {

			result, err := reconciler.Reconcile(context.TODO(), limitRangeReconcileRequest)

			Eventually(func() []string {
				var actualEvents []string
				actualEvents = append(actualEvents, (<-projectEvents).Object.GetName())
				actualEvents = append(actualEvents, (<-projectEvents).Object.GetName())
				return actualEvents
			}, 2).Should(ConsistOf(project1.Name, project2.Name))

			Expect(result).To(BeZero())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
