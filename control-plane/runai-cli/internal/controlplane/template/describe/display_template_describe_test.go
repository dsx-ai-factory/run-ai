/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/describe"
	workloads_openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplateDisplayer", func() {
	var (
		displayer      *describe.TemplateDisplayer
		template       *describe.Template
		originalStdout *os.File
		r              *os.File
		w              *os.File
	)

	// Helper function to capture stdout
	captureOutput := func(f func() error) (string, error) {
		originalStdout = os.Stdout
		var err error
		r, w, err = os.Pipe()
		if err != nil {
			return "", err
		}
		os.Stdout = w

		execErr := f()

		w.Close()
		os.Stdout = originalStdout

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		if err != nil {
			return "", err
		}

		return buf.String(), execErr
	}

	// Helper to create a basic template with proper API Raw type
	createTemplate := func(id, name, templateType, scope, scopeID, createdBy string) *describe.Template {
		createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		// Create raw API template
		rawTemplate := &workloads_openapi.SupersetTemplate{
			Meta: workloads_openapi.TemplateMetaFields{
				Id:        id,
				Name:      name,
				ScopeType: workloads_openapi.ScopeType(scope),
				ScopeId:   scopeID,
				CreatedAt: createdAt,
				CreatedBy: createdBy,
			},
		}
		rawTemplate.SetCombinedSpec(workloads_openapi.SupersetSpec{})

		tmpl := &describe.Template{
			Meta: describe.TemplateMeta{
				ID:        id,
				Name:      name,
				Type:      templateType,
				Scope:     scope,
				ScopeID:   scopeID,
				CreatedAt: createdAt,
				CreatedBy: createdBy,
			},
			Spec: &describe.TemplateSpec{},
			Raw:  rawTemplate,
		}
		return tmpl
	}

	Context("when displaying a single template", func() {
		BeforeEach(func() {
			// Create a single template (highest priority after sorting)
			template = createTemplate("template-1", "my-template", "Training", "project", "project-123", "user@example.com")
			displayer = describe.NewTemplateDisplayer(template)
		})

		Describe("AsTable", func() {
			It("should display template in table format", func() {
				output, err := captureOutput(func() error {
					return displayer.AsTable()
				})

				Expect(err).To(BeNil())

				// Verify no multi-template header
				Expect(output).NotTo(ContainSubstring("Found"))
				Expect(output).NotTo(ContainSubstring("Template 1 of"))

				// Verify template is displayed
				Expect(output).To(ContainSubstring("template-1"))
				Expect(output).To(ContainSubstring("my-template"))
				Expect(output).To(ContainSubstring("project"))
				Expect(output).To(ContainSubstring("user@example.com"))
			})
		})

		Describe("AsJSON", func() {
			It("should display template in JSON format", func() {
				output, err := captureOutput(func() error {
					return displayer.AsJSON()
				})

				Expect(err).To(BeNil())

				// Verify JSON contains expected fields
				Expect(output).To(ContainSubstring("template-1"))
				Expect(output).To(ContainSubstring("my-template"))
				Expect(output).To(ContainSubstring("project"))
			})
		})

		Describe("AsYAML", func() {
			It("should display template in YAML format", func() {
				output, err := captureOutput(func() error {
					return displayer.AsYAML()
				})

				Expect(err).To(BeNil())

				// Verify YAML contains expected fields
				Expect(output).To(ContainSubstring("template-1"))
				Expect(output).To(ContainSubstring("my-template"))
				Expect(output).To(ContainSubstring("project"))
			})
		})
	})

	Context("when displaying template with detailed spec", func() {
		BeforeEach(func() {
			createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
			updatedAt := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

			image := "ubuntu:22.04"
			command := "python train.py"
			args := "--epochs 100"
			priorityClass := "high"

			gpuDevices := int32(2)
			gpuMemRequest := "16Gi"
			gpuMemLimit := "32Gi"
			cpuCoreRequest := 4.0
			cpuCoreLimit := 8.0
			cpuMemRequest := "32Gi"
			cpuMemLimit := "64Gi"

			description := "Test description"
			clusterId := "cluster-123"
			projectId := int32(456)
			departmentId := "dept-789"
			updatedBy := "admin@example.com"

			// Create raw API template with all fields
			compute := workloads_openapi.NewWorkspaceSpecSpecAllOfCompute()
			compute.SetGpuDevicesRequest(gpuDevices)
			compute.SetGpuMemoryRequest(gpuMemRequest)
			compute.SetGpuMemoryLimit(gpuMemLimit)
			compute.SetCpuCoreRequest(cpuCoreRequest)
			compute.SetCpuCoreLimit(cpuCoreLimit)
			compute.SetCpuMemoryRequest(cpuMemRequest)
			compute.SetCpuMemoryLimit(cpuMemLimit)

			envVar := workloads_openapi.EnvironmentVariable{}
			envVar.Name = *workloads_openapi.NewNullableString(ptr("EPOCHS"))
			envVar.Value = *workloads_openapi.NewNullableString(ptr("100"))

			envVar2 := workloads_openapi.EnvironmentVariable{}
			envVar2.Name = *workloads_openapi.NewNullableString(ptr("MODEL"))
			envVar2.Value = *workloads_openapi.NewNullableString(ptr("resnet50"))

			port := workloads_openapi.NewPort()
			port.SetContainer(int32(8080))

			exposedUrl := workloads_openapi.NewExposedUrl()
			exposedUrl.SetUrl("https://jupyter.example.com")
			exposedUrl.SetName("Jupyter")

			combinedSpec := workloads_openapi.SupersetSpec{
				Image:                *workloads_openapi.NewNullableString(&image),
				Command:              *workloads_openapi.NewNullableString(&command),
				Args:                 *workloads_openapi.NewNullableString(&args),
				Preemptibility:       *workloads_openapi.NewNullablePreemptibility(ptr(workloads_openapi.PREEMPTIBILITY_PREEMPTIBLE)),
				PriorityClass:        *workloads_openapi.NewNullableString(&priorityClass),
				NodePools:            []string{"gpu-pool"},
				EnvironmentVariables: []workloads_openapi.EnvironmentVariable{envVar, envVar2},
				Compute:              *workloads_openapi.NewNullableWorkspaceSpecSpecAllOfCompute(compute),
				Ports:                []workloads_openapi.Port{*port},
				ExposedUrls:          []workloads_openapi.ExposedUrl{*exposedUrl},
			}

			rawTemplate := &workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "template-1",
					Name:         "my-template",
					Description:  *workloads_openapi.NewNullableString(&description),
					WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					ScopeId:      "project-123",
					ClusterId:    *workloads_openapi.NewNullableString(&clusterId),
					ProjectId:    *workloads_openapi.NewNullableInt32(&projectId),
					DepartmentId: *workloads_openapi.NewNullableString(&departmentId),
					CreatedAt:    createdAt,
					CreatedBy:    "user@example.com",
					UpdatedAt:    *workloads_openapi.NewNullableTime(&updatedAt),
					UpdatedBy:    *workloads_openapi.NewNullableString(&updatedBy),
				},
			}
			rawTemplate.SetCombinedSpec(combinedSpec)

			preemptible := true
			template = &describe.Template{
				Meta: describe.TemplateMeta{
					ID:           "template-1",
					Name:         "my-template",
					Description:  &description,
					Type:         "Training",
					Scope:        "project",
					ScopeID:      "project-123",
					ClusterID:    &clusterId,
					ProjectID:    &projectId,
					DepartmentID: &departmentId,
					CreatedAt:    createdAt,
					CreatedBy:    "user@example.com",
					UpdatedAt:    &updatedAt,
					UpdatedBy:    &updatedBy,
				},
				Spec: &describe.TemplateSpec{
					Image:         &image,
					Command:       &command,
					Args:          &args,
					PriorityClass: &priorityClass,
					Preemptible:   &preemptible,
					NodePools:     []string{"gpu-pool"},
					EnvironmentVariables: map[string]string{
						"EPOCHS": "100",
						"MODEL":  "resnet50",
					},
					Compute: &describe.Compute{
						GpuDevicesRequest: &gpuDevices,
						GpuMemoryRequest:  &gpuMemRequest,
						GpuMemoryLimit:    &gpuMemLimit,
						CpuCoreRequest:    &cpuCoreRequest,
						CpuCoreLimit:      &cpuCoreLimit,
						CpuMemoryRequest:  &cpuMemRequest,
						CpuMemoryLimit:    &cpuMemLimit,
					},
					Ports: []describe.Port{
						{Container: 8080, External: nil},
					},
					ExposedUrls: []describe.ExposedUrl{
						{Url: "https://jupyter.example.com", Name: "Jupyter"},
					},
				},
				Raw: rawTemplate,
			}
			displayer = describe.NewTemplateDisplayer(template)
		})

		Describe("AsTable", func() {
			It("should display all template fields", func() {
				output, err := captureOutput(func() error {
					return displayer.AsTable()
				})

				Expect(err).To(BeNil())

				// Meta fields
				Expect(output).To(ContainSubstring("template-1"))
				Expect(output).To(ContainSubstring("my-template"))
				Expect(output).To(ContainSubstring("Test description"))
				Expect(output).To(ContainSubstring("project"))
				Expect(output).To(ContainSubstring("456"))
				Expect(output).To(ContainSubstring("dept-789"))

				// Spec fields
				Expect(output).To(ContainSubstring("ubuntu:22.04"))
				Expect(output).To(ContainSubstring("python train.py"))
				Expect(output).To(ContainSubstring("--epochs 100"))
				Expect(output).To(ContainSubstring("high"))
				Expect(output).To(ContainSubstring("true"))
				Expect(output).To(ContainSubstring("gpu-pool"))

				// Environment variables
				Expect(output).To(ContainSubstring("EPOCHS"))
				Expect(output).To(ContainSubstring("100"))
				Expect(output).To(ContainSubstring("MODEL"))
				Expect(output).To(ContainSubstring("resnet50"))

				// Compute section
				Expect(output).To(ContainSubstring("Compute Resources"))
				Expect(output).To(ContainSubstring("GPU Compute"))
				Expect(output).To(ContainSubstring("CPU Cores"))
				Expect(output).To(ContainSubstring("GPU Memory"))
				Expect(output).To(ContainSubstring("CPU Memory"))

				// Networks section
				Expect(output).To(ContainSubstring("Networks"))
				Expect(output).To(ContainSubstring("8080"))
				Expect(output).To(ContainSubstring("jupyter.example.com"))
			})
		})

		Describe("AsJSON", func() {
			It("should display complete template in JSON format", func() {
				output, err := captureOutput(func() error {
					return displayer.AsJSON()
				})

				Expect(err).To(BeNil())

				// Verify JSON contains expected fields
				Expect(output).To(ContainSubstring("template-1"))
				Expect(output).To(ContainSubstring("my-template"))
				Expect(output).To(ContainSubstring("Test description"))
				Expect(output).To(ContainSubstring("ubuntu:22.04"))
				Expect(output).To(ContainSubstring("python train.py"))
				Expect(output).To(ContainSubstring("gpu-pool"))
				Expect(output).To(ContainSubstring("jupyter.example.com"))
			})
		})

		Describe("AsYAML", func() {
			It("should display complete template in YAML format", func() {
				output, err := captureOutput(func() error {
					return displayer.AsYAML()
				})

				Expect(err).To(BeNil())

				// Verify YAML contains expected fields
				Expect(output).To(ContainSubstring("template-1"))
				Expect(output).To(ContainSubstring("my-template"))
				Expect(output).To(ContainSubstring("Test description"))
				Expect(output).To(ContainSubstring("ubuntu:22.04"))
				Expect(output).To(ContainSubstring("python train.py"))
				Expect(output).To(ContainSubstring("gpu-pool"))
				Expect(output).To(ContainSubstring("jupyter.example.com"))
			})
		})
	})

	Context("when displaying distributed template with different master spec", func() {
		BeforeEach(func() {
			createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

			// Worker spec values
			workerImage := "worker:latest"
			workerCommand := "python worker.py"

			// Master spec values - intentionally different from worker
			masterImage := "master:latest"
			masterCommand := "python master.py"
			masterArgs := "--master-mode"
			masterWorkingDir := "/app/master"

			// Worker compute
			workerGpuDevices := int32(4)
			workerCpuCoreRequest := 8.0
			workerCpuMemRequest := "64Gi"

			// Master compute - different from worker
			masterGpuDevices := int32(1)
			masterCpuCoreRequest := 2.0
			masterCpuMemRequest := "16Gi"

			// Create raw API template with master spec
			workerCompute := workloads_openapi.NewWorkspaceSpecSpecAllOfCompute()
			workerCompute.SetGpuDevicesRequest(workerGpuDevices)
			workerCompute.SetCpuCoreRequest(workerCpuCoreRequest)
			workerCompute.SetCpuMemoryRequest(workerCpuMemRequest)

			workerEnvVar := workloads_openapi.EnvironmentVariable{}
			workerEnvVar.Name = *workloads_openapi.NewNullableString(ptr("WORKER_VAR"))
			workerEnvVar.Value = *workloads_openapi.NewNullableString(ptr("worker_value"))

			combinedSpec := workloads_openapi.SupersetSpec{
				Image:                *workloads_openapi.NewNullableString(&workerImage),
				Command:              *workloads_openapi.NewNullableString(&workerCommand),
				NodePools:            []string{"worker-pool"},
				EnvironmentVariables: []workloads_openapi.EnvironmentVariable{workerEnvVar},
				Compute:              *workloads_openapi.NewNullableWorkspaceSpecSpecAllOfCompute(workerCompute),
			}

			// Master spec with different values
			masterCompute := workloads_openapi.NewWorkspaceSpecSpecAllOfCompute()
			masterCompute.SetGpuDevicesRequest(masterGpuDevices)
			masterCompute.SetCpuCoreRequest(masterCpuCoreRequest)
			masterCompute.SetCpuMemoryRequest(masterCpuMemRequest)

			masterEnvVar := workloads_openapi.EnvironmentVariable{}
			masterEnvVar.Name = *workloads_openapi.NewNullableString(ptr("MASTER_VAR"))
			masterEnvVar.Value = *workloads_openapi.NewNullableString(ptr("master_value"))

			masterPort := workloads_openapi.NewPort()
			masterPort.SetContainer(int32(9090))

			masterUrl := workloads_openapi.NewExposedUrl()
			masterUrl.SetUrl("https://master.example.com")
			masterUrl.SetName("master-ui")

			combinedMasterSpec := workloads_openapi.NewMasterTemplateSpec()
			combinedMasterSpec.SetImage(masterImage)
			combinedMasterSpec.SetCommand(masterCommand)
			combinedMasterSpec.SetArgs(masterArgs)
			combinedMasterSpec.SetWorkingDir(masterWorkingDir)
			combinedMasterSpec.SetNodePools([]string{"master-pool"})
			combinedMasterSpec.SetEnvironmentVariables([]workloads_openapi.EnvironmentVariable{masterEnvVar})
			combinedMasterSpec.SetCompute(*masterCompute)
			combinedMasterSpec.SetPorts([]workloads_openapi.Port{*masterPort})
			combinedMasterSpec.SetExposedUrls([]workloads_openapi.ExposedUrl{*masterUrl})

			rawTemplate := &workloads_openapi.SupersetTemplate{
				Meta: workloads_openapi.TemplateMetaFields{
					Id:           "distributed-template-1",
					Name:         "distributed-training",
					WorkloadType: workloads_openapi.WORKLOADTYPE_DISTRIBUTED,
					ScopeType:    workloads_openapi.SCOPETYPE_PROJECT,
					ScopeId:      "project-123",
					CreatedAt:    createdAt,
					CreatedBy:    "user@example.com",
				},
			}
			rawTemplate.SetCombinedSpec(combinedSpec)
			rawTemplate.SetCombinedMasterSpec(*combinedMasterSpec)

			template = &describe.Template{
				Meta: describe.TemplateMeta{
					ID:        "distributed-template-1",
					Name:      "distributed-training",
					Type:      "Distributed",
					Scope:     "project",
					ScopeID:   "project-123",
					CreatedAt: createdAt,
					CreatedBy: "user@example.com",
				},
				Spec: &describe.TemplateSpec{
					Image:     &workerImage,
					Command:   &workerCommand,
					NodePools: []string{"worker-pool"},
					EnvironmentVariables: map[string]string{
						"WORKER_VAR": "worker_value",
					},
					Compute: &describe.Compute{
						GpuDevicesRequest: &workerGpuDevices,
						CpuCoreRequest:    &workerCpuCoreRequest,
						CpuMemoryRequest:  &workerCpuMemRequest,
					},
				},
				MasterSpec: &describe.MasterSpec{
					Image:      &masterImage,
					Command:    &masterCommand,
					Args:       &masterArgs,
					WorkingDir: &masterWorkingDir,
					NodePools:  []string{"master-pool"},
					EnvironmentVariables: map[string]string{
						"MASTER_VAR": "master_value",
					},
					Compute: &describe.Compute{
						GpuDevicesRequest: &masterGpuDevices,
						CpuCoreRequest:    &masterCpuCoreRequest,
						CpuMemoryRequest:  &masterCpuMemRequest,
					},
					Ports: []describe.Port{
						{Container: 9090},
					},
					ExposedUrls: []describe.ExposedUrl{
						{Url: "https://master.example.com", Name: "master-ui"},
					},
				},
				Raw: rawTemplate,
			}
			displayer = describe.NewTemplateDisplayer(template)
		})

		Describe("AsTable", func() {
			It("should display both worker and master spec fields when they differ", func() {
				output, err := captureOutput(func() error {
					return displayer.AsTable()
				})

				Expect(err).To(BeNil())

				// Verify worker spec fields are displayed
				Expect(output).To(ContainSubstring("worker:latest"))
				Expect(output).To(ContainSubstring("python worker.py"))
				Expect(output).To(ContainSubstring("worker-pool"))
				Expect(output).To(ContainSubstring("WORKER_VAR"))
				Expect(output).To(ContainSubstring("worker_value"))

				// Verify worker compute section
				Expect(output).To(ContainSubstring("Compute Resources"))

				// Verify master configuration section header
				Expect(output).To(ContainSubstring("Master Configuration"))

				// Verify master spec fields are displayed (different from worker)
				Expect(output).To(ContainSubstring("master:latest"))
				Expect(output).To(ContainSubstring("python master.py"))
				Expect(output).To(ContainSubstring("--master-mode"))
				Expect(output).To(ContainSubstring("/app/master"))
				Expect(output).To(ContainSubstring("master-pool"))
				Expect(output).To(ContainSubstring("MASTER_VAR"))
				Expect(output).To(ContainSubstring("master_value"))

				// Verify master compute section
				Expect(output).To(ContainSubstring("Master Compute"))

				// Verify master network section
				Expect(output).To(ContainSubstring("Master Networks"))
				Expect(output).To(ContainSubstring("9090"))
				Expect(output).To(ContainSubstring("master.example.com"))
				Expect(output).To(ContainSubstring("master-ui"))
			})

			It("should display master compute resources separately from worker", func() {
				output, err := captureOutput(func() error {
					return displayer.AsTable()
				})

				Expect(err).To(BeNil())

				// Both compute sections should appear
				Expect(output).To(ContainSubstring("Compute Resources"))
				Expect(output).To(ContainSubstring("Master Compute"))

				// Worker has 4 GPUs, master has 1 GPU - both should be visible
				// Note: We can't easily distinguish which "4" or "1" belongs to which section
				// in a simple substring check, but the sections should both be present
				Expect(output).To(ContainSubstring("GPU Compute"))
			})
		})

		Describe("AsJSON", func() {
			It("should include masterSpec in JSON output", func() {
				output, err := captureOutput(func() error {
					return displayer.AsJSON()
				})

				Expect(err).To(BeNil())

				// Verify JSON contains masterSpec section
				Expect(output).To(ContainSubstring("masterSpec"))
				Expect(output).To(ContainSubstring("master:latest"))
				Expect(output).To(ContainSubstring("python master.py"))
				Expect(output).To(ContainSubstring("master-pool"))
			})
		})

		Describe("AsYAML", func() {
			It("should include masterSpec in YAML output", func() {
				output, err := captureOutput(func() error {
					return displayer.AsYAML()
				})

				Expect(err).To(BeNil())

				// Verify YAML contains masterSpec section
				Expect(output).To(ContainSubstring("masterSpec"))
				Expect(output).To(ContainSubstring("master:latest"))
				Expect(output).To(ContainSubstring("python master.py"))
				Expect(output).To(ContainSubstring("master-pool"))
			})
		})
	})
})
