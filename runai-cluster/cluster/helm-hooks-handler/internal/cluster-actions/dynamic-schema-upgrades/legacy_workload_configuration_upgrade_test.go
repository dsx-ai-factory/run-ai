package dynamic_schema_upgrades

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestLegacyWorkloadConfigurationUpgradeCopiesLegacyValue(t *testing.T) {
	resource := runaiConfigWithRouteConfigurations("passthrough", "")
	upgrader := newLegacyWorkloadConfigurationUpgrader()

	require.True(t, upgrader.shouldPerformUpgrade(resource))
	updated, err := upgrader.updateRunaiConfig(resource)
	require.NoError(t, err)
	require.Equal(t, "passthrough", routeTerminationPolicy(t, updated, anyworkloadControllerOperandName))
}

func TestLegacyWorkloadConfigurationUpgradePreservesAnyworkloadValue(t *testing.T) {
	resource := runaiConfigWithRouteConfigurations("passthrough", "edge")
	upgrader := newLegacyWorkloadConfigurationUpgrader()

	require.False(t, upgrader.shouldPerformUpgrade(resource))
	require.Equal(t, "edge", routeTerminationPolicy(t, resource, anyworkloadControllerOperandName))
}

func TestLegacyWorkloadConfigurationUpgradeCopiesOtherConsumedFields(t *testing.T) {
	resource := runaiConfigWithRouteConfigurations("", "")
	resource.Object["spec"].(map[string]any)[workloadControllerOperandName] = map[string]any{
		"gitSyncImage":           map[string]any{"name": "custom-git-sync"},
		"externalAuthUrlEnabled": false,
	}
	upgrader := newLegacyWorkloadConfigurationUpgrader()

	require.True(t, upgrader.shouldPerformUpgrade(resource))
	updated, err := upgrader.updateRunaiConfig(resource)
	require.NoError(t, err)
	imageName, found, err := unstructured.NestedString(
		updated.Object, "spec", anyworkloadControllerOperandName, "gitSyncImage", "name",
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "custom-git-sync", imageName)
	externalAuth, found, err := unstructured.NestedBool(
		updated.Object, "spec", anyworkloadControllerOperandName, "externalAuthUrlEnabled",
	)
	require.NoError(t, err)
	require.True(t, found)
	require.False(t, externalAuth)
}

func TestLegacyWorkloadConfigurationUpgradeCopiesInferenceControllerValue(t *testing.T) {
	resource := runaiConfigWithRouteConfigurations("", "")
	resource.Object["spec"].(map[string]any)[inferenceWorkloadControllerOperandName] = map[string]any{
		"gitSyncImage": map[string]any{"name": "inference-git-sync"},
	}
	upgrader := newLegacyWorkloadConfigurationUpgrader()

	require.True(t, upgrader.shouldPerformUpgrade(resource))
	updated, err := upgrader.updateRunaiConfig(resource)
	require.NoError(t, err)
	imageName, found, err := unstructured.NestedString(
		updated.Object, "spec", anyworkloadControllerOperandName, "gitSyncImage", "name",
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "inference-git-sync", imageName)
}

func TestLegacyWorkloadConfigurationUpgradePrefersWorkloadControllerValue(t *testing.T) {
	resource := runaiConfigWithRouteConfigurations("", "")
	spec := resource.Object["spec"].(map[string]any)
	spec[workloadControllerOperandName] = map[string]any{
		"gitSyncImage": map[string]any{"name": "workload-git-sync"},
	}
	spec[inferenceWorkloadControllerOperandName] = map[string]any{
		"gitSyncImage": map[string]any{"name": "inference-git-sync"},
	}
	upgrader := newLegacyWorkloadConfigurationUpgrader()

	updated, err := upgrader.updateRunaiConfig(resource)
	require.NoError(t, err)
	imageName, found, err := unstructured.NestedString(
		updated.Object, "spec", anyworkloadControllerOperandName, "gitSyncImage", "name",
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "workload-git-sync", imageName)
}

func runaiConfigWithRouteConfigurations(workloadPolicy, anyworkloadPolicy string) *unstructured.Unstructured {
	spec := map[string]any{}
	if workloadPolicy != "" {
		spec[workloadControllerOperandName] = map[string]any{
			openShiftRouteConfigurationField: map[string]any{
				openShiftRouteTerminationPolicyField: workloadPolicy,
			},
		}
	}
	if anyworkloadPolicy != "" {
		spec[anyworkloadControllerOperandName] = map[string]any{
			openShiftRouteConfigurationField: map[string]any{
				openShiftRouteTerminationPolicyField: anyworkloadPolicy,
			},
		}
	}
	return &unstructured.Unstructured{Object: map[string]any{"spec": spec}}
}

func routeTerminationPolicy(t *testing.T, resource *unstructured.Unstructured, controller string) string {
	t.Helper()
	policy, found, err := unstructured.NestedString(
		resource.Object,
		"spec",
		controller,
		openShiftRouteConfigurationField,
		openShiftRouteTerminationPolicyField,
	)
	require.NoError(t, err)
	require.True(t, found)
	return policy
}
