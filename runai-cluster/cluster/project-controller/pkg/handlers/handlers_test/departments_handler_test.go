/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Department Handler Tests", func() {

	var (
		handler           handlers.DepartmentHandler
		k8sClient         client.Client
		dep1, dep2        kaiv1alpha1.Department
		queue1d, queuenp1 kaiv2.Queue
	)

	BeforeEach(func() {
		dep1 = *test.KaiTestDepartment1.DeepCopy()
		dep2 = *test.KaiTestDepartment2.DeepCopy()
		queue1d = *test.KaiTestQueueDep1NpD.DeepCopy()
		queuenp1 = *test.KaiTestQueueDep1Np1.DeepCopy()

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		handler = handlers.NewDepartmentHandler(k8sClient)
	})

	Describe("Department Handler Tests - validate correct creation of queues", func() {
		It("Creating new queues matching the Department queues spec", func() {

			// Given
			beforeCreation, beforeCreationErr := getQueue(k8sClient,
				handlers.GetQueueName(dep2.Spec.Queues[0], dep2.Name))
			Expect(beforeCreationErr).To(HaveOccurred())
			Expect(beforeCreation).To(BeZero())
			beforeCreation, beforeCreationErr = getQueue(k8sClient,
				handlers.GetQueueName(dep2.Spec.Queues[1], dep2.Name))
			Expect(beforeCreationErr).To(HaveOccurred())
			Expect(beforeCreation).To(BeZero())

			// When
			err := handler.Handle(context.Background(), &dep2)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToDepartmentSpec(k8sClient, &dep2, test.TestDepartment2Id)
		})

		It("1 queue spec is identical, 1 is not - not modifying/modifying accordingly", func() {
			Expect(k8sClient.Create(context.TODO(), &dep1)).To(Succeed())
			Expect(k8sClient.Create(context.TODO(), &queue1d)).To(Succeed())
			Expect(k8sClient.Create(context.TODO(), &queuenp1)).To(Succeed())

			// The fake client strips TypeMeta on Create; restore it so the handler
			// builds the queue owner reference with the department's GVK, as it would
			// with a real client.
			populateGVK(&dep1)

			// When
			err := handler.Handle(context.Background(), &dep1)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToDepartmentSpec(k8sClient, &dep1, test.TestDepartment1Id)

			// original queue for dep1-np1 is equal to actual - was already equal before
			actual, err := getQueue(k8sClient,
				handlers.GetQueueName(dep1.Spec.Queues[1], dep1.Name))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeZero())
			expectQueuesEqualV2(actual, queuenp1)

			// original queue for dep1-default-np - not equal to actual - the handler reconciled it
			actual, err = getQueue(k8sClient,
				handlers.GetQueueName(dep1.Spec.Queues[0], dep1.Name))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeZero())
			Expect(queue1d.Spec.Resources).ToNot(Equal(actual.Spec.Resources))
			Expect(queue1d.Labels).ToNot(Equal(actual.Labels))
		})

		It("Not modifying the queues when the department name changes (only owner ref name)", func() {
			err := handler.Handle(context.Background(), &dep2)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToDepartmentSpec(k8sClient, &dep2, test.TestDepartment2Id)

			// "rename" department - meaning create the new one and delete the old one
			renamedDep := test.KaiTestDepartment2.DeepCopy()
			renamedDep.Name = "different-name"

			// validate spec as expected for new department name after Handle
			err = handler.Handle(context.Background(), renamedDep)
			Expect(err).Should(Succeed())
			assertQueuesIdenticalToDepartmentSpec(k8sClient, renamedDep, test.TestDepartment2Id)

			actual, err := getQueue(k8sClient,
				handlers.GetQueueName(renamedDep.Spec.Queues[1], renamedDep.Name))
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeZero())
			Expect(actual.OwnerReferences).ToNot(BeZero())
			Expect(len(actual.OwnerReferences)).To(Equal(1))
			Expect(actual.OwnerReferences[0].Name).To(Equal(renamedDep.Name))
		})

		It("Queue with suggested name already exists for other resource - generating different name", func() {
			queueWithNameOfSuggestedQueueNameForDep2 := test.KaiTestQueueDep1NpD.DeepCopy()
			queueWithNameOfSuggestedQueueNameForDep2.Name = handlers.GetQueueName(
				dep2.Spec.Queues[0], dep2.Name)
			err := k8sClient.Create(context.Background(), queueWithNameOfSuggestedQueueNameForDep2)
			Expect(err).Should(Succeed())

			queueWithNameOfSuggestedQueueNameForDep2 = test.KaiTestQueueDep1NpD.DeepCopy()
			queueWithNameOfSuggestedQueueNameForDep2.Name = handlers.GetQueueName(
				dep2.Spec.Queues[1], dep2.Name)
			err = k8sClient.Create(context.Background(), queueWithNameOfSuggestedQueueNameForDep2)
			Expect(err).Should(Succeed())

			// When
			err = handler.Handle(context.Background(), &dep2)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToDepartmentSpecInner(k8sClient, &dep2,
				test.TestDepartment2Id, false)

			queue, err := getQueueForDepartmentByLabel(k8sClient, test.TestDepartment2Name, test.DefaultNodePoolName)
			Expect(err).Should(Succeed())
			createdDepartmentQueueNameDefaultNodepool := queue.Name
			queue, err = getQueueForDepartmentByLabel(k8sClient, test.TestDepartment2Name, test.SomeNodePoolName)
			Expect(err).Should(Succeed())
			createdDepartmentQueueNameSomeNodepool := queue.Name

			err = handler.Handle(context.Background(), &dep2)
			Expect(err).Should(Succeed())
			err = handler.Handle(context.Background(), &dep2)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToDepartmentSpecInner(k8sClient, &dep2,
				test.TestDepartment2Id, false)

			queue, err = getQueueForDepartmentByLabel(k8sClient, test.TestDepartment2Name, test.DefaultNodePoolName)
			Expect(err).Should(Succeed())
			Expect(queue.Name).To(Equal(createdDepartmentQueueNameDefaultNodepool))
			queue, err = getQueueForDepartmentByLabel(k8sClient, test.TestDepartment2Name, test.SomeNodePoolName)
			Expect(err).Should(Succeed())
			Expect(queue.Name).To(Equal(createdDepartmentQueueNameSomeNodepool))
		})

		It("Validate department queue too long name gets trimmed with random suffix", func() {
			// Given - clear the queue names so the "<departmentName>-<nodepool>" fallback applies;
			// combined with the long nodepool names this yields an over-long queue name that must
			// be trimmed and given a random suffix.
			dep3 := test.KaiTestDepartment3.DeepCopy()
			dep3.Spec.Queues[0].Name = ""
			dep3.Spec.Queues[1].Name = ""
			dep3.Spec.Queues[0].Nodepool = "the-longest-possible-nodepool-name-that-is-exactly-63-charslong"
			dep3.Spec.Queues[1].Nodepool = "the-longest-possible-nodepool-name-that-is-exactly-63-charslon2"

			// When
			err := handler.Handle(context.Background(), dep3)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToDepartmentSpecInner(k8sClient, dep3, test.TestDepartment3Id, false)

			queue0, err := getQueueForDepartmentByLabel(k8sClient, test.TestDepartment3Name, dep3.Spec.Queues[0].Nodepool)
			Expect(err).Should(Succeed())
			Expect(queue0.Name).To(HavePrefix(test.TestDepartment3Name + "-the-longest-possible-nodepool-name-that"))
			queue0Name := queue0.Name
			queue1, err := getQueueForDepartmentByLabel(k8sClient, test.TestDepartment3Name, dep3.Spec.Queues[1].Nodepool)
			Expect(err).Should(Succeed())
			Expect(queue1.Name).To(HavePrefix(test.TestDepartment3Name + "-the-longest-possible-nodepool-name-that"))
			queue1Name := queue1.Name

			// Handle again to ensure no changes
			err = handler.Handle(context.Background(), dep3)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToDepartmentSpecInner(k8sClient, dep3, test.TestDepartment3Id, false)

			queue0, err = getQueueForDepartmentByLabel(k8sClient, test.TestDepartment3Name, dep3.Spec.Queues[0].Nodepool)
			Expect(err).Should(Succeed())
			Expect(queue0.Name).To(Equal(queue0Name))
			queue1, err = getQueueForDepartmentByLabel(k8sClient, test.TestDepartment3Name, dep3.Spec.Queues[1].Nodepool)
			Expect(err).Should(Succeed())
			Expect(queue1.Name).To(Equal(queue1Name))
		})
	})

	Describe("Department Handler Tests - validate correct deletion of queues", func() {
		It("Delete queue if nodepool deleted - removed from department spec", func() {
			// When
			err := handler.Handle(context.Background(), &dep2)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToDepartmentSpec(k8sClient, &dep2, test.TestDepartment2Id)
			npQueueName := handlers.GetQueueName(dep2.Spec.Queues[1], dep2.Name)
			actual, err := getQueue(k8sClient, npQueueName)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeZero())

			// "delete" a nodepool - meaning delete it from the department spec
			modifiedDep := test.KaiTestDepartment2.DeepCopy()
			modifiedDep.Spec.Queues = []kaiv1alpha1.QueueConfig{modifiedDep.Spec.Queues[0]}

			// validate spec as expected after Handle
			err = handler.Handle(context.Background(), modifiedDep)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToDepartmentSpec(k8sClient, modifiedDep, test.TestDepartment2Id)
			actual, err = getQueue(k8sClient, npQueueName)
			Expect(err).To(HaveOccurred())
			Expect(actual).To(BeZero())
		})
	})

	It("Fails to Handle queues", func() {
		badScheme := runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(badScheme)).Should(Succeed())
		Expect(kaiv1alpha1.AddToScheme(badScheme)).Should(Succeed())
		// not adding kaiv2 to scheme - so queue handler should fail
		badK8sClient := fake.NewClientBuilder().WithScheme(badScheme).Build()
		badHandler := handlers.NewDepartmentHandler(badK8sClient)

		Expect(badK8sClient.Create(context.TODO(), &dep1)).To(Succeed())

		err := badHandler.Handle(context.TODO(), &dep1)
		Expect(err).To(HaveOccurred())
	})
})

func assertQueuesIdenticalToDepartmentSpec(k8sClient client.Client,
	department *kaiv1alpha1.Department, departmentId string) {
	assertQueuesIdenticalToDepartmentSpecInner(k8sClient, department, departmentId, true)
}

func assertQueuesIdenticalToDepartmentSpecInner(k8sClient client.Client,
	department *kaiv1alpha1.Department, departmentId string, queueNameEqualToSuggested bool) {
	for _, depQueueSpec := range department.Spec.Queues {
		var queue kaiv2.Queue
		var err error

		if queueNameEqualToSuggested {
			queue, err = getQueue(k8sClient, handlers.GetQueueName(depQueueSpec, department.Name))
			Expect(err).ToNot(HaveOccurred())
			Expect(queue).ToNot(BeZero())

			Expect(queue.Name).To(Equal(handlers.GetQueueName(depQueueSpec, department.Name)))
		} else {
			queue, err = getQueueForDepartmentByLabel(k8sClient, department.Name, depQueueSpec.Nodepool)
			Expect(err).ToNot(HaveOccurred())
			Expect(queue).ToNot(BeZero())

			Expect(queue.Name).ToNot(Equal(handlers.GetQueueName(depQueueSpec, department.Name)))

			prefixToExpect := handlers.GetQueueName(depQueueSpec, department.Name)
			if len(prefixToExpect) > 58 {
				prefixToExpect = prefixToExpect[:58]
			}
			Expect(queue.Name).To(HavePrefix(prefixToExpect))
		}

		Expect(queue.Spec.ParentQueue).To(Equal(""))
		if depQueueSpec.Priority == nil {
			Expect(queue.Spec.Priority).To(Equal(ptr.To(handlers.DefaultQueuePriority)))
		} else {
			Expect(queue.Spec.Priority).To(Equal(ptr.To(int(*depQueueSpec.Priority))))
		}

		Expect(queue.Labels[test.QueueDepartmentNameLabel]).To(Equal(department.Name))
		if depQueueSpec.Nodepool != test.DefaultNodePoolName {
			Expect(queue.Labels[test.NodePoolLabelKey]).To(Equal(depQueueSpec.Nodepool))
		} else {
			Expect(queue.Labels).ToNot(HaveKey(test.NodePoolLabelKey))
		}

		Expect(queue.OwnerReferences).ToNot(BeZero())
		Expect(len(queue.OwnerReferences)).To(Equal(1))
		Expect(queue.OwnerReferences[0].APIVersion).To(Equal(department.APIVersion))
		Expect(queue.OwnerReferences[0].Kind).To(Equal(department.Kind))
		Expect(queue.OwnerReferences[0].Name).To(Equal(department.Name))
		Expect(queue.OwnerReferences[0].UID).To(Equal(department.UID))
		Expect(queue.OwnerReferences[0].Controller).To(Equal(&common.TrueRef))

		assertQueueResourcesIdenticalToDepartmentSpecResources(&queue, depQueueSpec)
	}
}

func assertQueueResourcesIdenticalToDepartmentSpecResources(
	queue *kaiv2.Queue, depQueueSpec kaiv1alpha1.QueueConfig) {
	Expect(queue.Spec.Resources.GPU.Limit).To(Equal(depQueueSpec.Resources.GPU.Limit))
	Expect(queue.Spec.Resources.GPU.OverQuotaWeight).To(Equal(depQueueSpec.Resources.GPU.OverQuotaWeight))
	Expect(queue.Spec.Resources.GPU.Quota).To(Equal(depQueueSpec.Resources.GPU.Deserved))

	Expect(queue.Spec.Resources.CPU.Limit).To(Equal(depQueueSpec.Resources.CPU.Limit))
	Expect(queue.Spec.Resources.CPU.OverQuotaWeight).To(Equal(depQueueSpec.Resources.CPU.OverQuotaWeight))
	Expect(queue.Spec.Resources.CPU.Quota).To(Equal(depQueueSpec.Resources.CPU.Deserved))

	Expect(queue.Spec.Resources.Memory.Limit).To(Equal(depQueueSpec.Resources.Memory.Limit))
	Expect(queue.Spec.Resources.Memory.OverQuotaWeight).To(Equal(depQueueSpec.Resources.Memory.OverQuotaWeight))
	Expect(queue.Spec.Resources.Memory.Quota).To(Equal(depQueueSpec.Resources.Memory.Deserved))
}

func expectQueuesEqualV2(left, right kaiv2.Queue) {
	var transformMap = func(m map[string]string) map[string]string {
		if m == nil {
			return map[string]string{}
		}
		return m
	}
	left.Labels = transformMap(left.Labels)
	right.Labels = transformMap(right.Labels)

	Expect(left.Name).To(Equal(right.Name))
	Expect(left.Labels).To(Equal(right.Labels))
	Expect(left.APIVersion).To(Equal(right.APIVersion))
	Expect(left.OwnerReferences).To(Equal(right.OwnerReferences))
	Expect(left.Spec).To(Equal(right.Spec))
}

func getQueue(k8sClient client.Client, queueName string) (kaiv2.Queue, error) {
	queue := kaiv2.Queue{}
	err := k8sClient.Get(context.Background(),
		client.ObjectKey{Name: queueName},
		&queue,
	)
	return queue, err
}

func getQueueForDepartmentByLabel(k8sClient client.Client, departmentName, nodepoolName string) (kaiv2.Queue, error) {
	queues := &kaiv2.QueueList{}
	err := k8sClient.List(context.Background(), queues,
		client.MatchingLabels(map[string]string{test.QueueDepartmentNameLabel: departmentName}))
	if err != nil {
		return kaiv2.Queue{}, err
	}

	for _, queue := range queues.Items {
		// The department-name label is shared with project queues; only consider queues owned
		// by a Department.
		if !isOwnedByDepartment(queue.OwnerReferences) {
			continue
		}

		nodePoolLabelValue, found := queue.Labels[test.NodePoolLabelKey]

		if nodepoolName == test.DefaultNodePoolName && !found {
			return queue, nil
		}

		if nodepoolName != test.DefaultNodePoolName && found && nodePoolLabelValue == nodepoolName {
			return queue, nil
		}
	}

	return kaiv2.Queue{}, fmt.Errorf("couldn't find queue for department %s and nodepool %s",
		departmentName, nodepoolName)
}

// isOwnedByDepartment reports whether the queue is owned by a Department (vs a Project), matching
// how the handlers distinguish department queues from project queues that share the department
// name label.
func isOwnedByDepartment(ownerRefs []metav1.OwnerReference) bool {
	for _, ref := range ownerRefs {
		if ref.Kind == common.DepartmentKind {
			return true
		}
	}
	return false
}
