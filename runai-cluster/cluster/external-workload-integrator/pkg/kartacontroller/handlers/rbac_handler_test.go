/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/test/fakecachedclient"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
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
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, cr, false, resourceNameFromGVK)).To(BeTrue())
			crCrud := getClusterRoleOrFail(ctx, k8s, existingCrud.Name)
			Expect(cr.Rules).To(HaveLen(1))
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, crCrud, true, resourceNameFromGVK)).To(BeTrue())

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
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, created, false, resourceNameFromGVK)).To(BeTrue())
			createdCrud := getClusterRoleOrFail(ctx, k8s, "role-created-crud")
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, createdCrud, true, resourceNameFromGVK)).To(BeTrue())

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
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, updated, false, resourceNameFromGVK)).To(BeTrue())
			updatedCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, updatedCrud, true, resourceNameFromGVK)).To(BeTrue())

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
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, current, false, resourceNameFromGVK)).To(BeFalse())
			existingRules := len(current.Rules)
			currentCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, current, true, resourceNameFromGVK)).To(BeFalse())
			existingCrudRules := len(currentCrud.Rules)

			// when
			err := handler.Handle(ctx, gvk, []schema.GroupVersionKind{})

			// then
			Expect(err).To(BeNil())
			updated := getClusterRoleOrFail(ctx, k8s, cr.Name)
			Expect(len(updated.Rules)).To(Equal(existingRules + 1))
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, updated, false, resourceNameFromGVK)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updated, "kubevirt.io", "virtualmachineinstances", false)).To(BeTrue())
			updatedCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(len(updatedCrud.Rules)).To(Equal(existingCrudRules + 1))
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvk}, updatedCrud, true, resourceNameFromGVK)).To(BeTrue())
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
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, updated, false, resourceNameFromGVK)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updated, "apps", "deployments", false)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updated, "batch", "jobs", false)).To(BeTrue())
			Expect(hasAnyRuleWithAllVerbs(updated, "kubevirt.io", "virtualmachineinstances", false)).To(BeTrue())

			updatedCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, updatedCrud, true, resourceNameFromGVK)).To(BeTrue())
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
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, cr, false, resourceNameFromGVK)).To(BeTrue())
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, crCrud, true, resourceNameFromGVK)).To(BeTrue())

			// when
			err := handler.HandleDeletion(ctx, gvkMain, additional)

			// then
			Expect(err).To(BeNil())
			updated := getClusterRoleOrFail(ctx, k8s, cr.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, updated, false, resourceNameFromGVK)).To(BeFalse())
			updatedCrud := getClusterRoleOrFail(ctx, k8s, crCrud.Name)
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{gvkMain, gvkAdditional1, gvkAdditional2}, updatedCrud, true, resourceNameFromGVK)).To(BeFalse())
		})
	})

	Context("resource names", func() {
		sandbox := schema.GroupVersionKind{Group: "agents.x-k8s.io", Version: "v1beta1", Kind: "Sandbox"}

		It("uses the resource name served by the API server instead of guessing it from the kind", func() {
			// given
			irregular := schema.GroupVersionKind{Group: "x.run.ai", Version: "v1", Kind: "Mouse"}
			mapper := meta.NewDefaultRESTMapper(nil)
			mapper.AddSpecific(sandbox, sandbox.GroupVersion().WithResource("sandboxes"),
				sandbox.GroupVersion().WithResource("sandbox"), meta.RESTScopeNamespace)
			mapper.AddSpecific(irregular, irregular.GroupVersion().WithResource("mice"),
				irregular.GroupVersion().WithResource("mouse"), meta.RESTScopeNamespace)
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).Build()
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, "role-mapped", "role-mapped-crud")

			// when
			err := handler.Handle(ctx, sandbox, []schema.GroupVersionKind{irregular})

			// then
			Expect(err).To(BeNil())
			created := getClusterRoleOrFail(ctx, k8s, "role-mapped")
			Expect(created.Rules).ToNot(BeEmpty())
			Expect(created.Rules[0].Resources).To(ConsistOf("sandboxes", "mice"))
			names, err := handler.resolveResourceNames([]schema.GroupVersionKind{sandbox, irregular})
			Expect(err).To(BeNil())
			Expect(hasGvksRbacInClusterRole([]schema.GroupVersionKind{sandbox, irregular}, created, false, names.of)).To(BeTrue())
		})

		It("guesses the resource name from the kind when the API server does not know it", func() {
			// given
			k8s := newFakeClientWithObjects(scheme)
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, "role-guessed", "role-guessed-crud")

			// when
			err := handler.Handle(ctx, sandbox, []schema.GroupVersionKind{})

			// then
			Expect(err).To(BeNil())
			created := getClusterRoleOrFail(ctx, k8s, "role-guessed")
			Expect(created.Rules).ToNot(BeEmpty())
			Expect(created.Rules[0].Resources).To(ConsistOf("sandboxes"))
		})

		It("guesses the resource name when discovery reports the group as missing, as aggregated discovery does", func() {
			// given: a CRD that is not installed yet comes back from the lazy REST mapper wrapped in a discovery error
			discoveryErr := apiutil.ErrResourceDiscoveryFailed{
				sandbox.GroupVersion(): &meta.NoResourceMatchError{PartialResource: sandbox.GroupVersion().WithResource("")},
			}
			mapper := &countingRESTMapper{RESTMapper: meta.NewDefaultRESTMapper(nil), err: &discoveryErr}
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).Build()
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, "role-missing-group", "role-missing-group-crud")

			// when
			err := handler.Handle(ctx, sandbox, nil)

			// then
			Expect(err).To(BeNil())
			Expect(getClusterRoleOrFail(ctx, k8s, "role-missing-group").Rules[0].Resources).To(ConsistOf("sandboxes"))
		})

		It("returns the error instead of guessing when the API server fails to resolve a name", func() {
			// given: the rule with the real, irregular name exists, and the lookup fails like a discovery timeout
			mouse := schema.GroupVersionKind{Group: "x.run.ai", Version: "v1", Kind: "Mouse"}
			base := meta.NewDefaultRESTMapper(nil)
			base.AddSpecific(mouse, mouse.GroupVersion().WithResource("mice"),
				mouse.GroupVersion().WithResource("mouse"), meta.RESTScopeNamespace)
			mapper := &countingRESTMapper{RESTMapper: base, err: errors.New("discovery timed out")}
			existing := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-transient"), Rules: []rbacv1.PolicyRule{{
				APIGroups: []string{"x.run.ai"}, Resources: []string{"mice"}, Verbs: []string{"get", "list", "watch"}}}}
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(existing).Build()
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, existing.Name, "")

			// when
			err := handler.Handle(ctx, mouse, nil)

			// then: no rule is written with the guessed name "mouses", and the reconcile is retried
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to resolve the resource name of <x.run.ai/v1, Kind=Mouse>"))
			Expect(getClusterRoleOrFail(ctx, k8s, existing.Name).Rules).To(HaveLen(1))

			// when the lookup works again
			mapper.err = nil
			err = handler.Handle(ctx, mouse, nil)

			// then the existing rule is found and nothing is added
			Expect(err).To(BeNil())
			Expect(getClusterRoleOrFail(ctx, k8s, existing.Name).Rules).To(HaveLen(1))
		})

		It("resolves each resource name once, however many rules the cluster roles hold", func() {
			// given: cluster roles with many unrelated rules, so that matching has to visit every one of them
			irregular := schema.GroupVersionKind{Group: "x.run.ai", Version: "v1", Kind: "Mouse"}
			base := meta.NewDefaultRESTMapper(nil)
			base.AddSpecific(sandbox, sandbox.GroupVersion().WithResource("sandboxes"),
				sandbox.GroupVersion().WithResource("sandbox"), meta.RESTScopeNamespace)
			base.AddSpecific(irregular, irregular.GroupVersion().WithResource("mice"),
				irregular.GroupVersion().WithResource("mouse"), meta.RESTScopeNamespace)
			mapper := &countingRESTMapper{RESTMapper: base}
			unrelatedRules := func(verbs []string) []rbacv1.PolicyRule {
				rules := []rbacv1.PolicyRule{}
				for i := 0; i < 10; i++ {
					rules = append(rules, rbacv1.PolicyRule{
						APIGroups: []string{"other.run.ai"}, Resources: []string{fmt.Sprintf("things%d", i)}, Verbs: verbs})
				}
				return rules
			}
			read := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-once-read"),
				Rules: unrelatedRules([]string{"get", "list", "watch"})}
			crud := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-once-crud"),
				Rules: unrelatedRules([]string{"get", "list", "watch", "create", "update", "patch", "delete"})}
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(read, crud).Build()
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, read.Name, crud.Name)

			// when
			err := handler.Handle(ctx, sandbox, []schema.GroupVersionKind{irregular})

			// then: one lookup for each of the two kinds, not one for every rule and cluster role
			Expect(err).To(BeNil())
			Expect(mapper.calls).To(Equal(2))
		})

		It("still removes the rules on deletion when the API server fails to resolve a name", func() {
			// given: a rule written with the guessed name, and a lookup that fails like a discovery timeout
			mapper := &countingRESTMapper{RESTMapper: meta.NewDefaultRESTMapper(nil), err: errors.New("discovery timed out")}
			read := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-del-failing"), Rules: []rbacv1.PolicyRule{{
				APIGroups: []string{"agents.x-k8s.io"}, Resources: []string{"sandboxes"}, Verbs: []string{"get", "list", "watch"}}}}
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(read).Build()
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, read.Name, "")

			// when
			err := handler.HandleDeletion(ctx, sandbox, nil)

			// then: deletion does not wait for the API server
			Expect(err).To(BeNil())
			Expect(getClusterRoleOrFail(ctx, k8s, read.Name).Rules).To(BeEmpty())
		})

		It("removes the rule for the resource name served by the API server", func() {
			// given
			mapper := meta.NewDefaultRESTMapper(nil)
			mapper.AddSpecific(sandbox, sandbox.GroupVersion().WithResource("sandboxes"),
				sandbox.GroupVersion().WithResource("sandbox"), meta.RESTScopeNamespace)
			cr := &rbacv1.ClusterRole{
				ObjectMeta: objMeta("role-remove"),
				Rules: []rbacv1.PolicyRule{{
					APIGroups: []string{"agents.x-k8s.io"},
					Resources: []string{"sandboxes"},
					Verbs:     []string{"get", "list", "watch"},
				}},
			}
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(cr).Build()
			handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, "role-remove", "")

			// when
			err := handler.removeRuleFromClusterRole(ctx, cr, handler.removalTargets([]schema.GroupVersionKind{sandbox}), false, "read")

			// then
			Expect(err).To(BeNil())
			Expect(getClusterRoleOrFail(ctx, k8s, "role-remove").Rules).To(BeEmpty())
		})

		Context("deletion", func() {
			readVerbs := []string{"get", "list", "watch"}
			crudVerbs := []string{"get", "list", "watch", "create", "update", "patch", "delete"}
			rolesWith := func(group, resource string) (*rbacv1.ClusterRole, *rbacv1.ClusterRole) {
				read := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-del-read"), Rules: []rbacv1.PolicyRule{
					{APIGroups: []string{group}, Resources: []string{resource}, Verbs: readVerbs}}}
				crud := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-del-crud"), Rules: []rbacv1.PolicyRule{
					{APIGroups: []string{group}, Resources: []string{resource}, Verbs: crudVerbs}}}
				return read, crud
			}

			It("removes the rule of an irregular plural when the API server serves the type", func() {
				// given
				mouse := schema.GroupVersionKind{Group: "x.run.ai", Version: "v1", Kind: "Mouse"}
				mapper := meta.NewDefaultRESTMapper(nil)
				mapper.AddSpecific(mouse, mouse.GroupVersion().WithResource("mice"),
					mouse.GroupVersion().WithResource("mouse"), meta.RESTScopeNamespace)
				read, crud := rolesWith("x.run.ai", "mice")
				k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(read, crud).Build()
				handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, read.Name, crud.Name)

				// when
				err := handler.HandleDeletion(ctx, mouse, nil)

				// then
				Expect(err).To(BeNil())
				Expect(getClusterRoleOrFail(ctx, k8s, read.Name).Rules).To(BeEmpty())
				Expect(getClusterRoleOrFail(ctx, k8s, crud.Name).Rules).To(BeEmpty())
			})

			It("removes rules that earlier versions wrote with the Kind + s name", func() {
				// given
				sandbox := schema.GroupVersionKind{Group: "agents.x-k8s.io", Version: "v1beta1", Kind: "Sandbox"}
				mapper := meta.NewDefaultRESTMapper(nil)
				mapper.AddSpecific(sandbox, sandbox.GroupVersion().WithResource("sandboxes"),
					sandbox.GroupVersion().WithResource("sandbox"), meta.RESTScopeNamespace)
				read, crud := rolesWith("agents.x-k8s.io", "sandboxs")
				k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(read, crud).Build()
				handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, read.Name, crud.Name)

				// when
				err := handler.HandleDeletion(ctx, sandbox, nil)

				// then
				Expect(err).To(BeNil())
				Expect(getClusterRoleOrFail(ctx, k8s, read.Name).Rules).To(BeEmpty())
				Expect(getClusterRoleOrFail(ctx, k8s, crud.Name).Rules).To(BeEmpty())
			})

			It("keeps the other resources of a rule that is shared with the removed type", func() {
				// given
				sandbox := schema.GroupVersionKind{Group: "agents.x-k8s.io", Version: "v1beta1", Kind: "Sandbox"}
				mapper := meta.NewDefaultRESTMapper(nil)
				mapper.AddSpecific(sandbox, sandbox.GroupVersion().WithResource("sandboxes"),
					sandbox.GroupVersion().WithResource("sandbox"), meta.RESTScopeNamespace)
				read := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-shared-read"), Rules: []rbacv1.PolicyRule{{
					APIGroups: []string{"agents.x-k8s.io"}, Resources: []string{"sandboxes", "widgets"}, Verbs: readVerbs}}}
				k8s := fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).WithObjects(read).Build()
				handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, read.Name, "")

				// when
				err := handler.HandleDeletion(ctx, sandbox, nil)

				// then
				Expect(err).To(BeNil())
				rules := getClusterRoleOrFail(ctx, k8s, read.Name).Rules
				Expect(rules).To(HaveLen(1))
				Expect(rules[0].Resources).To(ConsistOf("widgets"))
			})

			It("keeps a resource of the same name that belongs to another API group", func() {
				// given: one rule that grants widgets and gadgets in two API groups
				widget := schema.GroupVersionKind{Group: "a.run.ai", Version: "v1", Kind: "Widget"}
				read := &rbacv1.ClusterRole{ObjectMeta: objMeta("role-groups-read"), Rules: []rbacv1.PolicyRule{{
					APIGroups: []string{"a.run.ai", "b.run.ai"}, Resources: []string{"widgets", "gadgets"}, Verbs: readVerbs}}}
				k8s := newFakeClientWithObjects(scheme, read)
				handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, read.Name, "")

				// when
				err := handler.HandleDeletion(ctx, widget, nil)

				// then: a.run.ai keeps gadgets only, and b.run.ai still has both
				Expect(err).To(BeNil())
				rules := getClusterRoleOrFail(ctx, k8s, read.Name).Rules
				Expect(rules).To(HaveLen(2))
				Expect(rules[0].APIGroups).To(ConsistOf("a.run.ai"))
				Expect(rules[0].Resources).To(ConsistOf("gadgets"))
				Expect(rules[0].Verbs).To(ConsistOf(readVerbs))
				Expect(rules[1].APIGroups).To(ConsistOf("b.run.ai"))
				Expect(rules[1].Resources).To(ConsistOf("widgets", "gadgets"))
			})

			It("does not fail, and leaves the rule, when an irregular plural cannot be resolved", func() {
				// given: the API server no longer serves the type, so only the guessed names can be matched
				mouse := schema.GroupVersionKind{Group: "x.run.ai", Version: "v1", Kind: "Mouse"}
				read, crud := rolesWith("x.run.ai", "mice")
				k8s := newFakeClientWithObjects(scheme, read, crud)
				handler := NewKartaRBACHandler(k8s, fakeCachedClient, true, read.Name, crud.Name)

				// when
				err := handler.HandleDeletion(ctx, mouse, nil)

				// then: deletion is not blocked, and the rule it could not identify is kept (a warning is logged)
				Expect(err).To(BeNil())
				Expect(getClusterRoleOrFail(ctx, k8s, read.Name).Rules).To(HaveLen(1))
			})
		})

		DescribeTable("guesses the plural of a kind",
			func(kind, expected string) {
				gvk := schema.GroupVersionKind{Group: "x.run.ai", Version: "v1", Kind: kind}
				Expect(resourceNameFromGVK(gvk)).To(Equal(expected))
			},
			Entry("adds s", "Foo", "foos"),
			Entry("adds s to a kind ending in a vowel followed by y", "Gateway", "gateways"),
			Entry("adds es to a kind ending in x", "Sandbox", "sandboxes"),
			Entry("adds es to a kind ending in s", "Ingress", "ingresses"),
			Entry("adds es to a kind ending in ch", "Batch", "batches"),
			Entry("adds es to a kind ending in sh", "Bush", "bushes"),
			Entry("replaces a consonant followed by y with ies", "Policy", "policies"),
			Entry("lowercases the kind", "Deployment", "deployments"),
		)
	})
})

// helpers
// countingRESTMapper counts the lookups of a REST mapper, and can fail them like a discovery timeout.
type countingRESTMapper struct {
	meta.RESTMapper
	err   error
	calls int
}

func (m *countingRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.RESTMapper.RESTMapping(gk, versions...)
}

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
