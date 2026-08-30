package mutators

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"
	v2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	distributedpolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	distributed "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	training "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/errors"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	workloadutils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/run-ai/runai/runai-cluster/common/constants"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	projectName    = "test"
	workloadName   = "test"
	namespaceName  = "runai-test"
	clusterVersion = "2.10000"
)

var _ = Describe("WorkloadPolicyMutationHandler", func() {
	var (
		policyApplier mocks.WorkloadPolicy
		handler       *WorkloadMutatorHandler
		scheme        *runtime.Scheme

		ctx      context.Context
		workload *jobs.JobWorkload
	)
	scheme = runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)
	})

	Context("mocking unit testing", func() {
		BeforeEach(func() {

			workload = workloadutils.NewJobWorkloadBuilder().WithName(workloadName).WithTraining().WithNamespace(namespaceName).WithLabels(nil).Build()
			policyApplier = mocks.WorkloadPolicy{}
			handler = NewWorkloadMutatorHandler(&policyApplier, func() client.Object { return workload }, clusterVersion, &WorkloadStatusUpdater{}, false, false, false)

		})
		It("Should denied when workload has invalid workload policy", func() {

			policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(errors.NewWorkloadPolicyValidationError("invalid policy"))
			rawData, _ := json.Marshal(workload)
			response := handler.HandleCreate(ctx, workload, rawData)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("invalid policy"))
		})
		It("Should error when workload policy get a error", func() {
			policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(fmt.Errorf("unknown error"))

			rawData, _ := json.Marshal(workload)
			response := handler.HandleCreate(ctx, workload, rawData)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("known error"))
		})
		It("Should allowed when training workload policy is valid", func() {
			policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)
			trainingWork := training.NewFromJobWorkload(workload)

			rawData, _ := json.Marshal(trainingWork)
			response := handler.HandleCreate(ctx, trainingWork, rawData)
			Expect(response.Allowed).To(BeTrue())
			Expect(trainingWork.Status.Phase).To(Equal(common.PhaseCreating.String()))
		})
		It("Should allowed when interactive workload policy is valid", func() {
			policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)
			interactiveWorkload := v2alpha1.NewFromJobWorkload(workload)

			rawData, _ := json.Marshal(interactiveWorkload)
			response := handler.HandleCreate(ctx, interactiveWorkload, rawData)
			Expect(response.Allowed).To(BeTrue())
			Expect(interactiveWorkload.Status.Phase).To(Equal(common.PhaseCreating.String()))
		})
		It("Should allowed when distributed workload policy is valid", func() {
			policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)
			distributedWorkload := distributed.NewFromJobWorkload(workload)

			rawData, _ := json.Marshal(distributedWorkload)
			response := handler.HandleCreate(ctx, distributedWorkload, rawData)
			Expect(response.Allowed).To(BeTrue())
			Expect(distributedWorkload.Status.Phase).To(Equal(common.PhaseCreating.String()))
		})
		//It("should allowed when inference workload policy is valid", func() {
		//	policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)
		//	inferenceWork := inference.NewFromJobWorkload(workload)
		//
		//	rawData, _ := json.Marshal(inferenceWork)
		//	response := handler.HandleCreate(ctx, inferenceWork, rawData)
		//	Expect(response.Allowed).To(BeTrue())
		//	Expect(inferenceWork.Status.Phase).To(Equal(common.PhaseCreating.String()))
		//})
		It("Should allowed when job workload casting failed to match type", func() {
			policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)

			rawData, _ := json.Marshal(workload)
			response := handler.HandleCreate(ctx, workload, rawData)
			Expect(response.Allowed).To(BeTrue())
			Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(""))
		})
		It("Should error when workload is nil", func() {

			rawData, _ := json.Marshal(workload)
			response := handler.HandleCreate(ctx, nil, rawData)
			Expect(response.Allowed).To(BeFalse())
			Expect(response.Result.Message).To(ContainSubstring("workload cannot be nil"))
		})

		It("Should be allowed when workload has control plane managed label", func() {
			workload.ObjectMeta.Labels[constants.ControlPlaneManagedLabel] = "true"
			rawData, _ := json.Marshal(workload)
			response := handler.HandleCreate(ctx, workload, rawData)
			Expect(response.Allowed).To(BeTrue())
		})

		Context("cluster version label", func() {
			var trainingWork *training.TrainingWorkload

			BeforeEach(func() {
				policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)
				trainingWork = training.NewFromJobWorkload(workload)
			})

			It("should be added on create", func() {
				rawData, _ := json.Marshal(trainingWork)
				response := handler.HandleCreate(ctx, trainingWork, rawData)
				Expect(response.Allowed).To(BeTrue())

				value, ok := trainingWork.Labels[workloadcommon.CreatedAtClusterVersionLabel]
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal(clusterVersion))
			})

			It("should not fail if labels are nil", func() {
				trainingWork.Labels = nil

				rawData, _ := json.Marshal(trainingWork)
				response := handler.HandleCreate(ctx, trainingWork, rawData)
				Expect(response.Allowed).To(BeTrue())

				_, ok := trainingWork.Labels[workloadcommon.CreatedAtClusterVersionLabel]
				Expect(ok).To(BeTrue())
			})
		})

		Context("MNNVL", func() {
			var (
				distWork *distributed.DistributedWorkload
			)
			BeforeEach(func() {
				policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)
				distWork = distributed.NewFromJobWorkload(workload)
			})
			It("MNNVL label is added to meta labels", func() {
				rawData, _ := json.Marshal(distWork)
				response := handler.HandleCreate(ctx, distWork, rawData)
				Expect(response.Allowed).To(BeTrue())
				Expect(distWork.Labels[mnnvl.FeatureLabelKey]).To(Equal(mnnvl.None))
			})
			It("denied if different values in spec and master spec", func() {
				distWork.Spec = distributedpolicy.DistributedSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							Labels: field_types.NewDictionaryField(make(map[string]string)),
						},
					},
				}
				distWork.MasterSpec = &distributedpolicy.DistributedMasterSpec{
					WorkloadFields: field_groups.WorkloadFields{
						ContainerFields: field_groups.ContainerFields{
							Labels: field_types.NewDictionaryField(make(map[string]string)),
						},
					},
				}

				distWork.Spec.Labels.Put(mnnvl.FeatureLabelKey, "a")
				distWork.MasterSpec.Labels.Put(mnnvl.FeatureLabelKey, "b")

				rawData, _ := json.Marshal(distWork)
				response := handler.HandleCreate(ctx, distWork, rawData)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Code).To(Equal(int32(http.StatusBadRequest)))
				Expect(response.Result.Message).To(ContainSubstring("providing different values"))
			})
		})

		Context("requested node affinity annotation", func() {
			var nodeSelector *field_types.NodeSelector

			BeforeEach(func() {
				policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)
				nodeSelector = &field_types.NodeSelector{
					NodeSelectorTerms: []field_types.NodeSelectorTerm{
						{MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "node-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"gpu"}},
						}},
					},
				}
			})

			It("should set annotation on training workload with node affinity", func() {
				trainingWork := training.NewFromJobWorkload(workload)
				trainingWork.Spec.NodeAffinityRequired = nodeSelector

				rawData, _ := json.Marshal(trainingWork)
				response := handler.HandleCreate(ctx, trainingWork, rawData)
				Expect(response.Allowed).To(BeTrue())

				annotationValue, ok := trainingWork.Annotations[clusterConstants.AnnotationRequestedNodeAffinity]
				Expect(ok).To(BeTrue())
				expected, _ := json.Marshal(nodeSelector)
				Expect(annotationValue).To(Equal(string(expected)))
			})

			It("should set annotation on interactive workload with node affinity", func() {
				interactiveWork := v2alpha1.NewFromJobWorkload(workload)
				interactiveWork.Spec.NodeAffinityRequired = nodeSelector

				rawData, _ := json.Marshal(interactiveWork)
				response := handler.HandleCreate(ctx, interactiveWork, rawData)
				Expect(response.Allowed).To(BeTrue())

				_, ok := interactiveWork.Annotations[clusterConstants.AnnotationRequestedNodeAffinity]
				Expect(ok).To(BeTrue())
			})

			It("should set annotation on distributed workload with node affinity", func() {
				distWork := distributed.NewFromJobWorkload(workload)
				distWork.Spec.NodeAffinityRequired = nodeSelector

				rawData, _ := json.Marshal(distWork)
				response := handler.HandleCreate(ctx, distWork, rawData)
				Expect(response.Allowed).To(BeTrue())

				_, ok := distWork.Annotations[clusterConstants.AnnotationRequestedNodeAffinity]
				Expect(ok).To(BeTrue())
			})

			It("should not set annotation when node affinity is nil", func() {
				trainingWork := training.NewFromJobWorkload(workload)
				// NodeAffinityRequired is nil by default

				rawData, _ := json.Marshal(trainingWork)
				response := handler.HandleCreate(ctx, trainingWork, rawData)
				Expect(response.Allowed).To(BeTrue())

				_, ok := trainingWork.Annotations[clusterConstants.AnnotationRequestedNodeAffinity]
				Expect(ok).To(BeFalse())
			})

			It("should not overwrite existing annotation", func() {
				trainingWork := training.NewFromJobWorkload(workload)
				trainingWork.Spec.NodeAffinityRequired = nodeSelector
				trainingWork.Annotations = map[string]string{
					clusterConstants.AnnotationRequestedNodeAffinity: "existing-value",
				}

				rawData, _ := json.Marshal(trainingWork)
				response := handler.HandleCreate(ctx, trainingWork, rawData)
				Expect(response.Allowed).To(BeTrue())

				Expect(trainingWork.Annotations[clusterConstants.AnnotationRequestedNodeAffinity]).To(Equal("existing-value"))
			})
		})

		Context("Topology", func() {
			var (
				distWork *distributed.DistributedWorkload
			)

			BeforeEach(func() {
				handler = NewWorkloadMutatorHandler(&policyApplier, func() client.Object { return workload }, clusterVersion, &WorkloadStatusUpdater{}, false, true, true)
				policyApplier.On("ApplyPolicy", mock.Anything, mock.Anything).Return(nil)
				distWork = distributed.NewFromJobWorkload(workload)
			})

			Context("valid", func() {
				It("No annotations", func() {
					rawData, _ := json.Marshal(distWork)
					response := handler.HandleCreate(ctx, distWork, rawData)
					Expect(response.Allowed).To(BeTrue())
					Expect(distWork.Annotations).To(HaveLen(0))
				})

				It("Empty annotations", func() {
					distWork.Spec.Annotations = field_types.NewDictionaryField(make(map[string]string))

					rawData, _ := json.Marshal(distWork)
					response := handler.HandleCreate(ctx, distWork, rawData)
					Expect(response.Allowed).To(BeTrue())
					Expect(distWork.Annotations).To(HaveLen(0))
				})

				It("Annotations exist in spec, moved to workload", func() {
					distWork.Spec.Annotations = field_types.NewDictionaryField(make(map[string]string))
					distWork.Spec.Annotations.Put(topology.KAITopologyAnnotationKey, "topology-name")
					distWork.Spec.Annotations.Put(topology.KAITopologyAnnotationPreferredKey, "topology-preferred")

					rawData, _ := json.Marshal(distWork)
					response := handler.HandleCreate(ctx, distWork, rawData)
					Expect(response.Allowed).To(BeTrue())
					Expect(distWork.Annotations).To(HaveLen(2))
					Expect(distWork.Annotations[topology.KAITopologyAnnotationKey]).To(Equal("topology-name"))
					Expect(distWork.Annotations[topology.KAITopologyAnnotationPreferredKey]).To(Equal("topology-preferred"))
				})

				It("Annotations exist in spec workspace - but ignored", func() {
					workload = workloadutils.NewJobWorkloadBuilder().WithName(workloadName).WithInteractive().WithNamespace(namespaceName).WithLabels(nil).Build()
					interactive := v2alpha1.NewFromJobWorkload(workload)
					interactive.Spec.Annotations = field_types.NewDictionaryField(make(map[string]string))
					interactive.Spec.Annotations.Put(topology.KAITopologyAnnotationKey, "topology-name")
					interactive.Spec.Annotations.Put(topology.KAITopologyAnnotationPreferredKey, "topology-preferred")

					rawData, _ := json.Marshal(interactive)
					response := handler.HandleCreate(ctx, interactive, rawData)
					Expect(response.Allowed).To(BeTrue())
					Expect(distWork.Annotations).To(HaveLen(0))
				})

				It("Annotations exist in spec inference - but ignored", func() {
					workload = workloadutils.NewJobWorkloadBuilder().WithName(workloadName).WithInference().WithNamespace(namespaceName).WithLabels(nil).Build()
					inference := v2alpha2.NewFromJobWorkload(workload)
					inference.Spec.Annotations = field_types.NewDictionaryField(make(map[string]string))
					inference.Spec.Annotations.Put(topology.KAITopologyAnnotationKey, "topology-name")
					inference.Spec.Annotations.Put(topology.KAITopologyAnnotationPreferredKey, "topology-preferred")

					rawData, _ := json.Marshal(inference)
					response := handler.HandleCreate(ctx, inference, rawData)
					Expect(response.Allowed).To(BeTrue())
					Expect(distWork.Annotations).To(HaveLen(0))
				})
			})
		})
	})
})

var _ = Describe("WorkloadMutatorHandler", func() {
	var (
		policyApplier mocks.WorkloadPolicy
		handler       *WorkloadMutatorHandler
		workload      *jobs.JobWorkload
	)

	BeforeEach(func() {
		workload = workloadutils.NewJobWorkloadBuilder().WithName(workloadName).WithTraining().WithNamespace(namespaceName).WithLabels(nil).Build()
		policyApplier = mocks.WorkloadPolicy{}
		handler = NewWorkloadMutatorHandler(&policyApplier, func() client.Object { return workload }, clusterVersion, &WorkloadStatusUpdater{}, false, false, false)
	})

	It("Should create a new WorkloadMutatorHandler with the provided workload policy and getWorkloadObjFn", func() {
		Expect(handler.workloadPolicy).To(Equal(&policyApplier))
		Expect(handler.getWorkloadObjFn()).To(Equal(workload))
	})

	It("Should return nil when getWorkloadObjFn is called", func() {
		handler = NewWorkloadMutatorHandler(&policyApplier, func() client.Object { return nil }, clusterVersion, &WorkloadStatusUpdater{}, false, false, false)
		Expect(handler.getWorkloadObjFn()).To(BeNil())
	})
})
