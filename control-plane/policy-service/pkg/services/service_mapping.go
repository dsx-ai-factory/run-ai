/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
)

type MappingApiService struct{}

func NewMappingApiService() api.SpecMappingAPIServicer {
	return MappingApiService{}
}

const runaiPrefix = "runai-"

const (
	formatterListDefaultLimit = 50
	formatterListMaxLimit     = 500
)

// withTenantMappingWriteLock attempts tenant locking before doing an update action
func (ms MappingApiService) withTenantMappingWriteLock(ctx context.Context, action func() (api.ImplResponse, error)) (api.ImplResponse, error) {
	release, lockErr := dao.AcquireTenantMappingLock(ctx)
	if lockErr != nil {
		return NewErrorResponse(lockErr), nil
	}
	defer release()

	return action()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// formatters
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (ms MappingApiService) CreateFormatter(ctx context.Context, request api.FormatterCreationRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("creating formatter %s", request.Name)

	// API level is only used for tenant scoped formatters, make sure they are not using the runai prefix
	if strings.HasPrefix(request.Name, runaiPrefix) {
		return NewErrorResponse(api.NewValidationError(fmt.Sprintf("prefix '%s' is not allowed for tenant scoped formatter", runaiPrefix))), nil
	}

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		// Enrich ArgsInfo with any $ARGS.named references not already present
		enrichedFields := make(map[string]api.FormatterField, len(request.Fields))
		for key, field := range request.Fields {
			EnrichFormatterFieldArgsInfo(&field)
			enrichedFields[key] = field
		}

		formatter := dbmodel.MappingFormatters{
			ID:          uuid.New(),
			Name:        request.Name,
			Description: request.Description,
			ScopeType:   api.MAPPINGSCOPE_TENANT,
			TenantId:    auth.TenantIdOf(ctx),
			Fields:      enrichedFields,

			CreatedAt: time.Now(),
			CreatedBy: auth.SubjectIdOf(ctx),
		}

		formatter.UpdatedAt = formatter.CreatedAt
		formatter.UpdatedBy = formatter.CreatedBy

		updated, err := ms.createFormatter(ctx, &formatter)
		if err != nil {
			return NewErrorResponse(err), nil
		}

		return api.Response(http.StatusOK, *updated), nil
	})
}

func (ms MappingApiService) ListFormatters(ctx context.Context, offset int32, limit int32) (api.ImplResponse, error) {
	off, lim := normalizeListParams(offset, limit, formatterListDefaultLimit, formatterListMaxLimit)

	response, err := ms.listFormatters(ctx, off, lim)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *response), nil
}

func (ms MappingApiService) GetFormatter(ctx context.Context, name string) (api.ImplResponse, error) {
	response, err := ms.getFormatter(ctx, name)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *response), nil
}

func (ms MappingApiService) DeleteFormatter(ctx context.Context, name string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleting formatter %s", name)

	// API level is only used for tenant scoped formatters, make sure they are not using the runai prefix
	if strings.HasPrefix(name, runaiPrefix) {
		return NewErrorResponse(api.NewValidationError(fmt.Sprintf("prefix '%s' is not allowed for tenant scoped formatter", runaiPrefix))), nil
	}

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		err := ms.deleteFormatter(ctx, name)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusNoContent, nil), nil
	})
}

func (ms MappingApiService) PatchFormatter(ctx context.Context, name string, request api.FormatterUpdateRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("updating formatter %s", name)

	// API level is only used for tenant scoped formatters, make sure they are not using the runai prefix
	if strings.HasPrefix(name, runaiPrefix) {
		return NewErrorResponse(api.NewValidationError(fmt.Sprintf("prefix '%s' is not allowed for tenant scoped formatter", runaiPrefix))), nil
	}

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		updated, err := ms.patchFormatter(ctx, name, &request)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusOK, updated), nil
	})
}

func (ms MappingApiService) DeleteFormatterFields(ctx context.Context, name string, request api.FormatterDeleteFieldsRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("updating fields from formatter %s", name)

	// API level is only used for tenant scoped formatters, make sure they are not using the runai prefix
	if strings.HasPrefix(name, runaiPrefix) {
		return NewErrorResponse(api.NewValidationError(fmt.Sprintf("prefix '%s' is not allowed for tenant scoped formatter", runaiPrefix))), nil
	}

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		if err := ms.deleteFormatterFields(ctx, name, request.Fields); err != nil {
			return NewErrorResponse(err), nil
		}
		updated, gerr := ms.getFormatter(ctx, name)
		if gerr != nil {
			return NewErrorResponse(gerr), nil
		}
		return api.Response(http.StatusOK, *updated), nil
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// field groups
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (ms MappingApiService) ListFieldGroups(ctx context.Context, offset int32, limit int32) (api.ImplResponse, error) {
	off, lim := normalizeListParams(offset, limit, fieldGroupListDefaultLimit, fieldGroupListMaxLimit)

	items, err := ms.listFieldGroups(ctx, off, lim)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, api.FieldGroupListResponse{Items: items}), nil
}

func (ms MappingApiService) CreateFieldGroup(ctx context.Context, request api.FieldGroupCreationRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("creating field group %s", request.Name)

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		now := time.Now().UTC()
		subject := auth.SubjectIdOf(ctx)
		tenantID := auth.TenantIdOf(ctx)

		group := dbmodel.MappingFieldGroups{
			ID:          uuid.New(),
			Name:        request.Name,
			Description: request.Description,
			ScopeType:   api.MAPPINGSCOPE_TENANT,
			TenantId:    tenantID,
			CreatedAt:   now,
			CreatedBy:   subject,
			UpdatedAt:   now,
			UpdatedBy:   subject,
		}

		group.Fields = addScopeToFields(request.Fields, api.MAPPINGSCOPE_TENANT)
		group.ItemSelectors = addScopeToItemSelectors(request.ItemSelectors, api.MAPPINGSCOPE_TENANT)

		merged, err := ms.createFieldGroup(ctx, &group)
		if err != nil {
			return NewErrorResponse(err), nil
		}

		return api.Response(http.StatusOK, *merged), nil
	})
}

func (ms MappingApiService) GetFieldGroup(ctx context.Context, name string) (api.ImplResponse, error) {
	fg, err := ms.getFieldGroup(ctx, name)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *fg), nil
}

func (ms MappingApiService) DeleteFieldGroup(ctx context.Context, name string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleting field group %s", name)

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		if err := ms.deleteFieldGroup(ctx, name, api.MAPPINGSCOPE_TENANT); err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusNoContent, nil), nil
	})
}

func (ms MappingApiService) PatchFieldGroup(ctx context.Context, name string, request api.FieldGroupUpdateRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("patching field group %s", name)

	if len(request.Fields) == 0 && len(request.ItemSelectors) == 0 && request.Description == nil {
		return NewErrorResponse(api.NewValidationError("patch must include description, fields, or itemSelectors")), nil
	}

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		updated, err := ms.patchFieldGroup(ctx, name, api.MAPPINGSCOPE_TENANT, &request, NewMappingConsistencyValidator(ctx, api.MAPPINGSCOPE_TENANT))
		if err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusOK, *updated), nil
	})
}

func (ms MappingApiService) DeleteFieldGroupFields(ctx context.Context, name string, request api.FieldGroupDeleteFieldsRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleting fields from field group %s", name)

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		updated, err := ms.deleteFieldGroupFields(ctx, name, api.MAPPINGSCOPE_TENANT, &request)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusOK, *updated), nil
	})
}

func (ms MappingApiService) DeleteFieldGroupItemSelectors(ctx context.Context, name string, request api.FieldGroupDeleteItemSelectorsRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleting item selectors from field group %s", name)

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		updated, err := ms.deleteFieldGroupItemSelectors(ctx, name, api.MAPPINGSCOPE_TENANT, &request)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusOK, *updated), nil
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// workload spec mappings
////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (ms MappingApiService) ListMappings(ctx context.Context, filterBy []string) (api.ImplResponse, error) {
	resp, err := ms.listMappings(ctx, filterBy)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, *resp), nil
}

func (ms MappingApiService) CreateMapping(ctx context.Context, request api.MappingCreUpdRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("creating mapping for %s/%s/%s",
		convutils.Deref(request.WorkloadType.Group, ""), request.WorkloadType.Kind, request.WorkloadType.Version)

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		now := time.Now().UTC()
		subject := auth.SubjectIdOf(ctx)
		tenantID := auth.TenantIdOf(ctx)

		specSelectors := request.SpecSelectors
		if specSelectors == nil {
			specSelectors = map[string]api.SpecSelector{}
		}

		row := &dbmodel.MappingSelectors{
			ID:          uuid.New(),
			GvkGroup:    convutils.Deref(request.WorkloadType.Group, ""),
			GvkKind:     request.WorkloadType.Kind,
			GvkVersion:  request.WorkloadType.Version,
			Description: request.Description,
			ScopeType:   api.MAPPINGSCOPE_TENANT,
			TenantId:    tenantID,
			CreatedAt:   now,
			CreatedBy:   subject,
			UpdatedAt:   now,
			UpdatedBy:   subject,
		}

		row.SpecSelectors = addScopeToSpecSelectors(specSelectors, api.MAPPINGSCOPE_TENANT)

		merged, err := ms.createMapping(ctx, row, NewMappingConsistencyValidator(ctx, api.MAPPINGSCOPE_TENANT))
		if err != nil {
			return NewErrorResponse(err), nil
		}

		return api.Response(http.StatusOK, *merged), nil
	})
}

func (ms MappingApiService) UpdateMapping(ctx context.Context, request api.MappingCreUpdRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("updating mapping for %s/%s/%s",
		convutils.Deref(request.WorkloadType.Group, ""), request.WorkloadType.Kind, request.WorkloadType.Version)

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		merged, err := ms.updateMapping(ctx, &request, api.MAPPINGSCOPE_TENANT, NewMappingConsistencyValidator(ctx, api.MAPPINGSCOPE_TENANT))
		if err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusOK, *merged), nil
	})
}

func (ms MappingApiService) DeleteMapping(ctx context.Context, request api.MappingDeletionRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleting mapping for %s/%s/%s",
		convutils.Deref(request.WorkloadType.Group, ""), request.WorkloadType.Kind, request.WorkloadType.Version)

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		err := ms.deleteMapping(ctx, &request, api.MAPPINGSCOPE_TENANT)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusNoContent, nil), nil
	})
}

func (ms MappingApiService) DeleteMappingSpecSelectors(ctx context.Context, request api.MappingDeleteSpecSelectorsReq) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleting spec selectors from mapping %s/%s/%s",
		convutils.Deref(request.WorkloadType.Group, ""), request.WorkloadType.Kind, request.WorkloadType.Version)

	return ms.withTenantMappingWriteLock(ctx, func() (api.ImplResponse, error) {
		merged, err := ms.deleteMappingSpecSelectors(ctx, &request, api.MAPPINGSCOPE_TENANT)
		if err != nil {
			return NewErrorResponse(err), nil
		}
		return api.Response(http.StatusOK, *merged), nil
	})
}

func addScopeToFields(fields map[string]api.FieldGroupField, scopeType api.MappingScope) map[string]api.FieldGroupFieldsWithScopeValue {
	result := make(map[string]api.FieldGroupFieldsWithScopeValue)
	for key, value := range fields {
		withScope := api.FieldGroupFieldsWithScopeValue{
			Description:   value.Description,
			Formatter:     value.Formatter,
			SpecRef:       value.SpecRef,
			Type:          value.Type,
			ItemSelectors: value.ItemSelectors,
			Scope:         scopeType,
		}
		result[key] = withScope
	}
	return result
}

func addScopeToItemSelectors(is map[string]api.ItemSelector, scopeType api.MappingScope) map[string]api.ItemSelectorsWithScopeValue {
	result := make(map[string]api.ItemSelectorsWithScopeValue)
	for key, value := range is {
		// Enrich ArgsInfo with any $ARGS.named references not already present
		EnrichItemSelectorArgsInfo(&value)
		withScope := api.ItemSelectorsWithScopeValue{
			Description: value.Description,
			Query:       value.Query,
			ArgsInfo:    value.ArgsInfo,
			ArgsEnum:    value.ArgsEnum,
			Scope:       scopeType,
		}
		result[key] = withScope
	}
	return result
}

func addScopeToSpecSelectors(m map[string]api.SpecSelector, scopeType api.MappingScope) map[string]api.SpecSelectorsWithScopeValue {
	result := make(map[string]api.SpecSelectorsWithScopeValue, len(m))
	for key, value := range m {
		// Enrich ArgsInfo with any $ARGS.named references not already present
		EnrichSpecSelectorArgsInfo(&value)
		withScope := api.SpecSelectorsWithScopeValue{
			Description: value.Description,
			Query:       value.Query,
			ArgsInfo:    value.ArgsInfo,
			ArgsEnum:    value.ArgsEnum,
			FieldGroups: value.FieldGroups,
			Dependency:  value.Dependency,
			Scope:       scopeType,
		}
		result[key] = withScope
	}
	return result
}
