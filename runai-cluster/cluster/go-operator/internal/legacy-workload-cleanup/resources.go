package legacycleanup

import (
	"context"
	"fmt"
	"slices"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var legacyControllerDeployments = []string{
	"workload-controller",
	"inference-workload-controller",
	"runai-job-controller",
}

const legacyControllerPodPollInterval = time.Second

type namedResourceSet struct {
	gvr       schema.GroupVersionResource
	kind      string
	namespace string
	names     []string
}

var remainingLegacyResources = []namedResourceSet{
	{gvr: schema.GroupVersionResource{Version: "v1", Resource: "services"}, kind: "Service", namespace: "runai",
		names: []string{"workload-controller", "inference-workload-controller", "runai-job-controller"}},
	{gvr: schema.GroupVersionResource{Version: "v1", Resource: "serviceaccounts"}, kind: "ServiceAccount", namespace: "runai",
		names: []string{"workload-controller", "inference-workload-controller", "runai-job-controller"}},
	{gvr: schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, kind: "Secret", namespace: "runai",
		names: []string{"runai-workload-controller-tls-secret", "runai-inference-workload-controller-tls-secret"}},
	{gvr: schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}, kind: "Role", namespace: "runai",
		names: []string{"workload-controller", "inference-workload-controller", "workload-controller-migration"}},
	{gvr: schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}, kind: "RoleBinding", namespace: "runai",
		names: []string{"workload-controller", "inference-workload-controller", "workload-controller-migration"}},
	{gvr: schema.GroupVersionResource{Group: "monitoring.coreos.com", Version: "v1", Resource: "servicemonitors"}, kind: "ServiceMonitor", namespace: "runai",
		names: []string{"workload-controller", "inference-workload-controller"}},
	{gvr: schema.GroupVersionResource{Group: "autoscaling.k8s.io", Version: "v1", Resource: "verticalpodautoscalers"}, kind: "VerticalPodAutoscaler", namespace: "runai",
		names: []string{"workload-controller", "inference-workload-controller", "runai-job-controller"}},
	{gvr: schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, kind: "ClusterRole",
		names: []string{"workload-controller-rw", "workload-controller-ro", "runai-inference-workload-controller-project", "runai-inference-workload-controller", "runai-job-controller", "runai-job-controller-project"}},
	{gvr: schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, kind: "ClusterRoleBinding",
		names: []string{"workload-controller-ro", "runai-inference-workload-controller", "runai-job-controller"}},
	{gvr: schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations"}, kind: "ValidatingWebhookConfiguration",
		names: []string{
			"runai-interactive-policy-validation",
			"runai-interactive-workload-validation",
			"runai-training-policy-validation",
			"runai-training-workload-validation",
			"runai-distributed-policy-validation",
			"runai-distributed-workload-validation",
			"runai-inference-policy-validation-old",
			"runai-inference-policy-validation",
			"runai-inference-workload-validation-old",
			"runai-inference-workload-validation",
			"runai-distributed-inference-workload-validation",
		}},
	{gvr: schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"}, kind: "MutatingWebhookConfiguration",
		names: []string{
			"runai-interactive-workload-mutation",
			"runai-training-workload-mutation",
			"runai-distributed-workload-mutation",
			"runai-inference-workload-mutation-old",
			"runai-inference-workload-mutation",
			"runai-distributed-inference-workload-mutation",
		}},
}

type projectLegacyRoleBinding struct {
	name                string
	roleName            string
	serviceAccountNames []string
}

var projectLegacyRoleBindings = []projectLegacyRoleBinding{
	{name: "workload-controller-rw", roleName: "workload-controller-rw", serviceAccountNames: []string{"workload-controller"}},
	{name: "runai-job-controller-project", roleName: "runai-job-controller-project", serviceAccountNames: []string{"runai-job-controller", "runai-project-controller"}},
	{name: "inference-workload-controller", roleName: "runai-inference-workload-controller-project", serviceAccountNames: []string{"inference-workload-controller"}},
}

func deleteLegacyControllerDeployments(ctx context.Context, k8sClient client.Client) error {
	for _, name := range legacyControllerDeployments {
		deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "runai", Name: name}}
		if err := k8sClient.Delete(ctx, deployment); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete legacy Deployment runai/%s: %w", name, err)
		}
	}
	return nil
}

func LegacyControllerResourcesPresent(ctx context.Context, k8sClient client.Client) (bool, error) {
	for _, name := range legacyControllerDeployments {
		deployment := &appsv1.Deployment{}
		err := k8sClient.Get(ctx, client.ObjectKey{Namespace: "runai", Name: name}, deployment)
		if err == nil {
			return true, nil
		}
		if !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("get legacy Deployment runai/%s: %w", name, err)
		}
	}
	for _, resourceSet := range remainingLegacyResources {
		apiVersion := resourceSet.gvr.Version
		if resourceSet.gvr.Group != "" {
			apiVersion = resourceSet.gvr.Group + "/" + resourceSet.gvr.Version
		}
		for _, name := range resourceSet.names {
			object := &unstructured.Unstructured{}
			object.SetAPIVersion(apiVersion)
			object.SetKind(resourceSet.kind)
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: resourceSet.namespace, Name: name}, object)
			if err == nil {
				return true, nil
			}
			if !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
				return false, fmt.Errorf("get legacy %s %s: %w", resourceSet.kind, name, err)
			}
		}
	}
	roleBindings := &rbacv1.RoleBindingList{}
	if err := k8sClient.List(ctx, roleBindings); err != nil {
		return false, fmt.Errorf("list RoleBindings for legacy controller detection: %w", err)
	}
	for i := range roleBindings.Items {
		if isProjectLegacyRoleBinding(&roleBindings.Items[i]) {
			return true, nil
		}
	}
	return false, nil
}

func waitForLegacyControllerPods(ctx context.Context, k8sClient client.Client, interval time.Duration) error {
	return wait.PollUntilContextCancel(ctx, interval, true, func(ctx context.Context) (bool, error) {
		pods := &corev1.PodList{}
		if err := k8sClient.List(ctx, pods, client.InNamespace("runai")); err != nil {
			return false, fmt.Errorf("list pods while waiting for legacy controllers to stop: %w", err)
		}
		for i := range pods.Items {
			if slices.Contains(legacyControllerDeployments, pods.Items[i].Labels["app"]) {
				return false, nil
			}
		}
		return true, nil
	})
}

func deleteRemainingLegacyResources(ctx context.Context, dynamicClient dynamic.Interface) error {
	for _, resourceSet := range remainingLegacyResources {
		resource := dynamicClient.Resource(resourceSet.gvr)
		for _, name := range resourceSet.names {
			var err error
			if resourceSet.namespace == "" {
				err = resource.Delete(ctx, name, metav1.DeleteOptions{})
			} else {
				err = resource.Namespace(resourceSet.namespace).Delete(ctx, name, metav1.DeleteOptions{})
			}
			if err != nil && !apierrors.IsNotFound(err) {
				return fmt.Errorf("delete legacy %s %s: %w", resourceSet.gvr.Resource, name, err)
			}
		}
	}
	return deleteProjectLegacyRoleBindings(ctx, dynamicClient)
}

func deleteProjectLegacyRoleBindings(ctx context.Context, dynamicClient dynamic.Interface) error {
	gvr := schema.GroupVersionResource{
		Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings",
	}
	resource := dynamicClient.Resource(gvr)
	list, err := resource.Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("list RoleBindings for legacy controller cleanup: %w", err)
	}
	for i := range list.Items {
		item := &list.Items[i]
		roleBinding := &rbacv1.RoleBinding{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, roleBinding); err != nil ||
			!isProjectLegacyRoleBinding(roleBinding) {
			continue
		}
		if err := resource.Namespace(item.GetNamespace()).Delete(ctx, item.GetName(), metav1.DeleteOptions{}); err != nil &&
			!apierrors.IsNotFound(err) {
			return fmt.Errorf("delete legacy RoleBinding %s/%s: %w", item.GetNamespace(), item.GetName(), err)
		}
	}
	return nil
}

func isProjectLegacyRoleBinding(roleBinding *rbacv1.RoleBinding) bool {
	for _, legacy := range projectLegacyRoleBindings {
		if roleBinding.Name != legacy.name ||
			roleBinding.RoleRef.APIGroup != rbacv1.GroupName ||
			roleBinding.RoleRef.Kind != "ClusterRole" ||
			roleBinding.RoleRef.Name != legacy.roleName {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.APIGroup == "" &&
				subject.Kind == "ServiceAccount" &&
				subject.Namespace == "runai" &&
				slices.Contains(legacy.serviceAccountNames, subject.Name) {
				return true
			}
		}
	}
	return false
}
