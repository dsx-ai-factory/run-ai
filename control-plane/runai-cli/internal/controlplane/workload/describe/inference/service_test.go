/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	internalApi "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cpworkloadmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/golang/mock/gomock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Inference Workload Tests", func() {
	const (
		podLimit       = int32(10)
		eventLimit     = int32(20)
		testWorkloadId = "test-workload-id"
	)
	var (
		mockCtrl             *gomock.Controller
		mockWorkloadClient   *cpworkloadmocks.MockWorkloadClient
		MockWorkloadV2Getter *cpworkloadmocks.MockWorkloadV2Getter
		mockPodsClient       *cpworkloadmocks.MockPodsClient
		mockEventClient      *cpworkloadmocks.MockEventClient
		service              *describe.Service
		mockPods             []openapi.Pod
		mockEvents           []openapi.Event
	)
	var ctx = context.Background()

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockWorkloadClient = cpworkloadmocks.NewMockWorkloadClient(mockCtrl)
		MockWorkloadV2Getter = cpworkloadmocks.NewMockWorkloadV2Getter(mockCtrl)
		mockPodsClient = cpworkloadmocks.NewMockPodsClient(mockCtrl)
		mockEventClient = cpworkloadmocks.NewMockEventClient(mockCtrl)
		service = describe.NewService(mockWorkloadClient, MockWorkloadV2Getter, mockPodsClient, mockEventClient)
		mockPods = []openapi.Pod{
			{
				Name:     "pod-1",
				K8sPhase: "Running",
			},
		}

		mockEvents = []openapi.Event{
			{
				CreatedAt: time.Now(),
				Message:   "Test event",
			},
		}

	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("GetWorkload for inference workloads", func() {
		It("should return an inference workload when given a valid ID", func() {
			workloadType := internalApi.Inference.String()

			reqWorkload := &api.Workload{
				ID: testWorkloadId,
			}

			mockWorkloadDetailed := &openapi.WorkloadDetailed{
				Id:             testWorkloadId,
				Name:           "test-workload",
				Type:           workloadType,
				Phase:          openapi.PHASE_CREATING,
				ProjectName:    "test-project",
				DepartmentName: "test-department",
				CreatedAt:      time.Now(),
				Command:        ptr.To("command"),
				Arguments:      ptr.To("arguments"),
				EnvironmentVariables: &map[string]string{
					"env1": "value1",
					"env2": "value2",
				},
				AdditionalFields: map[string]interface{}{
					"latestCreatedRevisionName": "revision2",
					"latestReadyRevisionName":   "revision1",
				},
				RequestedPods: &openapi.RequestedPods{
					Min: *openapi.NewNullableInt32(ptr.To(int32(1))),
					Max: *openapi.NewNullableInt32(ptr.To(int32(2))),
				},
			}

			mockInference := &openapi.Inference{
				Name:       "test-workload",
				WorkloadId: "testWorkloadId",
				Spec:       openapi.NewInferenceSpecSpec(),
			}

			autoScaling := &openapi.AutoScaling{
				MetricThresholdPercentage:   *openapi.NewNullableFloat32(ptr.To(float32(80))),
				MinReplicas:                 *openapi.NewNullableInt32(ptr.To(int32(1))),
				InitialReplicas:             *openapi.NewNullableInt32(ptr.To(int32(2))),
				MaxReplicas:                 *openapi.NewNullableInt32(ptr.To(int32(3))),
				ActivationReplicas:          *openapi.NewNullableInt32(ptr.To(int32(1))),
				ConcurrencyHardLimit:        *openapi.NewNullableInt32(ptr.To(int32(4))),
				ScaleToZeroRetentionSeconds: *openapi.NewNullableInt32(ptr.To(int32(5))),
				ScaleDownDelaySeconds:       *openapi.NewNullableInt32(ptr.To(int32(6))),
				MetricThreshold:             *openapi.NewNullableInt32(ptr.To(int32(8))),
				Metric:                      *openapi.NewNullableString(ptr.To("metric")),
			}

			servingConfiguration := &openapi.ServingConfiguration{
				InitializationTimeoutSeconds: *openapi.NewNullableInt32(ptr.To(int32(7))),
				RequestTimeoutSeconds:        *openapi.NewNullableInt32(ptr.To(int32(9))),
			}

			mockInference.Spec.Autoscaling = *openapi.NewNullableAutoScaling(autoScaling)
			mockInference.Spec.ServingConfiguration = *openapi.NewNullableServingConfiguration(servingConfiguration)
			mockInference.Spec.Annotations = []openapi.Annotation{
				{
					Name:  *openapi.NewNullableString(ptr.To("runai-submit-command")),
					Value: *openapi.NewNullableString(ptr.To("runai bla submit --image=ubuntu")),
				},
			}
			mockWorkloadClient.EXPECT().GetWorkload(gomock.Any(), gomock.Any()).Return(mockWorkloadDetailed, nil)
			mockWorkloadClient.EXPECT().GetInference(gomock.Any(), gomock.Any()).Return(mockInference, nil)
			mockPodsClient.EXPECT().ListByLimit(gomock.Any(), gomock.Any(), podLimit).Return(mockPods, nil)
			mockEventClient.EXPECT().ListByLimit(gomock.Any(), gomock.Any(), eventLimit).Return(mockEvents, nil)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)

			Expect(err).To(BeNil())
			Expect(result).NotTo(BeNil())

			// general fields
			Expect(result.General.Name).To(Equal("test-workload"))
			Expect(result.General.Type).To(Equal(workloadType))
			Expect(result.General.Phase).To(Equal(displayer.Creating))
			Expect(result.General.Project).To(Equal("test-project"))
			Expect(result.General.Department).To(Equal("test-department"))
			Expect(result.PodsInfo.Pods).To(HaveLen(1))
			Expect(result.Events).To(HaveLen(1))

			// general fields from get workload
			Expect(result.General.Command).To(Equal(ptr.To("command")))
			Expect(result.General.Arguments).To(Equal(ptr.To("arguments")))
			Expect(result.General.EnvironmentVariables).To(Equal(&map[string]string{
				"env1": "value1",
				"env2": "value2",
			}))
			Expect(result.Spec.RunaiCliCommand).To(Equal("runai bla submit --image=ubuntu"))

			// inference fields
			Expect(result.Inference.LatestReadyRevisionName).To(Equal(ptr.To("revision1")))
			Expect(result.Inference.LatestCreatedRevisionName).To(Equal(ptr.To("revision2")))
			Expect(result.Inference.MinReplicas).To(Equal(ptr.To(int32(1))))
			Expect(result.Inference.MaxReplicas).To(Equal(ptr.To(int32(3))))
			// autoScaling
			Expect(result.Inference.MetricThresholdPercentage).To(Equal(ptr.To(float32(80))))
			Expect(result.Inference.ActivationReplicas).To(Equal(ptr.To(int32(1))))
			Expect(result.Inference.InitialReplicas).To(Equal(ptr.To(int32(2))))
			Expect(result.Inference.MaxReplicas).To(Equal(ptr.To(int32(3))))
			Expect(result.Inference.ConcurrencyHardLimit).To(Equal(ptr.To(int32(4))))
			Expect(result.Inference.ScaleToZeroRetentionSeconds).To(Equal(ptr.To(int32(5))))
			Expect(result.Inference.ScaleDownDelaySeconds).To(Equal(ptr.To(int32(6))))
			Expect(result.Inference.MetricThreshold).To(Equal(ptr.To(int32(8))))
			Expect(result.Inference.Metric).To(Equal(ptr.To("metric")))
			// servingConfiguration
			Expect(result.Inference.InitializationTimeoutSeconds).To(Equal(ptr.To(int32(7))))
			Expect(result.Inference.RequestTimeoutSeconds).To(Equal(ptr.To(int32(9))))
		})

		It("should return an error when workload is not found", func() {
			workloadID := "non-existent-id"

			reqWorkload := &api.Workload{
				ID: workloadID,
			}

			mockWorkloadClient.EXPECT().GetWorkload(ctx, workloadID).Return(nil, errors.New("workload not found"))

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("workload not found"))
			Expect(result).To(BeNil())
		})
	})

	Context("Spec Generation for inference workloads", func() {
		const (
			labelName1  = "label-name-1"
			labelValue1 = "label-value-1"
			labelName2  = "label-name-2"
			labelValue2 = "label-value-2"

			workloadName = "test-workload"
			projectName  = "test-project"
		)

		var (
			mockWorkloadDetailed *openapi.WorkloadDetailed
			reqWorkload          *api.Workload
			mockWorkloadInList   *openapi.Workload
		)

		BeforeEach(func() {
			mockWorkloadDetailed = &openapi.WorkloadDetailed{
				Id:          testWorkloadId,
				Name:        workloadName,
				Phase:       openapi.PHASE_RUNNING,
				ProjectName: projectName,
				CreatedAt:   time.Now(),
			}

			reqWorkload = &api.Workload{
				ClusterID: "test-cluster",
				Project:   projectName,
				Name:      workloadName,
			}

			mockWorkloadInList = &openapi.Workload{
				Id:          testWorkloadId,
				Name:        workloadName,
				Phase:       openapi.PHASE_RUNNING,
				ProjectName: projectName,
			}
			mockWorkloadList := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{*mockWorkloadInList},
			}

			mockWorkloadClient.EXPECT().GetWorkload(gomock.Any(), gomock.Any()).Return(mockWorkloadDetailed, nil)
			mockWorkloadClient.EXPECT().ListSlim(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(client.ToSlimResponse(mockWorkloadList), nil)
			mockPodsClient.EXPECT().ListByLimit(ctx, testWorkloadId, podLimit).Return(mockPods, nil)
			mockEventClient.EXPECT().ListByLimit(ctx, testWorkloadId, eventLimit).Return(mockEvents, nil)
		})

		It("Inference workload", func() {
			mockWorkloadDetailed.Type = internalApi.Inference.String()
			reqWorkload.Type = internalApi.Inference
			mockWorkloadInList.Type = internalApi.Inference.String()

			mockInference := &openapi.Inference{
				Name:       workloadName,
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewInferenceSpecSpec(),
			}
			mockInference.Spec.Labels = []openapi.Label{
				createApiLabel(labelName1, labelValue1),
				createApiLabel(labelName2, labelValue2),
			}

			mockWorkloadClient.EXPECT().GetInference(gomock.Any(), gomock.Any()).Return(mockInference, nil)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Spec).NotTo(BeNil())
			Expect(result.Spec.MasterLabels).To(BeNil())
			validateLabels(result.Spec.Labels, mockInference.Spec.Labels)
		})
	})
})

// Helper functions
func createApiLabel(name, value string) openapi.Label {
	return openapi.Label{
		Name:  *openapi.NewNullableString(ptr.To(name)),
		Value: *openapi.NewNullableString(ptr.To(value)),
	}
}

func validateLabels(specLabels []*openapi.Label, source []openapi.Label) {
	Expect(specLabels).To(HaveLen(len(source)))
	for i, sourceLabel := range source {
		specLabel := specLabels[i]
		Expect(specLabel.Name).To(Equal(sourceLabel.Name))
		Expect(specLabel.Value).To(Equal(sourceLabel.Value))
	}
}
