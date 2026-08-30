package pre_hooks

import (
	"context"
	"testing"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const legacyManagedNodesConfigFinalizer = "managednodesconfig.run.ai/finalize"

// oldManagedNodesConfig builds the legacy run.ai/v1alpha1 singleton with a
// representative inclusion criteria, so the copy into the kai twin can be asserted.
func oldManagedNodesConfig() *runv1alpha1.ManagedNodesConfig {
	return &runv1alpha1.ManagedNodesConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:        constants.ManagedNodesConfigName,
			Labels:      map[string]string{"team": "research"},
			Annotations: map[string]string{"run.ai/note": "hello"},
			Finalizers:  []string{legacyManagedNodesConfigFinalizer},
		},
		Spec: runv1alpha1.ManagedNodesConfigSpec{
			InclusionCriteria: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "node-role",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"gpu"},
					}},
				}},
			},
		},
	}
}

func oldManagedNodesConfigCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: hooks.OldManagedNodesConfigCRDName},
	}
}

func getKaiManagedNodesConfig(t *testing.T, k8sClient client.Client) (kaiv1alpha1.ManagedNodesConfig, bool) {
	t.Helper()
	var got kaiv1alpha1.ManagedNodesConfig
	err := k8sClient.Get(context.Background(),
		types.NamespacedName{Name: constants.ManagedNodesConfigName}, &got)
	if apierrors.IsNotFound(err) {
		return got, false
	}
	require.NoError(t, err)
	return got, true
}

func TestMigrateManagedNodesConfigNoOpWhenLegacyCRDAbsent(t *testing.T) {
	// Fresh installs have no legacy CRD, so the migration must not create a twin.
	c := newNodePoolMigrationFakeClient(t)

	require.NoError(t, migrateManagedNodesConfig(context.Background(), c))

	_, found := getKaiManagedNodesConfig(t, c)
	require.False(t, found)
}

func TestMigrateManagedNodesConfigNoOpWhenLegacyObjectAbsent(t *testing.T) {
	// The CRD can be installed with no singleton yet (nothing has written one).
	c := newNodePoolMigrationFakeClient(t, oldManagedNodesConfigCRD())

	require.NoError(t, migrateManagedNodesConfig(context.Background(), c))

	_, found := getKaiManagedNodesConfig(t, c)
	require.False(t, found)
}

// The copy keeps the singleton's name and changes only the group: go-operator pins the
// name for the deployed controller via --managed-nodes-config-name, so a renamed twin
// would be read by nobody.
func TestMigrateManagedNodesConfigCopiesSpecUnderTheSameName(t *testing.T) {
	old := oldManagedNodesConfig()
	c := newNodePoolMigrationFakeClient(t, oldManagedNodesConfigCRD(), old)

	require.NoError(t, migrateManagedNodesConfig(context.Background(), c))

	got, found := getKaiManagedNodesConfig(t, c)
	require.True(t, found)
	require.Equal(t, constants.ManagedNodesConfigName, got.Name)
	require.Equal(t, old.Spec.InclusionCriteria, got.Spec.InclusionCriteria)
	require.Equal(t, old.Labels, got.Labels)
	require.Equal(t, old.Annotations, got.Annotations)
}

// The legacy finalizer must not be carried over: the kai object's controller adds its
// own, and a copied one would wedge deletion while that controller is down.
func TestMigrateManagedNodesConfigDoesNotCopyLegacyFinalizer(t *testing.T) {
	c := newNodePoolMigrationFakeClient(t, oldManagedNodesConfigCRD(), oldManagedNodesConfig())

	require.NoError(t, migrateManagedNodesConfig(context.Background(), c))

	got, found := getKaiManagedNodesConfig(t, c)
	require.True(t, found)
	require.Empty(t, got.Finalizers)
}

// The legacy controller is down for the migration window and would never clear its own
// finalizer, so the post-upgrade delete depends on the migration stripping it.
func TestMigrateManagedNodesConfigClearsLegacyFinalizers(t *testing.T) {
	c := newNodePoolMigrationFakeClient(t, oldManagedNodesConfigCRD(), oldManagedNodesConfig())

	require.NoError(t, migrateManagedNodesConfig(context.Background(), c))

	var legacy runv1alpha1.ManagedNodesConfig
	require.NoError(t, c.Get(context.Background(),
		types.NamespacedName{Name: constants.ManagedNodesConfigName}, &legacy))
	require.Empty(t, legacy.Finalizers)
}

// Re-running the hook (a retried upgrade) must not overwrite a twin the running
// controller may have since updated.
func TestMigrateManagedNodesConfigIsIdempotent(t *testing.T) {
	ctx := context.Background()
	existing := &kaiv1alpha1.ManagedNodesConfig{
		ObjectMeta: metav1.ObjectMeta{Name: constants.ManagedNodesConfigName},
		Spec: kaiv1alpha1.ManagedNodesConfigSpec{
			InclusionCriteria: corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "already",
						Operator: corev1.NodeSelectorOpExists,
					}},
				}},
			},
		},
	}
	c := newNodePoolMigrationFakeClient(t, oldManagedNodesConfigCRD(), oldManagedNodesConfig(), existing)

	require.NoError(t, migrateManagedNodesConfig(ctx, c))

	got, found := getKaiManagedNodesConfig(t, c)
	require.True(t, found)
	require.Equal(t, existing.Spec.InclusionCriteria, got.Spec.InclusionCriteria)

	// Even on the skip path the legacy finalizers must still be cleared, otherwise a
	// re-run after a partial migration would leave the legacy object undeletable.
	var legacy runv1alpha1.ManagedNodesConfig
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: constants.ManagedNodesConfigName}, &legacy))
	require.Empty(t, legacy.Finalizers)
}
