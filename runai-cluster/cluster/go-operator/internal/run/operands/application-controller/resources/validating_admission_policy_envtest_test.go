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

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestApplicationIdentityPolicyAdmission(t *testing.T) {
	binaryAssets := os.Getenv("KUBEBUILDER_ASSETS")
	if binaryAssets == "" {
		t.Skip("KUBEBUILDER_ASSETS is required for API-server-backed admission testing")
	}

	_, filename, _, ok := goruntime.Caller(0)
	require.True(t, ok)
	applicationCRD := filepath.Clean(filepath.Join(filepath.Dir(filename),
		"../../../../../../runai-chart/charts/crds/crds/run_ai_application.yaml"))
	testEnvironment := &envtest.Environment{
		BinaryAssetsDirectory: binaryAssets,
		CRDDirectoryPaths:     []string{applicationCRD},
		ErrorIfCRDPathMissing: true,
	}
	testEnvironment.ControlPlane.GetAPIServer().Configure().Append(
		"enable-admission-plugins", "ValidatingAdmissionPolicy")
	config, err := testEnvironment.Start()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testEnvironment.Stop()) })

	scheme := k8sruntime.NewScheme()
	require.NoError(t, runv1alpha1.AddToScheme(scheme))
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
	runaiConfig := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: installationNamespace}}
	definition := ValidatingAdmissionPolicyDefinition(runaiConfig)
	require.NoError(t, adminClient.Create(ctx, definition.Policy))
	require.NoError(t, adminClient.Create(ctx, definition.Binding))
	require.NoError(t, adminClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: workloadNamespace}}))
	createApplicationRoleBindings(t, ctx, adminClient, workloadNamespace, installationNamespace)

	unrelatedClient := impersonatedApplicationClient(t, config, scheme, workloadNamespace, "unrelated")
	require.NoError(t, unrelatedClient.Create(ctx, application("audit-allows-unrelated", workloadNamespace)))

	binding := &admissionv1.ValidatingAdmissionPolicyBinding{}
	require.NoError(t, adminClient.Get(ctx, client.ObjectKey{Name: definition.Binding.Name}, binding))
	binding.Spec.ValidationActions = []admissionv1.ValidationAction{admissionv1.Deny}
	require.NoError(t, adminClient.Update(ctx, binding))

	var deniedError error
	deniedAttempts := 0
	require.Eventually(t, func() bool {
		deniedAttempts++
		deniedError = unrelatedClient.Create(ctx,
			application(fmt.Sprintf("denied-%d", deniedAttempts), workloadNamespace))
		return deniedError != nil
	}, 10*time.Second, 100*time.Millisecond)
	assert.Contains(t, deniedError.Error(), "only application-controller may manage Applications")

	applicationControllerClient := impersonatedApplicationClient(t, config, scheme,
		installationNamespace, "application-controller")
	require.NoError(t, applicationControllerClient.Create(ctx, application("allowed", workloadNamespace)))
}

func createApplicationRoleBindings(t *testing.T, ctx context.Context, k8sClient client.Client,
	workloadNamespace, installationNamespace string) {
	t.Helper()
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "application-writer", Namespace: workloadNamespace},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"run.ai"},
				Resources: []string{"applications"},
				Verbs:     []string{"create"},
			},
		},
	}
	require.NoError(t, k8sClient.Create(ctx, role))
	require.NoError(t, k8sClient.Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "application-writers", Namespace: workloadNamespace},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     role.Name,
		},
		Subjects: []rbacv1.Subject{
			{Kind: rbacv1.ServiceAccountKind, Name: "application-controller", Namespace: installationNamespace},
			{Kind: rbacv1.ServiceAccountKind, Name: "unrelated", Namespace: workloadNamespace},
		},
	}))
}

func impersonatedApplicationClient(t *testing.T, config *rest.Config, scheme *k8sruntime.Scheme,
	namespace, serviceAccount string) client.Client {
	t.Helper()
	impersonatedConfig := rest.CopyConfig(config)
	impersonatedConfig.Impersonate = rest.ImpersonationConfig{
		UserName: "system:serviceaccount:" + namespace + ":" + serviceAccount,
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

func application(name, namespace string) *runv1alpha1.Application {
	return &runv1alpha1.Application{
		TypeMeta:   metav1.TypeMeta{APIVersion: schema.GroupVersion{Group: "run.ai", Version: "v1alpha1"}.String(), Kind: "Application"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
}
