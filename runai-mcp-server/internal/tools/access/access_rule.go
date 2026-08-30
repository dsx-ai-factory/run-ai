package access

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

func validSubjectType(s string) bool {
	switch authz.SubjectType(s) {
	case authz.SUBJECTTYPE_USER, authz.SUBJECTTYPE_SERVICE_ACCOUNT, authz.SUBJECTTYPE_GROUP:
		return true
	default:
		return false
	}
}

// validateScope enforces the scope identifiers each scope kind requires before the request
// reaches the authorization service. A name alternative (scopeName / clusterName) satisfies the
// requirement; the handler resolves it to an ID.
func validateScope(scopeType, scopeID, scopeName, clusterID, clusterName string) error {
	switch authz.ScopeType(scopeType) {
	case authz.SCOPETYPE_SYSTEM:
	case authz.SCOPETYPE_TENANT:
		if scopeID == "" {
			return fmt.Errorf("scopeId (tenant ID) is required for scopeType=tenant: %w", runaimcp.ErrValidation)
		}
	case authz.SCOPETYPE_CLUSTER:
		if scopeID == "" && clusterName == "" {
			return fmt.Errorf("scopeId (cluster ID) or clusterName is required for scopeType=cluster: %w", runaimcp.ErrValidation)
		}
	case authz.SCOPETYPE_DEPARTMENT, authz.SCOPETYPE_PROJECT:
		if scopeID == "" && scopeName == "" {
			return fmt.Errorf("scopeId or scopeName is required for scopeType=%s: %w", scopeType, runaimcp.ErrValidation)
		}
		if clusterID == "" && clusterName == "" {
			return fmt.Errorf("clusterId or clusterName is required for scopeType=%s: %w", scopeType, runaimcp.ErrValidation)
		}
	default:
		return fmt.Errorf("scopeType must be one of: system, tenant, cluster, department, project: %w", runaimcp.ErrValidation)
	}
	return nil
}

func manageAccessRule(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput, out manageUserAccessOutput) (manageUserAccessOutput, error) {
	if common.WriteAction(in.Action) == common.ActionCreate {
		return accessRuleCreate(ctx, client, in, out)
	}
	return accessRuleDelete(ctx, client, in, out)
}

// authorizeAccessRule fails fast when the caller can't manage access rules at the target
// scope, mapping the eventual upstream 403 to an early forbidden error.
func authorizeAccessRule(ctx context.Context, client *runaiclient.Client, action common.WriteAction, scopeType, scopeID, clusterID string) error {
	scope, err := common.PermittedScopeForAction(ctx, client, authz.RESOURCETYPE_ACCESS_RULES, action.Authz())
	if err != nil {
		return err
	}
	if !scopeAllowsTarget(scope, scopeType, scopeID, clusterID) {
		return fmt.Errorf("not permitted to %s access rules at %s scope %q: %w", action, scopeType, scopeID, runaimcp.ErrForbidden)
	}
	return nil
}

func scopeAllowsTarget(scope authz.PermittedScopes, scopeType, scopeID, clusterID string) bool {
	switch authz.ScopeType(scopeType) {
	case authz.SCOPETYPE_SYSTEM:
		// A system-scoped grant needs an actual system permission — a mere tenant grant is narrower.
		return scope.System != nil && *scope.System
	case authz.SCOPETYPE_TENANT:
		if scope.System != nil && *scope.System {
			return true
		}
		for _, t := range scope.Tenants {
			if t == scopeID {
				return true
			}
		}
		return false
	case authz.SCOPETYPE_CLUSTER:
		return common.ScopeAllowsCluster(scope, scopeID)
	case authz.SCOPETYPE_DEPARTMENT, authz.SCOPETYPE_PROJECT:
		// Defer to upstream: the permitted-scopes probe can't replicate the department→project
		// hierarchy walk without an extra org-unit fetch, so a caller permitted at a parent
		// department would be falsely denied here. Upstream is authoritative and returns 403.
		return true
	default:
		return false
	}
}

func accessRuleCreate(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput, out manageUserAccessOutput) (manageUserAccessOutput, error) {
	scopeID, clusterID, err := resolveScope(ctx, client, in)
	if err != nil {
		return manageUserAccessOutput{}, err
	}
	if err := authorizeAccessRule(ctx, client, common.ActionCreate, in.ScopeType, scopeID, clusterID); err != nil {
		return manageUserAccessOutput{}, err
	}

	var (
		roleID     int32
		roleName   string
		deprecated bool
	)
	if in.RoleID != nil {
		// Resolve by id too, so a dry-run catches an invalid role and the deprecation
		// warning fires on the roleId path as well as the roleName path.
		name, dep, err := resolveRoleByID(ctx, client, *in.RoleID)
		if err != nil {
			return manageUserAccessOutput{}, err
		}
		roleID, roleName, deprecated = *in.RoleID, name, dep
	} else {
		resolved, name, dep, err := resolveRoleID(ctx, client, in.RoleName)
		if err != nil {
			return manageUserAccessOutput{}, err
		}
		roleID, roleName, deprecated = resolved, name, dep
	}
	if deprecated {
		out.Warnings = append(out.Warnings, fmt.Sprintf("role %q is deprecated and should not be used for new assignments", roleName))
	}

	subjectID, err := resolveSubjectID(ctx, client, in)
	if err != nil {
		return manageUserAccessOutput{}, err
	}

	fields := authz.AccessRuleCreationFields{
		SubjectId:   subjectID,
		SubjectType: authz.SubjectType(in.SubjectType),
		RoleId:      roleID,
		ScopeId:     scopeID,
		ScopeType:   authz.ScopeType(in.ScopeType),
	}
	if clusterID != "" {
		fields.ClusterId = &clusterID
	}

	if in.DryRun {
		out.After = accessRuleFieldsPreview(fields, roleName)
		out.Diff = common.DiffMaps(nil, out.After)
		return out, nil
	}

	created, httpResp, err := client.Authorization.AccessRulesAPI.CreateAccessRule(ctx).AccessRuleCreationFields(fields).Execute()
	if err != nil {
		return manageUserAccessOutput{}, fmt.Errorf("create access-rule: %w", runaimcp.MapHTTPError(httpResp, err))
	}
	if created == nil {
		return manageUserAccessOutput{}, fmt.Errorf("create access-rule: nil response: %w", runaimcp.ErrUpstream)
	}
	out.ID = strconv.FormatInt(int64(created.Id), 10)
	out.After = accessRuleSnapshot(created)
	out.Applied = true
	out.Diff = common.DiffMaps(nil, out.After)
	return out, nil
}

// resolvedRule is the access-rule a delete targets, resolved by numeric id or by tuple.
type resolvedRule struct {
	id        int32
	snapshot  map[string]any
	scopeType string
	scopeID   string
	clusterID string
	roleName  string
}

func accessRuleDelete(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput, out manageUserAccessOutput) (manageUserAccessOutput, error) {
	// Tuple delete carries the scope in the input, so authorize before resolving — otherwise
	// resolution (and ignoreMissing) would let an unpermitted caller probe rule existence.
	// The id path can't pre-authorize (scope is only known after the fetch below).
	scopeID, clusterID := in.ScopeID, in.ClusterID
	if in.ID == "" {
		resolvedScope, resolvedCluster, rErr := resolveScope(ctx, client, in)
		if rErr != nil {
			if in.IgnoreMissing && common.IsNotFound(rErr) {
				out.Warnings = append(out.Warnings, "access-rule not found; nothing to delete")
				return out, nil
			}
			return manageUserAccessOutput{}, rErr
		}
		scopeID, clusterID = resolvedScope, resolvedCluster
		if err := authorizeAccessRule(ctx, client, common.ActionDelete, in.ScopeType, scopeID, clusterID); err != nil {
			return manageUserAccessOutput{}, err
		}
	}

	target, err := resolveAccessRuleForDelete(ctx, client, in, scopeID, clusterID)
	if err != nil {
		if in.IgnoreMissing && common.IsNotFound(err) {
			out.Warnings = append(out.Warnings, "access-rule not found; nothing to delete")
			return out, nil
		}
		return manageUserAccessOutput{}, err
	}
	if target == nil {
		if in.IgnoreMissing {
			out.Warnings = append(out.Warnings, "access-rule not found; nothing to delete")
			return out, nil
		}
		return manageUserAccessOutput{}, fmt.Errorf("access-rule not found: %w", runaimcp.ErrNotFound)
	}

	if in.ID != "" {
		if err := authorizeAccessRule(ctx, client, common.ActionDelete, target.scopeType, target.scopeID, target.clusterID); err != nil {
			return manageUserAccessOutput{}, err
		}
	}
	out.ID = strconv.FormatInt(int64(target.id), 10)
	out.Before = target.snapshot
	out.Diff = common.DiffMaps(out.Before, nil)

	if in.DryRun {
		return out, nil
	}
	if err := common.ConfirmWrite(ctx, in.Confirm, "delete", "access-rule", target.roleName, out.ID, out.Before); err != nil {
		return manageUserAccessOutput{}, err
	}

	if httpResp, err := client.Authorization.AccessRulesAPI.DeleteAccessRule(ctx, target.id).Execute(); err != nil {
		return manageUserAccessOutput{}, fmt.Errorf("delete access-rule %d: %w", target.id, runaimcp.MapHTTPError(httpResp, err))
	}
	out.Applied = true
	return out, nil
}

// resolveAccessRuleForDelete finds the target rule by numeric id, or by the subject+role+scope
// tuple. Returns (nil, nil) when no rule matches so the caller can honor ignoreMissing.
func resolveAccessRuleForDelete(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput, scopeID, clusterID string) (*resolvedRule, error) {
	if in.ID != "" {
		id64, err := strconv.ParseInt(in.ID, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("id %q is not a numeric access-rule ID: %w", in.ID, runaimcp.ErrValidation)
		}
		id := int32(id64)
		rule, httpResp, gErr := client.Authorization.AccessRulesAPI.GetAccessRule(ctx, id).Execute()
		if gErr != nil {
			mapped := runaimcp.MapHTTPError(httpResp, gErr)
			if common.IsNotFound(mapped) {
				return nil, nil
			}
			return nil, fmt.Errorf("get access-rule %d: %w", id, mapped)
		}
		if rule == nil {
			return nil, nil
		}
		return &resolvedRule{
			id: rule.Id, snapshot: accessRuleByIDSnapshot(rule), scopeType: string(rule.ScopeType),
			scopeID: rule.ScopeId, clusterID: strValue(rule.ClusterId), roleName: rule.RoleName,
		}, nil
	}
	return findAccessRuleByTuple(ctx, client, in, scopeID, clusterID)
}

// findAccessRuleByTuple resolves the subject and role, then matches the single rule binding that
// subject+role at the given scope. Returns (nil, nil) when no rule matches.
func findAccessRuleByTuple(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput, scopeID, clusterID string) (*resolvedRule, error) {
	subjectID, err := resolveSubjectID(ctx, client, in)
	if err != nil {
		return nil, err
	}
	roleID := int32(0)
	if in.RoleID != nil {
		roleID = *in.RoleID
	} else {
		rid, _, _, rErr := resolveRoleID(ctx, client, in.RoleName)
		if rErr != nil {
			return nil, rErr
		}
		roleID = rid
	}

	// Narrow server-side by subject+type; a large limit guards against the match falling
	// beyond the default page (which would spuriously report not-found).
	resp, httpResp, err := client.Authorization.AccessRulesAPI.GetAccessRules(ctx).
		SubjectIds([]string{subjectID}).SubjectType(in.SubjectType).ScopeId(scopeID).Limit(1000).Execute()
	if err != nil {
		return nil, fmt.Errorf("find access-rule: %w", runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil {
		return nil, nil
	}
	// Match the full tuple in-process rather than trusting the server-side filter: subjectId
	// guards against deleting a different subject's rule, scopeId is matched because a
	// system-scoped delete sends an empty scopeId (no server-side scope filter), and a subject
	// can hold the same role/scope under more than one subjectType.
	for i := range resp.AccessRules {
		r := resp.AccessRules[i]
		if r.RoleId == roleID && r.SubjectId == subjectID && string(r.ScopeType) == in.ScopeType &&
			r.ScopeId == scopeID && string(r.SubjectType) == in.SubjectType && strValue(r.ClusterId) == clusterID {
			return &resolvedRule{
				id: r.Id, snapshot: accessRuleSnapshot(&r), scopeType: string(r.ScopeType),
				scopeID: r.ScopeId, clusterID: strValue(r.ClusterId), roleName: r.RoleName,
			}, nil
		}
	}
	return nil, nil
}

func resolveSubjectID(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput) (string, error) {
	switch in.SubjectType {
	case "user":
		return resolveUserSubject(ctx, client, in)
	case "service-account":
		if in.SubjectID != "" {
			return in.SubjectID, nil
		}
		sa, err := findServiceAccount(ctx, client, "", in.SubjectName)
		if err != nil {
			return "", err
		}
		if sa == nil {
			return "", fmt.Errorf("subject service-account %q not found: %w", in.SubjectName, runaimcp.ErrNotFound)
		}
		// A service account authenticates with subject = its clientId (not the internal UUID),
		// so the access-rule subjectId must be the clientId for the grant to actually match.
		return sa.ClientId, nil
	case "group":
		if in.SubjectID != "" {
			return in.SubjectID, nil
		}
		// Validate() already requires subjectId for groups; this is defense-in-depth so a group
		// can never fall through to another lookup.
		return "", fmt.Errorf("subjectId is required for subjectType=group: %w", runaimcp.ErrValidation)
	default:
		return "", fmt.Errorf("unsupported subjectType %q: %w", in.SubjectType, runaimcp.ErrValidation)
	}
}

// resolveUserSubject returns the user's login username (email) — the identifier the authorization
// service keys user access-rules by. subjectEmail is verified to exist; a subjectId is resolved
// whether it's the internal UUID or the email.
func resolveUserSubject(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput) (string, error) {
	lookup := in.SubjectEmail
	if lookup == "" {
		lookup = in.SubjectID
	}
	u, err := findUserByIDOrEmail(ctx, client, lookup)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", fmt.Errorf("subject user %q not found: %w", lookup, runaimcp.ErrNotFound)
	}
	return u.Username, nil
}

// resolveScope resolves the access-rule scope identifiers, accepting names as alternatives to
// IDs: clusterName → cluster ID, and (for department/project) scopeName → org-unit ID. For
// scopeType=cluster the scope *is* the cluster, so scopeId defaults to the resolved cluster ID.
func resolveScope(ctx context.Context, client *runaiclient.Client, in manageUserAccessInput) (scopeID, clusterID string, err error) {
	clusterID = in.ClusterID
	if clusterID == "" && in.ClusterName != "" {
		clusterID, err = resolveClusterID(ctx, client, in.ClusterName)
		if err != nil {
			return "", "", err
		}
	}
	scopeID = in.ScopeID
	switch authz.ScopeType(in.ScopeType) {
	case authz.SCOPETYPE_CLUSTER:
		if scopeID == "" {
			scopeID = clusterID
		}
	case authz.SCOPETYPE_DEPARTMENT:
		if scopeID == "" && in.ScopeName != "" {
			scopeID, err = common.ResolveOrgUnitID(ctx, client, common.OrgTypeDepartment, in.ScopeName, clusterID)
		}
	case authz.SCOPETYPE_PROJECT:
		if scopeID == "" && in.ScopeName != "" {
			scopeID, err = common.ResolveOrgUnitID(ctx, client, common.OrgTypeProject, in.ScopeName, clusterID)
		}
	}
	return scopeID, clusterID, err
}

func resolveClusterID(ctx context.Context, client *runaiclient.Client, name string) (string, error) {
	clusters, httpResp, err := client.Cluster.ClustersAPI.GetClusters(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("resolve cluster %q: %w", name, runaimcp.MapHTTPError(httpResp, err))
	}
	for _, c := range clusters {
		if c.Name == name {
			return c.Uuid, nil
		}
	}
	return "", fmt.Errorf("cluster %q not found: %w", name, runaimcp.ErrNotFound)
}

// resolveRoleID maps a role name to its ID, erroring on no match so a typo can't silently grant
// the wrong role. Matching prefers an exact hit but falls back to case-insensitive (role names
// like "System administrator" are easy to mis-case). The deprecated flag lets the caller warn
// when binding a role the platform is retiring. Discovery of valid names is via the list_roles tool.
func resolveRoleID(ctx context.Context, client *runaiclient.Client, name string) (id int32, resolvedName string, deprecated bool, err error) {
	resp, httpResp, err := client.Authorization.RolesAPI.GetRoles(ctx).Execute()
	if err != nil {
		return 0, "", false, fmt.Errorf("resolve role %q: %w", name, runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil {
		return 0, "", false, fmt.Errorf("resolve role %q: nil response: %w", name, runaimcp.ErrUpstream)
	}
	fallback := -1
	for i := range resp.Roles {
		switch {
		case resp.Roles[i].Name == name:
			r := resp.Roles[i]
			return r.Id, r.Name, r.Deprecated != nil && *r.Deprecated, nil
		case fallback < 0 && strings.EqualFold(resp.Roles[i].Name, name):
			fallback = i
		}
	}
	if fallback >= 0 {
		r := resp.Roles[fallback]
		return r.Id, r.Name, r.Deprecated != nil && *r.Deprecated, nil
	}
	return 0, "", false, fmt.Errorf("role %q not found (use list_roles to see valid roles): %w", name, runaimcp.ErrNotFound)
}

// resolveRoleByID looks up a role by its ID (name + deprecation), so the roleId create path
// gets the same existence check and deprecated-role warning as the roleName path.
func resolveRoleByID(ctx context.Context, client *runaiclient.Client, id int32) (name string, deprecated bool, err error) {
	resp, httpResp, err := client.Authorization.RolesAPI.GetRoles(ctx).Execute()
	if err != nil {
		return "", false, fmt.Errorf("resolve role %d: %w", id, runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil {
		return "", false, fmt.Errorf("resolve role %d: nil response: %w", id, runaimcp.ErrUpstream)
	}
	for _, role := range resp.Roles {
		if role.Id == id {
			return role.Name, role.Deprecated != nil && *role.Deprecated, nil
		}
	}
	return "", false, fmt.Errorf("role %d not found: %w", id, runaimcp.ErrNotFound)
}

func accessRuleFieldsPreview(f authz.AccessRuleCreationFields, roleName string) map[string]any {
	m := map[string]any{
		"subjectId":   f.SubjectId,
		"subjectType": string(f.SubjectType),
		"roleId":      f.RoleId,
		"scopeId":     f.ScopeId,
		"scopeType":   string(f.ScopeType),
	}
	if roleName != "" {
		m["roleName"] = roleName
	}
	if f.ClusterId != nil {
		m["clusterId"] = *f.ClusterId
	}
	return m
}

func accessRuleSnapshot(r *authz.AccessRule) map[string]any {
	m := map[string]any{
		"id":          r.Id,
		"subjectId":   r.SubjectId,
		"subjectType": string(r.SubjectType),
		"roleId":      r.RoleId,
		"roleName":    r.RoleName,
		"scopeId":     r.ScopeId,
		"scopeType":   string(r.ScopeType),
		"scopeName":   r.ScopeName,
	}
	if r.ClusterId != nil {
		m["clusterId"] = *r.ClusterId
	}
	return m
}

func accessRuleByIDSnapshot(r *authz.AccessRuleById) map[string]any {
	m := map[string]any{
		"id":          r.Id,
		"subjectId":   r.SubjectId,
		"subjectType": string(r.SubjectType),
		"roleId":      r.RoleId,
		"roleName":    r.RoleName,
		"scopeId":     r.ScopeId,
		"scopeType":   string(r.ScopeType),
		"scopeName":   r.ScopeName,
	}
	if r.ClusterId != nil {
		m["clusterId"] = *r.ClusterId
	}
	return m
}
