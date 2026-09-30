/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package asset_cluster_status

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	k8s_objects_tracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// we expect asset to be updated in objects tracker within 60 seconds. if less time elapsed
// since the recent update of the asset and the objects tracker is not updated - indicate no
// status rather than failed status
const timeToGiveUpCreationStat = 60 * time.Second

type AssetClusterStatus struct {
	// status of resources of given resource type (secret, pvc, etc)
	resourceStatusPerType map[openapi.ClusterResourceType]AssetIdStatusMap

	// the set of assets that were sent with given kind, in order to make sure that calling the
	// calculation multiple time with same kind -> assets per kind does not change (this is done
	// because we want to do the API call to objects tracker once per kind)
	assetIdsPerKind map[api.AssetKind]datastructs.Set[string]

	// for some cases it is needed to look on the status that was determined based on the sync info table
	syncBasedStatus map[string]*dbmodel.SyncStatusRec

	patchingInfo map[string]response_patching.ResponsePatchingPtrs

	capabilitiesCheckerPerCluster map[string]*checker.ClusterCapabilitiesChecker

	permittedScopes *auth.PermittedScopes
}

type AssetIdStatusMap map[string]*api.AssetClusterStatusInfo

// When k8s-objects-tracker learns about a new asset kind (e.g.
// resource-claim-template, and any future k8s-object asset), FOUR maps in this
// file must be updated together, or the status projection will silently fall
// through to nil / to the wrong message:
//
//  1. assetKindToClusterResourceType (below)      — required for
//     HasClusterResourceMapping(kind) to return true so PatchStatusInfo stops
//     short-circuiting.
//  2. assetKindToTargetAssetType                  — routes to
//     calculateAssetStatus vs calculateAssetStatusFromCredential. For k8s
//     resources (RCT, etc.) use targetAssetType_K8SResource.
//  3. getAssetClusterResourceName switch          — tells the calculator the
//     name(s) to look up in the tracker's response.
//  4. getAssetIssuesFoundMessage / getAssetNoIssuesFoundMessage /
//     getAssetNotCreatedMessage                   — the K8SResource branch
//     hardcodes "data source ..." today; RCT needs distinct copy
//     ("resource claim template ...") or the tracker path will surface
//     wrong-family messages.
var assetKindToClusterResourceType = map[api.AssetKind]openapi.ClusterResourceType{
	api.ASSETKIND_PVC:             openapi.CLUSTERRESOURCETYPE_PVC,
	api.ASSETKIND_ACCESS_KEY:      openapi.CLUSTERRESOURCETYPE_SECRET,
	api.ASSETKIND_PASSWORD:        openapi.CLUSTERRESOURCETYPE_SECRET,
	api.ASSETKIND_DOCKER_REGISTRY: openapi.CLUSTERRESOURCETYPE_SECRET,
	api.ASSETKIND_S3:              openapi.CLUSTERRESOURCETYPE_SECRET,
	api.ASSETKIND_GIT:             openapi.CLUSTERRESOURCETYPE_SECRET,
	api.ASSETKIND_GENERIC_SECRET:  openapi.CLUSTERRESOURCETYPE_SECRET,
	api.ASSETKIND_CONFIG_MAP:      openapi.CLUSTERRESOURCETYPE_CONFIG_MAP,
	api.ASSETKIND_SECRET_VOLUME:   openapi.CLUSTERRESOURCETYPE_SECRET,
	api.ASSETKIND_ENVIRONMENT:     openapi.CLUSTERRESOURCETYPE_SECRET,
	api.ASSETKIND_NGC_API_KEY:     openapi.CLUSTERRESOURCETYPE_SECRET,
	// resource-claim-template: intentionally absent until the tracker learns
	// about resource.k8s.io/v1 resources. See the doc block above for the full
	// checklist of maps that must be updated together.
}

// HasClusterResourceMapping reports whether the given asset kind has a
// k8s-objects-tracker projection (a native k8s resource type the tracker can
// report status for). Kinds without a mapping — RCT today — have their status
// driven entirely by sync_status and must not fall through to the tracker
// call, which would fire an HTTP request with an empty resourceType.
func HasClusterResourceMapping(kind api.AssetKind) bool {
	_, ok := assetKindToClusterResourceType[kind]
	return ok
}

type UniqueKey struct {
	ScopeId      string
	ScopeType    string
	resourceName string
}

// targetAssetType is an enum that represents the underlying(target) asset(resource) behind the asset
// for example, S3 replies on a credential assets, so the target asset type is credential
// AccessKey, Password, DockerRegistry, GenericSecret are all base on secret assets,
// and PVC, ConfigMap are general k8s resources
type targetAssetType int

const (
	// TargetAssetType_K8SResource implies the underlying asset is a k8s resource such as PVC and ConfigMap
	targetAssetType_K8SResource = iota
	// targetAssetType_Secret implies the underlying asset is a k8s secret resource
	targetAssetType_Secret
	// targetAssetType_Secret implies the underlying asset is a credential asset
	targetAssetType_Credential
)

var assetKindToTargetAssetType = map[api.AssetKind]targetAssetType{
	api.ASSETKIND_PVC:             targetAssetType_K8SResource,
	api.ASSETKIND_ACCESS_KEY:      targetAssetType_Secret,
	api.ASSETKIND_PASSWORD:        targetAssetType_Secret,
	api.ASSETKIND_DOCKER_REGISTRY: targetAssetType_Secret,
	api.ASSETKIND_S3:              targetAssetType_Credential,
	api.ASSETKIND_GIT:             targetAssetType_Credential,
	api.ASSETKIND_GENERIC_SECRET:  targetAssetType_Secret,
	api.ASSETKIND_CONFIG_MAP:      targetAssetType_K8SResource,
	api.ASSETKIND_SECRET_VOLUME:   targetAssetType_Credential,
	api.ASSETKIND_ENVIRONMENT:     targetAssetType_Credential,
	api.ASSETKIND_NGC_API_KEY:     targetAssetType_Secret,
}

func NewAssetClusterStatus(patchingInfo map[string]response_patching.ResponsePatchingPtrs, permittedScopes *auth.PermittedScopes) *AssetClusterStatus {
	result := AssetClusterStatus{
		resourceStatusPerType:         make(map[openapi.ClusterResourceType]AssetIdStatusMap),
		assetIdsPerKind:               make(map[api.AssetKind]datastructs.Set[string]),
		patchingInfo:                  patchingInfo,
		capabilitiesCheckerPerCluster: make(map[string]*checker.ClusterCapabilitiesChecker),
		permittedScopes:               permittedScopes,
	}

	for _, curPatching := range patchingInfo {
		clusterId := curPatching.Meta.ClusterId
		if clusterId != nil {
			if result.capabilitiesCheckerPerCluster[*clusterId] == nil {
				result.capabilitiesCheckerPerCluster[*clusterId] = checker.NewClusterCapabilitiesChecker(*clusterId)
			}
		}
	}
	return &result
}

func (acs *AssetClusterStatus) StoreSyncBasedInfo(syncBasedStatus map[string]*dbmodel.SyncStatusRec) {
	acs.syncBasedStatus = syncBasedStatus
}

// LoadAssetsStatus - Load can be executed multiple times, if it is executed with the same assetKind but different assets it will throw an error.
// So it can be loaded with the same assets for each assetKind.
// For example if pvcs 1,3,6 were loaded and the next time we call Load with pvcs 1,2,3, it will throw an error
// If pvcs 1,3,5 were loaded and the next time we want to load secrets 1,6,9 it will load them
func (acs *AssetClusterStatus) LoadAssetsStatus(ctx context.Context, assetKind api.AssetKind) *api.Error {
	// This condition check that make sure that if the function is called with the same assetKind,
	// the same set of assets is given
	assetIdsOfThisKind := datastructs.NewSet[string]()
	for assetId, curPatch := range acs.patchingInfo {
		if curPatch.Meta.Kind == assetKind {
			assetIdsOfThisKind.Add(assetId)
		}
	}

	if acs.assetIdsPerKind[assetKind] != nil {
		if len(acs.assetIdsPerKind) != len(assetIdsOfThisKind) {
			return api.NewError(http.StatusInternalServerError, "patchingInfo contains assets that were not loaded", nil)
		}
		for assetId := range assetIdsOfThisKind {
			if !acs.assetIdsPerKind[assetKind].Has(assetId) {
				return api.NewError(http.StatusInternalServerError, "patchingInfo contains assets that were not loaded", nil)
			}
		}
		return nil // all assets of this kind were already loaded, nothing else to do
	}

	acs.assetIdsPerKind[assetKind] = assetIdsOfThisKind

	clusterResourceType := assetKindToClusterResourceType[assetKind]

	clusterResources, err := k8s_objects_tracker.Get().GetClusterResourcesStatus(ctx, &clusterResourceType, nil, nil, nil)
	if err != nil {
		return api.NewError(int(err.Code), err.Message, nil)
	}

	cleanUnauthorizedIssues(ctx, acs.permittedScopes, clusterResources.Entries)

	targetType := assetKindToTargetAssetType[assetKind]
	if targetType == targetAssetType_Credential {
		err := acs.calculateAssetStatusFromCredential(ctx, clusterResources.Entries, assetKind)
		if err != nil {
			return api.NewError(int(err.Code), err.Message, nil)
		}
	} else {
		err := acs.calculateAssetStatus(ctx, clusterResources.Entries, assetKind)
		if err != nil {
			return api.NewError(int(err.Code), err.Message, nil)
		}
	}

	return nil
}

func (acs *AssetClusterStatus) GetAssetsStatus() (map[string]*api.AssetClusterStatusInfo, *api.Error) {
	assetIdStatusMap := make(map[string]*api.AssetClusterStatusInfo)
	for id, curPatchInfo := range acs.patchingInfo {
		curPatchClusterResourceType, found := assetKindToClusterResourceType[curPatchInfo.Meta.Kind]
		if !found {
			// If the asset kind is not supported, we skip it
			continue
		}
		assetStatusMap, found := acs.resourceStatusPerType[curPatchClusterResourceType]
		if !found {
			// If we don't have the asset status, we skip it
			continue
		} else {
			status, found := assetStatusMap[id]
			if !found {
				// If we don't have the asset status, we skip it
				continue
			}
			assetIdStatusMap[id] = status
		}
	}
	return assetIdStatusMap, nil
}

func (acs *AssetClusterStatus) calculateAssetStatusFromCredential(ctx context.Context,
	clusterResources []openapi.ClusterObjectStatus,
	assetKind api.AssetKind,
) *api.Error {
	clusterResourceMap := getClusterResourceMap(clusterResources)

	assetUsedCredentialIdMap, credentialIdStatusMap, err := acs.getAssetCredentialsStatus(ctx, clusterResourceMap)
	if err != nil {
		return err
	}

	for assetId, curPatchInfo := range acs.patchingInfo {
		assetCredentialIds := assetUsedCredentialIdMap[assetId]
		for _, assetCredentialId := range assetCredentialIds {
			statusInfo, found := credentialIdStatusMap[assetCredentialId]
			if !found {
				if curPatchInfo.Meta.Scope == api.SCOPE_TENANT {
					statusInfo = newTenantScopeStatus(assetKind)
					acs.addAssetStatusToMap(ctx, curPatchInfo.Meta.Id, assetKindToClusterResourceType[curPatchInfo.Meta.Kind], statusInfo)
				}
			} else {
				clusterResourceType := assetKindToClusterResourceType[curPatchInfo.Meta.Kind]
				acs.addAssetStatusToMap(ctx, assetId, clusterResourceType, statusInfo)
			}
		}
	}

	return nil
}

func (acs *AssetClusterStatus) calculateAssetStatus(ctx context.Context,
	clusterResources []openapi.ClusterObjectStatus,
	assetKind api.AssetKind,
) *api.Error {
	logger := logging.WithCtx(runaictx.LoggerHook(ctx))
	clusterResourceMap := getClusterResourceMap(clusterResources)
	for assetId, curPatchInfo := range acs.patchingInfo {
		if curPatchInfo.Meta.Kind != assetKind {
			continue
		}

		clusterResourceType := assetKindToClusterResourceType[curPatchInfo.Meta.Kind]
		//
		// for tenant scope assets we have a generic status
		// if the asset is private (=tenant scope), we continue in the asset status calculation
		//
		clusterId := convutils.Deref(curPatchInfo.Meta.ClusterId, "")
		isPrivate := convutils.BoolIsTrue(curPatchInfo.Meta.Private)
		if clusterId == "" && !isPrivate {
			acs.addAssetStatusToMap(ctx, curPatchInfo.Meta.Id, clusterResourceType, newTenantScopeStatus(assetKind))
			continue
		}

		//
		//  calculate status based on agent sync info, if such info exists
		//
		syncBasedStatus := acs.syncBasedStatus[assetId]
		if syncBasedStatus != nil {
			recentSentToAgent := max(syncBasedStatus.SentToAgentApply, syncBasedStatus.SentToAgentDelete)
			if curPatchInfo.Meta.UpdateCount > recentSentToAgent {
				acs.addAssetStatusToMap(ctx, curPatchInfo.Meta.Id, clusterResourceType, acs.newAgentNotUpdated(curPatchInfo.Meta))
				continue
			}

			if curPatchInfo.Meta.UpdateCount == syncBasedStatus.ProcessedByAgent && syncBasedStatus.Message != "" {
				acs.addAssetStatusToMap(ctx, curPatchInfo.Meta.Id, clusterResourceType, acs.newAgentSyncFailed(curPatchInfo.Meta, syncBasedStatus.Message))
				continue
			}
		}

		//
		//   calculate status based on k8s objects tracker
		//
		scopeInfo := getScopeInfoForStatus(curPatchInfo.Meta)
		resources := getAssetClusterResourceName(ctx, curPatchInfo)
		if resources == nil {
			logger.Errorf("asset spec type is not supported %s", assetKind)
			continue
		}

		for _, resource := range resources {
			uniqueKey := newUniqueKey(scopeInfo.ScopeId, string(scopeInfo.Scope), resource)
			var statusInfo api.AssetClusterStatusInfo
			currClusterResource, found := clusterResourceMap[uniqueKey]
			// If we didn't find the resource in the cluster,
			// we check the cluster version to see if it supports the resource status
			if !found {
				capabilitiesChecker := checker.NewClusterCapabilitiesChecker(clusterId)

				isClusterVersionSupportStatus, err := capabilitiesChecker.ValidateClusterCapability(ctx, checker.MinVersionOfClusterReportingResourceStatus)
				if err != nil && !isPrivate {
					return api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
				}

				if !isClusterVersionSupportStatus && !isPrivate {
					statusInfo = newClusterVersionNotSupportedStatus()
				} else {
					// If the cluster version supports the resource status and didn't have the resource, there are two situations:
					// (1) information about the asset was not yet sent to k8s objects tracker. this is reasonable in proximity
					//     to the creation of the asset and we will just write Creating/Deleting (on new agent) NoStatus (asset sync)
					// (2) the asset was not created at all. this is mostly likely in assets-sync mechanism, in the new agent
					//     mechanism if there was an issue creating the asset the agent will report that immediately. but since
					//     the asset will not appear in the list of assets as long as it is not known to objects tracker, we
					//     must indicate not created in this case
					timeFromUpdate := time.Since(curPatchInfo.Meta.UpdatedAt)
					if !curPatchInfo.Meta.UpdatedAt.IsZero() && timeFromUpdate < timeToGiveUpCreationStat {
						statusInfo = acs.newNotYetUpdated(ctx, curPatchInfo.Meta)
					} else {
						statusInfo = newNotCreatedStatusInfo(scopeInfo.ScopeId, string(scopeInfo.Scope), assetKind)
					}
				}

			} else {
				statusInfo = api.AssetClusterStatusInfo{
					Status: getStatusFromClusterResource(currClusterResource),
					Issues: getResourceIssues(currClusterResource),
				}
				if len(statusInfo.Issues) == 0 {
					getAssetNoIssuesFoundMessage(&statusInfo, assetKind)
				} else {
					getAssetIssuesFoundMessage(&statusInfo, assetKind)
				}
			}
			acs.addAssetStatusToMap(ctx, assetId, clusterResourceType, statusInfo)
		}
	}

	return nil
}

func (acs *AssetClusterStatus) addAssetStatusToMap(ctx context.Context, assetId string, resourceType openapi.ClusterResourceType, status api.AssetClusterStatusInfo) {
	assetStatusMap, found := acs.resourceStatusPerType[resourceType]
	if !found {
		assetStatusMap = make(AssetIdStatusMap)
		acs.resourceStatusPerType[resourceType] = assetStatusMap
	}
	// The status of the asset may be effected by multiple resources, keep the first non NO_STATUS status
	// If there isn't a status, use the first status
	if assetStatusMap[assetId] == nil {
		assetStatusMap[assetId] = &status
		return
	}
	// if the current status is no status, override with a status
	if assetStatusMap[assetId].Status == api.ASSETSYNCSTATUSENUM_NO_STATUS {
		assetStatusMap[assetId] = &status
		return
	}
	logger := logging.WithCtx(runaictx.LoggerHook(ctx))
	logger.Debugf("status %+v isn't store on asset id %s as it already has a no issue start", status, assetId)
}

func (acs *AssetClusterStatus) getAssetCredentialsStatus(ctx context.Context,
	clusterResourceMap map[UniqueKey]openapi.ClusterObjectStatus,
) (map[string][]string, map[string]api.AssetClusterStatusInfo, *api.Error) {
	logger := logging.WithCtx(runaictx.LoggerHook(ctx))
	credentialIds := make([]string, 0)
	credentialIdMap := make(map[string][]string)
	for assetId, curPatchInfo := range acs.patchingInfo {
		// We don't calculate the status of tenant/system scope assets
		if curPatchInfo.Meta.Scope == api.SCOPE_TENANT || curPatchInfo.Meta.Scope == api.SCOPE_SYSTEM {
			continue
		}

		if assetKindToTargetAssetType[curPatchInfo.Meta.Kind] != targetAssetType_Credential {
			logger.Debugf("asset %s isn't based on a credential", curPatchInfo.Meta.Name)
			continue
		}

		curCredentialIds := getCredentialIdUsedByAsset(curPatchInfo)
		if curCredentialIds == nil {
			logger.Debugf("asset %s isn't based on a credential", curPatchInfo.Meta.Name)
			continue
		}
		credentialIds = append(credentialIds, curCredentialIds...)
		credentialIdMap[assetId] = curCredentialIds
	}

	credentials, err := dao.ListCredentialsAssets(credentialIds, runaidb.DB().WithContext(ctx))
	if err != nil {
		return nil, nil, api.NewError(int(err.Code), err.Message, nil)
	}

	credentialStatusMap := make(map[string]api.AssetClusterStatusInfo)
	for _, credential := range credentials {
		scopeInfo := scope.ScopeInfoOf(scope.Scope(credential.Scope), credential.ProjectId, credential.DepartmentId, credential.ClusterId, credential.TenantId)
		resourceName := getCredentialResourceName(credential)
		uniqueKey := newUniqueKey(scopeInfo.ScopeId, string(scopeInfo.Scope), resourceName)

		var statusInfo api.AssetClusterStatusInfo
		currClusterResource, found := clusterResourceMap[uniqueKey]
		// If we didn't find the resource in the cluster,
		// we check the cluster version to see if it supports the resource status
		if !found {
			clusterId := convutils.Deref(credential.ClusterId, "")
			// If it is tenant scope asset, we don't calculate its status
			if clusterId == "" {
				statusInfo = newTenantScopeStatus(api.ASSETKIND_S3)
				credentialStatusMap[credential.Id] = statusInfo
				continue
			}

			isClusterSupportStatus := false

			if acs.capabilitiesCheckerPerCluster[clusterId] != nil {
				var clusterErr error

				isClusterSupportStatus, clusterErr = acs.capabilitiesCheckerPerCluster[clusterId].ValidateClusterCapability(ctx, checker.MinVersionOfClusterReportingResourceStatus)
				if clusterErr != nil {
					return nil, nil, api.NewError(http.StatusServiceUnavailable, clusterErr.Error(), nil)
				}
			}

			if !isClusterSupportStatus {
				statusInfo = newClusterVersionNotSupportedStatus()
			} else {
				// If the cluster version supports the resource status and didn't have the resource, we set the status to NOT_CREATED
				// Sending S3 hardcoded because it's the same message for all datasource that contains a credential (s3/git)
				statusInfo = newNotCreatedStatusInfo(scopeInfo.ScopeId, string(scopeInfo.Scope), api.ASSETKIND_S3)
			}
		} else {
			statusInfo = api.AssetClusterStatusInfo{
				Status: getStatusFromClusterResource(currClusterResource),
				Issues: getResourceIssues(currClusterResource),
			}
			if len(statusInfo.Issues) == 0 {
				// Sending S3 hardcoded because it's the same message for all datasource that contains a credential (s3/git)
				getAssetNoIssuesFoundMessage(&statusInfo, api.ASSETKIND_S3)
			} else {
				// Sending S3 hardcoded because it's the same message for all datasource that contains a credential (s3/git)
				getAssetIssuesFoundMessage(&statusInfo, api.ASSETKIND_S3)
			}
		}
		credentialStatusMap[credential.Id] = statusInfo
	}
	return credentialIdMap, credentialStatusMap, nil
}

func cleanUnauthorizedIssues(ctx context.Context, permittedScopes *auth.PermittedScopes, clusterResources []openapi.ClusterObjectStatus) *api.Error {
	// get the user permitted scopes to read the resource type and clean non permitted scope issues from the resources

	for _, clusterResource := range clusterResources {
		if clusterResource.Issues == nil {
			continue
		} else {
			authorizedIssues := make([]openapi.ClusterObjectIssue, 0)
			for _, issue := range clusterResource.Issues {
				if issue.ScopeType == string(api.SCOPETYPE_PROJECT) {
					if permittedScopes.AllowedProjects().Has(issue.ScopeId) {
						authorizedIssues = append(authorizedIssues, issue)
					}
				}
				if issue.ScopeType == string(api.SCOPE_DEPARTMENT) {
					if permittedScopes.AllowedDepartments().Has(issue.ScopeId) {
						authorizedIssues = append(authorizedIssues, issue)
					}
				}
				if issue.ScopeType == string(api.SCOPETYPE_CLUSTER) {
					if permittedScopes.AllowedCluster().Has(issue.ScopeId) {
						authorizedIssues = append(authorizedIssues, issue)
					}
				}
			}
			clusterResource.Issues = authorizedIssues
		}
	}

	return nil
}

func getStatusFromClusterResource(clusterResource openapi.ClusterObjectStatus) api.AssetSyncStatusEnum {
	if len(clusterResource.Issues) == 0 {
		return api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND
	}
	return api.ASSETSYNCSTATUSENUM_ISSUES_FOUND
}

func getResourceIssues(clusterResource openapi.ClusterObjectStatus) []api.AssetClusterStatusIssue {
	issues := make([]api.AssetClusterStatusIssue, 0)
	for _, clusterIssue := range clusterResource.Issues {
		assetIssue := api.AssetClusterStatusIssue{
			ScopeType: api.ScopeType(clusterIssue.ScopeType),
			ScopeId:   clusterIssue.ScopeId,
			Issue:     api.AssetClusterStatusIssueEnum(clusterIssue.Issue),
		}
		issues = append(issues, assetIssue)
	}
	return issues
}

func newNotCreatedStatusInfo(scopeId string, scopeType string, assetKind api.AssetKind) api.AssetClusterStatusInfo {
	statusInfo := api.AssetClusterStatusInfo{
		Status: api.ASSETSYNCSTATUSENUM_ISSUES_FOUND,
		Issues: []api.AssetClusterStatusIssue{
			{
				ScopeType: api.ScopeType(scopeType),
				ScopeId:   scopeId,
				Issue:     api.ASSETCLUSTERSTATUSISSUEENUM_NOT_CREATED,
			},
		},
	}
	getAssetNotCreatedMessage(&statusInfo, assetKind)
	return statusInfo
}

func (acs *AssetClusterStatus) newAgentNotUpdated(assetMeta *api.AssetMeta) api.AssetClusterStatusInfo {
	if assetMeta.DeletedAt.IsZero() {
		return api.AssetClusterStatusInfo{
			Status: api.ASSETSYNCSTATUSENUM_CREATING,
		}
	} else {
		return api.AssetClusterStatusInfo{
			Status: api.ASSETSYNCSTATUSENUM_DELETING,
		}
	}
}

func (acs *AssetClusterStatus) newAgentSyncFailed(assetMeta *api.AssetMeta, message string) api.AssetClusterStatusInfo {
	if assetMeta.DeletedAt.IsZero() {
		return api.AssetClusterStatusInfo{
			Status:  api.ASSETSYNCSTATUSENUM_FAILED,
			Message: message,
		}
	} else {
		return api.AssetClusterStatusInfo{
			Status:  api.ASSETSYNCSTATUSENUM_DELETION_FAILED,
			Message: message,
		}
	}
}

func (acs *AssetClusterStatus) newNotYetUpdated(ctx context.Context, assetMeta *api.AssetMeta) api.AssetClusterStatusInfo {
	//
	//  for clusters supporting agent sync, we can use the soft deletion indication to figure out if they
	//  are doing creating/deleting of assets
	//
	if assetMeta != nil {
		clusterId := assetMeta.ClusterId
		if clusterId != nil {
			capabilities := acs.capabilitiesCheckerPerCluster[*clusterId]
			agentSyncSupport, err := capabilities.ValidateClusterCapability(ctx, checker.MinVersionOfAssetsAgentSync)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to check agent sync capability: %v", err.Error())
			} else if agentSyncSupport {
				if assetMeta.DeletedAt.IsZero() {
					return api.AssetClusterStatusInfo{
						Status: api.ASSETSYNCSTATUSENUM_CREATING,
					}
				} else {
					return api.AssetClusterStatusInfo{
						Status: api.ASSETSYNCSTATUSENUM_DELETING,
					}
				}
			}
		}
	}

	//
	//   in any other case of no clarity just write no status updated yet
	//
	statusInfo := api.AssetClusterStatusInfo{
		Status:  api.ASSETSYNCSTATUSENUM_CREATING,
		Message: "Status of the asset not yet updated from the cluster",
	}

	return statusInfo
}

func newClusterVersionNotSupportedStatus() api.AssetClusterStatusInfo {
	return api.AssetClusterStatusInfo{
		Status:  api.ASSETSYNCSTATUSENUM_NO_STATUS,
		Message: "Status cannot be displayed because the current version of the cluster is not up to date",
	}
}

func newUniqueKey(scopeId string, scopeType string, resourceName string) UniqueKey {
	return UniqueKey{ScopeId: scopeId, ScopeType: scopeType, resourceName: resourceName}
}

func getCredentialResourceName(credential dbmodel.CredentialsAssetSpecs) string {
	if credential.AssetsPassword != nil && credential.AssetsPassword.ExistingSecretName != nil {
		return *credential.AssetsPassword.ExistingSecretName
	}
	if credential.AssetsAccessKey != nil && credential.AssetsAccessKey.ExistingSecretName != nil {
		return *credential.AssetsAccessKey.ExistingSecretName
	}
	if credential.AssetsGenericSecret != nil && credential.AssetsGenericSecret.ExistingSecretName != nil {
		return *credential.AssetsGenericSecret.ExistingSecretName
	}

	resourceName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[credential.Kind], credential.Name)
	return resourceName
}

func getClusterResourceMap(clusterResources []openapi.ClusterObjectStatus) map[UniqueKey]openapi.ClusterObjectStatus {
	clusterResourceMap := make(map[UniqueKey]openapi.ClusterObjectStatus)
	for _, clusterResource := range clusterResources {
		uniqueKey := newUniqueKey(clusterResource.ScopeId, clusterResource.ScopeType, clusterResource.ResourceName)
		clusterResourceMap[uniqueKey] = clusterResource
	}
	return clusterResourceMap
}

const runaiSupportUrl = "https://www.nvidia.com/en-eu/support/enterprise/#contact-us"

func newTenantScopeStatus(assetKind api.AssetKind) api.AssetClusterStatusInfo {
	return api.AssetClusterStatusInfo{
		Status:  api.ASSETSYNCSTATUSENUM_NO_STATUS,
		Message: getAccountScopeMessage(assetKind),
	}
}

func getAccountScopeMessage(assetKind api.AssetKind) string {
	switch assetKindToTargetAssetType[assetKind] {
	case targetAssetType_K8SResource:
		return "Status cannot be displayed because the data source's scope is an account"
	case targetAssetType_Secret:
		return "Status cannot be displayed because the credentials' scope is an account"
	case targetAssetType_Credential:
		if assetKind == api.ASSETKIND_ENVIRONMENT {
			return "Status cannot be displayed because the environment's credentials' scope is account"
		}
		return "Status cannot be displayed because the data source's credentials' scope is account"
	default:
		return ""
	}
}

// The K8SResource branches in the three getAsset*Message helpers below
// hardcode "data source ..." copy — that's correct for PVC/ConfigMap but
// wrong for future k8s-object kinds (resource-claim-template, etc.). When
// adding a new k8s-object kind to assetKindToClusterResourceType, extend
// each K8SResource branch to key by assetKind (or add a new target type).
func getAssetNotCreatedMessage(statusInfo *api.AssetClusterStatusInfo, assetKind api.AssetKind) {
	statusInfo.Url = runaiSupportUrl
	switch assetKindToTargetAssetType[assetKind] {
	case targetAssetType_K8SResource:
		statusInfo.Message = "The data source couldn't be synchronized in the cluster"
	case targetAssetType_Secret:
		statusInfo.Message = "The credential couldn't be synchronized in the cluster"
	case targetAssetType_Credential:
		if assetKind == api.ASSETKIND_ENVIRONMENT {
			statusInfo.Message = "The environment's credentials couldn't be created in the cluster"
			return
		}
		statusInfo.Message = "The data source's credentials couldn't be created in the cluster"
	}
}

func getAssetIssuesFoundMessage(statusInfo *api.AssetClusterStatusInfo, assetKind api.AssetKind) {
	statusInfo.Url = runaiSupportUrl
	switch assetKindToTargetAssetType[assetKind] {
	case targetAssetType_K8SResource:
		statusInfo.Message = "Issues were found while propagating the data source to the data source's scope"
	case targetAssetType_Secret:
		statusInfo.Message = "Issues were found while propagating the credential to the credentials' scope"
	case targetAssetType_Credential:
		if assetKind == api.ASSETKIND_ENVIRONMENT {
			statusInfo.Message = "Issues were found while propagating the environment's credentials to the credentials' scope"
			return
		}
		statusInfo.Message = "Issues were found while propagating the data source's credentials to the credentials' scope"
	}
}

func getAssetNoIssuesFoundMessage(statusInfo *api.AssetClusterStatusInfo, assetKind api.AssetKind) {
	switch assetKindToTargetAssetType[assetKind] {
	case targetAssetType_K8SResource:
		statusInfo.Message = "No issues were found while propagating the data source to the data source's scope"
	case targetAssetType_Secret:
		statusInfo.Message = "No issues were found while propagating the credential to the credentials' scope"
	case targetAssetType_Credential:
		if assetKind == api.ASSETKIND_ENVIRONMENT {
			statusInfo.Message = "No issues were found while propagating the environment's credentials to the credentials' scope"
			return
		}
		statusInfo.Message = "No issues were found while propagating the data source's credentials to the credentials' scope"
	}
}

// getAssetClusterResourceName returns the resource name of the asset in the cluster
// if the asset is a datasource that uses a credential, it returns the credential id
func getAssetClusterResourceName(ctx context.Context, patchingInfo response_patching.ResponsePatchingPtrs) []string {
	switch patchingInfo.Meta.Kind {
	case api.ASSETKIND_PVC:
		pvcResourceName := getPVCResourceName(patchingInfo)
		if pvcResourceName == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("no resource name found for pvc %s", patchingInfo.Meta.Name)
			return []string{}
		}
		return []string{*pvcResourceName}
	case api.ASSETKIND_CONFIG_MAP:
		return []string{*getConfigMapResourceName(patchingInfo)}
	case api.ASSETKIND_ACCESS_KEY:
		return []string{*getAccessKeyResourceName(patchingInfo)}
	case api.ASSETKIND_PASSWORD:
		return []string{*getPasswordResourceName(patchingInfo)}
	case api.ASSETKIND_GENERIC_SECRET:
		return []string{*getGenericSecretResourceName(patchingInfo)}
	case api.ASSETKIND_DOCKER_REGISTRY:
		return []string{*getDockerRegistryResourceName(patchingInfo)}
	case api.ASSETKIND_NGC_API_KEY:
		return getNGCApiKeyResourcesName(patchingInfo)
	}
	return nil
}

// getCredentialIdUsedByAsset returns a list of all credentials used an asset
func getCredentialIdUsedByAsset(patchingInfo response_patching.ResponsePatchingPtrs) []string {
	switch patchingInfo.Meta.Kind {
	case api.ASSETKIND_S3, api.ASSETKIND_GIT, api.ASSETKIND_SECRET_VOLUME:
		credentialId := getDatasourceCredentialId(patchingInfo)
		if credentialId == nil {
			return nil
		}
		return []string{*credentialId}
	case api.ASSETKIND_ENVIRONMENT:
		return getCredentialIdsFromEnvironment(patchingInfo)
	}
	return nil
}

func getPVCResourceName(patchingInfo response_patching.ResponsePatchingPtrs) *string {
	switch assetSpec := patchingInfo.Spec.(type) {
	case *api.PvcAssetSpec:
		return assetSpec.ClaimName
	case *api.DatasourceListResponseAssetSpec:
		return assetSpec.Pvc.ClaimName
	}
	return nil
}

func getDatasourceCredentialId(patchingInfo response_patching.ResponsePatchingPtrs) *string {
	switch assetSpec := patchingInfo.Spec.(type) {
	case *api.S3AssetSpec:
		if assetSpec.AccessKeyAssetId != nil {
			return assetSpec.AccessKeyAssetId
		}
	case *api.GitAssetSpec:
		return dbmodel.GetGitCredentialId(*assetSpec)
	case *api.SecretAssetSpec:
		if assetSpec.CredentialAssetId != nil {
			return assetSpec.CredentialAssetId
		}
	case *api.DatasourceListResponseAssetSpec:
		if assetSpec.Git != nil {
			return dbmodel.GetGitCredentialId(*assetSpec.Git)
		}
		if assetSpec.S3 != nil && assetSpec.S3.AccessKeyAssetId != nil {
			return assetSpec.S3.AccessKeyAssetId
		}
		if assetSpec.Secret != nil && assetSpec.Secret.CredentialAssetId != nil {
			return assetSpec.Secret.CredentialAssetId
		}
	}
	return nil
}

func getCredentialIdsFromEnvironment(patchingInfo response_patching.ResponsePatchingPtrs) []string {
	var credentialIds []string
	switch assetSpec := patchingInfo.Spec.(type) {
	case *api.EnvironmentAssetSpec:
		if assetSpec.EnvironmentVariables == nil {
			return nil
		}
		for _, env := range *assetSpec.EnvironmentVariables {
			if env.Credential == nil {
				continue
			}
			credentialIds = append(credentialIds, *env.Credential.AssetId)
		}
	}
	return credentialIds
}

func getConfigMapResourceName(patchingInfo response_patching.ResponsePatchingPtrs) *string {
	switch assetSpec := patchingInfo.Spec.(type) {
	case *api.ConfigMapAssetSpec:
		return assetSpec.ConfigMap
	case *api.DatasourceListResponseAssetSpec:
		return assetSpec.ConfigMap.ConfigMap
	}
	return nil
}

func getAccessKeyResourceName(patchingInfo response_patching.ResponsePatchingPtrs) *string {
	switch assetSpec := patchingInfo.Spec.(type) {
	case *api.AccessKeySpec:
		if assetSpec.ExistingSecretName != nil {
			return assetSpec.ExistingSecretName
		} else {
			secretName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_ACCESS_KEY], patchingInfo.Meta.Name)
			return &secretName
		}
	case *api.CredentialsListResponseAssetSpec:
		if assetSpec.AccessKey != nil {
			if assetSpec.AccessKey.ExistingSecretName != nil {
				return assetSpec.AccessKey.ExistingSecretName
			} else {
				secretName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_ACCESS_KEY], patchingInfo.Meta.Name)
				return &secretName
			}
		}
	}
	return nil
}

func getPasswordResourceName(patchingInfo response_patching.ResponsePatchingPtrs) *string {
	switch assetSpec := patchingInfo.Spec.(type) {
	case *api.PasswordSpec:
		if assetSpec.ExistingSecretName != nil {
			return assetSpec.ExistingSecretName
		} else {
			secretName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_PASSWORD], patchingInfo.Meta.Name)
			return &secretName
		}
	case *api.CredentialsListResponseAssetSpec:
		if assetSpec.Password != nil {
			if assetSpec.Password.ExistingSecretName != nil {
				return assetSpec.Password.ExistingSecretName
			} else {
				secretName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_PASSWORD], patchingInfo.Meta.Name)
				return &secretName
			}
		}
	}
	return nil
}

func getGenericSecretResourceName(patchingInfo response_patching.ResponsePatchingPtrs) *string {
	switch assetSpec := patchingInfo.Spec.(type) {
	case *api.GenericSecretSpec:
		if assetSpec.ExistingSecretName != nil {
			return assetSpec.ExistingSecretName
		} else {
			if convutils.BoolIsTrue(patchingInfo.Meta.Private) {
				secretName := fmt.Sprintf("%s-%s", patchingInfo.Meta.Name, patchingInfo.Meta.Id[:10])
				return &secretName
			}
			secretName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_GENERIC_SECRET], patchingInfo.Meta.Name)
			return &secretName
		}
	case *api.CredentialsListResponseAssetSpec:
		if assetSpec.GenericSecret != nil {
			if assetSpec.GenericSecret.ExistingSecretName != nil {
				return assetSpec.GenericSecret.ExistingSecretName
			} else {
				if convutils.BoolIsTrue(patchingInfo.Meta.Private) {
					secretName := fmt.Sprintf("%s-%s", patchingInfo.Meta.Name, patchingInfo.Meta.Id[:10])
					return &secretName
				}
				secretName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_GENERIC_SECRET], patchingInfo.Meta.Name)
				return &secretName
			}
		}
	}
	return nil
}

func getDockerRegistryResourceName(patchingInfo response_patching.ResponsePatchingPtrs) *string {
	switch assetSpec := patchingInfo.Spec.(type) {
	case *api.DockerRegistrySpec:
		if assetSpec.ExistingSecretName != nil {
			return assetSpec.ExistingSecretName
		} else {
			if convutils.BoolIsTrue(patchingInfo.Meta.Private) {
				secretName := fmt.Sprintf("%s-%s", patchingInfo.Meta.Name, patchingInfo.Meta.Id[:10])
				return &secretName
			}
			secretName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_DOCKER_REGISTRY], patchingInfo.Meta.Name)
			return &secretName
		}
	case *api.CredentialsListResponseAssetSpec:
		if assetSpec.DockerRegistry != nil {
			if assetSpec.DockerRegistry.ExistingSecretName != nil {
				return assetSpec.DockerRegistry.ExistingSecretName
			} else {
				if convutils.BoolIsTrue(patchingInfo.Meta.Private) {
					secretName := fmt.Sprintf("%s-%s", patchingInfo.Meta.Name, patchingInfo.Meta.Id[:10])
					return &secretName
				}
				secretName := fmt.Sprintf("%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_DOCKER_REGISTRY], patchingInfo.Meta.Name)
				return &secretName
			}
		}
	}
	return nil
}

func getNGCApiKeyResourcesName(patchingInfo response_patching.ResponsePatchingPtrs) []string {
	var baseName string
	switch patchingInfo.Spec.(type) {
	case *api.NgcApiKeySpec:
		baseName = patchingInfo.Meta.Name
	case *api.CredentialsListResponseAssetSpec:
		baseName = patchingInfo.Meta.Name
	default:
		return nil
	}
	if convutils.BoolIsTrue(patchingInfo.Meta.Private) {
		return []string{
			fmt.Sprintf("%s-%s-%s", composer.NgcApiKeyGenericSecretPrefix, baseName, patchingInfo.Meta.Id[:10]),
			fmt.Sprintf("%s-%s-%s", composer.NgcApiKeyDockerRegistryPrefix, baseName, patchingInfo.Meta.Id[:10]),
		}
	}
	return []string{
		fmt.Sprintf("%s-%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_GENERIC_SECRET], composer.NgcApiKeyGenericSecretPrefix, baseName),
		fmt.Sprintf("%s-%s-%s", dbmodel.SecretTypeMap[api.ASSETKIND_DOCKER_REGISTRY], composer.NgcApiKeyDockerRegistryPrefix, baseName),
	}
}

func getScopeInfoForStatus(meta *api.AssetMeta) *scope.ScopeInfo {
	scopeInfo := scope.ScopeInfoOf(scope.Scope(meta.Scope), meta.ProjectId, meta.DepartmentId, meta.ClusterId, meta.TenantId)
	if convutils.BoolIsTrue(meta.Private) {
		scopeInfo.ScopeId = fmt.Sprintf("%d", *meta.TenantId)
	}
	return scopeInfo
}
