package podgroup

import (
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	RunaiProjectLabel = "project"
	RunaiQueueLabel   = "runai/queue"

	testNodePoolLabelKey    = "test/node-pool"
	testDefaultNodePoolName = "default"
	testUnexistingNodePool  = "test-unexisting-node-pool"
)

type TestData struct {
	testName              string
	podGroupLabels        map[string]string
	expectedNodePoolLabel string
}

var _ = Describe("TestNodePoolOptionsConverter", Ordered, func() {
	var (
		podGroupHandler *PodGroupMutator
		podGroup        *kaiv2alpha2.PodGroup
	)

	BeforeAll(func() {
		DeferCleanup(config.SetForTest(config.PodGroupAssignerConfig{
			NodePoolLabelKey:           testNodePoolLabelKey,
			QueueLabelKey:              RunaiQueueLabel,
			NamespaceProjectLabelKey:   RunaiQueueLabel,
			ProjectLabelKey:            RunaiProjectLabel,
			UnexistingNodepoolSentinel: testUnexistingNodePool,
			DefaultNodepoolName:        testDefaultNodePoolName,
		}))

		podGroupHandler = NewPodGroupMutator()

		podGroup = &kaiv2alpha2.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pg1",
				Namespace: "some-ns",
				Labels:    map[string]string{},
			},
			Spec: kaiv2alpha2.PodGroupSpec{
				Queue:             "some-queue",
				MarkUnschedulable: ptr.To(true),
				SchedulingBackoff: ptr.To(int32(-1)),
			},
		}
	})

	DescribeTable("TestNodePoolOptionsConverter",
		func(testData *TestData) {
			podGroup.Labels = testData.podGroupLabels
			podGroupHandler.mutatePodGroup(podGroup)

			Expect(podGroup.Spec.MarkUnschedulable).NotTo(BeNil(), "Test <%s>: expected MarkUnschedulable to be <false> after the webhook, but is was: <%t>",
				testData.testName, podGroup.Spec.MarkUnschedulable)
			Expect(*podGroup.Spec.MarkUnschedulable).To(BeFalse(), "Test <%s>: expected MarkUnschedulable to be <false> after the webhook, but is was: <%t>",
				testData.testName, *podGroup.Spec.MarkUnschedulable)

			Expect(podGroup.Spec.SchedulingBackoff).NotTo(BeNil(), "Test <%s>: expected SchedulingBackoff to be <1> after the webhook, but is was: <%t>",
				testData.testName, podGroup.Spec.SchedulingBackoff)
			Expect(*podGroup.Spec.SchedulingBackoff).To(BeFalse(), "Test <%s>: expected SchedulingBackoff to be <1> after the webhook, but is was: <%t>",
				testData.testName, *podGroup.Spec.SchedulingBackoff)

			nodePoolFromLabel, found := podGroup.Labels[testNodePoolLabelKey]
			Expect(found).To(BeTrue())
			Expect(nodePoolFromLabel).To(Equal(testData.expectedNodePoolLabel), "Test <%s>: expected nodePoolFromLabel to be <%s> after the webhook, but is was: <%s>",
				testData.testName, testData.expectedNodePoolLabel, nodePoolFromLabel)
		},
		Entry("No Node Pool Label", &TestData{
			testName:              "No Node Pool Label",
			podGroupLabels:        map[string]string{},
			expectedNodePoolLabel: testUnexistingNodePool,
		}),
		Entry("Nil Labels", &TestData{
			testName:              "Nil Labels",
			podGroupLabels:        nil,
			expectedNodePoolLabel: testUnexistingNodePool,
		}),
		Entry("Node Pool Label Exists", &TestData{
			testName: "Node Pool Label Exists",
			podGroupLabels: map[string]string{
				testNodePoolLabelKey: "some-node-pool",
			},
			expectedNodePoolLabel: "some-node-pool",
		}),
	)
})
