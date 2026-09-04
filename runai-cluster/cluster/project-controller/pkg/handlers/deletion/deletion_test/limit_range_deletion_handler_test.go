/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deletion_test

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers/deletion"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("LimitRange Deletion Handler", func() {

	var (
		handler             LimitRangeDeletionHandler
		client              client.Client
		project             kaiv1alpha1.Project
		namespace           corev1.Namespace
		limitRange          corev1.LimitRange
		unrelatedLimitRange corev1.LimitRange
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		namespace = *TestNamespace.DeepCopy()
		limitRange = *NamespacedLimitRange.DeepCopy()
		unrelatedLimitRange = *NamespacedLimitRange.DeepCopy()
		unrelatedLimitRange.Name = "unrelated-limit-range"
		unrelatedLimitRange.OwnerReferences = UnrelatedProjectOwnerRef

		client = fake.NewClientBuilder().WithScheme(scheme).Build()
		handler = NewLimitRangeDeletionHandler(client)
	})

	It("Handles Project deletion", func() {

		// Given
		Expect(client.Create(context.TODO(), &project)).To(Succeed())
		Expect(client.Create(context.TODO(), &namespace)).To(Succeed())
		Expect(client.Create(context.TODO(), &limitRange)).To(Succeed())
		Expect(client.Create(context.TODO(), &unrelatedLimitRange)).To(Succeed())

		//When
		conditions, err := handler.OnDelete(&project)
		Expect(err).To(BeNil())

		// Then
		By("ProjectConditions are correct", func() {
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.ProjectConditionType("LimitRangeReady")))
			Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))
			Expect(conditions[0].Reason).To(Equal(""))
			Expect(conditions[0].Message).To(Equal(""))
		})

		By("Deleting the Project's OwnerReference from namespaced LimitRange", func() {
			actualLimitRange := getNamespacedLimitRange(limitRange.Name, limitRange.Namespace, client)
			Expect(actualLimitRange.Name).To(Equal(limitRange.Name))
			Expect(actualLimitRange.Namespace).To(Equal(limitRange.Namespace))
			Expect(common.IsProjectOwner(&project, &actualLimitRange)).To(BeNumerically("==", -1))
			Expect(actualLimitRange.OwnerReferences).To(HaveLen(0))
		})

		By("Not modifying namespaced LimitRange with unrelated OwnerRef", func() {
			actualUnrelatedLimitRange := getNamespacedLimitRange(unrelatedLimitRange.Name, unrelatedLimitRange.Namespace, client)
			Expect(common.IsProjectOwner(&project, &actualUnrelatedLimitRange)).To(BeNumerically("==", -1))
			Expect(actualUnrelatedLimitRange.OwnerReferences).To(HaveLen(1))
		})
	})
})

func getNamespacedLimitRange(limitRangeName string, limitRangeNamespace string, cli client.Client) (limitRange corev1.LimitRange) {
	_ = cli.Get(context.Background(),
		client.ObjectKey{
			Name:      limitRangeName,
			Namespace: limitRangeNamespace,
		}, &limitRange)
	return limitRange
}
