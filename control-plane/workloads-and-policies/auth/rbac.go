package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"
	assets "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	orgUnitClient "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	WorkloadTypeWorkspace   = "Workspace"
	WorkloadTypeTraining    = "Training"
	WorkloadTypeDistributed = "Distributed"
	WorkloadTypeInference   = "Inference"

	contextKeyAncestors = "cachedAncestors"
)

// ContextWithAncestors allows a caller that already have the ancestors of a resource to
// pass it in the context, in order to avoid excessive call to the org unit service
func ContextWithAncestors(ctx context.Context, ancestors *orgUnitClient.Ancestors) context.Context {
	// ignore the error asking us to create type instead of string
	// nolint:staticcheck
	return context.WithValue(ctx, contextKeyAncestors, ancestors)
}

func AncestorsFromContext(ctx context.Context) *orgUnitClient.Ancestors {
	result, ok := ctx.Value(contextKeyAncestors).(*orgUnitClient.Ancestors)
	if ok {
		return result
	} else {
		return nil
	}
}

const (
	WorkspaceAsAssetKind = "workspace"
	TrainingAsAssetKind  = "training"
	InferenceAsAssetKind = "inference"
)

var WorkloadTypeToAssetKind = map[string]assets.AssetKind{
	WorkloadTypeWorkspace:   WorkspaceAsAssetKind,
	WorkloadTypeTraining:    TrainingAsAssetKind,
	WorkloadTypeDistributed: TrainingAsAssetKind,
	WorkloadTypeInference:   InferenceAsAssetKind,
}

// Mapping b/w asset kind/workload type and the resource type used by the authorization service
var resourceTypeOfKind = map[assets.AssetKind]authClient.ResourceType{
	assets.ASSETKIND_ACCESS_KEY:        authClient.RESOURCETYPE_CREDENTIALS,
	assets.ASSETKIND_GENERIC_SECRET:    authClient.RESOURCETYPE_CREDENTIALS,
	assets.ASSETKIND_COMPUTE:           authClient.RESOURCETYPE_COMPUTE_RESOURCES,
	assets.ASSETKIND_DOCKER_REGISTRY:   authClient.RESOURCETYPE_CREDENTIALS,
	assets.ASSETKIND_ENVIRONMENT:       authClient.RESOURCETYPE_ENVIRONMENTS,
	assets.ASSETKIND_HOST_PATH:         authClient.RESOURCETYPE_HOST_PATH_ASSETS,
	assets.ASSETKIND_GIT:               authClient.RESOURCETYPE_GIT_ASSETS,
	assets.ASSETKIND_NFS:               authClient.RESOURCETYPE_NFS_ASSETS,
	assets.ASSETKIND_PASSWORD:          authClient.RESOURCETYPE_CREDENTIALS,
	assets.ASSETKIND_PVC:               authClient.RESOURCETYPE_PVC_ASSETS,
	assets.ASSETKIND_REGISTRY:          authClient.RESOURCETYPE_CREDENTIALS,
	assets.ASSETKIND_S3:                authClient.RESOURCETYPE_S3_ASSETS,
	assets.ASSETKIND_WORKLOAD_TEMPLATE: authClient.RESOURCETYPE_TEMPLATES,
	assets.ASSETKIND_MODEL:             authClient.RESOURCETYPE_INFERENCES,
	assets.ASSETKIND_CONFIG_MAP:        authClient.RESOURCETYPE_CM_VOLUME_ASSETS,
	assets.ASSETKIND_SECRET_VOLUME:     authClient.RESOURCETYPE_SECRET_VOLUME_ASSETS,
	assets.ASSETKIND_NGC_API_KEY:       authClient.RESOURCETYPE_CREDENTIALS,
	WorkspaceAsAssetKind:               authClient.RESOURCETYPE_WORKSPACES,
	TrainingAsAssetKind:                authClient.RESOURCETYPE_TRAININGS,
	InferenceAsAssetKind:               authClient.RESOURCETYPE_INFERENCES,
}

var TestingAllowCreateSystemScopedAssets = false

func ResourceTypeOfKind(assetKind assets.AssetKind) (*authClient.ResourceType, *Error) {
	resourceType, ok := resourceTypeOfKind[assetKind]
	if !ok {
		return nil, &Error{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("unsupported assetKind '%s'", assetKind),
		}
	}
	return &resourceType, nil
}

// LoadAllowedOrgUnits finds out the complete set of projects and departments the caller is allowed to read,
// and add it to the context, because it is used by many functions while running assets APIs
func LoadAllowedOrgUnits(ctx context.Context) (*PermittedScopes, *Error) {
	rawScopes := authClient.PermittedScopes{}

	orgUnits, err := orgunit.Get().Client().GetOrgTree(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get org tree: %s", err.Message)
		return nil, &Error{
			Code:    http.StatusForbidden,
			Message: "failed to interrogate the organizational tree",
		}
	}

	for _, orgUnit := range orgUnits {
		if orgUnit.GetType() == orgUnitClient.SCOPETYPE_PROJECT {
			rawScopes.Projects = append(rawScopes.Projects, orgUnit.GetId())
		} else if orgUnit.GetType() == orgUnitClient.SCOPETYPE_DEPARTMENT {
			rawScopes.Departments = append(rawScopes.Departments, orgUnit.GetId())
		} else if orgUnit.GetType() == orgUnitClient.SCOPETYPE_CLUSTER {
			rawScopes.Clusters = append(rawScopes.Clusters, orgUnit.GetId())
		} else if orgUnit.GetType() == orgUnitClient.SCOPETYPE_TENANT {
			rawScopes.Tenants = append(rawScopes.Tenants, orgUnit.GetId())
		} else if orgUnit.GetType() == orgUnitClient.SCOPETYPE_SYSTEM {
			rawScopes.System = convutils.Ref(true)
		}
	}

	return NewPermittedScopes(&rawScopes), nil
}

// ValidateAccessByActionAndAssetKind check permission of the user to perform an action on a given resource in a given scope.
func ValidateAccessByActionAndAssetKind(ctx context.Context, scopeInfo *scope.ScopeInfo, assetKind string, action authClient.Action, isPrivate bool) *Error {
	resourceType, err := ResourceTypeOfKind(assets.AssetKind(assetKind))
	if err != nil {
		return err
	}
	return ValidateAccessByResourceTypeAndAction(ctx, scopeInfo, *resourceType, action, isPrivate)
}

func ValidateAccessByResourceTypeAndAction(ctx context.Context, scopeInfo *scope.ScopeInfo, resourceType authClient.ResourceType, action authClient.Action, isPrivate bool) *Error {
	rawPermittedScopes, err := authorization.Get().Client().CheckPermittedScopes(ctx, resourceType, action)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("check permitted scopes failed: %s: %v", err.Message, err.Details)
		return &ForbiddenError
	}

	// system scoped assets can be read by anyone and created/updated/deleted by no one
	if scopeInfo.Scope == scope.System {
		if action == authClient.ACTION_READ || TestingAllowCreateSystemScopedAssets {
			return nil
		} else {
			return &ForbiddenError
		}
	}

	// access to private assets is allowed if there request is allowed to perform an action at least on a scope.
	if isPrivate {
		if len(rawPermittedScopes.Projects) > 0 || len(rawPermittedScopes.Departments) > 0 || len(rawPermittedScopes.Clusters) > 0 || len(rawPermittedScopes.Tenants) > 0 {
			return nil
		}
		return &ForbiddenError
	}

	permittedScopes := NewPermittedScopes(rawPermittedScopes)

	//
	//   round 1: "naive" validation. department scope VS allowed departments, project scope VS allowed projects
	//
	switch scopeInfo.Scope {
	case scope.Tenant:
		// For tenant-scoped assets, ScopeId should contain the asset's tenant ID.
		// This ensures we validate against the ASSET's tenant, not the requester's tenant.
		// If ScopeId is empty (e.g., CREATE where asset doesn't exist yet), fall back to context tenant.
		tenantId := scopeInfo.ScopeId
		if tenantId == "" {
			tenantId = fmt.Sprintf("%d", TenantIdOf(ctx))
		}
		if isAllowedTenant(rawPermittedScopes, tenantId) {
			return nil
		}
		return &ForbiddenError

	case scope.Cluster:
		if permittedScopes.isAllowedCluster(scopeInfo.ScopeId) {
			return nil
		}

	case scope.Department:
		if permittedScopes.isAllowedDepartment(scopeInfo.ScopeId) {
			return nil
		}

	case scope.Project:
		if permittedScopes.isAllowedProject(scopeInfo.ScopeId) {
			return nil
		}
	}

	//
	//   round 2: we did not have a direct match, so we need to check the ancestors of the asset to see
	//   maybe the permission can be granted on the basis of the ancestors.
	//
	ancestors := AncestorsFromContext(ctx)
	if ancestors == nil {
		var getErr *orgUnitClient.Error
		ancestors, getErr = orgunit.Get().Client().GetAncestors(ctx, string(scopeInfo.Scope), scopeInfo.ScopeId)
		if getErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to obtain ancestors of %s/%s: %s", scopeInfo.Scope, scopeInfo, getErr.Message)
			if len(rawPermittedScopes.Clusters) == 0 && len(rawPermittedScopes.Departments) == 0 && len(rawPermittedScopes.Projects) == 0 {
				if isAllowedTenant(rawPermittedScopes, fmt.Sprintf("%d", TenantIdOf(ctx))) {
					return nil
				}
			}
			return &ForbiddenError
		}
	}

	// Check if user has tenant-level permission for the asset's tenant (fallback for round 2)
	if isAllowedTenant(rawPermittedScopes, ancestors.TenantId) {
		return nil
	}

	switch scopeInfo.Scope {
	case scope.Department:
		if permittedScopes.isAllowedCluster(convutils.Deref(ancestors.ClusterUuid.Get(), "")) {
			return nil
		}
	case scope.Project:
		if permittedScopes.isAllowedDepartment(convutils.Deref(ancestors.DepartmentId.Get(), "")) ||
			permittedScopes.isAllowedCluster(convutils.Deref(ancestors.ClusterUuid.Get(), "")) {
			return nil
		}
	}

	return &ForbiddenError
}

// isAllowedTenant checks if the user has permission for the specific tenant
func isAllowedTenant(permittedScopes *authClient.PermittedScopes, tenantId string) bool {
	for _, t := range permittedScopes.Tenants {
		if t == tenantId {
			return true
		}
	}
	return false
}
