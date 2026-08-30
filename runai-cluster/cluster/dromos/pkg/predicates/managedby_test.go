package predicates

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestPredicates(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Predicates Suite")
}

func newRunnerManagedBy(managedBy string) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		Spec: dromosv1alpha1.WorkloadRunnerSpec{ManagedBy: managedBy},
	}
}

var _ = Describe("ManagedByPredicate", func() {
	const otherController = "other-controller"

	DescribeTable("create and update events",
		func(obj client.Object, want string, expected bool) {
			p := ManagedByPredicate(want)
			Expect(p.Create(event.CreateEvent{Object: obj})).To(Equal(expected))
			Expect(p.Update(event.UpdateEvent{ObjectNew: obj})).To(Equal(expected))
		},
		Entry("managedBy == want", newRunnerManagedBy(otherController), otherController, true),
		Entry("managedBy != want", newRunnerManagedBy(otherController), "", false),
		Entry("non-WLR object", &corev1.Pod{}, otherController, true),
	)

	It("filters updates on the NEW object's managedBy", func() {
		p := ManagedByPredicate("")
		e := event.UpdateEvent{
			ObjectOld: newRunnerManagedBy(""),
			ObjectNew: newRunnerManagedBy(otherController),
		}

		Expect(p.Update(e)).To(BeFalse())

		back := event.UpdateEvent{
			ObjectOld: newRunnerManagedBy(otherController),
			ObjectNew: newRunnerManagedBy(""),
		}
		Expect(p.Update(back)).To(BeTrue())
	})

	It("passes delete and generic events through (predicate.Funcs default)", func() {
		p := ManagedByPredicate("anything")
		Expect(p.Delete(event.DeleteEvent{Object: newRunnerManagedBy("")})).To(BeTrue())
		Expect(p.Generic(event.GenericEvent{Object: newRunnerManagedBy("")})).To(BeTrue())
	})
})
