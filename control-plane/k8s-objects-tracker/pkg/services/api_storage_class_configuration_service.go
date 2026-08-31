/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"github.com/jinzhu/copier"
	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	authorizationApi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/clusters"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/common"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-application-packages/patching"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"github.com/run-ai/runai/runai-application-packages/validation"
	k8sObjectsService "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
	k8sObjectsApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var (
	supportedUnits = []string{
		"MB", "MiB", "GB", "GiB", "TB", "TiB",
	}

	AllSupportedUnits = datastructs.NewSet(supportedUnits...)

	defaultSupportedAccessModes = api.PvcSupportedAccessModes{
		ReadWriteOnce: convutils.Ref(true),
		ReadOnlyMany:  convutils.Ref(true),
		ReadWriteMany: convutils.Ref(true),
	}

	defaultAccessModes = api.PvcAccessModes{
		ReadWriteOnce: convutils.Ref(false),
		ReadOnlyMany:  convutils.Ref(false),
		ReadWriteMany: convutils.Ref(false),
	}

	defaultSupportedVolumeMode = api.VolumeModeCustomizationSupportedValues{
		Filesystem: convutils.Ref(true),
		Block:      convutils.Ref(true),
	}

	onlyLetters = regexp.MustCompile(`[a-zA-Z]+$`)
)

// StorageClassConfigurationAPIService is a service that implements the logic for the StorageClassConfigurationAPIServicer
// This service should implement the business logic for every endpoint for the StorageClassConfigurationAPI API.
// Include any external packages or services that will be required by this service.
type StorageClassConfigurationAPIService struct{}

// NewStorageClassConfigurationAPIService creates a default api service
func NewStorageClassConfigurationAPIService() *StorageClassConfigurationAPIService {
	return &StorageClassConfigurationAPIService{}
}

// CreateStorageClassConfiguration - create storage class configuration
func (s *StorageClassConfigurationAPIService) CreateStorageClassConfiguration(ctx context.Context, storageClassConfigurationCreationRequest api.StorageClassConfigurationCreationRequest) (api.ImplResponse, error) {
	// command can be executed by CSP admin, and in this case, we take the tenantId that he provided in the request
	// of the request. otherwise, just take it from the context
	tenantIdForDao, err := s.extractTenantIdFromCtxAndRequestedId(ctx, storageClassConfigurationCreationRequest.TenantId)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	_, clusterSvcErr := clusters.Get().GetClusterByUuid(ctx, storageClassConfigurationCreationRequest.ScopeId)
	if clusterSvcErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get cluster, err: %v", clusterSvcErr)
		return NewErrorResponse(api.NewError(int(clusterSvcErr.Code), clusterSvcErr.Message, nil)), nil
	}

	if storageClassConfigurationCreationRequest.ScopeType == "" {
		storageClassConfigurationCreationRequest.ScopeType = string(authorizationApi.SCOPETYPE_CLUSTER)
	}

	if storageClassConfigurationCreationRequest.Name == "" {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "mandatory name not provided", nil)), nil
	}

	isPermitted, authErr := isPermittedScope(ctx, *tenantIdForDao, storageClassConfigurationCreationRequest.ScopeId)
	if authErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get permitted scopes", authErr)), nil
	}

	if !isPermitted {
		return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
	}

	record := &models.StorageClassConfiguration{
		StorageClassConfigurationFilters: models.StorageClassConfigurationFilters{
			TenantId:  *tenantIdForDao,
			ScopeType: &storageClassConfigurationCreationRequest.ScopeType,
			ScopeId:   &storageClassConfigurationCreationRequest.ScopeId,
			Name:      &storageClassConfigurationCreationRequest.Name,
		},
		Permissions:   storageClassConfigurationCreationRequest.Permissions,
		Customization: storageClassConfigurationCreationRequest.Customization,
	}

	res, err := dao.CreateStorageClassConfiguration(ctx, record)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, mapModelToApi(res)), nil
}

// DeleteStorageClassConfigurationById - delete storage class configuration by id
func (s *StorageClassConfigurationAPIService) DeleteStorageClassConfigurationById(ctx context.Context, id string, tenantId int32) (api.ImplResponse, error) {
	// command can be executed by CSP admin, and in this case, we take the tenantId that he provided in the request
	// of the request. otherwise, just take it from the context
	tenantIdForDao, err := s.extractTenantIdFromCtxAndRequestedId(ctx, convutils.Ref(tenantId))
	if err != nil {
		return NewErrorResponse(err), nil
	}

	isPermitted, authErr := isPermittedScope(ctx, *tenantIdForDao, id)
	if authErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get permitted scopes", authErr)), nil
	}

	if !isPermitted {
		return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
	}

	err = dao.DeleteStorageClassConfiguration(ctx, id, *tenantIdForDao)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

// GetStorageClassConfigurationById - get storage class configuration by id
func (s *StorageClassConfigurationAPIService) GetStorageClassConfigurationById(ctx context.Context, id string, tenantId int32) (api.ImplResponse, error) {
	// CSP is allowed to send tenantId which is not the same as the tenantId in the context
	tenantIdForDao, err := s.extractTenantIdFromCtxAndRequestedId(ctx, convutils.Ref(tenantId))
	if err != nil {
		return NewErrorResponse(err), nil
	}

	isPermitted, authErr := isPermittedScope(ctx, *tenantIdForDao, id)
	if authErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get permitted scopes", authErr)), nil
	}

	if !isPermitted {
		return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
	}

	res, err := dao.GetStorageClassConfigurationById(ctx, id, *tenantIdForDao)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, mapModelToApi(res)), nil
}

// ListStorageClassConfiguration - list of storage class configuration
func (s *StorageClassConfigurationAPIService) ListStorageClassConfiguration(ctx context.Context, filterBy []string, sortBy string, sortOrder string, offset int32, limit int32) (api.ImplResponse, error) {
	permittedScopes, authErr := authorizationClient.ParsePermittedScopes(ctx)
	if authErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get permitted scopes", authErr)), nil
	}

	return s.listStorageClassConfiguration(ctx, filterBy, sortBy, sortOrder, offset, limit, permittedScopes)
}

func (s *StorageClassConfigurationAPIService) ListStorageClassConfigurationInt(ctx context.Context, filterBy []string, sortBy string, sortOrder string, offset int32, limit int32) (api.ImplResponse, error) {
	return s.listStorageClassConfiguration(ctx, filterBy, sortBy, sortOrder, offset, limit, nil)
}

func (s *StorageClassConfigurationAPIService) listStorageClassConfiguration(ctx context.Context,
	filterBy []string, sortBy string, sortOrder string,
	offset int32, limit int32,
	permittedScopes *authorizationApi.PermittedScopes,
) (api.ImplResponse, error) {
	configuration := models.StorageClassConfiguration{}
	parser := queryapi.NewParser(configuration.SortAndFilterFieldsToColumn())

	filters, filterErr := parser.Filters(filterBy)
	if filterErr != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy), filterErr)), nil
	}

	var sortColumn string
	if sortBy != "" {
		var sortError error
		sortColumn, sortError = parser.GetMapping(sortBy)
		if sortError != nil {
			return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("sort %s invalid", sortBy), sortError)), nil
		}
	}

	storageClassConfigurations, err := dao.ListStorageClassConfigurations(ctx, filters, sortColumn, sortOrder, offset, limit, permittedScopes)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	response := api.StorageClassConfigurationListResponse{
		StorageClassConfigurations: make([]api.StorageClassConfiguration, 0),
	}

	for _, storageClassConfiguration := range storageClassConfigurations {
		response.StorageClassConfigurations = append(response.StorageClassConfigurations, mapModelToApi(&storageClassConfiguration))
	}

	recordsCount := int32(len(storageClassConfigurations))
	if recordsCount == limit {
		response.Next = recordsCount + offset
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *StorageClassConfigurationAPIService) UpdateStorageClassConfigurationById(ctx context.Context, id string,
	request api.StorageClassConfigurationUpdateRequest, tenantId int32,
) (api.ImplResponse, error) {
	tenantIdForDao, err := s.extractTenantIdFromCtxAndRequestedId(ctx, convutils.Ref(tenantId))
	if err != nil {
		return NewErrorResponse(err), nil
	}

	isPermitted, authErr := isPermittedScope(ctx, *tenantIdForDao, id)
	if authErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get permitted scopes", authErr)), nil
	}

	if !isPermitted {
		return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
	}

	record, err := dao.GetStorageClassConfigurationById(ctx, id, *tenantIdForDao)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	record.Permissions = request.Permissions

	if request.Customization != nil {
		if record.Customization == nil {
			record.Customization = &api.StorageClassCustomization{}
		}

		copyErr := copier.Copy(record.Customization, request.Customization)
		if copyErr != nil {
			return api.Response(http.StatusInternalServerError, copyErr), nil
		}

		var attrsErr *api.Error
		s.validateCustomization(ctx, &attrsErr, record.ScopeId, request.Customization)
		if attrsErr != nil {
			return api.Response(int(attrsErr.Code), attrsErr), nil
		}
	}

	if patching.CheckIfAllMembersAreNil(record.Permissions) {
		record.Permissions = nil
	}

	if patching.CheckIfAllMembersAreNil(record.Customization) {
		record.Customization = nil
	}

	res, err := dao.UpdateStorageClassConfiguration(ctx, id, record, *tenantIdForDao)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, mapModelToApi(res)), nil
}

// PatchStorageClassConfigurationById - patch storage class configuration by id
func (s *StorageClassConfigurationAPIService) PatchStorageClassConfigurationById(ctx context.Context, id string,
	storageClassConfigurationPatchRequest api.StorageClassConfigurationPatchRequest, tenantId int32,
) (api.ImplResponse, error) {
	// CSP is allowed to send tenantId which is not the same as the tenantId in the context
	tenantIdForDao, err := s.extractTenantIdFromCtxAndRequestedId(ctx, convutils.Ref(tenantId))
	if err != nil {
		return NewErrorResponse(err), nil
	}

	isPermitted, authErr := isPermittedScope(ctx, *tenantIdForDao, id)
	if authErr != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get permitted scopes", authErr)), nil
	}

	if !isPermitted {
		return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
	}

	record, err := dao.GetStorageClassConfigurationById(ctx, id, *tenantIdForDao)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	s.patchPermissions(record, storageClassConfigurationPatchRequest.Permissions)

	customizationErr := s.patchCustomization(ctx, record, storageClassConfigurationPatchRequest.Customization)
	if customizationErr != nil {
		return NewErrorResponse(customizationErr), nil
	}

	res, err := dao.UpdateStorageClassConfiguration(ctx, id, record, *tenantIdForDao)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, mapModelToApi(res)), nil
}

func (s *StorageClassConfigurationAPIService) patchPermissions(record *models.StorageClassConfiguration, permissions *api.StorageClassPermissions) {
	if permissions == nil {
		return
	}

	if record.Permissions == nil {
		record.Permissions = &api.StorageClassPermissions{}
	}

	patching.PatchField(&record.Permissions.AllowedForAssets, permissions.AllowedForAssets)
	patching.PatchField(&record.Permissions.AllowedForWorkloads, permissions.AllowedForWorkloads)
	patching.PatchField(&record.Permissions.AllowedForPersistentVolumes, permissions.AllowedForPersistentVolumes)
	patching.PatchField(&record.Permissions.AllowedForEphemeralVolumes, permissions.AllowedForEphemeralVolumes)
	patching.PatchField(&record.Permissions.AllowDataSharing, permissions.AllowDataSharing)

	if patching.CheckIfAllMembersAreNil(record.Permissions) {
		record.Permissions = nil
	}
}

func (s *StorageClassConfigurationAPIService) patchCustomization(ctx context.Context, record *models.StorageClassConfiguration, customization *api.StorageClassConfigurationPatchRequestCustomization) *api.Error {
	if customization == nil {
		return nil
	}

	if record.Customization == nil {
		record.Customization = &api.StorageClassCustomization{}
	}

	// we are doing the attributes patch separately because of its unique logic, and then we clear
	// the attributes from the request so it will not affect the result when we do the "regular" patch
	// ideally, we should enhance the patching library to support itemized fields but I prefer to have
	// some more user cases to figure out how to do code it as a general solution
	//
	s.patchAttributes(record, customization)

	customization.Attributes = nil

	patching.ApplyPatch(record.Customization, customization)

	if patching.CheckIfAllMembersAreNil(record.Customization) {
		record.Customization = nil
		return nil
	}

	//
	//   check the fields
	//
	var err *api.Error
	s.validateCustomization(ctx, &err, record.ScopeId, record.Customization)
	return err
}

func (s *StorageClassConfigurationAPIService) patchAttributes(record *models.StorageClassConfiguration, customization *api.StorageClassConfigurationPatchRequestCustomization) *api.Error {
	// no attributes to patch
	if customization.Attributes == nil && len(customization.AttributesToRemove) == 0 {
		return nil
	}

	// if this is the first time attributes are set, take them from the patch request as is
	if record.Customization.Attributes == nil {
		record.Customization.Attributes = customization.Attributes
		return nil
	}

	// there are attributes in both the request and the existing record, so need to apply the patched
	// attributes over the existing attributes.
	existingKeys, existingAttributes := attributesToMap(record.Customization.Attributes)
	_, toPatchAttributes := attributesToMap(customization.Attributes)

	for toPatchKey, toPatchAttr := range toPatchAttributes {
		existingAttributes[toPatchKey] = toPatchAttr
		existingKeys.Add(toPatchKey)
	}

	// remove those attributes that user requested to remove
	for _, toRemoveKey := range customization.AttributesToRemove {
		delete(existingAttributes, toRemoveKey)
		delete(existingKeys, toRemoveKey)
	}

	// now create the updated array. we sort it by the key name which is not a real requirement but
	// will make life easier on both the administrator and our test, to get a sorted list
	updatedSetOfKeys := existingKeys.Keys()
	sort.Strings(updatedSetOfKeys)

	record.Customization.Attributes = make([]api.PvcAddedAttribute, 0, len(existingAttributes))

	for _, key := range updatedSetOfKeys {
		record.Customization.Attributes = append(record.Customization.Attributes, existingAttributes[key])
	}

	if len(record.Customization.Attributes) == 0 {
		record.Customization.Attributes = nil
	}

	return nil
}

func attributesToMap(attrs []api.PvcAddedAttribute) (datastructs.Set[string], map[string]api.PvcAddedAttribute) {
	keys := datastructs.NewSet[string]()
	asMap := make(map[string]api.PvcAddedAttribute)

	for _, attr := range attrs {
		asMap[attr.Key] = attr
		keys.Add(attr.Key)
	}

	return keys, asMap
}

func (s *StorageClassConfigurationAPIService) validateCustomization(ctx context.Context, err **api.Error, clusterId *string, customization *api.StorageClassCustomization) {
	if err == nil || *err != nil {
		return
	}

	if customization == nil {
		return
	}

	var capabilitiesChecker *checker.ClusterCapabilitiesChecker

	if clusterId != nil {
		capabilitiesChecker = checker.NewClusterCapabilitiesChecker(*clusterId)
	}

	s.validateAccessModes(err, customization.AccessMode)
	s.validateVolumeModes(err, customization.VolumeMode)
	s.validateClaimSize(err, customization.ClaimSize)
	s.validateSpecAttributes(ctx, err, capabilitiesChecker, customization.Attributes)
}

func (s *StorageClassConfigurationAPIService) validateAccessModes(err **api.Error, accessModes *api.AccessModeCustomization) {
	// the function is part of a chain of validations, if previous validation raised error no need to add more errors
	if err == nil || *err != nil {
		return
	}

	if accessModes == nil {
		return
	}

	//
	//  for required accessMode, at least one supported access mode sohuld be provided.
	//  in addition, align the defaults with the supported access modes.
	//
	if accessModes.Default == nil {
		accessModes.Default = &api.PvcAccessModes{}
	}

	if accessModes.SupportedValues != nil {
		supportedValues := 0

		if isSupportedValue(accessModes.SupportedValues.ReadOnlyMany) {
			supportedValues += 1
		} else {
			accessModes.Default.ReadOnlyMany = nil
		}

		if isSupportedValue(accessModes.SupportedValues.ReadWriteMany) {
			supportedValues += 1
		} else {
			accessModes.Default.ReadWriteMany = nil
		}

		if isSupportedValue(accessModes.SupportedValues.ReadWriteOnce) {
			supportedValues += 1
		} else {
			accessModes.Default.ReadWriteOnce = nil
		}

		if supportedValues == 0 && convutils.BoolIsTrue(accessModes.Required) {
			*err = api.NewError(http.StatusBadRequest, "at least one accessMode should be supported when accessMode is required", nil)
			return
		}

		if patching.CheckIfAllMembersAreNil(accessModes.Default) {
			accessModes.Default = nil
		}
	}
}

func (s *StorageClassConfigurationAPIService) validateVolumeModes(err **api.Error, volumeModes *api.VolumeModeCustomization) {
	// the function is part of a chain of validations, if previous validation raised error no need to add more errors
	if err == nil || *err != nil {
		return
	}

	if volumeModes == nil {
		return
	}

	//
	//  for required volumeMode, at least one supported modes should be provided.
	//  in addition, align the defaults with the supported modes.
	//
	if volumeModes.SupportedValues != nil {
		supportedValues := 0

		if isSupportedValue(volumeModes.SupportedValues.Filesystem) {
			supportedValues += 1
		} else if convutils.Deref(volumeModes.Default, "") == api.PVCVOLUMEMODE_FILESYSTEM {
			*err = api.NewError(http.StatusBadRequest, fmt.Sprintf("default value '%s' is not marked as supported", api.PVCVOLUMEMODE_FILESYSTEM), nil)
			return
		}

		if isSupportedValue(volumeModes.SupportedValues.Block) {
			supportedValues += 1
		} else if convutils.Deref(volumeModes.Default, "") == api.PVCVOLUMEMODE_BLOCK {
			*err = api.NewError(http.StatusBadRequest, fmt.Sprintf("default value '%s' is not marked as supported", api.PVCVOLUMEMODE_BLOCK), nil)
			return
		}

		if supportedValues == 0 && convutils.BoolIsTrue(volumeModes.Required) {
			*err = api.NewError(http.StatusBadRequest, "at least one volumeMode should be supported when volumeMode is required", nil)
			return
		}
	}
}

func (s *StorageClassConfigurationAPIService) validateClaimSize(err **api.Error, claimSize *api.ClaimSizeCustomization) {
	// the function is part of a chain of validations, if previous validation raised error no need to add more errors
	if err == nil || *err != nil {
		return
	}

	if claimSize == nil {
		return
	}

	var supportedUnits datastructs.Set[string]

	if claimSize.SupportedUnits != nil {
		supportedUnits = datastructs.NewSet[string](claimSize.SupportedUnits...)
	} else {
		supportedUnits = AllSupportedUnits
	}

	if len(supportedUnits) == 0 {
		*err = api.NewError(http.StatusBadRequest, "claimSize: at least one unit should be supported", nil)
		return
	}

	defaultSize := matchValueToUnit(err, "default", claimSize.Default, supportedUnits)
	minSize := matchValueToUnit(err, "min", claimSize.Min, supportedUnits)
	maxSize := matchValueToUnit(err, "max", claimSize.Max, supportedUnits)
	stepSize := matchValueToUnit(err, "step", claimSize.Step, supportedUnits)

	if *err != nil {
		return
	}

	if minSize == 0 || maxSize == 0 || stepSize == 0 || defaultSize == 0 {
		*err = api.NewError(http.StatusBadRequest, "claimSize: default, min, max and step have positive values", nil)
		return
	}

	if minSize > 0 && maxSize > 0 {
		if minSize > maxSize {
			*err = api.NewError(http.StatusBadRequest, "claimSize: max should be no less than min", nil)
			return
		}

		if stepSize > 0 && stepSize > maxSize-minSize {
			*err = api.NewError(http.StatusBadRequest, "claimSize: invalid value for step: should be no greater than the difference between max and min", nil)
			return
		}
	}

	if defaultSize > 0 {
		if minSize > 0 && defaultSize < minSize {
			*err = api.NewError(http.StatusBadRequest, "claimSize: default should be no less than min", nil)
			return
		}

		if maxSize > 0 && defaultSize > maxSize {
			*err = api.NewError(http.StatusBadRequest, "claimSize: default should be no greater than max", nil)
			return
		}
	}

	if stepSize > 0 && defaultSize > 0 {
		//
		//  if min value is not specified take 0 for the calculation, so step 3 will be valid for values 3, 6, 9, ...
		//
		var realMinSize = minSize
		if minSize < 0 {
			realMinSize = 0
		}

		diffFromMin := int64(defaultSize - realMinSize)
		if diffFromMin%int64(stepSize) != 0 {
			if minSize > 0 {
				*err = api.NewError(http.StatusBadRequest, fmt.Sprintf("claimSize: default value is not aligned with %s steps from the minimal value %s", *claimSize.Step, *claimSize.Min), nil)
			} else {
				*err = api.NewError(http.StatusBadRequest, fmt.Sprintf("claimSize: default value is not aligned with steps of %s", *claimSize.Step), nil)
			}
			return
		}
	}
}

func (s *StorageClassConfigurationAPIService) validateSpecAttributes(ctx context.Context, err **api.Error, capabilities *checker.ClusterCapabilitiesChecker, attrs []api.PvcAddedAttribute) {
	if err == nil || *err != nil {
		return
	}

	if attrs == nil {
		return
	}

	supportAttr, checkerErr := capabilities.ValidateClusterCapability(ctx, checker.MinVersionOfPvcAddedAttributes)
	if checkerErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to validate cluster version: %v", checkerErr)
	} else if !supportAttr {
		*err = api.NewValidationError("the cluster hosting this storage class does not support PVC added attributes yet")
		return
	}

	keys := datastructs.NewSet[string]()

	for _, attr := range attrs {
		keyPrefix := fmt.Sprintf("attribute '%s': ", attr.Key)

		if keys.Has(attr.Key) {
			*err = api.NewValidationError(keyPrefix + "key is multiply defined")
			return
		}
		keys.Add(attr.Key)

		if attr.Format == nil {
			*err = api.NewValidationError(keyPrefix + "mandatory format is not provided")
			return
		}

		if attr.MinValue != nil || attr.MaxValue != nil || attr.Step != nil {
			if *attr.Format != api.PVCADDEDATTRIBUTEFORMAT_NUMBER {
				*err = api.NewValidationError(keyPrefix + "minValue, maxValue and step are only supported for number attributes")
				return
			}

			if attr.MinValue != nil && attr.MaxValue != nil && *attr.MinValue > *attr.MaxValue {
				*err = api.NewValidationError(keyPrefix + "minValue should be no less than maxValue")
				return
			}

			if attr.Step != nil && *attr.Step < 0 {
				*err = api.NewValidationError(keyPrefix + "step should be greater than zero")
				return
			}
		}

		if convutils.IsStringHasValue(attr.ValidationRegexp) {
			if *attr.Format != api.PVCADDEDATTRIBUTEFORMAT_TEXT {
				*err = api.NewValidationError(keyPrefix + "validationRegexp is only supported for text attributes")
				return
			}

			_, regexpErr := regexp.Compile(*attr.ValidationRegexp)
			if regexpErr != nil {
				*err = api.NewValidationError(keyPrefix + fmt.Sprintf("invalid validationRegexp: %s", *attr.ValidationRegexp))
				return
			}
		}

		//
		//  check the validity of the default value
		//
		if attr.DefaultValue != nil {
			attrConfig, copyErr := copyToClientFormat(&attr, keyPrefix)
			if copyErr != nil {
				*err = copyErr
				return
			}

			defaultPrefix := fmt.Sprintf("default of attribute '%s': ", attr.Key)
			attrConfig.ValidationError.Unset() // for default, we only want the system pre-defined errors

			defaultErr := k8sObjectsService.ValidatePvcAttValueVsConfig(attrConfig, defaultPrefix, *attr.DefaultValue)
			if defaultErr != nil {
				*err = api.NewValidationError(defaultErr.Message)
				return
			}
		}
	}
}

func copyToClientFormat(src *api.PvcAddedAttribute, keyPrefix string) (*k8sObjectsApi.PvcAddedAttribute, *api.Error) {
	asJson, jsonErr := json.Marshal(src)
	if jsonErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, keyPrefix+"failed to load attribute info", jsonErr)
	}

	var result k8sObjectsApi.PvcAddedAttribute

	jsonErr = json.Unmarshal(asJson, &result)
	if jsonErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, keyPrefix+"failed to load attribute info", jsonErr)
	}

	return &result, nil
}

// extractTenantIdFromCtxAndRequestedId - Validates that the user is allowed to send tenantId and returns the requested tenantId or the tenantId from the context
func (s *StorageClassConfigurationAPIService) extractTenantIdFromCtxAndRequestedId(ctx context.Context, requestedTenantId *int32) (*int32, *api.Error) {
	if requestedTenantId != nil && runaictx.IsManagementTenant(ctx) && *requestedTenantId != 0 {
		return requestedTenantId, nil
	}

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &tenantId, nil
}

func isPermittedScope(ctx context.Context, tenantId int32, clusterId string) (bool, error) {
	permittedScopes, authErr := authorizationClient.ParsePermittedScopes(ctx)
	if authErr != nil {
		return false, authErr
	}

	if permittedScopes.System != nil {
		return true, nil
	}

	if permittedScopes.Tenants != nil {
		for _, currTenant := range permittedScopes.Tenants {
			if currTenant == strconv.Itoa(int(tenantId)) {
				return true, nil
			}
		}
	}

	if permittedScopes.Clusters != nil {
		for _, currCluster := range permittedScopes.Clusters {
			if currCluster == clusterId {
				return true, nil
			}
		}
	}

	return false, nil
}

func mapModelToApi(storageClassConfiguration *models.StorageClassConfiguration) api.StorageClassConfiguration {
	result := api.StorageClassConfiguration{
		Id:            storageClassConfiguration.ID.String(),
		TenantId:      storageClassConfiguration.TenantId,
		Name:          *storageClassConfiguration.Name,
		ScopeId:       *storageClassConfiguration.ScopeId,
		ScopeType:     *storageClassConfiguration.ScopeType,
		Permissions:   storageClassConfiguration.Permissions,
		Customization: storageClassConfiguration.Customization,
	}

	// for more clarity, we write defaults of some fields in the response implicitly
	if result.Permissions != nil {
		result.Permissions.AllowedForAssets = convutils.Ref(convutils.Deref(result.Permissions.AllowedForAssets, true))
		result.Permissions.AllowedForWorkloads = convutils.Ref(convutils.Deref(result.Permissions.AllowedForWorkloads, true))
		result.Permissions.AllowedForEphemeralVolumes = convutils.Ref(convutils.Deref(result.Permissions.AllowedForEphemeralVolumes, true))
		result.Permissions.AllowedForPersistentVolumes = convutils.Ref(convutils.Deref(result.Permissions.AllowedForPersistentVolumes, true))
		result.Permissions.AllowDataSharing = convutils.Ref(convutils.Deref(result.Permissions.AllowDataSharing, false))
	}

	if result.Customization != nil {
		if result.Customization.AccessMode != nil {
			result.Customization.AccessMode.Required = convutils.Ref(convutils.Deref(result.Customization.AccessMode.Required, false))

			if result.Customization.AccessMode.SupportedValues != nil {
				result.Customization.AccessMode.SupportedValues.ReadWriteOnce = convutils.Ref(convutils.Deref(result.Customization.AccessMode.SupportedValues.ReadWriteOnce, true))
				result.Customization.AccessMode.SupportedValues.ReadWriteMany = convutils.Ref(convutils.Deref(result.Customization.AccessMode.SupportedValues.ReadWriteMany, true))
				result.Customization.AccessMode.SupportedValues.ReadOnlyMany = convutils.Ref(convutils.Deref(result.Customization.AccessMode.SupportedValues.ReadOnlyMany, true))
			} else {
				result.Customization.AccessMode.SupportedValues = &defaultSupportedAccessModes
			}

			if result.Customization.AccessMode.Default != nil {
				result.Customization.AccessMode.Default.ReadWriteOnce = convutils.Ref(convutils.Deref(result.Customization.AccessMode.Default.ReadWriteOnce, false))
				result.Customization.AccessMode.Default.ReadWriteMany = convutils.Ref(convutils.Deref(result.Customization.AccessMode.Default.ReadWriteMany, false))
				result.Customization.AccessMode.Default.ReadOnlyMany = convutils.Ref(convutils.Deref(result.Customization.AccessMode.Default.ReadOnlyMany, false))
			} else {
				result.Customization.AccessMode.Default = &defaultAccessModes
			}
		}

		if result.Customization.VolumeMode != nil {
			result.Customization.VolumeMode.Required = convutils.Ref(convutils.Deref(result.Customization.VolumeMode.Required, false))

			if result.Customization.VolumeMode.SupportedValues != nil {
				result.Customization.VolumeMode.SupportedValues.Filesystem = convutils.Ref(convutils.Deref(result.Customization.VolumeMode.SupportedValues.Filesystem, true))
				result.Customization.VolumeMode.SupportedValues.Block = convutils.Ref(convutils.Deref(result.Customization.VolumeMode.SupportedValues.Block, true))
			} else {
				result.Customization.VolumeMode.SupportedValues = &defaultSupportedVolumeMode
			}
		}

		if result.Customization.ClaimSize != nil {
			if result.Customization.ClaimSize.SupportedUnits == nil {
				result.Customization.ClaimSize.SupportedUnits = supportedUnits
			}
		}

	}

	return result
}

// in storage class configuration, unspecified flag means that it is allowed, supported, etc
func isSupportedValue(flag *bool) bool {
	return flag == nil || *flag
}

func matchValueToUnit(err **api.Error, description string, value *string, supportedValues datastructs.Set[string]) float64 {
	if err == nil || *err != nil {
		return -1
	}

	if value == nil {
		return -1
	}

	//
	//  note that the units are MB, MiB, etc but the value is given without the B
	//
	unit := onlyLetters.FindString(*value)
	if !supportedValues.Has(unit + "B") {
		*err = api.NewError(http.StatusBadRequest, fmt.Sprintf("%s: unit suffix '%s' is not supported", description, unit), nil)
		return -1
	}

	realSize, convErr := validation.ToFloat(*value)
	if convErr != nil {
		*err = api.NewError(http.StatusBadRequest, convErr.Error(), nil)
		return -1
	}

	return realSize
}
