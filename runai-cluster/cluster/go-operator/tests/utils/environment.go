package testing_utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap/zapcore"

	appsv1 "k8s.io/api/apps/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	kcfake "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client/fake"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
)

func EnvironmentSetup(suite string) (*envtest.Environment, kc.HybridClient, context.Context) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true),
		func(o *zap.Options) {
			o.TimeEncoder = zapcore.ISO8601TimeEncoder
		}))

	By(fmt.Sprintf("bootstrapping test environment for [%s]", suite))
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:       findCRDsPaths(),
		ErrorIfCRDPathMissing:   true,
		ControlPlaneStopTimeout: 10 * TimeoutUnit,
		CRDInstallOptions: envtest.CRDInstallOptions{
			CleanUpAfterUse: true,
		},
	}

	cfg, err := testEnv.Start()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, cfg).NotTo(BeNil())

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme_init.Scheme()})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, k8sClient).NotTo(BeNil())

	hybridClient := kcfake.New(k8sClient, k8sClient, nil)

	By("creating testing namespaces")
	runaiNamespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: TestingNamespace,
			Annotations: map[string]string{
				"suite": suite,
			},
		},
	}

	ExpectCreateResource(k8sClient, runaiNamespace)

	By("waiting for dcgm-exporter to be ready")
	EventuallyWithTimeout(func() bool {
		ds := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nvidia-dcgm-exporter",
				Namespace: consts.GPUOperatorNamespace,
			},
		}
		ExpectGetResource(k8sClient, ds)

		return ds.Status.CurrentNumberScheduled == ds.Status.DesiredNumberScheduled
	}).Should(BeTrue())

	return testEnv, hybridClient, context.Background()
}

func EnvironmentTeardown(testEnv *envtest.Environment, suite string) {
	By(fmt.Sprintf("tearing down the test environment for [%s]", suite))
	k8sClient, err := client.New(testEnv.Config, client.Options{Scheme: scheme_init.Scheme()})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, k8sClient).NotTo(BeNil())

	By("deleting testing namespaces")
	testingNamespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: TestingNamespace,
		},
	}

	DeleteAndPollUntilDeleted(k8sClient, testingNamespace, 20*TimeoutUnit)

	Expect(testEnv.Stop()).To(Succeed())
}

func findCRDsPaths() []string {
	crdsBasePath := []string{"config", "crd"}

	for i := 0; i < 5; i++ {
		if _, err := os.ReadDir(filepath.Join(crdsBasePath...)); err != nil {
			if os.IsNotExist(err) {
				crdsBasePath = append([]string{".."}, crdsBasePath...)
			} else {
				panic(fmt.Errorf("could not find CRD definitions"))
			}
		} else {
			break
		}
	}

	return []string{
		filepath.Join(append(crdsBasePath, "bases")...),
	}
}
