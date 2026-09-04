/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers_test

import (
	"context"
	"fmt"

	"k8s.io/utils/ptr"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Queue Resource Handler", func() {

	var (
		handler                  QueueResourceHandler
		departmentHandler        DepartmentHandler
		k8sClient                client.Client
		project, modifiedProject kaiv1alpha1.Project
		queue                    kaiv2.Queue
		dep1, dep2               kaiv1alpha1.Department
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		modifiedProject = *KaiModifiedTestProject.DeepCopy()
		queue = *TestQueue.DeepCopy()
		dep1 = *KaiTestDepartment1.DeepCopy()
		dep2 = *KaiTestDepartment2.DeepCopy()

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		handler = NewQueueResourceHandler(k8sClient)
		departmentHandler = NewDepartmentHandler(k8sClient)

		// will create the parent queues that project handler is expecting
		err := departmentHandler.Handle(context.Background(), &dep1)
		Expect(err).Should(Succeed())
		err = departmentHandler.Handle(context.Background(), &dep2)
		Expect(err).Should(Succeed())

		// persist the department objects so the project handler can look up the parent queue
		// by fetching the owning department (parent resolution is keyed on project.Spec.Parent).
		Expect(k8sClient.Create(context.Background(), &dep1)).To(Succeed())
		Expect(k8sClient.Create(context.Background(), &dep2)).To(Succeed())
	})

	It("Handles a new, non-existing Queue", func() {
		By("Creating a new Queue matching the Project Queue spec", func() {

			// Given
			beforeCreation, beforeCreationErr := getQueue(k8sClient, project.Name)
			Expect(beforeCreationErr).To(HaveOccurred())
			Expect(beforeCreation).To(BeZero())

			// When
			conditions, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			// Then
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.QueuesReady))
			Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))
			Expect(conditions[0].Reason).To(Equal(""))
			Expect(conditions[0].Message).To(Equal(""))

			assertQueuesIdenticalToProjectSpec(k8sClient, &project, TestDepartment1Id)
		})
	})

	It("Handles multiples node groups for project", func() {
		By("Creating a queues for each node group", func() {
			// Given
			beforeCreation, beforeCreationErr := getQueue(k8sClient, project.Name)
			Expect(beforeCreationErr).To(HaveOccurred())
			Expect(beforeCreation).To(BeZero())

			// When
			_, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToProjectSpec(k8sClient, &project, TestDepartment1Id)
		})
	})

	It("Handles an existing Queue", func() {
		By("Not modifying the Queue when the incoming project Queue spec is identical", func() {

			// Given

			// Creating resources via the fake client causes it to put them in its own object
			// store so it can return them when queried
			Expect(k8sClient.Create(context.TODO(), &project)).To(Succeed())

			// When
			_, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToProjectSpec(k8sClient, &project, TestDepartment1Id)
		})
	})

	It("Handles an existing Queue", func() {
		By("Modifying the Queue when the incoming project Queue spec has differences", func() {

			// Given
			// Creating resources via the fake client causes it to put them in its own object
			// store so it can return them when queried
			Expect(k8sClient.Create(context.TODO(), &project)).To(Succeed())
			Expect(k8sClient.Create(context.TODO(), &queue)).To(Succeed())

			// When
			_, err := handler.HandleResource(modifiedProject)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToProjectSpec(k8sClient, &modifiedProject, TestDepartment1Id)
		})

		By("Validate correct parent queue when project assigned to different department", func() {
			// When
			_, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToProjectSpec(k8sClient, &project, TestDepartment1Id)

			// assign the project to a different department
			modifiedProj := TestProject.DeepCopy()
			modifiedProj.Spec.Parent = TestDepartment2Name

			// validate spec as expected after Handle
			_, err = handler.HandleResource(*modifiedProj)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToProjectSpec(k8sClient, modifiedProj, TestDepartment2Id)
		})
	})

	It("Validate naming of queues", func() {
		By("Create another project with name identical to proj-nodepool different queue name", func() {
			// When
			_, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToProjectSpec(k8sClient, &project, TestDepartment1Id)
			actual, err := getQueue(k8sClient, project.Spec.Queues[1].Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeZero())

			// create another project with the name being the queue name of the first proj
			modifiedProj := TestProject.DeepCopy()
			modifiedProj.Name = project.Spec.Queues[1].Name
			modifiedProj.UID = "2"
			modifiedProj.Spec.Queues[0].Name = modifiedProj.Name
			modifiedProj.Spec.Queues[1].Name = fmt.Sprintf("%s-%s", modifiedProj.Name, SomeNodePoolName)

			// validate spec as expected after Handle
			_, err = handler.HandleResource(*modifiedProj)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToProjectSpecInner(k8sClient, modifiedProj, TestDepartment1Id,
				false, true)
			actual, err = getQueue(k8sClient, modifiedProj.Spec.Queues[0].Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.Labels[config.Get().ProjectLabelKey]).ToNot(Equal(modifiedProj.Name))
			Expect(actual.Labels[config.Get().ProjectIdLabelKey]).ToNot(Equal(string(modifiedProj.UID)))

			actualQueue, err := getQueueForProjectByLabel(k8sClient, project.Name, modifiedProj.Spec.Queues[0].Nodepool)
			Expect(err).Should(Succeed())
			createdProjectQueueNameDefaultNodepool := actualQueue.Name
			actualQueue, err = getQueueForProjectByLabel(k8sClient, project.Name, modifiedProj.Spec.Queues[1].Nodepool)
			Expect(err).Should(Succeed())
			createdProjectQueueNameSomeNodepool := actualQueue.Name

			_, err = handler.HandleResource(*modifiedProj)
			Expect(err).Should(Succeed())
			_, err = handler.HandleResource(*modifiedProj)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToProjectSpecInner(k8sClient, modifiedProj, TestDepartment1Id,
				false, true)

			actualQueue, err = getQueueForProjectByLabel(k8sClient, project.Name, modifiedProj.Spec.Queues[0].Nodepool)
			Expect(err).Should(Succeed())
			Expect(actualQueue.Name).To(Equal(createdProjectQueueNameDefaultNodepool))
			actualQueue, err = getQueueForProjectByLabel(k8sClient, project.Name, modifiedProj.Spec.Queues[1].Nodepool)
			Expect(err).Should(Succeed())
			Expect(actualQueue.Name).To(Equal(createdProjectQueueNameSomeNodepool))
		})
	})

	It("Validate naming of queues", func() {
		By("validate Parent queue name correct when department's queue has different name", func() {
			// create a project whose queues are named the same as dep3's future queues, so the
			// department's queues must be created with a random suffix (name collision).
			dep3DefaultQueueName := GetQueueName(KaiTestDepartment3.Spec.Queues[0], TestDepartment3Name)
			dep3SomeQueueName := GetQueueName(KaiTestDepartment3.Spec.Queues[1], TestDepartment3Name)

			projectNamedAsDepartmentsQueue := TestProject.DeepCopy()
			projectNamedAsDepartmentsQueue.Name = "project-colliding-with-dep3-queue"
			projectNamedAsDepartmentsQueue.UID = "56"
			projectNamedAsDepartmentsQueue.Spec.Queues[0].Name = dep3DefaultQueueName
			projectNamedAsDepartmentsQueue.Spec.Queues[1].Name = dep3SomeQueueName

			// When
			_, err := handler.HandleResource(*projectNamedAsDepartmentsQueue)
			Expect(err).Should(Succeed())

			// create the department (queues + persisted object for parent lookup)
			newDep := KaiTestDepartment3.DeepCopy()
			err = departmentHandler.Handle(context.Background(), newDep)
			Expect(err).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), newDep)).To(Succeed())

			// the queues names should be different from the suggested name - but have its prefix
			actualQueue, err := getQueueForDepartmentByLabel(k8sClient, TestDepartment3Name, DefaultNodePoolName)
			Expect(err).Should(Succeed())
			createdDep3QueueNameDefaultNodePool := actualQueue.Name
			Expect(createdDep3QueueNameDefaultNodePool).ToNot(Equal(dep3DefaultQueueName))
			Expect(createdDep3QueueNameDefaultNodePool).To(HavePrefix(dep3DefaultQueueName))
			actualQueue, err = getQueueForDepartmentByLabel(k8sClient, TestDepartment3Name, SomeNodePoolName)
			Expect(err).Should(Succeed())
			createdDep3QueueNameSomeNodePool := actualQueue.Name
			Expect(createdDep3QueueNameSomeNodePool).ToNot(Equal(dep3SomeQueueName))
			Expect(createdDep3QueueNameSomeNodePool).To(HavePrefix(dep3SomeQueueName))

			// create a project in dep3
			projInDep3 := TestProject.DeepCopy()
			projInDep3.Spec.Parent = TestDepartment3Name

			_, err = handler.HandleResource(*projInDep3)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToProjectSpecInner(k8sClient, projInDep3, TestDepartment3Id,
				true, false)

			// validate the parent queue points to the correct queue (that has different name from the "default" name)
			projQueue, err := getQueue(k8sClient, projInDep3.Spec.Queues[0].Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(projQueue.Spec.ParentQueue).To(Equal(createdDep3QueueNameDefaultNodePool))
			projQueue, err = getQueue(k8sClient, projInDep3.Spec.Queues[1].Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(projQueue.Spec.ParentQueue).To(Equal(createdDep3QueueNameSomeNodePool))
		})
	})

	It("Validate naming of queues", func() {
		By("validate project queue too long name gets trimmed with random suffix", func() {
			// create a project with a queue name that is too long
			projectWithLongQueueName := TestProject.DeepCopy()
			longQueueName := "a-very-long-project-name-that-exceeds-the-allowed-length-for-queue-names"
			projectWithLongQueueName.Name = "test10-with-long-queue-names"
			projectWithLongQueueName.UID = "78"
			projectWithLongQueueName.Spec.Queues[0].Name = longQueueName
			projectWithLongQueueName.Spec.Queues[1].Name = fmt.Sprintf("%s-%s",
				longQueueName, SomeNodePoolName)

			// When
			_, err := handler.HandleResource(*projectWithLongQueueName)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToProjectSpecInner(k8sClient, projectWithLongQueueName, TestDepartment1Id,
				false, true)

			actualQueue, err := getQueueForProjectByLabel(k8sClient, projectWithLongQueueName.Name,
				DefaultNodePoolName)
			Expect(err).Should(Succeed())
			Expect(actualQueue.Name).ToNot(HavePrefix(projectWithLongQueueName.Spec.Queues[0].Name))
			Expect(actualQueue.Name).To(HavePrefix(projectWithLongQueueName.Spec.Queues[0].Name[:55]))
			queue0Name := actualQueue.Name
			actualQueue, err = getQueueForProjectByLabel(k8sClient, projectWithLongQueueName.Name,
				SomeNodePoolName)
			Expect(err).Should(Succeed())
			Expect(actualQueue.Name).ToNot(HavePrefix(projectWithLongQueueName.Spec.Queues[1].Name))
			Expect(actualQueue.Name).To(HavePrefix(projectWithLongQueueName.Spec.Queues[1].Name[:55]))
			queue1Name := actualQueue.Name

			// Handle resource again to ensure the names remain the same
			_, err = handler.HandleResource(*projectWithLongQueueName)
			Expect(err).Should(Succeed())

			actualQueue, err = getQueueForProjectByLabel(k8sClient, projectWithLongQueueName.Name,
				DefaultNodePoolName)
			Expect(err).Should(Succeed())
			Expect(actualQueue.Name).To(Equal(queue0Name))
			actualQueue, err = getQueueForProjectByLabel(k8sClient, projectWithLongQueueName.Name,
				SomeNodePoolName)
			Expect(err).Should(Succeed())
			Expect(actualQueue.Name).To(Equal(queue1Name))
		})
	})

	It("Validate correct deletion of queues", func() {
		By("Delete queue if nodepool deleted - removed from project spec", func() {
			// When
			_, err := handler.HandleResource(project)
			Expect(err).Should(Succeed())

			// Then
			assertQueuesIdenticalToProjectSpec(k8sClient, &project, TestDepartment1Id)
			actual, err := getQueue(k8sClient, project.Spec.Queues[1].Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeZero())

			// "delete" a nodepool - meaning delete it from the department spec
			modifiedProj := TestProject.DeepCopy()
			modifiedProj.Spec.Queues = []kaiv1alpha1.QueueConfig{modifiedProj.Spec.Queues[0]}

			// validate spec as expected after Handle
			_, err = handler.HandleResource(*modifiedProj)
			Expect(err).Should(Succeed())

			assertQueuesIdenticalToProjectSpec(k8sClient, modifiedProj, TestDepartment1Id)
			actual, err = getQueue(k8sClient, project.Spec.Queues[1].Name)
			Expect(err).To(HaveOccurred())
			Expect(actual).To(BeZero())
		})
	})

	Specify("Owner Reference Creation", func() {
		// When
		_, _ = handler.HandleResource(project)

		// Then
		currentQueue, err := getQueue(k8sClient, project.Spec.Queues[0].Name)
		Expect(err).ToNot(HaveOccurred())
		Expect(currentQueue.OwnerReferences).ToNot(BeZero())
		Expect(len(currentQueue.OwnerReferences)).To(Equal(1))
		Expect(currentQueue.OwnerReferences[0].APIVersion).To(Equal(project.APIVersion))
		Expect(currentQueue.OwnerReferences[0].Kind).To(Equal(project.Kind))
		Expect(currentQueue.OwnerReferences[0].Name).To(Equal(project.Name))
		Expect(currentQueue.OwnerReferences[0].UID).To(Equal(project.UID))
		Expect(currentQueue.OwnerReferences[0].Controller).To(Equal(&common.TrueRef))
	})

	It("Handles a project without a parent department", func() {
		By("Creating the queues without a parent queue and not failing", func() {
			// Given - a project with no parent, so no parent department queue is expected
			projNoDep := TestProject.DeepCopy()
			projNoDep.Spec.Parent = ""

			// When
			conditions, err := handler.HandleResource(*projNoDep)
			Expect(err).Should(Succeed())

			// Then
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))

			for _, projQueueSpec := range projNoDep.Spec.Queues {
				actualQueue, err := getQueueForProjectByLabel(k8sClient, projNoDep.Name, projQueueSpec.Nodepool)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualQueue.Spec.ParentQueue).To(BeEmpty())
			}
		})
	})

	It("Handles a project whose parent department does not exist", func() {
		By("Creating the queues without a parent queue and not failing", func() {
			// Given - a project pointing to a department that does not exist in the cluster
			projUnknownDep := TestProject.DeepCopy()
			projUnknownDep.Spec.Parent = "non-existing-department"

			// When
			conditions, err := handler.HandleResource(*projUnknownDep)
			Expect(err).Should(Succeed())

			// Then
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0].Status).To(Equal(corev1.ConditionTrue))

			for _, projQueueSpec := range projUnknownDep.Spec.Queues {
				actualQueue, err := getQueueForProjectByLabel(k8sClient, projUnknownDep.Name, projQueueSpec.Nodepool)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualQueue.Spec.ParentQueue).To(BeEmpty())
			}
		})
	})

	It("Fails to Handle Queues", func() {
		badScheme := runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(badScheme)).Should(Succeed())
		Expect(kaiv1alpha1.AddToScheme(badScheme)).Should(Succeed())
		// not adding schedulingv2 to scheme - so queue handler should fail
		badK8sClient := fake.NewClientBuilder().WithScheme(badScheme).Build()
		badHandler := NewQueueResourceHandler(badK8sClient)

		Expect(badK8sClient.Create(context.TODO(), &project)).To(Succeed())

		conditions, err := badHandler.HandleResource(project)
		Expect(err).To(HaveOccurred())

		Expect(conditions).To(HaveLen(1))
		Expect(conditions[0]).ToNot(BeNil())
		Expect(conditions[0].Type).To(Equal(kaiv1alpha1.QueuesReady))
		Expect(conditions[0].Status).To(Equal(corev1.ConditionFalse))
		Expect(conditions[0].Reason).To(Equal(string(QueuesHandlerFailed)))
		Expect(conditions[0].Message).ToNot(Equal(""))
	})
})

func assertQueuesIdenticalToProjectSpec(k8sclient client.Client,
	project *kaiv1alpha1.Project, departmentId string) {
	assertQueuesIdenticalToProjectSpecInner(k8sclient, project, departmentId,
		true, true)
}

func assertQueuesIdenticalToProjectSpecInner(k8sclient client.Client,
	project *kaiv1alpha1.Project, departmentId string, queueNameEqualToSuggested, parentQueueNameEqualToSuggested bool) {
	// The fake client strips TypeMeta when the project is created/read; restore it
	// so owner-reference assertions against the project's GVK hold, as with a real client.
	populateGVK(project)
	allQueueNamesWithoutPrefix := true
	allParentQueueNamesWithoutPrefix := true

	for _, projQueueSpec := range project.Spec.Queues {
		var queue kaiv2.Queue
		var err error
		if queueNameEqualToSuggested {
			queue, err = getQueue(k8sclient, projQueueSpec.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(queue).ToNot(BeZero())

			Expect(queue.Name).To(Equal(projQueueSpec.Name))
		} else {
			queue, err = getQueueForProjectByLabel(k8sclient, project.Name, projQueueSpec.Nodepool)
			Expect(err).ToNot(HaveOccurred())
			Expect(queue).ToNot(BeZero())

			prefixToExpect := projQueueSpec.Name
			if len(prefixToExpect) > 58 {
				prefixToExpect = projQueueSpec.Name[:58]
			}

			Expect(queue.Name).To(HavePrefix(prefixToExpect))
			if queue.Name != projQueueSpec.Name {
				allQueueNamesWithoutPrefix = false
			}
		}

		expectedParentQueueName := expectedDepartmentQueueName(departmentId, projQueueSpec.Nodepool)
		if parentQueueNameEqualToSuggested {
			Expect(queue.Spec.ParentQueue).To(Equal(expectedParentQueueName))
		} else {
			Expect(queue.Spec.ParentQueue).To(HavePrefix(expectedParentQueueName))
			if queue.Spec.ParentQueue != expectedParentQueueName {
				allParentQueueNamesWithoutPrefix = false
			}
		}

		if projQueueSpec.Priority == nil {
			Expect(queue.Spec.Priority).To(Equal(ptr.To(DefaultQueuePriority)))
		} else {
			Expect(queue.Spec.Priority).To(Equal(ptr.To(int(*projQueueSpec.Priority))))
		}

		Expect(queue.Labels[config.Get().ProjectLabelKey]).To(Equal(project.Name))
		Expect(queue.Labels[config.Get().ProjectIdLabelKey]).To(Equal(string(project.UID)))
		Expect(queue.Labels[QueueDepartmentNameLabel]).To(Equal(project.Spec.Parent))
		if projQueueSpec.Nodepool != DefaultNodePoolName {
			Expect(queue.Labels[NodePoolLabelKey]).To(Equal(projQueueSpec.Nodepool))
		} else {
			Expect(queue.Labels).ToNot(HaveKey(NodePoolLabelKey))
		}

		Expect(queue.OwnerReferences).ToNot(BeZero())
		Expect(len(queue.OwnerReferences)).To(Equal(1))
		Expect(queue.OwnerReferences[0].APIVersion).To(Equal(project.APIVersion))
		Expect(queue.OwnerReferences[0].Kind).To(Equal(project.Kind))
		Expect(queue.OwnerReferences[0].Name).To(Equal(project.Name))
		Expect(queue.OwnerReferences[0].UID).To(Equal(project.UID))
		Expect(queue.OwnerReferences[0].Controller).To(Equal(&common.TrueRef))

		assertQueueResourcesIdenticalToProjectSpecResources(&queue, projQueueSpec)
	}

	if !queueNameEqualToSuggested && allQueueNamesWithoutPrefix {
		Expect(allQueueNamesWithoutPrefix).To(BeFalse())
	}
	if !parentQueueNameEqualToSuggested && allParentQueueNamesWithoutPrefix {
		Expect(allParentQueueNamesWithoutPrefix).To(BeFalse())
	}
}

func assertQueueResourcesIdenticalToProjectSpecResources(
	queue *kaiv2.Queue, projQueueSpec kaiv1alpha1.QueueConfig) {
	Expect(queue.Spec.Resources.GPU.Limit).To(Equal(projQueueSpec.Resources.GPU.Limit))
	Expect(queue.Spec.Resources.GPU.OverQuotaWeight).To(Equal(projQueueSpec.Resources.GPU.OverQuotaWeight))
	Expect(queue.Spec.Resources.GPU.Quota).To(Equal(projQueueSpec.Resources.GPU.Deserved))

	Expect(queue.Spec.Resources.CPU.Limit).To(Equal(projQueueSpec.Resources.CPU.Limit))
	Expect(queue.Spec.Resources.CPU.OverQuotaWeight).To(Equal(projQueueSpec.Resources.CPU.OverQuotaWeight))
	Expect(queue.Spec.Resources.CPU.Quota).To(Equal(projQueueSpec.Resources.CPU.Deserved))

	Expect(queue.Spec.Resources.Memory.Limit).To(Equal(projQueueSpec.Resources.Memory.Limit))
	Expect(queue.Spec.Resources.Memory.OverQuotaWeight).To(Equal(projQueueSpec.Resources.Memory.OverQuotaWeight))
	Expect(queue.Spec.Resources.Memory.Quota).To(Equal(projQueueSpec.Resources.Memory.Deserved))
}

func getQueueForProjectByLabel(k8sclient client.Client, projectName, nodepoolName string) (kaiv2.Queue, error) {
	queues := &kaiv2.QueueList{}
	err := k8sclient.List(context.Background(), queues,
		client.MatchingLabels(map[string]string{config.Get().ProjectLabelKey: projectName}))
	if err != nil {
		return kaiv2.Queue{}, err
	}

	for _, queue := range queues.Items {
		nodePoolLabelValue, found := queue.Labels[NodePoolLabelKey]

		if nodepoolName == DefaultNodePoolName && !found {
			return queue, nil
		}

		if nodepoolName != DefaultNodePoolName && found && nodePoolLabelValue == nodepoolName {
			return queue, nil
		}
	}

	return kaiv2.Queue{}, fmt.Errorf("couldn't find queue for project %s and nodepool %s",
		projectName, nodepoolName)
}

// expectedDepartmentQueueName returns the base (pre-suffix) queue name the department handler
// assigns to the given department's queue for the given nodepool, per the unified naming rule
// (queue spec name, or "<departmentName>[-<nodepool>]" fallback). Used to derive the expected
// parent-queue name for a project.
func expectedDepartmentQueueName(departmentId, nodepool string) string {
	var dep *kaiv1alpha1.Department
	switch departmentId {
	case TestDepartment1Id:
		dep = KaiTestDepartment1
	case TestDepartment2Id:
		dep = KaiTestDepartment2
	case TestDepartment3Id:
		dep = KaiTestDepartment3
	}
	if dep == nil {
		return ""
	}
	for _, q := range dep.Spec.Queues {
		if q.Nodepool == nodepool {
			return GetQueueName(q, dep.Name)
		}
	}
	return ""
}
