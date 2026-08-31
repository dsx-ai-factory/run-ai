//go:build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package uninstall_integration_test

// Integration tests for the uninstall K8sWrapper against a real cluster.
// Requires a live cluster pointed to by KUBECONFIG (a local kind cluster is fine).
//
// Run directly:
//
//	KUBECONFIG=<path> go test -v -tags integration -timeout 2m \
//	    ./internal/installations/uninstall/deps_integration_tests/
//
// Or via the Makefile target (creates and tears down a kind cluster automatically):
//
//	make test-uninstall-integration

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	testGroup     = "run.ai"
	testVersion   = "v1alpha1"
	testFinalizer = "test.run.ai/hold"

	clusterCRDName  = "testwidgets.run.ai"
	clusterResource = "testwidgets"
	clusterCRName   = "test-widget-with-finalizer"

	namespacedCRDName  = "testgadgets.run.ai"
	namespacedResource = "testgadgets"
)

var testNamespaces = []string{"test-strip-ns-alpha", "test-strip-ns-beta"}

func TestStripRunAICRFinalizers_Integration(t *testing.T) {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
	require.NoError(t, err, "load kubeconfig")

	apiExtClient, err := apiextensionsclient.NewForConfig(cfg)
	require.NoError(t, err)
	dynClient, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err)
	k8sClient, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	// --- setup: create both CRDs ---
	createCRD(t, apiExtClient, clusterCRDName, clusterResource, "TestWidget", apiextensionsv1.ClusterScoped)
	createCRD(t, apiExtClient, namespacedCRDName, namespacedResource, "TestGadget", apiextensionsv1.NamespaceScoped)

	clusterGVR := schema.GroupVersionResource{Group: testGroup, Version: testVersion, Resource: clusterResource}
	namespacedGVR := schema.GroupVersionResource{Group: testGroup, Version: testVersion, Resource: namespacedResource}

	require.NoError(t, waitForCRDEstablished(t, apiExtClient, clusterCRDName))
	require.NoError(t, waitForCRDEstablished(t, apiExtClient, namespacedCRDName))

	// --- setup: create test namespaces ---
	for _, ns := range testNamespaces {
		createNamespace(t, k8sClient, ns)
	}

	// --- setup: create CRs with finalizers and put them in Terminating ---
	// Cluster-scoped: one CR
	createAndDeleteCR(t, dynClient, clusterGVR, "", clusterCRName, "TestWidget")

	// Namespace-scoped: one CR per namespace
	for _, ns := range testNamespaces {
		createAndDeleteCR(t, dynClient, namespacedGVR, ns, "test-gadget-in-"+ns, "TestGadget")
	}

	// --- verify all CRs are stuck in Terminating ---
	assertStuck(t, dynClient, clusterGVR, "", clusterCRName)
	for _, ns := range testNamespaces {
		assertStuck(t, dynClient, namespacedGVR, ns, "test-gadget-in-"+ns)
	}

	// --- call the production implementation ---
	deps, err := uninstall.ProductionDeps()
	require.NoError(t, err)

	require.NoError(t, deps.K8sWrapper.StripRunAICRFinalizers(context.Background()))

	// --- verify all CRs are now gone ---
	t.Run("cluster-scoped CR deleted", func(t *testing.T) {
		require.NoError(t, waitForCRDeleted(t, dynClient, clusterGVR, "", clusterCRName))
		t.Logf("cluster-scoped CR %s deleted", clusterCRName)
	})
	for _, ns := range testNamespaces {
		ns := ns
		t.Run("namespaced CR deleted in "+ns, func(t *testing.T) {
			require.NoError(t, waitForCRDeleted(t, dynClient, namespacedGVR, ns, "test-gadget-in-"+ns))
			t.Logf("namespaced CR test-gadget-in-%s deleted", ns)
		})
	}
}

// createCRD registers a throwaway CRD and removes it in t.Cleanup.
func createCRD(t *testing.T, client apiextensionsclient.Interface, name, plural, kind string, scope apiextensionsv1.ResourceScope) {
	t.Helper()
	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: testGroup,
			Scope: scope,
			Names: apiextensionsv1.CustomResourceDefinitionNames{Plural: plural, Kind: kind},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
				Name:    testVersion,
				Served:  true,
				Storage: true,
				Schema: &apiextensionsv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{Type: "object"},
				},
			}},
		},
	}
	_, err := client.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), crd, metav1.CreateOptions{})
	require.NoError(t, err, "create CRD %s", name)
	t.Cleanup(func() {
		_ = client.ApiextensionsV1().CustomResourceDefinitions().Delete(context.Background(), name, metav1.DeleteOptions{})
	})
}

// createNamespace creates a namespace and removes it in t.Cleanup.
func createNamespace(t *testing.T, client kubernetes.Interface, name string) {
	t.Helper()
	_, err := client.CoreV1().Namespaces().Create(context.Background(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}},
		metav1.CreateOptions{},
	)
	require.NoError(t, err, "create namespace %s", name)
	t.Cleanup(func() {
		_ = client.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
	})
}

// createAndDeleteCR creates a CR with a finalizer then immediately issues a
// delete so that it enters Terminating state.
// namespace is empty for cluster-scoped CRDs.
func createAndDeleteCR(t *testing.T, client dynamic.Interface, gvr schema.GroupVersionResource, namespace, name, kind string) {
	t.Helper()
	cr := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": testGroup + "/" + testVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name":       name,
			"finalizers": []interface{}{testFinalizer},
		},
	}}
	iface := dynIface(client, gvr, namespace)
	_, err := iface.Create(context.Background(), cr, metav1.CreateOptions{})
	require.NoError(t, err, "create CR %s/%s", namespace, name)

	err = iface.Delete(context.Background(), name, metav1.DeleteOptions{})
	require.NoError(t, err, "delete CR %s/%s", namespace, name)
}

// assertStuck verifies a CR is still present with a deletionTimestamp set.
func assertStuck(t *testing.T, client dynamic.Interface, gvr schema.GroupVersionResource, namespace, name string) {
	t.Helper()
	got, err := dynIface(client, gvr, namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err, "CR %s/%s should still exist", namespace, name)
	require.NotNil(t, got.GetDeletionTimestamp(), "CR %s/%s should be in Terminating", namespace, name)
	t.Logf("CR %s/%s is stuck in Terminating as expected", namespace, name)
}

func waitForCRDEstablished(t *testing.T, client apiextensionsclient.Interface, name string) error {
	t.Helper()
	return wait.PollUntilContextTimeout(context.Background(), time.Second, 30*time.Second, true,
		func(ctx context.Context) (bool, error) {
			crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return false, nil
			}
			for _, cond := range crd.Status.Conditions {
				if cond.Type == apiextensionsv1.Established && cond.Status == apiextensionsv1.ConditionTrue {
					return true, nil
				}
			}
			return false, nil
		},
	)
}

func waitForCRDeleted(t *testing.T, client dynamic.Interface, gvr schema.GroupVersionResource, namespace, name string) error {
	t.Helper()
	return wait.PollUntilContextTimeout(context.Background(), time.Second, 15*time.Second, true,
		func(ctx context.Context) (bool, error) {
			_, err := dynIface(client, gvr, namespace).Get(ctx, name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, nil
		},
	)
}

// dynIface returns the right dynamic resource interface depending on whether
// namespace is empty (cluster-scoped) or set (namespace-scoped).
func dynIface(client dynamic.Interface, gvr schema.GroupVersionResource, namespace string) dynamic.ResourceInterface {
	if namespace == "" {
		return client.Resource(gvr)
	}
	return client.Resource(gvr).Namespace(namespace)
}
