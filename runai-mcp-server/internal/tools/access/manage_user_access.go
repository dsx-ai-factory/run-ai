/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package access

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type entity string

const (
	entityUser           entity = "user"
	entityServiceAccount entity = "service-account"
	entityAccessRule     entity = "access-rule"
)

func (e entity) valid() bool {
	switch e {
	case entityUser, entityServiceAccount, entityAccessRule:
		return true
	default:
		return false
	}
}

type manageUserAccessInput struct {
	Entity string `json:"entity" jsonschema:"What to manage. One of: user, service-account, access-rule. user and service-account are identities (principals); access-rule grants a role to a principal at a scope."`
	Action string `json:"action" jsonschema:"What to do. One of: create, delete."`

	ID        string `json:"id,omitempty"        jsonschema:"Delete-only target: the user/service-account ID, or the numeric access-rule ID. Identities may instead be identified by email (user) or name (service-account)."`
	Email     string `json:"email,omitempty"     jsonschema:"For entity=user: email of the user to create; also resolves a user for delete when id is omitted."`
	Name      string `json:"name,omitempty"      jsonschema:"For entity=service-account: name to create (lowercase alphanumeric and hyphens, e.g. ci-bot); also resolves the service-account for delete."`
	FirstName string `json:"firstName,omitempty" jsonschema:"For entity=user, create-only: optional first name."`
	LastName  string `json:"lastName,omitempty"  jsonschema:"For entity=user, create-only: optional last name."`

	SubjectID    string `json:"subjectId,omitempty"    jsonschema:"For entity=access-rule: the principal — a user's email (NOT the UUID), a service-account clientId, or a group name. Or use subjectEmail/subjectName."`
	SubjectEmail string `json:"subjectEmail,omitempty" jsonschema:"For entity=access-rule with subjectType=user: the subject user's email, resolved to its ID. Alternative to subjectId."`
	SubjectName  string `json:"subjectName,omitempty"  jsonschema:"For entity=access-rule with subjectType=service-account: the subject's name, resolved to its ID. Alternative to subjectId."`
	SubjectType  string `json:"subjectType,omitempty"  jsonschema:"For entity=access-rule: kind of principal. One of: user, service-account, group."`
	RoleName     string `json:"roleName,omitempty"    jsonschema:"For entity=access-rule, create-only: role to grant by name (e.g. 'L1 researcher'); resolved to its ID. Mutually exclusive with roleId."`
	RoleID       *int32 `json:"roleId,omitempty"      jsonschema:"For entity=access-rule, create-only: role to grant by ID. Mutually exclusive with roleName."`
	ScopeType    string `json:"scopeType,omitempty"   jsonschema:"For entity=access-rule, create-only: scope kind the role applies at. One of: system, tenant, cluster, department, project."`
	ScopeID      string `json:"scopeId,omitempty"     jsonschema:"For entity=access-rule, create-only: scope ID as a string even when numeric — org-unit ID for project/dept, cluster ID, or tenant ID. Omit for system."`
	ScopeName    string `json:"scopeName,omitempty"   jsonschema:"For entity=access-rule, scopeType=department|project: org-unit name, resolved to scopeId (needs clusterId or clusterName). Alternative to scopeId."`
	ClusterID    string `json:"clusterId,omitempty"   jsonschema:"For entity=access-rule, create-only: cluster ID, required for cluster/department/project scoped rules."`
	ClusterName  string `json:"clusterName,omitempty" jsonschema:"For entity=access-rule: cluster name, resolved to its ID (fills clusterId, and scopeId when scopeType=cluster). Alternative to clusterId."`

	DryRun        bool `json:"dryRun,omitempty"        jsonschema:"When true, preview the change (before/after/diff) without mutating anything upstream."`
	IgnoreMissing bool `json:"ignoreMissing,omitempty" jsonschema:"Delete-only. When true, deleting a target that no longer exists is a successful no-op instead of not_found."`
	Confirm       bool `json:"confirm,omitempty"       jsonschema:"Required for a non-dry-run delete when the client can't prompt. Without it the call returns confirmation_required with the target snapshot."`
}

func (in manageUserAccessInput) Validate() error {
	if !entity(in.Entity).valid() {
		return fmt.Errorf("entity must be one of: user, service-account, access-rule")
	}
	switch common.WriteAction(in.Action) {
	case common.ActionCreate, common.ActionDelete:
	default:
		return fmt.Errorf("action must be one of: create, delete")
	}
	if in.IgnoreMissing && common.WriteAction(in.Action) != common.ActionDelete {
		return fmt.Errorf("ignoreMissing is only valid with action=delete")
	}

	switch entity(in.Entity) {
	case entityUser, entityServiceAccount:
		return in.validateIdentity()
	case entityAccessRule:
		return in.validateAccessRule()
	}
	return nil
}

func (in manageUserAccessInput) validateIdentity() error {
	if err := in.rejectAccessRuleFields(); err != nil {
		return err
	}
	isUser := entity(in.Entity) == entityUser
	if !isUser && (in.FirstName != "" || in.LastName != "") {
		return fmt.Errorf("firstName and lastName are only valid for entity=user")
	}
	if isUser && in.Name != "" {
		return fmt.Errorf("use email (not name) to identify a user")
	}
	if !isUser && in.Email != "" {
		return fmt.Errorf("email is only valid for entity=user")
	}

	switch common.WriteAction(in.Action) {
	case common.ActionCreate:
		if in.ID != "" {
			return fmt.Errorf("id must not be set for create")
		}
		if isUser {
			if strings.TrimSpace(in.Email) == "" {
				return fmt.Errorf("email is required to create a user")
			}
		} else if strings.TrimSpace(in.Name) == "" {
			return fmt.Errorf("name is required to create a %s", in.Entity)
		}
	case common.ActionDelete:
		if in.FirstName != "" || in.LastName != "" {
			return fmt.Errorf("firstName and lastName are only valid on create")
		}
		if isUser {
			if in.ID == "" && in.Email == "" {
				return fmt.Errorf("one of id or email is required to delete a user")
			}
		} else if in.ID == "" && in.Name == "" {
			return fmt.Errorf("one of id or name is required to delete a %s", in.Entity)
		}
	}
	if in.Email != "" {
		if err := common.ValidateFilterValue("email", in.Email); err != nil {
			return err
		}
	}
	if in.Name != "" {
		if err := common.ValidateFilterValue("name", in.Name); err != nil {
			return err
		}
	}
	return nil
}

func (in manageUserAccessInput) validateAccessRule() error {
	if in.Email != "" || in.Name != "" || in.FirstName != "" || in.LastName != "" {
		return fmt.Errorf("email, name, firstName, lastName are not valid for entity=access-rule")
	}
	switch common.WriteAction(in.Action) {
	case common.ActionCreate:
		if in.ID != "" {
			return fmt.Errorf("id must not be set for create")
		}
		if err := in.validateSubject(); err != nil {
			return err
		}
		return in.validateRoleAndScope()
	case common.ActionDelete:
		// Delete by numeric id, or by the (subject + role + scope) tuple — not both.
		if in.ID != "" {
			if in.subjectGiven() || in.roleGiven() || in.ScopeType != "" || in.ScopeID != "" || in.ClusterID != "" {
				return fmt.Errorf("delete an access-rule by id alone, or by the subject+role+scope tuple — not both")
			}
			return nil
		}
		if !in.subjectGiven() && !in.roleGiven() && in.ScopeType == "" {
			return fmt.Errorf("to delete an access-rule, give its numeric id, or the subject+role+scope tuple")
		}
		if err := in.validateSubject(); err != nil {
			return err
		}
		return in.validateRoleAndScope()
	}
	return nil
}

func (in manageUserAccessInput) subjectGiven() bool {
	return in.SubjectID != "" || in.SubjectEmail != "" || in.SubjectName != "" || in.SubjectType != ""
}

func (in manageUserAccessInput) roleGiven() bool {
	return in.RoleName != "" || in.RoleID != nil
}

// validateSubject enforces exactly one subject identifier appropriate to subjectType.
func (in manageUserAccessInput) validateSubject() error {
	if !validSubjectType(in.SubjectType) {
		return fmt.Errorf("subjectType must be one of: user, service-account, group")
	}
	switch in.SubjectType {
	case "user":
		if in.SubjectName != "" {
			return fmt.Errorf("subjectName is for subjectType=service-account; a user subject uses subjectId or subjectEmail")
		}
		if (in.SubjectID == "") == (in.SubjectEmail == "") {
			return fmt.Errorf("provide exactly one of subjectId or subjectEmail for subjectType=user")
		}
	case "service-account":
		if in.SubjectEmail != "" {
			return fmt.Errorf("subjectEmail is for subjectType=user; a service-account subject uses subjectId or subjectName")
		}
		if (in.SubjectID == "") == (in.SubjectName == "") {
			return fmt.Errorf("provide exactly one of subjectId or subjectName for subjectType=service-account")
		}
	default: // group
		if in.SubjectEmail != "" || in.SubjectName != "" {
			return fmt.Errorf("a group subject is identified by subjectId only")
		}
		if in.SubjectID == "" {
			return fmt.Errorf("subjectId is required for subjectType=group")
		}
	}
	return nil
}

func (in manageUserAccessInput) validateRoleAndScope() error {
	if (in.RoleName == "") == (in.RoleID == nil) {
		return fmt.Errorf("exactly one of roleName or roleId is required")
	}
	if in.RoleID != nil && *in.RoleID <= 0 {
		return fmt.Errorf("roleId must be positive")
	}
	return validateScope(in.ScopeType, in.ScopeID, in.ScopeName, in.ClusterID, in.ClusterName)
}

func (in manageUserAccessInput) rejectAccessRuleFields() error {
	if in.SubjectID != "" || in.SubjectEmail != "" || in.SubjectName != "" || in.SubjectType != "" ||
		in.RoleName != "" || in.RoleID != nil || in.ScopeType != "" || in.ScopeID != "" || in.ScopeName != "" ||
		in.ClusterID != "" || in.ClusterName != "" {
		return fmt.Errorf("subject*, role*, scope*, and cluster* fields are only valid for entity=access-rule")
	}
	return nil
}

type manageUserAccessOutput struct {
	Entity   string         `json:"entity"             jsonschema:"The entity acted on: user, service-account, or access-rule."`
	Action   string         `json:"action"             jsonschema:"The action performed: create or delete."`
	DryRun   bool           `json:"dryRun"             jsonschema:"True when this was a preview; no upstream change was made."`
	Applied  bool           `json:"applied"            jsonschema:"True when the change was actually applied upstream. False on dry-run or an ignored-missing delete."`
	ID       string         `json:"id,omitempty"       jsonschema:"Resolved identifier of the target. Empty on a create dry-run."`
	Before   map[string]any `json:"before"             jsonschema:"Snapshot before the change. Empty object on create."`
	After    map[string]any `json:"after"              jsonschema:"Snapshot after the change. Empty object on delete. On identity create it carries the one-time credential (tempPassword / clientSecret / secret)."`
	Diff     map[string]any `json:"diff"               jsonschema:"Field-level changes as path -> {before, after}."`
	Warnings []string       `json:"warnings,omitempty" jsonschema:"Non-fatal notes, e.g. a delete target that was already missing, or a reminder to store a one-time credential."`
}

func registerManageUserAccess(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageUserAccessInput, manageUserAccessOutput]{
		Name: "manage_user_access",
		Description: "Create or delete a Run:ai identity (user or service-account) or an access-rule that grants a role to " +
			"a principal at a scope. The entity field selects what to manage; the action field selects create or delete. To " +
			"grant access, first ensure the principal exists (entity=user|service-account), then create an access-rule " +
			"(entity=access-rule) binding it to a role at a scope. Subjects, roles, and scopes may be given by name " +
			"(subjectEmail/subjectName, roleName, clusterName/scopeName) and are resolved to IDs. For high-impact changes, call once with dryRun=true to " +
			"preview, then again with dryRun=false to apply. A non-dry-run delete requires confirm=true. Creating a user or " +
			"service-account returns its generated credential (tempPassword or clientSecret) once in the after snapshot — it " +
			"cannot be retrieved again, so store it securely. There is no read action: the create response's after snapshot is " +
			"the only echo of the new ID and credential. Do NOT call to list users/roles/access-rules (use the read tools).",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageUserAccessInput) (manageUserAccessOutput, error) {
			return manageUserAccessRun(ctx, client, in)
		},
	})
}

func manageUserAccessRun(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput) (manageUserAccessOutput, error) {
	out := manageUserAccessOutput{
		Entity: in.Entity,
		Action: in.Action,
		DryRun: in.DryRun,
		Before: map[string]any{},
		After:  map[string]any{},
		Diff:   map[string]any{},
	}

	switch entity(in.Entity) {
	case entityUser, entityServiceAccount:
		return manageIdentity(ctx, client, in, out)
	case entityAccessRule:
		return manageAccessRule(ctx, client, in, out)
	default:
		return manageUserAccessOutput{}, fmt.Errorf("unsupported entity %q: %w", in.Entity, runaimcp.ErrValidation)
	}
}
