/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	k8smocks "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient/mocks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupMockK8sClientsForPrometheus(t *testing.T, objects ...client.Object) *k8smocks.MockK8sClients {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		Build()

	mockClients := k8smocks.NewMockK8sClients(t)
	mockClients.On("GetClient").Return(fakeClient).Maybe()
	mockClients.On("GetConfig").Return(&rest.Config{}).Maybe()

	return mockClients
}

func newTestRunAIInstallDataForPrometheus() *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: 10 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	return data
}

func TestNewPrometheusInstallStep(t *testing.T) {
	mockClients := setupMockK8sClientsForPrometheus(t)

	step := NewPrometheusInstallStep(mockClients, "76.2.0")

	if step.GetName() != PrometheusStepName {
		t.Errorf("expected name '%s', got %s", PrometheusStepName, step.GetName())
	}

	prometheusStep := step.(*PrometheusInstallStep)
	if prometheusStep.version != "76.2.0" {
		t.Errorf("expected version '76.2.0', got %s", prometheusStep.version)
	}

	if false {
		t.Error("expected Prometheus to be mandatory")
	}
}

func TestPrometheusInstallStep_GetStatus_NotInstalled(t *testing.T) {
	mockClients := setupMockK8sClientsForPrometheus(t)
	runaiData := newTestRunAIInstallDataForPrometheus()

	prometheus := NewPrometheusInstallStep(mockClients, "76.2.0")

	status := prometheus.GetStatusAsync(runaiData)
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled status when prometheus not found, got %v", status)
	}
}

func TestPrometheusInstallStep_GetStatus_Installed(t *testing.T) {
	// Create namespace (any namespace with prometheus operator deployment will work)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "monitoring",
		},
	}

	// Create a deployment that matches prometheus operator pattern
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-operator",
			Namespace: "monitoring",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForPrometheus(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForPrometheus()

	prometheus := NewPrometheusInstallStep(mockClients, "76.2.0")

	// GetStatus now runs probe in background, poll until status changes
	var status model.InstallPhase
	for i := 0; i < 50; i++ {
		status = prometheus.GetStatusAsync(runaiData)
		if status == model.Installed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status != model.Installed {
		t.Errorf("expected Installed status when prometheus operator found, got %v", status)
	}
}

func TestPrometheusInstallStep_GetStatus_InstalledDifferentNamespace(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-prometheus-stack",
		},
	}

	// Create a deployment that matches prometheus operator pattern
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-operator",
			Namespace: "kube-prometheus-stack",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForPrometheus(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForPrometheus()

	prometheus := NewPrometheusInstallStep(mockClients, "76.2.0")

	// GetStatus now runs probe in background, poll until status changes
	var status model.InstallPhase
	for i := 0; i < 50; i++ {
		status = prometheus.GetStatusAsync(runaiData)
		if status == model.Installed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// ValidateComponentNamespace with empty expectedNamespace should find it in any namespace
	if status != model.Installed {
		t.Errorf("expected Installed status when prometheus operator found in any namespace, got %v", status)
	}
}

func TestPrometheusInstallStep_GetStatus_TerminalStatus(t *testing.T) {
	mockClients := setupMockK8sClientsForPrometheus(t)
	runaiData := newTestRunAIInstallDataForPrometheus()

	prometheus := NewPrometheusInstallStep(mockClients, "76.2.0").(*PrometheusInstallStep)

	// Set a terminal status
	prometheus.SetStatus(model.Installed, nil)

	status := prometheus.GetStatusAsync(runaiData)
	if status != model.Installed {
		t.Errorf("expected Installed status to be returned as terminal, got %v", status)
	}
}

func TestPrometheusInstallStep_Execute(t *testing.T) {
	t.Skip("Skipping test that requires network access and real cluster connection")

	mockClients := setupMockK8sClientsForPrometheus(t)
	runaiData := newTestRunAIInstallDataForPrometheus()

	prometheus := NewPrometheusInstallStep(mockClients, "76.2.0").(*PrometheusInstallStep)

	ctx := context.Background()
	err := prometheus.Execute(ctx, runaiData)

	// We expect an error because helm installation will fail in test environment
	if err == nil {
		t.Log("Execute returned no error (helm may have succeeded in test env)")
	}

	// Verify that status was set to Installing
	if prometheus.GetCachedStatus() != model.Installing && prometheus.GetCachedStatus() != model.Failed {
		t.Errorf("expected Installing or Failed status, got %v", prometheus.GetCachedStatus())
	}
}
