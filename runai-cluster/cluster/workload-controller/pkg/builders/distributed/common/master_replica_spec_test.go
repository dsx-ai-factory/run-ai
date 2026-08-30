package common_test

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	fakeClient "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/runtime"

	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const projectName = "test-project"

var _ = Describe("DistributedBuilderMaster", func() {
	var (
		builder    *common.DistributedBuilder
		masterSpec *kubeflowcommon.ReplicaSpec
		ctx        context.Context
	)

	BeforeEach(func() {
		builder = &common.DistributedBuilder{}
		builder.MergedWorkload = &dwv2.DistributedWorkload{}
		builder.ObjectNaming = &metav1.ObjectMeta{
			Name:      "test",
			Namespace: "runai" + projectName,
		}
		masterSpec = &kubeflowcommon.ReplicaSpec{}
		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)
	})

	Context("BuildMasterReplicaSpec", func() {
		When("master spec is nil", func() {
			It("returns an error", func() {
				builder.MergedWorkload.MasterSpec = nil

				err := builder.BuildMasterReplicaSpec(context.Background(), masterSpec, 0)
				Expect(err).To(MatchError("master spec is nil, but NoMaster is not set"))
			})
		})

		When("master spec is valid", func() {
			BeforeEach(func() {
				builder.MergedWorkload.MasterSpec = &v2alpha1.DistributedMasterSpec{}
			})

			It("sets replicas to one", func() {
				err := builder.BuildMasterReplicaSpec(ctx, masterSpec, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(*masterSpec.Replicas).To(Equal(int32(1)))
			})

			DescribeTable("default restart policy",
				func(jobType string, expectedPolicy kubeflowcommon.RestartPolicy) {
					builder.MergedWorkload.Spec.JobType = jobType

					err := builder.BuildMasterReplicaSpec(ctx, masterSpec, 0)
					Expect(err).NotTo(HaveOccurred())
					Expect(masterSpec.RestartPolicy).To(Equal(expectedPolicy))
				},
				// empty policy == framework default in training/mpi-operator
				Entry("TF", constants.KindTFJob, kubeflowcommon.RestartPolicy("")),
				Entry("MPI", constants.KindMpiJob, kubeflowcommon.RestartPolicy("")),
				Entry("PyTorch", constants.KindPyTorchJob, kubeflowcommon.RestartPolicy("")),
				Entry("XGBoost", constants.KindXGBoostJob, kubeflowcommon.RestartPolicy("")),
			)

			// Add more tests here for other scenarios
		})

		When("master spec has readinessProbe", func() {
			BeforeEach(func() {
				builder.MergedWorkload.MasterSpec = &v2alpha1.DistributedMasterSpec{}
				builder.MergedWorkload.MasterSpec.ReadinessProbe = &flds.ProbeField{
					Values: &flds.ProbeMembers{
						HttpGetPort: *flds.NewOptionalIntField(pointer.Int(8080)),
					},
				}
			})

			It("sets readinessProbe", func() {
				err := builder.BuildMasterReplicaSpec(ctx, masterSpec, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(masterSpec.Template.Spec.Containers[0].ReadinessProbe).NotTo(BeNil())
				Expect(masterSpec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet).NotTo(BeNil())
				Expect(masterSpec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal).To(Equal(int32(8080)))
			})
		})

		Context("Nodepools annotation", func() {
			const (
				nodepoolsList = "pool-a pool-b"
				nodepool      = "pool-c"
			)

			BeforeEach(func() {
				scheme := runtime.NewScheme()
				register.InitSchema(scheme)
				nodepoolA := resource.CreateNodePool("pool-a", "pool", "a")
				nodepoolB := resource.CreateNodePool("pool-b", "pool", "b")
				nodepoolC := resource.CreateNodePool("pool-c", "pool", "c")
				// TODO: Drop legacy nodepool once clusterconstants.UseKaiNodePools=true
				nodepoolAKai := resource.CreateKaiNodePool("pool-a", "pool", "a")
				nodepoolBKai := resource.CreateKaiNodePool("pool-b", "pool", "b")
				nodepoolCKai := resource.CreateKaiNodePool("pool-c", "pool", "c")

				builder.K8sClient = fakeClient.InitFakeClientWithObjectNoStatus(scheme, nodepoolA, nodepoolB, nodepoolC, nodepoolAKai, nodepoolBKai, nodepoolCKai)
				builder.MergedWorkload.MasterSpec = &v2alpha1.DistributedMasterSpec{}
			})

			Context("When has no nodepools", func() {
				It("when has no nodepools", func() {
					Expect(builder.BuildMasterReplicaSpec(ctx, masterSpec, 0)).To(Succeed())

					_, exist := masterSpec.Template.Annotations[commonconstants.AnnotationNodepools]
					Expect(exist).To(BeFalse())
				})
			})

			Context("When has nodepools", func() {
				It("when has Nodepools", func() {
					builder.MergedWorkload.MasterSpec.NodePools = flds.NewOptionalStringField(nodepoolsList, "")
					Expect(builder.BuildMasterReplicaSpec(ctx, masterSpec, 0)).To(Succeed())

					value, exist := masterSpec.Template.Annotations[commonconstants.AnnotationNodepools]
					Expect(exist).To(BeTrue())
					Expect(value).To(Equal(nodepoolsList))
				})

				It("when has Nodepool", func() {
					builder.MergedWorkload.MasterSpec.NodePool = flds.NewOptionalStringField(nodepool, "")
					Expect(builder.BuildMasterReplicaSpec(ctx, masterSpec, 0)).To(Succeed())

					value, exist := masterSpec.Template.Annotations[commonconstants.AnnotationNodepools]
					Expect(exist).To(BeTrue())
					Expect(value).To(Equal(nodepool))
				})

				It("when has Nodepools and Nodepool - uses Nodepools", func() {
					builder.MergedWorkload.MasterSpec.NodePools = flds.NewOptionalStringField(nodepoolsList, "")
					builder.MergedWorkload.MasterSpec.NodePool = flds.NewOptionalStringField(nodepool, "")
					Expect(builder.BuildMasterReplicaSpec(ctx, masterSpec, 0)).To(Succeed())

					value, exist := masterSpec.Template.Annotations[commonconstants.AnnotationNodepools]
					Expect(exist).To(BeTrue())
					Expect(value).To(Equal(nodepoolsList))
				})
			})
		})
	})
})
