/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"

	orgapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	clusterMock "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	backendMock "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	workloadsapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	workloadsMgr "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	workloadtest "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	manager "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerCluster "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

type gpuManifestRecorder struct {
	workloadsMgrMock
	createdIDs []string
}

func (m *gpuManifestRecorder) CreateWorkloadV2(_ context.Context, req manager.WorkloadV2CreateRequest) (*manager.WorkloadV2, *manager.Error) {
	workload := *m.createResponse
	// Stand in for the separate workloads-manager creation notification, which
	// populates the tracking row used by GET, suspend, and resume.
	workload.Metadata.Id = insertV2Workload(uint(workload.Metadata.TenantId), workload.Metadata.ClusterId, workload.Metadata.ProjectId, true)
	m.createdIDs = append(m.createdIDs, workload.Metadata.Id)
	workload.Metadata.Name = req.Metadata.GetName()
	workload.WorkloadManifest = req.Manifest
	m.getResponse = &manager.WorkloadV2GetResponse{Workload: workload}
	return &workload, nil
}

func (m *gpuManifestRecorder) UpdateWorkloadInternal(_ context.Context, _ string, req manager.WorkloadV2UpdateRequest) (*manager.WorkloadV2, *manager.Error) {
	m.getResponse.Workload.WorkloadManifest = req.Manifest
	return &m.getResponse.Workload, nil
}

type placementLookup struct {
	clusterMock.ClientInterface
	clusterID string
	nodes     []map[string]string
	err       error
}

func (c placementLookup) MatchGPUPlacement(_ context.Context, clusterID string, pools []string, labels map[string]string) (bool, error) {
	if c.err != nil {
		return false, c.err
	}
	if clusterID != c.clusterID || !slices.Contains(pools, "eligible") {
		return false, nil
	}
	for _, node := range c.nodes {
		matches := true
		for key, value := range labels {
			if node[key] != value {
				matches = false
				break
			}
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

var _ = Describe("GPU selection over the workload HTTP API", func() {
	var (
		ctx      context.Context
		router   http.Handler
		project  *testutils.RandomValues
		recorder *gpuManifestRecorder
	)
	const amd = "amd.com/gpu"
	const amdLabel = "amd.com/gpu.product-name"
	const nvidia = "nvidia.com/gpu"
	const nvidiaLabel = "nvidia.com/gpu.product"

	BeforeEach(func() {
		project = testutils.NewRandomValues()
		ctx = workloadtest.SetContext(project.TenantIdStr, workloadtest.TestUser, nil)
		auth := testing_crud.NewRandomsBasedAuthMockClient().With(project)
		auth.AllowedProjectsToContext(&ctx)
		authorization.SetMock(auth)
		orgunit.SetMock(auth.OrgUnitMock())
		policies.SetMock(policies.SimpleMockClientFactory{})
		assets.SetMock(assets.NewSimpleMock())
		cluster := workloadtest.BuildDisplayedCluster(project.ClusterId, "3.1.0", nil)
		checkerCluster.SetMock(checkerCluster.NewSimpleMockClient().WithCluster(project.ClusterId, cluster))
		checker.InitClusterCapabilitiesChecker("", workloadtest.GetDefaultClustersCacheConfig(), "")
		clusterservice.SetMock(placementLookup{ClientInterface: clusterMock.NewSimpleMock().WithDefaultNodePool(project.ClusterId).WithNodePools(project.ClusterId, "h100", "cpu").WithCluster(project.ClusterId, &cluster),
			clusterID: project.ClusterId, nodes: []map[string]string{
				{clusterconstants.AcceleratorResourceNodeLabel(nvidia): "true", nvidiaLabel: "H100"},
				{clusterconstants.AcceleratorResourceNodeLabel(amd): "true", amdLabel: "MI300X"},
				{clusterconstants.AcceleratorResourceNodeLabel("gpu.intel.com/i915"): "true", "hardware.example.com/model": "Flex170"},
			}})
		projectInfo := GetProjectResponse(*project)
		projectInfo.DefaultNodePools = []string{"h100"}
		for _, pool := range []struct{ name, id string }{{"h100", "eligible"}, {"cpu", "other"}} {
			resource := orgapi.Resources{}
			nodePool := orgapi.NewResourcesNodePool(pool.id)
			nodePool.SetName(pool.name)
			resource.SetNodePool(*nodePool)
			resource.Gpu.SetLimit(0)
			projectInfo.Resources = append(projectInfo.Resources, resource)
		}
		org_unit.SetMock(org_unit.NewSimpleMock().WithRandoms(project).WithProject(project.ProjectIdStr, projectInfo))
		recorder = &gpuManifestRecorder{workloadsMgrMock: workloadsMgrMock{createResponse: &manager.WorkloadV2{
			Metadata:     manager.WorkloadV2Metadata{ProjectId: project.ProjectIdStr, ProjectName: GetProjectResponse(*project).Name, TenantId: int32(project.TenantId), ClusterId: project.ClusterId, DepartmentId: project.DepartmentId, CreatedBy: workloadtest.TestUser},
			DesiredPhase: manager.DESIREDPHASE_RUNNING,
		}}}
		workloadsMgr.SetMock(recorder)
		settings := &backendMock.TenantSettingsClient{}
		router = api.NewRouter(api.NewWorkspacesAPIController(NewWorkspacesService(settings)), api.NewTrainingsAPIController(NewTrainingsService(settings)),
			api.NewInferencesAPIController(NewInferencesService(settings)), api.NewWorkloadTemplatesAPIController(NewTemplatesService()))
		DeferCleanup(func() {
			Expect(dao_crud.DbDeleteWorkloadV2Snapshots(ctx, project.TenantId, recorder.createdIDs, nil)).To(BeNil())
			workloadtest.CleanTables()
		})
	})

	request := func(method, path string, body any, wantStatus int) []byte {
		GinkgoHelper()
		req := httptest.NewRequest(method, path, makePayload(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		result := httptest.NewRecorder()
		router.ServeHTTP(result, req)
		Expect(result.Code).To(Equal(wantStatus), "%s %s: %s", method, path, result.Body.String())
		return result.Body.Bytes()
	}

	submit := func(collection, name string, spec any) workloadsapi.SupersetWorkload {
		GinkgoHelper()
		raw := request(http.MethodPost, "/api/v1/workloads/"+collection, map[string]any{
			"name": name, "projectId": project.ProjectIdStr, "clusterId": project.ClusterId, "spec": spec,
		}, http.StatusAccepted)
		var created workloadsapi.SupersetWorkload
		Expect(json.Unmarshal(raw, &created)).To(Succeed())
		return created
	}

	assertPlacement := func(resource, key, model string) {
		GinkgoHelper()
		var manifest struct {
			Spec struct{ Template corev1.PodTemplateSpec }
		}
		Expect(recorder.getResponse).NotTo(BeNil())
		raw, err := json.Marshal(recorder.getResponse.Workload.WorkloadManifest)
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Unmarshal(raw, &manifest)).To(Succeed())
		template := manifest.Spec.Template
		Expect(template.Spec.Containers).NotTo(BeEmpty())
		for _, list := range []corev1.ResourceList{template.Spec.Containers[0].Resources.Requests, template.Spec.Containers[0].Resources.Limits} {
			quantity, ok := list[corev1.ResourceName(resource)]
			Expect(ok).To(BeTrue())
			Expect(quantity.Value()).To(Equal(int64(1)))
		}
		Expect(template.Annotations).To(HaveKeyWithValue("run.ai/gpu-type-label", key+"="+model))
		Expect(template.Spec.Affinity).NotTo(BeNil())
		Expect(template.Spec.Affinity.NodeAffinity).NotTo(BeNil())
		required := template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		Expect(required).NotTo(BeNil())
		Expect(required.NodeSelectorTerms).NotTo(BeEmpty())
		for _, term := range required.NodeSelectorTerms {
			Expect(term.MatchExpressions).To(ContainElement(corev1.NodeSelectorRequirement{
				Key: key, Operator: corev1.NodeSelectorOpIn, Values: []string{model},
			}))
		}
	}

	DescribeTable("rejects invalid explicit device label keys before model selection",
		func(resource, model, key string) {
			accelerator := map[string]string{"deviceLabelKey": key}
			if resource != "" {
				accelerator["resourceName"] = resource
			}
			spec := map[string]any{"image": "ubuntu", "compute": map[string]int{"gpuDevicesRequest": 1}, "accelerator": accelerator}
			if model != "" {
				spec["gpuType"] = model
			}
			raw := request(http.MethodPost, "/api/v1/workloads/trainings", map[string]any{
				"name": "invalid-key", "clusterId": project.ClusterId, "projectId": project.ProjectIdStr, "spec": spec,
			}, http.StatusBadRequest)
			var problem workloadsapi.SubmissionErrorV2
			Expect(json.Unmarshal(raw, &problem)).To(Succeed())
			Expect(problem.Message).To(ContainSubstring("spec.accelerator.deviceLabelKey"))
			Expect(recorder.createdIDs).To(BeEmpty())
		},
		Entry("empty key with an otherwise discoverable model", amd, "MI300X", ""),
		Entry("empty key without a model", amd, "", ""),
		Entry("malformed key with Any", amd, "any", "invalid key"),
		Entry("malformed key without a resource or model", "", "", "invalid key"),
	)

	DescribeTable("checks H100 in selected or default pools even when the project GPU limit is zero",
		func(selected []string, expectWarning bool) {
			spec := map[string]any{"image": "ubuntu", "compute": map[string]int{"gpuDevicesRequest": 1}, "gpuType": "H100"}
			if selected != nil {
				spec["nodePools"] = selected
			}
			created := submit("trainings", "h100-placement", spec)
			Expect(created.Spec.GetGpuType()).To(Equal("H100"))
			if expectWarning {
				Expect(created.Warnings).To(HaveLen(1))
				Expect(created.Warnings[0].Code).To(Equal(gpuTypeNotFound))
			} else {
				Expect(created.Warnings).To(BeEmpty())
			}
		},
		Entry("selected zero-limit H100 pool", []string{"h100"}, false),
		Entry("H100 in the second selected pool", []string{"cpu", "h100"}, false),
		Entry("implicit project default is the zero-limit H100 pool", []string(nil), false),
		Entry("H100 exists only outside the selected pool", []string{"cpu"}, true),
	)

	It("warns when the model contradicts the resource presence label", func() {
		spec := map[string]any{"image": "ubuntu", "compute": map[string]int{"gpuDevicesRequest": 1},
			"gpuType": "other", "accelerator": map[string]string{"deviceLabelKey": clusterconstants.AcceleratorResourceNodeLabel(nvidia)}}
		created := submit("trainings", "conflicting-selector", spec)
		Expect(created.Warnings).To(HaveLen(1))
		Expect(created.Warnings[0].Code).To(Equal(gpuTypeNotFound))
	})
	It("requires an explicit label key for a non-NVIDIA model", func() {
		raw := request(http.MethodPost, "/api/v1/workloads/trainings", map[string]any{
			"name": "explicit-key", "clusterId": project.ClusterId, "projectId": project.ProjectIdStr,
			"spec": map[string]any{"image": "ubuntu", "compute": map[string]int{"gpuDevicesRequest": 1}, "gpuType": "MI300X", "accelerator": map[string]string{"resourceName": amd}},
		}, http.StatusBadRequest)
		Expect(string(raw)).To(ContainSubstring("accelerator.deviceLabelKey"))
	})
	It("preserves placement when inventory cannot be checked", func() {
		clusterservice.SetMock(placementLookup{ClientInterface: clusterservice.Get(), err: errors.New("unavailable")})
		created := submit("trainings", "inventory-unavailable", map[string]any{"image": "ubuntu", "compute": map[string]int{"gpuDevicesRequest": 1}, "gpuType": "H100"})
		Expect(created.Warnings).To(HaveLen(1))
		Expect(created.Warnings[0].Code).To(Equal(gpuInventoryUnavailable))
		assertPlacement(nvidia, nvidiaLabel, "H100")
	})

	DescribeTable("checks explicit device label keys and preserves unavailable selections",
		func(requestKey any, expectedKey string, expectWarning bool) {
			created := submit("trainings", "device-key", map[string]any{"image": "ubuntu",
				"compute": map[string]int{"gpuDevicesRequest": 1}, "gpuType": "MI300X",
				"accelerator": map[string]any{"resourceName": amd, "deviceLabelKey": requestKey}})
			if expectWarning {
				Expect(created.Warnings).To(HaveLen(1))
				Expect(created.Warnings[0].Code).To(Equal("GPU_TYPE_NOT_FOUND_IN_PROJECT_POOLS"))
			} else {
				Expect(created.Warnings).To(BeEmpty())
			}
			assertPlacement(amd, expectedKey, "MI300X")
		},
		Entry("explicit AMD product label", amdLabel, amdLabel, false),
		Entry("valid unavailable label is preserved with a warning", "lab.example.com/model", "lab.example.com/model", true),
	)

	It("matches nonstandard explicit resources with and without a custom model", func() {
		for _, model := range []string{"", "Flex170"} {
			resource := map[string]string{"resourceName": "gpu.intel.com/i915"}
			spec := map[string]any{"image": "ubuntu", "compute": map[string]int{"gpuDevicesRequest": 1}, "accelerator": resource}
			if model != "" {
				spec["gpuType"] = model
				resource["deviceLabelKey"] = "hardware.example.com/model"
			}
			created := submit("trainings", "typed-custom-"+strings.ToLower(model)+"gpu", spec)
			Expect(created.Warnings).To(BeEmpty())
			selection := created.Spec.GetAccelerator()
			Expect(selection.GetResourceName()).To(Equal("gpu.intel.com/i915"))
		}
	})

	DescribeTable("accepts JSON and submits the resolved Pod template",
		func(collection, model, resource, key string, warning bool) {
			spec := map[string]any{"image": "ubuntu", "compute": map[string]int{"gpuDevicesRequest": 1}, "gpuType": model}
			if resource != nvidia {
				spec["accelerator"] = map[string]string{"resourceName": resource, "deviceLabelKey": key}
			}
			if collection == "inferences" {
				spec["servingPort"] = map[string]any{"container": 8080, "protocol": "http"}
			}
			created := submit(collection, "gpu-selection", spec)
			Expect(created.Spec.GetGpuType()).To(Equal(model))
			if resource != nvidia {
				Expect(created.Spec.Accelerator.Get().GetDeviceLabelKey()).To(Equal(key))
			}
			if warning {
				Expect(created.Warnings).To(HaveLen(1))
				Expect(created.Warnings[0].Code).To(Equal("GPU_TYPE_NOT_FOUND_IN_PROJECT_POOLS"))
				Expect(created.Warnings[0].Field).To(Equal("spec.gpuType"))
			} else {
				Expect(created.Warnings).To(BeEmpty())
			}
			assertPlacement(resource, key, model)
		},
		Entry("workspace defaults an H100 to NVIDIA", "workspaces", "H100", nvidia, nvidiaLabel, false),
		Entry("training honors the explicit AMD device label", "trainings", "MI300X", amd, amdLabel, false),
		Entry("inference warns for AMD H100 even when the project has NVIDIA H100", "inferences", "H100", amd, amdLabel, true),
	)
	It("retains AMD placement from template through read, copy and resume", func() {
		const templates = "/api/v1/workload-templates/workspaces"
		const workspaces = "/api/v1/workloads/workspaces"
		spec := map[string]any{
			"image": "ubuntu", "compute": map[string]int{"gpuDevicesRequest": 1},
			"gpuType": "MI300X", "accelerator": map[string]string{"resourceName": amd, "deviceLabelKey": amdLabel},
		}
		raw := request(http.MethodPost, templates, map[string]any{
			"meta": map[string]string{"name": "amd-workspace-template", "scopeType": "project", "scopeId": project.ProjectIdStr}, "spec": spec,
		}, http.StatusOK)
		var template api.WorkspaceTemplate
		Expect(json.Unmarshal(raw, &template)).To(Succeed())
		Expect(template.Meta.Id).NotTo(BeEmpty())
		raw = request(http.MethodPost, workspaces, map[string]any{
			"name": "template-workspace", "projectId": project.ProjectIdStr, "clusterId": project.ClusterId, "templateId": template.Meta.Id, "spec": map[string]any{},
		}, http.StatusAccepted)
		var created workloadsapi.Workspace
		Expect(json.Unmarshal(raw, &created)).To(Succeed())
		assertPlacement(amd, amdLabel, "MI300X")
		path := workspaces + "/" + created.WorkloadId
		read := func() *workloadsapi.WorkspaceSpecSpec {
			GinkgoHelper()
			var saved struct {
				Spec *workloadsapi.WorkspaceSpecSpec `json:"spec"`
			}
			Expect(json.Unmarshal(request(http.MethodGet, path, nil, http.StatusOK), &saved)).To(Succeed())
			Expect(saved.Spec).NotTo(BeNil())
			Expect(saved.Spec.GetGpuType()).To(Equal("MI300X"))
			Expect(saved.Spec.Accelerator.Get().GetResourceName()).To(Equal(amd))
			Expect(saved.Spec.Accelerator.Get().GetDeviceLabelKey()).To(Equal(amdLabel))
			return saved.Spec
		}
		read()
		request(http.MethodPost, path+"/suspend", nil, http.StatusAccepted)
		read()
		request(http.MethodPost, path+"/resume", nil, http.StatusAccepted)
		Expect(recorder.lastIds).To(ConsistOf(created.WorkloadId))

		copied := submit("workspaces", "copied-workspace", read())
		Expect(copied.WorkloadId).NotTo(Equal(created.WorkloadId))
		Expect(copied.Warnings).To(BeEmpty())
		assertPlacement(amd, amdLabel, "MI300X")
	})

	It("recomposes inference placement from the saved spec during a partial update", func() {
		created := submit("inferences", "update-inference", map[string]any{
			"image": "ubuntu", "servingPort": map[string]any{"container": 8080}, "compute": map[string]int{"gpuDevicesRequest": 1},
			"gpuType": "MI300X", "accelerator": map[string]string{"resourceName": amd, "deviceLabelKey": amdLabel},
		})
		path := "/api/v1/workloads/inferences/" + created.WorkloadId
		raw := request(http.MethodPatch, path, map[string]any{"spec": map[string]string{"image": "ubuntu:next"}}, http.StatusAccepted)
		var updated workloadsapi.SupersetWorkload
		Expect(json.Unmarshal(raw, &updated)).To(Succeed())
		Expect(updated.Spec.GetImage()).To(Equal("ubuntu:next"))
		Expect(updated.Spec.GetGpuType()).To(Equal("MI300X"))
		Expect(updated.Spec.Accelerator.Get().GetDeviceLabelKey()).To(Equal(amdLabel))
		assertPlacement(amd, amdLabel, "MI300X")
	})

	It("checks the saved GPU type when only the inference node pools change", func() {
		created := submit("inferences", "update-pools", map[string]any{
			"image": "ubuntu", "servingPort": map[string]any{"container": 8080},
			"compute": map[string]int{"gpuDevicesRequest": 1}, "gpuType": "H100",
		})
		Expect(created.Warnings).To(BeEmpty())
		raw := request(http.MethodPatch, "/api/v1/workloads/inferences/"+created.WorkloadId,
			map[string]any{"spec": map[string]any{"nodePools": []string{"cpu"}}}, http.StatusAccepted)
		var updated workloadsapi.SupersetWorkload
		Expect(json.Unmarshal(raw, &updated)).To(Succeed())
		Expect(updated.Warnings).To(HaveLen(1))
		Expect(updated.Warnings[0].Code).To(Equal(gpuTypeNotFound))
	})
})
