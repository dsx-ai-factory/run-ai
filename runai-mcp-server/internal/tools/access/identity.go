/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package access

import (
	"context"
	"fmt"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	identity "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

func identityResourceType(e entity) authz.ResourceType {
	if e == entityServiceAccount {
		return authz.RESOURCETYPE_SERVICE_ACCOUNT
	}
	return authz.RESOURCETYPE_USERS
}

// authorizeIdentity fails fast when the caller lacks the tenant-level permission for the
// action, turning an eventual upstream 403 into an early, well-categorized forbidden error.
func authorizeIdentity(ctx context.Context, client *runaiclient.Client, e entity, action common.WriteAction) error {
	scope, err := common.PermittedScopeForAction(ctx, client, identityResourceType(e), action.Authz())
	if err != nil {
		return err
	}
	if !common.ScopeAllowsTenant(scope) {
		return fmt.Errorf("not permitted to %s %s in this tenant: %w", action, e, runaimcp.ErrForbidden)
	}
	return nil
}

func manageIdentity(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput, out manageUserAccessOutput) (manageUserAccessOutput, error) {
	if common.WriteAction(in.Action) == common.ActionCreate {
		return identityCreate(ctx, client, in, out)
	}
	return identityDelete(ctx, client, in, out)
}

func identityCreate(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput, out manageUserAccessOutput) (manageUserAccessOutput, error) {
	if err := authorizeIdentity(ctx, client, entity(in.Entity), common.ActionCreate); err != nil {
		return manageUserAccessOutput{}, err
	}

	if in.DryRun {
		out.After = createIdentityPreview(in)
		out.Diff = common.DiffMaps(nil, out.After)
		return out, nil
	}

	snap, id, err := doIdentityCreate(ctx, client, in)
	if err != nil {
		return manageUserAccessOutput{}, err
	}
	out.ID = id
	out.After = snap
	out.Applied = true
	out.Diff = common.DiffMaps(nil, out.After)
	out.Warnings = append(out.Warnings, credentialWarning(entity(in.Entity)))
	return out, nil
}

func doIdentityCreate(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput) (map[string]any, string, error) {
	if entity(in.Entity) == entityUser {
		req := identity.UserCreationRequest{Email: in.Email}
		if in.FirstName != "" {
			req.FirstName = &in.FirstName
		}
		if in.LastName != "" {
			req.LastName = &in.LastName
		}
		created, httpResp, err := client.Identity.UsersAPI.CreateUser(ctx).UserCreationRequest(req).Execute()
		if err != nil {
			return nil, "", fmt.Errorf("create user %q: %w", in.Email, runaimcp.MapHTTPError(httpResp, err))
		}
		if created == nil {
			return nil, "", fmt.Errorf("create user %q: nil response: %w", in.Email, runaimcp.ErrUpstream)
		}
		return map[string]any{"id": created.Id, "username": created.Username, "tempPassword": created.TempPassword}, created.Id, nil
	}
	created, httpResp, err := client.Identity.ServiceAccountsAPI.CreateServiceAccount(ctx).
		ServiceAccountCreationRequest(identity.ServiceAccountCreationRequest{Name: in.Name}).Execute()
	if err != nil {
		return nil, "", fmt.Errorf("create service-account %q: %w", in.Name, runaimcp.MapHTTPError(httpResp, err))
	}
	if created == nil {
		return nil, "", fmt.Errorf("create service-account %q: nil response: %w", in.Name, runaimcp.ErrUpstream)
	}
	return map[string]any{"id": created.Id, "name": created.Name, "clientId": strValue(created.ClientId), "clientSecret": created.ClientSecret}, created.Id, nil
}

func identityDelete(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput, out manageUserAccessOutput) (manageUserAccessOutput, error) {
	// Authorize before resolving — identity delete is tenant-scoped (no target needed), and this
	// stops an unpermitted caller from probing existence (incl. via ignoreMissing).
	if err := authorizeIdentity(ctx, client, entity(in.Entity), common.ActionDelete); err != nil {
		return manageUserAccessOutput{}, err
	}

	id, snap, found, err := resolveIdentity(ctx, client, in)
	if err != nil {
		return manageUserAccessOutput{}, err
	}
	if !found {
		if in.IgnoreMissing {
			out.Warnings = append(out.Warnings, fmt.Sprintf("%s not found; nothing to delete", in.Entity))
			return out, nil
		}
		return manageUserAccessOutput{}, fmt.Errorf("%s not found: %w", in.Entity, runaimcp.ErrNotFound)
	}

	out.ID = id
	out.Before = snap
	out.Diff = common.DiffMaps(out.Before, nil)

	if in.DryRun {
		return out, nil
	}
	if err := common.ConfirmWrite(ctx, in.Confirm, "delete", in.Entity, identityLabel(in), id, out.Before); err != nil {
		return manageUserAccessOutput{}, err
	}

	if err := doIdentityDelete(ctx, client, entity(in.Entity), id); err != nil {
		return manageUserAccessOutput{}, err
	}
	out.Applied = true
	return out, nil
}

// resolveIdentity finds the delete target by id when given, otherwise by email (user) or
// name (service-account), returning its snapshot. found=false means no match.
func resolveIdentity(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput) (id string, snap map[string]any, found bool, err error) {
	if entity(in.Entity) == entityUser {
		if in.ID != "" {
			u, httpResp, gErr := client.Identity.UsersAPI.GetUserById(ctx, in.ID).Execute()
			if gErr != nil {
				if common.IsNotFound(runaimcp.MapHTTPError(httpResp, gErr)) {
					return "", nil, false, nil
				}
				return "", nil, false, fmt.Errorf("get user %q: %w", in.ID, runaimcp.MapHTTPError(httpResp, gErr))
			}
			if u == nil {
				return "", nil, false, nil
			}
			return u.Id, userSnapshot(*u), true, nil
		}
		u, gErr := findUserByEmail(ctx, client, in.Email)
		if gErr != nil || u == nil {
			return "", nil, false, gErr
		}
		return u.Id, userSnapshot(*u), true, nil
	}
	sa, gErr := findServiceAccount(ctx, client, in.ID, in.Name)
	if gErr != nil || sa == nil {
		return "", nil, false, gErr
	}
	return sa.Id, map[string]any{"id": sa.Id, "name": sa.Name, "clientId": sa.ClientId, "enabled": sa.Enabled}, true, nil
}

// findUserByIDOrEmail resolves a user by internal UUID first, then falls back to username
// (email). Lets an access-rule subject be given either way and still resolve to the real user.
func findUserByIDOrEmail(ctx context.Context, client *runaiclient.Client, idOrEmail string) (*identity.User, error) {
	if u, _, err := client.Identity.UsersAPI.GetUserById(ctx, idOrEmail).Execute(); err == nil && u != nil {
		return u, nil
	}
	return findUserByEmail(ctx, client, idOrEmail)
}

// findUserByEmail returns the user whose username exactly equals email, or nil if none match.
func findUserByEmail(ctx context.Context, client *runaiclient.Client, email string) (*identity.User, error) {
	users, httpResp, err := client.Identity.UsersAPI.GetUsers(ctx).Search(email).Execute()
	if err != nil {
		return nil, fmt.Errorf("find user %q: %w", email, runaimcp.MapHTTPError(httpResp, err))
	}
	for i := range users {
		if users[i].Username == email {
			return &users[i], nil
		}
	}
	return nil, nil
}

// findServiceAccount returns the service-account matching id (preferred) or name, or nil if none match.
func findServiceAccount(ctx context.Context, client *runaiclient.Client, id, name string) (*identity.ServiceAccount, error) {
	sas, httpResp, err := client.Identity.ServiceAccountsAPI.GetServiceAccounts(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("list service-accounts: %w", runaimcp.MapHTTPError(httpResp, err))
	}
	for i := range sas {
		if sas[i].Id == id || (id == "" && sas[i].Name == name) {
			return &sas[i], nil
		}
	}
	return nil, nil
}

func doIdentityDelete(ctx context.Context, client *runaiclient.Client, e entity, id string) error {
	if e == entityUser {
		httpResp, err := client.Identity.UsersAPI.DeleteUserById(ctx, id).Execute()
		if err != nil {
			return fmt.Errorf("delete user %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
		}
		return nil
	}
	httpResp, err := client.Identity.ServiceAccountsAPI.DeleteServiceAccountById(ctx, id).Execute()
	if err != nil {
		return fmt.Errorf("delete service-account %q: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	return nil
}

func createIdentityPreview(in manageUserAccessInput) map[string]any {
	switch entity(in.Entity) {
	case entityUser:
		m := map[string]any{"username": in.Email}
		if in.FirstName != "" {
			m["firstName"] = in.FirstName
		}
		if in.LastName != "" {
			m["lastName"] = in.LastName
		}
		return m
	default:
		return map[string]any{"name": in.Name}
	}
}

func userSnapshot(u identity.User) map[string]any {
	snap := map[string]any{"id": u.Id, "username": u.Username}
	if u.FirstName != nil {
		snap["firstName"] = *u.FirstName
	}
	if u.LastName != nil {
		snap["lastName"] = *u.LastName
	}
	return snap
}

func identityLabel(in manageUserAccessInput) string {
	if entity(in.Entity) == entityUser {
		if in.Email != "" {
			return in.Email
		}
		return in.ID
	}
	if in.Name != "" {
		return in.Name
	}
	return in.ID
}

func credentialWarning(e entity) string {
	if e == entityUser {
		return "tempPassword is shown once and cannot be retrieved again — store it securely now"
	}
	return "clientSecret is shown once and cannot be retrieved again — store it securely now"
}

func strValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
