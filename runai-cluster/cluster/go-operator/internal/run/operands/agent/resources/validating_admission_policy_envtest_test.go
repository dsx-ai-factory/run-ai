/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// TestAssetIdentityAdmissionPolicies verifies the shipped Deny bindings against
// a real API server without installing any HTTP admission webhook.
func TestAssetIdentityAdmissionPolicies(t *testing.T) {
	binaryAssets := os.Getenv("KUBEBUILDER_ASSETS")
	if binaryAssets == "" {
		t.Skip("KUBEBUILDER_ASSETS is required for API-server-backed admission testing")
	}
	testEnvironment := &envtest.Environment{BinaryAssetsDirectory: binaryAssets}
	testEnvironment.ControlPlane.GetAPIServer().Configure().Append("enable-admission-plugins", "ValidatingAdmissionPolicy")
	config, err := testEnvironment.Start()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testEnvironment.Stop()) })

	scheme := k8sruntime.NewScheme()
	require.NoError(t, admissionv1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, rbacv1.AddToScheme(scheme))
	admin, err := client.New(config, client.Options{Scheme: scheme})
	require.NoError(t, err)
	ctx := context.Background()
	const namespace = "asset-policy-test"
	require.NoError(t, admin.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}))
	// Give every test identity identical RBAC so denied requests prove admission,
	// rather than missing Kubernetes permissions.
	require.NoError(t, admin.Create(ctx, &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "asset-writer", Namespace: namespace},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""}, Resources: []string{"secrets", "persistentvolumeclaims"},
			Verbs: []string{"get", "create", "update", "patch", "delete"},
		}},
	}))
	require.NoError(t, admin.Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "asset-writers", Namespace: namespace},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "asset-writer"},
		Subjects:   []rbacv1.Subject{{APIGroup: rbacv1.GroupName, Kind: rbacv1.GroupKind, Name: "asset-test-identities"}},
	}))

	runaiConfig := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "custom-runai"}}
	runaiConfig.SetDefaultsWhereNeeded()
	runaiConfig.Spec.Internal.Global.NMC.ServiceAccountName = "custom-nmc"
	runaiConfig.Spec.Internal.Global.NMC.AllowedServiceAccounts = []string{" nmc-ops, trusted-user ", "", "  "}
	for _, definition := range AssetIdentityAdmissionPolicyDefinitions(runaiConfig) {
		require.Equal(t, []admissionv1.ValidationAction{admissionv1.Deny}, definition.Binding.Spec.ValidationActions)
		require.NoError(t, admin.Create(ctx, definition.Policy))
		require.NoError(t, admin.Create(ctx, definition.Binding))
		// envtest does not run kube-controller-manager's type-checking
		// controller. The denial probes below prove actual CEL execution;
		// resource existence or an absent status is not treated as success.
	}

	identities := []struct {
		name, username string
		allowed        bool
	}{
		{"agent", "system:serviceaccount:runai:runai-agent", true},
		{"shared-controller", "system:serviceaccount:runai:shared-objects-controller", true},
		{"anyworkload-controller", "system:serviceaccount:runai:anyworkload-controller", true},
		{"unrelated-sa", "system:serviceaccount:tenant-a:unrelated", true},
		{"nmc-override", "system:serviceaccount:custom-runai:custom-nmc", true},
		{"volume-binder", "system:serviceaccount:kube-system:persistent-volume-binder", true},
		{"garbage-collector", "system:serviceaccount:kube-system:generic-garbage-collector", true},
		{"nmc-user", "external:nmc-ops:writer", true},
		{"second-pattern", "external:trusted-user:writer", true},
		{"ordinary-user", "alice", false},
		{"controller-manager", "system:kube-controller-manager", false},
		{"scheduler", "system:kube-scheduler", false},
		{"fake-sa-prefix", "prefix-system:serviceaccount:default:fake", false},
	}
	denied := assetImpersonatedClient(t, config, scheme, "alice")
	writer := assetImpersonatedClient(t, config, scheme, identities[0].username)

	for _, kind := range []string{"secret", "pvc"} {
		t.Run(kind, func(t *testing.T) {
			message := "with the run.ai/asset label"
			// Wait for admission's policy/binding cache before running assertions.
			require.Eventually(t, func() bool {
				err := denied.Create(ctx, testAsset(kind, "activation-probe", namespace), client.DryRunAll)
				return apierrors.IsForbidden(err) && strings.Contains(err.Error(), message)
			}, 15*time.Second, 100*time.Millisecond)

			for _, identity := range identities {
				t.Run(identity.name, func(t *testing.T) {
					requester := assetImpersonatedClient(t, config, scheme, identity.username)
					obj := testAsset(kind, identity.name, namespace)
					err := requester.Create(ctx, obj)
					if identity.allowed {
						require.NoError(t, err)
					} else {
						requireAssetDenial(t, err)
						require.NoError(t, writer.Create(ctx, obj))
					}
					obj.SetAnnotations(map[string]string{"updated": "true"})
					err = requester.Update(ctx, obj)
					if identity.allowed {
						require.NoError(t, err)
					} else {
						requireAssetDenial(t, err)
					}
					err = requester.Delete(ctx, obj)
					if identity.allowed {
						require.NoError(t, err)
					} else {
						requireAssetDenial(t, err)
						require.NoError(t, writer.Delete(ctx, obj))
					}
				})
			}

			t.Run("unlabeled-and-label-transitions", func(t *testing.T) {
				obj := testAsset(kind, "label-transitions", namespace)
				obj.SetLabels(nil)
				require.NoError(t, denied.Create(ctx, obj))
				obj.SetAnnotations(map[string]string{"updated": "true"})
				require.NoError(t, denied.Update(ctx, obj))
				obj.SetLabels(map[string]string{constants.RunaiAssetLabel: ""})
				requireAssetDenial(t, denied.Update(ctx, obj))
				require.NoError(t, writer.Update(ctx, obj))
				// The old labeled object must also match when the new object
				// tries to remove the label.
				obj.SetLabels(nil)
				requireAssetDenial(t, denied.Update(ctx, obj))
				require.NoError(t, writer.Update(ctx, obj))
				require.NoError(t, denied.Delete(ctx, obj))
			})

			t.Run("patch-is-validated-as-update", func(t *testing.T) {
				obj := testAsset(kind, "patch-test", namespace)
				require.NoError(t, writer.Create(ctx, obj))
				base := obj.DeepCopyObject().(client.Object)
				obj.SetAnnotations(map[string]string{"patched": "true"})
				requireAssetDenial(t, denied.Patch(ctx, obj, client.MergeFrom(base)))
				require.NoError(t, writer.Patch(ctx, obj, client.MergeFrom(base)))
				require.NoError(t, writer.Delete(ctx, obj))
			})
		})
	}

	// Reconcile back to the default configuration and verify that removing a
	// pattern revokes ordinary-user access while keeping ServiceAccounts allowed.
	runaiConfig.Spec.Internal.Global.NMC.AllowedServiceAccounts = nil
	for _, definition := range AssetIdentityAdmissionPolicyDefinitions(runaiConfig) {
		current := &admissionv1.ValidatingAdmissionPolicy{}
		require.NoError(t, admin.Get(ctx, client.ObjectKey{Name: definition.Policy.Name}, current))
		current.Spec = definition.Policy.Spec
		require.NoError(t, admin.Update(ctx, current))
	}
	for _, kind := range []string{"secret", "pvc"} {
		t.Run(kind+"-default-configuration", func(t *testing.T) {
			patternUser := assetImpersonatedClient(t, config, scheme, "external:nmc-ops:writer")
			require.Eventually(t, func() bool {
				err := patternUser.Create(ctx, testAsset(kind, "revoked-pattern", namespace), client.DryRunAll)
				return apierrors.IsForbidden(err) && strings.Contains(err.Error(), "with the run.ai/asset label")
			}, 15*time.Second, 100*time.Millisecond)
			unrelatedSA := assetImpersonatedClient(t, config, scheme, "system:serviceaccount:tenant-a:unrelated")
			obj := testAsset(kind, "default-sa", namespace)
			require.NoError(t, unrelatedSA.Create(ctx, obj))
			obj.SetAnnotations(map[string]string{"updated": "true"})
			require.NoError(t, unrelatedSA.Update(ctx, obj))
			require.NoError(t, unrelatedSA.Delete(ctx, obj))
		})
	}
}

func requireAssetDenial(t *testing.T, err error) {
	t.Helper()
	require.True(t, apierrors.IsForbidden(err), "expected admission Forbidden, got %v", err)
	require.Contains(t, err.Error(), "with the run.ai/asset label")
	require.Contains(t, err.Error(), "runai-")
}

func assetImpersonatedClient(t *testing.T, config *rest.Config, scheme *k8sruntime.Scheme, username string) client.Client {
	t.Helper()
	impersonated := rest.CopyConfig(config)
	impersonated.Impersonate = rest.ImpersonationConfig{
		UserName: username, Groups: []string{"system:authenticated", "asset-test-identities"},
	}
	c, err := client.New(impersonated, client.Options{Scheme: scheme})
	require.NoError(t, err)
	return c
}

func testAsset(kind, name, namespace string) client.Object {
	meta := metav1.ObjectMeta{Name: kind + "-" + name, Namespace: namespace, Labels: map[string]string{constants.RunaiAssetLabel: kind}}
	switch kind {
	case "secret":
		return &corev1.Secret{ObjectMeta: meta}
	case "pvc":
		return &corev1.PersistentVolumeClaim{
			ObjectMeta: meta,
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: new(string),
				AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Mi")},
				},
			},
		}
	default:
		panic(fmt.Sprintf("unsupported test asset kind %q", kind))
	}
}
