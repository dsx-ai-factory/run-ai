package distributed

import (
	"context"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions"
	"github.com/run-ai/runai/runai-cluster/cluster/common/runaipublic"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("job extension client", func() {
	var (
		extClient extensions.Client
		ctx       context.Context
		k8sClient client.Client
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(v1.AddToScheme(scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		ctx = context.Background()
		extClient = NewClient(k8sClient)
	})

	Context("public runai is not exists", func() {
		Context("Get", func() {
			It("should return error", func() {
				_, err := extClient.Get(ctx, Pytorch)
				Expect(err).To(HaveOccurred())
			})
		})
		Context("List", func() {
			It("should return error", func() {
				_, err := extClient.List(ctx, []extensions.Type{Pytorch})
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Context("framework does not exists", func() {

		BeforeEach(func() {
			jobStatusMap := map[string]runaiJobV1.PublicStatus{
				runaiJobV1.RunaiPublicMPI: {
					Available: ptr.To(true),
				},
				runaiJobV1.RunaiPublicTF: {
					Available: ptr.To(true),
				},
				runaiJobV1.RunaiPublicPyTorch: {
					Available: ptr.To(true),
				},
				runaiJobV1.RunaiPublicXGBoost: {
					Available: ptr.To(true),
				},
			}
			configMap := createRunaiPublicWithJobStatus(true, true, jobStatusMap)
			err := k8sClient.Create(ctx, &configMap)
			Expect(err).ToNot(HaveOccurred())
		})
		Context("Get", func() {
			It("should return error", func() {
				_, err := extClient.Get(ctx, Jax)
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, runaipublic.FrameworkNotFound)).To(BeTrue())
			})
		})
		Context("List", func() {
			It("should return error", func() {
				_, err := extClient.List(ctx, []extensions.Type{Jax, Pytorch})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("all the distributed framework enabled", func() {
		BeforeEach(func() {
			jobStatusMap := map[string]runaiJobV1.PublicStatus{
				runaiJobV1.RunaiPublicMPI: {
					Available: ptr.To(true),
					Reason:    "some reason",
					Components: map[string]runaiJobV1.PublicStatus{
						runaiJobV1.DistributedFrameworkSuspend: {
							Available: ptr.To(true),
						},
					},
				},
				runaiJobV1.RunaiPublicTF: {
					Available: ptr.To(true),
					Components: map[string]runaiJobV1.PublicStatus{
						runaiJobV1.DistributedFrameworkSuspend: {
							Available: ptr.To(true),
						},
					},
				},
				runaiJobV1.RunaiPublicPyTorch: {
					Available: ptr.To(true),
					Components: map[string]runaiJobV1.PublicStatus{
						runaiJobV1.DistributedFrameworkSuspend: {
							Available: ptr.To(true),
						},
					},
				},
				runaiJobV1.RunaiPublicXGBoost: {
					Available: ptr.To(true),
					Components: map[string]runaiJobV1.PublicStatus{
						runaiJobV1.DistributedFrameworkSuspend: {
							Available: ptr.To(true),
						},
					},
				},
				runaiJobV1.RunaiPublicJax: {
					Available: ptr.To(true),
					Components: map[string]runaiJobV1.PublicStatus{
						runaiJobV1.DistributedFrameworkSuspend: {
							Available: ptr.To(true),
						},
					},
				},
			}
			configMap := createRunaiPublicWithJobStatus(true, true, jobStatusMap)
			err := k8sClient.Create(ctx, &configMap)
			Expect(err).ToNot(HaveOccurred())

		})
		Context("Get", func() {
			DescribeTable("should validate suspension is enabled for all frameworks",
				func(framework string, expectedAvailable bool) {
					extension, err := extClient.Get(ctx, extensions.Type(framework))
					Expect(err).ToNot(HaveOccurred())
					Expect(extension.Status.Enabled).To(BeTrue())
					Expect(extension.Status.Capabilities[Suspension]).To(BeTrue())
				},
				Entry("MPI", runaiJobV1.RunaiPublicMPI, true),
				Entry("TensorFlow", runaiJobV1.RunaiPublicTF, true),
				Entry("PyTorch", runaiJobV1.RunaiPublicPyTorch, true),
				Entry("XGBoost", runaiJobV1.RunaiPublicXGBoost, true),
				Entry("Jax", runaiJobV1.RunaiPublicJax, true),
			)
		})
		Context("List", func() {
			It("should return all the frameworks", func() {
				group, err := extClient.List(ctx, []extensions.Type{Jax, Pytorch, XGBoost, MPI, Tensorflow})
				Expect(err).ToNot(HaveOccurred())
				for _, framework := range []extensions.Type{Jax, Pytorch, XGBoost, MPI, Tensorflow} {
					Expect(group[framework].Status.Enabled).To(BeTrue())
					Expect(group[framework].Status.Capabilities[Suspension]).To(BeTrue())
				}
			})
			It("should return subset", func() {
				group, err := extClient.List(ctx, []extensions.Type{Jax, Pytorch})
				Expect(err).ToNot(HaveOccurred())
				for _, framework := range []extensions.Type{Jax, Pytorch} {
					Expect(group[framework].Status.Enabled).To(BeTrue())
				}
			})
		})

	})
})

func createRunaiPublicWithJobStatus(addDependencies bool, addOptional bool, statusMap map[string]runaiJobV1.PublicStatus) v1.ConfigMap {
	var publicDependencies *runaiJobV1.PublicDependencies
	var publicOptional map[string]runaiJobV1.PublicStatus
	var publicRequired map[string]runaiJobV1.PublicStatus
	if addDependencies {

		if addOptional {
			publicOptional = statusMap
		}
		publicDependencies = &runaiJobV1.PublicDependencies{
			Required: publicRequired,
			Optional: publicOptional,
		}
	}

	runaiPublic := runaiJobV1.RunaiPublic{
		Version:      "123",
		Dependencies: publicDependencies,
	}

	dumpJson, _ := yaml.Marshal(runaiPublic)

	runaiPublicConfigMap := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunaiPublicConfigMapName,
			Namespace: constants.NamespaceRunAi,
		},
		Data: map[string]string{
			constants.RunaiPublicConfigMapName: string(dumpJson),
		},
	}
	return runaiPublicConfigMap
}
