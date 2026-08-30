package certificates

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("GenerateRandomString", func() {
	It("should generate a random string of length 10 (hex)", func() {
		str, err := GenerateRandomString()
		Expect(err).To(BeNil())
		Expect(len(str)).To(Equal(10))
	})
})

var _ = Describe("createOrPatchK8sTLSSecret", func() {
	var (
		ctx       context.Context
		client    *fake.Clientset
		namespace string
		cert      []byte
		key       []byte
	)

	BeforeEach(func() {
		ctx = context.Background()
		client = fake.NewSimpleClientset()
		namespace = "test-ns"
		cert = []byte("cert-bytes")
		key = []byte("key-bytes")
	})

	Context("when secret does not exist", func() {
		It("should create a new secret", func() {
			secretName := "new-secret"
			err := createOrPatchK8sTLSSecret(ctx, client, secretName, namespace, true, cert, key)
			Expect(err).To(BeNil())
			secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(secret.Data["tls.crt"]).To(Equal(cert))
			Expect(secret.Data["tls.key"]).To(Equal(key))
		})
	})

	Context("when secret already exists", func() {
		secretName := "existing-secret"
		BeforeEach(func() {
			_, err := client.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"tls.crt": []byte("old-cert"),
					"tls.key": []byte("old-key"),
				},
			}, metav1.CreateOptions{})
			Expect(err).To(BeNil())
		})

		It("should patch the existing secret", func() {
			err := createOrPatchK8sTLSSecret(ctx, client, secretName, namespace, false, cert, key)
			Expect(err).To(BeNil())
			secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(secret.Data["tls.crt"]).To(Equal(cert))
			Expect(secret.Data["tls.key"]).To(Equal(key))
		})
	})

	Context("when Get returns an error other than NotFound", func() {
		It("should return the error", func() {
			secretName := "existing-secret"
			badClient := fake.NewSimpleClientset()
			badClient.Fake.PrependReactor("get", "secrets", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, k8serrors.NewBadRequest("bad request")
			})
			err := createOrPatchK8sTLSSecret(ctx, badClient, secretName, namespace, true, cert, key)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("when createIfNotExists is false and secret does not exist", func() {
		It("should return an error indicating the secret is expected to exist", func() {
			secretName := "non-existent-secret"
			err := createOrPatchK8sTLSSecret(ctx, client, secretName, namespace, false, cert, key)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("the secret is expected to exist"))
		})
	})
})
