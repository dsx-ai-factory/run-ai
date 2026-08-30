package assigner

import (
	"context"
	"testing"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAssigner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod Group Assigner Unit Tests")
}

// The Split Flags suite proves the namespace-project / project / queue label
// keys are read from their own configured flag — splitting the historical
// runai/queue overload (one constant used for both workload queue and
// namespace project) into three independently configurable label keys.
var _ = Describe("Split flag wiring", func() {
	const (
		namespaceProjectLabelKey = "test.example.com/project-on-namespace"
		projectLabelKey          = "test.example.com/project-on-workload"
		queueLabelKey            = "test.example.com/queue"
		nodePoolLabelKey         = "test.example.com/node-pool"
		defaultNodepoolName      = "default"
		sentinelValue            = "no-such-nodepool"
	)

	BeforeEach(func() {
		DeferCleanup(config.SetForTest(config.PodGroupAssignerConfig{
			NodePoolLabelKey:           nodePoolLabelKey,
			QueueLabelKey:              queueLabelKey,
			NamespaceProjectLabelKey:   namespaceProjectLabelKey,
			ProjectLabelKey:            projectLabelKey,
			UnexistingNodepoolSentinel: sentinelValue,
			DefaultNodepoolName:        defaultNodepoolName,
		}))
	})

	Describe("getProjectOfNamespace", func() {
		It("reads from --namespace-project-label-key, not from --queue-label-key", func() {
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "team-a",
					Labels: map[string]string{
						namespaceProjectLabelKey: "team-a-project",
						queueLabelKey:            "wrong-value-from-queue-key",
					},
				},
			}

			pga := newPGAForTest(namespace)

			projectName, err := pga.getProjectOfNamespace(context.Background(), "team-a")

			Expect(err).NotTo(HaveOccurred())
			Expect(projectName).To(Equal("team-a-project"),
				"getProjectOfNamespace must read NamespaceProjectLabelKey; queue-key value must be ignored")
		})

		It("errors when the namespace lacks the configured label key", func() {
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "team-b",
					Labels: map[string]string{queueLabelKey: "ignored"},
				},
			}

			pga := newPGAForTest(namespace)

			_, err := pga.getProjectOfNamespace(context.Background(), "team-b")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("calculatePodGroupQueueNameOfNodePool", func() {
		It("reads project name from --project-label-key on the podgroup (not from --queue-label-key)", func() {
			podGroup := &kaiv2alpha2.PodGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pg-1",
					Namespace: "team-a",
					Labels: map[string]string{
						projectLabelKey: "team-a-project",
						queueLabelKey:   "wrong-value-from-queue-key",
					},
				},
			}

			matchingQueue := &kaiv2.Queue{
				ObjectMeta: metav1.ObjectMeta{
					Name: "team-a-nodepool-x-queue",
					Labels: map[string]string{
						projectLabelKey:  "team-a-project",
						nodePoolLabelKey: "nodepool-x",
					},
				},
			}

			pga := newPGAForTest(matchingQueue)

			queueName, err := pga.calculatePodGroupQueueNameOfNodePool(context.Background(), podGroup, "nodepool-x")
			Expect(err).NotTo(HaveOccurred())
			Expect(queueName).To(Equal("team-a-nodepool-x-queue"),
				"calculatePodGroupQueueNameOfNodePool must read ProjectLabelKey from the podgroup")
		})

		It("falls back to namespace via NamespaceProjectLabelKey when podgroup has no project label", func() {
			podGroup := &kaiv2alpha2.PodGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pg-2",
					Namespace: "team-a",
					Labels:    map[string]string{},
				},
			}

			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "team-a",
					Labels: map[string]string{
						namespaceProjectLabelKey: "team-a-project",
					},
				},
			}

			matchingQueue := &kaiv2.Queue{
				ObjectMeta: metav1.ObjectMeta{
					Name: "team-a-nodepool-x-queue",
					Labels: map[string]string{
						projectLabelKey:  "team-a-project",
						nodePoolLabelKey: "nodepool-x",
					},
				},
			}

			pga := newPGAForTest(namespace, matchingQueue)

			queueName, err := pga.calculatePodGroupQueueNameOfNodePool(context.Background(), podGroup, "nodepool-x")
			Expect(err).NotTo(HaveOccurred())
			Expect(queueName).To(Equal("team-a-nodepool-x-queue"))
		})
	})
})

func newPGAForTest(initObjs ...client.Object) *PodGroupAssigner {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(kaiv2.AddToScheme(scheme)).To(Succeed())
	Expect(kaiv2alpha2.AddToScheme(scheme)).To(Succeed())

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		Build()

	return NewPodGroupAssigner(fakeClient)
}
