package client

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

const mockKubeConfigWithSingleCluster = `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:12345
  name: local
contexts:
- context:
    cluster: local
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: dummy
`

var _ = Describe("Client", func() {
	Describe("NewClient", func() {
		It("should return a client when kubeconfig is valid", func() {
			path := writeKubeConfigFile(mockKubeConfigWithSingleCluster)
			defer os.Remove(path)

			os.Setenv("KUBECONFIG", path)
			defer os.Unsetenv("KUBECONFIG")

			c, err := NewClient()
			Expect(err).NotTo(HaveOccurred())
			Expect(c).NotTo(BeNil())
		})

		It("should return an error when kubeconfig is invalid", func() {
			path := writeKubeConfigFile("bad yaml content")
			defer os.Remove(path)

			os.Setenv("KUBECONFIG", path)
			defer os.Unsetenv("KUBECONFIG")

			c, err := NewClient()
			Expect(c).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error loading config file"))
		})

		It("should return an error when kubeconfig file is missing", func() {
			os.Setenv("KUBECONFIG", "/tmp/non_existing_kubeconfig_123")
			defer os.Unsetenv("KUBECONFIG")

			c, err := NewClient()
			Expect(c).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no configuration has been provided"))
		})
	})

	Describe("ListPods", func() {
		It("should return pods when list is successful", func() {
			clientset := fake.NewSimpleClientset(
				&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
			)
			kc := &Client{clientset: clientset}

			pods, err := kc.ListPods(context.Background(), "default")
			Expect(err).NotTo(HaveOccurred())
			Expect(pods.Items).To(HaveLen(1))
			Expect(pods.Items[0].Name).To(Equal("pod1"))
		})

		It("should return an error when listing pods fails", func() {
			clientset := fake.NewSimpleClientset()
			clientset.Fake.PrependReactor("list", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("failed to list pods")
			})
			kc := &Client{clientset: clientset}

			pods, err := kc.ListPods(context.Background(), "default")
			Expect(pods).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to list pods"))
		})
	})

	Describe("ListNamespaces", func() {
		It("should return namespaces when list is successful", func() {
			clientset := fake.NewSimpleClientset(
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
			)
			kc := &Client{clientset: clientset}

			nss, err := kc.ListNamespaces(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(nss.Items).To(HaveLen(1))
			Expect(nss.Items[0].Name).To(Equal("ns1"))
		})

		It("should return an error when listing namespaces fails", func() {
			clientset := fake.NewSimpleClientset()
			clientset.Fake.PrependReactor("list", "namespaces", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("failed to list namespaces")
			})
			kc := &Client{clientset: clientset}

			nss, err := kc.ListNamespaces(context.Background())
			Expect(nss).To(BeNil())
			Expect(err).To(MatchError("failed to list namespaces"))
		})
	})

	Describe("VerifyAccess", func() {
		It("should not return an error when access is allowed", func() {
			clientset := fake.NewSimpleClientset()
			clientset.Fake.PrependReactor("create", "selfsubjectaccessreviews", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &authorizationv1.SelfSubjectAccessReview{
					Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true},
				}, nil
			})
			kc := &Client{clientset: clientset}

			allowed, err := kc.VerifyAccess(context.Background(), getListPodsSAR("default"))
			Expect(err).NotTo(HaveOccurred())
			Expect(allowed).To(BeTrue())
		})

		It("should return an error when access is denied", func() {
			clientset := fake.NewSimpleClientset()
			clientset.Fake.PrependReactor("create", "selfsubjectaccessreviews", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &authorizationv1.SelfSubjectAccessReview{
					Status: authorizationv1.SubjectAccessReviewStatus{Allowed: false},
				}, nil
			})
			kc := &Client{clientset: clientset}

			allowed, err := kc.VerifyAccess(context.Background(), getListPodsSAR("restricted"))
			Expect(err).NotTo(HaveOccurred())
			Expect(allowed).To(BeFalse())
		})

		It("should return an error when API request fails", func() {
			clientset := fake.NewSimpleClientset()
			clientset.Fake.PrependReactor("create", "selfsubjectaccessreviews", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("api request failed")
			})
			kc := &Client{clientset: clientset}

			allowed, err := kc.VerifyAccess(context.Background(), getListPodsSAR("default"))
			Expect(err).To(MatchError("failed to verify access permissions: api request failed"))
			Expect(allowed).To(BeFalse())
		})
	})

	Describe("GetServerUrl", func() {
		It("should return the correct server when kubeconfig has a single cluster", func() {
			path := writeKubeConfigFile(mockKubeConfigWithSingleCluster)
			defer os.Remove(path)

			os.Setenv("KUBECONFIG", path)
			defer os.Unsetenv("KUBECONFIG")

			c, err := NewClient()
			Expect(err).NotTo(HaveOccurred())

			server := c.GetServerUrl(context.Background())
			Expect(server).To(Equal("https://127.0.0.1:12345"))
		})

		It("should return the correct server when kubeconfig has multiple clusters", func() {
			kubeConfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:12345
  name: local
- cluster:
    server: https://10.0.0.2:6443
  name: remote
contexts:
- context:
    cluster: local
    user: test-user
  name: local-context
- context:
    cluster: remote
    user: remote-user
  name: remote-context
current-context: remote-context
users:
- name: test-user
  user:
    token: dummy-local
- name: remote-user
  user:
    token: dummy-remote
`
			path := writeKubeConfigFile(kubeConfig)
			defer os.Remove(path)

			os.Setenv("KUBECONFIG", path)
			defer os.Unsetenv("KUBECONFIG")

			c, err := NewClient()
			Expect(err).NotTo(HaveOccurred())

			server := c.GetServerUrl(context.Background())
			Expect(server).To(Equal("https://10.0.0.2:6443"))
		})
	})

	Describe("CheckConnectivity", func() {
		It("should not return an error when connectivity check succeeds", func() {
			clientset := fake.NewSimpleClientset()
			c := &Client{clientset: clientset}

			err := c.CheckConnectivity(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func writeKubeConfigFile(content string) string {
	f, err := os.CreateTemp("", "kubeconfig-*")
	Expect(err).NotTo(HaveOccurred())
	defer f.Close()

	_, err = f.WriteString(content)
	Expect(err).NotTo(HaveOccurred())
	return f.Name()
}

func getListPodsSAR(namespace string) *authorizationv1.SelfSubjectAccessReview {
	return &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "list",
				Group:     "",
				Version:   "v1",
				Resource:  "pods",
			},
		},
	}
}
