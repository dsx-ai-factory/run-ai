/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers_test

import (
	"context"
	"reflect"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("LimitRange Resource Handler", func() {

	var (
		handler                                                       LimitRangeResourceHandler
		client                                                        client.Client
		project                                                       kaiv1alpha1.Project
		defaultLimitRangeConfigMap, anotherDefaultLimitRangeConfigMap corev1.ConfigMap
		namespacedLimitRange, anotherNamespacedLimitRange             corev1.LimitRange
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		defaultLimitRangeConfigMap = *DefaultLimitRangeConfigMap.DeepCopy()
		namespacedLimitRange = *NamespacedLimitRange.DeepCopy()
		anotherDefaultLimitRangeConfigMap = *AnotherDefaultLimitRangeConfigMap.DeepCopy()
		anotherNamespacedLimitRange = *AnotherNamespacedLimitRange.DeepCopy()

		client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(TestNamespace.DeepCopy()).Build()
		handler = NewLimitRangeResourceHandler(client)
	})

	When("no default LimitRange Config Map exists in namespace 'runai'", func() {
		It("don't create namespaced LimitRange resources", func() {

			// Given
			Expect(client.Get(context.TODO(), types.NamespacedName{Name: DefaultLimitRangeConfigMapName, Namespace: config.Get().InstallNamespace}, &corev1.ConfigMap{})).ToNot(Succeed()) //sanity

			// When
			conditions, err := handler.HandleResource(project)
			Expect(err).To(Succeed())

			// Then
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.ProjectConditionType("LimitRangeReady")))
			Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))
			Expect(conditions[0].Reason).To(Equal(""))
			Expect(conditions[0].Message).To(Equal(""))

			actualLimitRange := corev1.LimitRange{}
			Expect(client.Get(context.TODO(), types.NamespacedName{Name: RunaiLimitRangeName, Namespace: TestNamespace.Name}, &actualLimitRange)).ToNot(Succeed())
			Expect(actualLimitRange).To(BeZero())
		})
	})

	When("namespaced LimitRange doesn't exit", func() {
		It("create it", func() {

			// Given
			Expect(client.Create(context.TODO(), &defaultLimitRangeConfigMap)).To(Succeed())
			Expect(client.Create(context.TODO(), &corev1.LimitRange{})).NotTo(Succeed()) // sanity

			// When
			_, err := handler.HandleResource(project)
			Expect(err).To(Succeed())

			// Then
			actualLimitRange := corev1.LimitRange{}
			Expect(client.Get(context.TODO(), types.NamespacedName{Name: RunaiLimitRangeName, Namespace: TestNamespace.Name}, &actualLimitRange)).To(Succeed())
			assertLimitRangesEqual(actualLimitRange, namespacedLimitRange)
		})
	})

	When("namespaced LimitRange does exit, and is identical to the global Config Map", func() {
		It("the handler doesn't modify it", func() {

			// Given
			Expect(client.Create(context.TODO(), &defaultLimitRangeConfigMap)).To(Succeed())
			Expect(client.Create(context.TODO(), &namespacedLimitRange)).To(Succeed())

			// When
			_, err := handler.HandleResource(project)
			Expect(err).To(Succeed())

			// Then
			actualLimitRange := corev1.LimitRange{}
			Expect(client.Get(context.TODO(), types.NamespacedName{Name: RunaiLimitRangeName, Namespace: TestNamespace.Name}, &actualLimitRange)).To(Succeed())
			assertLimitRangesEqual(actualLimitRange, namespacedLimitRange)
		})
	})

	When("namespaced LimitRange does exit, and is different from the global Config Map", func() {
		It("the handler modifies it", func() {

			// Given
			Expect(client.Create(context.TODO(), &anotherDefaultLimitRangeConfigMap)).To(Succeed())
			Expect(client.Create(context.TODO(), &namespacedLimitRange)).To(Succeed())

			// When
			_, err := handler.HandleResource(project)
			Expect(err).To(Succeed())

			// Then
			actualLimitRange := corev1.LimitRange{}
			Expect(client.Get(context.TODO(), types.NamespacedName{Name: RunaiLimitRangeName, Namespace: TestNamespace.Name}, &actualLimitRange)).To(Succeed())
			assertLimitRangesEqual(actualLimitRange, anotherNamespacedLimitRange)
		})
	})

})

func assertLimitRangesEqual(actualLimitRange corev1.LimitRange, namespacedLimitRange corev1.LimitRange) {
	Expect(actualLimitRange.Name).To(Equal(namespacedLimitRange.Name))
	Expect(actualLimitRange.Namespace).To(Equal(namespacedLimitRange.Namespace))
	Expect(actualLimitRange.Spec.Limits).To(HaveLen(1))
	Expect(actualLimitRange.Spec.Limits[0].Type).To(Equal(namespacedLimitRange.Spec.Limits[0].Type))
	Expect(reflect.DeepEqual(actualLimitRange.Spec.Limits[0].Max, namespacedLimitRange.Spec.Limits[0].Max)).To(BeTrue())
	Expect(reflect.DeepEqual(actualLimitRange.Spec.Limits[0].Default, namespacedLimitRange.Spec.Limits[0].Default)).To(BeTrue())
	Expect(reflect.DeepEqual(actualLimitRange.Spec.Limits[0].DefaultRequest, namespacedLimitRange.Spec.Limits[0].DefaultRequest)).To(BeTrue())
}
