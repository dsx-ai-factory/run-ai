//go:build e2e


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// prereqNamespacePatterns, prereqHelmReleasePatterns, prereqWebhookPatterns,
// prereqCRDPatterns and prereqFinalizerHeavyCRDKinds mirror cleanup_env.py but
// intentionally exclude everything matching runai/run.ai: Run:ai itself is
// removed by Uninstall, this only wipes third-party prerequisites so a
// subsequent install exercises their detect+install paths from a clean cluster.
var (
	prereqNamespacePatterns = []string{
		"gpu-operator", "fake-gpu-operator", "knative-operator", "knative-serving",
		"knative-eventing", "kubeflow", "lws-system", "monitoring", "mpi-operator",
		"nim-operator", "nemo-operator", "network-operator", "nvidia-network-operator",
		"node-feature-discovery", "cert-manager", "kai-scheduler", "kueue-system",
		"grove", "dynamo",
	}

	prereqHelmReleasePatterns = []string{
		"fake-gpu-operator", "gpu-operator", "knative-operator", "kube-prometheus-stack",
		"prometheus-operator-crds", "lws", "nim-operator", "k8s-nim-operator",
		"nemo-operator", "network-operator", "nvidia-network-operator", "nfd",
		"node-feature-discovery", "nvsentinel", "cert-manager", "kai-scheduler",
		"kueue", "grove", "dynamo",
	}

	prereqWebhookPatterns = []string{
		"gpu-operator", "network-operator", "nvidia", "knative", "cert-manager",
		"node-feature-discovery", "kueue", "kai-scheduler", "grove", "podclique",
		"clustertopology",
	}

	prereqCRDPatterns = []string{
		".nvidia.com", ".knative.dev", ".kubeflow.org", "leaderworkerset.x-k8s.io",
		"kai.scheduler", ".kueue.x-k8s.io", "node-feature-discovery", "nodefeature",
		"nodefeaturerule", "cert-manager.io", "acme.cert-manager.io", "monitoring.coreos.com",
		".grove.io", "scheduler.grove.io",
	}

	prereqFinalizerHeavyCRDKinds = []string{
		"clusterpolicies.nvidia.com", "nvidiadrivers.nvidia.com",
		"knativeservings.operator.knative.dev", "knativeeventings.operator.knative.dev",
		"trainingruntimes.kubeflow.org", "issuers.cert-manager.io",
		"clusterissuers.cert-manager.io", "certificates.cert-manager.io",
		"certificaterequests.cert-manager.io",
	}

	prereqNonHelmManifests = []struct {
		Namespace string
		Name      string
	}{
		{"kubeflow", "training-operator"},
		{"mpi-operator", "mpi-operator"},
	}
)

const (
	prereqDeleteTimeout = 30 * time.Second
	prereqHelmTimeout   = 120 * time.Second
	prereqSettleDelay   = 20 * time.Second
)

var (
	deletePropagationBackground = metav1.DeletePropagationBackground
	prereqDeleteOptions         = metav1.DeleteOptions{PropagationPolicy: &deletePropagationBackground}
)

// Namespaces owned by the OpenShift platform or Kubernetes itself are never
// deleted or finalizer-stripped, even when they substring-match a pattern above
// (e.g. "monitoring" matches "openshift-monitoring"). Tearing them down on OCP
// removes platform components like metrics-server, which breaks API discovery
// and wedges every namespace in Terminating.
var (
	protectedNamespacePrefixes = []string{"openshift", "kube-"}
	protectedNamespaceNames    = map[string]bool{
		"default":         true,
		"openshift":       true,
		"kube-system":     true,
		"kube-public":     true,
		"kube-node-lease": true,
	}
	// ocpOwnedCRDSubstrings are CRD groups owned by OpenShift platform operators
	// (the installer reuses the cluster's built-in monitoring on OCP), so they
	// must never be deleted there.
	ocpOwnedCRDSubstrings = []string{"monitoring.coreos.com"}
)

func isProtectedNamespace(name string) bool {
	if protectedNamespaceNames[name] {
		return true
	}
	for _, p := range protectedNamespacePrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func isOpenShiftCluster(k8s kubernetes.Interface) bool {
	groups, err := k8s.Discovery().ServerGroups()
	if err != nil || groups == nil {
		return false
	}
	for _, g := range groups.Groups {
		if g.Name == "config.openshift.io" || g.Name == "security.openshift.io" {
			return true
		}
	}
	return false
}

// CleanupPrerequisites removes the third-party prerequisites the install-new
// flow can install (GPU/network operators, Prometheus, Knative, Kubeflow, LWS,
// cert-manager, etc.) from the current kube context, following the same phase
// ordering as cleanup_env.py: webhooks, finalizer-heavy CRs, Helm releases,
// non-helm manifests, namespaces, broken APIServices + finalizer stripping,
// then CRDs. Phases run sequentially because finalizers depend on operators
// being alive when their CRs are deleted. Individual deletions are best-effort
// and only logged on failure; an error is returned only when the Kubernetes
// clients cannot be built.
func CleanupPrerequisites() error {
	cfg, err := GetK8sConfig()
	if err != nil {
		return fmt.Errorf("failed to get k8s config: %w", err)
	}
	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}
	apiext, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %w", err)
	}

	ctx := context.Background()

	deletePrereqWebhooks(ctx, dyn)
	deletePrereqFinalizerHeavyCRs(ctx, dyn, apiext)
	uninstallPrereqHelmReleases()
	deletePrereqManifestInstalls(ctx, k8s)
	targetNamespaces := deletePrereqNamespaces(ctx, k8s)
	deleteBrokenAPIServices(ctx, dyn)
	time.Sleep(prereqSettleDelay)
	stripStuckPrereqFinalizers(ctx, dyn, apiext, k8s, targetNamespaces)
	deletePrereqCRDs(ctx, apiext, isOpenShiftCluster(k8s))

	return nil
}

var (
	mutatingWebhookGVR = schema.GroupVersionResource{
		Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations",
	}
	validatingWebhookGVR = schema.GroupVersionResource{
		Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations",
	}
	apiServiceGVR = schema.GroupVersionResource{
		Group: "apiregistration.k8s.io", Version: "v1", Resource: "apiservices",
	}
)

func deletePrereqWebhooks(ctx context.Context, dyn dynamic.Interface) {
	prereqLogf("deleting admission webhook configurations")
	for _, gvr := range []schema.GroupVersionResource{mutatingWebhookGVR, validatingWebhookGVR} {
		list, err := dyn.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			prereqLogf("  list %s failed: %v", gvr.Resource, err)
			continue
		}
		for _, item := range list.Items {
			name := item.GetName()
			if !matchesAnySubstring(name, prereqWebhookPatterns) {
				continue
			}
			deleteClusterScoped(ctx, dyn, gvr, name)
		}
	}
}

func deletePrereqFinalizerHeavyCRs(ctx context.Context, dyn dynamic.Interface, apiext apiextensionsclient.Interface) {
	prereqLogf("deleting CRs of finalizer-heavy CRDs (before their operators die)")
	for _, crdName := range prereqFinalizerHeavyCRDKinds {
		gvr, namespaced, ok := crdGVR(ctx, apiext, crdName)
		if !ok {
			continue
		}
		forEachCR(ctx, dyn, gvr, namespaced, func(namespace, name string) {
			deleteOne(ctx, dyn, gvr, namespace, name)
		})
	}
}

func uninstallPrereqHelmReleases() {
	prereqLogf("uninstalling Helm releases")
	settings := cli.New()
	listCfg := &action.Configuration{}
	if err := listCfg.Init(settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), helmNoopLog); err != nil {
		prereqLogf("  helm init failed: %v", err)
		return
	}
	list := action.NewList(listCfg)
	list.AllNamespaces = true
	list.All = true
	list.SetStateMask()
	releases, err := list.Run()
	if err != nil {
		prereqLogf("  helm list failed: %v", err)
		return
	}
	for _, rel := range releases {
		if !matchesAnySubstring(rel.Name, prereqHelmReleasePatterns) {
			continue
		}
		uninstallHelmRelease(rel.Name, rel.Namespace)
	}
}

func uninstallHelmRelease(name, namespace string) {
	settings := cli.New()
	settings.SetNamespace(namespace)
	cfg := &action.Configuration{}
	if err := cfg.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), helmNoopLog); err != nil {
		prereqLogf("  helm init for %s/%s failed: %v", namespace, name, err)
		return
	}
	un := action.NewUninstall(cfg)
	un.Wait = true
	un.Timeout = prereqHelmTimeout
	un.IgnoreNotFound = true
	if _, err := un.Run(name); err != nil {
		prereqLogf("  uninstall %s/%s failed: %v", namespace, name, err)
		return
	}
	prereqLogf("  uninstalled %s/%s", namespace, name)
}

func deletePrereqManifestInstalls(ctx context.Context, k8s kubernetes.Interface) {
	prereqLogf("deleting non-helm manifest installs")
	for _, m := range prereqNonHelmManifests {
		err := k8s.AppsV1().Deployments(m.Namespace).Delete(ctx, m.Name, prereqDeleteOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			prereqLogf("  delete deployment %s/%s failed: %v", m.Namespace, m.Name, err)
		}
	}
}

func deletePrereqNamespaces(ctx context.Context, k8s kubernetes.Interface) []string {
	prereqLogf("deleting namespaces (background)")
	list, err := k8s.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		prereqLogf("  list namespaces failed: %v", err)
		return nil
	}
	var targets []string
	for _, ns := range list.Items {
		if !matchesAnySubstring(ns.Name, prereqNamespacePatterns) {
			continue
		}
		if isProtectedNamespace(ns.Name) {
			prereqLogf("  protected (OpenShift/system), NOT deleting: %s", ns.Name)
			continue
		}
		targets = append(targets, ns.Name)
		err := k8s.CoreV1().Namespaces().Delete(ctx, ns.Name, prereqDeleteOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			prereqLogf("  delete namespace %s failed: %v", ns.Name, err)
		}
	}
	return targets
}

func deleteBrokenAPIServices(ctx context.Context, dyn dynamic.Interface) {
	prereqLogf("deleting broken aggregated APIServices")
	list, err := dyn.Resource(apiServiceGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		prereqLogf("  list apiservices failed: %v", err)
		return
	}
	for _, item := range list.Items {
		service, found, _ := nestedMap(item.Object, "spec", "service")
		if !found || service == nil {
			continue
		}
		// Never delete an aggregated API served by an OpenShift/kube platform
		// namespace (e.g. metrics-server in openshift-monitoring); it may be
		// transiently unavailable and is not ours to remove.
		if svcNS, _ := service["namespace"].(string); isProtectedNamespace(svcNS) {
			continue
		}
		if apiServiceAvailable(item.Object) {
			continue
		}
		deleteClusterScoped(ctx, dyn, apiServiceGVR, item.GetName())
	}
}

func stripStuckPrereqFinalizers(
	ctx context.Context,
	dyn dynamic.Interface,
	apiext apiextensionsclient.Interface,
	k8s kubernetes.Interface,
	targetNamespaces []string,
) {
	prereqLogf("stripping finalizers from anything stuck")
	for _, crdName := range prereqFinalizerHeavyCRDKinds {
		gvr, namespaced, ok := crdGVR(ctx, apiext, crdName)
		if !ok {
			continue
		}
		forEachCR(ctx, dyn, gvr, namespaced, func(namespace, name string) {
			stripCRFinalizers(ctx, dyn, gvr, namespace, name)
		})
	}

	for _, name := range targetNamespaces {
		stripNamespaceFinalizers(ctx, k8s, name)
	}
}

func deletePrereqCRDs(ctx context.Context, apiext apiextensionsclient.Interface, openshift bool) {
	prereqLogf("deleting matching CRDs")
	list, err := apiext.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		prereqLogf("  list CRDs failed: %v", err)
		return
	}
	for _, crd := range list.Items {
		if !matchesAnySubstring(crd.Name, prereqCRDPatterns) {
			continue
		}
		if openshift && matchesAnySubstring(crd.Name, ocpOwnedCRDSubstrings) {
			prereqLogf("  protected (OpenShift-owned), NOT deleting: %s", crd.Name)
			continue
		}
		err := apiext.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, crd.Name, prereqDeleteOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			prereqLogf("  delete CRD %s failed: %v", crd.Name, err)
		}
	}
}

// crdGVR resolves a CRD by name into the GroupVersionResource of its served
// (storage) version and reports whether the resource is namespaced. ok is false
// when the CRD is absent so callers can skip it.
func crdGVR(ctx context.Context, apiext apiextensionsclient.Interface, crdName string) (schema.GroupVersionResource, bool, bool) {
	crd, err := apiext.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		return schema.GroupVersionResource{}, false, false
	}
	version := servedVersion(crd)
	if version == "" {
		return schema.GroupVersionResource{}, false, false
	}
	gvr := schema.GroupVersionResource{
		Group:    crd.Spec.Group,
		Version:  version,
		Resource: crd.Spec.Names.Plural,
	}
	return gvr, crd.Spec.Scope == apiextensionsv1.NamespaceScoped, true
}

func servedVersion(crd *apiextensionsv1.CustomResourceDefinition) string {
	for _, v := range crd.Spec.Versions {
		if v.Storage && v.Served {
			return v.Name
		}
	}
	for _, v := range crd.Spec.Versions {
		if v.Served {
			return v.Name
		}
	}
	return ""
}

func forEachCR(ctx context.Context, dyn dynamic.Interface, gvr schema.GroupVersionResource, namespaced bool, fn func(namespace, name string)) {
	list, err := dyn.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			prereqLogf("  list %s failed: %v", gvr.Resource, err)
		}
		return
	}
	for _, item := range list.Items {
		namespace := ""
		if namespaced {
			namespace = item.GetNamespace()
		}
		fn(namespace, item.GetName())
	}
}

func deleteOne(ctx context.Context, dyn dynamic.Interface, gvr schema.GroupVersionResource, namespace, name string) {
	var err error
	if namespace == "" {
		err = dyn.Resource(gvr).Delete(ctx, name, prereqDeleteOptions)
	} else {
		err = dyn.Resource(gvr).Namespace(namespace).Delete(ctx, name, prereqDeleteOptions)
	}
	if err != nil && !k8serrors.IsNotFound(err) {
		prereqLogf("  delete %s %s/%s failed: %v", gvr.Resource, namespace, name, err)
	}
}

func deleteClusterScoped(ctx context.Context, dyn dynamic.Interface, gvr schema.GroupVersionResource, name string) {
	if err := dyn.Resource(gvr).Delete(ctx, name, prereqDeleteOptions); err != nil && !k8serrors.IsNotFound(err) {
		prereqLogf("  delete %s %s failed: %v", gvr.Resource, name, err)
	}
}

func stripCRFinalizers(ctx context.Context, dyn dynamic.Interface, gvr schema.GroupVersionResource, namespace, name string) {
	patch := []byte(`{"metadata":{"finalizers":null}}`)
	var err error
	if namespace == "" {
		_, err = dyn.Resource(gvr).Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	} else {
		_, err = dyn.Resource(gvr).Namespace(namespace).Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	}
	if err != nil && !k8serrors.IsNotFound(err) {
		prereqLogf("  strip finalizers %s %s/%s failed: %v", gvr.Resource, namespace, name, err)
	}
}

func stripNamespaceFinalizers(ctx context.Context, k8s kubernetes.Interface, name string) {
	if isProtectedNamespace(name) {
		return
	}
	ns, err := k8s.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return
	}
	if err != nil {
		prereqLogf("  get namespace %s failed: %v", name, err)
		return
	}
	if len(ns.Spec.Finalizers) == 0 && len(ns.Finalizers) == 0 {
		return
	}
	ns.Spec.Finalizers = nil
	ns.Finalizers = nil
	if _, err := k8s.CoreV1().Namespaces().Finalize(ctx, ns, metav1.UpdateOptions{}); err != nil {
		prereqLogf("  finalize namespace %s failed: %v", name, err)
	}
}

func apiServiceAvailable(obj map[string]interface{}) bool {
	conditions, found, _ := nestedSlice(obj, "status", "conditions")
	if !found {
		return true
	}
	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if cond["type"] == "Available" {
			return cond["status"] == "True"
		}
	}
	return true
}

func nestedMap(obj map[string]interface{}, keys ...string) (map[string]interface{}, bool, error) {
	cur := obj
	for i, k := range keys {
		v, ok := cur[k]
		if !ok || v == nil {
			return nil, false, nil
		}
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, false, nil
		}
		if i == len(keys)-1 {
			return m, true, nil
		}
		cur = m
	}
	return nil, false, nil
}

func nestedSlice(obj map[string]interface{}, keys ...string) ([]interface{}, bool, error) {
	last := keys[len(keys)-1]
	parent := obj
	if len(keys) > 1 {
		m, found, _ := nestedMap(obj, keys[:len(keys)-1]...)
		if !found {
			return nil, false, nil
		}
		parent = m
	}
	v, ok := parent[last]
	if !ok || v == nil {
		return nil, false, nil
	}
	s, ok := v.([]interface{})
	if !ok {
		return nil, false, nil
	}
	return s, true, nil
}

func matchesAnySubstring(name string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

func prereqLogf(format string, args ...interface{}) {
	fmt.Printf("[CleanupPrerequisites] "+format+"\n", args...)
}

func helmNoopLog(string, ...interface{}) {}
