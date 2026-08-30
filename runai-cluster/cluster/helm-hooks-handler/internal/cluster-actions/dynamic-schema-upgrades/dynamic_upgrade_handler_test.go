package dynamic_schema_upgrades

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	testu "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

func TestRecoverRunaiConfigWithBackupUpgradeHandler(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	obj := testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_16.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}
	dynamicResource := dynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)

	mpsUpgrader := newMpsServerUpgrader()
	runaiUpgradeHandler := &runaiConfigUpgrade{
		specUpgraders: []runaiConfigDynamicSchemaUpgrader{mpsUpgrader},
	}
	err := runaiUpgradeHandler.Upgrade(context.TODO(), k8sClients)
	assert.Nil(t, err)

	shouldUpgrade := mpsUpgrader.shouldPerformUpgrade(obj)
	assert.Equal(t, shouldUpgrade, true)
	imageStr := mpsUpgrader.extractImageStringAfterValidation(obj)
	assert.Equal(t, imageStr, "registry.access.redhat.com/ubi8/ubi-minimal:8.9")

	val, _ := dynamicResource.Get(ctx, obj.GetName(), metav1.GetOptions{})
	res := mpsUpgrader.shouldPerformUpgrade(val)
	assert.Equal(t, res, false)

	image := val.Object["spec"].(map[string]interface{})["mps-server"].(map[string]interface{})["image"]
	imageStruct, ok := image.(map[string]interface{})
	assert.Equal(t, ok, true)
	assert.Equal(t, imageStruct["registry"], "registry.access.redhat.com")
	assert.Equal(t, imageStruct["tag"], "8.9")
	assert.Equal(t, imageStruct["name"], "ubi8/ubi-minimal")
	assert.Equal(t, imageStruct["pullPolicy"], "Always")
}

func TestNPCDisabledInPreCrdUpgrade(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	_ = testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_22_nodepool_controller_enabled.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}

	// sanity
	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled before pre-crd-upgrade",
		dynamicClient, t)
	legacyCompatibilityDataBefore := getLegacyCompatibilityData(dynamicClient, t)

	runaiUpgradeHandler := NewRunaiConfigDisableNodepoolController() // This includes newOperandEnableUpgrader(true, "nodepool-controller)
	err := runaiUpgradeHandler.Upgrade(ctx, k8sClients)
	assert.Nil(t, err)

	validateNPCEnabledInRunaiConfig(false, "nodepool-controller should be disabled after pre-crd-upgrade",
		dynamicClient, t)
	assert.Equal(t, legacyCompatibilityDataBefore, getLegacyCompatibilityData(dynamicClient, t),
		"pre-upgrade RunaiConfig rewrite must preserve deprecated Engine/Core compatibility data")
}

func TestNPCDisabledInPreCrdUpgradeNPCConfigMissing(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	// config missing - means nodepool-controller is enabled by default
	_ = testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_22_nodepool_controller_config_missing.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}

	// sanity
	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled before pre-crd-upgrade",
		dynamicClient, t)

	runaiUpgradeHandler := NewRunaiConfigDisableNodepoolController() // This includes newOperandEnableUpgrader(true, "nodepool-controller)
	err := runaiUpgradeHandler.Upgrade(ctx, k8sClients)
	assert.Nil(t, err)

	validateNPCEnabledInRunaiConfig(false, "nodepool-controller should be disabled after pre-crd-upgrade with nodepool-controller config missing",
		dynamicClient, t)
}

func TestNPCDisabledInPreCrdUpgradeNPCConfigEnabledMissing(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	// enabled config missing - means nodepool-controller is enabled by default
	_ = testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_22_nodepool_controller_config_enabled_missing.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}

	// sanity
	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled before pre-crd-upgrade",
		dynamicClient, t)

	runaiUpgradeHandler := NewRunaiConfigDisableNodepoolController() // This includes newOperandEnableUpgrader(true, "nodepool-controller)
	err := runaiUpgradeHandler.Upgrade(ctx, k8sClients)
	assert.Nil(t, err)

	validateNPCEnabledInRunaiConfig(false, "nodepool-controller should be disabled after pre-crd-upgrade with nodepool-controller config enabled missing",
		dynamicClient, t)
}

func TestNPCDisabledInPreCrdUpgradeNPCAlreadyDisabled(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	// enabled config missing - means nodepool-controller is enabled by default
	_ = testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_22_nodepool_controller_disabled.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}

	// sanity
	validateNPCEnabledInRunaiConfig(false, "nodepool-controller should be disabled sanity",
		dynamicClient, t)

	runaiUpgradeHandler := NewRunaiConfigDisableNodepoolController() // This includes newOperandEnableUpgrader(true, "nodepool-controller)
	err := runaiUpgradeHandler.Upgrade(ctx, k8sClients)
	assert.Nil(t, err)

	validateNPCEnabledInRunaiConfig(false, "nodepool-controller should be disabled after pre-crd-upgrade with nodepool-controller already disabled",
		dynamicClient, t)
}

func TestNPCEnabledInPostUpgrade(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	_ = testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_22_nodepool_controller_disabled.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}

	// sanity
	validateNPCEnabledInRunaiConfig(false, "nodepool-controller should be disabled before post-upgrade",
		dynamicClient, t)

	runaiUpgradeHandler := NewRunaiConfigEnableNodepoolController() // This includes newOperandEnableUpgrader(false, "nodepool-controller)
	err := runaiUpgradeHandler.Upgrade(ctx, k8sClients)
	assert.Nil(t, err)

	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled in post-upgrade",
		dynamicClient, t)
}

func TestNPCEnabledInPostUpgradeNPCConfigMissing(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	_ = testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_22_nodepool_controller_config_missing.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}

	// sanity
	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled before post-upgrade with nodepool-controller config missing",
		dynamicClient, t)

	runaiUpgradeHandler := NewRunaiConfigEnableNodepoolController() // This includes newOperandEnableUpgrader(false, "nodepool-controller)
	err := runaiUpgradeHandler.Upgrade(ctx, k8sClients)
	assert.Nil(t, err)

	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled in post-upgrade with nodepool-controller config missing",
		dynamicClient, t)
}

func TestNPCEnabledInPostUpgradeNPCConfigEnabledMissing(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	_ = testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_22_nodepool_controller_config_enabled_missing.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}

	// sanity
	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled before post-upgrade with nodepool-controller config enabled missing",
		dynamicClient, t)

	runaiUpgradeHandler := NewRunaiConfigEnableNodepoolController() // This includes newOperandEnableUpgrader(false, "nodepool-controller)
	err := runaiUpgradeHandler.Upgrade(ctx, k8sClients)
	assert.Nil(t, err)

	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled in post-upgrade with nodepool-controller config enabled missing",
		dynamicClient, t)
}

func TestNPCEnabledInPostUpgradeNPCAlreadyEnabled(t *testing.T) {
	dynamicClient := testu.NewTestDynamicClient(false, false)
	_ = testu.AddYamlRunaiConf(dynamicClient, "../../test-util/test_resources/runaiconfig2_22_nodepool_controller_enabled.yaml")
	ctx := context.TODO()
	k8sClients := &clients.Clients{
		LocalClusterDynamicClient: dynamicClient,
	}

	// sanity
	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled before post-upgrade sanity",
		dynamicClient, t)

	runaiUpgradeHandler := NewRunaiConfigEnableNodepoolController() // This includes newOperandEnableUpgrader(false, "nodepool-controller)
	err := runaiUpgradeHandler.Upgrade(ctx, k8sClients)
	assert.Nil(t, err)

	validateNPCEnabledInRunaiConfig(true, "nodepool-controller should be enabled in post-upgrade with nodepool-controller already enabled",
		dynamicClient, t)
}

func validateNPCEnabledInRunaiConfig(expectedEnabled bool, message string,
	dynamicClient dynamic.Interface, t *testing.T) {
	dynamicResource := dynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
	val, _ := dynamicResource.Get(context.TODO(), runaiv1.DefaultName, metav1.GetOptions{})
	spec, ok := val.Object["spec"].(map[string]interface{})
	assert.True(t, ok, "spec should be a map[string]interface{}")

	runaiNPCConf, ok := spec["nodepool-controller"].(map[string]interface{})
	if !ok {
		if expectedEnabled {
			// default is enabled
			return
		} else {
			assert.Fail(t, message)
		}
	}

	enabled, ok := runaiNPCConf["enabled"].(bool)
	if !ok {
		if expectedEnabled {
			// default is enabled
			return
		} else {
			assert.Fail(t, message)
		}
	}

	if expectedEnabled {
		assert.True(t, enabled, message)
	} else {
		assert.False(t, enabled, message)
	}
}

func getLegacyCompatibilityData(dynamicClient dynamic.Interface, t *testing.T) map[string]interface{} {
	t.Helper()
	dynamicResource := dynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
	value, err := dynamicResource.Get(context.TODO(), runaiv1.DefaultName, metav1.GetOptions{})
	assert.NoError(t, err)
	spec, ok := value.Object["spec"].(map[string]interface{})
	assert.True(t, ok, "spec should be a map[string]interface{}")
	globalSpec, ok := spec["global"].(map[string]interface{})
	assert.True(t, ok, "global should be a map[string]interface{}")

	return map[string]interface{}{
		"engine":                  spec["engine"],
		"global.core":             globalSpec["core"],
		"runai-container-toolkit": spec["runai-container-toolkit"],
		"runai-node-exporter":     spec["runai-node-exporter"],
	}
}
