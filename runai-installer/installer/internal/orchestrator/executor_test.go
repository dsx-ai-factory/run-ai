/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/flow"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8smocks "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
	"github.com/stretchr/testify/assert"
)

// Helper functions for executor tests

func createTestRunAIDataForExecutor(stepTimeout time.Duration) *model.RunAIInstallData {
	return createTestRunAIDataForExecutorWithExecTimeout(stepTimeout, 50*time.Minute)
}

func createTestRunAIDataForExecutorWithExecTimeout(stepTimeout, executionTimeout time.Duration) *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:          true,
		StepTimeout:      stepTimeout,
		ExecutionTimeout: executionTimeout,
		GPUType:          config.GPUTypeReal,
		LogsDir:          "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	return data
}

func createMockInstallerForExecutor(t *testing.T) *Installer {
	mockK8sClient := k8smocks.NewMockK8sClients(t)
	mockK8sClient.On("GetContextName").Return("test-context").Maybe()

	return &Installer{
		k8sClient: mockK8sClient,
		runaiData: createTestRunAIDataForExecutor(7 * time.Minute),
	}
}

// Test basic executor creation

func TestNewInstallationExecutor(t *testing.T) {
	installer := createMockInstallerForExecutor(t)
	runaiData := createTestRunAIDataForExecutor(7 * time.Minute)

	executor := NewInstallationExecutor(installer, runaiData)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.ctx)
	assert.NotNil(t, executor.cancel)
	assert.NotNil(t, executor.externalEvents)
	assert.Equal(t, 2*time.Second, executor.statusPollInterval)
}

func TestExecutor_WithTimeout(t *testing.T) {
	installer := createMockInstallerForExecutor(t)
	runaiData := createTestRunAIDataForExecutorWithExecTimeout(7*time.Minute, 100*time.Millisecond) // Very short execution timeout

	executor := NewInstallationExecutor(installer, runaiData)

	// Verify context has deadline
	deadline, ok := executor.ctx.Deadline()
	assert.True(t, ok, "Context should have a deadline")
	assert.True(t, time.Until(deadline) <= 100*time.Millisecond, "Deadline should be within 100ms")
}

// Test cancellation

func TestExecutor_Cancel(t *testing.T) {
	installer := createMockInstallerForExecutor(t)
	runaiData := createTestRunAIDataForExecutor(7 * time.Minute)
	executor := NewInstallationExecutor(installer, runaiData)

	// Verify context is not done initially
	select {
	case <-executor.ctx.Done():
		t.Fatal("Context should not be done initially")
	default:
	}

	// Cancel the executor
	executor.Cancel()

	// Verify context is now done
	select {
	case <-executor.ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be done after cancel")
	}
}

// Test timeout functionality

func TestExecutor_Timeout(t *testing.T) {
	installer := createMockInstallerForExecutor(t)
	runaiData := createTestRunAIDataForExecutorWithExecTimeout(7*time.Minute, 50*time.Millisecond) // Very short execution timeout
	executor := NewInstallationExecutor(installer, runaiData)

	// Wait for timeout
	select {
	case <-executor.ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, executor.ctx.Err())
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timeout should have occurred")
	}
}

func TestNewFlowGroup(t *testing.T) {
	flows := []*flow.InstallFlow{}
	externalEvents := make(chan *flow.FlowStatusReportingMsg, 100)

	flowGroup := NewFlowGroup(flows, externalEvents)

	assert.NotNil(t, flowGroup)
	assert.Equal(t, flows, flowGroup.Flows)
	assert.NotNil(t, flowGroup.Manager)
}

func TestExecutor_StopsOnFlowGroupFailure(t *testing.T) {
	installer := createMockInstallerForExecutor(t)
	runaiData := createTestRunAIDataForExecutor(7 * time.Minute)

	originalBuilder := BuildFlowGroups
	defer func() { BuildFlowGroups = originalBuilder }()

	var capturedExternalEvents chan *flow.FlowStatusReportingMsg
	BuildFlowGroups = func(deps InstallDeps, runaiData *model.RunAIInstallData, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup {
		capturedExternalEvents = externalEvents
		flowManager1 := flow.NewFlowManagerWithExternalEvents([]*flow.InstallFlow{}, externalEvents, false)
		flowManager2 := flow.NewFlowManagerWithExternalEvents([]*flow.InstallFlow{}, externalEvents, false)
		return []*FlowGroup{
			{Flows: []*flow.InstallFlow{}, Manager: flowManager1},
			{Flows: []*flow.InstallFlow{}, Manager: flowManager2},
		}
	}

	executor := NewInstallationExecutor(installer, runaiData)
	executor.Start()
	executor.Cancel()

	assert.NotNil(t, capturedExternalEvents)
	assert.NotNil(t, executor.flowGroups[0].Manager)
	assert.NotNil(t, executor.flowGroups[1].Manager)
	assert.False(t, executor.flowGroups[0].Manager.HasFailed(), "Empty flow group should not have failed")
	assert.False(t, executor.flowGroups[1].Manager.HasFailed(), "Empty flow group should not have failed")
}

func TestExecutor_EmptyPrerequisites_SkipsPrereqFlows(t *testing.T) {
	installer := createMockInstallerForExecutor(t)
	opts := &config.InstallConfig{
		DevMode:     true,
		StepTimeout: 7 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
	}
	runaiData, _ := model.NewRunAIInstallData(opts)
	runaiData.Prerequisites = []string{}

	executor := NewInstallationExecutor(installer, runaiData)
	executor.Start()
	executor.Cancel()

	assert.Equal(t, 1, len(executor.flowGroups), "Should have only 1 flow group when Prerequisites is empty")
}

func TestExecutor_WithPrerequisites_IncludesPrereqFlows(t *testing.T) {
	installer := createMockInstallerForExecutor(t)
	runaiData := createTestRunAIDataForExecutor(7 * time.Minute)
	runaiData.Prerequisites = []string{"ingress-nginx"}

	originalBuilder := BuildFlowGroups
	defer func() { BuildFlowGroups = originalBuilder }()

	BuildFlowGroups = func(deps InstallDeps, data *model.RunAIInstallData, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup {
		prereqFlowGroup := &FlowGroup{
			Flows:   []*flow.InstallFlow{},
			Manager: flow.NewFlowManagerWithExternalEvents([]*flow.InstallFlow{}, externalEvents, false),
		}
		runaiFlowGroup := &FlowGroup{
			Flows:   []*flow.InstallFlow{},
			Manager: flow.NewFlowManagerWithExternalEvents([]*flow.InstallFlow{}, externalEvents, false),
		}
		return []*FlowGroup{prereqFlowGroup, runaiFlowGroup}
	}

	executor := NewInstallationExecutor(installer, runaiData)
	executor.Start()
	executor.Cancel()

	assert.Equal(t, 2, len(executor.flowGroups), "Should have 2 flow groups when Prerequisites is not empty")
}

func TestExecutor_ContextTimeout_BeforeFlowGroup_SendsTimeoutEvent(t *testing.T) {
	installer := createMockInstallerForExecutor(t)
	runaiData := createTestRunAIDataForExecutorWithExecTimeout(7*time.Minute, 1*time.Millisecond)

	originalBuilder := BuildFlowGroups
	defer func() { BuildFlowGroups = originalBuilder }()

	BuildFlowGroups = func(deps InstallDeps, data *model.RunAIInstallData, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup {
		time.Sleep(50 * time.Millisecond)

		fl := flow.NewInstallFlow("TestFlow", []step.InstallStep{}, data)
		flowManager := flow.NewFlowManagerWithExternalEvents([]*flow.InstallFlow{fl}, externalEvents, false)
		return []*FlowGroup{
			{Flows: []*flow.InstallFlow{fl}, Manager: flowManager},
		}
	}

	executor := NewInstallationExecutor(installer, runaiData)
	eventsChan := executor.Start()

	var receivedTimeout bool
	for event := range eventsChan {
		if event.FlowStatus == model.Timeout {
			receivedTimeout = true
			assert.Equal(t, "TestFlow", event.FlowName)
			break
		}
	}

	executor.Wait()

	assert.True(t, receivedTimeout, "Should have received a timeout event")
}
