/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package uninstall

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/helmclient"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type HelmWrapper interface {
	// FixPendingRelease transitions a release stuck in a pending state (pending-install,
	// pending-upgrade, or pending-rollback) to failed so that a subsequent
	// UninstallRelease call can proceed without being blocked by the pending-state
	// protection. It is a no-op when the release does not exist or is not pending.
	FixPendingRelease(releaseName, namespace string) error
	UninstallRelease(releaseName, namespace string, waitTimeout time.Duration) error
}

type K8sWrapper interface {
	// ForceDeleteNamespace clears finalizers from all resources in the namespace
	// before deleting it. This prevents the namespace from getting stuck in
	// Terminating state when controllers that own finalizers have already been
	// removed (e.g. after helm uninstall).
	ForceDeleteNamespace(namespace string) error
	NamespaceExists(namespace string) (bool, error)
	ListCRDs() ([]string, error)
	DeleteCRD(ctx context.Context, name string) error

	// A webhook configuration is Run:AI's when it matches the label selector
	// (e.g. "app=runai") or routes a webhook to a service in one of the given
	// namespaces. The label alone misses configurations written by older
	// versions, which init-ca creates at runtime and Helm therefore never owns.
	ListMutatingWebhookConfigurations(labelSelector string, serviceNamespaces []string) ([]string, error)
	DeleteMutatingWebhookConfiguration(name string) error
	ListValidatingWebhookConfigurations(labelSelector string, serviceNamespaces []string) ([]string, error)
	DeleteValidatingWebhookConfiguration(name string) error

	// ListResourceInterfaces accepts a label key prefix rather than a full
	// Kubernetes label selector because Kubernetes selectors do not support
	// key-prefix matching.  All Run:AI ResourceInterface objects carry labels
	// whose keys start with "run.ai/" (e.g. "run.ai/resource-interface-group"),
	// but the specific key names may change over time.  Filtering by prefix
	// client-side ensures we catch every Run:AI-owned object regardless of
	// which exact label key is used, and tests can assert on the same broad
	// criterion instead of a single hard-coded key.
	ListResourceInterfaces(labelKeyPrefix string) ([]string, error)
	DeleteResourceInterface(name string) error
	// ListClusterRoles accepts a label key prefix for the same reason as
	// ListResourceInterfaces: Kubernetes selectors do not support key-prefix
	// matching, and Run:AI ClusterRoles are identified by labels whose keys
	// start with "run.ai/" rather than by a single stable label key.
	ListClusterRoles(labelKeyPrefix string) ([]string, error)
	DeleteClusterRole(name string) error
	// ListPriorityClasses accepts a label key prefix for the same reason as
	// ListClusterRoles and ListResourceInterfaces: Kubernetes selectors do not
	// support key-prefix matching, and Run:AI PriorityClasses are identified
	// only by the presence of labels whose keys start with "run.ai/".
	ListPriorityClasses(labelKeyPrefix string) ([]string, error)
	DeletePriorityClass(name string) error

	// StripRunAICRFinalizers removes finalizers from all custom resources whose
	// CRD belongs to the run.ai API group. This must be called before CRDs are
	// deleted so that the garbage collector can finish cleaning up CRs instead
	// of leaving CRDs stuck in Terminating state.
	StripRunAICRFinalizers(ctx context.Context) error
}

type Deps struct {
	HelmWrapper HelmWrapper
	K8sWrapper  K8sWrapper
}

type helmWrapperImpl struct{}

func (h *helmWrapperImpl) FixPendingRelease(releaseName, namespace string) error {
	return helmclient.FixPendingRelease(releaseName, namespace)
}

func (h *helmWrapperImpl) UninstallRelease(releaseName, namespace string, waitTimeout time.Duration) error {
	return helmclient.UninstallRelease(releaseName, namespace, waitTimeout)
}

type k8sWrapperImpl struct {
	k8sClient k8s.K8sClients
}

func (w *k8sWrapperImpl) newAPIExtClient() (apiextensionsclient.Interface, error) {
	return apiextensionsclient.NewForConfig(w.k8sClient.GetConfig())
}

func (w *k8sWrapperImpl) newDynamicClient() (dynamic.Interface, error) {
	return dynamic.NewForConfig(w.k8sClient.GetConfig())
}

func (w *k8sWrapperImpl) newKubernetesClient() (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(w.k8sClient.GetConfig())
}

func (w *k8sWrapperImpl) ListCRDs() ([]string, error) {
	apiExtClient, err := w.newAPIExtClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create apiextensions client: %w", err)
	}
	crdList, err := apiExtClient.ApiextensionsV1().CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CRDs: %w", err)
	}
	names := make([]string, 0, len(crdList.Items))
	for _, crd := range crdList.Items {
		names = append(names, crd.Name)
	}
	return names, nil
}

func (w *k8sWrapperImpl) DeleteCRD(ctx context.Context, name string) error {
	apiExtClient, err := w.newAPIExtClient()
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %w", err)
	}
	crds := apiExtClient.ApiextensionsV1().CustomResourceDefinitions()
	if err := crds.Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			logging.Infof("[Uninstall] CRD '%s' not found, skipping deletion", name)
			return nil
		}
		return fmt.Errorf("failed to delete CRD %s: %w", name, err)
	}
	var finalizers []string
	if err := wait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
		crd, err := crds.Get(ctx, name, metav1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		finalizers = append(finalizers[:0], crd.Finalizers...)
		return false, nil
	}); err != nil {
		return fmt.Errorf("CRD %s was not deleted; finalizers=%v: %w", name, finalizers, err)
	}
	logging.Infof("[Uninstall] CRD '%s' deletion completed", name)
	return nil
}

func ownsWebhookConfiguration(objLabels map[string]string, selector k8slabels.Selector,
	serviceNamespaces map[string]struct{}, webhookServiceNamespaces []string) bool {
	if selector.Matches(k8slabels.Set(objLabels)) {
		return true
	}
	for _, ns := range webhookServiceNamespaces {
		if _, ok := serviceNamespaces[ns]; ok {
			return true
		}
	}
	return false
}

func namespaceSet(namespaces []string) map[string]struct{} {
	set := make(map[string]struct{}, len(namespaces))
	for _, ns := range namespaces {
		set[ns] = struct{}{}
	}
	return set
}

func (w *k8sWrapperImpl) ListMutatingWebhookConfigurations(labelSelector string, serviceNamespaces []string) ([]string, error) {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	selector, err := k8slabels.Parse(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse label selector %q: %w", labelSelector, err)
	}
	list, err := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list MutatingWebhookConfigurations: %w", err)
	}
	nsSet := namespaceSet(serviceNamespaces)
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		hookNamespaces := make([]string, 0, len(item.Webhooks))
		for _, hook := range item.Webhooks {
			if hook.ClientConfig.Service != nil {
				hookNamespaces = append(hookNamespaces, hook.ClientConfig.Service.Namespace)
			}
		}
		if ownsWebhookConfiguration(item.Labels, selector, nsSet, hookNamespaces) {
			names = append(names, item.Name)
		}
	}
	return names, nil
}

func (w *k8sWrapperImpl) DeleteMutatingWebhookConfiguration(name string) error {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	if err := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			logging.Infof("[Uninstall] MutatingWebhookConfiguration '%s' not found, skipping deletion", name)
			return nil
		}
		return fmt.Errorf("failed to delete MutatingWebhookConfiguration %s: %w", name, err)
	}
	logging.Infof("[Uninstall] MutatingWebhookConfiguration '%s' deleted successfully", name)
	return nil
}

func (w *k8sWrapperImpl) ListValidatingWebhookConfigurations(labelSelector string, serviceNamespaces []string) ([]string, error) {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	selector, err := k8slabels.Parse(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse label selector %q: %w", labelSelector, err)
	}
	list, err := clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ValidatingWebhookConfigurations: %w", err)
	}
	nsSet := namespaceSet(serviceNamespaces)
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		hookNamespaces := make([]string, 0, len(item.Webhooks))
		for _, hook := range item.Webhooks {
			if hook.ClientConfig.Service != nil {
				hookNamespaces = append(hookNamespaces, hook.ClientConfig.Service.Namespace)
			}
		}
		if ownsWebhookConfiguration(item.Labels, selector, nsSet, hookNamespaces) {
			names = append(names, item.Name)
		}
	}
	return names, nil
}

func (w *k8sWrapperImpl) DeleteValidatingWebhookConfiguration(name string) error {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	if err := clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			logging.Infof("[Uninstall] ValidatingWebhookConfiguration '%s' not found, skipping deletion", name)
			return nil
		}
		return fmt.Errorf("failed to delete ValidatingWebhookConfiguration %s: %w", name, err)
	}
	logging.Infof("[Uninstall] ValidatingWebhookConfiguration '%s' deleted successfully", name)
	return nil
}

var resourceInterfaceGVR = schema.GroupVersionResource{
	Group:    "optimization.nvidia.com",
	Version:  "v1",
	Resource: "resourceinterfaces",
}

// ListResourceInterfaces returns the names of all ResourceInterface objects
// that have at least one label whose key starts with labelKeyPrefix.
// Kubernetes label selectors do not support key-prefix matching, so filtering
// is done client-side after a full list.  See K8sWrapper interface for the
// full rationale.
func (w *k8sWrapperImpl) ListResourceInterfaces(labelKeyPrefix string) ([]string, error) {
	dynamicClient, err := w.newDynamicClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	list, err := dynamicClient.Resource(resourceInterfaceGVR).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// The optimization.nvidia.com CRD is not installed on this cluster — no instances can exist.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list ResourceInterfaces: %w", err)
	}
	var names []string
	for _, item := range list.Items {
		for key := range item.GetLabels() {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, item.GetName())
				break
			}
		}
	}
	return names, nil
}

func (w *k8sWrapperImpl) DeleteResourceInterface(name string) error {
	dynamicClient, err := w.newDynamicClient()
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}
	if err := dynamicClient.Resource(resourceInterfaceGVR).Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			logging.Infof("[Uninstall] ResourceInterface '%s' not found, skipping deletion", name)
			return nil
		}
		return fmt.Errorf("failed to delete ResourceInterface %s: %w", name, err)
	}
	logging.Infof("[Uninstall] ResourceInterface '%s' deleted successfully", name)
	return nil
}

func (w *k8sWrapperImpl) ListClusterRoles(labelKeyPrefix string) ([]string, error) {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	list, err := clientset.RbacV1().ClusterRoles().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ClusterRoles: %w", err)
	}
	var names []string
	for _, item := range list.Items {
		for key := range item.GetLabels() {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, item.Name)
				break
			}
		}
	}
	return names, nil
}

func (w *k8sWrapperImpl) DeleteClusterRole(name string) error {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	if err := clientset.RbacV1().ClusterRoles().Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			logging.Infof("[Uninstall] ClusterRole '%s' not found, skipping deletion", name)
			return nil
		}
		return fmt.Errorf("failed to delete ClusterRole %s: %w", name, err)
	}
	logging.Infof("[Uninstall] ClusterRole '%s' deleted successfully", name)
	return nil
}

func (w *k8sWrapperImpl) ListPriorityClasses(labelKeyPrefix string) ([]string, error) {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	list, err := clientset.SchedulingV1().PriorityClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list PriorityClasses: %w", err)
	}
	var names []string
	for _, item := range list.Items {
		for key := range item.GetLabels() {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, item.Name)
				break
			}
		}
	}
	return names, nil
}

func (w *k8sWrapperImpl) DeletePriorityClass(name string) error {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	if err := clientset.SchedulingV1().PriorityClasses().Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			logging.Infof("[Uninstall] PriorityClass '%s' not found, skipping deletion", name)
			return nil
		}
		return fmt.Errorf("failed to delete PriorityClass %s: %w", name, err)
	}
	logging.Infof("[Uninstall] PriorityClass '%s' deleted successfully", name)
	return nil
}

// StripRunAICRFinalizers removes finalizers from all CRs belonging to the
// run.ai API group so that pending CRD deletions can complete.
func (w *k8sWrapperImpl) StripRunAICRFinalizers(ctx context.Context) error {
	apiExtClient, err := w.newAPIExtClient()
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %w", err)
	}

	crdList, err := apiExtClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list CRDs: %w", err)
	}

	dynamicClient, err := w.newDynamicClient()
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	patch := []byte(`{"metadata":{"finalizers":null}}`)
	var failures []error

	for _, crd := range crdList.Items {
		if !strings.HasSuffix(strings.ToLower(crd.Name), ".run.ai") {
			continue
		}

		// Prefer the storage version: listing through a non-storage version requires
		// a conversion webhook that may already be gone after helm uninstall.
		// Fall back to the first served version if no storage version is found.
		version := ""
		for _, v := range crd.Spec.Versions {
			if v.Served {
				if version == "" {
					version = v.Name
				}
				if v.Storage {
					version = v.Name
					break
				}
			}
		}
		if version == "" {
			logging.Infof("[Uninstall] CRD %s has no served version, skipping finalizer strip", crd.Name)
			continue
		}

		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  version,
			Resource: crd.Spec.Names.Plural,
		}

		// List across all namespaces for namespace-scoped CRDs; cluster-scoped
		// CRDs do not accept a namespace on the list call.
		var list *unstructured.UnstructuredList
		if crd.Spec.Scope == apiextensionsv1.ClusterScoped {
			list, err = dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
		} else {
			list, err = dynamicClient.Resource(gvr).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("list %s CRs: %w", crd.Name, err))
			continue
		}

		for _, item := range list.Items {
			if len(item.GetFinalizers()) == 0 {
				continue
			}
			logging.Infof("[Uninstall] Stripping finalizers from %s %s", crd.Spec.Names.Kind, item.GetName())
			if crd.Spec.Scope == apiextensionsv1.ClusterScoped {
				_, err = dynamicClient.Resource(gvr).Patch(ctx, item.GetName(), types.MergePatchType, patch, metav1.PatchOptions{})
			} else {
				_, err = dynamicClient.Resource(gvr).Namespace(item.GetNamespace()).Patch(ctx, item.GetName(), types.MergePatchType, patch, metav1.PatchOptions{})
			}
			if err != nil {
				failures = append(failures, fmt.Errorf("strip finalizers from %s %s: %w", crd.Spec.Names.Kind, item.GetName(), err))
			}
		}
	}

	return errors.Join(failures...)
}

// ForceDeleteNamespace clears finalizers from all resources in the namespace and
// then deletes it. Finalizer clearing is best-effort: failures are logged but do
// not stop the deletion. This prevents the namespace from getting stuck in
// Terminating state when the controllers that own those finalizers are already gone.
func (w *k8sWrapperImpl) ForceDeleteNamespace(namespace string) error {
	w.clearNamespaceResourceFinalizers(namespace)

	ctx := context.Background()
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	if err := w.k8sClient.GetClient().Delete(ctx, ns); err != nil {
		if k8serrors.IsNotFound(err) {
			logging.Infof("[Uninstall] Namespace '%s' not found, skipping deletion", namespace)
			return nil
		}
		logging.Errorf("[Uninstall] Failed to delete namespace '%s': %v", namespace, err)
		return fmt.Errorf("failed to delete namespace %s: %w", namespace, err)
	}
	logging.Infof("[Uninstall] Namespace '%s' deleted successfully", namespace)
	return nil
}

// clearNamespaceResourceFinalizers removes finalizers from all resources in the
// namespace. Uses the discovery API to enumerate resource types generically —
// works for any CRDs or built-in resources without hardcoding specific types.
// All errors are logged and skipped; we clear as many as possible.
func (w *k8sWrapperImpl) clearNamespaceResourceFinalizers(namespace string) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(w.k8sClient.GetConfig())
	if err != nil {
		logging.Infof("[Uninstall] Warning: could not create discovery client to clear finalizers in '%s': %v", namespace, err)
		return
	}

	// ServerGroupsAndResources may return partial results alongside an error
	// (e.g. an aggregated API is unavailable). Proceed with whatever we got.
	_, resourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil && resourceLists == nil {
		logging.Infof("[Uninstall] Warning: could not list API resources to clear finalizers in '%s': %v", namespace, err)
		return
	}

	dynamicClient, err := w.newDynamicClient()
	if err != nil {
		logging.Infof("[Uninstall] Warning: could not create dynamic client to clear finalizers in '%s': %v", namespace, err)
		return
	}

	patch := []byte(`{"metadata":{"finalizers":null}}`)
	ctx := context.Background()

	for _, resourceList := range resourceLists {
		// Skip if the group/version string is malformed (e.g. from a broken aggregated API).
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}
		for _, resource := range resourceList.APIResources {
			// Subresources (e.g. "pods/log") are operations on a parent, not
			// standalone objects — they never have finalizers, skip them.
			if strings.Contains(resource.Name, "/") || !resource.Namespaced {
				continue
			}
			// Only process resources that support list+patch; others would return 405.
			if !hasVerb(resource.Verbs, "list") || !hasVerb(resource.Verbs, "patch") {
				continue
			}

			gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name}
			list, err := dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				continue
			}

			for _, item := range list.Items {
				if len(item.GetFinalizers()) == 0 {
					continue
				}
				logging.Infof("[Uninstall] Clearing finalizers from %s/%s in namespace '%s'", resource.Name, item.GetName(), namespace)
				if _, err := dynamicClient.Resource(gvr).Namespace(namespace).Patch(
					ctx, item.GetName(), types.MergePatchType, patch, metav1.PatchOptions{},
				); err != nil {
					logging.Infof("[Uninstall] Warning: could not clear finalizers from %s/%s: %v", resource.Name, item.GetName(), err)
				}
			}
		}
	}
}

func hasVerb(verbs []string, verb string) bool {
	for _, v := range verbs {
		if v == verb {
			return true
		}
	}
	return false
}

func (w *k8sWrapperImpl) NamespaceExists(namespace string) (bool, error) {
	clientset, err := w.newKubernetesClient()
	if err != nil {
		return false, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	_, err = clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check namespace %s: %w", namespace, err)
	}
	return true, nil
}

func ProductionDeps() (Deps, error) {
	k8sClients, err := k8s.GetK8sClients()
	if err != nil {
		logging.Errorf("[Uninstall] Failed to initialize k8s clients: %v", err)
		return Deps{}, fmt.Errorf("failed to get k8s clients: %w", err)
	}
	return Deps{
		HelmWrapper: &helmWrapperImpl{},
		K8sWrapper:  &k8sWrapperImpl{k8sClient: k8sClients},
	}, nil
}
