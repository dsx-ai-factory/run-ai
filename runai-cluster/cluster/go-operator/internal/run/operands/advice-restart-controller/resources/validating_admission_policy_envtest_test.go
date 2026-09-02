/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"
	"time"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// TestWorkloadRunnerWorkflowIDPolicyAdmission tests the WLR workflow-id policy
// the label lives inside the embedded spec.template
func TestWorkloadRunnerWorkflowIDPolicyAdmission(t *testing.T) {
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
	k8sClient, err := client.New(config, client.Options{Scheme: scheme})
	require.NoError(t, err)

	ctx := context.Background()
	const workloadNamespace = "workloads"
	// the policy binds only to runai/queue-labeled namespaces
	require.NoError(t, k8sClient.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   workloadNamespace,
			Labels: map[string]string{"runai/queue": "test"},
		},
	}))

	definition := ValidatingAdmissionPolicyDefinition()
	require.NoError(t, k8sClient.Create(ctx, definition.Policy))
	require.NoError(t, k8sClient.Create(ctx, definition.Binding))

	setTemplate := func(runner *dromosv1alpha1.WorkloadRunner, templateLabels map[string]string) {
		runner.Spec.Template = k8sruntime.RawExtension{
			Object: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      runner.Name + "-pod",
					Namespace: workloadNamespace,
					Labels:    templateLabels,
				},
			},
		}
	}
	newRunner := func(name string, templateLabels map[string]string) *dromosv1alpha1.WorkloadRunner {
		runner := &dromosv1alpha1.WorkloadRunner{
			TypeMeta: metav1.TypeMeta{
				APIVersion: schema.GroupVersion{Group: "run.ai", Version: "v1alpha1"}.String(),
				Kind:       "WorkloadRunner",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: workloadNamespace,
				Labels:    map[string]string{"run.ai/workload-name": name},
			},
		}
		setTemplate(runner, templateLabels)
		return runner
	}
	getRunner := func(name string) *dromosv1alpha1.WorkloadRunner {
		runner := &dromosv1alpha1.WorkloadRunner{}
		require.NoError(t, k8sClient.Get(ctx,
			client.ObjectKey{Name: name, Namespace: workloadNamespace}, runner))
		return runner
	}

	labeled := newRunner("labeled", map[string]string{heimdallWorkflowIDLabel: "wf-1"})
	require.NoError(t, k8sClient.Create(ctx, labeled))

	var deniedError error
	labeledKey := client.ObjectKey{Name: "labeled", Namespace: workloadNamespace}
	require.Eventually(t, func() bool {
		runner := &dromosv1alpha1.WorkloadRunner{}
		if !assert.NoError(t, k8sClient.Get(ctx, labeledKey, runner)) {
			return false
		}
		setTemplate(runner, nil)
		deniedError = k8sClient.Update(ctx, runner)
		if deniedError == nil {
			restamp := &dromosv1alpha1.WorkloadRunner{}
			if assert.NoError(t, k8sClient.Get(ctx, labeledKey, restamp)) {
				setTemplate(restamp, map[string]string{heimdallWorkflowIDLabel: "wf-1"})
				assert.NoError(t, k8sClient.Update(ctx, restamp))
			}
			return false
		}
		return true
	}, 10*time.Second, 100*time.Millisecond)
	assert.Contains(t, deniedError.Error(), "cannot be removed once set")

	runner := getRunner("labeled")
	setTemplate(runner, map[string]string{heimdallWorkflowIDLabel: "wf-2"})
	err = k8sClient.Update(ctx, runner)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be modified once set")

	// unrelated template change with the label intact passes
	runner = getRunner("labeled")
	setTemplate(runner, map[string]string{heimdallWorkflowIDLabel: "wf-1", "app": "training"})
	require.NoError(t, k8sClient.Update(ctx, runner))

	// the WLR's own metadata labels are not this policy's business
	runner = getRunner("labeled")
	runner.Labels = map[string]string{"run.ai/workload-name": "renamed"}
	require.NoError(t, k8sClient.Update(ctx, runner))

	// first stamp on a previously unlabeled template is always allowed
	unlabeled := newRunner("unlabeled", nil)
	require.NoError(t, k8sClient.Create(ctx, unlabeled))
	runner = getRunner("unlabeled")
	setTemplate(runner, map[string]string{heimdallWorkflowIDLabel: "wf-3"})
	require.NoError(t, k8sClient.Update(ctx, runner))
}
