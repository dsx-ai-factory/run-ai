/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"context"
	"fmt"
	"strings"

	assets "github.com/run-ai/runai/runai-clients/assets-service-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type credentialMetadata struct {
	CredentialID   string  `json:"credentialId"           jsonschema:"Credential asset ID to pass to manage_credentials."`
	Name           string  `json:"name"                   jsonschema:"Credential name."`
	CredentialType string  `json:"credentialType"         jsonschema:"Credential kind."`
	Private        bool    `json:"private"                jsonschema:"True for a user credential; false for a shared credential."`
	Description    *string `json:"description,omitempty"  jsonschema:"Credential description."`
	Scope          string  `json:"scope"                  jsonschema:"Credential scope."`
	ClusterID      string  `json:"clusterId,omitempty"    jsonschema:"Owning cluster ID, when scoped below tenant."`
	DepartmentID   string  `json:"departmentId,omitempty" jsonschema:"Owning department ID, for department-scoped credentials."`
	ProjectID      *int32  `json:"projectId,omitempty"    jsonschema:"Owning project ID, for project-scoped credentials."`
}

type listCredentialsInput struct {
	CredentialID   string `json:"credentialId,omitempty"   jsonschema:"Filter by exact credential asset ID (UUID)."`
	Name           string `json:"name,omitempty"           jsonschema:"Filter by exact credential name."`
	CredentialType string `json:"credentialType,omitempty" jsonschema:"Filter by credential kind. One of: accessKey, password, dockerRegistry, genericSecret, ngcApiKey."`
	Private        *bool  `json:"private,omitempty"        jsonschema:"Filter by privacy. true returns user credentials; false returns shared credentials; omit to return both."`
	Scope          string `json:"scope,omitempty"          jsonschema:"Filter shared credentials by exact scope. One of: tenant, cluster, department, project."`
	ClusterID      string `json:"clusterId,omitempty"      jsonschema:"Filter by cluster ID. With scope=cluster, returns that cluster's credentials; otherwise inherited credentials may also be returned."`
	DepartmentID   string `json:"departmentId,omitempty"   jsonschema:"Filter by department ID. With scope=department, returns that department's credentials; otherwise inherited credentials may also be returned."`
	ProjectID      *int32 `json:"projectId,omitempty"      jsonschema:"Filter by project ID. With scope=project, returns that project's credentials; otherwise inherited credentials may also be returned."`
	Limit          int32  `json:"limit,omitempty"          jsonschema:"Maximum number of credentials to return. Defaults to 50, capped at 200."`
	Offset         int32  `json:"offset,omitempty"         jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listCredentialsInput) Validate() error {
	if in.CredentialID != "" {
		if err := common.ValidateUUID("credentialId", in.CredentialID); err != nil {
			return err
		}
	}
	if in.CredentialType != "" {
		if _, err := credentialAdapterByType(in.CredentialType); err != nil {
			return err
		}
	}
	switch assets.Scope(in.Scope) {
	case "", assets.SCOPE_TENANT, assets.SCOPE_CLUSTER, assets.SCOPE_DEPARTMENT, assets.SCOPE_PROJECT:
	default:
		return fmt.Errorf("scope must be one of: tenant, cluster, department, project")
	}
	if in.Private != nil && *in.Private &&
		(in.Scope != "" || in.ClusterID != "" || in.DepartmentID != "" || in.ProjectID != nil) {
		return fmt.Errorf("scope filters are not valid with private=true")
	}
	if in.ClusterID != "" {
		if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
			return err
		}
	}
	return common.ValidatePagination(in.Limit, in.Offset)
}

type listCredentialsOutput struct {
	Credentials []credentialMetadata `json:"credentials"          jsonschema:"Matching credential metadata. Secret values are never returned."`
	TotalCount  int32                `json:"totalCount"           jsonschema:"Total number of credentials matching the filters across all pages."`
	HasMore     bool                 `json:"hasMore"              jsonschema:"True when more credentials exist beyond the returned page; pass nextOffset back as offset."`
	NextOffset  int32                `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call to fetch more credentials. Only set when hasMore is true."`
}

func registerListCredentials(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listCredentialsInput, listCredentialsOutput]{
		Name: "list_credentials",
		Description: "List safe metadata for shared credentials and the caller's user credentials. " +
			"Call this before manage_credentials when an update or delete target's credentialId is unknown. " +
			"Supports exact ID and name filters, API-native kind, privacy, and scope filters, plus limit/offset pagination. " +
			"Call without filters to list all visible credentials. Never returns credential values.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listCredentialsInput) (listCredentialsOutput, error) {
			return listCredentialsRun(ctx, client, in)
		},
	})
}

func listCredentialsRun(
	ctx context.Context,
	client *runaiclient.Client,
	in listCredentialsInput,
) (listCredentialsOutput, error) {
	req := client.Assets.CredentialsAPI.ListCredentialsAssets(ctx)
	if in.Name != "" {
		req = req.Name(in.Name)
	}
	if in.CredentialType != "" {
		req = req.Type_(in.CredentialType)
	}
	req = applyPrivateFilter(req, in.Private)
	if in.Scope != "" {
		req = req.Scope(in.Scope)
	}
	if in.ClusterID != "" {
		req = req.ClusterId(in.ClusterID)
	}
	if in.DepartmentID != "" {
		req = req.DepartmentId(in.DepartmentID)
	}
	if in.ProjectID != nil {
		req = req.ProjectId(*in.ProjectID)
	}

	result, resp, err := req.Execute()
	if err != nil {
		return listCredentialsOutput{}, fmt.Errorf(
			"list credentials: %w", runaimcp.MapHTTPError(resp, err))
	}
	if result == nil {
		return listCredentialsOutput{}, fmt.Errorf(
			"list credentials: nil response: %w", runaimcp.ErrUpstream)
	}

	matches := []credentialMetadata{}
	for _, entry := range result.Entries {
		if in.CredentialID != "" && !strings.EqualFold(entry.Meta.Id, in.CredentialID) {
			continue
		}
		matches = append(matches, credentialMetadataFromAsset(entry.Meta))
	}
	// The API has no limit or offset parameters, so paginate its returned order.
	page, hasMore, nextOffset := common.Paginate(
		matches, in.Offset, common.EffectiveLimit(in.Limit))
	out := listCredentialsOutput{
		Credentials: page,
		TotalCount:  int32(len(matches)),
		HasMore:     hasMore,
		NextOffset:  nextOffset,
	}
	return out, nil
}

func applyPrivateFilter(
	req assets.ApiListCredentialsAssetsRequest,
	private *bool,
) assets.ApiListCredentialsAssetsRequest {
	if private == nil {
		return req.IncludePrivate(true)
	}
	if *private {
		return req.PrivateOnly(true)
	}
	return req.IncludePrivate(false)
}

func credentialMetadataFromAsset(meta assets.AssetMeta) credentialMetadata {
	out := credentialMetadata{
		CredentialID:   meta.Id,
		Name:           meta.Name,
		CredentialType: string(meta.Kind),
		Private:        meta.GetPrivate(),
		Scope:          string(meta.Scope),
		ClusterID:      meta.GetClusterId(),
		DepartmentID:   meta.GetDepartmentId(),
	}
	if meta.HasDescription() {
		description := meta.GetDescription()
		out.Description = &description
	}
	if meta.HasProjectId() {
		projectID := meta.GetProjectId()
		out.ProjectID = &projectID
	}
	return out
}

func credentialSnapshot(meta assets.AssetMeta) map[string]any {
	metadata := credentialMetadataFromAsset(meta)
	out := map[string]any{
		"name":           metadata.Name,
		"credentialType": metadata.CredentialType,
		"private":        metadata.Private,
		"scope":          metadata.Scope,
	}
	if metadata.CredentialID != "" {
		out["credentialId"] = metadata.CredentialID
	}
	if metadata.Description != nil {
		out["description"] = *metadata.Description
	}
	if metadata.ClusterID != "" {
		out["clusterId"] = metadata.ClusterID
	}
	if metadata.DepartmentID != "" {
		out["departmentId"] = metadata.DepartmentID
	}
	if metadata.ProjectID != nil {
		out["projectId"] = *metadata.ProjectID
	}
	return out
}
