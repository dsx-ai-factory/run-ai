/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package query_types

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

const assetIdsPattern = "^\\b(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})(?:,(?:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}))*\\b$"

var AssetIdsRegexp = regexp.MustCompile(assetIdsPattern)

type DistributedQueryParams struct {
	DistributedFramework   string
	IsDistributed          bool
	IsDistributedInference bool
	IsTraining             bool
	IsInference            bool
	IsWorkspace            bool
}

type AllowedOrgUnits struct {
	AllowedDepartments []string
	AllowedProjects    []string
	AllowedClusters    []string
}

func NewAllowedOrgUnits() *AllowedOrgUnits {
	return &AllowedOrgUnits{
		AllowedDepartments: make([]string, 0),
		AllowedProjects:    make([]string, 0),
		AllowedClusters:    make([]string, 0),
	}
}

type QueryParams struct {
	// filter by asset ids)
	AssetIds *[]string

	// filter by name
	Name *string

	// filter by scope
	Scope *api.Scope

	// include only assets that are accessible by a given cluster
	ClusterId *string

	// include only assets that are accessible by a given project
	//    if scope is given - only assets of this scope
	//    if scope not given - include assets from all levels (tenant, department and project)
	ProjectId *int32

	DepartmentId *string

	EnvironmentId *string

	// if true, include assets with auto-deletion flag
	IncludeAutoDeleteAssets bool

	// those parameters is used to filter by workload type
	DistributedQueryParams *DistributedQueryParams

	// this filtering is a bit different, it is not taken from query params themselves but from
	// the authorization info.
	AllowedOrgUnits *AllowedOrgUnits

	// IncludeDescendants includes assets from all nested layers within the specified scope
	IncludeDescendants bool

	// PrivateOnly indicates that only private assets should be returned (mutual exclusion with IncludePrivate)
	PrivateOnly bool

	// IncludePrivate indicates that private assets should be included in the result (mutual exclusion with PrivateOnly)
	IncludePrivate bool
}

func (params *QueryParams) OnlyAssetsOfScope(scope api.Scope) bool {
	return params.Scope != nil && *params.Scope == scope
}

func NewQueryParams(
	assetIds string,
	name string,
	scope string,
	includeDescendants bool,
	projectId int32,
	departmentId string,
	clusterId string,
	distributedQueryParams *DistributedQueryParams,
) (*QueryParams, *api.Error) {
	result := QueryParams{
		IncludeDescendants: includeDescendants,
	}

	if assetIds != "" {
		if !AssetIdsRegexp.MatchString(assetIds) {
			return nil, api.NewError(http.StatusBadRequest, "assetIds parameters must contain comma separated list of UUIDs", nil)
		}

		result.AssetIds = convutils.Ref(strings.Split(assetIds, ","))
	}

	if name != "" {
		result.Name = convutils.Ref(name)
	}

	var err *api.Error
	result.Scope, err = loadScopeParameter(scope)
	if err != nil {
		return nil, err
	}
	if result.Scope != nil && includeDescendants {
		return nil, api.NewError(http.StatusBadRequest, "The parameters includeDescendants and scope are mutual exclusive", nil)
	}

	if projectId != 0 {
		result.ProjectId = convutils.Ref(projectId)
	}
	if departmentId != "" {
		// TODO: add department auth check
		result.DepartmentId = convutils.Ref(departmentId)
	}
	if clusterId != "" {
		result.ClusterId = convutils.Ref(clusterId)
	}

	if distributedQueryParams != nil {
		if len(distributedQueryParams.DistributedFramework) > 0 {
			if !distributedQueryParams.IsDistributed {
				return nil, api.NewError(http.StatusBadRequest, "Invalid query parameters - cannot use distributedFramework without isDistributed set to true", nil)
			}
			distFramework := api.DistributedFramework(distributedQueryParams.DistributedFramework)
			if !distFramework.IsValid() {
				return nil, api.NewError(http.StatusBadRequest, "Invalid query parameters (distributedFramework)", nil)
			}
		}
		result.DistributedQueryParams = distributedQueryParams
	}

	return &result, nil
}

// WithPrivate sets the PrivateOnly and IncludePrivate parameters.
func (q *QueryParams) WithPrivate(privateOnly bool, includePrivate bool) (*QueryParams, *api.Error) {
	if privateOnly && includePrivate {
		return nil, api.NewError(http.StatusBadRequest, "The parameters privateOnly and includePrivate are mutual exclusive", nil)
	}
	q.PrivateOnly = privateOnly
	q.IncludePrivate = includePrivate
	return q, nil
}

func loadScopeParameter(scope string) (*api.Scope, *api.Error) {
	switch scope {
	case string(api.SCOPE_SYSTEM),
		string(api.SCOPE_TENANT),
		string(api.SCOPE_CLUSTER),
		string(api.SCOPE_DEPARTMENT),
		string(api.SCOPE_PROJECT):
		desiredScope := api.Scope(scope)
		return &desiredScope, nil
	case "":
		return nil, nil
	default:
		return nil, api.NewValidationError("invalid value provided for scope parameter")
	}
}

func LoadKindParameter(kind string) (*api.AssetKind, *api.Error) {
	switch kind {
	case string(api.ASSETKIND_ACCESS_KEY),
		string(api.ASSETKIND_COMPUTE),
		string(api.ASSETKIND_ENVIRONMENT),
		string(api.ASSETKIND_GIT),
		string(api.ASSETKIND_HOST_PATH),
		string(api.ASSETKIND_NFS),
		string(api.ASSETKIND_PASSWORD),
		string(api.ASSETKIND_REGISTRY),
		string(api.ASSETKIND_GENERIC_SECRET),
		string(api.ASSETKIND_PVC),
		string(api.ASSETKIND_CONFIG_MAP),
		string(api.ASSETKIND_SECRET_VOLUME),
		string(api.ASSETKIND_S3),
		string(api.ASSETKIND_WORKLOAD_TEMPLATE),
		string(api.ASSETKIND_DOCKER_REGISTRY),
		string(api.ASSETKIND_NGC_API_KEY):
		result := api.AssetKind(kind)
		return &result, nil
	case "":
		return nil, nil
	default:
		return nil, api.NewValidationError("invalid value provided for kind parameter")
	}
}
