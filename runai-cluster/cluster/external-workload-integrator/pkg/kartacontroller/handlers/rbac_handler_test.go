package handlers

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/test/fakecachedclient"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestKartaControllerHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KartaController Handlers Suite")
}

var _ = Describe("KartaRBACHandler", Ordered, func() {
	var (
		scheme           *runtime.Scheme
		ctx              context.Context
		fakeCachedClient cachedclient.CachedClient
	)

	BeforeAll(func() {
		scheme = runtime.NewScheme()
		Expect(rbacv1.AddToScheme(scheme)).To(Succeed())
		ctx = context.Background()

		fakeCachedClient = fakecachedclient.NewFakeCachedClient(nil, nil, false)
	})

	Context("HandleRBAC", func() {
		It("returns nil when RBAC already exists", func() {
			// given
			gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			existing := &rbacv1.ClusterRole{
				ObjectMeta: objMeta("role-existing"),
				Rules: []rbacv1.PolicyRule{{
					APIGroups: []string{"apps"},
					Resources: []string{"deployments"},
					Verbs:     []string{"get", "list", "watch"},
				}},
			}
			existingCrud := &rbacv1.ClusterRole{
				ObjectMeta: objMeta("role-existing-crud"),
				Rules: []rbacv1.PolicyRule{{
					APIGroups: []string{"apps"},
					Resources: []string{"deployments"},
					Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				}},
			}
			k8s := newFakeClientWithObjects(scheme, existing, existingCrud)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, false, existing.Name, existingCrud.Name)

			// when
			err := handler.Handle(ctx, gvk, []schema.GroupVersionKind{})

			// then
			Expect(err).To(BeNil())
			cr := getClusterRoleOrFail(ctx, k8s, existing.Name)
			Expect(cr.Rules).To(HaveLen(1))
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, cr, false)).To(BeTrue())
			crCrud := getClusterRoleOrFail(ctx, k8s, existingCrud.Name)
			Expect(cr.Rules).To(HaveLen(1))
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, crCrud, true)).To(BeTrue())

			// same when updater enabled
			handler = NewKartaRBACHandler(k8s, fakeCachedClient, true, existing.Name, existingCrud.Name)
			err = handler.Handle(ctx, gvk, []schema.GroupVersionKind{})
			Expect(err).To(BeNil())
		})

		It("fails when RBAC missing and updater disabled", func() {
			// given
			gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			cr := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-disabled")}
			crCrud := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-disabled-crud")}
			k8s := newFakeClientWithObjects(scheme, cr)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, false, cr.Name, crCrud.Name)

			// when
			err := handler.Handle(ctx, gvk, []schema.GroupVersionKind{})

			// then
			Expect(err).To(HaveOccurred())
		})

		It("fails when ClusterRole missing and updater disabled", func() {
			// given
			gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			k8s := newFakeClientWithObjects(scheme)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, false, "role-missing", "role-missing-crud")

			// when
			err := handler.Handle(ctx, gvk, []schema.GroupVersionKind{})

			// then
			Expect(err).To(HaveOccurred())

			// validate ClusterRoles does not exist though
			cr := &rbacv1.ClusterRole{}
			Expect(k8s.Get(ctx, client.ObjectKey{Name: "role-missing"}, cr)).ToNot(Succeed())
			Expect(k8s.Get(ctx, client.ObjectKey{Name: "role-missing-crud"}, cr)).ToNot(Succeed())
		})

		It("succeeds when ClusterRole missing and updater enabled", func() {
			// given
			gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			k8s := newFakeClientWithObjects(scheme)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, "role-created", "role-created-crud")

			// when
			err := handler.Handle(ctx, gvk, []schema.GroupVersionKind{})

			// then
			Expect(err).To(BeNil())
			created := getClusterRoleOrFail(ctx, k8s, "role-created")
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, created, false)).To(BeTrue())
			createdCrud := getClusterRoleOrFail(ctx, k8s, "role-created-crud")
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, createdCrud, true)).To(BeTrue())

			// ensure the exact rule was created
			rule, ok := findRule(created, "apps", "deployments")
			Expect(ok).To(BeTrue())
			Expect(HasAllVerbs(rule, false)).To(BeTrue())
			rule, ok = findRule(createdCrud, "apps", "deployments")
			Expect(ok).To(BeTrue())
			Expect(HasAllVerbs(rule, true)).To(BeTrue())
		})

		It("adds rule when RBAC missing and updater enabled", func() {
			// given
			gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			cr := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-enabled")}
			crCrud := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-enabled-crud")}
			k8s := newFakeClientWithObjects(scheme, cr, crCrud)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, cr.Name, crCrud.Name)

			// when
			err := handler.Handle(ctx, gvk, []schema.GroupVersionKind{})

			// then
			Expect(err).To(BeNil())
			updated := getClusterRoleOrFail(ctx, k8s, cr.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, updated, false)).To(BeTrue())
			updatedCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, updatedCrud, true)).To(BeTrue())

			// ensure the exact rule was created
			rule, ok := findRule(updated, "apps", "deployments")
			Expect(ok).To(BeTrue())
			Expect(HasAllVerbs(rule, false)).To(BeTrue())
			rule, ok = findRule(updatedCrud, "apps", "deployments")
			Expect(ok).To(BeTrue())
			Expect(HasAllVerbs(rule, true)).To(BeTrue())
		})

		It("adds a new rule when existing rule misses verbs", func() {
			// given
			gvk := schema.GroupVersionKind{Group: "kubevirt.io", Version: "v1", Kind: "VirtualMachineInstance"}
			cr := &rbacv1.ClusterRole{
				ObjectMeta: objMeta("role-partial"),
				Rules: []rbacv1.PolicyRule{{
					APIGroups: []string{"kubevirt.io"},
					Resources: []string{"virtualmachineinstances"},
					Verbs:     []string{"get", "list"}, // missing watch
				}},
			}
			crCrud := &rbacv1.ClusterRole{
				ObjectMeta: objMeta("role-partial-crud"),
				Rules: []rbacv1.PolicyRule{{
					APIGroups: []string{"kubevirt.io"},
					Resources: []string{"virtualmachineinstances"},
					Verbs:     []string{"create", "update"}, // missing watch
				}},
			}
			k8s := newFakeClientWithObjects(scheme, cr, crCrud)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, cr.Name, crCrud.Name)

			// sanity
			current := getClusterRoleOrFail(ctx, k8s, cr.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, current, false)).To(BeFalse())
			existingRules := len(current.Rules)
			currentCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, current, true)).To(BeFalse())
			existingCrudRules := len(currentCrud.Rules)

			// when
			err := handler.Handle(ctx, gvk, []schema.GroupVersionKind{})

			// then
			Expect(err).To(BeNil())
			updated := getClusterRoleOrFail(ctx, k8s, cr.Name)
			Expect(len(updated.Rules)).To(Equal(existingRules + 1))
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, updated, false)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updated, "kubevirt.io", "virtualmachineinstances", false)).To(BeTrue())
			updatedCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(len(updatedCrud.Rules)).To(Equal(existingCrudRules + 1))
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, updatedCrud, true)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updatedCrud, "kubevirt.io", "virtualmachineinstances", true)).To(BeTrue())
		})

		It("adds RBAC for multiple gvks (read and CRUD)", func() {
			// given
			gvkMain := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			gvkAdditional1 := schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}
			gvkAdditional2 := schema.GroupVersionKind{Group: "kubevirt.io", Version: "v1", Kind: "VirtualMachineInstance"}
			additional := []schema.GroupVersionKind{gvkAdditional1, gvkAdditional2}

			cr := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-multi-read")}
			crCrud := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-multi-crud")}
			k8s := newFakeClientWithObjects(scheme, cr, crCrud)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, cr.Name, crCrud.Name)

			// when
			err := handler.Handle(ctx, gvkMain, additional)

			// then
			Expect(err).To(BeNil())
			updated := getClusterRoleOrFail(ctx, k8s, cr.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, updated, false)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updated, "apps", "deployments", false)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updated, "batch", "jobs", false)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updated, "kubevirt.io", "virtualmachineinstances", false)).To(BeTrue())

			updatedCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, updatedCrud, true)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updatedCrud, "apps", "deployments", true)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updatedCrud, "batch", "jobs", true)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updatedCrud, "kubevirt.io", "virtualmachineinstances", true)).To(BeTrue())
		})

		It("removes RBAC for multiple gvks (read and CRUD)", func() {
			// given
			gvkMain := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
			gvkAdditional1 := schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"}
			gvkAdditional2 := schema.GroupVersionKind{Group: "kubevirt.io", Version: "v1", Kind: "VirtualMachineInstance"}
			additional := []schema.GroupVersionKind{gvkAdditional1, gvkAdditional2}

			// pre-existing roles with combined rule
			readRule := rbacv1.PolicyRule{
				APIGroups: []string{"apps", "batch", "kubevirt.io"},
				Resources: []string{"deployments", "jobs", "virtualmachineinstances"},
				Verbs:     []string{"get", "list", "watch"},
			}
			crudRule := rbacv1.PolicyRule{
				APIGroups: []string{"apps", "batch", "kubevirt.io"},
				Resources: []string{"deployments", "jobs", "virtualmachineinstances"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			}
			cr := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-multi-read-del"), Rules: []rbacv1.PolicyRule{readRule}}
			crCrud := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-multi-crud-del"), Rules: []rbacv1.PolicyRule{crudRule}}
			k8s := newFakeClientWithObjects(scheme, cr, crCrud)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, cr.Name, crCrud.Name)

			// sanity precondition
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, cr, false)).To(BeTrue())
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, crCrud, true)).To(BeTrue())

			// when
			err := handler.HandleDeletion(ctx, gvkMain, additional)

			// then
			Expect(err).To(BeNil())
			updated := getClusterRoleOrFail(ctx, k8s, cr.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, updated, false)).To(BeFalse())
			updatedCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, updatedCrud, true)).To(BeFalse())
		})
	})

	Context("helpers", func() {
		It("resourceNameFromGVK lowercases and pluralizes kind", func() {
			gvk := schema.GroupVersionKind{Group: "x.run.ai", Version: "v1", Kind: "Foo"}
			Expect(resourceNameFromGVK(gvk)).To(Equal("foos"))
		})
	})
})

// helpers
func newFakeClientWithObjects(s *runtime.Scheme, objs ...client.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(s).WithObjects(objs...).Build()
}

func objMeta(name string) metav1.ObjectMeta { return metav1.ObjectMeta{Name: name} }

func getClusterRoleOrFail(ctx context.Context, c client.Client, name string) *rbacv1.ClusterRole {
	cr := &rbacv1.ClusterRole{}
	Expect(c.Get(ctx, client.ObjectKey{Name: name}, cr)).To(Succeed())
	return cr
}

func findRule(cr *rbacv1.ClusterRole, group, resource string) (rbacv1.PolicyRule, bool) {
	for _, r := range cr.Rules {
		if contains(r.APIGroups, group) && contains(r.Resources, resource) {
			return r, true
		}
	}
	return rbacv1.PolicyRule{}, false
}

func hasAnyRuleWithAllVerbs(cr *rbacv1.ClusterRole, group, resource string, hasCrud bool) bool {
	for _, r := range cr.Rules {
		if contains(r.APIGroups, group) && contains(r.Resources, resource) && HasAllVerbs(r, hasCrud) {
			return true
		}
	}
	return false
}

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
