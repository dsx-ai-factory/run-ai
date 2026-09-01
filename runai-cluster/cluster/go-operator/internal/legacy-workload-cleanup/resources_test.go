package legacycleanup

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

var _ = Describe("Legacy controller resources", func() {
	It("waits while a legacy controller pod still exists", func() {
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Namespace: "runai",
			Name:      "workload-controller-old",
			Labels:    map[string]string{"app": "workload-controller"},
		}}
		k8sClient := runnerClient(pod)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		DeferCleanup(cancel)

		Expect(waitForLegacyControllerPods(ctx, k8sClient, time.Millisecond)).To(MatchError(context.DeadlineExceeded))
	})

	It("completes when no legacy controller pods remain", func() {
		k8sClient := runnerClient()

		Expect(waitForLegacyControllerPods(context.Background(), k8sClient, time.Millisecond)).To(Succeed())
	})

	It("deletes only allowlisted legacy resources and supports retries", func() {
		serviceGVR := schema.GroupVersionResource{Version: "v1", Resource: "services"}
		roleBindingGVR := schema.GroupVersionResource{
			Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings",
		}
		legacy := newObject("v1", "Service", "runai", "workload-controller", "legacy-service")
		unrelated := newObject("v1", "Service", "runai", "workload-controller-copy", "unrelated-service")
		dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{roleBindingGVR: "RoleBindingList"},
			legacy,
			unrelated,
		)

		Expect(deleteRemainingLegacyResources(context.Background(), dynamicClient)).To(Succeed())
		Expect(deleteRemainingLegacyResources(context.Background(), dynamicClient)).To(Succeed())

		_, err := dynamicClient.Resource(serviceGVR).Namespace("runai").Get(
			context.Background(), "workload-controller", metav1.GetOptions{},
		)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
		_, err = dynamicClient.Resource(serviceGVR).Namespace("runai").Get(
			context.Background(), "workload-controller-copy", metav1.GetOptions{},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("preserves a same-named RoleBinding with a different identity", func() {
		roleBindingGVR := schema.GroupVersionResource{
			Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings",
		}
		lookalike := legacyRoleBinding("team-a", "workload-controller-rw", "unrelated-role", "unrelated-controller")
		dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{roleBindingGVR: "RoleBindingList"},
			lookalike,
		)

		Expect(deleteProjectLegacyRoleBindings(context.Background(), dynamicClient)).To(Succeed())
		_, err := dynamicClient.Resource(roleBindingGVR).Namespace("team-a").Get(
			context.Background(), "workload-controller-rw", metav1.GetOptions{},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("deletes a RoleBinding only when its full legacy identity matches", func() {
		roleBindingGVR := schema.GroupVersionResource{
			Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings",
		}
		legacy := legacyRoleBinding("team-a", "workload-controller-rw", "workload-controller-rw", "workload-controller")
		dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{roleBindingGVR: "RoleBindingList"},
			legacy,
		)

		Expect(deleteProjectLegacyRoleBindings(context.Background(), dynamicClient)).To(Succeed())
		_, err := dynamicClient.Resource(roleBindingGVR).Namespace("team-a").Get(
			context.Background(), "workload-controller-rw", metav1.GetOptions{},
		)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("does not detect a same-named RoleBinding with a different identity", func() {
		lookalike := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Namespace: "team-a", Name: "workload-controller-rw"},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     "unrelated-role",
			},
			Subjects: []rbacv1.Subject{{
				Kind:      "ServiceAccount",
				Name:      "unrelated-controller",
				Namespace: "runai",
			}},
		}

		present, err := LegacyControllerResourcesPresent(context.Background(), runnerClient(lookalike))
		Expect(err).NotTo(HaveOccurred())
		Expect(present).To(BeFalse())
	})
})

func legacyRoleBinding(namespace, name, roleName, subjectName string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "rbac.authorization.k8s.io/v1",
		"kind":       "RoleBinding",
		"metadata": map[string]any{
			"namespace": namespace,
			"name":      name,
		},
		"roleRef": map[string]any{
			"apiGroup": rbacv1.GroupName,
			"kind":     "ClusterRole",
			"name":     roleName,
		},
		"subjects": []any{map[string]any{
			"kind":      "ServiceAccount",
			"name":      subjectName,
			"namespace": "runai",
		}},
	}}
}
