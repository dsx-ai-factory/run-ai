/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jsschema "github.com/santhosh-tekuri/jsonschema/v6"
	resourcev1 "k8s.io/api/resource/v1"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/projects"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/scope"
	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	backend "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// rctObjectSpecSchemaRaw is the Kubernetes OpenAPI v3 document for the
// resource.k8s.io/v1 API group, captured once from a live cluster. Loaded
// at process start and used to build rctObjectSpecSchema. See
// schemas/README.md for how to refresh.
//
//go:embed schemas/rct_object_spec.schema.json
var rctObjectSpecSchemaRaw []byte

// rctObjectSpecKey is the fully-qualified schema name (matches the K8s
// OpenAPI naming convention) of the ResourceClaimSpec type inside the
// embedded document. The compiler resolves refs like
// "#/components/schemas/io.k8s.api.resource.v1.DeviceClaim" natively.
const rctObjectSpecKey = "io.k8s.api.resource.v1.ResourceClaimSpec"

// rctObjectSpecSchema is compiled once from the embedded schema at init and
// reused by every CreateRct/UpdateRctById call. Nil until initRctObjectSpecValidator
// succeeds; nil-check before use.
var rctObjectSpecSchema *jsschema.Schema

func init() {
	initRctObjectSpecValidator()
}

// RctApiService is a service that implements the logic for the ResourceClaimTemplates servicer.
// This service should implement the business logic for every endpoint for the ResourceClaimTemplates API.
type RctApiService struct{}

// NewRctApiService creates a default api service.
func NewRctApiService() *RctApiService {
	return &RctApiService{}
}

// CreateRct - Create resource claim template.
func (s *RctApiService) CreateRct(ctx context.Context, request api.RctCreationRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityName(request.Meta.Name).WithClusterId(safeStrPtr(request.Meta.ClusterId))

	validationError := ValidateAssetCreateRequestMeta(ctx, &request.Meta, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfCreateRequest(&request.Meta), string(assetKind), backend.ACTION_CREATE, convutils.Deref(request.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	if metaErr := validateK8sObjectMeta(request.Spec.ObjectMeta); metaErr != nil {
		return NewErrorResponse(metaErr), nil
	}

	if specErr := validateRctObjectSpec(request.Spec.ObjectSpec); specErr != nil {
		return NewErrorResponse(specErr), nil
	}

	nameValidationQuery, err := scope.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		request.Meta.Scope,
		request.Meta.ProjectId,
		request.Meta.DepartmentId,
		request.Meta.ClusterId,
		request.Meta.Private,
		assetKind,
		nil,
	)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := dao.CreateRct(ctx, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	audit.AuditFromCtx(ctx).WithEntityId(result.Meta.Id)

	return api.Response(http.StatusAccepted, result), nil
}

// ListRcts - List resource claim templates.
func (s *RctApiService) ListRcts(ctx context.Context, offset int32, limit int32) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE

	// WAIT_FOR_OFER this API does not work with the "classic" queryParams so just initial it to empty
	// when we work with the "standard" filter this needs to be refactored
	queryParams, queryError := query_types.NewQueryParams("", "", "", false, 0, "", "", nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	_, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	// WAIT_FOR_OFER currently the code does not support filtering by scope, so the
	// list will filter the result by the entire scope the user is allowed to see.
	// once we support filtering, we need to filter by scope as well.
	result, actionError := dao.ListRcts(ctx, offset, limit)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	patchingPointers := response_patching.LoadResponsePatchingPtrs(result)

	patchErr := projects.PatchProjectNames(ctx, patchingPointers)
	if patchErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("list %s: Failed to patch project names: %v", assetKind, patchErr)
	}

	return api.Response(http.StatusOK, result), nil
}

// GetRctById - Retrieve a resource claim template.
func (s *RctApiService) GetRctById(ctx context.Context, id string) (api.ImplResponse, error) {
	asset, actionError := dao.GetRctById(ctx, id)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	if _, _, authError := query_types.ApplyRbacRulesGetById(ctx, asset.Meta); authError != nil {
		return NewErrorResponse(authError), nil
	}

	return api.Response(http.StatusOK, asset), nil
}

// UpdateRctById - Update a resource claim template.
func (s *RctApiService) UpdateRctById(ctx context.Context, id string, request api.RctUpdateRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id).WithEntityName(request.Meta.Name)

	if metaErr := validateK8sObjectMeta(request.Spec.ObjectMeta); metaErr != nil {
		return NewErrorResponse(metaErr), nil
	}

	if specErr := validateRctObjectSpec(request.Spec.ObjectSpec); specErr != nil {
		return NewErrorResponse(specErr), nil
	}

	assetMeta, getError := dao.GetAssetMetaById(ctx, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	if kindErr := requireRctKind(assetMeta, id); kindErr != nil {
		return NewErrorResponse(kindErr), nil
	}
	audit.AuditFromCtx(ctx).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_UPDATE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	nameValidationQuery, err := scope.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		assetMeta.Scope,
		assetMeta.ProjectId,
		assetMeta.DepartmentId,
		assetMeta.ClusterId,
		assetMeta.Private,
		assetKind,
		&id,
	)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := dao.UpdateRct(ctx, id, request, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	return api.Response(http.StatusAccepted, result), nil
}

// DeleteRctById - Delete a resource claim template.
func (s *RctApiService) DeleteRctById(ctx context.Context, id string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id)

	assetMeta, getError := dao.GetAssetMetaById(ctx, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	if kindErr := requireRctKind(assetMeta, id); kindErr != nil {
		return NewErrorResponse(kindErr), nil
	}
	audit.AuditFromCtx(ctx).WithEntityName(assetMeta.Name).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_DELETE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	if actionError := dao.DeleteRct(ctx, assetMeta, dao.HardDelete); actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	return NewDeletionResponse(), nil
}

// requireRctKind guards Update/Delete against IDs that belong to a different
// asset kind. dao.GetAssetMetaById is kind-agnostic — without this check a
// caller could POST/DELETE at /resource-claim-templates/{id} with the id of
// e.g. a compute asset and drive it through RCT's RBAC and DAO paths. We
// return the same 404 shape that dao.GetRctById returns for a missing RCT so
// the endpoint's namespace behaves symmetrically and doesn't leak the fact
// that the id exists under a different kind.
func requireRctKind(assetMeta *dbmodel.AssetMeta, id string) *api.Error {
	if assetMeta.Kind == api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE {
		return nil
	}
	return api.NewError(http.StatusNotFound,
		fmt.Sprintf("resource claim template with id %s not found", id), nil)
}

// validateRctObjectSpec is a thin adapter over the generic
// validateK8sObjectSpec, specialised to the K8s ResourceClaimSpec type.
// Kept here (rather than inlined at each callsite) so future RCT-only
// validation can slot in alongside the generic call without changing the
// callers, and so the RCT-specific schema label lives with the RCT service.
func validateRctObjectSpec(objectSpec map[string]any) *api.Error {
	return validateK8sObjectSpec[resourcev1.ResourceClaimSpec](
		objectSpec, rctObjectSpecSchema,
		"spec.objectSpec", "resource.k8s.io/v1 ResourceClaimSpec",
	)
}

func initRctObjectSpecValidator() {
	// santhosh-tekuri/jsonschema natively resolves "#/components/schemas/..."
	// refs against the same document once we add it as a compiler resource.
	// No ref-normalization or manual inlining needed — the compiler handles
	// the full graph.
	var doc any
	if err := json.Unmarshal(rctObjectSpecSchemaRaw, &doc); err != nil {
		// Leaves rctObjectSpecSchema nil; runtime validation path returns
		// 503 so the misconfig is obvious in dev builds still on the placeholder.
		logging.Errorf("rct: failed to parse embedded object-spec schema: %v", err)
		return
	}

	compiler := jsschema.NewCompiler()
	const resourceURL = "mem://rct_object_spec.schema.json"
	if err := compiler.AddResource(resourceURL, doc); err != nil {
		logging.Errorf("rct: failed to add object-spec schema resource: %v", err)
		return
	}

	schema, err := compiler.Compile(resourceURL + "#/components/schemas/" + rctObjectSpecKey)
	if err != nil {
		logging.Errorf("rct: failed to compile object-spec schema for %s: %v", rctObjectSpecKey, err)
		return
	}
	rctObjectSpecSchema = schema
	logging.Infof("rct: object-spec validator initialized")
}
