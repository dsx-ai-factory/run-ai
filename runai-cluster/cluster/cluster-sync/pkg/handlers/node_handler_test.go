package handlers

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

// nodeWithNodePool builds a node carrying the given runai/node-pool label value.
// An empty nodePool produces a node with no node-pool label at all.
func nodeWithNodePool(name, nodePool string) corev1.Node {
	labels := map[string]string{}
	if nodePool != "" {
		labels[v1alpha1.NodePoolNameLabel] = nodePool
	}
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func nodeNames(nodes []corev1.Node) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}

var _ = Describe("filterOutExcludedNodes", func() {
	It("keeps all nodes when none are excluded", func() {
		nodes := []corev1.Node{
			nodeWithNodePool("node-a", "default"),
			nodeWithNodePool("node-b", "gpu-pool"),
		}

		managed := filterOutExcludedNodes(nodes)

		Expect(nodeNames(managed)).To(ConsistOf("node-a", "node-b"))
	})

	It("drops only the excluded nodes from a mixed list", func() {
		nodes := []corev1.Node{
			nodeWithNodePool("managed-1", "default"),
			nodeWithNodePool("excluded-1", clusterconstants.ExcludedNodepoolName),
			nodeWithNodePool("managed-2", "gpu-pool"),
			nodeWithNodePool("excluded-2", clusterconstants.ExcludedNodepoolName),
		}

		managed := filterOutExcludedNodes(nodes)

		Expect(nodeNames(managed)).To(ConsistOf("managed-1", "managed-2"))
	})

	It("returns an empty (non-nil) slice when every node is excluded", func() {
		nodes := []corev1.Node{
			nodeWithNodePool("excluded-1", clusterconstants.ExcludedNodepoolName),
			nodeWithNodePool("excluded-2", clusterconstants.ExcludedNodepoolName),
		}

		managed := filterOutExcludedNodes(nodes)

		Expect(managed).NotTo(BeNil())
		Expect(managed).To(BeEmpty())
	})

	It("keeps nodes that have no node-pool label at all", func() {
		nodes := []corev1.Node{
			nodeWithNodePool("no-label-node", ""),
			nodeWithNodePool("excluded-1", clusterconstants.ExcludedNodepoolName),
		}

		managed := filterOutExcludedNodes(nodes)

		Expect(nodeNames(managed)).To(ConsistOf("no-label-node"))
	})

	It("keeps a node whose node-pool is a real pool, not the sentinel", func() {
		nodes := []corev1.Node{
			nodeWithNodePool("regular-node", "runai-excluded-nodes-but-not-really"),
			nodeWithNodePool("excluded-1", clusterconstants.ExcludedNodepoolName),
		}

		managed := filterOutExcludedNodes(nodes)

		Expect(nodeNames(managed)).To(ConsistOf("regular-node"))
	})

	It("handles an empty input", func() {
		managed := filterOutExcludedNodes([]corev1.Node{})

		Expect(managed).NotTo(BeNil())
		Expect(managed).To(BeEmpty())
	})
})
