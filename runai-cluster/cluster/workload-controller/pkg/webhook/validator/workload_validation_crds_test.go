/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/controller"
	"k8s.io/utils/ptr"

	kfv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Workload webhook Validator", func() {
	const (
		AnyServiceName = "any-name-is-good"
	)

	var (
		currentWorkload         *job.JobWorkload
		currentWorkload2        *job.JobWorkload
		workloadParamsValidator *WorkloadValidatorHandler
		rawObject               runtime.RawExtension
		ctx                     context.Context
		scheme                  *runtime.Scheme
		k8sClient               client.Client
	)
	scheme = runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)

	})

	Context("Workload Duplicate Existence Validation", func() {
		BeforeEach(func() {
			workloadParamsValidator = NewWorkloadParamsValidator(k8sClient, []string{})
			workloadBuilder := workload.NewJobWorkloadBuilder()
			currentWorkload = workloadBuilder.
				WithInteractive().
				WithName(testWorkloadName).
				WithNamespace(namespace).
				SetActive().Build().DeepCopy()

			currentWorkload2 = workloadBuilder.
				WithTraining().
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
		It("Create Workload Already Exist", func() {
			labels := map[string]string{constants.LabelRelease: currentWorkload.Name}

			dumpJson, err := json.Marshal(iwv2.NewFromJobWorkload(currentWorkload))
			Expect(err).ToNot(HaveOccurred())
			rawObject = runtime.RawExtension{Raw: dumpJson}

			response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
			Expect(response.Allowed).To(BeTrue())
			Expect(string(response.Result.Message)).To(Equal("Ok"))

			service := CreateDumbServiceWithLabel(AnyServiceName, namespace, currentWorkload.UID, currentWorkload.Kind, labels)
			err = k8sClient.Create(ctx, service)
			Expect(err).ToNot(HaveOccurred())

			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: AnyServiceName}, service)
			Expect(err).ToNot(HaveOccurred())

			dumpJson, err = json.Marshal(twv2.NewFromJobWorkload(currentWorkload2))
			Expect(err).ToNot(HaveOccurred())
			rawObject = runtime.RawExtension{Raw: dumpJson}

			response = workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
			Expect(response.Allowed).To(BeFalse())
			Expect(string(response.Result.Message)).To(Equal("A workload with the name test-workload already exists. Please choose a different name for your workload."))
		})
	})

	Context("Distributed CRDs verification", func() {
		var (
			workloadBuilder    *workload.JobWorkloadBuilder
			frameworkValidator *workloads.WorkloadFrameworkValidator
		)

		BeforeEach(func() {
			enabledFrameworks := make(map[string]controller.DistributedJobFrameworksStatus)
			enabledFrameworks[constants.KindTFJob] = controller.DistributedJobFrameworksStatus{JobObject: &kfv1.TFJob{}, SuspendEnabled: true}
			enabledFrameworks[constants.KindPyTorchJob] = controller.DistributedJobFrameworksStatus{JobObject: &kfv1.PyTorchJob{}}
			enabledFrameworks[constants.KindXGBoostJob] = controller.DistributedJobFrameworksStatus{JobObject: &kfv1.XGBoostJob{}, SuspendEnabled: false}
			frameworkValidator = workloads.NewWorkloadFrameworkValidator(k8sClient, enabledFrameworks)
			workloadParamsValidator = NewWorkloadParamsValidator(k8sClient, []string{}, frameworkValidator)
			workloadBuilder = workload.NewJobWorkloadBuilder()
		})
		Context("Create distributed workload", func() {
			Context("TFJob installed and suspendable", func() {
				It("Create TFJob Success - Active", func() {
					currentWorkload = workloadBuilder.
						WithDistribution().
						WithName(testWorkloadName).
						WithNamespace(namespace).
						WithJobType(constants.KindTFJob).
						SetActive().Build()
					dumpJson, err := json.Marshal(dwv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}

					response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeTrue())
					Expect(response.Result.Message).To(Equal("Ok"))
				})

				It("Create TFJob Success - Suspended", func() {
					currentWorkload = workloadBuilder.
						WithDistribution().
						WithName(testWorkloadName).
						WithNamespace(namespace).
						WithJobType(constants.KindTFJob).Build()
					dumpJson, err := json.Marshal(dwv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}

					response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeTrue())
					Expect(response.Result.Message).To(Equal("Ok"))
				})
			})
			Context("MPI not installed", func() {
				It("should fail to create MPIJob", func() {
					currentWorkload = workloadBuilder.
						WithDistribution().
						WithName(testWorkloadName).
						WithNamespace(namespace).
						WithJobType(constants.KindMpiJob).
						SetActive().Build()
					dumpJson, err := json.Marshal(dwv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}

					response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeFalse())
					Expect(response.Result.Message).To(Equal("distributed job MPIJob is not available, make sure that MPIJob is installed on your cluster"))
				})
			})
			Context("XGBoost installed and not suspendable", func() {
				It("should fail to create suspended xgboost job", func() {
					currentWorkload = workloadBuilder.
						WithDistribution().
						WithName(testWorkloadName).
						WithNamespace(namespace).
						WithJobType(constants.KindXGBoostJob).Build()
					dumpJson, err := json.Marshal(dwv2.NewFromJobWorkload(currentWorkload))
					Expect(err).ToNot(HaveOccurred())
					rawObject = runtime.RawExtension{Raw: dumpJson}

					response := workloadParamsValidator.HandleCreate(ctx, rawObject, currentWorkload.Kind, nonSystemUser)
					Expect(response.Allowed).To(BeFalse())
					Expect(response.Result.Message).To(Equal(fmt.Sprintf(workloads.NotSuspendable, constants.KindXGBoostJob)))
				})
			})
		})

		Context("Update distributed workload", func() {
			Context("Update distributed workload", func() {
				Context("XGBoost installed and not suspendable", func() {
					var (
						oldRawObject runtime.RawExtension
						newRawObject runtime.RawExtension
					)
					BeforeEach(func() {
						workload := workloadBuilder.
							WithDistribution().
							WithName(testWorkloadName).
							WithNamespace(namespace).
							WithJobType(constants.KindXGBoostJob).
							SetActive().Build()
						dumpJson, err := json.Marshal(dwv2.NewFromJobWorkload(workload))
						Expect(err).ToNot(HaveOccurred())
						oldRawObject = runtime.RawExtension{Raw: dumpJson}

						workload.Spec.Active = field_types.NewOptionalBoolField(ptr.To(false))
						dumpJson, err = json.Marshal(dwv2.NewFromJobWorkload(workload))
						Expect(err).ToNot(HaveOccurred())
						newRawObject = runtime.RawExtension{Raw: dumpJson}
					})
					It("should not fail to update if update validator not registered", func() {
						response := workloadParamsValidator.HandleUpdate(ctx, oldRawObject, newRawObject, currentWorkload.Kind, nonSystemUser)
						Expect(response.Allowed).To(BeTrue())
					})
					It("should fail to update if update validator registered", func() {
						workloadParamsValidator.RegisterUpdateValidators(frameworkValidator)

						response := workloadParamsValidator.HandleUpdate(ctx, oldRawObject, newRawObject, currentWorkload.Kind, nonSystemUser)
						Expect(response.Allowed).To(BeFalse())
						Expect(response.Result.Message).To(Equal(fmt.Sprintf(workloads.NotSuspendable, constants.KindXGBoostJob)))
					})
				})
			})
		})
	})
})
