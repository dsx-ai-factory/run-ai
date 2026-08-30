package resources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	"context"

	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ConfigMaps", func() {
	var (
		k8s         rclient.Client
		runaiConfig *runv1.RunaiConfig
	)

	BeforeEach(func() {
		k8s = fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		runaiConfig = &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai",
				Namespace: "runai",
			},
		}
		runaiConfig.SetDefaultsWhereNeeded()
	})

	It("should return a configmap", func() {
		res, err := ConfigMapForRunaiConfig(context.Background(), k8s, runaiConfig)
		cm := res.(*v1.ConfigMap)
		Expect(err).ToNot(HaveOccurred())
		Expect(cm).ToNot(BeNil())
		Expect(len(cm.Annotations)).To(Equal(1))
	})
})
