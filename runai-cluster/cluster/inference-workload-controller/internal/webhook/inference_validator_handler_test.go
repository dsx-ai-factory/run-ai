/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/webhook/validator"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	inferencewv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
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
	imageTest           = "test-image"
	Port8080            = 8080
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

			DescribeTable("should allow to create workload with fraction gpu", func(assetsTest AssetsTest) {
				assetsTest.currentWorkload.Spec.Gpu = field_types.NewOptionalNumberField(ptr.To[float64](0.5))
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				workloadParamsValidator := InferenceValidator(assetsTest.k8sClient)

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)

				Expect(response.Allowed).To(BeTrue())
			},
				Entry("Inference", AssetsTest{
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithInference().WithName(testWorkloadName).WithImage(imageTest).WithServingPort(Port8080).WithNamespace(namespace).SetActive().Build(),
				}),
			)

			DescribeTable("should allow to create workload with fraction gpu and gpu limit", func(assetsTest AssetsTest) {
				assetsTest.currentWorkload.Spec.Gpu = field_types.NewOptionalNumberField(ptr.To[float64](0.5))
				assetsTest.currentWorkload.Spec.GpuLimit = field_types.NewOptionalNumberField(ptr.To[float64](1.0))
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				workloadParamsValidator := InferenceValidator(assetsTest.k8sClient)

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)

				Expect(response.Allowed).To(BeTrue())
			},
				Entry("Inference", AssetsTest{
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithInference().WithName(testWorkloadName).WithImage(imageTest).WithServingPort(Port8080).WithNamespace(namespace).SetActive().Build(),
				}),
			)

			DescribeTable("should allow to create workload with integer gpu", func(assetsTest AssetsTest) {
				assetsTest.currentWorkload.Spec.Gpu = field_types.NewOptionalNumberField(ptr.To[float64](1.0))
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}

				workloadParamsValidator := InferenceValidator(assetsTest.k8sClient)

				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			},
				Entry("Inference", AssetsTest{
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithInference().WithName(testWorkloadName).WithImage(imageTest).WithServingPort(Port8080).WithNamespace(namespace).SetActive().Build(),
				}),
			)

			DescribeTable("should allow create inference workload when runai job exits (same name)", func(assetsTest AssetsTest) {
				job := resourceUtils.CreateDumbRunaiJob(assetsTest.currentWorkload.Name, assetsTest.currentWorkload.Namespace, "", "", runaiJobV1.JobSpec{})
				Expect(assetsTest.k8sClient.Create(ctx, job)).To(Succeed())

				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(assetsTest.currentWorkload))
				Expect(err).ToNot(HaveOccurred())

				workloadParamsValidator := InferenceValidator(assetsTest.k8sClient)

				rawObject = runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleCreate(ctx, rawObject, assetsTest.currentWorkload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())

			},
				Entry("Inference", AssetsTest{
					k8sClient:       client2.InitFakeClientWithObjectNoStatus(scheme),
					currentWorkload: workloadUtils.NewJobWorkloadBuilder().WithInference().WithName(testWorkloadName).WithImage(imageTest).WithServingPort(Port8080).WithNamespace(namespace).SetActive().Build(),
				}),
			)
		})

		Context("Update workload", func() {
			var (
				workload                *job.JobWorkload
				workloadParamsValidator *InferenceValidatorHandler
				currentWorkload         *job.JobWorkload
				k8sClient               client.Client
			)

			k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
			workloadParamsValidator = InferenceValidator(k8sClient)

			BeforeEach(func() {
				builder := workloadUtils.NewJobWorkloadBuilder()
				currentWorkload = builder.WithInference().WithName(testWorkloadName).WithImage(imageTest).WithServingPort(Port8080).WithNamespace(namespace).SetActive().Build()
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(currentWorkload))
				Expect(err).ToNot(HaveOccurred())
				rawObject = runtime.RawExtension{Raw: dumpJson}
			})

			It("should allow update workload with no changes", func() {
				workload = currentWorkload.DeepCopy()
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())
			})
			It("should allow update workload with change in active", func() {
				workload = currentWorkload.DeepCopy()
				workload.Spec.Active = field_types.NewOptionalBoolField(ptr.To(false))
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())

				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())

			})
			It("should allow update workload with active not set", func() {
				workload = currentWorkload.DeepCopy()
				workload.Spec.Active = nil
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())
			})
			It("should allow update workload with change in metadata labels", func() {
				workload = currentWorkload.DeepCopy()
				workload.ObjectMeta.Labels = map[string]string{"test": "test"}
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})
			It("should allow update workload with change in metadata annotations", func() {
				workload = currentWorkload.DeepCopy()
				workload.ObjectMeta.Annotations = map[string]string{"test": "test"}
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				newObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObject, newObjectOld, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())

			})
			It("should allow update workload with change in nodepools or node affinity", func() {
				workload = currentWorkload.DeepCopy()
				workload.Spec.NodePools = field_types.NewOptionalStringField("new-pool-1 new-pool-2", "")
				workload.Spec.NodeAffinityRequired = &field_types.NodeSelector{NodeSelectorTerms: []field_types.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				}}
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				newObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObject, newObjectOld, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
				Expect(response.Patch).To(BeNil())

			})
			It("should deny update workload with change in metadata namespace", func() {
				workload = currentWorkload.DeepCopy()
				workload.ObjectMeta.Namespace = "new-namespace"
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
			})
			It("should deny update workload with change in metadata name", func() {
				workload = currentWorkload.DeepCopy()
				workload.ObjectMeta.Name = testNewName
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
			})
			It("should deny update workload with change in spec", func() {
				workload = currentWorkload.DeepCopy()
				workload.Name = testNewName
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
			})

			It("should allow update workload with change in spec (ServingConfigurationFields mutable fields)", func() {
				workload = currentWorkload.DeepCopy()
				workload.Spec.ServingConfigurationFields.InitializationTimeoutSeconds = field_types.NewOptionalIntField(ptr.To(100))
				workload.Spec.ServingConfigurationFields.RequestTimeoutSeconds = field_types.NewOptionalIntField(ptr.To(100))
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			// The full tests for AutoScalingFields in other files,
			// the next two tests were added just to ensure that we run the inference_params_validator on "update" as well.
			It("should allow update workload with change in spec (AutoScalingFields mutable fields)", func() {
				workload = currentWorkload.DeepCopy()
				workload.Spec.AutoScalingFields.MinScale = field_types.NewOptionalIntField(ptr.To(1))
				workload.Spec.AutoScalingFields.MaxScale = field_types.NewOptionalIntField(ptr.To(2))
				workload.Spec.AutoScalingFields.InitialScale = field_types.NewOptionalIntField(ptr.To(2))
				workload.Spec.AutoScalingFields.ActivationScale = field_types.NewOptionalIntField(ptr.To(2))
				workload.Spec.AutoScalingFields.Metric = &field_types.StringField{Value: ptr.To("latency")}
				workload.Spec.AutoScalingFields.MetricName = &field_types.StringField{Value: ptr.To("MetricName")}
				workload.Spec.AutoScalingFields.Target = field_types.NewOptionalIntField(ptr.To(1))
				workload.Spec.AutoScalingFields.TargetUtilizationPercentage = field_types.NewOptionalNumberField(ptr.To(80.0))
				workload.Spec.AutoScalingFields.ConcurrencyHardLimit = field_types.NewOptionalIntField(ptr.To(1))
				workload.Spec.AutoScalingFields.ScaleToZeroRetentionSeconds = field_types.NewOptionalIntField(ptr.To(100))
				workload.Spec.AutoScalingFields.ScaleDownDelaySeconds = field_types.NewOptionalIntField(ptr.To(100))
				workload.Spec.ServingConfigurationFields.InitializationTimeoutSeconds = field_types.NewOptionalIntField(ptr.To(100))
				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}
				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeTrue())
			})

			It("should deny update workload with change in spec (AutoScalingFields invalid mutable fields value)", func() {
				workload = currentWorkload.DeepCopy()

				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}

				// invalid value of minScale > maxScale
				workload.Spec.AutoScalingFields.MaxScale = field_types.NewOptionalIntField(ptr.To(2))
				workload.Spec.AutoScalingFields.MinScale = field_types.NewOptionalIntField(ptr.To(3))
				dumpJson, err = json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				newObjectOld := runtime.RawExtension{Raw: dumpJson}

				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, newObjectOld, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(ContainSubstring("maxScale must be greater than or equal to minScale"))
			})
			It("should deny update workload with change in spec (immutable fields value)", func() {
				workload = currentWorkload.DeepCopy()

				dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				rawObjectOld := runtime.RawExtension{Raw: dumpJson}

				// security fields are immutable
				workload.Spec.RunAsUid = field_types.NewOptionalIntField(ptr.To(1000))
				workload.Spec.AutoScalingFields.MinScale = field_types.NewOptionalIntField(ptr.To(3))
				dumpJson, err = json.Marshal(inferencewv2.NewFromJobWorkload(workload))
				Expect(err).ToNot(HaveOccurred())
				newObjectOld := runtime.RawExtension{Raw: dumpJson}

				response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, newObjectOld, workload.Kind, nonSystemUser)
				Expect(response.Allowed).To(BeFalse())
				Expect(response.Result.Message).To(ContainSubstring("cannot update /spec/runAsUid, field is immutable"))
			})

			Context("workload is run:ai control-plane workload (update)", func() {
				BeforeEach(func() {
					workload = currentWorkload.DeepCopy()
					workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}
				})

				It("should allow system service account to update the workload", func() {
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, systemUser)
					Expect(response.Allowed).To(BeTrue())
				})

				It("should not allow normal account to update the workload", func() {
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
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
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}
				})
				It("should allow update workload with change in metadata labels", func() {
					workload.ObjectMeta.Labels = map[string]string{"test": "test"}
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeTrue())
				})
				It("should allow update workload with change in metadata annotations", func() {
					workload.ObjectMeta.Annotations = map[string]string{"test": "test"}
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeTrue())
				})
				It("should allow update workload with change in spec", func() {
					workload.Name = testNewName
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeTrue())
				})
				It("should deny update workload with change workload usage", func() {
					workload.Spec.Usage = common.UsedForSubmit
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := workloadParamsValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeFalse())
				})
			})

			Context("when allowedServiceAccount is not empty", func() {
				var (
					allowedServiceAccountWorkloadValidator *InferenceValidatorHandler
					allowedServiceAccount                  = "test-namespace"
				)

				BeforeEach(func() {
					allowedServiceAccountWorkloadValidator = InferenceValidator(k8sClient, allowedServiceAccount)
				})

				It("should allow update workload when user is in allowedServiceAccount list", func() {
					workload = currentWorkload.DeepCopy()
					workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := allowedServiceAccountWorkloadValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, "test-namespace:test-sa")
					Expect(response.Allowed).To(BeTrue())
				})

				It("should allow update workload when user is system user", func() {
					workload = currentWorkload.DeepCopy()
					workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := allowedServiceAccountWorkloadValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, systemUser)
					Expect(response.Allowed).To(BeTrue())
				})

				It("should deny update workload when user is not in allowedServiceAccount list", func() {
					workload = currentWorkload.DeepCopy()
					workload.ObjectMeta.Labels = map[string]string{"run.ai/workload": "test"}
					dumpJson, err := json.Marshal(inferencewv2.NewFromJobWorkload(workload))
					Expect(err).ToNot(HaveOccurred())
					rawObjectOld := runtime.RawExtension{Raw: dumpJson}
					response := allowedServiceAccountWorkloadValidator.HandleUpdate(ctx, rawObjectOld, rawObject, workload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeFalse())
					Expect(response.Result.Message).To(ContainSubstring("not authorized"))
				})
			})

		})
	})
})

func InferenceValidator(k8sClient client.Client, allowedServiceAccounts ...string) *InferenceValidatorHandler {
	GpuValidator := job.NewWorkloadGpuValidator()
	genericValidator := workloads.NewGenericParamsValidator()
	inferenceParamsValidator := validator.NewInferenceParamsValidator()
	return NewInferenceValidator(k8sClient, allowedServiceAccounts, genericValidator, GpuValidator, inferenceParamsValidator)
}
