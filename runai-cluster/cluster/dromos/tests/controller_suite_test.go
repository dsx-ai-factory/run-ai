package controller

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1/testcerts"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

var (
	cfg                   *rest.Config
	testEnv               *envtest.Environment
	ctx                   context.Context
	cancel                context.CancelFunc
	webhookPort           int
	webhookCertDir        string
	webhookUrlValidate    string
	webhookUrlMutate      string
	validatingWebhookName = "workloadrunner-validating-webhook"
	mutatingWebhookName   = "workloadrunner-mutating-webhook"
)

// getFreePort returns a free loopback port
func getFreePort() int {
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())
	port := listener.Addr().(*net.TCPAddr).Port
	Expect(listener.Close()).To(Succeed())
	return port
}

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	// Kernel-allocated port so concurrent envtest suites never collide.
	webhookPort = getFreePort()
	webhookUrlValidate = fmt.Sprintf("https://127.0.0.1:%d/validate-workloadrunner", webhookPort)
	webhookUrlMutate = fmt.Sprintf("https://127.0.0.1:%d/mutate-workloadrunner", webhookPort)

	ctx, cancel = context.WithCancel(context.TODO())

	var certErr error
	webhookCertDir, certErr = testcerts.GenerateWebhookCertDir()
	Expect(certErr).NotTo(HaveOccurred())
	DeferCleanup(func() { _ = os.RemoveAll(webhookCertDir) })

	By("bootstrapping test environment")

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "sdk", "config", "crd", "bases"),
			filepath.Join("..", "..", "runai-crd-upgrader", "crds"),
			filepath.Join("testdata", "generated", "crds"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = dromosv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = kartav1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apiextensionsv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = jobsetv1alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = nvapi.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	setupWebhookConfigurations(webhookCertDir, ctx, k8sClient, webhookPort, validatingWebhookName, mutatingWebhookName)
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
