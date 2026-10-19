/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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
	workloadmanager "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/golang/mock/gomock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Describe Service", func() {
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

	Context("GetWorkload", func() {
		It("should return a workload when given a valid ID", func() {
			workloadType := internalApi.Training.String()

			reqWorkload := &api.Workload{
				ID: testWorkloadId,
			}

			mockWorkloadDetailed := &openapi.WorkloadDetailed{
				Id:             testWorkloadId,
				Name:           "test-workload",
				Type:           workloadType,
				Phase:          openapi.PHASE_RUNNING,
				ProjectName:    "test-project",
				DepartmentName: "test-department",
				CreatedAt:      time.Now(),
				Preemptibility: *openapi.NewNullableString(ptr.To("semi-preemptible")),
			}

			mockTraining := &openapi.Training{
				Name:       "test-workload",
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewTrainingSpecSpec(),
			}

			mockWorkloadClient.EXPECT().GetWorkload(gomock.Any(), gomock.Any()).Return(mockWorkloadDetailed, nil)
			mockWorkloadClient.EXPECT().GetTraining(gomock.Any(), gomock.Any()).Return(mockTraining, nil)
			mockPodsClient.EXPECT().ListByLimit(gomock.Any(), gomock.Any(), podLimit).Return(mockPods, nil)
			mockEventClient.EXPECT().ListByLimit(gomock.Any(), gomock.Any(), eventLimit).Return(mockEvents, nil)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)

			Expect(err).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(result.General.Name).To(Equal("test-workload"))
			Expect(result.General.Type).To(Equal(workloadType))
			Expect(result.General.Phase).To(Equal(displayer.Running))
			Expect(result.General.Project).To(Equal("test-project"))
			Expect(result.General.Department).To(Equal("test-department"))
			Expect(result.General.Preemptibility).To(Equal(ptr.To("semi-preemptible")))
			Expect(result.PodsInfo.Pods).To(HaveLen(1))
			Expect(result.Events).To(HaveLen(1))
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

		It("should return a workload when given valid filter criteria", func() {
			workloadType := internalApi.Training.String()

			reqWorkload := &api.Workload{
				ClusterID: "test-cluster",
				Project:   "test-project",
				Name:      "test-workload",
				Type:      internalApi.Training,
			}

			mockWorkloadList := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{
					{
						Id:             testWorkloadId,
						Name:           "test-workload",
						Type:           workloadType,
						Phase:          openapi.PHASE_RUNNING,
						ProjectName:    "test-project",
						DepartmentName: "test-department",
						CreatedAt:      time.Now(),
					},
				},
			}

			mockWorkloadDetailed := &openapi.WorkloadDetailed{
				Id:             testWorkloadId,
				Name:           "test-workload",
				Type:           workloadType,
				Phase:          openapi.PHASE_RUNNING,
				ProjectName:    "test-project",
				DepartmentName: "test-department",
				CreatedAt:      time.Now(),
			}

			mockTraining := &openapi.Training{
				Name:       "test-workload",
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewTrainingSpecSpec(),
			}

			mockWorkloadClient.EXPECT().ListSlim(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(client.ToSlimResponse(mockWorkloadList), nil)
			mockWorkloadClient.EXPECT().GetWorkload(gomock.Any(), gomock.Any()).Return(mockWorkloadDetailed, nil)
			mockWorkloadClient.EXPECT().GetTraining(gomock.Any(), gomock.Any()).Return(mockTraining, nil)
			mockPodsClient.EXPECT().ListByLimit(ctx, testWorkloadId, podLimit).Return(mockPods, nil)
			mockEventClient.EXPECT().ListByLimit(ctx, testWorkloadId, eventLimit).Return(mockEvents, nil)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)

			Expect(err).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(result.General.Name).To(Equal("test-workload"))
			Expect(result.General.Type).To(Equal(workloadType))
			Expect(result.General.Phase).To(Equal(displayer.Running))
			Expect(result.General.Project).To(Equal("test-project"))
			Expect(result.General.Department).To(Equal("test-department"))
			Expect(result.PodsInfo.Pods).To(HaveLen(1))
			Expect(result.Events).To(HaveLen(1))
		})

		It("should return an error when no workload matches the filter criteria", func() {
			reqWorkload := &api.Workload{
				ClusterID: "test-cluster",
				Project:   "non-existent-project",
				Name:      "non-existent-workload",
				Type:      internalApi.Training,
			}

			mockWorkloadList := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{},
			}

			mockWorkloadClient.EXPECT().ListSlim(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(client.ToSlimResponse(mockWorkloadList), nil)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("workload not found"))
			Expect(result).To(BeNil())
		})
	})

	Context("Workloads V2 spec resolution", func() {
		var mockWorkloadDetailed *openapi.WorkloadDetailed

		BeforeEach(func() {
			staleCommand := "stale-command"
			staleArguments := "stale-arguments"
			mockWorkloadDetailed = &openapi.WorkloadDetailed{
				Id:               testWorkloadId,
				Name:             "test-inference",
				Type:             internalApi.Inference.String(),
				Phase:            openapi.PHASE_RUNNING,
				ProjectName:      "test-project",
				CreatedAt:        time.Now(),
				AdditionalFields: map[string]interface{}{},
				Command:          &staleCommand,
				Arguments:        &staleArguments,
				RequestedNodePools: []string{
					"stale-pool",
				},
			}
			mockWorkloadDetailed.SetSourceApi(openapi.SOURCEAPI_WORKLOADS_V2)

			mockWorkloadClient.EXPECT().GetWorkload(ctx, testWorkloadId).Return(mockWorkloadDetailed, nil)
			mockPodsClient.EXPECT().ListByLimit(ctx, testWorkloadId, podLimit).Return(nil, nil)
			mockEventClient.EXPECT().ListByLimit(ctx, testWorkloadId, eventLimit).Return(nil, nil)
		})

		It("uses the typed response for workloads translated to V2", func() {
			mockWorkloadDetailed.Images = []string{"stale-image:latest"}

			inferenceSpec := openapi.NewInferenceSpecSpec()
			inferenceSpec.Image = *openapi.NewNullableString(ptr.To("snapshot-image:latest"))
			inferenceSpec.ImagePullPolicy = *openapi.NewNullableImagePullPolicy(ptr.To(openapi.IMAGEPULLPOLICY_ALWAYS))
			inferenceSpec.WorkingDir = *openapi.NewNullableString(ptr.To("/workspace"))
			inferenceSpec.Command = *openapi.NewNullableString(ptr.To("python"))
			inferenceSpec.Args = *openapi.NewNullableString(ptr.To("server.py"))
			inferenceSpec.NodePools = []string{"snapshot-pool"}
			inferenceSpec.Labels = []openapi.Label{createApiLabel("source", "typed-snapshot")}
			inferenceSpec.EnvironmentVariables = []openapi.EnvironmentVariable{
				{
					Name:  *openapi.NewNullableString(ptr.To("SOURCE")),
					Value: *openapi.NewNullableString(ptr.To("typed-snapshot")),
				},
			}
			compute := openapi.NewWorkspaceSpecSpecAllOfCompute()
			compute.SetGpuPortionRequest(0.5)
			compute.SetGpuPortionLimit(0.75)
			compute.SetCpuCoreRequest(0.25)
			compute.SetCpuMemoryRequest("128Mi")
			inferenceSpec.Compute.Set(compute)

			pvc := openapi.NewPvcInstance()
			pvc.ClaimName = *openapi.NewNullableString(ptr.To("models"))
			pvc.Path = *openapi.NewNullableString(ptr.To("/models"))
			storage := openapi.NewInferenceSpecSpecAllOfStorage()
			storage.Pvc = []openapi.PvcInstance{*pvc}
			inferenceSpec.Storage.Set(storage)

			mockInference := &openapi.Inference{
				Name:       "test-inference",
				WorkloadId: testWorkloadId,
				Spec:       inferenceSpec,
			}
			mockWorkloadClient.EXPECT().GetInference(ctx, testWorkloadId).Return(mockInference, nil)

			result, err := service.GetWorkload(ctx, &api.Workload{ID: testWorkloadId}, podLimit, eventLimit)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Runtime.Images).To(Equal([]string{"snapshot-image:latest"}))
			Expect(result.Runtime.ImagePullPolicy).To(Equal(ptr.To("Always")))
			Expect(result.Runtime.WorkingDir).To(Equal(ptr.To("/workspace")))
			Expect(result.General.Command).To(Equal(ptr.To("python")))
			Expect(result.General.Arguments).To(Equal(ptr.To("server.py")))
			Expect(result.General.EnvironmentVariables).NotTo(BeNil())
			Expect(*result.General.EnvironmentVariables).To(HaveKeyWithValue("SOURCE", "typed-snapshot"))
			Expect(result.Compute.RequestedResources.GPUCompute.Requested).To(Equal(0.5))
			Expect(result.Compute.RequestedResources.GPUCompute.Limit).NotTo(BeNil())
			Expect(*result.Compute.RequestedResources.GPUCompute.Limit).To(Equal(0.75))
			Expect(result.Compute.RequestedResources.CPUCores.Requested).To(Equal(0.25))
			Expect(result.Compute.RequestedResources.CPUMemory.Requested).To(Equal("128Mi"))
			Expect(result.Compute.RequestedNodePools).To(Equal([]string{"snapshot-pool"}))
			Expect(result.Storage.PVCs).To(HaveLen(1))
			Expect(result.Storage.PVCs[0].Name).To(Equal("models"))
			Expect(result.Spec.Labels).To(HaveLen(1))
			Expect(*result.Spec.Labels[0].Value.Get()).To(Equal("typed-snapshot"))
		})

		It("handles an explicitly null compute spec", func() {
			inferenceSpec := openapi.NewInferenceSpecSpec()
			inferenceSpec.SetComputeNil()

			mockInference := &openapi.Inference{
				Name:       "test-inference",
				WorkloadId: testWorkloadId,
				Spec:       inferenceSpec,
			}
			mockWorkloadClient.EXPECT().GetInference(ctx, testWorkloadId).Return(mockInference, nil)

			result, err := service.GetWorkload(ctx, &api.Workload{ID: testWorkloadId}, podLimit, eventLimit)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Compute.RequestedResources.GPUCompute.Requested).To(BeZero())
			Expect(result.Compute.RequestedResources.CPUCores.Requested).To(BeZero())
		})

		It("renders typed environment variable references", func() {
			inferenceSpec := openapi.NewInferenceSpecSpec()
			inferenceSpec.EnvironmentVariables = []openapi.EnvironmentVariable{
				{
					Name:  *openapi.NewNullableString(ptr.To("LITERAL")),
					Value: *openapi.NewNullableString(ptr.To("value")),
				},
				{
					Name: *openapi.NewNullableString(ptr.To("FROM_SECRET")),
					Secret: *openapi.NewNullableEnvironmentVariableSecret(&openapi.EnvironmentVariableSecret{
						Name: *openapi.NewNullableString(ptr.To("secret-name")),
						Key:  *openapi.NewNullableString(ptr.To("secret-key")),
					}),
				},
				{
					Name: *openapi.NewNullableString(ptr.To("FROM_CONFIG_MAP")),
					ConfigMap: *openapi.NewNullableEnvironmentVariableConfigMap(&openapi.EnvironmentVariableConfigMap{
						Name: *openapi.NewNullableString(ptr.To("config-map-name")),
						Key:  *openapi.NewNullableString(ptr.To("config-map-key")),
					}),
				},
				{
					Name: *openapi.NewNullableString(ptr.To("FROM_POD_FIELD")),
					PodFieldRef: *openapi.NewNullableEnvironmentVariablePodFieldReference(&openapi.EnvironmentVariablePodFieldReference{
						Path: *openapi.NewNullableString(ptr.To("metadata.name")),
					}),
				},
				{
					Name: *openapi.NewNullableString(ptr.To("FROM_CREDENTIAL")),
					UserCredential: *openapi.NewNullableEnvironmentVariableUserCredential(&openapi.EnvironmentVariableUserCredential{
						Name: *openapi.NewNullableString(ptr.To("credential-name")),
						Key:  *openapi.NewNullableString(ptr.To("credential-key")),
					}),
				},
				{
					Name:    *openapi.NewNullableString(ptr.To("EXCLUDED")),
					Value:   *openapi.NewNullableString(ptr.To("excluded-value")),
					Exclude: *openapi.NewNullableBool(ptr.To(true)),
				},
			}

			mockInference := &openapi.Inference{
				Name:       "test-inference",
				WorkloadId: testWorkloadId,
				Spec:       inferenceSpec,
			}
			mockWorkloadClient.EXPECT().GetInference(ctx, testWorkloadId).Return(mockInference, nil)

			result, err := service.GetWorkload(ctx, &api.Workload{ID: testWorkloadId}, podLimit, eventLimit)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.General.EnvironmentVariables).NotTo(BeNil())
			Expect(*result.General.EnvironmentVariables).To(Equal(map[string]string{
				"LITERAL":         "value",
				"FROM_SECRET":     "SECRET:secret-name,secret-key",
				"FROM_CONFIG_MAP": "CONFIGMAP:config-map-name,config-map-key",
				"FROM_POD_FIELD":  "PODFIELDREF:metadata.name",
				"FROM_CREDENTIAL": "USERCREDENTIAL:credential-name,credential-key",
			}))
		})

		It("returns the typed response error for recognized workload types", func() {
			mockWorkloadClient.EXPECT().GetInference(ctx, testWorkloadId).Return(nil, errors.New("typed response not found"))

			result, err := service.GetWorkload(ctx, &api.Workload{ID: testWorkloadId}, podLimit, eventLimit)

			Expect(err).To(MatchError(ContainSubstring("typed response not found")))
			Expect(result).To(BeNil())
		})

		It("uses the V2 manifest directly for raw workload types", func() {
			mockWorkloadDetailed.Type = "KubernetesJob"
			MockWorkloadV2Getter.EXPECT().Get(ctx, testWorkloadId).Return(workloadV2ResponseWithMetadata(
				map[string]interface{}{"source": "raw-manifest"},
				nil,
			), nil)

			result, err := service.GetWorkload(ctx, &api.Workload{ID: testWorkloadId}, podLimit, eventLimit)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.General.Type).To(Equal("KubernetesJob"))
			Expect(result.Spec.Labels).To(HaveLen(1))
			Expect(*result.Spec.Labels[0].Value.Get()).To(Equal("raw-manifest"))
		})
	})

	Context("Spec Generation", func() {
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

		It("Training", func() {
			mockWorkloadDetailed.Type = internalApi.Training.String()
			reqWorkload.Type = internalApi.Training
			mockWorkloadInList.Type = internalApi.Training.String()

			mockTraining := &openapi.Training{
				Name:       workloadName,
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewTrainingSpecSpec(),
			}
			mockTraining.Spec.Labels = []openapi.Label{
				createApiLabel(labelName1, labelValue1),
				createApiLabel(labelName2, labelValue2),
			}

			mockWorkloadClient.EXPECT().GetTraining(gomock.Any(), gomock.Any()).Return(mockTraining, nil)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			Expect(result.Spec).NotTo(BeNil())
			Expect(result.Spec.MasterLabels).To(BeNil())
			validateLabels(result.Spec.Labels, mockTraining.Spec.Labels)
		})

		It("Workspace", func() {
			mockWorkloadDetailed.Type = internalApi.Workspace.String()
			reqWorkload.Type = internalApi.Workspace
			mockWorkloadInList.Type = internalApi.Workspace.String()

			mockWorkspace := &openapi.Workspace{
				Name:       workloadName,
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewWorkspaceSpecSpec(),
			}
			mockWorkspace.Spec.Labels = []openapi.Label{
				createApiLabel(labelName1, labelValue1),
				createApiLabel(labelName2, labelValue2),
			}

			mockWorkloadClient.EXPECT().GetWorkspace(gomock.Any(), gomock.Any()).Return(mockWorkspace, nil)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			Expect(result.Spec).NotTo(BeNil())
			Expect(result.Spec.MasterLabels).To(BeNil())
			validateLabels(result.Spec.Labels, mockWorkspace.Spec.Labels)
		})

		Context("Distributed", func() {
			var (
				mockDistributed *openapi.DistributedWorkload
			)

			BeforeEach(func() {
				mockWorkloadDetailed.Type = internalApi.Distributed.String()
				reqWorkload.Type = internalApi.Distributed
				mockWorkloadInList.Type = internalApi.Distributed.String()

				mockDistributed = &openapi.DistributedWorkload{
					Name:       workloadName,
					WorkloadId: testWorkloadId,
					Spec:       openapi.NewDistributedSpecSpec(),
				}
				mockDistributed.Spec.Labels = []openapi.Label{
					createApiLabel(labelName1, labelValue1),
					createApiLabel(labelName2, labelValue2),
				}
			})

			It("with master spec", func() {
				masterSpec := openapi.NewMasterSpec()
				masterSpec.Labels = []openapi.Label{
					createApiLabel(labelName1, labelValue1),
					createApiLabel(labelName2, labelValue2),
				}
				mockDistributed.MasterSpec.Set(masterSpec)

				mockWorkloadClient.EXPECT().GetDistributed(gomock.Any(), gomock.Any()).Return(mockDistributed, nil)

				result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				Expect(result.Spec).NotTo(BeNil())
				validateLabels(result.Spec.Labels, mockDistributed.Spec.Labels)
				validateLabels(result.Spec.MasterLabels, mockDistributed.MasterSpec.Get().Labels)
			})

			It("without master spec", func() {
				mockWorkloadClient.EXPECT().GetDistributed(gomock.Any(), gomock.Any()).Return(mockDistributed, nil)

				result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				Expect(result.Spec).NotTo(BeNil())
				Expect(result.Spec.MasterLabels).To(BeNil())
				validateLabels(result.Spec.Labels, mockDistributed.Spec.Labels)
			})

			It("with master config (different image, storage, env vars)", func() {
				// Set up worker spec with different values
				mockDistributed.Spec.Image = *openapi.NewNullableString(ptr.To("worker-image:latest"))
				mockDistributed.Spec.Command = *openapi.NewNullableString(ptr.To("python"))
				mockDistributed.Spec.Args = *openapi.NewNullableString(ptr.To("--worker"))
				mockDistributed.Spec.EnvironmentVariables = []openapi.EnvironmentVariable{
					{
						Name:  *openapi.NewNullableString(ptr.To("ROLE")),
						Value: *openapi.NewNullableString(ptr.To("worker")),
					},
				}

				// Create worker storage
				workerPvc := openapi.NewPvcInstance()
				workerPvc.ClaimName = *openapi.NewNullableString(ptr.To("worker-pvc"))
				workerPvc.Path = *openapi.NewNullableString(ptr.To("/worker/data"))
				workerStorage := openapi.NewWorkspaceSpecSpecAllOfStorage()
				workerStorage.Pvc = []openapi.PvcInstance{*workerPvc}
				mockDistributed.Spec.Storage.Set(workerStorage)

				// Create worker compute
				workerCompute := openapi.NewWorkspaceSpecSpecAllOfCompute()
				workerCompute.GpuDevicesRequest = *openapi.NewNullableInt32(ptr.To(int32(1)))
				workerCompute.CpuCoreRequest = *openapi.NewNullableFloat64(ptr.To(float64(2)))
				workerCompute.CpuMemoryRequest = *openapi.NewNullableString(ptr.To("4Gi"))
				mockDistributed.Spec.Compute.Set(workerCompute)

				// Create master spec with different values
				masterSpec := openapi.NewMasterSpec()
				masterSpec.Image = *openapi.NewNullableString(ptr.To("master-image:latest"))
				masterSpec.Command = *openapi.NewNullableString(ptr.To("python"))
				masterSpec.Args = *openapi.NewNullableString(ptr.To("--master"))
				masterSpec.WorkingDir = *openapi.NewNullableString(ptr.To("/app"))
				masterSpec.EnvironmentVariables = []openapi.EnvironmentVariable{
					{
						Name:  *openapi.NewNullableString(ptr.To("ROLE")),
						Value: *openapi.NewNullableString(ptr.To("master")),
					},
				}

				// Create master storage
				masterPvc := openapi.NewPvcInstance()
				masterPvc.ClaimName = *openapi.NewNullableString(ptr.To("master-pvc"))
				masterPvc.Path = *openapi.NewNullableString(ptr.To("/master/data"))
				masterStorage := openapi.NewWorkspaceSpecSpecAllOfStorage()
				masterStorage.Pvc = []openapi.PvcInstance{*masterPvc}
				masterSpec.Storage.Set(masterStorage)

				// Create master compute
				masterCompute := openapi.NewWorkspaceSpecSpecAllOfCompute()
				masterCompute.GpuDevicesRequest = *openapi.NewNullableInt32(ptr.To(int32(2)))
				masterCompute.CpuCoreRequest = *openapi.NewNullableFloat64(ptr.To(float64(4)))
				masterCompute.CpuMemoryRequest = *openapi.NewNullableString(ptr.To("8Gi"))
				masterSpec.Compute.Set(masterCompute)

				mockDistributed.MasterSpec.Set(masterSpec)

				mockWorkloadClient.EXPECT().GetDistributed(gomock.Any(), gomock.Any()).Return(mockDistributed, nil)

				result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				// Verify worker storage
				Expect(result.Storage).NotTo(BeNil())
				Expect(result.Storage.PVCs).To(HaveLen(1))
				Expect(result.Storage.PVCs[0].Name).To(Equal("worker-pvc"))
				Expect(result.Storage.PVCs[0].MountPath).To(Equal("/worker/data"))

				// Verify master config (full configuration)
				Expect(result.MasterConfig).NotTo(BeNil())
				Expect(*result.MasterConfig.Image).To(Equal("master-image:latest"))
				Expect(*result.MasterConfig.Command).To(Equal("python"))
				Expect(*result.MasterConfig.Arguments).To(Equal("--master"))
				Expect(*result.MasterConfig.WorkingDir).To(Equal("/app"))
				Expect(result.MasterConfig.EnvironmentVariables).To(HaveKeyWithValue("ROLE", "master"))

				// Verify master compute
				Expect(result.MasterConfig.Compute).NotTo(BeNil())
				Expect(result.MasterConfig.Compute.RequestedResources.GPUCompute.Requested).To(Equal(float64(2)))
				Expect(result.MasterConfig.Compute.RequestedResources.CPUCores.Requested).To(Equal(float64(4)))
				Expect(result.MasterConfig.Compute.RequestedResources.CPUMemory.Requested).To(Equal("8Gi"))

				// Verify master storage within master config
				Expect(result.MasterConfig.Storage).NotTo(BeNil())
				Expect(result.MasterConfig.Storage.PVCs).To(HaveLen(1))
				Expect(result.MasterConfig.Storage.PVCs[0].Name).To(Equal("master-pvc"))
				Expect(result.MasterConfig.Storage.PVCs[0].MountPath).To(Equal("/master/data"))
			})
		})

		Context("DistributedInference", func() {
			var (
				mockDistributedInference *openapi.DistributedInference
			)

			BeforeEach(func() {
				mockWorkloadDetailed.Type = internalApi.DistributedInference.StringCapitalize()
				reqWorkload.Type = internalApi.DistributedInference
				mockWorkloadInList.Type = internalApi.DistributedInference.StringCapitalize()

				mockDistributedInference = &openapi.DistributedInference{
					Name:       workloadName,
					WorkloadId: testWorkloadId,
					Spec:       openapi.NewDistributedInferenceSpecSpec(),
				}
			})

			It("with full spec", func() {
				servingPort := &openapi.DistributedInferenceServingPort{
					Port:              *openapi.NewNullableInt32(ptr.To(int32(8080))),
					AuthorizationType: *openapi.NewNullableDistributedInferenceServingPortAccessAuthorizationTypeEnum(ptr.To(openapi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC)),
					AuthorizedUsers:   []string{"user1", "user2"},
					AuthorizedGroups:  []string{"group1"},
					Protocol:          *openapi.NewNullableDistributedInferenceServingPortProtocol(ptr.To(openapi.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP)),
					ExposeExternally:  *openapi.NewNullableBool(ptr.To(true)),
				}

				spec := mockDistributedInference.Spec
				spec.RestartPolicy = *openapi.NewNullableDistributedInferenceRestartPolicy(
					ptr.To(openapi.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART))
				spec.StartupPolicy = *openapi.NewNullableDistributedInferenceStartupPolicy(
					ptr.To(openapi.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED))
				spec.Replicas = *openapi.NewNullableInt32(ptr.To(int32(3)))
				spec.Workers = *openapi.NewNullableInt32(ptr.To(int32(2)))
				spec.ServingPort = *openapi.NewNullableDistributedInferenceServingPort(servingPort)

				mockWorkloadClient.EXPECT().GetDistributedInference(gomock.Any(), gomock.Any()).Return(mockDistributedInference, nil)

				result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				// Verify distributed inference fields
				Expect(result.DistributedInference).NotTo(BeNil())
				Expect(result.DistributedInference.RestartPolicy).To(Equal(ptr.To("RecreateGroupOnPodRestart")))
				Expect(result.DistributedInference.StartupPolicy).To(Equal(ptr.To("LeaderCreated")))
				Expect(result.DistributedInference.Replicas).To(Equal(ptr.To(int32(3))))
				Expect(result.DistributedInference.Workers).To(Equal(ptr.To(int32(2))))
				Expect(result.DistributedInference.ServingPort).To(Equal(ptr.To(int32(8080))))
				Expect(result.DistributedInference.AuthorizationType).To(Equal(ptr.To("public")))
				Expect(result.DistributedInference.AuthorizedUsers).To(Equal([]string{"user1", "user2"}))
				Expect(result.DistributedInference.AuthorizedGroups).To(Equal([]string{"group1"}))
				Expect(result.DistributedInference.Protocol).To(Equal(ptr.To("http")))
				Expect(result.DistributedInference.ExposeExternally).To(Equal(ptr.To(true)))
			})

			It("with minimal spec", func() {
				// Use empty spec to test nil handling
				mockDistributedInference.Spec = &openapi.DistributedInferenceSpecSpec{}

				mockWorkloadClient.EXPECT().GetDistributedInference(gomock.Any(), gomock.Any()).Return(mockDistributedInference, nil)

				result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				// Verify distributed inference fields are nil for unset values
				Expect(result.DistributedInference).NotTo(BeNil())
				Expect(result.DistributedInference.RestartPolicy).To(BeNil())
				Expect(result.DistributedInference.StartupPolicy).To(BeNil())
				Expect(result.DistributedInference.Replicas).To(BeNil())
				Expect(result.DistributedInference.Workers).To(BeNil())
			})
		})
	})

	Context("issues", func() {
		const (
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
				Type:        internalApi.Training.String(),
				Phase:       openapi.PHASE_RUNNING,
				ProjectName: projectName,
				CreatedAt:   time.Now(),
			}

			reqWorkload = &api.Workload{
				ClusterID: "test-cluster",
				Project:   projectName,
				Name:      workloadName,
				Type:      internalApi.Training,
			}

			mockWorkloadInList = &openapi.Workload{
				Id:          testWorkloadId,
				Name:        workloadName,
				Phase:       openapi.PHASE_RUNNING,
				ProjectName: projectName,
				Type:        internalApi.Training.String(),
			}
			mockWorkloadList := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{*mockWorkloadInList},
			}
			mockTraining := &openapi.Training{
				Name:       workloadName,
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewTrainingSpecSpec(),
			}

			mockWorkloadClient.EXPECT().GetWorkload(gomock.Any(), gomock.Any()).Return(mockWorkloadDetailed, nil)
			mockWorkloadClient.EXPECT().ListSlim(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(client.ToSlimResponse(mockWorkloadList), nil)
			mockPodsClient.EXPECT().ListByLimit(ctx, testWorkloadId, podLimit).Return(mockPods, nil)
			mockEventClient.EXPECT().ListByLimit(ctx, testWorkloadId, eventLimit).Return(mockEvents, nil)
			mockWorkloadClient.EXPECT().GetTraining(gomock.Any(), gomock.Any()).Return(mockTraining, nil)
		})

		It("issues are sorted", func() {
			issues := getTestIssues()
			issuesStr, err := json.Marshal(issues)
			Expect(err).NotTo(HaveOccurred())
			mockWorkloadDetailed.AdditionalFields = map[string]interface{}{
				describe.AdditionalFieldsIssues: string(issuesStr),
			}

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Issues).NotTo(BeNil())
			Expect(result.Issues).To(HaveLen(len(issues.Compute) + len(issues.Internal) + len(issues.Storage) + len(issues.Network)))

			for i := 0; i+1 < len(result.Issues); i++ {
				first := result.Issues[i]
				second := result.Issues[i+1]

				typesCompareResult := strings.Compare(first.Type, second.Type)
				Expect(typesCompareResult <= 0).To(BeTrue())

				// if issue type is the same, make sure sorted by resource kind
				if typesCompareResult == 0 {
					kindCompareResult := strings.Compare(first.Kind, second.Kind)
					Expect(kindCompareResult <= 0).To(BeTrue())

					// if resource kind is the same, make sure sorted by resource name
					if kindCompareResult == 0 {
						Expect(strings.Compare(first.Name, second.Name) <= 0).To(BeTrue())
					}
				}
			}
		})

		It("no issues", func() {
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Issues).To(BeNil())
		})
	})

	Context("GPU Network Accelerator field", func() {
		const (
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
				Type:        internalApi.Distributed.String(),
				Phase:       openapi.PHASE_RUNNING,
				ProjectName: projectName,
				CreatedAt:   time.Now(),
			}

			reqWorkload = &api.Workload{
				ClusterID: "test-cluster",
				Project:   projectName,
				Name:      workloadName,
				Type:      internalApi.Distributed,
			}

			mockWorkloadInList = &openapi.Workload{
				Id:          testWorkloadId,
				Name:        workloadName,
				Phase:       openapi.PHASE_RUNNING,
				ProjectName: projectName,
				Type:        internalApi.Distributed.String(),
			}
			mockWorkloadList := &openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{*mockWorkloadInList},
			}
			mockDistributed := &openapi.DistributedWorkload{
				Name:       workloadName,
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewDistributedSpecSpec(),
			}

			mockWorkloadClient.EXPECT().GetWorkload(gomock.Any(), gomock.Any()).Return(mockWorkloadDetailed, nil)
			mockWorkloadClient.EXPECT().ListSlim(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(client.ToSlimResponse(mockWorkloadList), nil)
			mockPodsClient.EXPECT().ListByLimit(ctx, testWorkloadId, podLimit).Return(mockPods, nil)
			mockEventClient.EXPECT().ListByLimit(ctx, testWorkloadId, eventLimit).Return(mockEvents, nil)
			mockWorkloadClient.EXPECT().GetDistributed(gomock.Any(), gomock.Any()).Return(mockDistributed, nil)
		})

		It("Without GPU Acceleration field", func() {
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.GPUAcceleration).To(BeNil())
		})

		It("With GPU Acceleration field false", func() {
			field, err := json.Marshal(describe.GPUNetworkAcceleration{})
			Expect(err).NotTo(HaveOccurred())
			mockWorkloadDetailed.AdditionalFields = map[string]interface{}{
				describe.AdditionalFieldsGPUNetworkAccelerationStatus: string(field),
			}
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.GPUAcceleration).To(Equal(ptr.To(false)))
		})

		It("With GPU Acceleration field true", func() {
			field, err := json.Marshal(describe.GPUNetworkAcceleration{Used: true})
			Expect(err).NotTo(HaveOccurred())
			mockWorkloadDetailed.AdditionalFields = map[string]interface{}{
				describe.AdditionalFieldsGPUNetworkAccelerationStatus: string(field),
			}
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.GPUAcceleration).To(Equal(ptr.To(true)))
		})
	})

	Context("Network Topology", func() {
		const (
			workloadName = "test-workload"
			projectName  = "test-project"
		)

		var (
			mockWorkloadDetailed *openapi.WorkloadDetailed
			reqWorkload          *api.Workload
		)

		BeforeEach(func() {
			mockWorkloadDetailed = &openapi.WorkloadDetailed{
				Id:          testWorkloadId,
				Name:        workloadName,
				Type:        internalApi.Training.String(),
				Phase:       openapi.PHASE_RUNNING,
				ProjectName: projectName,
				CreatedAt:   time.Now(),
			}

			reqWorkload = &api.Workload{
				ID: testWorkloadId,
			}

			mockTraining := &openapi.Training{
				Name:       workloadName,
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewTrainingSpecSpec(),
			}

			mockWorkloadClient.EXPECT().GetWorkload(gomock.Any(), gomock.Any()).Return(mockWorkloadDetailed, nil)
			mockWorkloadClient.EXPECT().GetTraining(gomock.Any(), gomock.Any()).Return(mockTraining, nil)
			mockPodsClient.EXPECT().ListByLimit(gomock.Any(), gomock.Any(), podLimit).Return(mockPods, nil)
			mockEventClient.EXPECT().ListByLimit(gomock.Any(), gomock.Any(), eventLimit).Return(mockEvents, nil)
		})

		It("should set network topology when topology is present", func() {
			topology := openapi.NewWorkloadTopology()
			topology.Name = ptr.To("my-topology")
			topology.RequiredPlacement = ptr.To("rack")
			topology.PreferredPlacement = ptr.To("block")
			actualPlacement := &openapi.WorkloadTopologyActualPlacement{
				Level: ptr.To("rack"),
				Value: ptr.To("rack-1"),
			}
			topology.ActualPlacement = *openapi.NewNullableWorkloadTopologyActualPlacement(actualPlacement)
			mockWorkloadDetailed.Topology = *openapi.NewNullableWorkloadTopology(topology)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.NetworkTopology).NotTo(BeNil())
			Expect(*result.NetworkTopology.Name).To(Equal("my-topology"))
			Expect(*result.NetworkTopology.RequiredPlacement).To(Equal("rack"))
			Expect(*result.NetworkTopology.PreferredPlacement).To(Equal("block"))
			Expect(*result.NetworkTopology.ActualPlacement).To(Equal("rack (rack-1)"))
		})

		It("should not set network topology when topology is not set", func() {
			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.NetworkTopology).To(BeNil())
		})

		It("should not set network topology when name is empty", func() {
			topology := openapi.NewWorkloadTopology()
			topology.Name = ptr.To("")
			mockWorkloadDetailed.Topology = *openapi.NewNullableWorkloadTopology(topology)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.NetworkTopology).To(BeNil())
		})

		It("should set network topology without actual placement when not set", func() {
			topology := openapi.NewWorkloadTopology()
			topology.Name = ptr.To("my-topology")
			topology.RequiredPlacement = ptr.To("rack")
			mockWorkloadDetailed.Topology = *openapi.NewNullableWorkloadTopology(topology)

			result, err := service.GetWorkload(ctx, reqWorkload, podLimit, eventLimit)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.NetworkTopology).NotTo(BeNil())
			Expect(*result.NetworkTopology.Name).To(Equal("my-topology"))
			Expect(*result.NetworkTopology.RequiredPlacement).To(Equal("rack"))
			Expect(result.NetworkTopology.ActualPlacement).To(BeNil())
		})
	})

	Context("Pod Role Extraction", func() {
		It("should extract role from container name (leader/worker/none)", func() {
			mockWorkloadDetailed := &openapi.WorkloadDetailed{
				Id:               testWorkloadId,
				Name:             "test-role",
				Type:             internalApi.DistributedInference.StringCapitalize(),
				Phase:            openapi.PHASE_RUNNING,
				ProjectName:      "test-project",
				CreatedAt:        time.Now(),
				AdditionalFields: map[string]interface{}{},
			}

			mockDistributedInference := &openapi.DistributedInference{
				Name:       "test-role",
				WorkloadId: testWorkloadId,
				Spec:       openapi.NewDistributedInferenceSpecSpec(),
			}

			mockPods := []openapi.Pod{
				{Name: "pod-0", K8sPhase: "Running", Containers: []openapi.Container{{Name: "leader", Image: "img"}}},
				{Name: "pod-1", K8sPhase: "Running", Containers: []openapi.Container{{Name: "worker", Image: "img"}}},
				{Name: "pod-2", K8sPhase: "Running", Containers: []openapi.Container{{Name: "Worker", Image: "img"}}},
				{Name: "pod-3", K8sPhase: "Pending", Containers: []openapi.Container{}},
				{Name: "pod-4", K8sPhase: "Running", Containers: []openapi.Container{{Name: "main", Image: "img"}}},
			}

			mockWorkloadClient.EXPECT().GetWorkload(gomock.Any(), gomock.Any()).Return(mockWorkloadDetailed, nil)
			mockWorkloadClient.EXPECT().GetDistributedInference(gomock.Any(), gomock.Any()).Return(mockDistributedInference, nil)
			mockPodsClient.EXPECT().ListByLimit(gomock.Any(), gomock.Any(), podLimit).Return(mockPods, nil)
			mockEventClient.EXPECT().ListByLimit(gomock.Any(), gomock.Any(), eventLimit).Return([]openapi.Event{}, nil)

			result, err := service.GetWorkload(ctx, &api.Workload{ID: testWorkloadId}, podLimit, eventLimit)

			Expect(err).To(BeNil())
			Expect(result.PodsInfo.Pods).To(HaveLen(5))
			Expect(*result.PodsInfo.Pods[0].Role).To(Equal("Leader"))
			Expect(*result.PodsInfo.Pods[1].Role).To(Equal("Worker"))
			Expect(*result.PodsInfo.Pods[2].Role).To(Equal("Worker"))
			Expect(result.PodsInfo.Pods[3].Role).To(BeNil())
			Expect(result.PodsInfo.Pods[4].Role).To(BeNil())
		})
	})
})

func validateLabels(specLabels []*openapi.Label, source []openapi.Label) {
	Expect(specLabels).To(HaveLen(len(source)))
	for i, sourceLabel := range source {
		specLabel := specLabels[i]
		Expect(specLabel.Name).To(Equal(sourceLabel.Name))
		Expect(specLabel.Value).To(Equal(sourceLabel.Value))
	}
}

func createApiLabel(name, value string) openapi.Label {
	return openapi.Label{
		Name:  *openapi.NewNullableString(ptr.To(name)),
		Value: *openapi.NewNullableString(ptr.To(value)),
	}
}

func workloadV2ResponseWithMetadata(
	labels map[string]interface{},
	annotations map[string]interface{},
) *workloadmanager.WorkloadV2GetResponse {
	manifest := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels":      labels,
			"annotations": annotations,
		},
	}

	return &workloadmanager.WorkloadV2GetResponse{
		Workload: workloadmanager.WorkloadV2{
			WorkloadManifest: manifest,
		},
	}
}

func getTestIssues() describe.Issues {
	issues := describe.Issues{
		Compute: []displayer.Issue{
			{
				Kind:        constants.KindRunaiJob,
				Name:        "job2",
				Description: "abc",
			},
			{
				Kind:        constants.KindRunaiJob,
				Name:        "job1",
				Description: "abc",
			},
		},
		Network: []displayer.Issue{
			{
				Kind:        constants.KindService,
				Name:        "s1",
				Description: "abc",
			},
			{
				Kind:        constants.KindIngress,
				Name:        "ing",
				Description: "abc",
			},
			{
				Kind:        constants.KindService,
				Name:        "aaa",
				Description: "abc",
			},
		},
		Internal: []displayer.Issue{
			{
				Kind:        constants.KindTrainingWorkload,
				Name:        "w1",
				Description: "abc",
			},
		},
		Storage: []displayer.Issue{
			{
				Kind:        constants.KindPvc,
				Name:        "pvc1234",
				Description: "abc",
			},
		},
	}

	return issues
}
