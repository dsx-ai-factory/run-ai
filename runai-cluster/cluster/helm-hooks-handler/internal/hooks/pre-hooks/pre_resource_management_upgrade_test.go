/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	testu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/resourcing"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	testDeptID  = "dept-id-1"
	testDeptUID = "old-dept-uid"
)

// oldDepartment builds a v2alpha2 department with the labels owned queues key on,
// carrying two fully-populated queue specs so migration of every field can be
// asserted.
func oldDepartment(name string) *schedv2alpha2.Department {
	return &schedv2alpha2.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			UID:         testDeptUID,
			Labels:      map[string]string{schedv2alpha2.DepartmentIdLabel: testDeptID, "team": "research"},
			Annotations: map[string]string{"run.ai/note": "hello"},
			Finalizers:  []string{"run.ai/department-finalize"},
		},
		Spec: schedv2alpha2.DepartmentSpec{
			Queues: []schedv2alpha2.QueueDepartmentSpec{
				{
					Name:        "q-default",
					Nodepool:    "default",
					Priority:    ptr.To(int32(50)),
					SwapEnabled: true,
					Resources: &resourcing.QueueResources{
						GPU:    resourcing.SystemResource{Deserved: 2, OverQuotaWeight: 1, Limit: 4},
						CPU:    resourcing.SystemResource{Deserved: 1000, OverQuotaWeight: 2, Limit: 2000},
						Memory: resourcing.SystemResource{Deserved: 100, OverQuotaWeight: 3, Limit: 200},
					},
				},
				{
					Name:     "q-research",
					Nodepool: "gpu-pool",
					Priority: ptr.To(int32(75)),
					Resources: &resourcing.QueueResources{
						GPU: resourcing.SystemResource{Deserved: 8, OverQuotaWeight: 2, Limit: 16},
					},
				},
			},
		},
	}
}

// ownedQueue builds a queue owned by the given old department.
func ownedQueue(name string, dept *schedv2alpha2.Department) *schedv2.Queue {
	return &schedv2.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				schedv2alpha2.DepartmentIdLabel:   testDeptID,
				schedv2alpha2.DepartmentNameLabel: dept.Name,
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "scheduling.run.ai/v2alpha2",
				Kind:       departmentKind,
				Name:       dept.Name,
				UID:        dept.UID,
				Controller: &migrationTrueRef,
			}},
		},
	}
}

func oldDepartmentCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: hooks.OldDepartmentCRDName},
	}
}

// migrationScheme registers every type the department and project migration
// touches, so fake clients exercise the migration against the same registrations
// production uses (the legacy department/project, the kai twins, the re-parented
// queues and core/rbac owned objects, and the apiextensions CRD types). Production
// registers these on the shared client scheme at InitClients time rather than
// through a dedicated migration scheme.
func migrationScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	adders := []func(*runtime.Scheme) error{
		clientgoscheme.AddToScheme,
		apiextensionsv1.AddToScheme,
		schedv2alpha2.AddToScheme,
		runaiv2.AddToScheme,
		kaiv1alpha1.AddToScheme,
		schedv2.AddToScheme,
	}
	for _, add := range adders {
		if err := add(s); err != nil {
			t.Fatalf("build migration scheme: %v", err)
		}
	}
	return s
}

// newMigrationFakeClient builds a fake client wired with the migration scheme and
// seeded with the given objects.
func newMigrationFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(migrationScheme(t)).WithObjects(objs...).Build()
}

// operandEnabledInRunaiConfig reads spec.<operand>.enabled off the RunaiConfig in
// the dynamic client.
func operandEnabledInRunaiConfig(t *testing.T, dynamicClient dynamic.Interface, operand string) bool {
	t.Helper()
	cfg, err := dynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace).
		Get(context.Background(), runaiv1.DefaultName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get RunaiConfig: %v", err)
	}

	enabled, found, err := unstructured.NestedBool(cfg.Object, "spec", operand, "enabled")
	if err != nil {
		t.Fatalf("read spec.%s.enabled: %v", operand, err)
	}
	if !found {
		t.Fatalf("expected spec.%s.enabled to be set", operand)
	}
	return enabled
}

// mustGetDept fetches the kai department by name, failing the test if it is absent.
func mustGetDept(t *testing.T, k8sClient client.Client, name string) kaiv1alpha1.Department {
	t.Helper()
	var got kaiv1alpha1.Department
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name}, &got); err != nil {
		t.Fatalf("get kai department %s: %v", name, err)
	}
	return got
}

// mustGetQueue fetches the queue by name, failing the test if it is absent.
func mustGetQueue(t *testing.T, k8sClient client.Client, name string) schedv2.Queue {
	t.Helper()
	var got schedv2.Queue
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name}, &got); err != nil {
		t.Fatalf("get queue %s: %v", name, err)
	}
	return got
}

func TestBuildNewDepartmentQueueNames(t *testing.T) {
	t.Run("renames queues to the agent's id-based scheme", func(t *testing.T) {
		got, err := buildNewDepartment(oldDepartment("dep-a"))
		if err != nil {
			t.Fatalf("buildNewDepartment: %v", err)
		}
		want := []string{"q-" + testDeptID, "q-" + testDeptID + "-gpu-pool"}
		for i, name := range want {
			if got.Spec.Queues[i].Name != name {
				t.Errorf("queue[%d] name = %q, want %q", i, got.Spec.Queues[i].Name, name)
			}
		}
	})

	t.Run("keeps the original names when the department has no id label", func(t *testing.T) {
		old := oldDepartment("dep-a")
		delete(old.Labels, schedv2alpha2.DepartmentIdLabel)

		got, err := buildNewDepartment(old)
		if err != nil {
			t.Fatalf("buildNewDepartment: %v", err)
		}
		for i, name := range []string{"q-default", "q-research"} {
			if got.Spec.Queues[i].Name != name {
				t.Errorf("queue[%d] name = %q, want %q", i, got.Spec.Queues[i].Name, name)
			}
		}
	})
}

func TestMigrateDepartments(t *testing.T) {
	ctx := context.Background()

	t.Run("creates the new kai department and re-owns its queues", func(t *testing.T) {
		old := oldDepartment("dep-a")
		old.Status.SyncStatus.LastSyncTimestamp = metav1.NewTime(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC))
		q1 := ownedQueue("q-default", old)
		q2 := ownedQueue("q-research", old)
		c := newMigrationFakeClient(t, old, q1, q2)

		if err := migrateDepartments(ctx, c); err != nil {
			t.Fatalf("migrateDepartments: %v", err)
		}

		got := mustGetDept(t, c, "dep-a")

		// ObjectMeta: name, annotations, labels, finalizers all carried over,
		// and the status sync timestamp moved onto a label.
		if got.Name != "dep-a" {
			t.Errorf("name = %q, want dep-a", got.Name)
		}
		if got.Annotations["run.ai/note"] != "hello" {
			t.Errorf("annotation run.ai/note = %q, want hello", got.Annotations["run.ai/note"])
		}
		if got.Labels["team"] != "research" {
			t.Errorf("label team = %q, want research", got.Labels["team"])
		}
		if got.Labels[schedv2alpha2.DepartmentIdLabel] != testDeptID {
			t.Errorf("label %s = %q, want %q", schedv2alpha2.DepartmentIdLabel, got.Labels[schedv2alpha2.DepartmentIdLabel], testDeptID)
		}
		if got.Annotations[hooks.LastSyncTimestampAnnotation] != "2024-01-02T03-04-05Z" {
			t.Errorf("annotation %s = %q, want 2024-01-02T03-04-05Z", hooks.LastSyncTimestampAnnotation, got.Annotations[hooks.LastSyncTimestampAnnotation])
		}
		if len(got.Finalizers) != 1 || got.Finalizers[0] != "run.ai/department-finalize" {
			t.Errorf("finalizers = %v, want [run.ai/department-finalize]", got.Finalizers)
		}

		// Spec: both queues converted with every shared field preserved, renamed to
		// the id-based scheme the agent uses for kai department queues.
		if len(got.Spec.Queues) != 2 {
			t.Fatalf("queues = %d, want 2", len(got.Spec.Queues))
		}

		qDefault := got.Spec.Queues[0]
		if qDefault.Name != "q-"+testDeptID {
			t.Errorf("queue[0] name = %q, want q-%s", qDefault.Name, testDeptID)
		}
		if qDefault.Nodepool != "default" {
			t.Errorf("queue[0] nodepool = %q, want default", qDefault.Nodepool)
		}
		if qDefault.Priority == nil || *qDefault.Priority != 50 {
			t.Errorf("queue[0] priority = %v, want 50", qDefault.Priority)
		}
		if qDefault.Resources == nil {
			t.Fatal("queue[0] resources = nil, want non-nil")
		}
		if qDefault.Resources.GPU != (kaiv1alpha1.SystemResource{Deserved: 2, OverQuotaWeight: 1, Limit: 4}) {
			t.Errorf("queue[0] GPU = %+v", qDefault.Resources.GPU)
		}
		if qDefault.Resources.CPU != (kaiv1alpha1.SystemResource{Deserved: 1000, OverQuotaWeight: 2, Limit: 2000}) {
			t.Errorf("queue[0] CPU = %+v", qDefault.Resources.CPU)
		}
		if qDefault.Resources.Memory != (kaiv1alpha1.SystemResource{Deserved: 100, OverQuotaWeight: 3, Limit: 200}) {
			t.Errorf("queue[0] Memory = %+v", qDefault.Resources.Memory)
		}

		qResearch := got.Spec.Queues[1]
		if qResearch.Name != "q-"+testDeptID+"-gpu-pool" {
			t.Errorf("queue[1] name = %q, want q-%s-gpu-pool", qResearch.Name, testDeptID)
		}
		if qResearch.Nodepool != "gpu-pool" {
			t.Errorf("queue[1] nodepool = %q, want gpu-pool", qResearch.Nodepool)
		}
		if qResearch.Priority == nil || *qResearch.Priority != 75 {
			t.Errorf("queue[1] priority = %v, want 75", qResearch.Priority)
		}
		if qResearch.Resources == nil {
			t.Fatal("queue[1] resources = nil, want non-nil")
		}
		if qResearch.Resources.GPU != (kaiv1alpha1.SystemResource{Deserved: 8, OverQuotaWeight: 2, Limit: 16}) {
			t.Errorf("queue[1] GPU = %+v", qResearch.Resources.GPU)
		}

		// Every owned queue re-pointed at the new kai department with a full owner ref.
		for _, name := range []string{"q-default", "q-research"} {
			gotQueue := mustGetQueue(t, c, name)
			if len(gotQueue.OwnerReferences) != 1 {
				t.Fatalf("queue %s owner refs = %d, want 1", name, len(gotQueue.OwnerReferences))
			}
			owner := gotQueue.OwnerReferences[0]
			if owner.APIVersion != kaiv1alpha1.GroupVersion.String() {
				t.Errorf("queue %s owner APIVersion = %q, want %q", name, owner.APIVersion, kaiv1alpha1.GroupVersion.String())
			}
			if owner.Kind != departmentKind {
				t.Errorf("queue %s owner Kind = %q, want %q", name, owner.Kind, departmentKind)
			}
			if owner.Name != "dep-a" {
				t.Errorf("queue %s owner Name = %q, want dep-a", name, owner.Name)
			}
			if owner.UID != got.UID {
				t.Errorf("queue %s owner UID = %q, want %q", name, owner.UID, got.UID)
			}
			if owner.Controller == nil || !*owner.Controller {
				t.Errorf("queue %s owner Controller = %v, want true", name, owner.Controller)
			}
			if owner.BlockOwnerDeletion == nil || !*owner.BlockOwnerDeletion {
				t.Errorf("queue %s owner BlockOwnerDeletion = %v, want true", name, owner.BlockOwnerDeletion)
			}
		}
	})

	t.Run("skips a department that already has a kai twin", func(t *testing.T) {
		old := oldDepartment("dep-a")
		existing := &kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "dep-a"}}
		c := newMigrationFakeClient(t, old, existing)

		if err := migrateDepartments(ctx, c); err != nil {
			t.Fatalf("migrateDepartments: %v", err)
		}

		// The pre-existing kai twin is left untouched (no queues copied onto it).
		got := mustGetDept(t, c, "dep-a")
		if len(got.Spec.Queues) != 0 {
			t.Errorf("queues = %d, want 0 (twin left untouched)", len(got.Spec.Queues))
		}
	})

	t.Run("rolls back the new department when re-owning a queue fails", func(t *testing.T) {
		old := oldDepartment("dep-a")
		queue := ownedQueue("q-default", old)

		c := fake.NewClientBuilder().WithScheme(migrationScheme(t)).WithObjects(old, queue).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, cl client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					if _, ok := obj.(*schedv2.Queue); ok {
						return errors.New("forced queue patch failure")
					}
					return cl.Patch(ctx, obj, patch, opts...)
				},
			}).Build()

		if err := migrateDepartments(ctx, c); err == nil {
			t.Fatal("migrateDepartments succeeded, want error")
		}

		var got kaiv1alpha1.Department
		err := c.Get(ctx, types.NamespacedName{Name: "dep-a"}, &got)
		if !apierrors.IsNotFound(err) {
			t.Fatalf("expected new department to be rolled back (deleted), get err = %v", err)
		}
	})

	t.Run("rolls back every re-owned queue and deletes the new department when the last queue fails", func(t *testing.T) {
		old := oldDepartment("dep-a")
		q1 := ownedQueue("q-default", old)
		q2 := ownedQueue("q-research", old)

		// Fail only the last queue's update (q-research sorts after q-default),
		// so q-default is re-owned first and must be rolled back.
		c := fake.NewClientBuilder().WithScheme(migrationScheme(t)).WithObjects(old, q1, q2).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, cl client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					if q, ok := obj.(*schedv2.Queue); ok && q.Name == "q-research" {
						return errors.New("forced queue patch failure")
					}
					return cl.Patch(ctx, obj, patch, opts...)
				},
			}).Build()

		if err := migrateDepartments(ctx, c); err == nil {
			t.Fatal("migrateDepartments succeeded, want error")
		}

		// New department deleted.
		var got kaiv1alpha1.Department
		err := c.Get(ctx, types.NamespacedName{Name: "dep-a"}, &got)
		if !apierrors.IsNotFound(err) {
			t.Fatalf("expected new department to be deleted on rollback, get err = %v", err)
		}

		// Every queue's owner restored to the old (v2alpha2) department.
		for _, name := range []string{"q-default", "q-research"} {
			gotQueue := mustGetQueue(t, c, name)
			if len(gotQueue.OwnerReferences) != 1 {
				t.Fatalf("queue %s owner refs = %d, want 1", name, len(gotQueue.OwnerReferences))
			}
			owner := gotQueue.OwnerReferences[0]
			if owner.APIVersion != "scheduling.run.ai/v2alpha2" {
				t.Errorf("queue %s owner APIVersion = %q, want scheduling.run.ai/v2alpha2", name, owner.APIVersion)
			}
			if owner.Kind != departmentKind {
				t.Errorf("queue %s owner Kind = %q, want %q", name, owner.Kind, departmentKind)
			}
			if owner.Name != "dep-a" {
				t.Errorf("queue %s owner Name = %q, want dep-a", name, owner.Name)
			}
			if owner.UID != types.UID(testDeptUID) {
				t.Errorf("queue %s owner UID = %q, want %q", name, owner.UID, testDeptUID)
			}
		}
	})

	t.Run("migrates departments spanning more than one list page", func(t *testing.T) {
		// More than listPageSize departments forces the List to be served over
		// multiple pages followed via the Continue token.
		count := hooks.ListPageSize + 5
		objs := make([]client.Object, 0, count)
		for i := 0; i < count; i++ {
			objs = append(objs, &schedv2alpha2.Department{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("dep-%04d", i)},
			})
		}
		c := newMigrationFakeClient(t, objs...)

		if err := migrateDepartments(ctx, c); err != nil {
			t.Fatalf("migrateDepartments: %v", err)
		}

		// First and last department (the last sorts onto the second page) both
		// got a kai twin, proving every page was processed.
		for _, name := range []string{"dep-0000", fmt.Sprintf("dep-%04d", count-1)} {
			mustGetDept(t, c, name)
		}
	})

	t.Run("re-owns queues spanning more than one list page", func(t *testing.T) {
		old := oldDepartment("dep-a")
		count := hooks.ListPageSize + 5
		objs := make([]client.Object, 0, count+1)
		objs = append(objs, old)
		for i := 0; i < count; i++ {
			objs = append(objs, ownedQueue(fmt.Sprintf("q-%04d", i), old))
		}
		c := newMigrationFakeClient(t, objs...)

		if err := migrateDepartments(ctx, c); err != nil {
			t.Fatalf("migrateDepartments: %v", err)
		}

		// First and last queue (the last sorts onto the second page) both
		// re-owned by the new kai department.
		for _, name := range []string{"q-0000", fmt.Sprintf("q-%04d", count-1)} {
			gotQueue := mustGetQueue(t, c, name)
			if len(gotQueue.OwnerReferences) != 1 {
				t.Fatalf("queue %s owner refs = %d, want 1", name, len(gotQueue.OwnerReferences))
			}
			if gotQueue.OwnerReferences[0].APIVersion != kaiv1alpha1.GroupVersion.String() {
				t.Errorf("queue %s owner APIVersion = %q, want %q", name, gotQueue.OwnerReferences[0].APIVersion, kaiv1alpha1.GroupVersion.String())
			}
			if gotQueue.OwnerReferences[0].Name != "dep-a" {
				t.Errorf("queue %s owner Name = %q, want dep-a", name, gotQueue.OwnerReferences[0].Name)
			}
		}
	})
}

func TestMigrateDepartmentsAndProjects(t *testing.T) {
	ctx := context.Background()

	t.Run("is a no-op when the old department CRD is not installed", func(t *testing.T) {
		old := oldDepartment("dep-a")
		// No old department CRD seeded -> the migration must skip entirely,
		// before it ever touches the project-controller, so empty clients are fine.
		c := newMigrationFakeClient(t, old)
		h := &PreResourceManagementUpgradeHook{}

		if err := h.migrateResourceManagementResources(ctx, &clients.Clients{}, c, true, false); err != nil {
			t.Fatalf("migrateResourceManagementResources: %v", err)
		}

		var got kaiv1alpha1.Department
		err := c.Get(ctx, types.NamespacedName{Name: "dep-a"}, &got)
		if !apierrors.IsNotFound(err) {
			t.Fatalf("expected no migration when old CRD is absent, get err = %v", err)
		}
	})

	t.Run("disables project-controller and agent, migrates departments, and leaves them disabled for post-upgrade", func(t *testing.T) {
		old := oldDepartment("dep-a")
		queue := ownedQueue("q-default", old)
		// Migration client (kai scheme): old CRD present so the gate is open.
		c := newMigrationFakeClient(t, oldDepartmentCRD(), old, queue)

		// Local cluster client has no project-controller / agent pods, so the
		// down-wait returns on its first tick. Dynamic client holds a RunaiConfig
		// the disable/enable upgraders toggle.
		dynamicClient := testu.NewTestDynamicClient(true, false)
		resources := &clients.Clients{
			LocalClusterClient:        newMigrationFakeClient(t),
			LocalClusterDynamicClient: dynamicClient,
		}
		h := &PreResourceManagementUpgradeHook{
			params: &hooks.ClusterHookParams{OperandDownIntervalMillis: 1},
		}

		if err := h.migrateResourceManagementResources(ctx, resources, c, true, false); err != nil {
			t.Fatalf("migrateResourceManagementResources: %v", err)
		}

		// Department migrated and its queue re-owned by the kai department.
		mustGetDept(t, c, "dep-a")
		gotQueue := mustGetQueue(t, c, "q-default")
		if len(gotQueue.OwnerReferences) != 1 {
			t.Fatalf("queue owner refs = %d, want 1", len(gotQueue.OwnerReferences))
		}
		if gotQueue.OwnerReferences[0].APIVersion != kaiv1alpha1.GroupVersion.String() {
			t.Errorf("queue owner APIVersion = %q, want %q", gotQueue.OwnerReferences[0].APIVersion, kaiv1alpha1.GroupVersion.String())
		}

		// On success both operands stay disabled - the post-upgrade hook re-enables
		// them so they come back on the new version, not the old one.
		if operandEnabledInRunaiConfig(t, dynamicClient, "project-controller") {
			t.Error("expected project-controller to remain disabled after successful migration")
		}
		if operandEnabledInRunaiConfig(t, dynamicClient, "agent") {
			t.Error("expected agent to remain disabled after successful migration")
		}
	})
}
