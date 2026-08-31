/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/common"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/assigner"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/assignment_params"

	"k8s.io/apimachinery/pkg/types"
)

const (
	testProject  = "team-a"
	testProject2 = "team-b"

	// podGroupName is the legacy pod-name prefix used by afterEachCleanup.
	podGroupName = "test-pod-group"

	key1 = "key1"
	key2 = "key2"
	key3 = "key3"
	val1 = "val1"
	val2 = "val2"
	val3 = "val3"

	RunaiNamespace = "runai"
)

type TestNodePool struct {
	Name       string                 `json:"name,omitempty"`
	LabelKey   string                 `json:"labelKey,omitempty"`
	LabelValue string                 `json:"labelValue,omitempty"`
	Phase      v1alpha1.NodePoolPhase `json:"phase,omitempty"`
}

var (
	testNamespace  = fmt.Sprintf("runai-%s", testProject)
	testNamespace2 = fmt.Sprintf("runai-%s", testProject2)

	defaultNodePool = TestNodePool{testDefaultNodePoolName, "", "", v1alpha1.NodePoolReady}
	nodePoolA       = TestNodePool{"node-pool-a", key1, val1, v1alpha1.NodePoolReady}
	nodePoolB       = TestNodePool{"node-pool-b", key2, val2, v1alpha1.NodePoolReady}
	nodePoolC       = TestNodePool{"node-pool-c", key3, val3, v1alpha1.NodePoolReady}
	testNodePools   = []TestNodePool{nodePoolA, nodePoolB, nodePoolC, defaultNodePool}

	emptyNodePool      = TestNodePool{"node-pool-empty", "emptykey", "emptyval", v1alpha1.NodePoolEmpty}
	unschedNodePool    = TestNodePool{"node-pool-unsched", "unschedkey", "unschedval", v1alpha1.NodePoolUnschedulable}
	unschedNodePool2   = TestNodePool{"node-pool-unsched2", "unschedkey2", "unschedval2", v1alpha1.NodePoolUnschedulable}
	deletingNodePool   = TestNodePool{"node-pool-deleting", "deletingkey", "deletingval", v1alpha1.NodePoolDeleting}
	extraTestNodePools = []TestNodePool{emptyNodePool, unschedNodePool, deletingNodePool, unschedNodePool2}

	createdPodGroups []types.NamespacedName
	numberOfPods     int

	testNamespaces = []string{RunaiNamespace, testNamespace, testNamespace2}
)

var _ = Describe("Pod Group Assigner Tests", Ordered, func() {
	var (
		assignmentParams = &assigner.FullNodePoolAssignmentParams{
			NodePoolAssignmentParams: assignment_params.NodePoolAssignmentParams{
				MarkUnschedulable: false,
				SchedulingBackoff: common.SingleSchedulingBackoff,
			},
		}
	)

	BeforeAll(func() {
		controllerSetup()

		for _, testNodePool := range testNodePools {
			createTestNodePool("BeforeAll", &testNodePool, k8sClient)
		}
		for _, testNamespace := range testNamespaces {
			createNamespaceIfNeeded(testNamespace, k8sClient)
		}

		createProjectWithNodePoolsQueues(testProject, append(testNodePools, extraTestNodePools...), k8sClient)
		createProjectWithNodePoolsQueuesInner(testProject2, append(testNodePools, extraTestNodePools...), true, k8sClient)
	})

	AfterAll(func() {
		deleteProject(testProject2, k8sClient)
		deleteProject(testProject, k8sClient)

		for _, testNamespace := range testNamespaces {
			deleteNamespace(testNamespace, k8sClient)
		}
		for _, testNodePool := range testNodePools {
			deleteNodePool("AfterAll", testNodePool.Name, k8sClient)
		}
	})

	Context("Basic Tests", func() {
		AfterEach(func() {
			afterEachCleanup()

			assignmentParams.NodePoolName = ""
			assignmentParams.MarkUnschedulable = false
			assignmentParams.SchedulingBackoff = common.SingleSchedulingBackoff
		})

		It("Sanity - Assign Pod Group to Node Pools one by one", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{testNodePools[1].Name, testNodePools[0].Name, testNodePools[2].Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("Sanity - Assign Pod Group to Node Pools one by one", pg, nodePoolOptions, 2, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[1]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// MarkUnschedulable should always be true from now on
			assignmentParams.MarkUnschedulable = true
			assignmentParams.NodePoolName = nodePoolOptions[2]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// validate round-robin
			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)
		})

		It("default node pool is one of list", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{testNodePools[1].Name, testNodePools[2].Name, defaultNodePool.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("default node pool is one of list", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[1]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// MarkUnschedulable should always be true from now on
			assignmentParams.MarkUnschedulable = true
			assignmentParams.NodePoolName = nodePoolOptions[2]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)
		})

		It("Only one in NodePool options", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{nodePoolB.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("Only one in NodePool options", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			assignmentParams.MarkUnschedulable = true
			assignmentParams.SchedulingBackoff = common.NoSchedulingBackoff
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("Only default in NodePool options", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{testDefaultNodePoolName}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("Only default in NodePool options", pg, nodePoolOptions, 2, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			assignmentParams.MarkUnschedulable = true
			assignmentParams.SchedulingBackoff = common.NoSchedulingBackoff
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("Backwards Compatability - nodepool label on pod group", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolName := nodePoolB.Name

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupFromSpecWithSingleNPOldStyle(
				"Backwards Compatability - nodepool label on pod group",
				pg,
				nodePoolName,
				testProject,
				k8sClient)

			assignmentParams.NodePoolName = nodePoolName
			assignmentParams.MarkUnschedulable = true
			assignmentParams.SchedulingBackoff = common.NoSchedulingBackoff
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("No NodePool options - submit to default", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("No NodePool options - submit to default", pg, []string{}, 1, k8sClient)

			assignmentParams.NodePoolName = testDefaultNodePoolName
			assignmentParams.MarkUnschedulable = true
			assignmentParams.SchedulingBackoff = common.NoSchedulingBackoff
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("Not eligible for NodePool assignment", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{testNodePools[1].Name, testNodePools[0].Name, testNodePools[2].Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("Not eligible for NodePool assignment", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			// not mocking scheduler marking UnschedulableOnNodePool - our controller should ignore this pod group
			updateRandomLabelToTriggerController(pg, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[1]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			// not mocking scheduler marking UnschedulableOnNodePool - our controller should ignore this pod group
			updateRandomLabelToTriggerController(pg, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("Non Existing NodePool", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nonExistingNodePoolName := generateNodePoolName()
			nodePoolOptions := []string{nonExistingNodePoolName, nodePoolC.Name, nodePoolA.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("Non Existing NodePool", pg, nodePoolOptions, 1, k8sClient)

			// will ignore the first NP as it does not exist - will assign to second NP

			assignmentParams.NodePoolName = nodePoolOptions[1]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[2]
			assignmentParams.MarkUnschedulable = true
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)
		})

		It("No Pods at first - then pods are created with PodGroup - should be assigned", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{testNodePools[1].Name, testNodePools[0].Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("No Pods at first - then pods are created with PodGroup - should be assigned", pg, nodePoolOptions, 0, k8sClient)

			// pod group will not be assigned to any NP
			assignmentParams.NodePoolName = testUnexistingNodePool
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			// now add some pods
			createPodsForPodGroupForNodePoolOptions(pg.Name, pg.Namespace, nodePoolOptions, 2)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)
		})

		It("Queue names for project with suffix - Assign Pod Group to Node Pools one by one", func() {
			pg := types.NamespacedName{Namespace: testNamespace2, Name: generatePodGroupName()}
			nodePoolOptions := []string{testNodePools[1].Name, testNodePools[0].Name, testNodePools[2].Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpecWithProjName("Queue names for project with suffix - Assign Pod Group to Node Pools one by one",
				pg, nodePoolOptions, 2, testProject2, k8sClient)

			queues, err := getExistingQueuesByLabels(testProject2, nodePoolOptions[0], k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(queues).To(HaveLen(1))
			nodePool0ProjQueue := queues[0]
			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignmentWithQueueName(pg, nodePool0ProjQueue.Name, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			queues, err = getExistingQueuesByLabels(testProject2, nodePoolOptions[1], k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(queues).To(HaveLen(1))
			nodePool1ProjQueue := queues[0]
			assignmentParams.NodePoolName = nodePoolOptions[1]
			eventuallyExpectNodePoolAssignmentWithQueueName(pg, nodePool1ProjQueue.Name, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// MarkUnschedulable should always be true from now on
			queues, err = getExistingQueuesByLabels(testProject2, nodePoolOptions[2], k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(queues).To(HaveLen(1))
			nodePool2ProjQueue := queues[0]
			assignmentParams.MarkUnschedulable = true
			assignmentParams.NodePoolName = nodePoolOptions[2]
			eventuallyExpectNodePoolAssignmentWithQueueName(pg, nodePool2ProjQueue.Name, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// validate round-robin
			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignmentWithQueueName(pg, nodePool0ProjQueue.Name, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)
		})
	})

	Context("Non Ready NodePools", func() {
		BeforeAll(func() {
			for _, testNodePool := range extraTestNodePools {
				createTestNodePool("BeforeAll", &testNodePool, k8sClient)
			}
		})

		AfterAll(func() {
			for _, testNodePool := range extraTestNodePools {
				deleteNodePool("AfterAll", testNodePool.Name, k8sClient)
			}
		})

		AfterEach(func() {
			afterEachCleanup()

			for _, testNodePool := range extraTestNodePools {
				updateNodePoolStatus(testNodePool.Name, testNodePool.Phase)
			}
			for _, testNodePool := range testNodePools {
				updateNodePoolStatus(testNodePool.Name, testNodePool.Phase)
			}

			assignmentParams.NodePoolName = ""
			assignmentParams.MarkUnschedulable = false
			assignmentParams.SchedulingBackoff = common.SingleSchedulingBackoff
		})

		It("Non Ready NodePools Sanity", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{nodePoolC.Name, emptyNodePool.Name, unschedNodePool.Name, deletingNodePool.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("Non Ready NodePools Sanity", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// will assign to empty node pool - it is available for scheduling. just no nodes in it
			assignmentParams.NodePoolName = nodePoolOptions[1]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// will ignore all the non-(ready-or-empty) NodePools as they are not available for scheduling - will assign to next NP

			// round-robin - will assign to the first one again -
			// also validate MarkUnschedulable as it should be true now
			assignmentParams.NodePoolName = nodePoolOptions[0]
			assignmentParams.MarkUnschedulable = true
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)
		})

		It("All NodePool options not ready - then one becomes ready - validate assignment", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{unschedNodePool2.Name, unschedNodePool.Name, deletingNodePool.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("All NodePool options not ready - then one becomes ready - validate assignment", pg, nodePoolOptions, 1, k8sClient)

			// pod group will not be assigned to any NP
			assignmentParams.NodePoolName = testUnexistingNodePool
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			updateNodePoolStatus(unschedNodePool2.Name, v1alpha1.NodePoolReady)

			// now that one of the node pools is ready - expect assignment
			assignmentParams.NodePoolName = unschedNodePool2.Name
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("1 ready, 1 non-ready node pool", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{nodePoolB.Name, unschedNodePool.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("1 ready, 1 non-ready node pool", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePoolNoValidate(pg, assignmentParams, k8sClient)
			expectNoUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// will ignore all the non-(ready-or-empty) NodePools as they are not available for scheduling - will assign to next NP

			// round-robin - will assign to the first one again -
			// also validate MarkUnschedulable as it should be true now,
			// and RemoveUnschedulableOnNodePoolSchedulingCondition should be true now - cause only 1 NP is technically available
			assignmentParams.MarkUnschedulable = true
			assignmentParams.RemoveUnschedulableOnNodePoolSchedulingCondition = true
			assignmentParams.SchedulingBackoff = common.SingleSchedulingBackoff
			eventuallyExpectNodePoolAssignmentRemoveUnschedulableOnNodePoolSchedulingCondition(pg, assignmentParams, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			SchedulerMock.UnschedulableOnNodePoolNoValidate(pg, assignmentParams, k8sClient)
			expectNoUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			updateRandomLabelToTriggerController(pg, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("1 non-ready, 1 ready node pool", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{unschedNodePool.Name, nodePoolA.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("1 non-ready, 1 ready node pool", pg, nodePoolOptions, 1, k8sClient)

			// will ignore all the non-(ready-or-empty) NodePools as they are not available for scheduling - will assign to next NP

			assignmentParams.NodePoolName = nodePoolOptions[1]
			assignmentParams.MarkUnschedulable = true
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePoolNoValidate(pg, assignmentParams, k8sClient)
			expectNoUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			assignmentParams.SchedulingBackoff = common.SingleSchedulingBackoff
			assignmentParams.RemoveUnschedulableOnNodePoolSchedulingCondition = true
			eventuallyExpectNodePoolAssignmentRemoveUnschedulableOnNodePoolSchedulingCondition(pg, assignmentParams, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			SchedulerMock.UnschedulableOnNodePoolNoValidate(pg, assignmentParams, k8sClient)
			expectNoUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			updateRandomLabelToTriggerController(pg, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("1 ready, 1 non-ready node pool, suddenly the NP becomes Ready", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{nodePoolB.Name, unschedNodePool.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("1 ready, 1 non-ready node pool, suddenly the NP becomes Ready", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePoolNoValidate(pg, assignmentParams, k8sClient)
			expectNoUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// will ignore all the non-(ready-or-empty) NodePools as they are not available for scheduling - will assign to next NP

			// round-robin - will assign to the first one again -
			// also validate MarkUnschedulable as it should be true now,
			// and RemoveUnschedulableOnNodePoolSchedulingCondition should be true now - cause only 1 NP is technically available
			assignmentParams.MarkUnschedulable = true
			assignmentParams.RemoveUnschedulableOnNodePoolSchedulingCondition = true
			assignmentParams.SchedulingBackoff = common.SingleSchedulingBackoff
			eventuallyExpectNodePoolAssignmentRemoveUnschedulableOnNodePoolSchedulingCondition(pg, assignmentParams, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			SchedulerMock.UnschedulableOnNodePoolNoValidate(pg, assignmentParams, k8sClient)
			expectNoUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			// now the nodepool becomes Ready - validate assignment moves to this nodepool
			updateNodePoolStatus(nodePoolOptions[1], v1alpha1.NodePoolReady)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[1]
			assignmentParams.MarkUnschedulable = true
			assignmentParams.RemoveUnschedulableOnNodePoolSchedulingCondition = false
			assignmentParams.SchedulingBackoff = common.SingleSchedulingBackoff
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)
		})

		It("1 ready, 1 deleting node pool", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{nodePoolB.Name, deletingNodePool.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("1 ready, 1 deleting node pool", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			assignmentParams.MarkUnschedulable = true
			assignmentParams.SchedulingBackoff = common.NoSchedulingBackoff
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			updateRandomLabelToTriggerController(pg, k8sClient)

			eventuallyExpectNodePoolAssignmentNoLastEventCheck(pg, assignmentParams, k8sClient)

			updateRandomLabelToTriggerController(pg, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("1 deleting, 1 ready node pool", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{deletingNodePool.Name, nodePoolA.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("1 deleting, 1 ready node pool", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[1]
			assignmentParams.MarkUnschedulable = true
			assignmentParams.SchedulingBackoff = common.NoSchedulingBackoff
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			eventuallyExpectNodePoolAssignmentNoLastEventCheck(pg, assignmentParams, k8sClient)

			updateRandomLabelToTriggerController(pg, k8sClient)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("During scheduling cycle, suddenly the NP becomes Unschedulable/Deleting", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{nodePoolA.Name, nodePoolB.Name, unschedNodePool.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("During scheduling cycle, suddenly the NP becomes Unschedulable/Deleting", pg, nodePoolOptions, 2, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			// not mocking scheduler marking UnschedulableOnNodePool - our controller should ignore this pod group

			// now "suddenly" the NP becomes Deleting - move on to the next NP
			updateNodePoolStatus(nodePoolOptions[0], v1alpha1.NodePoolDeleting)

			assignmentParams.NodePoolName = nodePoolOptions[1]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			// now "suddenly" the NP becomes Unschedulable - move on to the next NP
			updateNodePoolStatus(nodePoolOptions[1], v1alpha1.NodePoolUnschedulable)

			// the next NP is Unschedulable - and the first one is still deleting
			// the PG will be left "hanging" in this NP
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			// now the same NP becomes Ready (Empty) again
			// the controller doesn't know it was already assigned to this NP - cause we don't have any events... so try again
			updateNodePoolStatus(nodePoolOptions[1], v1alpha1.NodePoolEmpty)
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})

		It("During scheduling cycle, suddenly the NP becomes Unschedulable/Deleting - reassigning to previous NP", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{nodePoolA.Name, nodePoolB.Name, unschedNodePool.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("During scheduling cycle, suddenly the NP becomes Unschedulable/Deleting - reassigning to previous NP", pg, nodePoolOptions, 2, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
			SchedulerMock.UnschedulableOnNodePool(pg, assignmentParams, k8sClient)
			expectUnschedulableOnNodePool(pg, assignmentParams, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[1]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			// not mocking scheduler marking UnschedulableOnNodePool - our controller should ignore this pod group

			// now "suddenly" the NP becomes Deleting - move on to the next NP - which will the first from the list
			updateNodePoolStatus(nodePoolOptions[1], v1alpha1.NodePoolDeleting)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			assignmentParams.MarkUnschedulable = true
			assignmentParams.SchedulingBackoff = common.SingleSchedulingBackoff
			assignmentParams.RemoveUnschedulableOnNodePoolSchedulingCondition = true
			eventuallyExpectNodePoolAssignmentRemoveUnschedulableOnNodePoolSchedulingCondition(pg, assignmentParams, k8sClient)
		})

		It("During scheduling cycle, suddenly the NP becomes Unschedulable - but pg is allocated, so don't re-assign", func() {
			pg := types.NamespacedName{Namespace: testNamespace, Name: generatePodGroupName()}
			nodePoolOptions := []string{nodePoolA.Name, nodePoolB.Name}

			createdPodGroups = append(createdPodGroups, pg)
			createPodGroupAndPodsFromSpec("During scheduling cycle, suddenly the NP becomes Unschedulable - but pg is allocated, so don't re-assign", pg, nodePoolOptions, 1, k8sClient)

			assignmentParams.NodePoolName = nodePoolOptions[0]
			eventuallyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)

			SchedulerMock.Running(pg, k8sClient)

			// now "suddenly" the NP becomes Unschedulable
			updateNodePoolStatus(nodePoolOptions[0], v1alpha1.NodePoolUnschedulable)

			// don't move to next NP - since the PG is Running!
			consistentlyExpectNodePoolAssignment(pg, assignmentParams, k8sClient)
		})
	})
})

func afterEachCleanup() {
	for _, podGroupNamespacedName := range createdPodGroups {
		deletePodGroup(podGroupNamespacedName, k8sClient)

		for i := 0; i < numberOfPods; i++ {
			deletePod(types.NamespacedName{Name: getPodNameForPodGroupIndex(podGroupName, i), Namespace: podGroupNamespacedName.Namespace},
				k8sClient)
		}

		numberOfPods = 0
	}

	createdPodGroups = []types.NamespacedName{}
}
