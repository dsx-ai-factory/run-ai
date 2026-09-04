/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package requested_nodepools_converter_test

import (
	"context"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/requested_nodepools_converter"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/testbuilders"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/node-pool-utils/converter"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestRequestedNodepoolsConverter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RequestedNodepoolsConverter Suite")
}

var _ = BeforeSuite(func() {
	// Configure the package-level config with the runai vocabulary so that
	// the assigner-side code (which reads from config.Config()) sees the
	// same identifiers the test data assumes.
	DeferCleanup(config.SetForTest(config.PodGroupAssignerConfig{
		NodePoolLabelKey:           testNodePoolLabelKey,
		DefaultNodepoolName:        testDefaultNodePoolName,
		UnexistingNodepoolSentinel: testUnexistingNodePool,
		AnnotationNodepoolsKey:     AnnotationNodepools,
	}))
})

const (
	testPodGroupName = "test-pod-group"
	testNamespace    = "runai-team-a"

	key1  = "key1"
	key2  = "key2"
	key3  = "key3"
	key30 = "key3-0"
	key4  = "key4"
	key5  = "key5"
	val1  = "val1"
	val10 = "val1-0"
	val2  = "val2"
	val21 = "val2-1"
	val20 = "val2-0"
	val3  = "val3"
	val4  = "val4"
	val5  = "val5"

	AnnotationNodepools = "test-nodepools"

	testNodePoolLabelKey    = "test/node-pool"
	testDefaultNodePoolName = "default"
	testUnexistingNodePool  = "test-unexisting-node-pool"
)

type testNodePool struct {
	name       string
	labelKey   string
	labelValue string
	phase      v1alpha1.NodePoolPhase
}

func (tnp testNodePool) toObjects() []client.Object {
	return []client.Object{
		testbuilders.BuildNodePool(tnp.name, tnp.labelKey, tnp.labelValue, tnp.phase),
	}
}

var (
	defaultNodePool  = testNodePool{testDefaultNodePoolName, "", "", v1alpha1.NodePoolReady}
	nodePoolA        = testNodePool{"node-pool-a", key1, val1, v1alpha1.NodePoolReady}
	nodePoolA0       = testNodePool{"node-pool-a-0", key1, val10, v1alpha1.NodePoolReady}
	nodePoolB        = testNodePool{"node-pool-b", key2, val2, v1alpha1.NodePoolReady}
	nodePoolB0       = testNodePool{"node-pool-b-0", key2, val20, v1alpha1.NodePoolReady}
	nodePoolB1       = testNodePool{"node-pool-b-1", key2, val21, v1alpha1.NodePoolReady}
	nodePoolC        = testNodePool{"node-pool-c", key3, val3, v1alpha1.NodePoolReady}
	nodePoolC0       = testNodePool{"node-pool-c-0", key30, val3, v1alpha1.NodePoolReady}
	nodePoolEmpty    = testNodePool{"node-pool-d-0", key4, val4, v1alpha1.NodePoolEmpty}
	nodePoolUnsch    = testNodePool{"node-pool-d-1", key4, val5, v1alpha1.NodePoolUnschedulable}
	nodePoolDeleting = testNodePool{"node-pool-d-2", key5, val4, v1alpha1.NodePoolDeleting}

	allTestNodePools = []testNodePool{
		defaultNodePool,
		nodePoolA, nodePoolA0, nodePoolB, nodePoolB0, nodePoolB1,
		nodePoolC, nodePoolC0, nodePoolEmpty, nodePoolUnsch, nodePoolDeleting,
	}
)

func newScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func newFakeClientWith(nps []testNodePool) client.Client {
	objs := make([]client.Object, 0, len(nps)*2)
	for _, np := range nps {
		objs = append(objs, np.toObjects()...)
	}
	return fakeclient.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(objs...).
		Build()
}

func buildPodList(numberOfPods int, matchExpressionss [][]corev1.NodeSelectorRequirement) *corev1.PodList {
	list := &corev1.PodList{}
	for i := 0; i < numberOfPods; i++ {
		pod := testbuilders.BuildPodWithNodeAffinity(
			testPodGroupName+"-"+strconv.Itoa(i),
			testPodGroupName,
			testNamespace,
			matchExpressionss,
		)
		list.Items = append(list.Items, *pod)
	}
	return list
}

type npcCase struct {
	name                         string
	numberOfPods                 int
	nodeMatchExpressionssForPod  [][]corev1.NodeSelectorRequirement
	podGroupLabels               map[string]string
	expectedNodePoolsForPodGroup []string
	expectError                  bool
	// onlyDefaultNodePool: if true, the converter sees a cluster with only the
	// default node pool (i.e., no seeded NodePool CRs at all).
	onlyDefaultNodePool bool
}

var _ = Describe("GetRequestedNodePoolsForPodGroup", func() {
	DescribeTable("returns the expected node pool options",
		func(tc *npcCase) {
			var fakeClient client.Client
			if tc.onlyDefaultNodePool {
				fakeClient = newFakeClientWith([]testNodePool{defaultNodePool})
			} else {
				fakeClient = newFakeClientWith(allTestNodePools)
			}

			converter := requested_nodepools_converter.NewRequestedNodePoolsConverter(fakeClient, converter.NodePoolIdentifiers{
				NodePoolAssignmentLabelKey: config.Config().NodePoolLabelKey,
				DefaultNodepoolName:        config.Config().DefaultNodepoolName,
				UnexistingNodepoolSentinel: config.Config().UnexistingNodepoolSentinel,
				AnnotationNodepoolsKey:     config.Config().AnnotationNodepoolsKey,
			})

			podGroupMeta := metav1.ObjectMeta{
				Name:      testPodGroupName,
				Namespace: testNamespace,
				Labels:    tc.podGroupLabels,
			}
			if podGroupMeta.Labels == nil {
				podGroupMeta.Labels = map[string]string{}
			}

			podList := buildPodList(tc.numberOfPods, tc.nodeMatchExpressionssForPod)

			result, err := converter.GetRequestedNodePoolsForPodGroup(context.Background(), podGroupMeta, podList)
			if tc.expectError {
				Expect(err).To(HaveOccurred(), "Test: <%s>", tc.name)
				return
			}
			Expect(err).NotTo(HaveOccurred(), "Test: <%s>", tc.name)
			Expect(result).To(ConsistOf(tc.expectedNodePoolsForPodGroup), "Test: <%s>", tc.name)
		},

		Entry("Sanity", &npcCase{
			name:                         "Sanity",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{nodePoolA.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key1, Operator: corev1.NodeSelectorOpIn, Values: []string{val1}}},
			},
		}),
		Entry("Only default node pool affinity", &npcCase{
			name:                         "Only default node pool affinity",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{testDefaultNodePoolName},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: testNodePoolLabelKey, Operator: corev1.NodeSelectorOpDoesNotExist}},
			},
		}),
		Entry("default node pool one of options in affinity", &npcCase{
			name:                         "default node pool one of options in affinity",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{nodePoolB.name, testDefaultNodePoolName, nodePoolC.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key2, Operator: corev1.NodeSelectorOpIn, Values: []string{val2}}},
				{{Key: testNodePoolLabelKey, Operator: corev1.NodeSelectorOpDoesNotExist}},
				{{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{val3}}},
			},
		}),
		Entry("No pods - error, use nothing", &npcCase{
			name:                         "No pods - error, use nothing",
			numberOfPods:                 0,
			expectedNodePoolsForPodGroup: []string{},
			expectError:                  true,
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key1, Operator: corev1.NodeSelectorOpIn, Values: []string{val1}}},
			},
		}),
		Entry("Pods exist - but no matching affinity - use default", &npcCase{
			name:                         "Pods exist - but no matching affinity - use default",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{testDefaultNodePoolName},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{val1}}},
			},
		}),
		Entry("Pods exist - but no affinity at all - use default", &npcCase{
			name:                         "Pods exist - but no affinity at all - use default",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{testDefaultNodePoolName},
			nodeMatchExpressionssForPod:  nil,
		}),
		Entry("Wrong operator - parse only the other node pool", &npcCase{
			name:                         "Wrong operator - parse only the other node pool",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{nodePoolB.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key3, Operator: corev1.NodeSelectorOpNotIn, Values: []string{val3}}},
				{{Key: key1, Operator: corev1.NodeSelectorOpExists}},
				{{Key: key2, Operator: corev1.NodeSelectorOpIn, Values: []string{val2}}},
			},
		}),
		Entry("Wrong operator - use default", &npcCase{
			name:                         "Wrong operator - use default",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{testDefaultNodePoolName},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key1, Operator: corev1.NodeSelectorOpNotIn, Values: []string{key1}}},
				{{Key: key2, Operator: corev1.NodeSelectorOpExists}},
				{{Key: key2, Operator: corev1.NodeSelectorOpDoesNotExist}},
				{{Key: key2, Operator: corev1.NodeSelectorOpGt, Values: []string{key1}}},
				{{Key: key2, Operator: corev1.NodeSelectorOpLt, Values: []string{key1}}},
			},
		}),
		Entry("Non-existing node pool in affinity - use default", &npcCase{
			name:                         "Non-existing node pool in affinity - use default",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{testDefaultNodePoolName},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: "non-existing-key", Operator: corev1.NodeSelectorOpIn, Values: []string{"non-existing-val", "non-existing-val2"}}},
			},
		}),
		Entry("Key does not exist - parse only the other node pool", &npcCase{
			name:                         "Key does not exist - parse only the other node pool",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{nodePoolC.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: "non-existing-key", Operator: corev1.NodeSelectorOpIn, Values: []string{val3}}},
				{{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{val3}}},
			},
		}),
		Entry("Value does not exist - parse only the other node pool", &npcCase{
			name:                         "Value does not exist - parse only the other node pool",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{nodePoolC.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{"non-existing-val"}}},
				{{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{val3}}},
			},
		}),
		Entry("Non ready node pool - still matches - except deleting", &npcCase{
			name:                         "Non ready node pool - still matches - except deleting",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{nodePoolEmpty.name, nodePoolUnsch.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key4, Operator: corev1.NodeSelectorOpIn, Values: []string{val4, val5}}},
				{{Key: key5, Operator: corev1.NodeSelectorOpIn, Values: []string{val4}}},
			},
		}),
		Entry("Same key - different values", &npcCase{
			name:                         "Same key - different values",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{nodePoolB0.name, nodePoolB.name, nodePoolB1.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key2, Operator: corev1.NodeSelectorOpIn, Values: []string{val2, val21, "non-existing-val", val20}}},
			},
		}),
		Entry("Same value - different keys", &npcCase{
			name:                         "Same value - different keys",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{nodePoolC0.name, nodePoolC.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{"non-existing-val", val3}}},
				{{Key: key30, Operator: corev1.NodeSelectorOpIn, Values: []string{val3}}},
			},
		}),
		Entry("Multiple match expressions - only consider the first - but consider the other nodeSelectorTerm", &npcCase{
			name:                         "Multiple match expressions - only consider the first - but consider the other nodeSelectorTerm",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{nodePoolB.name, nodePoolA.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{
					{Key: key2, Operator: corev1.NodeSelectorOpIn, Values: []string{val2}},
					{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{val3}},
				},
				{
					{Key: key1, Operator: corev1.NodeSelectorOpIn, Values: []string{val1}},
					{Key: key1, Operator: corev1.NodeSelectorOpIn, Values: []string{val10}},
				},
			},
		}),
		Entry("No affinity at all - use single node pool from label", &npcCase{
			name:                         "No affinity at all - use single node pool from label",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{nodePoolA.name},
			nodeMatchExpressionssForPod:  nil,
			podGroupLabels:               map[string]string{testNodePoolLabelKey: nodePoolA.name},
		}),
		Entry("No affinity and no label - use default", &npcCase{
			name:                         "No affinity and no label - use default",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{testDefaultNodePoolName},
			nodeMatchExpressionssForPod:  nil,
			podGroupLabels:               map[string]string{},
		}),
		Entry("Duplicate node pools in node affinity", &npcCase{
			name:                         "Duplicate node pools in node affinity",
			numberOfPods:                 1,
			expectedNodePoolsForPodGroup: []string{nodePoolC.name, nodePoolB.name},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key2, Operator: corev1.NodeSelectorOpIn, Values: []string{val2, val2}}},
				{{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{val3}}},
				{{Key: key2, Operator: corev1.NodeSelectorOpIn, Values: []string{val2}}},
				{{Key: key3, Operator: corev1.NodeSelectorOpIn, Values: []string{val3}}},
			},
		}),
		Entry("Only default node pool exists in cluster - affinity doesn't matter", &npcCase{
			name:                         "Only default node pool exists in cluster - affinity doesn't matter",
			numberOfPods:                 2,
			expectedNodePoolsForPodGroup: []string{testDefaultNodePoolName},
			nodeMatchExpressionssForPod: [][]corev1.NodeSelectorRequirement{
				{{Key: key1, Operator: corev1.NodeSelectorOpIn, Values: []string{val1}}},
			},
			podGroupLabels:      map[string]string{},
			onlyDefaultNodePool: true,
		}),
	)
})
