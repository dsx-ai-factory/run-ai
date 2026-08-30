package nodepool_controller

import (
	"context"
	"fmt"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func projectRefTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func projectWithQueues(name string, nodePoolNames ...string) *v1alpha1.Project {
	queues := make([]v1alpha1.QueueConfig, 0, len(nodePoolNames))
	for _, npName := range nodePoolNames {
		queues = append(queues, v1alpha1.QueueConfig{Name: name + "-" + npName, Nodepool: npName})
	}
	return &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       v1alpha1.ProjectSpec{Queues: queues},
	}
}

var _ = Describe("Project reference handling", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = projectRefTestScheme()
	})

	Describe("getProjectsReferencingNodePool", func() {
		It("returns projects referencing the nodepool via queues, sorted", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				projectWithQueues("proj-c", "node-pool-a"),
				projectWithQueues("proj-a", "node-pool-a", "node-pool-b"),
				projectWithQueues("proj-b", "node-pool-b"),
			).Build()
			npc := &NodePoolController{Client: fakeClient, Scheme: scheme}

			referencing, err := npc.getProjectsReferencingNodePool(ctx, "node-pool-a")
			Expect(err).ToNot(HaveOccurred())
			Expect(referencing).To(Equal([]string{"proj-a", "proj-c"}))
		})

		It("ignores a reference that exists only in the default nodepool list", func() {
			project := projectWithQueues("proj-a")
			project.Spec.DefaultNodePools = []string{"node-pool-a"}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(project).Build()
			npc := &NodePoolController{Client: fakeClient, Scheme: scheme}

			referencing, err := npc.getProjectsReferencingNodePool(ctx, "node-pool-a")
			Expect(err).ToNot(HaveOccurred())
			Expect(referencing).To(BeEmpty())
		})

		It("returns empty when no project references the nodepool", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				projectWithQueues("proj-a", "node-pool-b"),
			).Build()
			npc := &NodePoolController{Client: fakeClient, Scheme: scheme}

			referencing, err := npc.getProjectsReferencingNodePool(ctx, "node-pool-a")
			Expect(err).ToNot(HaveOccurred())
			Expect(referencing).To(BeEmpty())
		})
	})

	Describe("getProjectsReferenceMessage", func() {
		It("returns empty for no projects", func() {
			Expect(getProjectsReferenceMessage(nil)).To(BeEmpty())
		})

		It("lists all projects when under the cap", func() {
			Expect(getProjectsReferenceMessage([]string{"a", "b"})).To(
				Equal(fmt.Sprintf(common.ProjectsReferencingNodePoolMessage, "a, b")))
		})

		It("caps the list and appends a remainder count", func() {
			projects := make([]string, maxProjectNamesInMessage+3)
			for i := range projects {
				projects[i] = fmt.Sprintf("p%02d", i)
			}
			msg := getProjectsReferenceMessage(projects)
			Expect(msg).To(ContainSubstring("(and 3 more)"))
			Expect(msg).To(ContainSubstring("p00"))
			Expect(msg).ToNot(ContainSubstring("p10"))
		})
	})

	Describe("MapProjectToNodePoolEvents", func() {
		var (
			deletingNodePool *v1alpha1.NodePool
			readyNodePool    *v1alpha1.NodePool
		)

		BeforeEach(func() {
			deletingNodePool = &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: "deleting-np"},
				Status:     v1alpha1.NodePoolStatus{Phase: v1alpha1.NodePoolDeleting},
			}
			readyNodePool = &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{Name: "ready-np"},
				Status:     v1alpha1.NodePoolStatus{Phase: v1alpha1.NodePoolReady},
			}
		})

		It("enqueues exactly the nodepools that are being deleted", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).
				WithObjects(deletingNodePool, readyNodePool).
				WithIndex(&v1alpha1.NodePool{}, common.IsDeletingPhaseField, NodePoolIsDeletingPhaseIndexer).
				Build()
			npc := &NodePoolController{Client: fakeClient, Scheme: scheme}

			requests := npc.MapProjectToNodePoolEvents(ctx, projectWithQueues("proj-a", "deleting-np"))
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].Name).To(Equal("deleting-np"))
		})

		It("enqueues nothing when no nodepool is being deleted", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).
				WithObjects(readyNodePool).
				WithIndex(&v1alpha1.NodePool{}, common.IsDeletingPhaseField, NodePoolIsDeletingPhaseIndexer).
				Build()
			npc := &NodePoolController{Client: fakeClient, Scheme: scheme}

			requests := npc.MapProjectToNodePoolEvents(ctx, projectWithQueues("proj-a", "ready-np"))
			Expect(requests).To(BeEmpty())
		})
	})

	Describe("filterUpdateEvents for Project", func() {
		It("passes when a queue nodepool reference is added", func() {
			old := projectWithQueues("proj-a", "node-pool-a")
			updated := projectWithQueues("proj-a", "node-pool-a", "node-pool-b")
			Expect(filterUpdateEvents(old, updated)).To(BeTrue())
		})

		It("passes when a queue nodepool reference is removed", func() {
			old := projectWithQueues("proj-a", "node-pool-a", "node-pool-b")
			updated := projectWithQueues("proj-a", "node-pool-a")
			Expect(filterUpdateEvents(old, updated)).To(BeTrue())
		})

		It("drops updates that do not change queue nodepool references", func() {
			old := projectWithQueues("proj-a", "node-pool-a")
			updated := projectWithQueues("proj-a", "node-pool-a")
			updated.Spec.DefaultNodePools = []string{"node-pool-a"}
			updated.Labels = map[string]string{"changed": "true"}
			Expect(filterUpdateEvents(old, updated)).To(BeFalse())
		})

		It("drops updates that change only the queue's quota, not the nodepool reference", func() {
			old := projectWithQueues("proj-a", "node-pool-a")
			updated := projectWithQueues("proj-a", "node-pool-a")
			updated.Spec.Queues[0].Resources = &v1alpha1.QueueResourcesConfig{
				GPU: v1alpha1.SystemResource{Deserved: 5, Limit: -1},
			}
			Expect(filterUpdateEvents(old, updated)).To(BeFalse())
		})
	})
})
