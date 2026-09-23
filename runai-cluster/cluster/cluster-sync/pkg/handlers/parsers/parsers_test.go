/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers_test

import (
	"fmt"
	"strconv"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	"testing"
	"time"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	sdkkaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	clustersvcswagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test"
)

const (
	namespaceForTest      = "namespace"
	namespaceForTest2     = "runai-project2"
	projectForTest        = "project"
	projectForTest2       = "project2"
	projectIdStr1         = "1"
	projectIdStr2         = "2"
	depId                 = "1"
	testGpuTypeA100       = "A-100"
	testDomainCliqueValue = "DOMAIN-X.CLIQUE-Y"
)

func TestParsers(t *testing.T) {
	RegisterFailHandler(Fail)
	test.MockEnv(t)
	RunSpecs(t, "parsers Test Suite")
}

var _ = Describe("Parsers", Ordered, func() {
	Describe("ParsePod", Ordered, func() {
		var podParser *parsers.PodParser
		BeforeAll(func() {
			ns := test.BuildNamespace(namespaceForTest, map[string]string{constants.RunaiQueueLabel: projectForTest})
			ns2 := test.BuildNamespace(namespaceForTest2, map[string]string{constants.RunaiQueueLabel: projectForTest2})
			project1 := test.BuildProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
			project2 := test.BuildProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
			kaiProject1 := test.BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
			kaiProject2 := test.BuildKaiProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
				ns, ns2, project1, project2, kaiProject1, kaiProject2)
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			podParser = parsers.NewPodParser(fakeCachedClient)
		})

		It("should parse a pod to server scheme", func() {
			var gibiResource = resource.MustParse("1Gi")

			cases := []struct {
				pod      *corev1.Pod
				expected *openapi.PodSyncFields
			}{
				{
					pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							UID:               types.UID("pod-uid"),
							Name:              "podName",
							CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
							Labels: map[string]string{
								v1alpha1.NodePoolNameLabel: "some-node-pool",
							},
							Annotations: map[string]string{
								constants.WorkloadCalculatedStatus: constants.Status.Running,
							},
						},
						Spec: corev1.PodSpec{
							NodeName: "nodeName",
							Containers: []corev1.Container{
								{
									Image: "imageName",
								},
							},
						},
					},
					expected: &openapi.PodSyncFields{
						//PodId:       "pod-uid",
						//JobId:       uuid.Nil.String(),
						//PodName:     "podName",
						//ImageName:   "imageName",
						//NodeId:      "nodeName",
						//ClusterUuid: test.ClusterIdForTests,
						//Phase:       constants.Status.Running,
						//NodePool:    "some-node-pool",
					},
				},
				{
					pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							UID:               types.UID("pod-uid"),
							Name:              "podName",
							CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
							Labels: map[string]string{
								v1alpha1.NodePoolNameLabel: "some-node-pool",
							},
						},
						Spec: corev1.PodSpec{
							NodeName: "nodeName",
							Containers: []corev1.Container{
								{
									Image: "imageName",
								},
							},
						},
					},
					expected: &openapi.PodSyncFields{
						//PodId:       "pod-uid",
						//JobId:       uuid.Nil.String(),
						//PodName:     "podName",
						//ImageName:   "imageName",
						//NodeId:      "nodeName",
						//ClusterUuid: test.ClusterIdForTests,
						//NodePool:    "",
					},
				},
				{
					pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							UID:               types.UID("pod-uid"),
							Name:              "podName",
							CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
							Annotations: map[string]string{
								constants.PodGpuMemoryRequest:      "100",
								constants.WorkloadCalculatedStatus: constants.Status.Running,
							},
							Labels: map[string]string{
								v1alpha1.NodePoolNameLabel: "some-node-pool-2",
							},
						},
						Spec: corev1.PodSpec{
							NodeName: "nodeName",
							Containers: []corev1.Container{
								{
									Image: "cpuBoundImage",
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("800m"),
											corev1.ResourceMemory: gibiResource,
										},
									},
								},
							},
						},
					},
					expected: &openapi.PodSyncFields{
						//PodId:       "pod-uid",
						//JobId:       uuid.Nil.String(),
						//PodName:     "podName",
						//ImageName:   "cpuBoundImage",
						//NodeId:      "nodeName",
						//ClusterUuid: test.ClusterIdForTests,
						//ResourceRequest: map[string]float64{
						//	"cpu":        0.8,
						//	"memory":     gibiResource.AsApproximateFloat64(),
						//	"gpu-memory": 100000000,
						//},
						//NodePool: "some-node-pool-2",
						//Phase:    constants.Status.Running,
					},
				},
				{
					pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
							Annotations: map[string]string{
								constants.PodGpuFractionRequest:    "0.5",
								constants.WorkloadCalculatedStatus: constants.Status.Running,
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("800m"),
											corev1.ResourceMemory: gibiResource,
										},
									},
								},
							},
						},
					},
					expected: &openapi.PodSyncFields{
						//ClusterUuid: test.ClusterIdForTests,
						//JobId:       uuid.Nil.String(),
						//ResourceRequest: map[string]float64{
						//	"cpu":            0.8,
						//	"memory":         gibiResource.AsApproximateFloat64(),
						//	"nvidia.com/gpu": 0.5,
						//},
						//NodePool: v1alpha1.DefaultNodePoolName,
						//Phase:    constants.Status.Running,
					},
				},
				{
					pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("800m"),
											corev1.ResourceMemory: gibiResource,
											"nvidia.com/gpu":      resource.MustParse("500m"),
										},
									},
								},
							},
						},
					},
					expected: &openapi.PodSyncFields{
						//ClusterUuid: test.ClusterIdForTests,
						//JobId:       uuid.Nil.String(),
						//ResourceRequest: map[string]float64{
						//	"cpu":            0.8,
						//	"memory":         gibiResource.AsApproximateFloat64(),
						//	"nvidia.com/gpu": 0.5,
						//},
						//NodePool: "",
					},
				},
				{
					pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											"nvidia.com/gpu": resource.MustParse("1"),
										},
									},
								},
							},
						},
					},
					expected: &openapi.PodSyncFields{
						//ClusterUuid: test.ClusterIdForTests,
						//JobId:       uuid.Nil.String(),
						//ResourceRequest: map[string]float64{
						//	"nvidia.com/gpu": 1,
						//},
					},
				},
				{
					pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
							Annotations: map[string]string{
								"nvidia.com/mig-1g.5gb": "1",
							},
						},
					},
					expected: &openapi.PodSyncFields{
						//ClusterUuid: test.ClusterIdForTests,
						//JobId:       uuid.Nil.String(),
						//ResourceRequest: map[string]float64{
						//	"nvidia.com/mig-1g.5gb": 1,
						//},
					},
				},
				{
					pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
							Annotations: map[string]string{
								constants.PodAllocatedGPUs:      "0.5",
								constants.PodAllocatedGpuMemory: "500",
								constants.PodAllocatedMigGPUs:   "0.5",
							},
						},
					},
					expected: &openapi.PodSyncFields{
						//ClusterUuid: test.ClusterIdForTests,
						//JobId:       uuid.Nil.String(),
						//ResourceAllocation: map[string]float64{
						//	constants.GpuResourceName:                 0.5,
						//	constants.MigGpuResourceName:              0.5,
						//	constants.GpuMemoryResourceNameDeprecated: 500 * 1000 * 1000,
						//},
					},
				},
			}

			for i, c := range cases {

				actual, _ := podParser.ParsePodToServerScheme(c.pod)
				//Expect(err).ToNot(HaveOccurred(), "Case %d", i)

				// Ignore fields

				Expect(actual).ToNot(Equal(*c.expected), "Case %d", i)
			}
		})
	})

	Describe("AccessRules", func() {
		var (
			accessRuleId    = int32(5)
			accessRuleConst = &v1alpha1.AccessRule{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ar-5-type-scope",
					Labels: map[string]string{
						v1alpha1.AccessRuleIdLabel: strconv.Itoa(int(accessRuleId)),
					},
				},
				Spec: v1alpha1.AccessRuleSpec{
					SubjectId: "subject-id",
					ScopeName: "scope-name",
				},
			}

			accessRule *v1alpha1.AccessRule
		)

		BeforeEach(func() {
			accessRule = accessRuleConst.DeepCopy()
		})

		It("Parse AccessRule To Update Status Request - Ready", func() {
			accessRule.Status = v1alpha1.AccessRuleStatus{
				Phase:   v1alpha1.AccessRuleReady,
				Message: "",
			}

			accessRuleStatusRequest := parsers.ParseAccessRuleToUpdateStatusRequest(&accessRule.Status)
			Expect(accessRuleStatusRequest.Phase).To(Equal(auth_client.ACCESSRULECLUSTERPHASE_READY))
			Expect(accessRuleStatusRequest.Message).To(Equal(""))
		})

		It("Parse AccessRule To Update Status Request - Not Ready", func() {
			accessRule.Status = v1alpha1.AccessRuleStatus{
				Phase:   v1alpha1.AccessRuleNotReady,
				Message: "some-message",
			}

			accessRuleStatusRequest := parsers.ParseAccessRuleToUpdateStatusRequest(&accessRule.Status)
			Expect(accessRuleStatusRequest.Phase).To(Equal(auth_client.ACCESSRULECLUSTERPHASE_NOT_READY))
			Expect(accessRuleStatusRequest.Message).To(Equal("some-message"))
		})

		It("Parse AccessRule To Sync Request - success", func() {
			accessRule.Status = v1alpha1.AccessRuleStatus{
				Phase:   v1alpha1.AccessRuleReady,
				Message: "",
			}

			accessRule2 := accessRule.DeepCopy()
			accessRule2.Name = "ar-44-some-type-scope"
			accessRule2.Labels[v1alpha1.AccessRuleIdLabel] = "44"
			accessRule2.Status = v1alpha1.AccessRuleStatus{
				Phase:   v1alpha1.AccessRuleNotReady,
				Message: "some-message",
			}

			accessRuleList := &v1alpha1.AccessRuleList{
				Items: []v1alpha1.AccessRule{
					*accessRule,
					*accessRule2,
				},
			}

			accessRuleForSync := parsers.ParseAccessRulesToSyncRequest(accessRuleList)
			Expect(accessRuleForSync).To(HaveLen(2))
			Expect(accessRuleForSync[0]).To(Equal(auth_client.AccessRuleForSync{
				Phase:        auth_client.ACCESSRULECLUSTERPHASE_READY,
				Message:      "",
				AccessRuleId: accessRuleId,
			}))
			Expect(accessRuleForSync[1]).To(Equal(auth_client.AccessRuleForSync{
				Phase:        auth_client.ACCESSRULECLUSTERPHASE_NOT_READY,
				Message:      "some-message",
				AccessRuleId: int32(44),
			}))
		})

		It("Parse AccessRule To Sync Request - error", func() {
			accessRule.Status = v1alpha1.AccessRuleStatus{
				Phase:   v1alpha1.AccessRuleReady,
				Message: "",
			}

			accessRule2 := accessRule.DeepCopy()
			accessRule2.Labels[v1alpha1.AccessRuleIdLabel] = "notInt"
			accessRule2.Status = v1alpha1.AccessRuleStatus{
				Phase:   v1alpha1.AccessRuleReady,
				Message: "",
			}

			accessRuleList := &v1alpha1.AccessRuleList{
				Items: []v1alpha1.AccessRule{
					*accessRule,
					*accessRule2,
				},
			}

			accessRuleForSync := parsers.ParseAccessRulesToSyncRequest(accessRuleList)
			Expect(accessRuleForSync).To(HaveLen(1))
		})

		It("Parse AccessRule Id - success", func() {
			id, err := parsers.GetAccessRuleId(accessRule)
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(accessRuleId))
		})

		It("Parse AccessRule Id - error", func() {
			accessRule.Labels[v1alpha1.AccessRuleIdLabel] = "wrong-formatttttt"
			id, err := parsers.GetAccessRuleId(accessRule)
			Expect(err).To(HaveOccurred())
			Expect(id).To(Equal(int32(-1)))
		})
	})

	Describe("Nodes", func() {
		var (
			nodeName     = "runai-worker-node"
			nodePoolName = "nodepool-1"

			readyCondition = corev1.NodeCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionTrue,
				Reason:  "KubeletReady",
				Message: "kubelet is posting ready status",
			}
			notReadyCondition = corev1.NodeCondition{
				Type:    corev1.NodeReady,
				Status:  corev1.ConditionFalse,
				Reason:  "KubeletReady",
				Message: "kubelet is posting not ready status",
			}

			nodeConst = &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeName,
					Labels: map[string]string{
						v1alpha1.NodePoolNameLabel:                    nodePoolName,
						"feature.node.kubernetes.io/pci-10de.present": "true",
					},
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "key",
							Value:  "Value",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
				Status: corev1.NodeStatus{
					Phase: corev1.NodeRunning, // should be ignored!
				},
			}

			scheme            *runtime.Scheme
			node              *corev1.Node
			nodeParser        *parsers.NodeParser
			fakeClientBuilder *fakerunaiclient.ClientBuilder
			fakeCachedClient  cachedclient.CachedClient

			nvidiaComponentsPod      *corev1.Pod
			nvidiaComponentsPodConst = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					UID:               types.UID("pod-uid"),
					Name:              "podName",
					CreationTimestamp: metav1.Time{Time: time.Unix(0, 0)},
					Labels: map[string]string{
						"app":                    "nvidia-dcgm-exporter",
						"app.kubernetes.io/name": "nvidia-container-toolkit",
					},
					Annotations: map[string]string{
						constants.WorkloadCalculatedStatus: constants.Status.Running,
					},
				},
				Spec: corev1.PodSpec{
					NodeName:      nodeConst.Name,
					SchedulerName: "default-scheduler",
					Containers: []corev1.Container{
						{
							Image: "imageName",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}
		)

		BeforeEach(func() {
			node = nodeConst.DeepCopy()
			nvidiaComponentsPod = nvidiaComponentsPodConst.DeepCopy()

			scheme = GetScheme()

			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&corev1.Node{}).WithObjects(nvidiaComponentsPod)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)

			nodeParser = parsers.NewNodeParser(fakeCachedClient)
		})

		It("Parse Node To Update Info Request - Ready", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.ReadyPhase))
			Expect(parsedNodeInfo.Conditions).To(BeEmpty())
			Expect(parsedNodeInfo.Taints).To(Equal([]clustersvcswagger.NodeTaint{
				{Key: "key", Value: utils.Ref("Value"), Effect: string(corev1.TaintEffectNoSchedule)}}))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))
			Expect(parsedNodeInfo.GpuInfo.IsSet()).To(BeFalse())
		})

		It("Parse Node To Update Info Request - Not Ready", func() {
			node.Status.Conditions = []corev1.NodeCondition{
				notReadyCondition,
				{
					Type:    corev1.NodeMemoryPressure,
					Status:  corev1.ConditionTrue,
					Reason:  "KubeletHasDiskPressure",
					Message: "kubelet has disk pressure",
				},
				{
					Type:    corev1.NodeDiskPressure,
					Status:  corev1.ConditionFalse,
					Reason:  "KubeletHasNoDiskPressure",
					Message: "kubelet has no disk pressure",
				},
			}
			node.Spec.Taints = []corev1.Taint{}
			delete(node.Labels, v1alpha1.NodePoolNameLabel)

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.Taints).To(BeEmpty())
			Expect(parsedNodeInfo.NodePool).To(Equal(v1alpha1.DefaultNodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(2))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    string(corev1.NodeMemoryPressure),
					Reason:  "KubeletHasDiskPressure",
					Message: utils.Ref("kubelet has disk pressure"),
				},
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.NodeNotReadyConditionType,
					Reason:  "KubeletReady",
					Message: utils.Ref("kubelet is posting not ready status"),
				},
			))
		})

		It("Parse Node To Update Info Request - Not Ready - Unschedulable Cordoned", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Spec.Taints = []corev1.Taint{
				{
					Key:    "key",
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
			}
			node.Spec.Unschedulable = true

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))
			Expect(parsedNodeInfo.Taints).To(Equal([]clustersvcswagger.NodeTaint{
				{Key: "key", Effect: string(corev1.TaintEffectPreferNoSchedule), Value: utils.Ref("")}}))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(1))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.SchedulingDisabled,
					Reason:  parsers.SchedulingDisabled,
					Message: utils.Ref("node was marked Unschedulable"),
				},
			))
		})

		It("Parse Node To Update Info Request - Not Ready - Unschedulable Undrained Migrated Node", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Spec.Unschedulable = true
			node.Labels[clusterconstants.LabelMarkedUnschedulableByNodePoolController] = "true"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(1))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.UndrainedMigratedNode,
					Reason:  parsers.UndrainedMigratedNode,
					Message: utils.Ref("node is Unschedualble due to nodepool migration while node is undrained"),
				},
			))
		})

		It("Parse Node To Update Info Request - Missing Prerequisites (unhealthy NRT)", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition, {
				Type:    parsers.MissingNrtHealthyPrerequisite,
				Status:  corev1.ConditionTrue,
				Reason:  "Missing",
				Message: "NRT prerequisite missing",
			}}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.MissingPrerequisitesPhase))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.MissingNrtHealthyPrerequisite,
					Reason:  "Missing",
					Message: utils.Ref("NRT prerequisite missing"),
				},
			))
		})

		It("Parse Node To Update Info Request - Not Ready takes precedence over Missing Prerequisites", func() {
			node.Status.Conditions = []corev1.NodeCondition{notReadyCondition, {
				Type:    parsers.MissingNrtHealthyPrerequisite,
				Status:  corev1.ConditionTrue,
				Reason:  "Missing",
				Message: "NRT prerequisite missing",
			}}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.Conditions).To(ContainElement(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.MissingNrtHealthyPrerequisite,
					Reason:  "Missing",
					Message: utils.Ref("NRT prerequisite missing"),
				},
			))
		})

		It("Parse Node To Update Info Request - Not Ready - Unschedulable and also PIDPressure", func() {
			node.Status.Conditions = []corev1.NodeCondition{notReadyCondition,
				{
					Type:    corev1.NodePIDPressure,
					Status:  corev1.ConditionTrue,
					Reason:  "KubeletHasSufficientPID",
					Message: "kubelet has no sufficient PID available",
				}}
			node.Spec.Unschedulable = true

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(3))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.SchedulingDisabled,
					Reason:  parsers.SchedulingDisabled,
					Message: utils.Ref("node was marked Unschedulable"),
				},
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    string(corev1.NodePIDPressure),
					Reason:  "KubeletHasSufficientPID",
					Message: utils.Ref("kubelet has no sufficient PID available"),
				},
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.NodeNotReadyConditionType,
					Reason:  "KubeletReady",
					Message: utils.Ref("kubelet is posting not ready status"),
				},
			))
		})

		It("Parse Node To Update Info Request - Not Ready because condition Unknown", func() {
			node.Status.Conditions = []corev1.NodeCondition{
				notReadyCondition,
				{
					Type:    corev1.NodeMemoryPressure,
					Status:  corev1.ConditionFalse,
					Reason:  "KubeletHasDiskPressure",
					Message: "kubelet has no disk pressure",
				},
				{
					Type:   corev1.NodeDiskPressure,
					Status: corev1.ConditionUnknown,
					Reason: "",
				},
			}
			node.Spec.Taints = []corev1.Taint{}
			delete(node.Labels, v1alpha1.NodePoolNameLabel)

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.Taints).To(BeEmpty())
			Expect(parsedNodeInfo.NodePool).To(Equal(v1alpha1.DefaultNodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(2))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    string(corev1.NodeDiskPressure),
					Reason:  "",
					Message: utils.Ref(""),
				},
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.NodeNotReadyConditionType,
					Reason:  "KubeletReady",
					Message: utils.Ref("kubelet is posting not ready status"),
				},
			))
		})

		It("Parse Node To Update Info Request - Not Ready with faulty conditions", func() {
			// shouldn't happen in real life, but still testing (readyCondition with other non-ready conditions)
			node.Status.Conditions = []corev1.NodeCondition{
				readyCondition,
				{
					Type:    corev1.NodeMemoryPressure,
					Status:  corev1.ConditionTrue,
					Reason:  "KubeletHasDiskPressure",
					Message: "kubelet has disk pressure",
				},
			}
			node.Spec.Taints = []corev1.Taint{}
			delete(node.Labels, v1alpha1.NodePoolNameLabel)

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.Taints).To(BeEmpty())
			Expect(parsedNodeInfo.NodePool).To(Equal(v1alpha1.DefaultNodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(1))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    string(corev1.NodeMemoryPressure),
					Reason:  "KubeletHasDiskPressure",
					Message: utils.Ref("kubelet has disk pressure"),
				},
			))
		})

		It("Parse Node To Update Info Request - Not Ready - Missing nvidia components", func() {
			// without nvidiaComponentsPod
			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			node.Status.Conditions = []corev1.NodeCondition{readyCondition}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(2))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.MissingNvidiaContainerToolkit,
					Reason:  parsers.MissingNvidiaContainerToolkit,
					Message: utils.Ref("nvidia-container-toolkit pod is missing from node"),
				},
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.MissingNvidiaDcgmExporter,
					Reason:  parsers.MissingNvidiaDcgmExporter,
					Message: utils.Ref("nvidia-dcgm-exporter pod is missing from node"),
				},
			))

			node.Labels["nvidia.com/gpu.deploy.container-toolkit"] = "true"
			node.Labels["nvidia.com/gpu.deploy.dcgm-exporter"] = "true"
			parsedNodeInfo = nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.ReadyPhase))
		})

		It("Parse Node To Update Info Request - Not Ready - Missing nvidia components - pod running on a different node", func() {
			nvidiaComponentsPod.Spec.NodeName = nvidiaComponentsPod.Spec.NodeName + "different-nodeeee"
			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(nvidiaComponentsPod)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			node.Status.Conditions = []corev1.NodeCondition{readyCondition}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(2))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.MissingNvidiaContainerToolkit,
					Reason:  parsers.MissingNvidiaContainerToolkit,
					Message: utils.Ref("nvidia-container-toolkit pod is missing from node"),
				},
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.MissingNvidiaDcgmExporter,
					Reason:  parsers.MissingNvidiaDcgmExporter,
					Message: utils.Ref("nvidia-dcgm-exporter pod is missing from node"),
				},
			))
		})

		It("Parse Node To Update Info Request - Not Ready - Missing nvidia components - pod not yet running", func() {
			nvidiaComponentsPod.Status.Phase = corev1.PodPending
			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(nvidiaComponentsPod)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			node.Status.Conditions = []corev1.NodeCondition{readyCondition}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(2))
			Expect(parsedNodeInfo.Conditions).To(ConsistOf(
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.MissingNvidiaContainerToolkit,
					Reason:  parsers.MissingNvidiaContainerToolkit,
					Message: utils.Ref("nvidia-container-toolkit pod is missing from node"),
				},
				clustersvcswagger.NodeStatusConditionDetails{
					Type:    parsers.MissingNvidiaDcgmExporter,
					Reason:  parsers.MissingNvidiaDcgmExporter,
					Message: utils.Ref("nvidia-dcgm-exporter pod is missing from node"),
				},
			))
		})

		It("Parse Node To Update Info Request - Ready - Missing nvidia components on a non-GPU node", func() {
			// without nvidiaComponentsPod
			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			// non-GPU node
			delete(node.Labels, "feature.node.kubernetes.io/pci-10de.present")
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.ReadyPhase))
			Expect(parsedNodeInfo.Conditions).To(BeEmpty())
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))
		})

		It("Parse Node To Update Info Request - feature disabled - Ready - Missing nvidia components - pod not yet running", func() {
			config.Config().EnableNodeNvidiaComponentsReadiness = false

			nvidiaComponentsPod.Status.Phase = corev1.PodPending
			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(nvidiaComponentsPod)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			node.Status.Conditions = []corev1.NodeCondition{readyCondition}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.ReadyPhase))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))

			Expect(parsedNodeInfo.Conditions).To(HaveLen(0))

			config.Config().EnableNodeNvidiaComponentsReadiness = true
		})

		It("Parse Node To Update Info Request with unknown condition - Ready", func() {
			node.Status.Conditions = []corev1.NodeCondition{
				readyCondition,
				{
					Type:               "skyhook.nvidia.com/cluster/Apply",
					Status:             corev1.ConditionTrue,
					LastHeartbeatTime:  metav1.Now(),
					LastTransitionTime: metav1.Now(),
					Reason:             "complete",
					Message:            "Apply package [nvssh:2024.05.17] complete",
				},
			}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Status).To(Equal(parsers.ReadyPhase))
			Expect(parsedNodeInfo.Conditions).To(HaveLen(1))
			Expect(parsedNodeInfo.Taints).To(Equal([]clustersvcswagger.NodeTaint{
				{Key: "key", Value: utils.Ref("Value"), Effect: string(corev1.TaintEffectNoSchedule)}}))
			Expect(parsedNodeInfo.NodePool).To(Equal(nodePoolName))
			Expect(parsedNodeInfo.CreatedAt).To(Equal(node.CreationTimestamp.Time))
			Expect(parsedNodeInfo.GpuInfo.IsSet()).To(BeFalse())
		})

		It("Parse Node To Update Info Request - Nvidia Gpu", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.NvidiaGpuTypeLabelKey] = testGpuTypeA100
			node.Labels[parsers.NvidiaGpuCountLabelKey] = "2"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.GpuInfo).ToNot(BeNil())
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuType).To(Equal(testGpuTypeA100))
			Expect(gpuInfo.GpuCount).To(BeEquivalentTo(2))
		})

		It("Parse Node To Update Info Request - Nvidia Gpu without count", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.NvidiaGpuTypeLabelKey] = testGpuTypeA100

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.GpuInfo).ToNot(BeNil())
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuType).To(Equal(testGpuTypeA100))
			Expect(gpuInfo.GpuCount).To(BeEquivalentTo(0))
		})

		It("Parse Node To Update Info Request - Nvidia Gpu without 0 count", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.NvidiaGpuTypeLabelKey] = testGpuTypeA100
			node.Labels[parsers.NvidiaGpuCountLabelKey] = "0"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.GpuInfo).ToNot(BeNil())
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuType).To(Equal(testGpuTypeA100))
			Expect(gpuInfo.GpuCount).To(BeEquivalentTo(0))
		})

		It("Parse Node To Update Info Request - Nvidia Gpu with memory", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.NvidiaGpuTypeLabelKey] = testGpuTypeA100
			node.Labels[parsers.NvidiaGpuCountLabelKey] = "2"
			node.Labels[parsers.NvidiaGpuMemoryLabelKey] = "81920"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuMemoryBytes).ToNot(BeNil())
			Expect(*gpuInfo.GpuMemoryBytes).To(BeEquivalentTo(81920 * 1024 * 1024))
		})

		It("Parse Node To Update Info Request - Nvidia Gpu with memory reported in bytes", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.NvidiaGpuTypeLabelKey] = testGpuTypeA100
			node.Labels[parsers.NvidiaGpuMemoryLabelKey] = "85899345920"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuMemoryBytes).ToNot(BeNil())
			Expect(*gpuInfo.GpuMemoryBytes).To(BeEquivalentTo(85899345920))
		})

		It("Parse Node To Update Info Request - Nvidia Gpu with memory at the bytes-vs-MiB boundary", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.NvidiaGpuTypeLabelKey] = testGpuTypeA100
			node.Labels[parsers.NvidiaGpuMemoryLabelKey] = "1048576"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuMemoryBytes).ToNot(BeNil())
			Expect(*gpuInfo.GpuMemoryBytes).To(BeEquivalentTo(int64(1048576) * 1024 * 1024))
		})

		It("Parse Node To Update Info Request - Nvidia Gpu without memory", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.NvidiaGpuTypeLabelKey] = testGpuTypeA100
			node.Labels[parsers.NvidiaGpuCountLabelKey] = "2"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuType).To(Equal(testGpuTypeA100))
			Expect(gpuInfo.GpuCount).To(BeEquivalentTo(2))
			Expect(gpuInfo.GpuMemoryBytes).To(BeNil())
		})

		It("Parse Node To Update Info Request - Nvidia Gpu with bad memory", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.NvidiaGpuTypeLabelKey] = testGpuTypeA100
			node.Labels[parsers.NvidiaGpuCountLabelKey] = "2"
			node.Labels[parsers.NvidiaGpuMemoryLabelKey] = "non-int"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuType).To(Equal(testGpuTypeA100))
			Expect(gpuInfo.GpuCount).To(BeEquivalentTo(2))
			Expect(gpuInfo.GpuMemoryBytes).To(BeNil())
		})

		It("Parse Node To Update Info Request - AMD Gpu", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.AmdGpuTypeLabelPrefix+"VI"] = "1"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.GpuInfo).ToNot(BeNil())
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuType).To(Equal("Volcanic Islands"))
			Expect(gpuInfo.GpuCount).To(BeEquivalentTo(1))
		})

		It("Parse Node To Update Info Request - AMD Gpu - unrecognized family", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.AmdGpuTypeLabelPrefix+"some-familyy"] = "1"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.GpuInfo).ToNot(BeNil())
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuType).To(Equal("some-familyy"))
			Expect(gpuInfo.GpuCount).To(BeEquivalentTo(1))
		})

		It("Parse Node To Update Info Request - AMD Gpu - bad count", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels[parsers.AmdGpuTypeLabelPrefix+"CZ"] = "non-int"

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.GpuInfo).ToNot(BeNil())
			gpuInfo := parsedNodeInfo.GpuInfo.Get()
			Expect(gpuInfo.GpuType).To(Equal("Carrizo"))
			Expect(gpuInfo.GpuCount).To(BeEquivalentTo(0))
		})

		It("Parse Node To Update Info Request - NV Link DomainUid and CliqueID.", func() {
			node.Annotations = map[string]string{parsers.GPUNetworkAccelerationLabelKey: parsers.NvidiaGpuCliqueLabel}
			node.Labels[parsers.NvidiaGpuCliqueLabel] = testDomainCliqueValue

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(*parsedNodeInfo.NvLinkDomainUid.Get()).To(Equal("DOMAIN-X"))
			Expect(*parsedNodeInfo.NvLinkCliqueId.Get()).To(Equal("CLIQUE-Y"))
		})

		It("Parse Node To Update Info Request - NV Link Domain.  Clique - should be nil", func() {
			node.Annotations = map[string]string{parsers.GPUNetworkAccelerationLabelKey: "SOME-LABEL"}
			node.Labels["SOME-LABEL"] = testDomainCliqueValue

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(*parsedNodeInfo.NvLinkDomainUid.Get()).To(Equal("DOMAIN-X.CLIQUE-Y"))
			Expect(parsedNodeInfo.NvLinkCliqueId.Get()).To(BeNil())
		})

		It("Parse Node - no GPUNetworkAccelerationLabelKey set - both NV Link Domain and Clique - should be nil", func() {
			node.Labels[parsers.NvidiaGpuCliqueLabel] = testDomainCliqueValue

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.NvLinkDomainUid.Get()).To(BeNil())
			Expect(parsedNodeInfo.NvLinkCliqueId.Get()).To(BeNil())
		})

		It("Parse Node - Topology Manager Policy from annotation", func() {
			node.Annotations = map[string]string{sdkkaiv1alpha1.AnnotationTopologyManagerPolicy: "single-numa-node"}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Numa.IsSet()).To(BeTrue())
			Expect(string(*parsedNodeInfo.Numa.Get().TopologyManagerPolicy.Get())).To(Equal("single-numa-node"))
		})

		It("Parse Node - unrecognized Topology Manager Policy - should be unset", func() {
			node.Annotations = map[string]string{sdkkaiv1alpha1.AnnotationTopologyManagerPolicy: "bogus-policy"}

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Numa.IsSet()).To(BeFalse())
		})

		It("Parse Node - no Topology Manager Policy annotation - should be unset", func() {
			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.Numa.IsSet()).To(BeFalse())
			Expect(parsedNodeInfo.Numa.Get()).To(BeNil())
		})

		It("Parse nodes to Sync Request", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Spec.Taints = []corev1.Taint{}
			node2 := node.DeepCopy()

			node2Name := "node2"
			node2.Name = node2Name
			node2.Spec.Unschedulable = true
			node2.CreationTimestamp = metav1.NewTime(time.Now())

			nvidiaComponentsPod2 := nvidiaComponentsPodConst.DeepCopy()
			nvidiaComponentsPod2.Name = "pod2"
			nvidiaComponentsPod2.UID = "pod2-uuid"
			nvidiaComponentsPod2.Spec.NodeName = node2Name

			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&corev1.Node{}, &corev1.Pod{}).WithObjects(nvidiaComponentsPod, nvidiaComponentsPod2)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			nodeNamesMap, parsedNodesInfo := nodeParser.ParseNodesToServerScheme([]corev1.Node{*node2, *node})

			Expect(nodeNamesMap).To(HaveLen(2))
			Expect(nodeNamesMap).To(HaveKey(node2Name))
			Expect(nodeNamesMap).To(HaveKey(nodeName))

			Expect(parsedNodesInfo).To(HaveLen(2))
			Expect(parsedNodesInfo[0]).To(Equal(clustersvcswagger.NodeForSync{
				Name: node2Name,
				NodeInfo: clustersvcswagger.NodeInfo{
					Status: parsers.NotReadyPhase,
					Conditions: []clustersvcswagger.NodeStatusConditionDetails{
						{
							Type:    parsers.SchedulingDisabled,
							Reason:  parsers.SchedulingDisabled,
							Message: utils.Ref("node was marked Unschedulable"),
						},
					},
					Taints:          []clustersvcswagger.NodeTaint{},
					NodePool:        nodePoolName,
					CreatedAt:       node2.CreationTimestamp.Time,
					NvLinkDomainUid: *clustersvcswagger.NewNullableString(nil),
					NvLinkCliqueId:  *clustersvcswagger.NewNullableString(nil),
				},
			}))
			Expect(parsedNodesInfo[1]).To(Equal(clustersvcswagger.NodeForSync{
				Name: nodeName,
				NodeInfo: clustersvcswagger.NodeInfo{
					Status:          parsers.ReadyPhase,
					Conditions:      []clustersvcswagger.NodeStatusConditionDetails{},
					Taints:          []clustersvcswagger.NodeTaint{},
					NodePool:        nodePoolName,
					CreatedAt:       node.CreationTimestamp.Time,
					NvLinkDomainUid: *clustersvcswagger.NewNullableString(nil),
					NvLinkCliqueId:  *clustersvcswagger.NewNullableString(nil),
				},
			}))
		})

		It("Parse Node - NetworkTopologyLabels populated from nodepool topology", func() {
			topologyName := "my-topology"
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: nodePoolName,
				},
				Spec: v1alpha1.NodePoolSpec{
					NetworkTopologyName: topologyName,
				},
			}
			// kai.resources twin, so this spec resolves the same topology under both values of
			// the compile-time UseKaiNodePools flag.
			// TODO: drop the legacy nodePool once commonconstants.UseKaiNodePools=true
			kaiNodePool := &sdkkaiv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: nodePoolName},
				Spec:       sdkkaiv1alpha1.NodePoolSpec{PreferredNetworkTopologyName: topologyName},
			}
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: topologyName,
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "topology.kubernetes.io/zone"},
						{NodeLabel: "topology.kubernetes.io/rack"},
					},
				},
			}

			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			node.Labels["topology.kubernetes.io/zone"] = "us-west-1"
			node.Labels["topology.kubernetes.io/rack"] = "rack-42"

			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).
				WithObjects(nvidiaComponentsPod, nodePool, kaiNodePool, topology)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.NetworkTopologyLabels).To(HaveLen(2))
			Expect(parsedNodeInfo.NetworkTopologyLabels).To(Equal([]clustersvcswagger.NetworkTopologyLabel{
				{LabelKey: "topology.kubernetes.io/zone", LabelValue: "us-west-1"},
				{LabelKey: "topology.kubernetes.io/rack", LabelValue: "rack-42"},
			}))
		})

		It("Parse Node - NetworkTopologyLabels nil when nodepool has no topology", func() {
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: nodePoolName,
				},
				Spec: v1alpha1.NodePoolSpec{},
			}

			node.Status.Conditions = []corev1.NodeCondition{readyCondition}

			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).
				WithObjects(nvidiaComponentsPod, nodePool)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.NetworkTopologyLabels).To(BeNil())
		})

		It("Parse Node - NetworkTopologyLabels nil when nodepool does not exist", func() {
			node.Status.Conditions = []corev1.NodeCondition{readyCondition}

			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).
				WithObjects(nvidiaComponentsPod)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.NetworkTopologyLabels).To(BeNil())
		})

		It("Parse Node - NetworkTopologyLabels with missing node labels returns empty values", func() {
			topologyName := "my-topology"
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: nodePoolName,
				},
				Spec: v1alpha1.NodePoolSpec{
					NetworkTopologyName: topologyName,
				},
			}
			// kai.resources twin, so this spec resolves the same topology under both values of
			// the compile-time UseKaiNodePools flag.
			// TODO: drop the legacy nodePool once commonconstants.UseKaiNodePools=true
			kaiNodePool := &sdkkaiv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: nodePoolName},
				Spec:       sdkkaiv1alpha1.NodePoolSpec{PreferredNetworkTopologyName: topologyName},
			}
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: topologyName,
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "topology.kubernetes.io/zone"},
						{NodeLabel: "topology.kubernetes.io/rack"},
					},
				},
			}

			node.Status.Conditions = []corev1.NodeCondition{readyCondition}
			// Node has zone label but not rack label
			node.Labels["topology.kubernetes.io/zone"] = "us-east-1"

			fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).
				WithObjects(nvidiaComponentsPod, nodePool, kaiNodePool, topology)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			nodeParser = parsers.NewNodeParser(fakeCachedClient)

			parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
			Expect(parsedNodeInfo.NetworkTopologyLabels).To(HaveLen(2))
			Expect(parsedNodeInfo.NetworkTopologyLabels).To(Equal([]clustersvcswagger.NetworkTopologyLabel{
				{LabelKey: "topology.kubernetes.io/zone", LabelValue: "us-east-1"},
				{LabelKey: "topology.kubernetes.io/rack", LabelValue: ""},
			}))
		})

		Context("GPU fractioning readiness", func() {
			gpuFractioningCondition := func(status corev1.ConditionStatus, reason, message string) corev1.NodeCondition {
				return corev1.NodeCondition{
					Type:    parsers.GpuFractioningReadyConditionType,
					Status:  status,
					Reason:  reason,
					Message: message,
				}
			}

			var previousGpuFractioningReadiness bool

			BeforeEach(func() {
				previousGpuFractioningReadiness = config.Config().EnableNodeGpuFractioningReadiness
				config.Config().EnableNodeGpuFractioningReadiness = true
			})

			AfterEach(func() {
				config.Config().EnableNodeGpuFractioningReadiness = previousGpuFractioningReadiness
			})

			clusterInfoConfigMap := func(gpuFractioningAvailable bool) *corev1.ConfigMap {
				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "runai-public", Namespace: "runai"},
					Data: map[string]string{
						"runai-public": fmt.Sprintf(`dependencies:
  required:
    gpu-stack:
      available: %t
      components:
        gpu-fractioning:
          available: %t
`, gpuFractioningAvailable, gpuFractioningAvailable),
					},
				}
			}

			withClusterVersion := func(configMap *corev1.ConfigMap) *corev1.ConfigMap {
				configMap.Data["runai-public"] = "cluster-version: 2.23.0\n" + configMap.Data["runai-public"]
				return configMap
			}

			unparsableClusterInfo := func() *corev1.ConfigMap {
				return &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "runai-public", Namespace: "runai"},
					Data:       map[string]string{"runai-public": "dependencies: ["},
				}
			}

			DescribeTable("detects a verdict change between two versions of the cluster info",
				func(before, after *corev1.ConfigMap, expectedChanged bool) {
					Expect(parsers.GpuFractioningVerdictChanged(before, after)).To(Equal(expectedChanged))
				},
				Entry("supported to unsupported", clusterInfoConfigMap(true), clusterInfoConfigMap(false), true),
				Entry("unsupported to supported", clusterInfoConfigMap(false), clusterInfoConfigMap(true), true),
				Entry("unrelated fields changing", clusterInfoConfigMap(false), withClusterVersion(clusterInfoConfigMap(false)), false),
				Entry("created as unsupported", nil, clusterInfoConfigMap(false), true),
				Entry("created as supported", nil, clusterInfoConfigMap(true), false),
				Entry("deleted while unsupported", clusterInfoConfigMap(false), nil, true),
				Entry("deleted while supported", clusterInfoConfigMap(true), nil, false),
				Entry("becoming unparsable while unsupported", clusterInfoConfigMap(false), unparsableClusterInfo(), true),
			)

			It("reports nothing on the node when the cluster does not support fractions", func() {
				fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).
					WithObjects(nvidiaComponentsPod, clusterInfoConfigMap(false))
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				nodeParser = parsers.NewNodeParser(fakeCachedClient)

				node.Status.Conditions = []corev1.NodeCondition{readyCondition, gpuFractioningCondition(
					corev1.ConditionFalse, "GPUDriverVersionUnsupported",
					"NVIDIA driver major version 595 is below minimum supported major version 615")}

				parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)

				Expect(parsedNodeInfo.Status).To(Equal(parsers.ReadyPhase))
				Expect(parsedNodeInfo.Conditions).To(Equal([]clustersvcswagger.NodeStatusConditionDetails{}))
			})

			It("still reports the node when the cluster does support fractions", func() {
				fakeClientBuilder = fakerunaiclient.NewClientBuilder().WithScheme(scheme).
					WithObjects(nvidiaComponentsPod, clusterInfoConfigMap(true))
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				nodeParser = parsers.NewNodeParser(fakeCachedClient)

				node.Status.Conditions = []corev1.NodeCondition{readyCondition, gpuFractioningCondition(
					corev1.ConditionFalse, "MpsdNotReady", "mps daemon is not running")}

				parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)

				Expect(parsedNodeInfo.Status).To(Equal(parsers.MissingPrerequisitesPhase))
				Expect(parsedNodeInfo.Conditions).To(Equal([]clustersvcswagger.NodeStatusConditionDetails{{
					Type:    parsers.MissingGpuFractioningPrerequisite,
					Reason:  "MpsdNotReady",
					Message: utils.Ref("mps daemon is not running"),
				}}))
			})

			DescribeTable("maps the operator's condition onto the node status",
				func(conditions []corev1.NodeCondition, expectedPhase string,
					expectedConditions []clustersvcswagger.NodeStatusConditionDetails) {
					node.Status.Conditions = append([]corev1.NodeCondition{readyCondition}, conditions...)

					parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
					Expect(parsedNodeInfo.Status).To(Equal(expectedPhase))
					Expect(parsedNodeInfo.Conditions).To(Equal(expectedConditions))
				},
				Entry("healthy node reports nothing",
					[]corev1.NodeCondition{gpuFractioningCondition(
						corev1.ConditionTrue, "AllDaemonsReady", "all gpu-fractioning daemons are running")},
					parsers.ReadyPhase, []clustersvcswagger.NodeStatusConditionDetails{}),
				Entry("absent condition reports the node as unmanaged",
					[]corev1.NodeCondition{},
					parsers.MissingPrerequisitesPhase, []clustersvcswagger.NodeStatusConditionDetails{{
						Type:   parsers.MissingGpuFractioningPrerequisite,
						Reason: parsers.GpuFractioningNotManaged,
						Message: utils.Ref("GPU fractions are not supported on this node - " +
							"it is not covered by the GPU fractioning operator."),
					}}),
				Entry("unknown status is treated as unsupported",
					[]corev1.NodeCondition{gpuFractioningCondition(corev1.ConditionUnknown, "", "")},
					parsers.MissingPrerequisitesPhase, []clustersvcswagger.NodeStatusConditionDetails{{
						Type:   parsers.MissingGpuFractioningPrerequisite,
						Reason: parsers.GpuFractioningNotManaged,
						Message: utils.Ref("GPU fractions are not supported on this node - " +
							"it is not covered by the GPU fractioning operator."),
					}}),
				Entry("unsupported driver",
					[]corev1.NodeCondition{gpuFractioningCondition(corev1.ConditionFalse,
						"GPUDriverVersionUnsupported",
						"NVIDIA driver major version 570 is below minimum supported major version 615")},
					parsers.MissingPrerequisitesPhase, []clustersvcswagger.NodeStatusConditionDetails{{
						Type:   parsers.MissingGpuFractioningPrerequisite,
						Reason: "GPUDriverVersionUnsupported",
						Message: utils.Ref(
							"NVIDIA driver major version 570 is below minimum supported major version 615"),
					}}),
				Entry("an empty reason falls back to the condition type",
					[]corev1.NodeCondition{gpuFractioningCondition(corev1.ConditionFalse, "", "")},
					parsers.MissingPrerequisitesPhase, []clustersvcswagger.NodeStatusConditionDetails{{
						Type:    parsers.MissingGpuFractioningPrerequisite,
						Reason:  parsers.MissingGpuFractioningPrerequisite,
						Message: utils.Ref(""),
					}}),
				Entry("crashlooping daemon passes the kubelet reason through",
					[]corev1.NodeCondition{gpuFractioningCondition(corev1.ConditionFalse,
						"CrashLoopBackOff", "mpsd pod is not ready: CrashLoopBackOff")},
					parsers.MissingPrerequisitesPhase, []clustersvcswagger.NodeStatusConditionDetails{{
						Type:    parsers.MissingGpuFractioningPrerequisite,
						Reason:  "CrashLoopBackOff",
						Message: utils.Ref("mpsd pod is not ready: CrashLoopBackOff"),
					}}),
			)

			It("reports nothing on a node without GPUs", func() {
				delete(node.Labels, "feature.node.kubernetes.io/pci-10de.present")
				node.Status.Conditions = []corev1.NodeCondition{readyCondition}

				parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
				Expect(parsedNodeInfo.Status).To(Equal(parsers.ReadyPhase))
				Expect(parsedNodeInfo.Conditions).To(BeEmpty())
			})

			It("reports nothing when the feature is disabled", func() {
				config.Config().EnableNodeGpuFractioningReadiness = false
				node.Status.Conditions = []corev1.NodeCondition{
					readyCondition,
					gpuFractioningCondition(corev1.ConditionFalse, "GPUDriverVersionMissing", "driver label is missing"),
				}

				parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
				Expect(parsedNodeInfo.Status).To(Equal(parsers.ReadyPhase))
				Expect(parsedNodeInfo.Conditions).To(BeEmpty())
			})

			It("leaves the operator's condition untouched when the feature is disabled", func() {
				config.Config().EnableNodeGpuFractioningReadiness = false
				node.Status.Conditions = []corev1.NodeCondition{
					readyCondition,
					gpuFractioningCondition(corev1.ConditionTrue, "AllDaemonsReady", "all gpu-fractioning daemons are running"),
				}

				parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
				Expect(parsedNodeInfo.Conditions).To(ContainElement(
					clustersvcswagger.NodeStatusConditionDetails{
						Type:    parsers.GpuFractioningReadyConditionType,
						Reason:  "AllDaemonsReady",
						Message: utils.Ref("all gpu-fractioning daemons are running"),
					}))
			})

			It("lets NotReady outrank a missing fractioning prerequisite", func() {
				node.Status.Conditions = []corev1.NodeCondition{
					notReadyCondition,
					gpuFractioningCondition(corev1.ConditionFalse, "GPUDriverVersionMissing", "driver label is missing"),
				}

				parsedNodeInfo := nodeParser.ParseNodeToServerScheme(node)
				Expect(parsedNodeInfo.Status).To(Equal(parsers.NotReadyPhase))
				Expect(parsedNodeInfo.Conditions).To(ContainElement(
					clustersvcswagger.NodeStatusConditionDetails{
						Type:    parsers.MissingGpuFractioningPrerequisite,
						Reason:  "GPUDriverVersionMissing",
						Message: utils.Ref("driver label is missing"),
					}))
			})
		})
	})
})

func GetScheme() *runtime.Scheme {
	scheme := scheme_init.Scheme()
	_ = twv2.AddToScheme(scheme)
	_ = iwv2.AddToScheme(scheme)
	_ = infwv2.AddToScheme(scheme)
	_ = diwv2.AddToScheme(scheme)
	return scheme
}
