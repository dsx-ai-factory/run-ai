/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

type TemplatesService struct {
	tenantSettingsClient clients.TenantSettingsClient
}

func NewTemplatesService(tenantSettingsClient clients.TenantSettingsClient) api.WorkloadTemplatesAPIServicer {
	return &TemplatesService{
		tenantSettingsClient: tenantSettingsClient,
	}
}

func (ts *TemplatesService) CountTemplates(ctx context.Context, filterByParam []string, searchParam string, includeAncestors bool) (api.ImplResponse, error) {
	result, err := ts.CountTemplatesUnified(ctx, filterByParam, searchParam, includeAncestors)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, api.CountResponse{
		Count: result,
	}), nil
}

func (ts *TemplatesService) CreateDistributedTemplate(ctx context.Context, request api.DistributedTemplateCreationRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	err = ts.CreateTemplate(ctx, unifiedRequest)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(unifiedRequest.template)
}

func (ts *TemplatesService) CreateInferenceTemplate(ctx context.Context, request api.InferenceTemplateCreationRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_INFERENCE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	err = ts.CreateTemplate(ctx, unifiedRequest)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(unifiedRequest.template)
}

func (ts *TemplatesService) CreateDistributedInferenceTemplate(ctx context.Context, request api.DistributedInferenceTemplateCreationRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED_INFERENCE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	err = ts.CreateTemplate(ctx, unifiedRequest)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(unifiedRequest.template)
}

func (ts *TemplatesService) CreateTrainingTemplate(ctx context.Context, request api.TrainingTemplateCreationRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_TRAINING, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	err = ts.CreateTemplate(ctx, unifiedRequest)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(unifiedRequest.template)
}

func (ts *TemplatesService) CreateWorkspaceTemplate(ctx context.Context, request api.WorkspaceTemplateCreationRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	err = ts.CreateTemplate(ctx, unifiedRequest)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(unifiedRequest.template)
}

func (ts *TemplatesService) DeleteDistributedTemplate(ctx context.Context, id string) (api.ImplResponse, error) {
	err := ts.DeleteTemplate(ctx, api.WORKLOADTYPE_DISTRIBUTED, id)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (ts *TemplatesService) DeleteInferenceTemplate(ctx context.Context, id string) (api.ImplResponse, error) {
	err := ts.DeleteTemplate(ctx, api.WORKLOADTYPE_INFERENCE, id)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (ts *TemplatesService) DeleteDistributedInferenceTemplate(ctx context.Context, id string) (api.ImplResponse, error) {
	err := ts.DeleteTemplate(ctx, api.WORKLOADTYPE_DISTRIBUTED_INFERENCE, id)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (ts *TemplatesService) DeleteTrainingTemplate(ctx context.Context, id string) (api.ImplResponse, error) {
	err := ts.DeleteTemplate(ctx, api.WORKLOADTYPE_TRAINING, id)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (ts *TemplatesService) DeleteWorkspaceTemplate(ctx context.Context, id string) (api.ImplResponse, error) {
	err := ts.DeleteTemplate(ctx, api.WORKLOADTYPE_WORKSPACE, id)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (ts *TemplatesService) GetTemplateById(ctx context.Context, templateId string) (api.ImplResponse, error) {
	template, err := ts.GetTemplate(ctx, nil, templateId)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(template.template)
}

func (ts *TemplatesService) GetDistributedTemplate(ctx context.Context, templateId string) (api.ImplResponse, error) {
	template, err := ts.GetTemplate(ctx, convutils.Ref(api.WORKLOADTYPE_DISTRIBUTED), templateId)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(template.template)
}

func (ts *TemplatesService) GetInferenceTemplate(ctx context.Context, templateId string) (api.ImplResponse, error) {
	template, err := ts.GetTemplate(ctx, convutils.Ref(api.WORKLOADTYPE_INFERENCE), templateId)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(template.template)
}

func (ts *TemplatesService) GetDistributedInferenceTemplate(ctx context.Context, templateId string) (api.ImplResponse, error) {
	template, err := ts.GetTemplate(ctx, convutils.Ref(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE), templateId)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(template.template)
}

func (ts *TemplatesService) GetTrainingTemplate(ctx context.Context, templateId string) (api.ImplResponse, error) {
	template, err := ts.GetTemplate(ctx, convutils.Ref(api.WORKLOADTYPE_TRAINING), templateId)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(template.template)
}

func (ts *TemplatesService) GetWorkspaceTemplate(ctx context.Context, templateId string) (api.ImplResponse, error) {
	template, err := ts.GetTemplate(ctx, convutils.Ref(api.WORKLOADTYPE_WORKSPACE), templateId)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(template.template)
}

// TemplateNameAvailability - Template name availability
func (ts *TemplatesService) TemplateNameAvailability(ctx context.Context, templateName string, scopeType string, scopeId string) (api.ImplResponse, error) {
	err := checkTemplateNameAvailability(ctx, templateName, scopeType, scopeId)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

type ComplianceParams struct {
	projectId    int32
	workloadType api.WorkloadType
	replica      string
}

func NewComplianceParams(projectId int32, workloadType api.WorkloadType, replica string) (*ComplianceParams, *api.Error) {
	isProjectIdGiven := projectId > 0
	if !isProjectIdGiven {
		return nil, nil
	}

	switch workloadType {
	case "":
		return nil, api.NewValidationError("complyToWorkloadType should be provided provided along with complyToProjectId")
	case api.WORKLOADTYPE_DISTRIBUTED:
		if replica == "" {
			replica = string(api.REPLICATYPE_WORKER)
		}
	}

	return &ComplianceParams{
		projectId:    projectId,
		workloadType: workloadType,
		replica:      replica,
	}, nil
}

func (ts *TemplatesService) ListTemplates(ctx context.Context, offsetParam int32, limitParam int32, sortOrderParam string, sortByParam string,
	filterByParam []string, searchParam string, includeAncestors bool,
	complyToProjectId int32, complyToWorkloadType string, complyToWorkloadReplica string,
) (api.ImplResponse, error) {
	complianceParams, paramsErr := NewComplianceParams(complyToProjectId, api.WorkloadType(complyToWorkloadType), complyToWorkloadReplica)
	if paramsErr != nil {
		return api.Response(http.StatusBadRequest, paramsErr), nil
	}

	result, err := ts.ListTemplatesUnified(ctx, nil, offsetParam, limitParam, sortOrderParam, sortByParam,
		filterByParam, searchParam, complianceParams, includeAncestors)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (ts *TemplatesService) ListDistributedTemplates(ctx context.Context, offsetParam int32, limitParam int32, sortOrderParam string, sortByParam string,
	filterByParam []string, searchParam string, includeAncestors bool,
	complyToProjectId int32, complyToReplicaType string,
) (api.ImplResponse, error) {
	complianceInfo, complianceErr := NewComplianceParams(complyToProjectId, api.WORKLOADTYPE_DISTRIBUTED, complyToReplicaType)
	if complianceErr != nil {
		return api.Response(http.StatusBadRequest, complianceErr), nil
	}

	result, err := ts.ListTemplatesUnified(ctx, convutils.Ref(api.WORKLOADTYPE_DISTRIBUTED), offsetParam, limitParam, sortOrderParam, sortByParam,
		filterByParam, searchParam, complianceInfo, includeAncestors)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (ts *TemplatesService) ListInferenceTemplates(ctx context.Context, offsetParam int32, limitParam int32, sortOrderParam string, sortByParam string,
	filterByParam []string, searchParam string, includeAncestors bool,
	complyToProjectId int32,
) (api.ImplResponse, error) {
	complianceInfo, complianceErr := NewComplianceParams(complyToProjectId, api.WORKLOADTYPE_INFERENCE, "")
	if complianceErr != nil {
		return api.Response(http.StatusBadRequest, complianceErr), nil
	}

	result, err := ts.ListTemplatesUnified(ctx, convutils.Ref(api.WORKLOADTYPE_INFERENCE), offsetParam, limitParam, sortOrderParam, sortByParam,
		filterByParam, searchParam, complianceInfo, includeAncestors)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (ts *TemplatesService) ListDistributedInferenceTemplates(ctx context.Context, offsetParam int32, limitParam int32, sortOrderParam string, sortByParam string,
	filterByParam []string, searchParam string, includeAncestors bool,
	complyToProjectId int32,
) (api.ImplResponse, error) {
	complianceInfo, complianceErr := NewComplianceParams(complyToProjectId, api.WORKLOADTYPE_DISTRIBUTED_INFERENCE, "")
	if complianceErr != nil {
		return api.Response(http.StatusBadRequest, complianceErr), nil
	}

	result, err := ts.ListTemplatesUnified(ctx, convutils.Ref(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE), offsetParam, limitParam, sortOrderParam, sortByParam,
		filterByParam, searchParam, complianceInfo, includeAncestors)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (ts *TemplatesService) ListTrainingTemplates(ctx context.Context, offsetParam int32, limitParam int32, sortOrderParam string, sortByParam string,
	filterByParam []string, searchParam string, includeAncestors bool,
	complyToProjectId int32,
) (api.ImplResponse, error) {
	complianceInfo, complianceErr := NewComplianceParams(complyToProjectId, api.WORKLOADTYPE_TRAINING, "")
	if complianceErr != nil {
		return api.Response(http.StatusBadRequest, complianceErr), nil
	}

	result, err := ts.ListTemplatesUnified(ctx, convutils.Ref(api.WORKLOADTYPE_TRAINING), offsetParam, limitParam, sortOrderParam, sortByParam,
		filterByParam, searchParam, complianceInfo, includeAncestors)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (ts *TemplatesService) ListWorkspaceTemplates(ctx context.Context, offsetParam int32, limitParam int32, sortOrderParam string, sortByParam string,
	filterByParam []string, searchParam string, includeAncestors bool,
	complyToProjectId int32,
) (api.ImplResponse, error) {
	complianceInfo, complianceErr := NewComplianceParams(complyToProjectId, api.WORKLOADTYPE_WORKSPACE, "")
	if complianceErr != nil {
		return api.Response(http.StatusBadRequest, complianceErr), nil
	}

	result, err := ts.ListTemplatesUnified(ctx, convutils.Ref(api.WORKLOADTYPE_WORKSPACE), offsetParam, limitParam, sortOrderParam, sortByParam,
		filterByParam, searchParam, complianceInfo, includeAncestors)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (ts *TemplatesService) PatchDistributedTemplate(ctx context.Context, templateId string, request api.DistributedTemplatePatchRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{
		patchMode:       true,
		fieldsToClear:   request.FieldsToClear,
		patchMetaFields: request.Meta,
	})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) PatchInferenceTemplate(ctx context.Context, templateId string, request api.InferenceTemplatePatchRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_INFERENCE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{
		patchMode:       true,
		fieldsToClear:   request.FieldsToClear,
		patchMetaFields: request.Meta,
	})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) PatchDistributedInferenceTemplate(ctx context.Context, templateId string, request api.DistributedInferenceTemplatePatchRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED_INFERENCE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{
		patchMode:       true,
		fieldsToClear:   request.FieldsToClear,
		patchMetaFields: request.Meta,
	})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) PatchTrainingTemplate(ctx context.Context, templateId string, request api.TrainingTemplatePatchRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_TRAINING, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{
		patchMode:       true,
		fieldsToClear:   request.FieldsToClear,
		patchMetaFields: request.Meta,
	})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) PatchWorkspaceTemplate(ctx context.Context, templateId string, request api.WorkspaceTemplatePatchRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{
		patchMode:       true,
		fieldsToClear:   request.FieldsToClear,
		patchMetaFields: request.Meta,
	})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) UpdateDistributedTemplate(ctx context.Context, templateId string, request api.DistributedTemplateUpdateRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) UpdateInferenceTemplate(ctx context.Context, templateId string, request api.InferenceTemplateUpdateRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_INFERENCE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) UpdateDistributedInferenceTemplate(ctx context.Context, templateId string, request api.DistributedInferenceTemplateUpdateRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_DISTRIBUTED_INFERENCE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) UpdateTrainingTemplate(ctx context.Context, templateId string, request api.TrainingTemplateUpdateRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_TRAINING, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}

func (ts *TemplatesService) UpdateWorkspaceTemplate(ctx context.Context, templateId string, request api.WorkspaceTemplateUpdateRequest) (api.ImplResponse, error) {
	unifiedRequest, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, &request)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	updated, err := ts.UpdateTemplate(ctx, templateId, unifiedRequest, updateOptions{})
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return legacyCompatibleTemplateResponse(updated.template)
}
