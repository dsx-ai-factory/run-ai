/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	k8smocks "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient/mocks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupMockK8sClientsForIngress(t *testing.T, objects ...client.Object) *k8smocks.MockK8sClients {
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

func newTestRunAIInstallDataForIngress() *model.RunAIInstallData {
	opts := &config.InstallConfig{
		DevMode:     false,
		StepTimeout: 10 * time.Minute,
		GPUType:     config.GPUTypeReal,
		LogsDir:     "/tmp",
	}
	data, _ := model.NewRunAIInstallData(opts)
	return data
}

func TestNewNgnixIngress(t *testing.T) {
	mockClients := setupMockK8sClientsForIngress(t)

	step := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	if step.GetName() != "Ingress NGINX" {
		t.Errorf("expected name '%s', got %s", "Ingress NGINX", step.GetName())
	}
}

func TestNgnixIngress_GetStatus_NotInstalled(t *testing.T) {
	mockClients := setupMockK8sClientsForIngress(t)
	runaiData := newTestRunAIInstallDataForIngress()

	ingress := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	status := ingress.GetStatusAsync(runaiData)
	if status != model.NotInstalled {
		t.Errorf("expected NotInstalled status when ingress not found, got %v", status)
	}
}

func TestNgnixIngress_GetStatus_InstalledInNginxIngress(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-ingress",
		},
	}

	// Create a deployment that matches nginx controller pattern
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-controller",
			Namespace: "nginx-ingress",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForIngress(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForIngress()

	ingress := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	// GetStatus now runs probe in background, poll until status changes
	var status model.InstallPhase
	for i := 0; i < 50; i++ {
		status = ingress.GetStatusAsync(runaiData)
		if status == model.Installed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status != model.Installed {
		t.Errorf("expected Installed status when nginx-ingress controller found, got %v", status)
	}
}

func TestNgnixIngress_GetStatus_InstalledInIngressNginx(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-nginx",
		},
	}

	// Create a deployment that matches nginx controller pattern
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-controller",
			Namespace: "ingress-nginx",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
		},
	}

	mockClients := setupMockK8sClientsForIngress(t, namespace, deployment)
	runaiData := newTestRunAIInstallDataForIngress()

	ingress := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	// GetStatus now runs probe in background, poll until status changes
	var status model.InstallPhase
	for i := 0; i < 50; i++ {
		status = ingress.GetStatusAsync(runaiData)
		if status == model.Installed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status != model.Installed {
		t.Errorf("expected Installed status when ingress-nginx controller found, got %v", status)
	}
}

func TestNgnixIngress_GetStatus_TerminalStatus(t *testing.T) {
	mockClients := setupMockK8sClientsForIngress(t)
	runaiData := newTestRunAIInstallDataForIngress()

	ingress := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	// Set a terminal status
	ingress.SetStatus(model.Failed, nil)

	status := ingress.GetStatusAsync(runaiData)
	if status != model.Failed {
		t.Errorf("expected Failed status to be returned as terminal, got %v", status)
	}
}

func TestNgnixIngress_getValuesForEnvironment_Default(t *testing.T) {
	mockClients := setupMockK8sClientsForIngress(t)
	runaiData := newTestRunAIInstallDataForIngress()

	ingress := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	values := ingress.(*NgnixIngressInstallStep).getValuesForEnvironment(runaiData)

	if len(values) != 1 {
		t.Errorf("expected 1 value for default (on-premise), got %d", len(values))
	}

	expected := []string{"controller.service.type=NodePort"}
	for i, exp := range expected {
		if values[i] != exp {
			t.Errorf("expected value[%d] to be '%s', got '%s'", i, exp, values[i])
		}
	}
}

func TestNgnixIngress_getValuesForEnvironment_WithCustomPort(t *testing.T) {
	mockClients := setupMockK8sClientsForIngress(t)
	runaiData := newTestRunAIInstallDataForIngress()
	runaiData.Inputs[configinputs.InstallInputIngressControllerHTTPSPort] = "30443"

	ingress := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	values := ingress.(*NgnixIngressInstallStep).getValuesForEnvironment(runaiData)

	if len(values) != 2 {
		t.Errorf("expected 2 values for on-premise with custom port, got %d", len(values))
	}

	expectedValues := []string{
		"controller.service.type=NodePort",
		"controller.service.nodePorts.https=30443",
	}

	for i, expected := range expectedValues {
		if values[i] != expected {
			t.Errorf("expected value[%d] to be '%s', got '%s'", i, expected, values[i])
		}
	}
}

func TestNgnixIngress_getOnPremiseValues(t *testing.T) {
	mockClients := setupMockK8sClientsForIngress(t)
	ingress := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	tests := []struct {
		name           string
		httpsPort      string
		expectedValues []string
	}{
		{
			name:      "without ingress controller https port",
			httpsPort: "",
			expectedValues: []string{
				"controller.service.type=NodePort"},
		},
		{
			name:      "with ingress controller https port",
			httpsPort: "30443",
			expectedValues: []string{
				"controller.service.type=NodePort",
				"controller.service.nodePorts.https=30443",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := ingress.(*NgnixIngressInstallStep).getOnPremiseValues(tt.httpsPort)

			if len(values) != len(tt.expectedValues) {
				t.Errorf("expected %d values, got %d", len(tt.expectedValues), len(values))
			}

			for i, expected := range tt.expectedValues {
				if values[i] != expected {
					t.Errorf("expected value[%d] to be '%s', got '%s'", i, expected, values[i])
				}
			}
		})
	}
}

func TestNgnixIngress_PostInstall_NodePortService(t *testing.T) {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-nginx",
		},
	}

	// Create a NodePort service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress-nginx-controller",
			Namespace: "ingress-nginx",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Name:     "https",
					NodePort: 30443,
				},
			},
		},
	}

	mockClients := setupMockK8sClientsForIngress(t, namespace, service)
	runaiData := newTestRunAIInstallDataForIngress()

	ingress := NewNgnixIngressInstallStep(mockClients, "4.13.0")

	// This should retrieve the node port and store it in runaiData
	ingress.PostInstall(runaiData)

	port := runaiData.Inputs[configinputs.InstallInputIngressControllerHTTPSPort]
	if port != "30443" {
		t.Errorf("expected port to be 30443, got %s", port)
	}
}

// Helper function
func int32Ptr(i int32) *int32 {
	return &i
}
