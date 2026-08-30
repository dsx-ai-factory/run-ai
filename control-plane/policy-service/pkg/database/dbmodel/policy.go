package dbmodel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	authApi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"gorm.io/gorm"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const (
	IncludeDeletedPolicies      = true
	NotToIncludeDeletedPolicies = false
	GvkRulesJsonKey             = "rules"
)

type Policy struct {
	dao_utils.BaseModel // enables soft-delete
	Name                string
	WorkloadType        string
	Scope               api.ScopeType
	TenantId            int32
	ClusterId           *string
	DepartmentId        *string
	ProjectId           *int32
	DefaultsV2          dao_utils.JsonBMap
	RulesV2             dao_utils.JsonBMap
	ImposedAssets       *api.DistributedImposedAssets `gorm:"serializer:json; type:jsonb;"`
	GvkWorkload         bool                          `gorm:"default:false"`
	ValidationStatus    *api.PolicyPhase              `gorm:"column:validation_status;type:text"`
	ValidationIssues    []string                      `gorm:"serializer:json;type:jsonb;column:validation_issues"`
}

func (Policy) TableName() string {
	return dbtables.PoliciesTable
}

func (p Policy) IsDeleted() bool {
	return p.DeletedAt.Valid
}

func (p Policy) IsEmpty() bool {
	if len(p.RulesV2) != 0 {
		return false
	}
	if len(p.DefaultsV2) != 0 {
		return false
	}

	if p.ImposedAssets == nil {
		return true
	}

	if p.ImposedAssets.Worker != nil && len(*p.ImposedAssets.Worker) != 0 {
		return false
	}
	if p.ImposedAssets.Master != nil && len(*p.ImposedAssets.Master) != 0 {
		return false
	}

	return true
}

// HasContent returns true if the policy has content
// meaning it has defined defaults, rules or imposed assets, its important beacuse in policy we have placeholder policies with no content
// and we need to be able to distinguish them
func (p Policy) HasContent() bool {
	return len(convutils.MapMe(p.DefaultsV2.AsMAp())) != 0 || len(convutils.MapMe(p.RulesV2.AsMAp())) != 0 || p.ImposedAssets != nil
}

func (p Policy) Compare(other interface{}) int {
	otherPolicy, ok := other.(Policy)
	if !ok {
		logging.Fatalf("Comparison can only be done between two policy instances")
	}
	return strings.Compare(p.ID.String(), otherPolicy.ID.String())
}

func (p Policy) OrgUnit() scopes.OrgUnit {
	result := scopes.OrgUnit{
		Scope: p.Scope,
	}
	if p.Scope == api.SCOPETYPE_TENANT {
		result.Id = fmt.Sprintf("%d", p.TenantId)
	}
	if p.Scope == api.SCOPETYPE_DEPARTMENT {
		result.Id = convutils.Deref(p.DepartmentId, "")
	}
	if p.Scope == api.SCOPETYPE_PROJECT {
		result.Id = fmt.Sprintf("%d", convutils.Deref(p.ProjectId, 0))
	}
	if p.Scope == api.SCOPETYPE_CLUSTER {
		result.Id = convutils.Deref(p.ClusterId, "")
	}
	return result
}

func (p *Policy) ToWorkspaceCreationFields() *api.PolicyCreationFields {
	return &api.PolicyCreationFields{
		Scope:        p.Scope,
		ProjectId:    p.ProjectId,
		DepartmentId: p.DepartmentId,
		ClusterId:    p.ClusterId,
		Name:         p.Name,
	}
}

func (p *Policy) ApplyScopeQueryParams(workloadType string, scope string, departmentId string, projectId int32, clusterId string) {
	if workloadType != "" {
		p.WorkloadType = workloadType
	}

	if scope != "" {
		p.Scope = api.ScopeType(scope)
	}

	if departmentId != "" {
		p.DepartmentId = &departmentId
	}

	if projectId != 0 {
		p.ProjectId = &projectId
	}

	if clusterId != "" {
		p.ClusterId = &clusterId
	}
}

func (Policy) LeftJoinEffectivePolicy(includeDeletedRecords bool) string {
	result := fmt.Sprintf("LEFT JOIN %s ON (%s.%s = %s.%s)",
		dbtables.EffectivePoliciesTable,
		dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnPolicyId,
		dbtables.PoliciesTable, dbtables.PoliciesColumnId)

	if !includeDeletedRecords {
		result += fmt.Sprintf(" AND (%s.%s IS NULL)", dbtables.EffectivePoliciesTable, dbtables.DeletedAtColumn)
	}

	return result
}

func (EffectivePolicy) JoinWithPolicy() string {
	return fmt.Sprintf("JOIN %s ON %s.%s = %s.%s",
		dbtables.PoliciesTable,
		dbtables.EffectivePoliciesTable, dbtables.EffectivePoliciesColumnPolicyId,
		dbtables.PoliciesTable, dbtables.PoliciesColumnId)
}

func ApplyPermittedScopes(tx *gorm.DB, permittedScopes *authApi.PermittedScopes) {
	if permittedScopes == nil || convutils.BoolIsTrue(permittedScopes.System) {
		return
	}

	if len(permittedScopes.Tenants) != 0 {
		tx.Where(fmt.Sprintf("%s.%s in ?", dbtables.PoliciesTable, dbtables.PoliciesColumnTenantId), permittedScopes.Tenants)
		return
	}

	if len(permittedScopes.Clusters) != 0 {
		tx.Where(fmt.Sprintf("%s.%s in ?", dbtables.PoliciesTable, dbtables.PoliciesColumnClusterId), permittedScopes.Clusters)
	}

	if len(permittedScopes.Departments) != 0 {
		tx.Where(fmt.Sprintf("%s.%s in ?", dbtables.PoliciesTable, dbtables.PoliciesColumnDepartmentId), permittedScopes.Departments)
	}

	if len(permittedScopes.Projects) != 0 {
		tx.Where(fmt.Sprintf("%s.%s in ?", dbtables.PoliciesTable, dbtables.PoliciesColumnProjectId), permittedScopes.Projects)
	}
}

func (p *PolicyWithOptionalEffective) ToGvkPolicyApi() (*api.Policy, *api.Error) {
	if !p.GvkWorkload {
		return nil, api.NewError(http.StatusInternalServerError, "stored policy is not a GVK workload", nil)
	}

	gvk, err := DecodeGvkWorkloadType(p.WorkloadType)
	if err != nil {
		return nil, err
	}

	rules, rulesErr := FromGvkRulesApi[api.PolicyRule](p.Policy.RulesV2)
	if rulesErr != nil {
		return nil, rulesErr
	}

	effectiveRules, effectiveRulesErr := FromGvkRulesApi[api.EffectiveRulesEffectiveRulesInner](p.EffectivePolicy.RulesV2)
	if effectiveRulesErr != nil {
		return nil, effectiveRulesErr
	}

	var projectId *string
	if p.ProjectId != nil {
		s := fmt.Sprintf("%d", *p.ProjectId)
		projectId = &s
	}

	var status *api.PolicyStatus
	if p.ValidationStatus != nil {
		status = &api.PolicyStatus{
			Phase:  *p.ValidationStatus,
			Issues: p.ValidationIssues,
		}
	}

	return &api.Policy{
		Description:    convutils.Ref(p.Name),
		WorkloadType:   gvk,
		Scope:          p.Scope,
		ProjectId:      projectId,
		Rules:          rules,
		EffectiveRules: effectiveRules,
		Id:             p.ID.String(),
		CreatedBy:      p.CreatedBy,
		CreatedAt:      p.CreatedAt,
		UpdatedBy:      p.UpdatedBy,
		UpdatedAt:      p.UpdatedAt,
		Status:         status,
	}, nil
}

func DecodeGvkWorkloadType(workloadType string) (api.Gvk, *api.Error) {
	parts := strings.Split(workloadType, "/")
	if len(parts) != 3 {
		return api.Gvk{}, api.NewError(http.StatusInternalServerError, fmt.Sprintf("stored workload type '%s' is not valid GVK", workloadType), nil)
	}

	return api.Gvk{
		Group:   convutils.Ref(parts[0]),
		Kind:    parts[1],
		Version: parts[2],
	}, nil
}

type GvkPolicyRules interface {
	api.PolicyRule | api.EffectiveRulesEffectiveRulesInner
}

func FromGvkRulesApi[T GvkPolicyRules](m dao_utils.JsonBMap) ([]T, *api.Error) {
	if len(m) == 0 {
		return nil, nil
	}

	rawRules, ok := m[GvkRulesJsonKey]
	if !ok {
		return nil, api.NewError(http.StatusInternalServerError, "stored rules are missing the rules wrapper", nil)
	}

	b, err := json.Marshal(rawRules)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to read stored rules", err)
	}

	var rules []T
	if err := json.Unmarshal(b, &rules); err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "stored rules are not in the expected format", err)
	}

	return rules, nil
}

func ToGvkRulesApi[T GvkPolicyRules](rules []T, scope *scopes.OrgUnit) (dao_utils.JsonBMap, *api.Error) {
	if len(rules) == 0 {
		return nil, nil
	}

	wrap := map[string][]T{
		GvkRulesJsonKey: rules,
	}

	b, err := json.Marshal(wrap)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to serialize policy rules", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to build rules map", err)
	}

	return out, nil
}

func (p Policy) FilterFieldsToColumn() map[string]string {
	return map[string]string{
		"id":    dbtables.PoliciesColumnId,
		"scope": dbtables.PoliciesColumnScope,
		"type":  dbtables.PoliciesColumnWorkloadType,
	}
}

func (p Policy) SortFieldsToColumn() map[string]string {
	return map[string]string{
		"id":        dbtables.PoliciesColumnId,
		"scope":     dbtables.PoliciesColumnScope,
		"type":      dbtables.PoliciesColumnWorkloadType,
		"createdAt": dbtables.CreatedAtColumn,
		"createdBy": dbtables.CreatedByColumn,
		"updatedAt": dbtables.UpdatedAtColumn,
		"updatedBy": dbtables.UpdatedByColumn,
	}
}
