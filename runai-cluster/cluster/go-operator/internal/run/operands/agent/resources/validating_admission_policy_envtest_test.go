/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestAssetIdentityAdmissionPolicies(t *testing.T) {
	binaryAssets := os.Getenv("KUBEBUILDER_ASSETS")
	if binaryAssets == "" {
		t.Skip("KUBEBUILDER_ASSETS is required for API-server-backed admission testing")
	}

	testEnvironment := &envtest.Environment{BinaryAssetsDirectory: binaryAssets}
	testEnvironment.ControlPlane.GetAPIServer().Configure().Append(
		"enable-admission-plugins", "ValidatingAdmissionPolicy")
	config, err := testEnvironment.Start()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testEnvironment.Stop()) })

	scheme := k8sruntime.NewScheme()
	require.NoError(t, admissionv1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, rbacv1.AddToScheme(scheme))
	adminClient, err := client.New(config, client.Options{Scheme: scheme})
	require.NoError(t, err)

	ctx := context.Background()
	const (
		installationNamespace   = "runai-system"
		workloadNamespace       = "workloads"
		unrelatedServiceAccount = "unrelated"
	)
	writerServiceAccounts := []string{
		"runai-agent",
		"shared-objects-controller",
		"anyworkload-controller",
	}
	require.NoError(t, adminClient.Create(ctx,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: installationNamespace}}))
	require.NoError(t, adminClient.Create(ctx,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: workloadNamespace}}))
	for _, serviceAccountName := range writerServiceAccounts {
		require.NoError(t, adminClient.Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: serviceAccountName, Namespace: installationNamespace},
		}))
	}
	require.NoError(t, adminClient.Create(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: unrelatedServiceAccount, Namespace: workloadNamespace},
	}))
	createAssetWriterRoleBinding(t, ctx, adminClient, workloadNamespace, installationNamespace,
		writerServiceAccounts, unrelatedServiceAccount)

	runaiConfig := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: installationNamespace}}
	runaiConfig.SetDefaultsWhereNeeded()
	definitions := AssetIdentityAdmissionPolicyDefinitions(runaiConfig)
	for _, definition := range definitions {
		require.NoError(t, adminClient.Create(ctx, definition.Policy))
		require.NoError(t, adminClient.Create(ctx, definition.Binding))
		requirePolicyTypeCheckingClean(t, ctx, adminClient, definition.Policy.Name)
	}

	unrelatedClient := impersonatedAssetClient(t, config, scheme,
		workloadNamespace, unrelatedServiceAccount)
	auditSecret := assetSecret("audit-secret", workloadNamespace)
	auditPVC := assetPVC("audit-pvc", workloadNamespace)
	require.NoError(t, unrelatedClient.Create(ctx, auditSecret))
	require.NoError(t, unrelatedClient.Create(ctx, auditPVC))

	for _, definition := range definitions {
		binding := &admissionv1.ValidatingAdmissionPolicyBinding{}
		require.NoError(t, adminClient.Get(ctx, client.ObjectKey{Name: definition.Binding.Name}, binding))
		binding.Spec.ValidationActions = []admissionv1.ValidationAction{admissionv1.Deny}
		require.NoError(t, adminClient.Update(ctx, binding))
	}

	requireEventuallyDenied(t, func(attempt int) error {
		return unrelatedClient.Create(ctx, assetSecret(fmt.Sprintf("denied-secret-%d", attempt), workloadNamespace))
	}, "only Run:ai-managed service accounts may modify asset Secrets")
	requireEventuallyDenied(t, func(attempt int) error {
		return unrelatedClient.Create(ctx, assetPVC(fmt.Sprintf("denied-pvc-%d", attempt), workloadNamespace))
	}, "only Run:ai-managed service accounts may modify asset PersistentVolumeClaims")

	require.Error(t, unrelatedClient.Delete(ctx, auditSecret))
	auditPVC.Annotations = map[string]string{"updated": "true"}
	require.Error(t, unrelatedClient.Update(ctx, auditPVC))

	require.NoError(t, unrelatedClient.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "unlabeled-secret", Namespace: workloadNamespace},
	}))
	require.NoError(t, unrelatedClient.Create(ctx, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "unlabeled-pvc", Namespace: workloadNamespace},
		Spec:       validPVCSpec(),
	}))

	for _, serviceAccountName := range writerServiceAccounts {
		writerClient := impersonatedAssetClient(t, config, scheme,
			installationNamespace, serviceAccountName)
		require.NoError(t, writerClient.Create(ctx,
			assetSecret("allowed-secret-"+serviceAccountName, workloadNamespace)))
	}
	for _, serviceAccountName := range []string{"runai-agent", "shared-objects-controller"} {
		writerClient := impersonatedAssetClient(t, config, scheme,
			installationNamespace, serviceAccountName)
		require.NoError(t, writerClient.Create(ctx,
			assetPVC("allowed-pvc-"+serviceAccountName, workloadNamespace)))
	}
}

func createAssetWriterRoleBinding(
	t *testing.T,
	ctx context.Context,
	k8sClient client.Client,
	workloadNamespace string,
	installationNamespace string,
	writerServiceAccounts []string,
	unrelatedServiceAccount string,
) {
	t.Helper()
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "asset-writer", Namespace: workloadNamespace},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "persistentvolumeclaims"},
				Verbs:     []string{"create", "get", "update", "delete"},
			},
		},
	}
	require.NoError(t, k8sClient.Create(ctx, role))
	subjects := make([]rbacv1.Subject, 0, len(writerServiceAccounts)+1)
	for _, serviceAccountName := range writerServiceAccounts {
		subjects = append(subjects, rbacv1.Subject{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: installationNamespace,
		})
	}
	subjects = append(subjects, rbacv1.Subject{
		Kind:      rbacv1.ServiceAccountKind,
		Name:      unrelatedServiceAccount,
		Namespace: workloadNamespace,
	})
	require.NoError(t, k8sClient.Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "asset-writers", Namespace: workloadNamespace},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     role.Name,
		},
		Subjects: subjects,
	}))
}

func requirePolicyTypeCheckingClean(
	t *testing.T,
	ctx context.Context,
	k8sClient client.Client,
	policyName string,
) {
	t.Helper()
	policy := &admissionv1.ValidatingAdmissionPolicy{}
	require.NoError(t, k8sClient.Get(ctx, client.ObjectKey{Name: policyName}, policy))
	if policy.Status.TypeChecking != nil {
		assert.Empty(t, policy.Status.TypeChecking.ExpressionWarnings)
	}
}

func requireEventuallyDenied(t *testing.T, request func(attempt int) error, message string) {
	t.Helper()
	var deniedError error
	attempt := 0
	require.Eventually(t, func() bool {
		attempt++
		deniedError = request(attempt)
		return deniedError != nil
	}, 10*time.Second, 100*time.Millisecond)
	assert.Contains(t, deniedError.Error(), message)
}

func impersonatedAssetClient(
	t *testing.T,
	config *rest.Config,
	scheme *k8sruntime.Scheme,
	namespace string,
	serviceAccount string,
) client.Client {
	t.Helper()
	impersonatedConfig := rest.CopyConfig(config)
	impersonatedConfig.Impersonate = rest.ImpersonationConfig{
		UserName: serviceAccountUsername(namespace, serviceAccount),
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

func assetSecret(name, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{constants.RunaiAssetLabel: "secret"},
		},
	}
}

func assetPVC(name, namespace string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{constants.RunaiAssetLabel: "pvc"},
		},
		Spec: validPVCSpec(),
	}
}

func validPVCSpec() corev1.PersistentVolumeClaimSpec {
	return corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Mi")},
		},
	}
}
