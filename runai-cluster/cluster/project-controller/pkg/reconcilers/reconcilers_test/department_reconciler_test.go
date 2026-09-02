/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcilers_test

import (
	"context"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/reconcilers"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Department Reconciler Tests", func() {

	var (
		reconciler *DepartmentReconciler
		k8sClient  client.Client
		department kaiv1alpha1.Department
		project    kaiv1alpha1.Project

		departmentReconcileRequest ctrl.Request
	)

	BeforeEach(func() {
		k8sClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&kaiv1alpha1.Department{}).
			WithIndex(&kaiv1alpha1.Project{}, ProjectDepartmentIndexField, IndexProjectByDepartment).
			Build()
		reconciler = NewDepartmentReconciler(k8sClient)

		department = *KaiTestDepartment1.DeepCopy()
		project = *TestProject.DeepCopy()
		// Sanity: the test project references the test department by name.
		Expect(project.Spec.Parent).To(Equal(department.Name))

		departmentReconcileRequest = ctrl.Request{
			NamespacedName: types.NamespacedName{Name: department.Name},
		}
	})

	It("Adds the controller to the department finalizers list", func() {
		// Given
		Expect(k8sClient.Create(context.TODO(), &department)).To(Succeed())

		// When
		result, err := reconciler.Reconcile(context.TODO(), departmentReconcileRequest)

		// Then
		Expect(result).To(BeZero())
		Expect(err).ToNot(HaveOccurred())

		actual := kaiv1alpha1.Department{}
		Expect(k8sClient.Get(context.TODO(), departmentReconcileRequest.NamespacedName, &actual)).To(Succeed())
		Expect(actual.Finalizers).To(ContainElement(config.DepartmentFinalizerName()))
	})

	When("A department is being deleted", func() {
		It("Blocks deletion while a project still references the department", func() {
			// Given - a department being deleted, with our finalizer set, and a project that references it
			deletedDepartment := department.DeepCopy()
			deletedDepartment.DeletionTimestamp = &now
			deletedDepartment.Finalizers = []string{config.DepartmentFinalizerName()}

			k8sClientWithDeleted := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&kaiv1alpha1.Department{}).
				WithIndex(&kaiv1alpha1.Project{}, ProjectDepartmentIndexField, IndexProjectByDepartment).
				WithObjects(deletedDepartment, &project).
				Build()
			reconcilerWithDeleted := NewDepartmentReconciler(k8sClientWithDeleted)

			// When
			result, err := reconcilerWithDeleted.Reconcile(context.TODO(), departmentReconcileRequest)

			// Then - mirrors the Project reconciler: an error is returned so the request
			// is requeued with backoff, and our finalizer is kept in place.
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeZero())

			By("Keeping the finalizer in place so the department is not removed + Setting a deletion-blocked condition on the department status", func() {
				actual := kaiv1alpha1.Department{}
				Expect(k8sClientWithDeleted.Get(context.TODO(),
					departmentReconcileRequest.NamespacedName, &actual)).To(Succeed())
				Expect(actual.Finalizers).To(ContainElement(config.DepartmentFinalizerName()))

				var blockedCondition *kaiv1alpha1.DepartmentCondition
				for i := range actual.Status.Conditions {
					if actual.Status.Conditions[i].Type == kaiv1alpha1.DepartmentDeletionBlocked {
						blockedCondition = &actual.Status.Conditions[i]
						break
					}
				}
				Expect(blockedCondition).ToNot(BeNil())
				Expect(blockedCondition.Status).To(Equal(corev1.ConditionTrue))
				Expect(blockedCondition.Reason).To(Equal("ProjectsStillAssigned"))
				Expect(blockedCondition.Message).ToNot(BeEmpty())
			})
		})

		It("Removes the finalizer when no project references the department", func() {
			// Given - a department being deleted, with our finalizer plus another finalizer so the
			// fake client keeps the object around after we remove ours, and no project referencing it.
			deletedDepartment := department.DeepCopy()
			deletedDepartment.DeletionTimestamp = &now
			deletedDepartment.Finalizers = []string{config.DepartmentFinalizerName(), "other-finalizer"}

			k8sClientWithDeleted := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&kaiv1alpha1.Department{}).
				WithIndex(&kaiv1alpha1.Project{}, ProjectDepartmentIndexField, IndexProjectByDepartment).
				WithObjects(deletedDepartment).
				Build()
			reconcilerWithDeleted := NewDepartmentReconciler(k8sClientWithDeleted)

			// When
			result, err := reconcilerWithDeleted.Reconcile(context.TODO(), departmentReconcileRequest)

			// Then
			Expect(result).To(BeZero())
			Expect(err).ToNot(HaveOccurred())

			By("Removing our finalizer from the department", func() {
				actual := kaiv1alpha1.Department{}
				Expect(k8sClientWithDeleted.Get(context.TODO(),
					departmentReconcileRequest.NamespacedName, &actual)).To(Succeed())
				Expect(actual.Finalizers).ToNot(ContainElement(config.DepartmentFinalizerName()))
				Expect(actual.Finalizers).To(ContainElement("other-finalizer"))

				Expect(actual.Status.Conditions).To(BeEmpty())
			})
		})

		It("Deletes the department once our finalizer is the last one removed", func() {
			// Given - a department being deleted whose only finalizer is ours, and no
			// project referencing it. Removing our finalizer should let the fake client
			// actually delete the object.
			deletedDepartment := department.DeepCopy()
			deletedDepartment.DeletionTimestamp = &now
			deletedDepartment.Finalizers = []string{config.DepartmentFinalizerName()}

			k8sClientWithDeleted := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&kaiv1alpha1.Department{}).
				WithIndex(&kaiv1alpha1.Project{}, ProjectDepartmentIndexField, IndexProjectByDepartment).
				WithObjects(deletedDepartment).
				Build()
			reconcilerWithDeleted := NewDepartmentReconciler(k8sClientWithDeleted)

			// When
			result, err := reconcilerWithDeleted.Reconcile(context.TODO(), departmentReconcileRequest)

			// Then
			Expect(result).To(BeZero())
			Expect(err).ToNot(HaveOccurred())

			By("Removing the department from the cluster", func() {
				actual := kaiv1alpha1.Department{}
				getErr := k8sClientWithDeleted.Get(context.TODO(),
					departmentReconcileRequest.NamespacedName, &actual)
				Expect(apierrors.IsNotFound(getErr)).To(BeTrue())
			})
		})

		It("Does not block deletion for projects that reference a different department", func() {
			// Given - a project assigned to a different department
			otherProject := project.DeepCopy()
			otherProject.Spec.Parent = "some-other-department"

			deletedDepartment := department.DeepCopy()
			deletedDepartment.DeletionTimestamp = &now
			deletedDepartment.Finalizers = []string{config.DepartmentFinalizerName(), "other-finalizer"}

			k8sClientWithDeleted := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&kaiv1alpha1.Department{}).
				WithIndex(&kaiv1alpha1.Project{}, ProjectDepartmentIndexField, IndexProjectByDepartment).
				WithObjects(deletedDepartment, otherProject).
				Build()
			reconcilerWithDeleted := NewDepartmentReconciler(k8sClientWithDeleted)

			// When
			result, err := reconcilerWithDeleted.Reconcile(context.TODO(), departmentReconcileRequest)

			// Then
			Expect(result).To(BeZero())
			Expect(err).ToNot(HaveOccurred())

			actual := kaiv1alpha1.Department{}
			Expect(k8sClientWithDeleted.Get(context.TODO(),
				departmentReconcileRequest.NamespacedName, &actual)).To(Succeed())
			Expect(actual.Finalizers).ToNot(ContainElement(config.DepartmentFinalizerName()))
		})
	})
})
