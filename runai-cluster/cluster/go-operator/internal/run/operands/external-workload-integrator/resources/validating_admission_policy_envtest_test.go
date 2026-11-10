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

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestPodOwnershipPolicyAdmission(t *testing.T) {
	assets := os.Getenv("KUBEBUILDER_ASSETS")
	if assets == "" {
		t.Skip("KUBEBUILDER_ASSETS is required for API-server-backed admission testing")
	}
	env := &envtest.Environment{BinaryAssetsDirectory: assets}
	env.ControlPlane.GetAPIServer().Configure().Append("enable-admission-plugins", "ValidatingAdmissionPolicy")
	cfg, err := env.Start()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, env.Stop()) })
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, admissionv1.AddToScheme(scheme))
	require.NoError(t, rbacv1.AddToScheme(scheme))
	admin, err := client.New(cfg, client.Options{Scheme: scheme})
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, admin.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "workloads", Labels: map[string]string{"runai/queue": "project-a"}}}))
	// Give the impersonated callers Pod permissions so denials come from admission, not RBAC.
	require.NoError(t, admin.Create(ctx, &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-writer", Namespace: "workloads"},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{""}, Resources: []string{"pods", "pods/status"}, Verbs: []string{"get", "create", "update", "patch", "delete"}}},
	}))
	require.NoError(t, admin.Create(ctx, &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-writers", Namespace: "workloads"},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "pod-writer"},
		Subjects:   []rbacv1.Subject{{Kind: "Group", APIGroup: rbacv1.GroupName, Name: "system:authenticated"}},
	}))
	caller := func(username string) client.Client {
		c := rest.CopyConfig(cfg)
		c.Impersonate = rest.ImpersonationConfig{UserName: username, Groups: []string{"system:authenticated"}}
		result, err := client.New(c, client.Options{Scheme: scheme})
		require.NoError(t, err)
		return result
	}
	other := caller("system:serviceaccount:workloads:other")
	ewi := caller("system:serviceaccount:custom-runai:custom-nmc")
	wrongNamespace := caller("system:serviceaccount:workloads:custom-nmc")
	config := &runv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{Namespace: "custom-runai"}}
	config.SetDefaultsWhereNeeded()
	config.Spec.Internal.Global.NMC.ServiceAccountName = "custom-nmc"
	definition := PodOwnershipPolicyDefinition(config)
	require.NoError(t, admin.Create(ctx, definition.Policy))
	// Exercise enforcement explicitly; the shipped binding starts in Audit.
	definition.Binding.Spec.ValidationActions = []admissionv1.ValidationAction{admissionv1.Deny}
	require.NoError(t, admin.Create(ctx, definition.Binding))
	pod := func(name string, labels map[string]string) *corev1.Pod {
		return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "workloads", Labels: labels},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "main", Image: "test-image"}}}}
	}
	const workloadLabel = "run.ai/workload-id"
	const ownerLabel = "run.ai/top-owner-uid"
	original := map[string]string{workloadLabel: "W42", ownerLabel: "W42"}
	probe := pod("readiness", original)
	require.NoError(t, admin.Create(ctx, probe))
	require.Eventually(t, func() bool {
		err := other.Patch(ctx, probe, client.RawPatch("application/merge-patch+json", []byte(`{"metadata":{"labels":{"run.ai/workload-id":"W99"}}}`)), client.DryRunAll)
		return apierrors.IsForbidden(err)
	}, 15*time.Second, 100*time.Millisecond)
	// Ownership-looking labels outside Run:ai namespaces must not trigger the policy.
	require.NoError(t, admin.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "outside-runai"}}))
	outside := pod("outside", original)
	outside.Namespace = "outside-runai"
	require.NoError(t, admin.Create(ctx, outside))
	require.NoError(t, admin.Patch(ctx, outside, client.RawPatch("application/merge-patch+json", []byte(`{"metadata":{"labels":{"run.ai/workload-id":"W99"}}}`))))
	require.NoError(t, admin.Patch(ctx, outside, client.RawPatch("application/merge-patch+json", []byte(`{"metadata":{"labels":null}}`))))

	cases := []struct {
		name, patch string
		writer      client.Client
		denied      bool
	}{
		{"change-workload", `{"metadata":{"labels":{"run.ai/workload-id":"W99"}}}`, other, true},
		{"remove-owner", `{"metadata":{"labels":{"run.ai/top-owner-uid":null}}}`, other, true},
		{"remove-all", `{"metadata":{"labels":null}}`, other, true},
		{"empty-owner", `{"metadata":{"labels":{"run.ai/top-owner-uid":""}}}`, other, true},
		{"same-labels", `{"metadata":{"labels":{"run.ai/workload-id":"W42","run.ai/top-owner-uid":"W42"}}}`, other, false},
		{"unrelated-label", `{"metadata":{"labels":{"team":"research"}}}`, other, false},
		{"ewi-complete", `{"metadata":{"labels":{"run.ai/workload-id":"W42","run.ai/top-owner-uid":"W42","run.ai/cluster-top-owner-uid":"kubernetes-uid"}}}`, ewi, false},
		{"ewi-correct-partial", `{"metadata":{"labels":{"run.ai/workload-id":"W99","run.ai/top-owner-uid":"W99","run.ai/cluster-top-owner-uid":"kubernetes-uid"}}}`, ewi, false},
		{"wrong-namespace", `{"metadata":{"labels":{"run.ai/workload-id":"W99"}}}`, wrongNamespace, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := pod(tc.name, original)
			require.NoError(t, other.Create(ctx, p))
			err := tc.writer.Patch(ctx, p, client.RawPatch("application/merge-patch+json", []byte(tc.patch)))
			if tc.denied {
				require.True(t, apierrors.IsForbidden(err), "%v", err)
				assert.Contains(t, err.Error(), "ownership label")
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, other.Delete(ctx, p))
		})
	}
	for i, labels := range []map[string]string{nil, {workloadLabel: "", ownerLabel: ""}} {
		p := pod(fmt.Sprintf("legacy-%d", i), labels)
		require.NoError(t, other.Create(ctx, p))
		require.NoError(t, other.Patch(ctx, p, client.RawPatch("application/merge-patch+json", []byte(`{"metadata":{"labels":{"team":"legacy"}}}`))))
		require.NoError(t, ewi.Patch(ctx, p, client.RawPatch("application/merge-patch+json", []byte(`{"metadata":{"labels":{"run.ai/workload-id":"W42","run.ai/top-owner-uid":"W42"}}}`))))
		require.NoError(t, other.Delete(ctx, p))
	}
	// Status writers must still be able to report progress without touching ownership.
	statusPod := pod("status-update", original)
	require.NoError(t, other.Create(ctx, statusPod))
	statusPod.Status.Phase = corev1.PodRunning
	require.NoError(t, other.Status().Update(ctx, statusPod))
	// Full updates must enforce the same rule as patches.
	require.NoError(t, other.Get(ctx, client.ObjectKeyFromObject(statusPod), statusPod))
	statusPod.Labels[ownerLabel] = "W99"
	require.True(t, apierrors.IsForbidden(other.Update(ctx, statusPod)))

	// A partial label set must not let removal of the existing label bypass validation.
	partial := pod("partial", map[string]string{ownerLabel: "W42"})
	require.NoError(t, other.Create(ctx, partial))
	err = other.Patch(ctx, partial, client.RawPatch("application/merge-patch+json", []byte(`{"metadata":{"labels":null}}`)))
	require.True(t, apierrors.IsForbidden(err), "%v", err)
	require.NoError(t, admin.Get(ctx, client.ObjectKeyFromObject(definition.Binding), definition.Binding))
	definition.Binding.Spec.ValidationActions = []admissionv1.ValidationAction{admissionv1.Audit}
	require.NoError(t, admin.Update(ctx, definition.Binding))
	require.Eventually(t, func() bool {
		return other.Patch(ctx, partial, client.RawPatch("application/merge-patch+json", []byte(`{"metadata":{"labels":null}}`))) == nil
	}, 10*time.Second, 100*time.Millisecond)
}
