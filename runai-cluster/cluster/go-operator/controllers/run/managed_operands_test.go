/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"testing"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	internalv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// Helper to create a basic RunaiConfig
func createTestRunaiConfig() runaiv1.RunaiConfig {
	config := runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runaiv1.DefaultName,
			Namespace: runaiv1.DefaultNamespace,
		},
		Spec: runaiv1.RunaiConfigSpec{
			Internal: &internalv1.Internal{
				Global: &internalv1.Global{
					Version: "test-version",
				},
			},
		},
	}
	config.SetDefaultsWhereNeeded()
	return config
}

// Helper to count enabled operands
func countEnabled(operandList []operands.Operand) int {
	count := 0
	for _, op := range operandList {
		if op.Enabled() {
			count++
		}
	}
	return count
}

// Helper to find operand by name
func findOperandByName(operandList []operands.Operand, name string) operands.Operand {
	for _, op := range operandList {
		if op.Name() == name {
			return op
		}
	}
	return nil
}

func TestNewManagedOperands_BasicConfiguration(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	// Verify interface implementation
	require.NotNil(t, mo)
	require.Implements(t, (*ManagedOperands)(nil), mo)

	// Verify that we have operands in both clusters
	localOps := mo.GetLocalClusterOperands()
	remoteOps := mo.GetRemoteClusterOperands()

	assert.NotEmpty(t, localOps, "Local cluster should have operands")
	assert.NotEmpty(t, remoteOps, "Remote cluster should have operands")
}

func TestNewManagedOperandsIncludesCleanupOnlyOnLocalCluster(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)
	cleanup := findOperandByName(mo.GetLocalClusterOperands(), "legacy-workload-cleanup")
	require.NotNil(t, cleanup)
	require.Empty(t, cleanup.Resources())
	require.Nil(t, findOperandByName(mo.GetRemoteClusterOperands(), "legacy-workload-cleanup"))
}

func TestNewManagedOperandsDeletesResearcherServiceOnRemoteCluster(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)
	cleanup := findOperandByName(mo.GetRemoteClusterOperands(), "researcher-service")
	require.NotNil(t, cleanup)
	require.False(t, cleanup.Enabled())
	require.NotEmpty(t, cleanup.Resources())
	require.Nil(t, findOperandByName(mo.GetLocalClusterOperands(), "researcher-service"))

	resources, err := ResourcesForOperands(ctx, []operands.Operand{cleanup})
	require.NoError(t, err)
	assert.Empty(t, resources.CreateOrUpdate)
	assert.Len(t, resources.Delete, len(cleanup.Resources()))
}

func TestManagedOperands_EnabledVsAllOperands(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	// Disable some operands
	falseVal := false
	runaiConfig.Spec.Agent.Enabled = &falseVal
	runaiConfig.Spec.ClusterSync.Enabled = &falseVal

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	// Check local cluster
	localAll := mo.GetLocalClusterOperands()
	localEnabled := mo.GetLocalClusterEnabled()

	assert.GreaterOrEqual(t, len(localAll), len(localEnabled),
		"All operands should be >= enabled operands")

	enabledCount := countEnabled(localAll)
	assert.Equal(t, len(localEnabled), enabledCount,
		"GetLocalClusterEnabled should return only enabled operands")

	// Check remote cluster
	remoteAll := mo.GetRemoteClusterOperands()
	remoteEnabled := mo.GetRemoteClusterEnabled()

	assert.GreaterOrEqual(t, len(remoteAll), len(remoteEnabled),
		"All operands should be >= enabled operands")

	enabledCountRemote := countEnabled(remoteAll)
	assert.Equal(t, len(remoteEnabled), enabledCountRemote,
		"GetRemoteClusterEnabled should return only enabled operands")
}

func TestManagedOperands_AgentDisabled(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	falseVal := false
	runaiConfig.Spec.Agent.Enabled = &falseVal

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	localEnabled := mo.GetLocalClusterEnabled()
	agent := findOperandByName(localEnabled, "runai-agent")
	assert.Nil(t, agent, "Agent should not be in enabled list when disabled")
}

func TestManagedOperands_KaiOperatorEnabledByDefault(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)

	runaiConfig := createTestRunaiConfig()
	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)
	kaiOperator := findOperandByName(mo.GetRemoteClusterEnabled(), "kai-operator")
	assert.NotNil(t, kaiOperator, "kai-operator should be enabled by default")
}

func TestManagedOperands_KaiOperatorDisabled(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)

	runaiConfig := createTestRunaiConfig()
	runaiConfig.Spec.KAIOperator.Enabled = ptr.To(false)
	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)
	kaiOperator := findOperandByName(mo.GetRemoteClusterEnabled(), "kai-operator")
	assert.Nil(t, kaiOperator, "kai-operator should not be enabled when disabled")
}

func TestManagedOperands_HeadlessMode(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	runaiConfig.Spec.Internal.Development.Headless = true

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	remoteEnabled := mo.GetRemoteClusterEnabled()

	// ResearcherService should not be enabled in headless mode
	researcherService := findOperandByName(remoteEnabled, "researcher-service")
	assert.Nil(t, researcherService, "ResearcherService should not be enabled in headless mode")

	// ClusterApi should not be enabled in headless mode
	clusterApi := findOperandByName(remoteEnabled, "cluster-api")
	assert.Nil(t, clusterApi, "ClusterApi should not be enabled in headless mode")
}

func TestManagedOperands_AccessRuleRequiresProjectControllerRoleBindings(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	trueVal := true
	falseVal := false
	runaiConfig.Spec.AccessRuleController.Enabled = &trueVal
	runaiConfig.Spec.ProjectController.Enabled = &trueVal
	runaiConfig.Spec.ProjectController.CreateRoleBindings = &falseVal

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	remoteEnabled := mo.GetRemoteClusterEnabled()

	// AccessRule should not be enabled without CreateRoleBindings
	accessRule := findOperandByName(remoteEnabled, "accessrule-controller")
	assert.Nil(t, accessRule, "AccessRule should not be enabled without CreateRoleBindings")
}

func TestManagedOperands_NMCEnabled(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	runaiConfig.Spec.Internal.Global.NMC = internalv1.NMC{
		RemoteAPIServerURL: "https://remote-api-server:6443",
	}

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	localOps := mo.GetLocalClusterOperands()
	nmcProm := findOperandByName(localOps, "nmc-mgmt-prometheus")
	assert.NotNil(t, nmcProm, "NMC Prometheus should be present when NMC is enabled")
	assert.True(t, nmcProm.Enabled(), "NMC Prometheus should be enabled")
}

func TestManagedOperands_NMCDisabled(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	// NMC is disabled by default (no RemoteAPIServerURL)
	runaiConfig.Spec.Internal.Global.NMC = internalv1.NMC{}

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	localOps := mo.GetLocalClusterOperands()
	nmcProm := findOperandByName(localOps, "nmc-mgmt-prometheus")
	assert.Nil(t, nmcProm, "NMC Prometheus should not be present when NMC is disabled")
}

func TestManagedOperands_OperandNamesAreUnique(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	// Check local cluster operands have unique names
	localOps := mo.GetLocalClusterOperands()
	localNames := make(map[string]bool)
	for _, op := range localOps {
		name := op.Name()
		assert.False(t, localNames[name],
			"Duplicate operand name in local cluster: %s", name)
		localNames[name] = true
	}

	// Check remote cluster operands have unique names
	remoteOps := mo.GetRemoteClusterOperands()
	remoteNames := make(map[string]bool)
	for _, op := range remoteOps {
		name := op.Name()
		assert.False(t, remoteNames[name],
			"Duplicate operand name in remote cluster: %s", name)
		remoteNames[name] = true
	}
}

func TestManagedOperands_EncryptionKeyInBothClusters(t *testing.T) {
	ctx := context.Background()
	apiContext := &ContextMock{}
	fc := createEmptyKubeClientMock(apiContext)
	kbClient := kc.New(fc, fc, nil)
	runaiConfig := createTestRunaiConfig()

	mo := NewManagedOperands(ctx, kbClient, kbClient, &runaiConfig)

	// According to the comment in the code, encryption-key should be in both clusters temporarily
	localOps := mo.GetLocalClusterOperands()
	remoteOps := mo.GetRemoteClusterOperands()

	localEncryption := findOperandByName(localOps, "encryption-key")
	remoteEncryption := findOperandByName(remoteOps, "encryption-key")

	assert.NotNil(t, localEncryption, "Encryption-key should be in local cluster")
	assert.NotNil(t, remoteEncryption, "Encryption-key should be in remote cluster")

	// Both should be enabled
	assert.True(t, localEncryption.Enabled(), "Local encryption-key should be enabled")
	assert.True(t, remoteEncryption.Enabled(), "Remote encryption-key should be enabled")
}

func TestGetEnabled_EmptyList(t *testing.T) {
	result := getEnabled([]operands.Operand{})
	assert.Equal(t, 0, len(result))
}

func TestGetEnabled_AllEnabled(t *testing.T) {
	ops := []operands.Operand{
		&testMockOperand{name: "op1", enabled: true},
		&testMockOperand{name: "op2", enabled: true},
		&testMockOperand{name: "op3", enabled: true},
	}

	result := getEnabled(ops)
	assert.Equal(t, 3, len(result))

	// Verify all returned operands are enabled
	for _, op := range result {
		assert.True(t, op.Enabled(), "Operand %s should be enabled", op.Name())
	}
}

func TestGetEnabled_AllDisabled(t *testing.T) {
	ops := []operands.Operand{
		&testMockOperand{name: "op1", enabled: false},
		&testMockOperand{name: "op2", enabled: false},
		&testMockOperand{name: "op3", enabled: false},
	}

	result := getEnabled(ops)
	assert.Equal(t, 0, len(result))
}

func TestGetEnabled_MixedEnabledDisabled(t *testing.T) {
	ops := []operands.Operand{
		&testMockOperand{name: "op1", enabled: true},
		&testMockOperand{name: "op2", enabled: false},
		&testMockOperand{name: "op3", enabled: true},
		&testMockOperand{name: "op4", enabled: false},
	}

	result := getEnabled(ops)
	assert.Equal(t, 2, len(result))

	// Verify all returned operands are enabled
	for _, op := range result {
		assert.True(t, op.Enabled(), "Operand %s should be enabled", op.Name())
	}
}

// Mock operand for testing
type testMockOperand struct {
	name      string
	enabled   bool
	resources []operands.Resource
}

func (m *testMockOperand) Name() string {
	return m.name
}

func (m *testMockOperand) Resources() []operands.Resource {
	return m.resources
}

func (m *testMockOperand) Status(ctx context.Context) (operands.Status, error) {
	return operands.ReadyStatus(), nil
}

func (m *testMockOperand) Enabled() bool {
	return m.enabled
}
