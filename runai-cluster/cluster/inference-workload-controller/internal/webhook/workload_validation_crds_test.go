package validator

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	register "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/testing_utils"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/webhook/validator"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	inferencewv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Workload webhook Validator", func() {
	var (
		currentWorkload               *job.JobWorkload
		currentDistWorkload           *dinfw.DistributedInferenceWorkload
		workloadParamsValidator       *InferenceValidatorHandler
		distributedInferenceValidator *DistributedInferenceValidatorHandler
		rawObject                     runtime.RawExtension
		ctx                           context.Context
		scheme                        *runtime.Scheme
		k8sClient                     client.Client
	)

	scheme = runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)

	})

	Context("Workload Duplicate Existence Validation", func() {
		BeforeEach(func() {
			workloadParamsValidator = NewInferenceValidator(k8sClient, nil)
			workloadBuilder := workload.NewJobWorkloadBuilder()
			currentWorkload = workloadBuilder.
				WithInference().
				WithName(testWorkloadName).
				WithNamespace(namespace).
				SetActive().Build().DeepCopy()
		})
		It("Create Workload Success", func() {
			dumpJson, err := json.Marshal(iwv2.NewFromJobWorkload(currentWorkload))
			Expect(err).ToNot(HaveOccurred())
			rawObject = runtime.RawExtension{Raw: dumpJson}

			response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
			Expect(response.Allowed).To(BeTrue())
			Expect(string(response.Result.Message)).To(Equal("Ok"))
		})
		Context("Knative validation", func() {
			BeforeEach(func() {
				knativeValidator := validator.NewKnativeValidator(k8sClient)
				workloadParamsValidator = NewInferenceValidator(k8sClient, []string{}, knativeValidator)

				workloadBuilder := workload.NewJobWorkloadBuilder()
				currentWorkload = workloadBuilder.
					WithInference().
					WithName(testWorkloadName).
					WithNamespace(namespace).
					SetActive().Build()
			})

			It("Runai Public Missing", func() {
				configMap := CreateRunaiPublicWithJobStatus(true, false, nil)
				_ = k8sClient.Create(ctx, &configMap)

				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("runai-public config, is missing \"Optional\" field"))
			})

			It("Runai Public not allow knative, with reason", func() {
				jobStatusMap := map[string]runaiJobV1.PublicStatus{
					"knative": {
						Available: ptr.To(false),
						Reason:    "controller not found in the cluster",
					},
				}

				configMap := CreateRunaiPublicWithJobStatus(true, true, jobStatusMap)
				_ = k8sClient.Create(ctx, &configMap)

				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("Knative is not available, reason: controller not found in the cluster"))
			})

			It("Runai Public not allow knative, without reason", func() {
				jobStatusMap := map[string]runaiJobV1.PublicStatus{
					"knative": {
						Available: ptr.To(false),
					},
				}

				configMap := CreateRunaiPublicWithJobStatus(true, true, jobStatusMap)
				_ = k8sClient.Create(ctx, &configMap)

				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("Knative is not available"))
			})

			It("Runai Public allow knative", func() {
				jobStatusMap := map[string]runaiJobV1.PublicStatus{
					"knative": {
						Available: ptr.To(true),
					},
				}

				configMap := CreateRunaiPublicWithJobStatus(true, true, jobStatusMap)
				_ = k8sClient.Create(ctx, &configMap)

				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("runai workload - allowed for system service account", func() {
				jobStatusMap := map[string]runaiJobV1.PublicStatus{
					"knative": {
						Available: ptr.To(true),
					},
				}

				configMap := CreateRunaiPublicWithJobStatus(true, true, jobStatusMap)
				_ = k8sClient.Create(ctx, &configMap)

				currentWorkload.Labels["run.ai/workload"] = "111222"

				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(ContainSubstring("not authorized"))

				response = workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, systemUser)
				Expect(response.Allowed).To(BeTrue())
			})

		})

		Context("LWS Validation", func() {
			BeforeEach(func() {
				lwsValidator := validator.NewLwsValidator(k8sClient)
				distributedInferenceValidator = NewDistributedInferenceValidator(k8sClient, []string{}, false, lwsValidator)

				currentDistWorkload = &dinfw.DistributedInferenceWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testWorkloadName,
						Namespace: namespace,
					},
				}
			})

			It("Runai Public Missing", func() {
				configMap := CreateRunaiPublicWithJobStatus(true, false, nil)
				_ = k8sClient.Create(ctx, &configMap)

				dumpJson, err := json.Marshal(currentDistWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := distributedInferenceValidator.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("runai-public config, is missing \"Optional\" field"))
			})

			It("Runai Public not allow LWS, with reason", func() {
				jobStatusMap := map[string]runaiJobV1.PublicStatus{
					"lws": {
						Available: ptr.To(false),
						Reason:    "controller not found in the cluster",
					},
				}

				configMap := CreateRunaiPublicWithJobStatus(true, true, jobStatusMap)
				_ = k8sClient.Create(ctx, &configMap)

				dumpJson, err := json.Marshal(currentDistWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := distributedInferenceValidator.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("LWS is not available, reason: controller not found in the cluster"))
			})

			It("Runai Public not allow LWS, without reason", func() {
				jobStatusMap := map[string]runaiJobV1.PublicStatus{
					"lws": {
						Available: ptr.To(false),
					},
				}

				configMap := CreateRunaiPublicWithJobStatus(true, true, jobStatusMap)
				_ = k8sClient.Create(ctx, &configMap)

				dumpJson, err := json.Marshal(currentDistWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := distributedInferenceValidator.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(Equal("LWS is not available"))
			})

			It("Runai Public allow LWS", func() {
				jobStatusMap := map[string]runaiJobV1.PublicStatus{
					"lws": {
						Available: ptr.To(true),
					},
				}

				configMap := CreateRunaiPublicWithJobStatus(true, true, jobStatusMap)
				_ = k8sClient.Create(ctx, &configMap)

				dumpJson, err := json.Marshal(currentDistWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := distributedInferenceValidator.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("runai workload - allowed for system service account", func() {
				jobStatusMap := map[string]runaiJobV1.PublicStatus{
					"lws": {
						Available: ptr.To(true),
					},
				}

				configMap := CreateRunaiPublicWithJobStatus(true, true, jobStatusMap)
				_ = k8sClient.Create(ctx, &configMap)

				if currentDistWorkload.Labels == nil {
					currentDistWorkload.Labels = make(map[string]string)
				}
				currentDistWorkload.Labels["run.ai/workload"] = "111222"

				dumpJson, err := json.Marshal(currentDistWorkload)
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				response := distributedInferenceValidator.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(ContainSubstring("not authorized"))

				response = distributedInferenceValidator.HandleCreate(ctx, rawObject, constants.KindDistributedInferenceWorkload, systemUser)
				Expect(response.Allowed).To(BeTrue())
			})

		})
	})
})

func CreateRunaiPublicWithJobStatus(addDependencies bool, addOptional bool, statusMap map[string]runaiJobV1.PublicStatus) v1.ConfigMap {
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
