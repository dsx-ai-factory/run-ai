/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

const (
	kindAccessKey      = "accessKey"
	kindPassword       = "password"
	kindDockerRegistry = "dockerRegistry"
	kindGenericSecret  = "genericSecret"
	kindNGCAPIKey      = "ngcApiKey"
)

type manageCredentialsInput struct {
	Action         string          `json:"action"                    jsonschema:"What to do. One of: create, update, delete."`
	Private        bool            `json:"private,omitempty"         jsonschema:"Create-only. Set true for a user credential visible only to its creator. Defaults to false for a shared credential."`
	CredentialType string          `json:"credentialType"            jsonschema:"Credential kind. Shared: accessKey, password, dockerRegistry, genericSecret, ngcApiKey. User: dockerRegistry, genericSecret, ngcApiKey."`
	CredentialID   string          `json:"credentialId,omitempty"    jsonschema:"Credential asset ID (UUID). Required for update and delete; use list_credentials to resolve it when unknown. Must not be set for create."`
	Name           string          `json:"name,omitempty"            jsonschema:"Credential name. Required for create; must not be set for update or delete."`
	Description    *string         `json:"description,omitempty"     jsonschema:"Description. Required for update; an empty string clears it. Updating secret values is not supported upstream."`
	Scope          string          `json:"scope,omitempty"           jsonschema:"Shared create scope: tenant, cluster, department, or project. genericSecret and ngcApiKey exclude tenant. User credentials use tenant."`
	ClusterID      string          `json:"clusterId,omitempty"       jsonschema:"Cluster ID (UUID). Required with scope=cluster; optional with scope=project. Do not set with scope=department; derived from departmentId."`
	DepartmentID   string          `json:"departmentId,omitempty"    jsonschema:"Department ID. Required for department-scoped shared credentials."`
	ProjectID      *int32          `json:"projectId,omitempty"       jsonschema:"Project ID. Required for project-scoped shared credentials."`
	Spec           *credentialSpec `json:"spec,omitempty"            jsonschema:"Secret material for create. Never accepted for update/delete because upstream credential updates support description metadata only."`
	IgnoreMissing  bool            `json:"ignoreMissing,omitempty"   jsonschema:"Delete-only. Treat an already-missing credential as a successful no-op."`
	Confirm        bool            `json:"confirm,omitempty"         jsonschema:"Required for update and delete when the client cannot show an interactive confirmation prompt. Create does not require confirmation."`
}

type manageCredentialsOutput struct {
	Action         string         `json:"action"                    jsonschema:"The action performed: create, update, or delete."`
	Applied        bool           `json:"applied"                   jsonschema:"True when the upstream service accepted the change."`
	CredentialID   string         `json:"credentialId,omitempty"    jsonschema:"Created or resolved credential asset ID."`
	Name           string         `json:"name,omitempty"            jsonschema:"Credential name."`
	CredentialType string         `json:"credentialType"            jsonschema:"Credential kind."`
	Private        bool           `json:"private"                   jsonschema:"True for a user credential; false for a shared credential."`
	Before         map[string]any `json:"before"                    jsonschema:"Safe metadata snapshot before the change. Secret values are never included."`
	After          map[string]any `json:"after"                     jsonschema:"Safe metadata snapshot after the change. Secret values are never included."`
	Diff           map[string]any `json:"diff"                      jsonschema:"Safe metadata changes. Secret material is omitted."`
	Warnings       []string       `json:"warnings,omitempty"        jsonschema:"Non-fatal notes, including ignored missing deletes."`
}

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerListCredentials(r, client)
	registerManageCredentials(r, client)
}

func registerManageCredentials(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageCredentialsInput, manageCredentialsOutput]{
		Name: "manage_credentials",
		Description: "Create, update, or delete shared and user credentials. " +
			"Shared credentials support accessKey, password, dockerRegistry, genericSecret, and ngcApiKey at cluster/department/project scope; " +
			"accessKey, password, and dockerRegistry also support tenant scope. " +
			"User credentials support dockerRegistry, genericSecret, and ngcApiKey and are visible only to their creator. " +
			"Create defaults to a shared credential; set private=true to create a user credential. " +
			"Create accepts secret material but never returns it and does not require confirmation. " +
			"Update changes description metadata only because the upstream API does not support rotating credential values. " +
			"Delete permanently removes the credential and can fail while it is in use. Update and delete require confirmation. " +
			"For update or delete, credentialId uniquely identifies the target and private is derived from the credential. " +
			"Call list_credentials first when credentialId is unknown. Do not use this tool to list or reveal credentials.",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageCredentialsInput) (manageCredentialsOutput, error) {
			return manageCredentialsRun(ctx, client, in)
		},
	})
}

func manageCredentialsRun(ctx context.Context, client *runaiclient.Client, in manageCredentialsInput) (manageCredentialsOutput, error) {
	action := common.WriteAction(in.Action)
	out := newManageCredentialsOutput(in)

	if action == common.ActionCreate {
		return createCredential(ctx, client, in, out)
	}

	// Update and delete need the current metadata for target validation,
	// confirmation details, and the before/after result.
	current, err := getCredentialMeta(ctx, client, in.CredentialType, in.CredentialID)
	if err != nil {
		if action == common.ActionDelete && in.IgnoreMissing && common.IsNotFound(err) {
			out.Warnings = append(out.Warnings, ignoredMissingCredentialWarning(in))
			return out, nil
		}
		return manageCredentialsOutput{}, err
	}

	out.CredentialID = current.Id
	out.Name = current.Name
	out.Private = current.GetPrivate()
	out.Before = credentialSnapshot(current)

	switch action {
	case common.ActionUpdate:
		return updateCredential(ctx, client, in, out, current)
	case common.ActionDelete:
		return deleteCredential(ctx, client, in, out, current)
	default:
		return manageCredentialsOutput{}, fmt.Errorf("unsupported action %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

func ignoredMissingCredentialWarning(in manageCredentialsInput) string {
	return fmt.Sprintf(
		"%s credential %q not found; nothing to delete",
		in.CredentialType,
		in.CredentialID,
	)
}

func newManageCredentialsOutput(in manageCredentialsInput) manageCredentialsOutput {
	return manageCredentialsOutput{
		Action:         in.Action,
		CredentialID:   in.CredentialID,
		Name:           in.Name,
		CredentialType: in.CredentialType,
		Private:        in.Private,
		Before:         map[string]any{},
		After:          map[string]any{},
		Diff:           map[string]any{},
	}
}
