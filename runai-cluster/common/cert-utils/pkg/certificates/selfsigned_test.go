package certificates

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("selfSignedSigner", func() {
	var (
		signer *selfSignedSigner
	)

	BeforeEach(func() {
		signer = &selfSignedSigner{
			kbClient: fake.NewSimpleClientset(),
		}
	})

	Context("Sign", func() {
		Context("Happy flow", func() {
			Context("First run / no previous secret exist", func() {
				It("should generate a self signed certificate and create a tls secret from it", func() {
					certificateDetails, err := signer.Sign(context.Background(), "service", "secret", "namespace", true)
					Expect(err).To(BeNil())
					Expect(certificateDetails.CaBundle).ToNot(BeNil())

					// PEM Parse the certificate and check the common name
					block, _ := pem.Decode(certificateDetails.CaBundle)
					Expect(block).ToNot(BeNil())
					Expect(block.Type).To(Equal("CERTIFICATE"))
					Expect(block.Bytes).ToNot(BeNil())

					x509Cert, err := x509.ParseCertificate(block.Bytes)
					Expect(err).To(BeNil())
					Expect(x509Cert.Subject.CommonName).To(Equal("service.namespace.svc"))

					secret, err := signer.kbClient.CoreV1().Secrets("namespace").Get(
						context.Background(), "secret", metav1.GetOptions{})
					Expect(err).To(BeNil())
					Expect(secret.Data["tls.crt"]).To(Equal(certificateDetails.CaBundle))
					Expect(secret.Data["tls.key"]).ToNot(BeNil())
				})
			})

			Context("run after secret already exist", func() {
				BeforeEach(func() {
					secret := &v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "secret",
							Namespace: "namespace",
						},
						Data: map[string][]byte{
							"tls.crt": []byte("old-cert"),
							"tls.key": []byte("old-key"),
						},
					}

					_, err := signer.kbClient.CoreV1().Secrets("namespace").Create(
						context.Background(), secret, metav1.CreateOptions{})
					Expect(err).To(BeNil())
				})

				It("should generate a new certificate and recreate the secret", func() {
					certificateDetails, err := signer.Sign(
						context.Background(), "service", "secret", "namespace", true)
					Expect(err).To(BeNil())
					Expect(certificateDetails.CaBundle).ToNot(BeNil())

					secret, err := signer.kbClient.CoreV1().Secrets("namespace").Get(
						context.Background(), "secret", metav1.GetOptions{})
					Expect(err).To(BeNil())
					Expect(secret.Data["tls.crt"]).To(Equal(certificateDetails.CaBundle))
					Expect(secret.Data["tls.crt"]).ToNot(Equal("old-cert"))
					Expect(secret.Data["tls.key"]).ToNot(BeNil())
					Expect(secret.Data["tls.key"]).ToNot(Equal("old-key"))
				})
			})
		})
	})
})
