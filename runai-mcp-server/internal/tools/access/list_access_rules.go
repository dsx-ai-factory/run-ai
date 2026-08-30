package access

import (
	"context"
	"fmt"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type listAccessRulesInput struct {
	SubjectID   string `json:"subjectId,omitempty"   jsonschema:"Filter to rules for this principal — a user/service-account ID or a group name."`
	SubjectType string `json:"subjectType,omitempty" jsonschema:"Filter by principal kind. One of: user, service-account, group, app."`
	ScopeID     string `json:"scopeId,omitempty"     jsonschema:"Filter to rules at this scope ID (org-unit ID, cluster ID, or tenant ID)."`
	ClusterID   string `json:"clusterId,omitempty"   jsonschema:"Filter to rules on this cluster ID."`
	Limit       int32  `json:"limit,omitempty"       jsonschema:"Maximum number of rules to return. Defaults to 50, capped at 200."`
	Offset      int32  `json:"offset,omitempty"      jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listAccessRulesInput) Validate() error {
	if err := common.ValidatePagination(in.Limit, in.Offset); err != nil {
		return err
	}
	for field, v := range map[string]string{"subjectId": in.SubjectID, "scopeId": in.ScopeID, "clusterId": in.ClusterID} {
		if err := common.ValidateNoCommas(field, v); err != nil {
			return err
		}
	}
	return nil
}

type accessRuleStatus struct {
	Phase   string `json:"phase,omitempty"   jsonschema:"Provisioning phase of the rule: Creating, Ready, NotReady, or Deleting. The console shows NotReady as 'issues found'."`
	Message string `json:"message,omitempty" jsonschema:"Backend status message explaining the phase, when present."`
	Details string `json:"details,omitempty" jsonschema:"Additional backend status details, when present."`
}

type accessRuleInfo struct {
	ID          int32             `json:"id"                  jsonschema:"Access-rule ID — pass as id to manage_user_access delete."`
	SubjectID   string            `json:"subjectId"           jsonschema:"ID of the principal the role is granted to."`
	SubjectType string            `json:"subjectType"         jsonschema:"Kind of principal: user, service-account, group, or app."`
	RoleID      int32             `json:"roleId"              jsonschema:"Granted role ID."`
	RoleName    string            `json:"roleName"            jsonschema:"Granted role name."`
	ScopeType   string            `json:"scopeType"           jsonschema:"Scope kind the role applies at: system, tenant, cluster, department, or project."`
	ScopeID     string            `json:"scopeId"             jsonschema:"ID of the scope the role applies at."`
	ScopeName   string            `json:"scopeName,omitempty" jsonschema:"Human-readable scope name, when available."`
	ClusterID   string            `json:"clusterId,omitempty" jsonschema:"Cluster the rule applies to, for cluster/department/project scopes."`
	Status      *accessRuleStatus `json:"status,omitempty"    jsonschema:"Provisioning status of the rule, when the backend reports one. Absent means the API returned no status — do not infer or invent one."`
}

type listAccessRulesOutput struct {
	AccessRules []accessRuleInfo `json:"accessRules"          jsonschema:"Access rules matching the filters. Always present; empty array when none match."`
	HasMore     bool             `json:"hasMore"              jsonschema:"True when more rules exist beyond the returned page; pass nextOffset back as offset."`
	NextOffset  int32            `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call to fetch more rules. Only set when hasMore is true."`
}

func registerListAccessRules(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listAccessRulesInput, listAccessRulesOutput]{
		Name: "list_access_rules",
		Description: "List access-rules (role grants) with optional filters by principal, scope, or cluster. Call this to " +
			"see who has which role where, to confirm a grant created via manage_user_access, or to find a rule's numeric " +
			"id before deleting it. Filter by subjectId to list a specific principal's grants.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listAccessRulesInput) (listAccessRulesOutput, error) {
			return listAccessRulesRun(ctx, client, in)
		},
	})
}

func listAccessRulesRun(ctx context.Context, client *runaiclient.Client, in listAccessRulesInput) (listAccessRulesOutput, error) {
	limit := common.EffectiveLimit(in.Limit)
	req := client.Authorization.AccessRulesAPI.GetAccessRules(ctx).Limit(limit).Offset(in.Offset)
	if in.SubjectID != "" {
		req = req.SubjectIds([]string{in.SubjectID})
	}
	if in.SubjectType != "" {
		req = req.SubjectType(in.SubjectType)
	}
	if in.ScopeID != "" {
		req = req.ScopeId(in.ScopeID)
	}
	if in.ClusterID != "" {
		req = req.ClusterId(in.ClusterID)
	}

	resp, httpResp, err := req.Execute()
	if err != nil {
		return listAccessRulesOutput{}, fmt.Errorf("list access-rules: %w", runaimcp.MapHTTPError(httpResp, err))
	}
	rules := []accessRuleInfo{}
	if resp != nil {
		for i := range resp.AccessRules {
			r := resp.AccessRules[i]
			rules = append(rules, accessRuleInfo{
				ID: r.Id, SubjectID: r.SubjectId, SubjectType: string(r.SubjectType),
				RoleID: r.RoleId, RoleName: r.RoleName, ScopeType: string(r.ScopeType),
				ScopeID: r.ScopeId, ScopeName: r.ScopeName, ClusterID: strValue(r.ClusterId),
				Status: accessRuleStatusFrom(r.Status),
			})
		}
	}

	out := listAccessRulesOutput{AccessRules: rules}
	if resp != nil && in.Offset+int32(len(rules)) < resp.TotalRecords {
		out.HasMore = true
		out.NextOffset = in.Offset + int32(len(rules))
	}
	return out, nil
}

func accessRuleStatusFrom(s *authz.AccessRuleStatusResponse) *accessRuleStatus {
	if s == nil {
		return nil
	}
	out := &accessRuleStatus{Message: strValue(s.Message), Details: strValue(s.Details)}
	if s.Phase != nil {
		out.Phase = string(*s.Phase)
	}
	if out.Phase == "" && out.Message == "" && out.Details == "" {
		return nil
	}
	return out
}
