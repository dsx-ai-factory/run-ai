/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"context"
	"errors"
	"time"

	cptemplatemocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/template"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/describe"
	workloads_openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Helper function to create a pointer
func ptr[T any](v T) *T {
	return &v
}

var _ = Describe("Service", func() {
	var (
		ctrl         *gomock.Controller
		mockClient   *cptemplatemocks.MockTemplateDescribeLister
		service      *describe.Service
		ctx          context.Context
		templateName string
		clusterId    string
		projectId    string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClient = cptemplatemocks.NewMockTemplateDescribeLister(ctrl)
		clusterId = "cluster-123"
		service = describe.NewService(mockClient, clusterId)
		ctx = context.Background()
		templateName = "test-template"
		projectId = "project-456"
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("GetByName", func() {
		It("should successfully retrieve template", func() {
			createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

			rawTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "template-123",
					Name:         "test-template",
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					ScopeId:      "scope-456",
					CreatedAt:    createdAt,
					CreatedBy:    "user@example.com",
				},
			}

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{rawTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.GetByName(ctx, templateName, projectId)

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
			Expect(template.Meta.ID).To(Equal("template-123"))
			Expect(template.Meta.Name).To(Equal("test-template"))
			Expect(template.Meta.Type).To(Equal("Training"))
			Expect(template.Meta.Scope).To(Equal("project"))
			Expect(template.Meta.CreatedBy).To(Equal("user@example.com"))
		})

		It("should return error when client fails", func() {
			// Setup List expectation - return error
			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("API error"))

			template, err := service.GetByName(ctx, templateName, projectId)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to fetch template"))
			Expect(template).To(BeNil())
		})

		It("should return 'not found' error when no templates found (describe)", func() {
			// Setup List expectation - return empty list
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.GetByName(ctx, templateName, projectId)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("template 'test-template' not found"))
			Expect(template).To(BeNil())
		})

		It("should return highest priority template when multiple templates with same name exist", func() {
			createdAt := time.Now()

			// Create templates with different scopes
			projectTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "template-project",
					Name:         "test-template",
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					ScopeId:      "scope-1",
					CreatedAt:    createdAt,
					CreatedBy:    "user@example.com",
				},
			}

			clusterTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "template-cluster",
					Name:         "test-template",
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					ScopeType:    workloads_openapi.SCOPETYPE_CLUSTER,
					ScopeId:      "scope-2",
					CreatedAt:    createdAt,
					CreatedBy:    "admin@example.com",
				},
			}

			departmentTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "template-department",
					Name:         "test-template",
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					ScopeType:    workloads_openapi.SCOPETYPE_DEPARTMENT,
					ScopeId:      "scope-3",
					CreatedAt:    createdAt,
					CreatedBy:    "dept@example.com",
				},
			}

			// Setup List expectation - return templates in random order
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{clusterTemplate, projectTemplate, departmentTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.GetByName(ctx, templateName, projectId)

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
			// Should return project scope (highest priority)
			Expect(template.Meta.ID).To(Equal("template-project"))
			Expect(template.Meta.Scope).To(Equal("project"))
			Expect(template.Meta.CreatedBy).To(Equal("user@example.com"))
		})

		It("should properly transform template with spec fields", func() {
			createdAt := time.Now()
			updatedAt := createdAt.Add(24 * time.Hour)

			image := "ubuntu:22.04"
			command := "python train.py"
			args := "--epochs 100"
			priorityClass := "high-priority"
			description := "Test description"
			clusterIdVal := "cluster-789"
			updatedBy := "admin@example.com"
			projectIdVal := int32(100)

			envVar := workloads_openapi.EnvironmentVariable{}
			envVar.Name = *workloads_openapi.NewNullableString(ptr("EPOCHS"))
			envVar.Value = *workloads_openapi.NewNullableString(ptr("100"))

			rawTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "template-123",
					Name:         "test-template",
					Description:  *workloads_openapi.NewNullableString(&description),
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					ScopeId:      "scope-456",
					ClusterId:    *workloads_openapi.NewNullableString(&clusterIdVal),
					ProjectId:    *workloads_openapi.NewNullableInt32(&projectIdVal),
					CreatedAt:    createdAt,
					CreatedBy:    "user@example.com",
					UpdatedAt:    *workloads_openapi.NewNullableTime(&updatedAt),
					UpdatedBy:    *workloads_openapi.NewNullableString(&updatedBy),
				},
				Spec: &workloads_openapi.SupersetSpec{
					Image:                *workloads_openapi.NewNullableString(&image),
					Command:              *workloads_openapi.NewNullableString(&command),
					Args:                 *workloads_openapi.NewNullableString(&args),
					Preemptibility:       *workloads_openapi.NewNullablePreemptibility(ptr(workloads_openapi.PREEMPTIBILITY_PREEMPTIBLE)),
					PriorityClass:        *workloads_openapi.NewNullableString(&priorityClass),
					NodePools:            []string{"gpu-pool"},
					EnvironmentVariables: []workloads_openapi.EnvironmentVariable{envVar},
				},
			}

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{rawTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.GetByName(ctx, templateName, projectId)

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())

			// Verify meta fields
			Expect(template.Meta.ID).To(Equal("template-123"))
			Expect(template.Meta.Name).To(Equal("test-template"))
			Expect(*template.Meta.Description).To(Equal("Test description"))
			Expect(template.Meta.Type).To(Equal("Training"))
			Expect(template.Meta.Scope).To(Equal("project"))
			Expect(*template.Meta.ClusterID).To(Equal("cluster-789"))
			Expect(*template.Meta.ProjectID).To(Equal(int32(100)))
			Expect(*template.Meta.UpdatedBy).To(Equal("admin@example.com"))

			// Verify spec fields
			Expect(*template.Spec.Image).To(Equal("ubuntu:22.04"))
			Expect(*template.Spec.Command).To(Equal("python train.py"))
			Expect(*template.Spec.Args).To(Equal("--epochs 100"))
			Expect(*template.Spec.Preemptible).To(BeTrue())
			Expect(*template.Spec.PriorityClass).To(Equal("high-priority"))
			Expect(template.Spec.NodePools).To(Equal([]string{"gpu-pool"}))
			Expect(template.Spec.EnvironmentVariables).To(HaveLen(1))
			Expect(template.Spec.EnvironmentVariables["EPOCHS"]).To(Equal("100"))
		})

		It("should handle template with compute resources", func() {
			gpuDevices := int32(4)
			gpuMemory := "16GB"
			cpuCores := float64(8.0)
			cpuMemory := "32GB"

			compute := workloads_openapi.NewWorkspaceSpecSpecAllOfCompute()
			compute.SetGpuDevicesRequest(gpuDevices)
			compute.SetGpuMemoryRequest(gpuMemory)
			compute.SetCpuCoreRequest(cpuCores)
			compute.SetCpuMemoryRequest(cpuMemory)

			rawTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "template-123",
					Name:         "compute-template",
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					CreatedAt:    time.Now(),
				},
				Spec: &workloads_openapi.SupersetSpec{
					Compute: *workloads_openapi.NewNullableWorkspaceSpecSpecAllOfCompute(compute),
				},
			}

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{rawTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.GetByName(ctx, "compute-template", projectId)

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
			Expect(template.Spec.Compute).NotTo(BeNil())
			Expect(*template.Spec.Compute.GpuDevicesRequest).To(Equal(int32(4)))
			Expect(*template.Spec.Compute.GpuMemoryRequest).To(Equal("16GB"))
			Expect(*template.Spec.Compute.CpuCoreRequest).To(Equal(float64(8.0)))
			Expect(*template.Spec.Compute.CpuMemoryRequest).To(Equal("32GB"))
		})

		It("should handle template with ports and exposed URLs", func() {
			containerPort := int32(8080)
			externalPort := int32(80)
			url := "https://example.com/app"
			urlName := "webapp"

			port := workloads_openapi.NewPort()
			port.SetContainer(containerPort)
			port.SetExternal(externalPort)

			exposedUrl := workloads_openapi.NewExposedUrl()
			exposedUrl.SetUrl(url)
			exposedUrl.SetName(urlName)

			rawTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "template-123",
					Name:         "network-template",
					WorkloadType: workloads_openapi.WORKLOADTYPE_WORKSPACE,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					CreatedAt:    time.Now(),
				},
				Spec: &workloads_openapi.SupersetSpec{
					Ports:       []workloads_openapi.Port{*port},
					ExposedUrls: []workloads_openapi.ExposedUrl{*exposedUrl},
				},
			}

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{rawTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.GetByName(ctx, "network-template", projectId)

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
			Expect(template.Spec.Ports).To(HaveLen(1))
			Expect(template.Spec.Ports[0].Container).To(Equal(int32(8080)))
			Expect(*template.Spec.Ports[0].External).To(Equal(int32(80)))
			Expect(template.Spec.ExposedUrls).To(HaveLen(1))
			Expect(template.Spec.ExposedUrls[0].Url).To(Equal("https://example.com/app"))
			Expect(template.Spec.ExposedUrls[0].Name).To(Equal("webapp"))
		})

		It("should handle distributed template with master spec (master != worker)", func() {
			// Worker spec configuration
			workerImage := "ubuntu-worker:latest"
			workerGpuDevices := int32(2)
			workerCpuCores := float64(4.0)
			workerCpuMemory := "16GB"

			workerCompute := workloads_openapi.NewWorkspaceSpecSpecAllOfCompute()
			workerCompute.SetGpuDevicesRequest(workerGpuDevices)
			workerCompute.SetCpuCoreRequest(workerCpuCores)
			workerCompute.SetCpuMemoryRequest(workerCpuMemory)

			// Master spec configuration (different from worker)
			masterImage := "ubuntu-master:latest"
			masterCommand := "python master.py"
			masterArgs := "--role master"
			masterWorkingDir := "/app/master"
			masterGpuDevices := int32(1)
			masterCpuCores := float64(2.0)
			masterCpuMemory := "8GB"

			masterCompute := workloads_openapi.NewWorkspaceSpecSpecAllOfCompute()
			masterCompute.SetGpuDevicesRequest(masterGpuDevices)
			masterCompute.SetCpuCoreRequest(masterCpuCores)
			masterCompute.SetCpuMemoryRequest(masterCpuMemory)

			masterEnvVar := workloads_openapi.EnvironmentVariable{}
			masterEnvVar.Name = *workloads_openapi.NewNullableString(ptr("ROLE"))
			masterEnvVar.Value = *workloads_openapi.NewNullableString(ptr("master"))

			masterPort := workloads_openapi.NewPort()
			masterPort.SetContainer(int32(8888))
			masterPort.SetExternal(int32(30888))

			masterUrl := workloads_openapi.NewExposedUrl()
			masterUrl.SetUrl("https://master.example.com")
			masterUrl.SetName("master-ui")

			combinedMasterSpec := workloads_openapi.NewMasterTemplateSpec()
			combinedMasterSpec.SetImage(masterImage)
			combinedMasterSpec.SetCommand(masterCommand)
			combinedMasterSpec.SetArgs(masterArgs)
			combinedMasterSpec.SetWorkingDir(masterWorkingDir)
			combinedMasterSpec.SetCompute(*masterCompute)
			combinedMasterSpec.SetEnvironmentVariables([]workloads_openapi.EnvironmentVariable{masterEnvVar})
			combinedMasterSpec.NodePools = []string{"master-pool"}
			combinedMasterSpec.Ports = []workloads_openapi.Port{*masterPort}
			combinedMasterSpec.ExposedUrls = []workloads_openapi.ExposedUrl{*masterUrl}

			rawTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "distributed-template-123",
					Name:         "distributed-template",
					WorkloadType: workloads_openapi.WORKLOADTYPE_DISTRIBUTED,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					CreatedAt:    time.Now(),
				},
				Spec: &workloads_openapi.SupersetSpec{
					Image:   *workloads_openapi.NewNullableString(&workerImage),
					Compute: *workloads_openapi.NewNullableWorkspaceSpecSpecAllOfCompute(workerCompute),
				},
			}
			rawTemplate.SetCombinedMasterSpec(*combinedMasterSpec)

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{rawTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.GetByName(ctx, "distributed-template", projectId)

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())

			// Verify worker spec
			Expect(*template.Spec.Image).To(Equal("ubuntu-worker:latest"))
			Expect(template.Spec.Compute).NotTo(BeNil())
			Expect(*template.Spec.Compute.GpuDevicesRequest).To(Equal(int32(2)))
			Expect(*template.Spec.Compute.CpuCoreRequest).To(Equal(float64(4.0)))
			Expect(*template.Spec.Compute.CpuMemoryRequest).To(Equal("16GB"))

			// Verify master spec is populated (this was the bug - masterSpec was not shown)
			Expect(template.MasterSpec).NotTo(BeNil(), "MasterSpec should be populated for distributed templates")
			Expect(*template.MasterSpec.Image).To(Equal("ubuntu-master:latest"))
			Expect(*template.MasterSpec.Command).To(Equal("python master.py"))
			Expect(*template.MasterSpec.Args).To(Equal("--role master"))
			Expect(*template.MasterSpec.WorkingDir).To(Equal("/app/master"))
			Expect(template.MasterSpec.NodePools).To(Equal([]string{"master-pool"}))
			Expect(template.MasterSpec.EnvironmentVariables).To(HaveKeyWithValue("ROLE", "master"))

			// Verify master compute
			Expect(template.MasterSpec.Compute).NotTo(BeNil())
			Expect(*template.MasterSpec.Compute.GpuDevicesRequest).To(Equal(int32(1)))
			Expect(*template.MasterSpec.Compute.CpuCoreRequest).To(Equal(float64(2.0)))
			Expect(*template.MasterSpec.Compute.CpuMemoryRequest).To(Equal("8GB"))

			// Verify master ports and exposed URLs
			Expect(template.MasterSpec.Ports).To(HaveLen(1))
			Expect(template.MasterSpec.Ports[0].Container).To(Equal(int32(8888)))
			Expect(*template.MasterSpec.Ports[0].External).To(Equal(int32(30888)))
			Expect(template.MasterSpec.ExposedUrls).To(HaveLen(1))
			Expect(template.MasterSpec.ExposedUrls[0].Url).To(Equal("https://master.example.com"))
			Expect(template.MasterSpec.ExposedUrls[0].Name).To(Equal("master-ui"))
		})

		It("should not populate master spec when CombinedMasterSpec is not set", func() {
			image := "ubuntu:latest"

			rawTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "training-template-123",
					Name:         "training-template",
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					CreatedAt:    time.Now(),
				},
				Spec: &workloads_openapi.SupersetSpec{
					Image: *workloads_openapi.NewNullableString(&image),
				},
			}
			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{rawTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.GetByName(ctx, "training-template", projectId)

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
			Expect(*template.Spec.Image).To(Equal("ubuntu:latest"))
			// MasterSpec should be nil when CombinedMasterSpec is not set
			Expect(template.MasterSpec).To(BeNil())
		})
	})

	Context("FindByName", func() {
		It("should return template", func() {
			rawTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:        "template-123",
					Name:      templateName,
					ScopeType: workloads_openapi.SCOPETYPE_PROJECT,
					CreatedAt: time.Now(),
				},
			}

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{rawTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.FindByName(ctx, templateName, projectId, "")

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
			Expect(template.Meta.GetId()).To(Equal("template-123"))
		})

		It("should return error when no templates found", func() {
			// Setup List expectation - return empty
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.FindByName(ctx, templateName, projectId, "")

			Expect(err).NotTo(BeNil())
			Expect(template).To(BeNil())
		})

		It("should return highest priority template when multiple exist", func() {
			projectTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:        "project-template-id",
					Name:      templateName,
					ScopeType: workloads_openapi.SCOPETYPE_PROJECT,
					CreatedAt: time.Now(),
				},
			}

			clusterTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:        "cluster-template-id",
					Name:      templateName,
					ScopeType: workloads_openapi.SCOPETYPE_CLUSTER,
					CreatedAt: time.Now(),
				},
			}

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{clusterTemplate, projectTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.FindByName(ctx, templateName, projectId, "")

			Expect(err).To(BeNil())
			// Should return project template (highest priority)
			Expect(template).NotTo(BeNil())
			Expect(template.Meta.GetId()).To(Equal("project-template-id"))
		})

		It("should return template when workload type matches (Training template for training submit)", func() {
			trainingTemplate := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "training-template-id",
					Name:         templateName,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					CreatedAt:    time.Now(),
				},
			}

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{trainingTemplate})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			template, err := service.FindByName(ctx, templateName, projectId, "Training")

			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
			Expect(template.Meta.GetId()).To(Equal("training-template-id"))
		})

		It("should return error when workload type does not match (Workspace template for training submit)", func() {
			// API filters by workloadType=Training, but only Workspace templates exist
			// So the API returns empty results
			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&workloads_openapi.TemplatesListResponse{}, nil)

			template, err := service.FindByName(ctx, templateName, projectId, "Training")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("does not support training workloads"))
			Expect(template).To(BeNil())
		})

		It("should return 'not found' error when template doesn't exist (describe - no workload type)", func() {
			// API returns empty results when template doesn't exist
			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&workloads_openapi.TemplatesListResponse{}, nil)

			// Empty workloadType simulates describe command
			template, err := service.FindByName(ctx, templateName, projectId, "")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("template 'test-template' not found"))
			Expect(template).To(BeNil())
		})
	})

	Context("FetchTemplates", func() {
		It("should return all matching templates", func() {
			template1 := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:        "template-1",
					Name:      templateName,
					ScopeType: workloads_openapi.SCOPETYPE_PROJECT,
					CreatedAt: time.Now(),
				},
			}

			template2 := workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:        "template-2",
					Name:      templateName,
					ScopeType: workloads_openapi.SCOPETYPE_CLUSTER,
					CreatedAt: time.Now(),
				},
			}

			// Setup List expectation
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{template1, template2})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			templates, err := service.FetchTemplates(ctx, templateName, projectId, "")

			Expect(err).To(BeNil())
			Expect(templates).To(HaveLen(2))
			Expect(templates[0].Meta.Id).To(Equal("template-1"))
			Expect(templates[1].Meta.Id).To(Equal("template-2"))
		})

		It("should return empty slice when template not found", func() {
			// Setup List expectation - return empty
			listResponse := &workloads_openapi.TemplatesListResponse{}
			listResponse.SetEntries([]workloads_openapi.SupersetTemplate{})

			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(listResponse, nil)

			templates, err := service.FetchTemplates(ctx, templateName, projectId, "")

			Expect(err).To(BeNil())
			Expect(templates).To(BeEmpty())
		})

		It("should return error when API fails", func() {
			// Setup List expectation - return error
			mockClient.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("API error"))

			templates, err := service.FetchTemplates(ctx, templateName, projectId, "")

			Expect(err).NotTo(BeNil())
			Expect(templates).To(BeNil())
		})
	})
})
