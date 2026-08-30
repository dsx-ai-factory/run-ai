package pre_hooks

import (
	"context"
	"errors"
	"testing"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	testu "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const testNodePoolUID = "old-nodepool-uid"

// oldNodePool builds a legacy run.ai/v1alpha1 nodepool with representative
// scheduling fields so the mapping to the kai nodepool can be asserted.
func oldNodePool(name string) *runv1alpha1.NodePool {
	detection := runv1alpha1.UseGPUNetworkAcceleration
	windowType := runv1alpha1.TumblingWindow
	return &runv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			UID:         testNodePoolUID,
			Labels:      map[string]string{"team": "research"},
			Annotations: map[string]string{"run.ai/note": "hello"},
		},
		Spec: runv1alpha1.NodePoolSpec{
			LabelKey:                        "role",
			LabelValue:                      "gpu",
			NetworkTopologyName:             "topo-a",
			MinGuaranteedRuntime:            "5m",
			NumaAwareScheduling:             true,
			GPUNetworkAccelerationDetection: &detection,
			GPUNetworkAccelerationLabelKey:  "custom/label",
			PlacementStrategy: runv1alpha1.PlacementStrategy{
				Gpu:       "binpack",
				GpuDevice: "spread",
				Cpu:       "spread",
			},
			TimeBasedFairShare: runv1alpha1.TimeBasedFairShare{
				Enabled:               true,
				HistoricalUsageWeight: ptr.To(float32(0.5)),
				DecayHalfLife:         ptr.To("24h"),
				WindowDuration:        ptr.To("7d"),
				WindowType:            &windowType,
			},
		},
	}
}

// ownedShard builds the scheduling shard owned (controller ref) by the given legacy
// nodepool, the way the legacy nodepool-controller stamps it - named after the
// nodepool (shard.Name == nodePool.Name).
func ownedShard(np *runv1alpha1.NodePool) *kaiv1.SchedulingShard {
	return &kaiv1.SchedulingShard{
		ObjectMeta: metav1.ObjectMeta{
			Name:   np.Name,
			Labels: map[string]string{"app": "kai-scheduler"},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: runv1alpha1.GroupVersion.String(),
				Kind:       nodePoolKind,
				Name:       np.Name,
				UID:        np.UID,
				Controller: &migrationTrueRef,
			}},
		},
	}
}

// nodePoolMigrationScheme registers every type the nodepool migration touches: the
// legacy nodepool, the kai nodepool twin, the re-parented scheduling shards, and
// the apiextensions CRD types.
func nodePoolMigrationScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		clientgoscheme.AddToScheme,
		apiextensionsv1.AddToScheme,
		runv1alpha1.AddToScheme,
		kaiv1alpha1.AddToScheme,
		kaiv1.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatalf("build nodepool migration scheme: %v", err)
		}
	}
	return s
}

func newNodePoolMigrationFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(nodePoolMigrationScheme(t)).WithObjects(objs...).Build()
}

func mustGetShard(t *testing.T, k8sClient client.Client, name string) kaiv1.SchedulingShard {
	t.Helper()
	var got kaiv1.SchedulingShard
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name}, &got); err != nil {
		t.Fatalf("get scheduling shard %s: %v", name, err)
	}
	return got
}

func oldNodePoolCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: hooks.OldNodePoolCRDName}}
}

// operandNotDisabled reports whether spec.<operand>.enabled is anything other than
// an explicit false - i.e. the operand was left enabled/untouched. An absent key
// means go-operator applies its default (enabled), so that also counts as "not
// disabled".
func operandNotDisabled(t *testing.T, dynamicClient dynamic.Interface, operand string) bool {
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
	return !found || enabled
}

// TestMigrateResourceManagementResourcesNodePoolsOnly proves the nodepool flow is
// independent of the org-unit flow: with only the nodepool flow enabled it migrates
// the nodepool, takes down the agent and the nodepool-controller, and leaves the
// project-controller untouched.
func TestMigrateResourceManagementResourcesNodePoolsOnly(t *testing.T) {
	ctx := context.Background()
	old := oldNodePool("np-a")
	shard := ownedShard(old)
	// Migration client with the legacy nodepool CRD present so the nodepool flow runs.
	c := newNodePoolMigrationFakeClient(t, oldNodePoolCRD(), old, shard)

	dynamicClient := testu.NewTestDynamicClient(true, false)
	resources := &clients.Clients{
		LocalClusterClient:        newNodePoolMigrationFakeClient(t),
		LocalClusterDynamicClient: dynamicClient,
	}
	h := &PreResourceManagementUpgradeHook{
		params: &hooks.ClusterHookParams{OperandDownIntervalMillis: 1},
	}

	// orgUnitsEnabled=false, nodePoolsEnabled=true.
	require.NoError(t, h.migrateResourceManagementResources(ctx, resources, c, false, true))

	// The nodepool was migrated and its shard re-parented to the kai nodepool.
	var twin kaiv1alpha1.NodePool
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: "np-a"}, &twin))
	got := mustGetShard(t, c, "np-a")
	require.Equal(t, kaiv1alpha1.GroupVersion.String(), got.OwnerReferences[0].APIVersion)

	// Agent and nodepool-controller are disabled; project-controller is left alone.
	require.False(t, operandEnabledInRunaiConfig(t, dynamicClient, "agent"))
	require.False(t, operandEnabledInRunaiConfig(t, dynamicClient, "nodepool-controller"))
	require.True(t, operandNotDisabled(t, dynamicClient, "project-controller"))
}

func TestBuildNewNodePoolMapsSpecAndAnnotations(t *testing.T) {
	got := buildNewNodePool(oldNodePool("np-a"))

	require.Equal(t, "np-a", got.Name)
	require.Equal(t, "role", got.Spec.LabelKey)
	require.Equal(t, "gpu", got.Spec.LabelValue)
	require.Equal(t, "topo-a", got.Spec.PreferredNetworkTopologyName)

	// Legacy labels are preserved; the legacy finalizer is not carried over.
	require.Equal(t, "research", got.Labels["team"])
	require.Empty(t, got.Finalizers)

	// GPU network acceleration spec fields moved onto kai annotations.
	require.Equal(t, "Use", got.Annotations[kaiv1alpha1.AnnotationGPUNetworkAccelerationDetection])
	require.Equal(t, "custom/label", got.Annotations[kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey])
	require.Equal(t, "hello", got.Annotations["run.ai/note"])

	cfg := got.Spec.SchedulingShardConfig
	require.NotNil(t, cfg)
	require.Equal(t, "binpack", ptr.Deref(cfg.PlacementStrategy.GPU, ""))
	require.Equal(t, "spread", ptr.Deref(cfg.PlacementStrategy.CPU, ""))
	require.Equal(t, "5m", ptr.Deref(cfg.MinRuntime.PreemptMinRuntime, ""))
	require.Equal(t, "5m", ptr.Deref(cfg.MinRuntime.ReclaimMinRuntime, ""))

	// Device strategy (spread) differs from the node strategy (binpack) -> device
	// placement plugins reflect spread; numa-aware scheduling -> numa plugin.
	require.False(t, ptr.Deref(cfg.Plugins["gpupack"].Enabled, true))
	require.True(t, ptr.Deref(cfg.Plugins["gpuspread"].Enabled, false))
	require.True(t, ptr.Deref(cfg.Plugins["numa"].Enabled, false))

	require.True(t, ptr.Deref(cfg.TimeBasedFairShare.Enabled, false))
	require.Equal(t, float32(0.5), ptr.Deref(cfg.TimeBasedFairShare.HistoricalUsageWeight, 0))
	require.Equal(t, "24h", ptr.Deref(cfg.TimeBasedFairShare.HalfLifePeriod, ""))
	require.Equal(t, "tumbling", ptr.Deref(cfg.TimeBasedFairShare.Window.Type, ""))
	require.Equal(t, "7d", ptr.Deref(cfg.TimeBasedFairShare.Window.Size, ""))
}

func TestBuildNewNodePoolDisabledFairShareIsNil(t *testing.T) {
	old := oldNodePool("np-a")
	old.Spec.TimeBasedFairShare = runv1alpha1.TimeBasedFairShare{Enabled: false}

	got := buildNewNodePool(old)
	require.Nil(t, got.Spec.SchedulingShardConfig.TimeBasedFairShare)
}

// TestBuildNewNodePoolFairShareCopiesValuesWithoutDefaults asserts the migration
// does not inject fair-share defaults: enabled fair share with every optional field
// unset stays unset (the controller defaults them downstream).
func TestBuildNewNodePoolFairShareCopiesValuesWithoutDefaults(t *testing.T) {
	old := oldNodePool("np-a")
	old.Spec.TimeBasedFairShare = runv1alpha1.TimeBasedFairShare{Enabled: true}

	tbfs := buildNewNodePool(old).Spec.SchedulingShardConfig.TimeBasedFairShare
	require.NotNil(t, tbfs)
	require.True(t, ptr.Deref(tbfs.Enabled, false))
	require.Nil(t, tbfs.HistoricalUsageWeight)
	require.Nil(t, tbfs.HalfLifePeriod)
	require.Nil(t, tbfs.Window)
}

func TestMigrateNodePoolsCreatesTwinAndReownsShards(t *testing.T) {
	old := oldNodePool("np-a")
	shard := ownedShard(old)
	c := newNodePoolMigrationFakeClient(t, old, shard)

	require.NoError(t, migrateNodePools(context.Background(), c))

	// The kai nodepool twin exists.
	var twin kaiv1alpha1.NodePool
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &twin))

	// The shard is re-parented to the kai nodepool.
	got := mustGetShard(t, c, "np-a")
	require.Len(t, got.OwnerReferences, 1)
	require.Equal(t, kaiv1alpha1.GroupVersion.String(), got.OwnerReferences[0].APIVersion)
	require.Equal(t, nodePoolKind, got.OwnerReferences[0].Kind)
	require.Equal(t, "np-a", got.OwnerReferences[0].Name)
}

func TestMigrateNodePoolsSkipsWhenTwinExists(t *testing.T) {
	old := oldNodePool("np-a")
	shard := ownedShard(old)
	existingTwin := &kaiv1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-a"}}
	c := newNodePoolMigrationFakeClient(t, old, shard, existingTwin)

	require.NoError(t, migrateNodePools(context.Background(), c))

	// The shard's owner is left pointing at the legacy nodepool - migration skipped.
	got := mustGetShard(t, c, "np-a")
	require.Len(t, got.OwnerReferences, 1)
	require.Equal(t, runv1alpha1.GroupVersion.String(), got.OwnerReferences[0].APIVersion)
}

func TestMigrateNodePoolsRollsBackOnShardPatchFailure(t *testing.T) {
	old := oldNodePool("np-a")
	shard := ownedShard(old)

	// Fail the shard owner-ref patch so the migration must roll back: the kai
	// nodepool it just created is deleted and the shard owner is left untouched.
	failing := interceptor.Funcs{
		Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			if _, ok := obj.(*kaiv1.SchedulingShard); ok {
				return errors.New("boom")
			}
			return c.Patch(ctx, obj, patch, opts...)
		},
	}
	c := fake.NewClientBuilder().WithScheme(nodePoolMigrationScheme(t)).
		WithObjects(old, shard).WithInterceptorFuncs(failing).Build()

	require.Error(t, migrateNodePools(context.Background(), c))

	// The kai nodepool twin was rolled back (deleted).
	var twin kaiv1alpha1.NodePool
	require.True(t, apierrors.IsNotFound(c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &twin)))

	// The shard still points at the legacy nodepool.
	got := mustGetShard(t, c, "np-a")
	require.Len(t, got.OwnerReferences, 1)
	require.Equal(t, runv1alpha1.GroupVersion.String(), got.OwnerReferences[0].APIVersion)
}

// TestMigrateNodePoolsWithNoShard covers a nodepool that has no scheduling shard:
// the kai twin is still created and the migration succeeds with nothing to reparent
// (the reownSchedulingShards Get returns NotFound).
func TestMigrateNodePoolsWithNoShard(t *testing.T) {
	old := oldNodePool("np-a")
	c := newNodePoolMigrationFakeClient(t, old) // no shard seeded

	require.NoError(t, migrateNodePools(context.Background(), c))

	// The twin is created even though the nodepool has no shard.
	var twin kaiv1alpha1.NodePool
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &twin))
	require.True(t, apierrors.IsNotFound(
		c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &kaiv1.SchedulingShard{})))
}

// TestMigrateNodePoolsClearsLegacyFinalizers - without this the cleanup's deletes
// never complete and the nodepool (then the CRD) stays stuck Terminating.
func TestMigrateNodePoolsClearsLegacyFinalizers(t *testing.T) {
	old := oldNodePool("np-a")
	old.Finalizers = []string{"nodepool.run.ai/finalize"}
	c := newNodePoolMigrationFakeClient(t, old)

	require.NoError(t, migrateNodePools(context.Background(), c))

	var got runv1alpha1.NodePool
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &got))
	require.Empty(t, got.Finalizers, "legacy nodepool finalizers should be cleared by the migration")
}

// TestMigrateNodePoolsClearsForeignFinalizers - finalizers owned by other controllers
// are cleared too.
func TestMigrateNodePoolsClearsForeignFinalizers(t *testing.T) {
	old := oldNodePool("np-a")
	old.Finalizers = []string{"something.else/finalize", "nodepool.run.ai/finalize"}
	c := newNodePoolMigrationFakeClient(t, old)

	require.NoError(t, migrateNodePools(context.Background(), c))

	var got runv1alpha1.NodePool
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &got))
	require.Empty(t, got.Finalizers)
}

// TestMigrateNodePoolsWithNoFinalizersIsNoOp - nothing to clear, nothing patched.
func TestMigrateNodePoolsWithNoFinalizersIsNoOp(t *testing.T) {
	old := oldNodePool("np-a") // no finalizers
	c := newNodePoolMigrationFakeClient(t, old)

	require.NoError(t, migrateNodePools(context.Background(), c))

	var got runv1alpha1.NodePool
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &got))
	require.Empty(t, got.Finalizers)
}

// noKaiShardCRD makes every SchedulingShard Get fail the way the API server does
// when the kai SchedulingShard CRD is not served yet - a RESTMapper no-match rather
// than a NotFound.
func noKaiShardCRD() interceptor.Funcs {
	return interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey,
			obj client.Object, opts ...client.GetOption) error {
			if _, ok := obj.(*kaiv1.SchedulingShard); ok {
				return &meta.NoKindMatchError{
					GroupKind:        schema.GroupKind{Group: "kai.scheduler", Kind: "SchedulingShard"},
					SearchedVersions: []string{"v1"},
				}
			}
			return c.Get(ctx, key, obj, opts...)
		},
	}
}

// TestMigrateNodePoolsWhenShardCRDNotServed covers the engine->kai cutover, where
// the kai SchedulingShard CRD is applied by a pre-upgrade hook that runs after this
// one. The migration must still create the twin and succeed - failing here fails the
// whole chart upgrade.
func TestMigrateNodePoolsWhenShardCRDNotServed(t *testing.T) {
	old := oldNodePool("np-a")
	c := fake.NewClientBuilder().WithScheme(nodePoolMigrationScheme(t)).
		WithObjects(old).WithInterceptorFuncs(noKaiShardCRD()).Build()

	require.NoError(t, migrateNodePools(context.Background(), c))

	// The twin is created and kept - the skip must not trigger a rollback.
	var twin kaiv1alpha1.NodePool
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &twin))
	require.Equal(t, old.Spec.LabelKey, twin.Spec.LabelKey)
	require.NotNil(t, twin.Spec.SchedulingShardConfig)
}

// TestMigrateNodePoolsWhenShardCRDNotServedClearsFinalizers - the skip must not cut
// the migration short: the legacy finalizers are still cleared afterwards, otherwise
// the later cleanup leaves the nodepools stuck Terminating.
func TestMigrateNodePoolsWhenShardCRDNotServedClearsFinalizers(t *testing.T) {
	old := oldNodePool("np-a")
	old.Finalizers = []string{"nodepool.run.ai/finalize"}
	c := fake.NewClientBuilder().WithScheme(nodePoolMigrationScheme(t)).
		WithObjects(old).WithInterceptorFuncs(noKaiShardCRD()).Build()

	require.NoError(t, migrateNodePools(context.Background(), c))

	var got runv1alpha1.NodePool
	require.NoError(t, c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &got))
	require.Empty(t, got.Finalizers)
}

// TestMigrateNodePoolsStillFailsOnOtherShardGetErrors - only a no-match is skipped;
// a transient API error must still fail (and roll back) so a shard is never left
// orphaned under the legacy owner with no second chance to reparent it.
func TestMigrateNodePoolsStillFailsOnOtherShardGetErrors(t *testing.T) {
	old := oldNodePool("np-a")
	failing := interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey,
			obj client.Object, opts ...client.GetOption) error {
			if _, ok := obj.(*kaiv1.SchedulingShard); ok {
				return errors.New("boom")
			}
			return c.Get(ctx, key, obj, opts...)
		},
	}
	c := fake.NewClientBuilder().WithScheme(nodePoolMigrationScheme(t)).
		WithObjects(old).WithInterceptorFuncs(failing).Build()

	require.Error(t, migrateNodePools(context.Background(), c))

	// The kai nodepool twin was rolled back (deleted).
	var twin kaiv1alpha1.NodePool
	require.True(t, apierrors.IsNotFound(
		c.Get(context.Background(), types.NamespacedName{Name: "np-a"}, &twin)))
}
