package accessrule

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestAccessRuleEventFilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AccessRule Event Filter Unit Tests")
}

var _ = Describe("ControllerPredicateFuncs", func() {
	funcs := ControllerPredicateFuncs()

	It("admits all create, delete and generic events", func() {
		Expect(funcs.Create(event.CreateEvent{Object: &v1alpha1.AccessRule{}})).To(BeTrue())
		Expect(funcs.Delete(event.DeleteEvent{Object: &v1alpha1.AccessRule{}})).To(BeTrue())
		Expect(funcs.Generic(event.GenericEvent{Object: &v1alpha1.AccessRule{}})).To(BeTrue())
	})

	Describe("update events", func() {
		It("admits an AccessRule update only when the resourceVersion changed", func() {
			oldAR := &v1alpha1.AccessRule{ObjectMeta: metav1.ObjectMeta{Name: "ar", ResourceVersion: "1"}}

			sameRV := &v1alpha1.AccessRule{ObjectMeta: metav1.ObjectMeta{Name: "ar", ResourceVersion: "1"}}
			Expect(funcs.Update(event.UpdateEvent{ObjectOld: oldAR, ObjectNew: sameRV})).To(BeFalse())

			newRV := &v1alpha1.AccessRule{ObjectMeta: metav1.ObjectMeta{Name: "ar", ResourceVersion: "2"}}
			Expect(funcs.Update(event.UpdateEvent{ObjectOld: oldAR, ObjectNew: newRV})).To(BeTrue())
		})

		It("always admits RoleBinding updates", func() {
			old := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "rb", Namespace: "ns"}}
			updated := old.DeepCopy()
			Expect(funcs.Update(event.UpdateEvent{ObjectOld: old, ObjectNew: updated})).To(BeTrue())
		})

		It("filters out updates of unrelated types", func() {
			old := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "cr"}}
			updated := old.DeepCopy()
			Expect(funcs.Update(event.UpdateEvent{ObjectOld: old, ObjectNew: updated})).To(BeFalse())
		})
	})
})

var _ = Describe("Map RoleBinding/ClusterRoleBinding events to AccessRule requests", func() {
	c := &AccessRuleController{}

	It("extracts access rule names from RoleBinding labels (both prefixes)", func() {
		rb := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rb",
				Namespace: "ns",
				Labels: map[string]string{
					common.AccessRuleAddedSubjectToProjectsRBPrefix + "ar-project": "true",
					common.AccessRuleAddedSubjectPrefix + "ar-runai":               "true",
					"unrelated-label": "true",
				},
			},
		}

		reqs := c.MapRoleBindingToAccessRuleEvent(context.Background(), rb)
		names := make([]string, 0, len(reqs))
		for _, r := range reqs {
			names = append(names, r.NamespacedName.Name)
		}
		Expect(names).To(ConsistOf("ar-project", "ar-runai"))
	})

	It("returns no requests for a RoleBinding without labels", func() {
		rb := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "rb", Namespace: "ns"}}
		Expect(c.MapRoleBindingToAccessRuleEvent(context.Background(), rb)).To(BeEmpty())
	})

	It("extracts access rule names from ClusterRoleBinding labels", func() {
		crb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "crb",
				Labels: map[string]string{
					common.AccessRuleAddedSubjectPrefix + "ar-cluster": "true",
					"unrelated-label": "true",
				},
			},
		}

		reqs := c.MapClusterRoleBindingToAccessRuleEvent(context.Background(), crb)
		names := make([]string, 0, len(reqs))
		for _, r := range reqs {
			names = append(names, r.NamespacedName.Name)
		}
		Expect(names).To(ConsistOf("ar-cluster"))
	})

	It("returns no requests for a ClusterRoleBinding without labels", func() {
		crb := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "crb"}}
		Expect(c.MapClusterRoleBindingToAccessRuleEvent(context.Background(), crb)).To(BeEmpty())
	})
})
