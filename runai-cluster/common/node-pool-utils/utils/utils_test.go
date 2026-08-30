package utils_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/common/node-pool-utils/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NodePool Utils Suite")
}

const (
	readyPool    = "np-ready"
	deletingPool = "np-deleting"
	labelKey     = "some/key"
	readyValue   = "ready-value"
	deletingVal  = "deleting-value"
)

// UseKaiNodePools is a compile-time const, so each CRD's readers are exercised
// directly rather than by flipping it. Delete the legacy entries alongside the
// legacy twins.
func newClient(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	Expect(runv1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(kaiv1alpha1.AddToScheme(scheme)).To(Succeed())
	return fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func legacyPools() []client.Object {
	return []client.Object{
		&runv1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{Name: readyPool},
			Spec:       runv1alpha1.NodePoolSpec{LabelKey: labelKey, LabelValue: readyValue},
			Status:     runv1alpha1.NodePoolStatus{Phase: runv1alpha1.NodePoolReady},
		},
		&runv1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{Name: deletingPool},
			Spec:       runv1alpha1.NodePoolSpec{LabelKey: labelKey, LabelValue: deletingVal},
			Status:     runv1alpha1.NodePoolStatus{Phase: runv1alpha1.NodePoolDeleting},
		},
	}
}

func kaiPools() []client.Object {
	return []client.Object{
		&kaiv1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{Name: readyPool},
			Spec:       kaiv1alpha1.NodePoolSpec{LabelKey: labelKey, LabelValue: readyValue},
			Status:     kaiv1alpha1.NodePoolStatus{Phase: kaiv1alpha1.NodePoolReady},
		},
		&kaiv1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{Name: deletingPool},
			Spec:       kaiv1alpha1.NodePoolSpec{LabelKey: labelKey, LabelValue: deletingVal},
			Status:     kaiv1alpha1.NodePoolStatus{Phase: kaiv1alpha1.NodePoolDeleting},
		},
	}
}

var _ = Describe("NodePool listing", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("GetAllNodePools", func() {
		It("lists only the kai CRD, ignoring the legacy one", func() {
			objects := append(legacyPools(), kaiPools()...)

			nodePools, err := utils.GetAllNodePools(ctx, newClient(objects...))
			Expect(err).ToNot(HaveOccurred())
			Expect(nodePools.Items).To(HaveLen(2))
		})

		// The caller resolves node affinity against these pools and cannot do so
		// without any, so an empty cluster is an error rather than an empty list.
		It("errors when no kai node pools exist", func() {
			_, err := utils.GetAllNodePools(ctx, newClient(legacyPools()...))
			Expect(err).To(MatchError(ContainSubstring("didn't find any node pools")))
		})
	})
})

var _ = Describe("NodePool indexing", func() {
	// Both CRDs index identically; the maps are what affinity resolution matches against.
	assertMaps := func(keyToValueToName map[string]map[string]string, deleting map[string]string) {
		Expect(keyToValueToName).To(Equal(map[string]map[string]string{
			labelKey: {
				readyValue:  readyPool,
				deletingVal: deletingPool,
			},
		}))
		// Deleting pools stay in the index but are flagged, so the caller can report
		// why a matching pool was skipped rather than silently dropping it.
		Expect(deleting).To(Equal(map[string]string{deletingPool: "Deleting"}))
	}

	It("GetNodePoolsMap indexes kai pools by label key and value", func() {
		nodePools := &kaiv1alpha1.NodePoolList{}
		for _, obj := range kaiPools() {
			nodePools.Items = append(nodePools.Items, *obj.(*kaiv1alpha1.NodePool))
		}

		assertMaps(utils.GetNodePoolsMap(nodePools.Items))
	})

	It("returns empty maps for no pools", func() {
		keyToValueToName, deleting := utils.GetNodePoolsMap(nil)
		Expect(keyToValueToName).To(BeEmpty())
		Expect(deleting).To(BeEmpty())
	})
})
