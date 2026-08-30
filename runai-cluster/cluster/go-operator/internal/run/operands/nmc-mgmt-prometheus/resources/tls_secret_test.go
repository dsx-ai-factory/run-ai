package resources

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prom_common "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/shared_data"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("TLS Secret Resource", func() {
	var (
		ctx         context.Context
		k8sClient   client.Client
		runaiConfig *runv1.RunaiConfig
		sharedData  *shared_data.SharedData
		resource    *TLSSecretResource
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient = fake.NewClientBuilder().Build()
		runaiConfig = &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "test-namespace",
			},
		}
		sharedData = shared_data.NewSharedData()
		resource = NewTLSSecretResource("test-operand", k8sClient, runaiConfig, sharedData)
	})

	Context("when creating a new TLS secret", func() {
		It("should generate a valid certificate and key", func() {
			obj, err := resource.Object(ctx)
			Expect(err).NotTo(HaveOccurred())

			secret := obj.(*corev1.Secret)
			Expect(secret.Name).To(Equal(tlsSecretName))
			Expect(secret.Namespace).To(Equal(runaiConfig.Namespace))
			Expect(secret.Type).To(Equal(corev1.SecretTypeTLS))

			// Verify certificate and key exist
			Expect(secret.Data).To(HaveKey("tls.crt"))
			Expect(secret.Data).To(HaveKey("tls.key"))

			// Verify certificate is valid
			block, _ := pem.Decode(secret.Data["tls.crt"])
			Expect(block).NotTo(BeNil())

			cert, err := x509.ParseCertificate(block.Bytes)
			Expect(err).NotTo(HaveOccurred())

			// Verify certificate properties
			Expect(cert.DNSNames).To(ContainElement(prom_common.MgmtPrometheusDomain))
			Expect(cert.NotAfter).To(BeTemporally(">", time.Now().AddDate(0, 1, 0)))
		})

		It("should store the certificate in shared data", func() {
			_, err := resource.Object(ctx)
			Expect(err).NotTo(HaveOccurred())

			certData, exists := sharedData.Get(shared_data.NmcMgmtPrometheusCA)
			Expect(exists).To(BeTrue())
			Expect(certData).NotTo(BeNil())
		})
	})

	Context("when secret already exists", func() {
		var existingSecret *corev1.Secret

		BeforeEach(func() {
			existingSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tlsSecretName,
					Namespace: runaiConfig.Namespace,
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					"tls.crt": []byte("existing-cert"),
					"tls.key": []byte("existing-key"),
				},
			}
			err := k8sClient.Create(ctx, existingSecret)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not regenerate certificate if it exists and is valid", func() {
			obj, err := resource.Object(ctx)
			Expect(err).NotTo(HaveOccurred())

			secret := obj.(*corev1.Secret)
			Expect(secret.Data["tls.crt"]).To(Equal([]byte("existing-cert")))
			Expect(secret.Data["tls.key"]).To(Equal([]byte("existing-key")))
		})
	})

	Context("Status", func() {
		It("should return error if secret does not exist", func() {
			err := resource.Status(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is missing"))
		})

		It("should return nil if secret exists", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tlsSecretName,
					Namespace: runaiConfig.Namespace,
					UID:       "test-uid",
				},
			}
			err := k8sClient.Create(ctx, secret)
			Expect(err).NotTo(HaveOccurred())

			err = resource.Status(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Certificate expiration", func() {
		It("should correctly identify expired certificates", func() {
			// Generate a private key for self signing
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			Expect(err).NotTo(HaveOccurred())

			// Create certificate template
			template := &x509.Certificate{
				SerialNumber: big.NewInt(1),
				Subject: pkix.Name{
					CommonName: "test.local",
				},
				NotBefore:             time.Now().AddDate(-1, 0, 0), // 1 year ago
				NotAfter:              time.Now().AddDate(0, -1, 0), // 1 month ago
				KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
				ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
				BasicConstraintsValid: true,
				IsCA:                  true,
			}

			// Create the self-signed certificate
			derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
			Expect(err).NotTo(HaveOccurred())

			// Encode as PEM
			certPEM := pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: derBytes,
			})

			result := isCertExpiredOrExpiringSoon(certPEM)
			Expect(result).To(BeTrue())
		})

		It("should correctly identify valid certificates", func() {
			obj, err := resource.Object(ctx)
			Expect(err).NotTo(HaveOccurred())

			secret := obj.(*corev1.Secret)
			result := isCertExpiredOrExpiringSoon(secret.Data["tls.crt"])
			Expect(result).To(BeFalse())
		})
	})
})
