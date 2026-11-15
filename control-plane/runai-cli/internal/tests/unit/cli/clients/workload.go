/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	cpworkloadprioritiesmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/priorities"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/priorities"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	apifilter "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	cpworkloadmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/distributed"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/distributedinference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/inference"
	nodes "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/training"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload"
	spec2 "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workspace"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/golang/mock/gomock"
	"k8s.io/utils/ptr"
)

const (
	WorkloadName01  = "workload-01"
	WorkloadName02  = "workload-02"
	WorkloadName03  = "workload-03"
	WorkloadName021 = "workload-021"
	ProjectNameP301 = "p301"
	ProjectIdP301   = "301"
	ProjectNameP302 = "p302"
	ProjectIdP302   = "302"
	clusterId       = "uuid-cluster-1"
	UuidExist       = "uuid_exist"
	UuidDontExist   = "dont_exist"
)

var (
	ClusterPrioritiesNames = []string{"low", "very-very-high"}
	WorkspaceWorkloadName  = WorkloadName01
	WorkspaceProjectName   = ProjectNameP301
	workload01             = openapi.Workload{
		TenantId:    1,
		RunningPods: 0,
		Phase:       openapi.PHASE_CREATING,
		Name:        WorkloadName01,
		Id:          UuidExist,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP301,
		ProjectId:   ProjectIdP301,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP301),
		Type:        api.Workspace.String(),
		CreatedAt:   time.Unix(10, 0),
		RequestedPods: &openapi.RequestedPods{
			Number: *openapi.NewNullableInt32(ptr.To[int32](4)),
		},
	}

	TrainingWorkloadName = WorkloadName02
	TrainingProjectName  = ProjectNameP302
	workload02           = openapi.Workload{
		TenantId:    1,
		RunningPods: 1,
		Phase:       openapi.PHASE_CREATING,
		Name:        WorkloadName02,
		Id:          UuidDontExist,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP302,
		ProjectId:   ProjectIdP302,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP302),
		Type:        api.Training.String(),
		CreatedAt:   time.Unix(11, 0),
		RequestedPods: &openapi.RequestedPods{
			Parallelism: *openapi.NewNullableInt32(ptr.To[int32](5)),
		},
	}

	TrainingWorkloadName03 = WorkloadName03
	workload03             = openapi.Workload{
		TenantId:    1,
		RunningPods: 1,
		Phase:       openapi.PHASE_CREATING,
		Name:        TrainingWorkloadName03,
		Id:          UuidDontExist,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP302,
		ProjectId:   ProjectIdP302,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP302),
		Type:        api.Training.String(),
		CreatedAt:   time.Unix(11, 0),
		RequestedPods: &openapi.RequestedPods{
			Completions: *openapi.NewNullableInt32(ptr.To[int32](5)),
		},
	}

	workload021 = openapi.Workload{
		TenantId:    1,
		RunningPods: 0,
		Phase:       openapi.PHASE_CREATING,
		Name:        WorkloadName021,
		Id:          UuidExist,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP302,
		ProjectId:   ProjectIdP302,
		CreatedAt:   time.Unix(10, 0),
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP302),
		Type:        api.Training.String(),
	}

	WorkloadList = []openapi.Workload{
		workload01,
		workload02,
	}

	specMap = map[string]json.Marshaler{}

	mockNodesService *nodes.BackendNodeService
)

type MockWorkloadFilter struct {
	Filter   []string
	Response openapi.GetWorkloads200Response
}

type MockWorkloadBatch struct {
	Ids      []string
	Response *openapi.BatchResponse
	Err      error
}

func MockWorkloadByList(mockCtrl *gomock.Controller, filterList []MockWorkloadFilter) {
	mockWsPager := cpworkloadmocks.NewMockPaginator(mockCtrl)
	for _, filter := range filterList {
		mockWsPager.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(filter.Filter), false).Return(&filter.Response, nil).AnyTimes()
	}
	workload.SetPaginator(mockWsPager)
}

func MockWorkloadBatchDelete(mockCtrl *gomock.Controller, batchList []MockWorkloadBatch) {
	mockWsClient := cpworkloadmocks.NewMockBatcher(mockCtrl)
	for _, batch := range batchList {
		mockWsClient.EXPECT().BatchDelete(gomock.Any(), gomock.InAnyOrder(batch.Ids)).Return(batch.Response, batch.Err).AnyTimes()
	}
	workload.SetBatcher(mockWsClient)
}

func MockWorkloadPriorities(mockCtrl *gomock.Controller) {
	mockWsPrioritiesClient := cpworkloadprioritiesmocks.NewMockLister(mockCtrl)
	prioritiesList := []openapi.ClusterPriority{}
	for _, clusterPriorityName := range ClusterPrioritiesNames {
		prioritiesList = append(prioritiesList, openapi.ClusterPriority{Name: clusterPriorityName})
	}
	mockWsPrioritiesClient.EXPECT().ListPriorities(gomock.Any(), gomock.Any()).Return(&openapi.GetWorkloadClusterPriorities200Response{
		Priorities: prioritiesList,
	}, nil).AnyTimes()
	priorities.SetClient(mockWsPrioritiesClient)
}

func MockWorkload(mockCtrl *gomock.Controller, filterList []string) {
	var emptyStrList []string
	emptyStrList = append(emptyStrList, "clusterId==uuid-cluster-1")
	emptyStrList = append(emptyStrList, filterList...)
	mockWsClient := cpworkloadmocks.NewMockPaginator(mockCtrl)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(emptyStrList), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: WorkloadList,
	}, nil).AnyTimes()

	var projectFilter301 = []string{"projectName==p301", "clusterId==uuid-cluster-1"}
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(projectFilter301), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload01},
	}, nil).AnyTimes()
	projectFilter301 = append(projectFilter301, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(projectFilter301), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload01},
	}, nil).AnyTimes()

	var projectFilter301wl01 = []string{"name==workload-01", "projectName==p301", "clusterId==uuid-cluster-1"}
	projectFilter301wl01 = append(projectFilter301wl01, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(projectFilter301wl01), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload01},
	}, nil).AnyTimes()

	var projectFilter301wl02 = []string{"name==workload-02", "projectName==p301", "clusterId==uuid-cluster-1"}
	projectFilter301wl02 = append(projectFilter301wl02, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(projectFilter301wl02), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload02},
	}, nil).AnyTimes()

	// Mock for scale command lookup
	var projectFilter301DistInfScale = []string{"name==dist-inf-scale", "projectName==p301", "clusterId==uuid-cluster-1"}
	projectFilter301DistInfScale = append(projectFilter301DistInfScale, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(projectFilter301DistInfScale), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload01}, // Return a valid workload for scale tests
	}, nil).AnyTimes()

	var projectFilter302 = []string{"projectName==p302", "clusterId==uuid-cluster-1"}
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(projectFilter302), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload02},
	}, nil).AnyTimes()
	projectFilter302 = append(projectFilter302, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(projectFilter302), false).Return(&openapi.GetWorkloads200Response{
		Workloads: []openapi.Workload{workload02},
	}, nil).AnyTimes()

	var projectFilter300 = []string{"projectName==p300", "clusterId==uuid-cluster-1"}
	projectFilter300 = append(projectFilter300, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(projectFilter300), false).Return(&openapi.GetWorkloads200Response{
		Workloads: []openapi.Workload{},
	}, nil).AnyTimes()

	var trainingFilterCatch = []string{"projectName==p301", "clusterId==uuid-cluster-1"}
	trainingFilterCatch = append(trainingFilterCatch, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(trainingFilterCatch), false).Return(&openapi.GetWorkloads200Response{
		Workloads: []openapi.Workload{workload02},
	}, nil).AnyTimes()

	var runaiFilterCatch = []string{"clusterId==uuid-cluster-1", "type==Training||Distributed"}
	runaiFilterCatch = append(runaiFilterCatch, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(runaiFilterCatch), false).Return(&openapi.GetWorkloads200Response{
		Workloads: []openapi.Workload{workload02},
	}, nil).AnyTimes()

	var runaiFrameworkFilterCatch = []string{"clusterId==uuid-cluster-1", "type==Training||Distributed", "projectName==p302"}
	runaiFrameworkFilterCatch = append(runaiFrameworkFilterCatch, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(runaiFrameworkFilterCatch), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload02},
	}, nil).AnyTimes()

	var runaiP301FilterCatch = []string{"projectName==p301", "clusterId==uuid-cluster-1", "type==Training||Distributed"}
	runaiP301FilterCatch = append(runaiP301FilterCatch, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(runaiP301FilterCatch), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{},
	}, nil).AnyTimes()

	var trainingP301FilterCatch = []string{"projectName==p301", "clusterId==uuid-cluster-1", "type==Training||Distributed"}
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(trainingP301FilterCatch), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{},
	}, nil).AnyTimes()

	var runaiP302FilterCatch = []string{"projectName==p302", "name==workload-02", "clusterId==uuid-cluster-1", "type==Training||Distributed"}
	runaiP302FilterCatch = append(runaiP302FilterCatch, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(runaiP302FilterCatch), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload02},
	}, nil).AnyTimes()

	var p301FilterCatch = []string{"projectName==p301", "name==workload-01", "clusterId==uuid-cluster-1", "type==Training||Distributed"}
	p301FilterCatch = append(p301FilterCatch, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(p301FilterCatch), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload02},
	}, nil).AnyTimes()

	var workload21FilterCatch = []string{"projectName==p302", "name==workload-021", "clusterId==uuid-cluster-1", "type==Training||Distributed"}
	workload21FilterCatch = append(workload21FilterCatch, filterList...)
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(workload21FilterCatch), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload021},
	}, nil).AnyTimes()

	var trainingFilterNoProjectCatch = []string{"clusterId==uuid-cluster-1", "type==Training||Distributed"}
	mockWsClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.InAnyOrder(trainingFilterNoProjectCatch), false).Return(&openapi.GetWorkloads200Response{
		Next:      nil,
		Workloads: []openapi.Workload{workload02, workload03},
	}, nil).AnyTimes()
	workload.SetPaginator(mockWsClient)
}

type MockClientCalls interface {
	GetType() string
	GetFramework() api.Framework
	GenerateClient(w *Workload)
}

type Workload struct {
	mockCtrl         *gomock.Controller
	submitWlNameList []string
	mockSubmit       MockClientCalls
}

func NewWorkloadSubmitMocker(mockCtrl *gomock.Controller, mockSubmit MockClientCalls, wlList []string) *Workload {
	wl := &Workload{mockCtrl: mockCtrl, mockSubmit: mockSubmit, submitWlNameList: wlList}
	wl.init()

	return wl
}

func (w *Workload) init() {
	var filters []string

	if w.mockSubmit.GetType() != api.Training.StringCapitalize() {
		filters = []string{fmt.Sprintf("%s==%s", apifilter.Type, w.mockSubmit.GetType())}
	}

	framework := w.mockSubmit.GetFramework()
	for _, frameworkFilter := range apifilter.GetFrameworkFilters(framework) {
		filters = append(filters, frameworkFilter.Get())
	}

	MockWorkload(w.mockCtrl, filters)
	w.mockSubmit.GenerateClient(w)
}

func clientCallFunc(ctx context.Context, name, _, _ string, spec interface{}) {
	specValue, ok := spec.(*spec2.GeneralSpec)
	if !ok {
		err := fmt.Errorf("failed to cast to GeneralSpec")
		log.Printf("clientCallFunc, error: %v\n", err.Error())
		panic(err)
	}

	switch specValue.GetType() {
	case spectype.Training:
		submitSpec, err := specValue.GenerateTrainingSpec(ctx, clusterId, mockNodesService)
		if err != nil {
			log.Printf("clientCallFunc, training generation failed, error: %v\n", err.Error())
			panic(err)
		}
		specMap[name] = *submitSpec
	case spectype.Workspace:
		submitSpec, err := specValue.GenerateWorkspaceSpec(ctx, clusterId, mockNodesService)
		if err != nil {
			log.Printf("clientCallFunc, workspace generation failed, error: %v\n", err.Error())
			panic(err)
		}
		specMap[name] = *submitSpec
	case spectype.Inference:
		submitSpec, err := specValue.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
		if err != nil {
			log.Printf("clientCallFunc, inference generation failed, error: %v\n", err.Error())
			panic(err)
		}
		specMap[name] = *submitSpec
	case spectype.InferenceDistributed:
		distInfSpec, err := specValue.GenerateDistributedInferenceSpec(ctx, clusterId, mockNodesService)
		if err != nil {
			log.Printf("clientCallFunc, distributed inference generation failed, error: %v\n", err.Error())
			panic(err)
		}
		request := openapi.NewDistributedInferenceSpec()
		request.SetSpec(*distInfSpec)
		specMap[name] = *request
	case spectype.TrainingDistributed:
		submitSpec, err := specValue.GenerateDistributedSpec(ctx, clusterId, mockNodesService)
		if err != nil {
			log.Printf("clientCallFunc, workspace generation failed, error: %v\n", err.Error())
			panic(err)
		}
		specMap[name] = *submitSpec
	default:
		err := fmt.Errorf("unknown workload type: %s", specValue.GetType().DisplayName())
		log.Printf("clientCallFunc, couldn't genreate spec, error: %v\n", err.Error())
		panic(err)
	}
}

type Workspace struct {
}

func (w *Workspace) GetType() string {
	return api.Workspace.QueryParam()
}

func (w *Workspace) GetFramework() api.Framework {
	return api.Runai
}

func (w *Workspace) GenerateClient(wl *Workload) {
	mockWsClient := generateMockClient(wl)
	workspace.SetClient(mockWsClient)
}

type Inference struct {
}

func (infw *Inference) GetType() string {
	return api.Inference.QueryParam()
}

func (infw *Inference) GetFramework() api.Framework {
	return api.Knative
}

func (infw *Inference) GenerateClient(wl *Workload) {
	mockInfwClient := generateUpdateMockClient(wl)
	inference.SetClient(mockInfwClient)
}

func generateUpdateMockClient(wl *Workload) *cpworkloadmocks.MockWithUpdateAction {
	mockNodesService = getMockNodesService(wl.mockCtrl)
	mockWsClient := cpworkloadmocks.NewMockWithUpdateAction(wl.mockCtrl)
	for _, wsName := range wl.submitWlNameList {
		mockWsClient.EXPECT().
			Submit(gomock.Any(), wsName, gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			DoAndReturn(func(ctx context.Context, name, _, _ string, specArg interface{}) (string, error) {
				// Call the original clientCallFunc to populate specMap
				clientCallFunc(ctx, name, "", "", specArg)

				if generalSpec, ok := specArg.(*spec2.GeneralSpec); ok {
					// For distributed inference, validate and generate spec
					if generalSpec.GetFramework() == api.LWS {
						distInfSpec, err := generalSpec.GenerateDistributedInferenceSpec(ctx, clusterId, mockNodesService)
						if err != nil {
							// Return error for validation failures
							return "", err
						}
						// For dry-run, return JSON
						if generalSpec.IsDryRun() {
							distributedInference := openapi.NewDistributedInferenceCreationRequest(name, "p301", clusterId)
							distributedInference.SetSpec(*distInfSpec)
							jsonBytes, err := json.Marshal(distributedInference)
							if err != nil {
								return fmt.Sprintf("{\"error\":\"failed to marshal spec in mock: %v\"}", err), nil
							}
							return string(jsonBytes), nil
						}
						// For non-dry-run, return name
						return name, nil
					}

					// For regular inference, check dry-run
					if generalSpec.GetType() == spectype.Inference && generalSpec.IsDryRun() {
						infwSpec, err := generalSpec.GenerateInferenceSpec(ctx, clusterId, mockNodesService)
						if err != nil {
							return fmt.Sprintf("{\"error\":\"failed to generate spec in mock: %v\"}", err), nil
						}
						inference := openapi.NewInferenceCreationRequest(name, "p301", clusterId)
						inference.SetSpec(*infwSpec)
						jsonBytes, err := json.Marshal(inference)
						if err != nil {
							return fmt.Sprintf("{\"error\":\"failed to marshal spec in mock: %v\"}", err), nil
						}
						return string(jsonBytes), nil
					}
				}

				// For non-dry-run, return workload name
				return name, nil
			})
	}

	mockWsClient.EXPECT().
		Delete(gomock.Any(), UuidExist).
		AnyTimes().
		Return(nil)

	mockWsClient.EXPECT().
		Delete(gomock.Any(), UuidDontExist).
		AnyTimes().
		Return(fmt.Errorf("generic error string"))

	mockWsClient.EXPECT().
		Update(gomock.Any(), UuidExist, gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(ctx context.Context, _, _ string, specArg interface{}) (string, error) {
			if generalSpec, ok := specArg.(*spec2.GeneralSpec); ok {
				// For distributed inference, generate update request
				if generalSpec.GetFramework() == api.LWS {
					updateReq := openapi.NewUpdateRequest()
					updateSpec := openapi.NewUpdateSpecSpec()
					if generalSpec.GetFlags().DistributedInference.Replicas >= 0 {
						updateSpec.SetReplicas(generalSpec.GetFlags().DistributedInference.Replicas)
					}
					updateReq.SetSpec(*updateSpec)

					// Store in updateRequestMap for testing
					updateRequestMap[WorkloadName01] = updateReq

					if generalSpec.IsDryRun() {
						jsonBytes, err := json.Marshal(updateReq)
						if err != nil {
							return fmt.Sprintf("{\"error\":\"failed to marshal spec in mock: %v\"}", err), nil
						}
						return string(jsonBytes), nil
					}
					return WorkloadName01, nil
				}

				// For standard inference
				if generalSpec.IsDryRun() {
					inferenceUpdatePayload, err := generalSpec.GenerateInferenceUpdateSpec(ctx, clusterId, mockNodesService)
					if err != nil {
						return fmt.Sprintf("{\"error\":\"failed to generate spec in mock: %v\"}", err), nil
					}
					jsonBytes, err := json.Marshal(inferenceUpdatePayload)
					if err != nil {
						return fmt.Sprintf("{\"error\":\"failed to marshal spec in mock: %v\"}", err), nil
					}
					return string(jsonBytes), nil
				}
			}
			return WorkloadName01, nil // For non-dry-run
		})

	mockWsClient.EXPECT().
		Update(gomock.Any(), UuidDontExist, gomock.Any(), gomock.Any()).
		AnyTimes().
		Return("", fmt.Errorf("generic error string"))

	return mockWsClient
}

type Training struct {
}

func (t *Training) GetType() string {
	return "Training"
}

func (t *Training) GetFramework() api.Framework {
	return api.Runai
}

func (t *Training) GenerateClient(wl *Workload) {
	mockWsClient := generateMockClient(wl)
	training.SetClient(mockWsClient)
}

type Distributed struct {
	Framework api.Framework
}

func (d *Distributed) GetType() string {
	return "Distributed"
}

func (d *Distributed) GetFramework() api.Framework {
	return d.Framework
}

func (d *Distributed) GenerateClient(wl *Workload) {
	mockWsClient := generateMockClient(wl)
	distributed.SetClient(mockWsClient)
}

type DistributedInference struct {
}

func (di *DistributedInference) GetType() string {
	return api.DistributedInference.QueryParam()
}

func (di *DistributedInference) GetFramework() api.Framework {
	return api.LWS
}

func (di *DistributedInference) GenerateClient(wl *Workload) {
	// Distributed inference uses update client to support scale operations
	mockDistInfClient := generateUpdateMockClient(wl)
	distributedinference.SetClient(mockDistInfClient)
}

func generateMockClient(wl *Workload) *cpworkloadmocks.MockWithSuspendableAction {
	mockNodesService = getMockNodesService(wl.mockCtrl)
	mockWsClient := cpworkloadmocks.NewMockWithSuspendableAction(wl.mockCtrl)
	for _, wsName := range wl.submitWlNameList {
		mockWsClient.EXPECT().
			Submit(gomock.Any(), wsName, gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(wsName, nil).
			Do(clientCallFunc)
	}

	mockWsClient.EXPECT().
		Delete(gomock.Any(), UuidExist).
		AnyTimes().
		Return(nil)

	mockWsClient.EXPECT().
		Delete(gomock.Any(), UuidDontExist).
		AnyTimes().
		Return(fmt.Errorf("generic error string"))

	mockWsClient.EXPECT().
		Resume(gomock.Any(), UuidExist).
		AnyTimes().
		Return(nil)

	mockWsClient.EXPECT().
		Resume(gomock.Any(), UuidDontExist).
		AnyTimes().
		Return(fmt.Errorf("generic error string"))

	mockWsClient.EXPECT().
		Suspend(gomock.Any(), UuidExist).
		AnyTimes().
		Return(nil)

	mockWsClient.EXPECT().
		Suspend(gomock.Any(), UuidDontExist).
		AnyTimes().
		Return(fmt.Errorf("generic error string"))
	return mockWsClient
}

func getMockNodesService(mockCtrl *gomock.Controller) *nodes.BackendNodeService {
	mockNodeClient := cpmocks.NewMockNodeClient(mockCtrl)
	mockNodes := getMockNodeList()
	mockNodeClient.EXPECT().GetNodeList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockNodes, nil).AnyTimes()
	return nodes.NewBackendNodeService(mockNodeClient)
}
