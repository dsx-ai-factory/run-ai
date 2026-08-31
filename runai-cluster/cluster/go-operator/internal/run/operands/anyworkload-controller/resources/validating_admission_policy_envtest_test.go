/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"
	"time"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestWorkloadRunnerCreationPolicyAdmission(t *testing.T) {
	binaryAssets := os.Getenv("KUBEBUILDER_ASSETS")
	if binaryAssets == "" {
		t.Skip("KUBEBUILDER_ASSETS is required for API-server-backed admission testing")
	}

	_, filename, _, ok := goruntime.Caller(0)
	require.True(t, ok)
	workloadRunnerCRD := filepath.Clean(filepath.Join(filepath.Dir(filename),
		"../../../../../../dromos/config/crd/bases/run.ai_workloadrunners.yaml"))
	testEnvironment := &envtest.Environment{
		BinaryAssetsDirectory: binaryAssets,
		CRDDirectoryPaths:     []string{workloadRunnerCRD},
		ErrorIfCRDPathMissing: true,
	}
	testEnvironment.ControlPlane.GetAPIServer().Configure().Append(
		"enable-admission-plugins", "ValidatingAdmissionPolicy")
	config, err := testEnvironment.Start()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testEnvironment.Stop()) })

	scheme := k8sruntime.NewScheme()
	require.NoError(t, dromosv1alpha1.AddToScheme(scheme))
	require.NoError(t, admissionv1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, rbacv1.AddToScheme(scheme))
	adminClient, err := client.New(config, client.Options{Scheme: scheme})
	require.NoError(t, err)

	ctx := context.Background()
	const (
		installationNamespace = "runai-system"
		workloadNamespace     = "workloads"
	)
	require.NoError(t, adminClient.Create(ctx,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: installationNamespace}}))
	require.NoError(t, adminClient.Create(ctx,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: workloadNamespace}}))

	agentServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "runai-agent", Namespace: installationNamespace},
	}
	otherServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: workloadNamespace},
	}
	require.NoError(t, adminClient.Create(ctx, agentServiceAccount))
	require.NoError(t, adminClient.Create(ctx, otherServiceAccount))
	createWorkloadRunnerRoleBinding(t, ctx, adminClient, workloadNamespace, installationNamespace)

	runaiConfig := &runv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: installationNamespace},
	}
	runaiConfig.SetDefaultsWhereNeeded()
	definition := ValidatingAdmissionPolicyDefinition(runaiConfig)
	require.NoError(t, adminClient.Create(ctx, definition.Policy))
	require.NoError(t, adminClient.Create(ctx, definition.Binding))

	otherClient := impersonatedWorkloadRunnerClient(
		t,
		config,
		scheme,
		workloadNamespace,
		otherServiceAccount.Name,
		otherServiceAccount.UID,
	)
	var deniedError error
	deniedAttempts := 0
	require.Eventually(t, func() bool {
		deniedAttempts++
		deniedError = otherClient.Create(ctx,
			workloadRunner(fmt.Sprintf("denied-%d", deniedAttempts), workloadNamespace))
		return deniedError != nil
	}, 10*time.Second, 100*time.Millisecond)
	assert.Contains(t, deniedError.Error(), workloadRunnerCreationDeniedMessage)

	spoofedAgentClient := impersonatedWorkloadRunnerClient(
		t,
		config,
		scheme,
		installationNamespace,
		agentServiceAccount.Name,
		otherServiceAccount.UID,
	)
	err = spoofedAgentClient.Create(ctx, workloadRunner("spoofed-agent", workloadNamespace))
	require.Error(t, err)
	assert.Contains(t, err.Error(), workloadRunnerCreationDeniedMessage)

	agentClient := impersonatedWorkloadRunnerClient(
		t,
		config,
		scheme,
		installationNamespace,
		agentServiceAccount.Name,
		agentServiceAccount.UID,
	)
	require.NoError(t, agentClient.Create(ctx, workloadRunner("allowed", workloadNamespace)))
	require.NoError(t, agentClient.Create(ctx, workloadRunner("update-delete", workloadNamespace)))

	runner := &dromosv1alpha1.WorkloadRunner{}
	require.NoError(t, otherClient.Get(ctx,
		client.ObjectKey{Name: "update-delete", Namespace: workloadNamespace}, runner))
	runner.Labels = map[string]string{"updated-by": "other"}
	require.NoError(t, otherClient.Update(ctx, runner))
	require.NoError(t, otherClient.Delete(ctx, runner))
}

func createWorkloadRunnerRoleBinding(
	t *testing.T,
	ctx context.Context,
	k8sClient client.Client,
	workloadNamespace string,
	installationNamespace string,
) {
	t.Helper()

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "workloadrunner-writer", Namespace: workloadNamespace},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"run.ai"},
				Resources: []string{"workloadrunners"},
				Verbs:     []string{"create", "get", "update", "delete"},
			},
		},
	}
	require.NoError(t, k8sClient.Create(ctx, role))
	require.NoError(t, k8sClient.Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "workloadrunner-writers", Namespace: workloadNamespace},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     role.Name,
		},
		Subjects: []rbacv1.Subject{
			{Kind: rbacv1.ServiceAccountKind, Name: "runai-agent", Namespace: installationNamespace},
			{Kind: rbacv1.ServiceAccountKind, Name: "other", Namespace: workloadNamespace},
		},
	}))
}

func impersonatedWorkloadRunnerClient(
	t *testing.T,
	config *rest.Config,
	scheme *k8sruntime.Scheme,
	namespace string,
	serviceAccount string,
	uid types.UID,
) client.Client {
	t.Helper()

	impersonatedConfig := rest.CopyConfig(config)
	impersonatedConfig.Impersonate = rest.ImpersonationConfig{
		UserName: "system:serviceaccount:" + namespace + ":" + serviceAccount,
		UID:      string(uid),
		Groups: []string{
			"system:serviceaccounts",
			"system:serviceaccounts:" + namespace,
			"system:authenticated",
		},
	}
	k8sClient, err := client.New(impersonatedConfig, client.Options{Scheme: scheme})
	require.NoError(t, err)
	return k8sClient
}

func workloadRunner(name string, namespace string) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		TypeMeta: metav1.TypeMeta{
			APIVersion: schema.GroupVersion{Group: "run.ai", Version: "v1alpha1"}.String(),
			Kind:       "WorkloadRunner",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: k8sruntime.RawExtension{
				Object: &corev1.Pod{
					TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
					ObjectMeta: metav1.ObjectMeta{
						Name:      name + "-pod",
						Namespace: namespace,
					},
				},
			},
		},
	}
}
