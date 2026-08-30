package validator

import (
	"context"
	"encoding/json"

	v1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/mpi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	fields "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	resourceUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	workloadUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testWorkloadName    = "test-workload"
	namespace           = "runai-default"
	runaiClusterVersion = "1.1.2"
	nonSystemUser       = "test@run.ai"
	systemUser          = "system:serviceaccount:agent"

	imageTest = "test-image"

	expectedManagedByFieldDeniedMessage = "The managedBy field is reserved for system controllers and cannot be modified manually."
)

type AssetsTest struct {
	k8sClient       client.Client
	currentWorkload *job.JobWorkload
}

var _ = Describe("Workload webhook Validator", func() {
	var (
		rawObject runtime.RawExtension
		ctx       context.Context
		scheme    *runtime.Scheme
	)
	scheme = runtime.NewScheme()
	register.InitSchema(scheme)

	Context("Nvidia Cluster", func() {

		ctx = context.Background()

		Context("Create workload", func() {

			DescribeTable("should allow/deny (gpuVendor) to create workload with fraction gpu", func(assetsTest AssetsTest) {
				assetsTest.currentWorkload.Spec.Gpu = fields.NewOptionalNumberField(ptr.To[float64](0.5))
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				workloadParamsValidator, err := WorkloadValidator(assetsTest.k8sClient)
				Expect(err).To(Succeed())
				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)

				Expect(response.Allowed).To(BeTrue())
			},
				Entry("Training", AssetsTest{
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithTraining().WithName(testWorkloadName).WithImage(imageTest).WithNamespace(namespace).SetActive().Build(),
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
				}),
			)

			DescribeTable("should allow to create workload with fraction gpu and gpu limit", func(assetsTest AssetsTest) {
				assetsTest.currentWorkload.Spec.Gpu = fields.NewOptionalNumberField(ptr.To[float64](0.5))
				assetsTest.currentWorkload.Spec.GpuLimit = fields.NewOptionalNumberField(ptr.To[float64](1.0))
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				workloadParamsValidator, err := WorkloadValidator(assetsTest.k8sClient)
				Expect(err).To(Succeed())
				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)

				Expect(response.Allowed).To(BeTrue())
			},
				Entry("Training", AssetsTest{
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithTraining().WithName(testWorkloadName).WithImage(imageTest).WithNamespace(namespace).SetActive().Build(),
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
				}),
			)

			DescribeTable("should allow/de to create workload with integer gpu", func(assetsTest AssetsTest) {
				assetsTest.currentWorkload.Spec.Gpu = fields.NewOptionalNumberField(ptr.To[float64](1.0))
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				workloadParamsValidator, err := WorkloadValidator(assetsTest.k8sClient)
				Expect(err).To(Succeed())
				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			},
				Entry("Training", AssetsTest{
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithTraining().WithName(testWorkloadName).WithImage(imageTest).WithNamespace(namespace).SetActive().Build(),
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
				}),
			)

			DescribeTable("should allow to create workload with integer gpu", func(assetsTest AssetsTest) {
				assetsTest.currentWorkload.Spec.Gpu = fields.NewOptionalNumberField(ptr.To[float64](1.0))
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				workloadParamsValidator, err := WorkloadValidator(assetsTest.k8sClient)
				Expect(err).To(Succeed())

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())

			},
				Entry("Training", AssetsTest{
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithTraining().WithName(testWorkloadName).WithImage(imageTest).WithNamespace(namespace).SetActive().Build(),
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
				}),
			)

			DescribeTable("should deny create workload when runai job exits (same name)", func(assetsTest AssetsTest) {
				job := resourceUtils.CreateDumbRunaiJob(assetsTest.currentWorkload.Name, assetsTest.currentWorkload.Namespace, "", "", runaiJobV1.JobSpec{})
				Expect(assetsTest.k8sClient.Create(ctx, job)).To(Succeed())

				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())

				workloadParamsValidator, err := WorkloadValidator(assetsTest.k8sClient)
				Expect(err).To(Succeed())

				rawObject = runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())

			},
				Entry("Training", AssetsTest{
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithTraining().WithName(testWorkloadName).WithImage(imageTest).WithNamespace(namespace).WithImage(imageTest).SetActive().Build(),
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
				}),
			)

			DescribeTable("should block MPI with master GPU", func(assetsTest AssetsTest) {
				dumpJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}
				workloadParamsValidator, err := WorkloadValidator(assetsTest.k8sClient)
				Expect(err).To(Succeed())
				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)
				if assetsTest.currentWorkload.MasterSpec != nil {
					Expect(response.Allowed).To(BeFalse())
					Expect(response.Result.Message).To(Equal(mpi.ErrorMpiMasterNoGPUAllowed.Error()))
				} else {
					Expect(response.Allowed).To(BeTrue())
				}

			},
				Entry("MPI with master GPU", AssetsTest{
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().
						WithJobType(v1.MPIJobKind).
						WithName(testWorkloadName).
						WithImage(imageTest).
						WithMasterSpec(&workloadUtils.NewJobWorkloadBuilder().WithGpuCount(1).Build().Spec).
						WithNamespace(namespace).SetActive().Build(),
					k8sClient: client2.InitFakeClientWithObjectNoStatus(scheme),
				}),
				Entry("MPI without GPU", AssetsTest{
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().
						WithJobType(v1.MPIJobKind).
						WithName(testWorkloadName).
						WithImage(imageTest).
						WithNamespace(namespace).SetActive().Build(),
					k8sClient: client2.InitFakeClientWithObjectNoStatus(scheme),
				}),
			)
		})

		Context("Update workload", func() {
			var (
				workload                *job.JobWorkload
				workloadParamsValidator *WorkloadValidatorHandler
				currentWorkload         *job.JobWorkload
			)

			BeforeEach(func() {
				builder := workloadUtils.NewJobWorkloadBuilder()
				currentWorkload = builder.WithTraining().WithName(testWorkloadName).WithImage(imageTest).WithNamespace(namespace).SetActive().Build()
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}
				workloadParamsValidator = NewWorkloadParamsValidator(client2.InitFakeClientWithObjectNoStatus(scheme), []string{})
			})

			It("should allow update workload with no changes", func() {
				workload = currentWorkload.DeepCopy()
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())
			})
			It("should allow update workload with change in active", func() {
				workload = currentWorkload.DeepCopy()
				workload.Spec.Active = fields.NewOptionalBoolField(ptr.To(false))
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())

				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())

			})
			It("should allow update workload with active not set", func() {
				workload = currentWorkload.DeepCopy()
				workload.Spec.Active = nil
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())
			})
			It("should allow update workload with change in metadata labels", func() {
				workload = currentWorkload.DeepCopy()
				workload.ObjectMeta.Labels = map[string]string{"test": "test"}
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})
			It("should allow update workload with change in metadata annotations", func() {
				workload = currentWorkload.DeepCopy()
				workload.ObjectMeta.Annotations = map[string]string{"test": "test"}
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				newObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObject, newObjectOld, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())

			})
			It("should denied update workload with change in metadata namespace", func() {
				workload = currentWorkload.DeepCopy()
				workload.ObjectMeta.Namespace = "new-namespace"
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
			})
			It("should denied update workload with change in metadata name", func() {
				workload = currentWorkload.DeepCopy()
				workload.ObjectMeta.Name = "new-name"
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
			})
			It("should denied update workload with change topology annotations", func() {
				currentWorkload.Annotations = map[string]string{"kai.scheduler/whatever": "name1"}
				workload := currentWorkload.DeepCopy()
				dumpJson, err := json.Marshal(workload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}

				currentWorkload.Annotations["kai.scheduler/whatever"] = "name2"
				dumpJson, err = json.Marshal(currentWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObjectNew := runtime.RawExtension{Raw: dumpJson}

				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObjectNew, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
			})
			It("should denied update workload with change in spec", func() {
				workload = currentWorkload.DeepCopy()
				workload.Spec.Name = fields.NewStringField("new-name")
				dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
			})
			Context("workload is run:ai control-plane workload (update)", func() {
				BeforeEach(func() {
					workload = currentWorkload.DeepCopy()
					workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}
				})

				It("should allow system service account to update the workload", func() {
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, systemUser)
					Expect(response.Allowed).To(BeTrue())
				})

				It("should not allow normal account to update the workload", func() {
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeFalse())
					Expect(response.Result.Message).To(ContainSubstring("not authorized"))
				})
			})

			Context("workload is used as template", func() {
				BeforeEach(func() {
					currentWorkload.Spec.Usage = common.UsedAsTemplate
					workload = currentWorkload.DeepCopy()
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}
				})
				It("should allow update workload with change in metadata labels", func() {
					workload.ObjectMeta.Labels = map[string]string{"test": "test"}
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeTrue())
				})
				It("should allow update workload with change in metadata annotations", func() {
					workload.ObjectMeta.Annotations = map[string]string{"test": "test"}
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeTrue())
				})
				It("should allow update workload with change in spec", func() {
					workload.Spec.Name = fields.NewStringField("new-name")
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeTrue())
				})
				It("should denied update workload with change workload usage ", func() {
					workload.Spec.Usage = common.UsedForSubmit
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeFalse())
				})
			})

			Context("with allowed service accounts", func() {
				var (
					allowedServiceAccount = "allowed-namespace:allowed-sa"
				)

				BeforeEach(func() {
					workload = currentWorkload.DeepCopy()
					workloadParamsValidator = NewWorkloadParamsValidator(client2.InitFakeClientWithObjectNoStatus(scheme), []string{allowedServiceAccount})
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}
				})

				It("should allow system service account to update the workload", func() {
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, systemUser)
					Expect(response.Allowed).To(BeTrue())
				})

				It("should allow explicitly allowed service account to update the workload", func() {
					workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, allowedServiceAccount)
					Expect(response.Allowed).To(BeTrue())
				})

				It("should deny regular user to update the workload", func() {
					workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeFalse())
					Expect(response.Result.Message).To(ContainSubstring("not authorized"))
				})
				It("should allow explicitly system service account to update the workload", func() {
					workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
					dumpJson, err := json.Marshal(twv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, systemUser)
					Expect(response.Allowed).To(BeTrue())
				})
			})

		})
	})

	Context("managedBy field protection", func() {
		var (
			workloadParamsValidator *WorkloadValidatorHandler
		)

		ctx = context.Background()

		BeforeEach(func() {
			workloadParamsValidator = NewWorkloadParamsValidator(client2.InitFakeClientWithObjectNoStatus(scheme), []string{})
		})

		Context("Create", func() {
			It("should deny user setting managedBy", func() {
				w := workloadUtils.NewJobWorkloadBuilder().
					WithJobType(v1.MPIJobKind).
					WithName(testWorkloadName).
					WithImage(imageTest).
					WithNamespace(namespace).
					SetActive().
					Build()
				w.Spec.ManagedBy = workloads.RestartControllerManagedBy
				dumpJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(w))
				Expect(err).ToNot(HaveOccurred())
				rawObj := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleCreate(ctx, rawObj, constants.KindDistributedWorkload, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal(expectedManagedByFieldDeniedMessage))
			})

			It("should allow controller to set managedBy", func() {
				w := workloadUtils.NewJobWorkloadBuilder().
					WithJobType(v1.MPIJobKind).
					WithName(testWorkloadName).
					WithImage(imageTest).
					WithNamespace(namespace).
					SetActive().
					Build()
				w.Spec.ManagedBy = workloads.RestartControllerManagedBy
				dumpJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(w))
				Expect(err).ToNot(HaveOccurred())
				rawObj := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleCreate(ctx, rawObj, constants.KindDistributedWorkload, systemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("should allow user to create with empty managedBy", func() {
				w := workloadUtils.NewJobWorkloadBuilder().
					WithJobType(v1.MPIJobKind).
					WithName(testWorkloadName).
					WithImage(imageTest).
					WithNamespace(namespace).
					SetActive().
					Build()
				dumpJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(w))
				Expect(err).ToNot(HaveOccurred())
				rawObj := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleCreate(ctx, rawObj, constants.KindDistributedWorkload, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})
		})

		Context("Update", func() {
			It("should deny user changing managedBy", func() {
				oldW := workloadUtils.NewJobWorkloadBuilder().
					WithJobType(v1.MPIJobKind).
					WithName(testWorkloadName).
					WithImage(imageTest).
					WithNamespace(namespace).
					SetActive().
					Build()
				newW := oldW.DeepCopy()
				newW.Spec.ManagedBy = workloads.RestartControllerManagedBy

				oldJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(oldW))
				Expect(err).ToNot(HaveOccurred())
				newJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(newW))
				Expect(err).ToNot(HaveOccurred())

				response := workloadParamsValidator.HandleUpdate(ctx,
					runtime.RawExtension{Raw: oldJson},
					runtime.RawExtension{Raw: newJson},
					constants.KindDistributedWorkload, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal(expectedManagedByFieldDeniedMessage))
			})

			It("should allow controller to change managedBy", func() {
				oldW := workloadUtils.NewJobWorkloadBuilder().
					WithJobType(v1.MPIJobKind).
					WithName(testWorkloadName).
					WithImage(imageTest).
					WithNamespace(namespace).
					SetActive().
					Build()
				newW := oldW.DeepCopy()
				newW.Spec.ManagedBy = workloads.RestartControllerManagedBy
				if newW.Labels == nil {
					newW.Labels = map[string]string{}
				}
				newW.Labels[workloads.RestartRequestUIDLabel] = "uid-123"

				oldJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(oldW))
				Expect(err).ToNot(HaveOccurred())
				newJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(newW))
				Expect(err).ToNot(HaveOccurred())

				response := workloadParamsValidator.HandleUpdate(ctx,
					runtime.RawExtension{Raw: oldJson},
					runtime.RawExtension{Raw: newJson},
					constants.KindDistributedWorkload, systemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("should allow spec change when managedBy is set (relaxed immutability)", func() {
				oldW := workloadUtils.NewJobWorkloadBuilder().
					WithJobType(v1.MPIJobKind).
					WithName(testWorkloadName).
					WithImage(imageTest).
					WithNamespace(namespace).
					SetActive().
					Build()
				oldW.Spec.ManagedBy = workloads.RestartControllerManagedBy
				oldW.Labels = map[string]string{workloads.RestartRequestUIDLabel: "uid-123"}

				newW := oldW.DeepCopy()
				newW.Spec.Image = fields.NewStringField("new-image")

				oldJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(oldW))
				Expect(err).ToNot(HaveOccurred())
				newJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(newW))
				Expect(err).ToNot(HaveOccurred())

				response := workloadParamsValidator.HandleUpdate(ctx,
					runtime.RawExtension{Raw: oldJson},
					runtime.RawExtension{Raw: newJson},
					constants.KindDistributedWorkload, systemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("should deny spec change when managedBy is empty (strict immutability)", func() {
				oldW := workloadUtils.NewJobWorkloadBuilder().
					WithJobType(v1.MPIJobKind).
					WithName(testWorkloadName).
					WithImage(imageTest).
					WithNamespace(namespace).
					SetActive().
					Build()
				newW := oldW.DeepCopy()
				newW.Spec.Image = fields.NewStringField("new-image")

				oldJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(oldW))
				Expect(err).ToNot(HaveOccurred())
				newJson, err := json.Marshal(v2alpha1.NewFromJobWorkload(newW))
				Expect(err).ToNot(HaveOccurred())

				response := workloadParamsValidator.HandleUpdate(ctx,
					runtime.RawExtension{Raw: oldJson},
					runtime.RawExtension{Raw: newJson},
					constants.KindDistributedWorkload, systemUser)
				Expect(response.Allowed).To(BeFalse())
			})
		})
	})
})

func WorkloadValidator(k8sClient client.Client) (*WorkloadValidatorHandler, error) {
	GpuValidator := job.NewWorkloadGpuValidator()
	runaiJobValidator := workloads.NewSingleRunaiJobValidator(k8sClient)
	genericValidator := workloads.NewGenericParamsValidator()
	distSpecValidator := workloads.NewDistSpecValidator()
	return NewWorkloadParamsValidator(k8sClient, []string{}, genericValidator, GpuValidator, runaiJobValidator, distSpecValidator), nil
}
