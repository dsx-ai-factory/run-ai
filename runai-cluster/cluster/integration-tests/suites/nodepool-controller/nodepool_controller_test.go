/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"fmt"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/integration-tests/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/integration-tests/config"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NodePool Controller Test", Ordered, func() {
	var (
		testConfig         *config.TestConfig
		kubeClient         client.Client
		ctx                context.Context
		testCreatedObjects = map[string]client.Object{}
		nodePoolName       string
		labelKey           string
		labelValue         string
		workerNodeName     string
	)

	BeforeAll(func() {
		testConfig = config.GetTestConfig(Default)
		kubeClient = testConfig.KubeClient
		ctx = testConfig.Ctx
		workerNodeName, _ = getWorkerNodeName(ctx, kubeClient)

	})

	AfterAll(func() {
		for _, obj := range testCreatedObjects {
			if err := kubeClient.Delete(ctx, obj); err != nil {
				fmt.Printf("Failed to delete resource %v: %v\n", obj, err)
			}
		}
		testConfig.Cleanup()
	})

	Context("Sanity nodepools tests", func() {

		AfterEach(func() {
			// remove labels from node
			err := removeNodeLabel(ctx, kubeClient, workerNodeName, labelKey)
			if err != nil {
				return
			}
			_ = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return removeNodeLabel(ctx, kubeClient, workerNodeName, "runai/node-pool")
			})
			// delete node-pool
			err = kubeClient.Delete(ctx, testCreatedObjects[nodePoolName])
			if err == nil {
				delete(testCreatedObjects, nodePoolName)
			}
		})

		It("Create node-pool and assign it a node", func() {
			nodePoolName = common.GetRandomName("node-pool")
			labelKey = "node-pool-key"
			labelValue = "node-pool-val"

			err := CreateNodePoolWithFullIntegration(ctx, kubeClient, NodePoolCreationConfig{
				NodePoolName:    nodePoolName,
				LabelKey:        labelKey,
				LabelValue:      labelValue,
				DepartmentNames: []string{testConfig.TestDepartment.Name},
				ProjectNames:    []string{testConfig.TestProject.Name},
			})
			Expect(err).To(BeNil())

			// Get the created nodepool for testCreatedObjects
			nodePool := &v1alpha1.NodePool{}
			err = kubeClient.Get(ctx, client.ObjectKey{Name: nodePoolName}, nodePool)
			Expect(err).To(BeNil())
			testCreatedObjects[nodePoolName] = nodePool

			err = labelNode(ctx, kubeClient, workerNodeName, labelKey, labelValue)
			Expect(err).To(BeNil())
			Eventually(func() []v1alpha1.NodeInNodePool {
				updatedNodePool := &v1alpha1.NodePool{}
				err := kubeClient.Get(ctx, client.ObjectKey{Name: nodePoolName}, updatedNodePool)
				Expect(err).To(BeNil())
				return updatedNodePool.Status.Nodes
			}, "30s", "2s").Should(ContainElement(v1alpha1.NodeInNodePool{Name: workerNodeName, Status: v1alpha1.NodeReady}))
			// Verify that the node has the correct node pool annotation
			Eventually(func() string {
				node := &v1.Node{}
				err := kubeClient.Get(ctx, client.ObjectKey{Name: workerNodeName}, node)
				Expect(err).To(BeNil())
				return node.Labels["runai/node-pool"]
			}, "30s", "2s").Should(Equal(nodePoolName))

			_ = validateNodeAndNodePoolCoupling(ctx, kubeClient, nodePoolName, workerNodeName)
		})
	})

	Context("NodePool deletion blocked by project references", func() {
		var deletionNodePoolName string

		AfterEach(func() {
			// best-effort cleanup in case the test failed before completing the deletion
			_ = RemoveNodePoolFromProjectResources(ctx, kubeClient, testConfig.TestProject.Name, deletionNodePoolName)
			nodePool := &v1alpha1.NodePool{}
			if err := kubeClient.Get(ctx, client.ObjectKey{Name: deletionNodePoolName}, nodePool); err == nil {
				_ = kubeClient.Delete(ctx, nodePool)
			}
		})

		It("stays in Deleting while a project references it and is deleted once the reference is removed", func() {
			deletionNodePoolName = common.GetRandomName("node-pool")

			err := CreateNodePoolWithFullIntegration(ctx, kubeClient, NodePoolCreationConfig{
				NodePoolName:    deletionNodePoolName,
				LabelKey:        "deletion-node-pool-key",
				LabelValue:      "deletion-node-pool-val",
				DepartmentNames: []string{testConfig.TestDepartment.Name},
				ProjectNames:    []string{testConfig.TestProject.Name},
			})
			Expect(err).To(BeNil())

			// Wait for the controller to add its finalizer, so the delete below is blocked
			// rather than completing immediately.
			Eventually(func() []string {
				nodePool := &v1alpha1.NodePool{}
				Expect(kubeClient.Get(ctx, client.ObjectKey{Name: deletionNodePoolName}, nodePool)).To(Succeed())
				return nodePool.Finalizers
			}, "30s", "2s").ShouldNot(BeEmpty())

			// Delete the nodepool while the project still references it.
			nodePool := &v1alpha1.NodePool{}
			Expect(kubeClient.Get(ctx, client.ObjectKey{Name: deletionNodePoolName}, nodePool)).To(Succeed())
			Expect(kubeClient.Delete(ctx, nodePool)).To(Succeed())

			// The nodepool must stay present in the Deleting phase and expose the
			// ProjectReferencesExist condition naming the blocking project.
			Eventually(func(g Gomega) {
				blocked := &v1alpha1.NodePool{}
				g.Expect(kubeClient.Get(ctx, client.ObjectKey{Name: deletionNodePoolName}, blocked)).To(Succeed())
				g.Expect(blocked.DeletionTimestamp.IsZero()).To(BeFalse())
				g.Expect(blocked.Status.Phase).To(Equal(v1alpha1.NodePoolDeleting))
				condition := getNodePoolCondition(blocked, v1alpha1.ProjectReferencesExist)
				g.Expect(condition).ToNot(BeNil())
				g.Expect(condition.Status).To(Equal(v1.ConditionTrue))
				g.Expect(condition.Message).To(ContainSubstring(testConfig.TestProject.Name))
			}, "30s", "2s").Should(Succeed())

			// It must remain blocked (not silently deleted) while the reference exists.
			Consistently(func() error {
				return kubeClient.Get(ctx, client.ObjectKey{Name: deletionNodePoolName}, &v1alpha1.NodePool{})
			}, "10s", "2s").Should(Succeed())

			// Removing the project reference should unblock deletion.
			Expect(RemoveNodePoolFromProjectResources(ctx, kubeClient, testConfig.TestProject.Name, deletionNodePoolName)).To(Succeed())

			Eventually(func() bool {
				err := kubeClient.Get(ctx, client.ObjectKey{Name: deletionNodePoolName}, &v1alpha1.NodePool{})
				return errors.IsNotFound(err)
			}, "60s", "2s").Should(BeTrue())
		})
	})

	Context("Deployment scheduling with multiple nodepools", func() {
		var (
			nodePool1Name  string
			nodePool2Name  string
			nodePool1Label string
			nodePool2Label string
			nodePool1Value string
			nodePool2Value string
			deploymentName string
			availableNodes []string
		)

		BeforeAll(func() {
			// Get available worker nodes
			nodeList := &v1.NodeList{}
			err := kubeClient.List(ctx, nodeList)
			Expect(err).To(BeNil())

			for _, node := range nodeList.Items {
				if _, exists := node.Labels["node-role.kubernetes.io/control-plane"]; !exists {
					availableNodes = append(availableNodes, node.Name)
				}
			}
			if len(availableNodes) < 2 {
				Skip(fmt.Sprintf("Skipping test: only %d worker nodes available, need at least 2", len(availableNodes)))
			}

			// Create two nodepools
			nodePool1Name = common.GetRandomName("nodepool-1")
			nodePool2Name = common.GetRandomName("nodepool-2")
			nodePool1Label = "nodepool-1-key"
			nodePool2Label = "nodepool-2-key"
			nodePool1Value = "nodepool-1-value"
			nodePool2Value = "nodepool-2-value"

			err = CreateNodePoolWithFullIntegration(ctx, kubeClient, NodePoolCreationConfig{
				NodePoolName:    nodePool1Name,
				LabelKey:        nodePool1Label,
				LabelValue:      nodePool1Value,
				DepartmentNames: []string{testConfig.TestDepartment.Name},
				ProjectNames:    []string{testConfig.TestProject.Name},
			})
			Expect(err).To(BeNil())

			// Get the created nodepool for testCreatedObjects
			nodePool1 := &v1alpha1.NodePool{}
			err = kubeClient.Get(ctx, client.ObjectKey{Name: nodePool1Name}, nodePool1)
			Expect(err).To(BeNil())
			testCreatedObjects[nodePool1Name] = nodePool1

			err = CreateNodePoolWithFullIntegration(ctx, kubeClient, NodePoolCreationConfig{
				NodePoolName:    nodePool2Name,
				LabelKey:        nodePool2Label,
				LabelValue:      nodePool2Value,
				DepartmentNames: []string{testConfig.TestDepartment.Name},
				ProjectNames:    []string{testConfig.TestProject.Name},
			})
			Expect(err).To(BeNil())

			// Get the created nodepool for testCreatedObjects
			nodePool2 := &v1alpha1.NodePool{}
			err = kubeClient.Get(ctx, client.ObjectKey{Name: nodePool2Name}, nodePool2)
			Expect(err).To(BeNil())
			testCreatedObjects[nodePool2Name] = nodePool2

			// Label nodes for different nodepools
			err = labelNode(ctx, kubeClient, availableNodes[0], nodePool1Label, nodePool1Value)
			Expect(err).To(BeNil())

			err = labelNode(ctx, kubeClient, availableNodes[1], nodePool2Label, nodePool2Value)
			Expect(err).To(BeNil())
		})

		AfterAll(func() {
			// Clean up deployment
			if deploymentName != "" {
				deployment := testCreatedObjects[deploymentName]
				if deployment != nil {
					if err := kubeClient.Delete(ctx, deployment); err != nil {
						fmt.Printf("Failed to delete deployment %s: %v\n", deploymentName, err)
					}
					delete(testCreatedObjects, deploymentName)

					// Wait for deployment pods to be deleted
					Eventually(func() int {
						podList := &v1.PodList{}
						err := kubeClient.List(ctx, podList, client.InNamespace(testConfig.TestNsName), client.MatchingLabels{"app": deploymentName})
						if err != nil {
							fmt.Printf("Failed to list pods for deployment %s: %v\n", deploymentName, err)
							return -1
						}
						return len(podList.Items)
					}, "60s", "2s").Should(BeZero(), "All deployment pods should be deleted")
				}
			}

			// Clean up labels with retry to handle race conditions
			for _, nodeName := range availableNodes {
				err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					return removeNodeLabel(ctx, kubeClient, nodeName, nodePool1Label)
				})
				Expect(err).To(BeNil())
				err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
					return removeNodeLabel(ctx, kubeClient, nodeName, nodePool2Label)
				})
				Expect(err).To(BeNil())
			}

			// Wait for node labels to not exist
			for _, nodeName := range availableNodes {
				eventuallyValidateNodeLabel(ctx, kubeClient, nodeName, nodePool1Label, "", false)
				eventuallyValidateNodeLabel(ctx, kubeClient, nodeName, nodePool2Label, "", false)
			}
			for _, nodeName := range availableNodes {
				eventuallyValidateNodeLabel(ctx, kubeClient, nodeName, "runai/node-pool", "", false)
			}

			// Clean up nodepools
			if nodePool1Name != "" {
				nodePool1 := testCreatedObjects[nodePool1Name]
				if nodePool1 != nil {
					if err := kubeClient.Delete(ctx, nodePool1); err != nil {
						fmt.Printf("Failed to delete nodepool %s: %v\n", nodePool1Name, err)
					}
					delete(testCreatedObjects, nodePool1Name)
				}
			}
			if nodePool2Name != "" {
				nodePool2 := testCreatedObjects[nodePool2Name]
				if nodePool2 != nil {
					if err := kubeClient.Delete(ctx, nodePool2); err != nil {
						fmt.Printf("Failed to delete nodepool %s: %v\n", nodePool2Name, err)
					}
					delete(testCreatedObjects, nodePool2Name)
				}
			}
		})

		It("Deploy application with nodepool affinity and validate scheduling", func() {
			deploymentName = common.GetRandomName("test-deployment")
			replicas := int32(1)

			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: testConfig.TestNsName,
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": deploymentName,
						},
					},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": deploymentName,
							},
						},
						Spec: v1.PodSpec{
							SchedulerName: "runai-scheduler",
							Containers: []v1.Container{
								{
									Name:    "test-container",
									Image:   "ubuntu",
									Command: []string{"sleep", "infinity"},
									Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											v1.ResourceCPU: resource.MustParse("10m"),
										},
									},
								},
							},
							Affinity: &v1.Affinity{
								NodeAffinity: &v1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
										NodeSelectorTerms: []v1.NodeSelectorTerm{
											{
												MatchExpressions: []v1.NodeSelectorRequirement{
													{
														Key:      nodePool1Label,
														Operator: v1.NodeSelectorOpIn,
														Values:   []string{nodePool1Value},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			err := kubeClient.Create(ctx, deployment)
			Expect(err).To(BeNil())
			testCreatedObjects[deploymentName] = deployment

			// Wait for deployment to be ready
			Eventually(func() bool {
				updatedDeployment := &appsv1.Deployment{}
				err := kubeClient.Get(ctx, client.ObjectKey{Name: deploymentName, Namespace: testConfig.TestNsName}, updatedDeployment)
				if err != nil {
					fmt.Printf("Failed to get deployment %s: %v\n", deploymentName, err)
					return false
				}
				// If deployment is not ready, log pod scheduling issues for debugging
				if updatedDeployment.Status.ReadyReplicas != replicas {
					podList := &v1.PodList{}
					err := kubeClient.List(ctx, podList, client.InNamespace(testConfig.TestNsName), client.MatchingLabels{"app": deploymentName})
					if err == nil && len(podList.Items) > 0 {
						for _, pod := range podList.Items {
							if pod.Status.Phase == v1.PodPending {
								for _, condition := range pod.Status.Conditions {
									if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionFalse {
										fmt.Printf("Pod scheduling issue: %s - %s\n", condition.Reason, condition.Message)
									}
								}
							}
						}
					}
				}
				return updatedDeployment.Status.ReadyReplicas == replicas
			}, "60s", "2s").Should(BeTrue(), "Deployment should become ready")

			// Verify pod is scheduled on the correct nodepool
			Eventually(func() string {
				podList := &v1.PodList{}
				err := kubeClient.List(ctx, podList, client.InNamespace(testConfig.TestNsName), client.MatchingLabels{"app": deploymentName})
				if err != nil {
					fmt.Printf("Failed to list pods for deployment %s: %v\n", deploymentName, err)
					return ""
				}
				if len(podList.Items) == 0 {
					return ""
				}

				pod := podList.Items[0]
				if pod.Spec.NodeName == "" {
					return ""
				}

				return pod.Spec.NodeName
			}, "60s", "2s").Should(Equal(availableNodes[0]), "Pod should be scheduled on the node with nodepool-1 label")

			// Verify that the node has the correct runai/node-pool label
			Eventually(func() string {
				node := &v1.Node{}
				err := kubeClient.Get(ctx, client.ObjectKey{Name: availableNodes[0]}, node)
				if err != nil {
					fmt.Printf("Failed to get node %s: %v\n", availableNodes[0], err)
					return ""
				}
				return node.Labels["runai/node-pool"]
			}, "30s", "2s").Should(Equal(nodePool1Name), "Node should have the correct runai/node-pool label")

			// Verify nodepool status shows the correct node
			Eventually(func() []v1alpha1.NodeInNodePool {
				updatedNodePool := &v1alpha1.NodePool{}
				err := kubeClient.Get(ctx, client.ObjectKey{Name: nodePool1Name}, updatedNodePool)
				if err != nil {
					fmt.Printf("Failed to get nodepool %s: %v\n", nodePool1Name, err)
					return nil
				}
				return updatedNodePool.Status.Nodes
			}, "30s", "2s").Should(ContainElement(v1alpha1.NodeInNodePool{Name: availableNodes[0], Status: v1alpha1.NodeReady}), "NodePool should show the correct node in its status")

			// Verify the second nodepool also has its node correctly labeled
			Eventually(func() string {
				node := &v1.Node{}
				err := kubeClient.Get(ctx, client.ObjectKey{Name: availableNodes[1]}, node)
				if err != nil {
					fmt.Printf("Failed to get second node %s: %v\n", availableNodes[1], err)
					return ""
				}
				return node.Labels["runai/node-pool"]
			}, "30s", "2s").Should(Equal(nodePool2Name), "Second node should have the correct runai/node-pool label")
		})
	})

	Context("Exclude node from Run:AI tests", func() {
		Context("Exclude node - sanity", func() {
			labelKey := "node-to-exclude"

			BeforeAll(func() {
				err := labelNode(ctx, kubeClient, workerNodeName, labelKey, "true")
				Expect(err).ToNot(HaveOccurred(), "Failed to label node:")
				err = editRunaiConfigAndExcludeNodesByLabelFromRunai(ctx, kubeClient, labelKey)
				Expect(err).ToNot(HaveOccurred(), "Failed to exclude nodes from Run:AI:")
			})

			AfterAll(func() {
				err := editRunaiConfigAndRemoveNodeExclusion(ctx, kubeClient)
				Expect(err).ToNot(HaveOccurred(), "Failed to remove node exclusion from Run:AI:")
				err = removeNodeLabel(ctx, kubeClient, workerNodeName, labelKey)
				Expect(err).ToNot(HaveOccurred(), "Failed to remove label from node:")
				Eventually(func() string {
					node := &v1.Node{}
					err := kubeClient.Get(ctx, client.ObjectKey{Name: workerNodeName}, node)
					if err != nil {
						return ""
					}
					return node.Labels["runai/node-pool"]
				}, "30s", "2s").Should(Equal(""))
			})

			It("Validate node is excluded from Run:AI", func() {
				var node v1.Node
				err := kubeClient.Get(ctx, client.ObjectKey{Name: workerNodeName}, &node)
				Expect(err).To(BeNil())

				eventuallyValidateNodeLabel(ctx, kubeClient, workerNodeName, "runai/node-pool", "runai-excluded-nodes", true)

				managedNodesConfig := &v1alpha1.ManagedNodesConfig{}
				err = kubeClient.Get(ctx, client.ObjectKey{Name: "runai-managed-nodes-config"}, managedNodesConfig)
				Expect(err).To(BeNil())
				expectedIncludeCriteriaTerm := v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      labelKey,
							Operator: v1.NodeSelectorOpDoesNotExist,
						},
					},
				}

				Eventually(func() []v1.NodeSelectorTerm {
					err := kubeClient.Get(ctx, client.ObjectKey{Name: "runai-managed-nodes-config"}, managedNodesConfig)
					if err != nil {
						return nil
					}
					return managedNodesConfig.Spec.InclusionCriteria.NodeSelectorTerms
				}, "30s", "2s").Should(ContainElement(expectedIncludeCriteriaTerm))

				Eventually(func() bool {
					err := kubeClient.Get(ctx, client.ObjectKey{Name: "runai-managed-nodes-config"}, managedNodesConfig)
					if err != nil {
						return false
					}
					for _, condition := range managedNodesConfig.Status.Conditions {
						if condition.Type == string(v1alpha1.MNCConditionTypeApplied) && condition.Reason == string(v1alpha1.MNCConditionReasonAllNodesIncludedCorrectly) {
							return true
						}
					}
					return false
				}, "30s", "2s").Should(BeTrue())
			})
		})

		Context("Exclude node - scheduling", Ordered, func() {
			var workerNodeName string

			BeforeAll(func() {
				Skip("")

				// Schedule a pod on the worker node
				pod, err := submitPod(ctx, kubeClient, testConfig.TestProject.Name)
				Expect(err).To(BeNil())
				testCreatedObjects["test-pod"] = pod

				// Get worker node name which is the node that the pod is scheduled on
				podFromCluster := &v1.Pod{}

				Eventually(func() string {
					err := kubeClient.Get(ctx, client.ObjectKey{Name: pod.Name, Namespace: pod.Namespace}, podFromCluster)
					if err != nil {
						fmt.Printf("Error getting pod from cluster: %v\n", err)
						return ""
					}
					workerNodeName = podFromCluster.ObjectMeta.Annotations["runai-used-nodes"]
					return workerNodeName
				}, "30s", "2s").Should(SatisfyAll(
					Not(BeEmpty()),
					Not(Equal("-")),
				))

				// Exclude the node from Run:AI
				labelKey := "node-to-exclude"
				err = labelNode(ctx, kubeClient, workerNodeName, labelKey, "true")
				Expect(err).ToNot(HaveOccurred(), "Failed to label node:")

				err = editRunaiConfigAndExcludeNodesByLabelFromRunai(ctx, kubeClient, labelKey)
				Expect(err).ToNot(HaveOccurred(), "Failed to exclude nodes from Run:AI:")
			})

			AfterAll(func() {
				// Delete the test pod
				pod := &v1.Pod{}
				err := kubeClient.Get(ctx, client.ObjectKey{Name: testCreatedObjects["test-pod"].GetName(), Namespace: testCreatedObjects["test-pod"].GetNamespace()}, pod)
				if err == nil {
					err = kubeClient.Delete(ctx, pod)
					Expect(err).ToNot(HaveOccurred(), "Failed to delete pod:")
				} else if !errors.IsNotFound(err) {
					Fail(fmt.Sprintf("Failed to get pod: %v", err))
				}

				// Wait for the specific annotation to be removed
				Eventually(func(g Gomega) {
					node := &v1.Node{}
					err := kubeClient.Get(ctx, client.ObjectKey{Name: workerNodeName}, node)
					g.Expect(err).To(BeNil(), "Failed to get node:")
					_, exists := node.Annotations["runai.ai/to-exclude"]
					g.Expect(exists).To(BeFalse(), "The annotation runai.ai/to-exclude was not removed")
				}, "30s", "5s").Should(Succeed(), "The annotation runai.ai/to-exclude was not removed in time")

				// Retry logic for removing the node exclusion with exponential backoff
				retryErr := retry.OnError(retry.DefaultBackoff, func(err error) bool {
					return errors.IsConflict(err)
				}, func() error {
					return editRunaiConfigAndRemoveNodeExclusion(ctx, kubeClient)
				})
				Expect(retryErr).ToNot(HaveOccurred(), "Failed to remove node exclusion from Run:AI:")
				Eventually(func(g Gomega) {
					node := &v1.Node{}
					err := kubeClient.Get(ctx, client.ObjectKey{Name: workerNodeName}, node)
					g.Expect(err).To(BeNil(), "Failed to get node:")

					delete(node.Labels, "node-to-exclude")
					err = kubeClient.Update(ctx, node)
					g.Expect(err).To(BeNil(), "Failed to update node:")
				}, "30s", "5s").Should(Succeed())
			})

			It("Validate node is excluded from Run:AI", func() {
				// Verify that the node has the correct label
				eventuallyValidateNodeLabel(ctx, kubeClient, workerNodeName, "runai.ai/to-exclude", "true", true)

				managedNodesConfig := &v1alpha1.ManagedNodesConfig{}
				err := kubeClient.Get(ctx, client.ObjectKey{Name: "runai-managed-nodes-config"}, managedNodesConfig)
				Expect(err).To(BeNil())
				Expect(len(managedNodesConfig.Status.Conditions)).To(Equal(1))
				Expect(managedNodesConfig.Status.Conditions[0].Type).To(Equal("Applied"))
				Eventually(func(g Gomega) {
					err := kubeClient.Get(ctx, client.ObjectKey{Name: "runai-managed-nodes-config"}, managedNodesConfig)
					g.Expect(err).To(BeNil(), "Failed to get node:")
					status := managedNodesConfig.Status.Conditions[0].Status
					g.Expect(status).To(BeEquivalentTo(v1.ConditionFalse), "The label value is not as expected")
				}, "30s", "2s").Should(Succeed())
				Expect(managedNodesConfig.Status.Conditions[0].Reason).To(Equal("ToBeExcludedNodes"))
				Expect(managedNodesConfig.Status.Conditions[0].Message).To(Equal("Some nodes need to be drained before gracefull exclusion: " + workerNodeName))
			})

			It("Validate node with scheduled pod and with runai.ai/to-exclude label is not excluded when managed nodes config is updated to include this node", func() {
				err := editRunaiConfigAndRemoveNodeExclusion(ctx, kubeClient)
				Expect(err).To(BeNil())
				podFromCluster := &v1.Pod{}
				err = kubeClient.Get(ctx, client.ObjectKey{Name: testCreatedObjects["test-pod"].GetName(), Namespace: testCreatedObjects["test-pod"].GetNamespace()}, podFromCluster)
				Expect(err).To(BeNil())
				workerNodeName = podFromCluster.ObjectMeta.Annotations["runai-used-nodes"]

				// Eventually runai.ai/to-exclude label should be removed
				eventuallyValidateNodeLabel(ctx, kubeClient, workerNodeName, "runai.ai/to-exclude", "", false)
			})
		})
	})
})
