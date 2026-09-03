/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/google/uuid"
	. "github.com/onsi/gomega" //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
	"k8s.io/utils/ptr"
)

const (
	DefaultType                  = "Training"
	WorkspaceType                = "Workspace"
	InferenceType                = "Inference"
	DefaultWorkloadName          = "job-1"
	DefaultPodName               = "job-1-1"
	DefaultContainerName         = "default"
	DefaultPriorityClassName     = "high-priority"
	DefaultNodeName              = "worker-0"
	DefaultNodepool              = "default"
	OtherNodepool                = "non-default"
	DefaultSubmittedBy           = "user@run.ai"
	DefaultProjectName           = "test"
	DefaultProjectId             = "1"
	DefaultDepartmentName        = "dep-1"
	DefaultDepartmentId          = "1"
	AnotherProjectId             = "2"
	DefaultNamespace             = "runai-test"
	DefaultGPURequestType        = api.GPUREQUESTTYPE_PORTION
	DefaultGPULimit              = 2
	DefaultGPURequest            = 1
	DefaultPhase                 = api.PHASE_RUNNING
	DefaultK8sPhase              = "Running"
	DefaultSource                = api.SOURCE_CLI
	DefaultNumberOfRequestedPods = 5
	DefaultTolerationKey         = "Key"
	DefaultTolerationOperator    = "Exists"
	DefaultTolerationEffect      = "Effect"
	DefaultTolerationSeconds     = 20
	DefaultCommand               = "sleep"
	DefaultArguments             = "10000"
	DefaultIP                    = "10.244.5.1"
	PytorchFramework             = "PyTorchJob"
	DefaultRevisionName          = "workload-00001"
	DefaultRevisionType          = "KnativeRevision"
)

var (
	DefaultCreationTime         = time.Now()
	DefaultGuaranteedEndsAtTime = DefaultCreationTime.Add(2 * time.Hour)
	DefaultRequestedNodePools   = []string{"default", "non-default"}
	DefaultImages               = []string{"ubuntu"} // must have at least 1 item
	DefaultUrls                 = []string{}
	DefaultChildrenIds          = []api.WorkloadChildrenIdsInner{
		{
			Id:   uuid.NewString(),
			Type: DefaultType,
		},
		{
			Id:   uuid.NewString(),
			Type: DefaultType,
		},
		{
			Id:   uuid.NewString(),
			Type: DefaultType,
		},
	}
	DefaultEnvVars = map[string]string{"ENV_VAR_1": "ENV_VALUE_1", "ENV_VAR_2": "ENV_VALUE_2"}
)

func SetupBaseWorkload(id string, projectId string, clusterId string) api.WorkloadSyncFields {
	priority := int32(50)
	return api.WorkloadSyncFields{
		Type:                 DefaultType,
		Name:                 DefaultWorkloadName,
		ExternalWorkloadName: fmt.Sprintf("ew-training-%s", id), // Just for uniqueness
		Source:               DefaultSource,
		Id:                   id,
		SubmittedBy:          DefaultSubmittedBy,
		ClusterId:            clusterId,
		ProjectName:          DefaultProjectName,
		ProjectId:            projectId,
		DepartmentId:         DefaultDepartmentId,
		DepartmentName:       DefaultDepartmentName,
		Namespace:            DefaultNamespace,
		CreatedAt:            DefaultCreationTime,
		RequestedResources: &api.WorkloadRequestResources{
			GpuRequestType: (*api.GpuRequestType)(ptr.To(string(DefaultGPURequestType))),
			Gpu: &api.RequestResourceCores{
				Limit:   ptr.To(float64(DefaultGPULimit)),
				Request: ptr.To(float64(DefaultGPURequest)),
			},
		},
		Phase:                   DefaultPhase,
		PhaseMessage:            "",
		K8sPhase:                DefaultK8sPhase,
		RequestedPods:           api.RequestedPods{Number: ptr.To(int32(DefaultNumberOfRequestedPods))},
		RequestedNodePools:      DefaultRequestedNodePools,
		Images:                  DefaultImages,
		ChildrenIds:             DefaultChildrenIds,
		Urls:                    DefaultUrls,
		AdditionalFields:        "",
		Priority:                &priority,
		ActionsSupport:          api.ActionsSupport{Delete: true, Suspend: true},
		Command:                 DefaultCommand,
		Arguments:               DefaultArguments,
		EnvironmentVariables:    DefaultEnvVars,
		PriorityClassName:       DefaultPriorityClassName,
		Category:                "Train",
		GuaranteedRuntimeEndsAt: &DefaultGuaranteedEndsAtTime,
	}
}

func ValidateBaseWorkload(workload api.WorkloadDetailed, id string, projectId string, clusterId string) {
	Expect(workload.Type).To(Equal(DefaultType))
	Expect(workload.Name).To(Equal(DefaultWorkloadName))
	Expect(workload.ExternalWorkloadName).To(Equal(fmt.Sprintf("ew-training-%s", id)))
	Expect(workload.Source).To(Equal(DefaultSource))
	Expect(workload.Id).To(Equal(id))
	Expect(workload.SubmittedBy).To(Equal(DefaultSubmittedBy))
	Expect(workload.ClusterId).To(Equal(clusterId))
	Expect(workload.ProjectName).To(Equal(DefaultProjectName))
	Expect(workload.ProjectId).To(Equal(projectId))
	Expect(workload.DepartmentName).To(Equal(DefaultDepartmentName))
	Expect(workload.DepartmentId).To(Equal(DefaultDepartmentId))
	Expect(workload.Namespace).To(Equal(DefaultNamespace))
	Expect(workload.CreatedAt.IsZero()).To(BeFalse())
	Expect(workload.CreatedAt.UnixMilli()).Should(BeNumerically(">=", workload.CreatedAt.Unix()*1000))
	Expect(workload.CreatedAt.UnixMilli()).Should(BeNumerically("<=", (workload.CreatedAt.Unix()+1)*1000))
	Expect(workload.WorkloadRequestedResources.GpuRequestType).ToNot(BeNil())
	Expect(*workload.WorkloadRequestedResources.GpuRequestType).To(Equal(DefaultGPURequestType))
	Expect(workload.WorkloadRequestedResources.Gpu.Limit).ToNot(BeNil())
	Expect(*workload.WorkloadRequestedResources.Gpu.Limit).To(Equal(float64(DefaultGPULimit)))
	Expect(workload.WorkloadRequestedResources.Gpu.Request).ToNot(BeNil())
	Expect(*workload.WorkloadRequestedResources.Gpu.Request).To(Equal(float64(DefaultGPURequest)))
	Expect(workload.WorkloadRequestedResources.GpuMemory).To(BeNil())
	Expect(workload.WorkloadRequestedResources.Cpu).To(BeNil())
	Expect(workload.WorkloadRequestedResources.CpuMemory).To(BeNil())
	Expect(workload.Phase).To(Equal(DefaultPhase))
	Expect(workload.PhaseMessage).To(Equal(""))
	Expect(workload.K8sPhase).To(Equal(DefaultK8sPhase))
	Expect(workload.RequestedPods.Number).To(Not(BeNil()))
	Expect(*workload.RequestedPods.Number).To(Equal(int32(DefaultNumberOfRequestedPods)))
	Expect(workload.RequestedNodePools).To(Equal(DefaultRequestedNodePools))
	Expect(workload.Images).To(Equal(DefaultImages))
	Expect(workload.Urls).To(BeNil())
	Expect(*workload.Preemptible).To(BeTrue())
	Expect(workload.Command).To(Equal(DefaultCommand))
	Expect(workload.Arguments).To(Equal(DefaultArguments))
	Expect(workload.EnvironmentVariables).To(BeEquivalentTo(DefaultEnvVars))
	Expect(workload.Priority).To(Not(BeNil()))
	Expect(*workload.Priority).To(Equal(int32(50)))
	Expect(workload.PriorityClassName).To(Equal(DefaultPriorityClassName))
	Expect(workload.PhaseReason).To(BeNil())
	Expect(workload.PendingSchedulingMessages).To(BeNil())

	var expectedAdditionalFields models.AdditionalFields = nil
	Expect(workload.AdditionalFields).To(BeEquivalentTo(expectedAdditionalFields))
}

func SetupBasePod(id string, workloadId string, projectId string, clusterId string) api.PodSyncFields {
	return api.PodSyncFields{
		Name:              DefaultPodName,
		PriorityClassName: DefaultPriorityClassName,
		Id:                id,
		WorkloadId:        workloadId,
		ClusterId:         clusterId,
		ProjectId:         projectId,
		NodeName:          ptr.To(DefaultNodeName),
		CreatedAt:         DefaultCreationTime,
		Containers: []api.Container{
			{
				Name:      DefaultContainerName,
				Image:     DefaultImages[0],
				StartedAt: &DefaultCreationTime,
			},
		},
		Tolerations: []api.PodToleration{
			{
				Key:               DefaultTolerationKey,
				Operator:          DefaultTolerationOperator,
				Effect:            DefaultTolerationEffect,
				TolerationSeconds: ptr.To(int32(DefaultTolerationSeconds)),
			},
		},
		CurrentNodePool:    ptr.To(DefaultNodepool),
		RequestedNodePools: DefaultRequestedNodePools,
		RequestedResources: &api.PodRequestResources{
			GpuRequestType: (*api.GpuRequestType)(ptr.To(string(DefaultGPURequestType))),
			Gpu: &api.RequestResourceCores{
				Limit:   ptr.To(float64(DefaultGPULimit)),
				Request: ptr.To(float64(DefaultGPURequest)),
			},
		},
		AllocatedResources: &api.AllocatedResources{
			Gpu: ptr.To(float64(DefaultGPULimit)),
		},
		K8sPhase: DefaultK8sPhase,
		Ip:       ptr.To(DefaultIP),
		Task:     ptr.To("Leader"),
		NodeAffinity: api.PodNodeAffinity{
			Required: &api.PodNodeAffinityRequired{NodeSelectorTerms: []api.NodeSelectorTerm{
				{
					MatchExpressions: []api.MatchExpression{
						{
							Key:      "k1",
							Operator: api.MATCHEXPRESSIONOPERATOR_NOT_IN,
							Values:   []string{"v1", "v2"},
						},
					},
				},
			}},
			Preferred: &api.PodNodeAffinityPreferred{NodeSelectorTerms: []api.PodPreferredNodeSelectorTerm{
				{
					Weight: 10,
					Preference: &api.NodeSelectorTerm{
						MatchExpressions: []api.MatchExpression{
							{
								Key:      "k2",
								Operator: api.MATCHEXPRESSIONOPERATOR_DOES_NOT_EXIST,
							},
						},
					},
				},
			}},
		},
	}
}

func ValidateBasePod(pod api.Pod, id string, workloadId string, projectId string, clusterId string, revisionId *string) {
	Expect(pod.Name).To(Equal(DefaultPodName))
	Expect(pod.PriorityClassName).To(Equal(DefaultPriorityClassName))
	Expect(pod.Id).To(Equal(id))
	Expect(pod.WorkloadId).To(Equal(workloadId))
	Expect(pod.ClusterId).To(Equal(clusterId))
	Expect(pod.ProjectId).To(Equal(projectId))
	Expect(pod.NodeName).To(Not(BeNil()))
	Expect(*pod.NodeName).To(Equal(DefaultNodeName))
	Expect(pod.CreatedAt.Unix()).To(Equal(DefaultCreationTime.Unix()))
	Expect(pod.Containers).ToNot(BeNil())
	Expect(len(pod.Containers)).To(Equal(1))
	Expect(pod.Containers[0].Name).To(Equal(DefaultContainerName))
	Expect(pod.Containers[0].Image).To(Equal(DefaultImages[0]))
	Expect(pod.Containers[0].StartedAt.Unix()).To(Equal(DefaultCreationTime.Unix()))
	Expect(len(pod.Tolerations)).To(Equal(1))
	Expect(pod.Tolerations[0].Key).To(Equal(DefaultTolerationKey))
	Expect(pod.Tolerations[0].Operator).To(Equal(DefaultTolerationOperator))
	Expect(pod.Tolerations[0].Value).To(Equal(""))
	Expect(pod.Tolerations[0].Effect).To(Equal(DefaultTolerationEffect))
	Expect(pod.Tolerations[0].TolerationSeconds).ToNot(BeNil())
	Expect(*pod.Tolerations[0].TolerationSeconds).To(Equal(int32(DefaultTolerationSeconds)))
	Expect(pod.CurrentNodePool).To(Not(BeNil()))
	Expect(*pod.CurrentNodePool).To(Equal(DefaultNodepool))
	Expect(pod.RequestedNodePools).To(Equal(DefaultRequestedNodePools))
	Expect(pod.RequestedResources.GpuRequestType).ToNot(BeNil())
	Expect(*pod.RequestedResources.GpuRequestType).To(Equal(DefaultGPURequestType))
	Expect(pod.RequestedResources.Gpu.Limit).ToNot(BeNil())
	Expect(*pod.RequestedResources.Gpu.Limit).To(Equal(float64(DefaultGPULimit)))
	Expect(pod.RequestedResources.Gpu.Request).ToNot(BeNil())
	Expect(*pod.RequestedResources.Gpu.Request).To(Equal(float64(DefaultGPURequest)))
	Expect(pod.RequestedResources.GpuMemory).To(BeNil())
	Expect(pod.RequestedResources.Cpu).To(BeNil())
	Expect(pod.RequestedResources.CpuMemory).To(BeNil())
	Expect(pod.AllocatedResources.Gpu).ToNot(BeNil())
	Expect(*pod.AllocatedResources.Gpu).To(Equal(float64(DefaultGPULimit)))
	Expect(pod.AllocatedResources.GpuMemory).To(BeNil())
	Expect(pod.AllocatedResources.Cpu).To(BeNil())
	Expect(pod.AllocatedResources.CpuMemory).To(BeNil())
	Expect(pod.K8sPhase).To(Equal(DefaultK8sPhase))
	Expect(pod.Ip).ToNot(BeNil())
	Expect(*pod.Ip).To(Equal(DefaultIP))
	Expect(pod.Task).ToNot(BeNil())
	Expect(*pod.Task).To(Equal("Leader"))

	if revisionId != nil {
		Expect(pod.RevisionId).ToNot(BeNil())
		Expect(*pod.RevisionId).To(Equal(*revisionId))
	} else {
		Expect(pod.RevisionId).To(BeNil())
	}

	Expect(pod.NodeAffinity).NotTo(BeNil())
	Expect(pod.NodeAffinity.Required).NotTo(BeNil())
	Expect(pod.NodeAffinity.Required.NodeSelectorTerms).To(HaveLen(1))
	Expect(pod.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
	Expect(pod.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Operator).To(Equal(api.MATCHEXPRESSIONOPERATOR_NOT_IN))
	Expect(pod.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values).To(HaveLen(2))

	Expect(pod.NodeAffinity.Preferred).NotTo(BeNil())
	Expect(pod.NodeAffinity.Preferred.NodeSelectorTerms).To(HaveLen(1))
	Expect(pod.NodeAffinity.Preferred.NodeSelectorTerms[0].Weight).To(Equal(float32(10)))
	Expect(pod.NodeAffinity.Preferred.NodeSelectorTerms[0].Preference.MatchExpressions).To(HaveLen(1))
	Expect(pod.NodeAffinity.Preferred.NodeSelectorTerms[0].Preference.MatchExpressions[0].Operator).To(Equal(api.MATCHEXPRESSIONOPERATOR_DOES_NOT_EXIST))
}

func SetupBaseRevision(id string, workloadId string, projectId string, clusterId string) api.RevisionSyncFields {
	return api.RevisionSyncFields{
		Name:               DefaultRevisionName,
		Id:                 id,
		Type:               DefaultRevisionType,
		WorkloadId:         workloadId,
		ClusterId:          clusterId,
		DepartmentId:       DefaultDepartmentId,
		ProjectId:          projectId,
		CreatedAt:          DefaultCreationTime,
		RequestedNodePools: DefaultRequestedNodePools,
		RequestedResources: &api.WorkloadRequestResources{
			GpuRequestType: (*api.GpuRequestType)(ptr.To(string(DefaultGPURequestType))),
			Gpu: &api.RequestResourceCores{
				Limit:   ptr.To(float64(DefaultGPULimit)),
				Request: ptr.To(float64(DefaultGPURequest)),
			},
		},
		Phase:                DefaultPhase,
		Images:               DefaultImages,
		Command:              DefaultCommand,
		Arguments:            DefaultArguments,
		EnvironmentVariables: DefaultEnvVars,
		ChildrenIds:          DefaultChildrenIds,
	}
}

func ValidateBaseRevision(ctx context.Context, revisionsService api.RevisionsAPIServicer, revisionId string, workloadId string, projectId string, clusterId string, deleted bool) {
	apiResponse, _ := revisionsService.GetRevision(ctx, revisionId)
	Expect(apiResponse.Code).To(Equal(http.StatusOK))
	Expect(apiResponse.Body).ToNot(BeNil())
	revision := apiResponse.Body.(api.Revision)

	Expect(revision.Name).To(Equal(DefaultRevisionName))
	Expect(revision.Id).To(Equal(revisionId))
	Expect(revision.WorkloadId).To(Equal(workloadId))
	Expect(revision.ClusterId).To(Equal(clusterId))
	Expect(revision.DepartmentId).To(Equal(DefaultDepartmentId))
	Expect(revision.ProjectId).To(Equal(projectId))
	Expect(revision.CreatedAt.Unix()).To(Equal(DefaultCreationTime.Unix()))
	Expect(revision.RevisionRequestedResources.GpuRequestType).ToNot(BeNil())
	Expect(*revision.RevisionRequestedResources.GpuRequestType).To(Equal(DefaultGPURequestType))
	Expect(revision.RevisionRequestedResources.Gpu.Limit).ToNot(BeNil())
	Expect(*revision.RevisionRequestedResources.Gpu.Limit).To(Equal(float64(DefaultGPULimit)))
	Expect(revision.RevisionRequestedResources.Gpu.Request).ToNot(BeNil())
	Expect(*revision.RevisionRequestedResources.Gpu.Request).To(Equal(float64(DefaultGPURequest)))
	Expect(revision.RevisionRequestedResources.GpuMemory).To(BeNil())
	Expect(revision.RevisionRequestedResources.Cpu).To(BeNil())
	Expect(revision.RevisionRequestedResources.CpuMemory).To(BeNil())
	Expect(revision.Phase).To(Equal(DefaultPhase))
	Expect(revision.PhaseMessage).To(Equal(""))
	Expect(revision.Images).To(Equal(DefaultImages))
	Expect(revision.Command).To(Equal(DefaultCommand))
	Expect(revision.Arguments).To(Equal(DefaultArguments))
	Expect(revision.EnvironmentVariables).To(BeEquivalentTo(DefaultEnvVars))

	if deleted {
		Expect(revision.DeletedAt).ToNot(BeNil())
	} else {
		Expect(revision.DeletedAt).To(BeNil())
	}
}

func ValidateBaseRevisionChildren(ctx context.Context, revisionId string, childrenIds []api.WorkloadChildrenIdsInner) {
	// validate the children, expecting an additional child for the revision itself
	children, err := dao.DbGetRevisionChildren(ctx, revisionId, nil)
	Expect(err).To(BeNil())
	Expect(children).ToNot(BeNil())

	var sentChildrenIds []string
	for _, child := range childrenIds {
		sentChildrenIds = append(sentChildrenIds, child.Id)
	}
	sentChildrenIds = append(sentChildrenIds, revisionId)

	var receivedChildrenIds []string
	for _, child := range *children {
		receivedChildrenIds = append(receivedChildrenIds, child.ChildID.String())
	}

	Expect(len(receivedChildrenIds)).To(Equal(len(sentChildrenIds)))
	Expect(receivedChildrenIds).To(ContainElements(sentChildrenIds))
}

func ValidateGetRevisions(ctx context.Context, revisionsService api.RevisionsAPIServicer, workloadId string, deleted bool, expectedIds []string) {
	// validate get
	getResponse, _ := revisionsService.GetInferenceWorkloadRevisions(ctx, workloadId, deleted)
	Expect(getResponse.Code).To(Equal(http.StatusOK))

	revisionsResponse := getResponse.Body.(api.GetInferenceWorkloadRevisions200Response)
	revisions := revisionsResponse.Revisions
	Expect(len(revisions)).To(Equal(len(expectedIds)))

	// the order of the revisions is not guaranteed so we collect them and use ContainElements to compare
	actualIds := make([]string, len(revisions))
	for i, revision := range revisions {
		actualIds[i] = revision.Id
	}
	Expect(actualIds).To(ContainElements(expectedIds))

	// validate count
	getResponse, _ = revisionsService.CountInferenceWorkloadRevisions(ctx, workloadId, deleted)
	Expect(getResponse.Code).To(Equal(http.StatusOK))
	Expect(getResponse.Body.(api.CountWorkloads200Response).Count).To(Equal(int64(len(expectedIds))))
}
