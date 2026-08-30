package policy_compliance

import (
	"fmt"
	"sort"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
)

// ComplianceError includes both the workload field and the policy field. for flat fields, they are the same
// for itemized, the policy may include also sections for instances/attributes in the rules so we have to add
// them so the user can locate the correct place in the policy that the error is related to.
type ComplianceError struct {
	WorkloadField string
	PolicyField   string
	Details       string
	Rule          *workloadsApi.PolicyRuleEnum
}

type ComplianceErrors []ComplianceError

func (cea *ComplianceErrors) Add(rule *workloadsApi.PolicyRuleEnum, details string, workloadFieldName string, policyField string) {

	*cea = append(*cea, ComplianceError{
		WorkloadField: workloadFieldName,
		PolicyField:   policyField,
		Details:       details,
		Rule:          rule,
	})
}

func (cse *ComplianceErrors) String() string {
	result := ""
	for _, complianceError := range *cse {
		if complianceError.WorkloadField != "" {
			if complianceError.WorkloadField != complianceError.PolicyField {
				result += fmt.Sprintf("%s -> %s: %s\n", complianceError.WorkloadField, complianceError.PolicyField, complianceError.Details)
			} else {
				result += fmt.Sprintf("%s -> %s\n", complianceError.PolicyField, complianceError.Details)
			}
		} else {
			result += fmt.Sprintf("%s\n", complianceError.Details)
		}
	}
	return strings.TrimSuffix(result, "\n")
}

func (cse *ComplianceErrors) AsSortedArray() []string {
	result := make([]string, 0, len(*cse))

	for _, complianceError := range *cse {
		rule := convutils.Deref(complianceError.Rule, "UnknownRule")

		if complianceError.WorkloadField != "" {
			if complianceError.WorkloadField != complianceError.PolicyField {
				result = append(result, fmt.Sprintf("%s -> %s: [%s] %s", complianceError.WorkloadField, complianceError.PolicyField, rule, complianceError.Details))
			} else {
				result = append(result, fmt.Sprintf("%s -> [%s] %s", complianceError.PolicyField, rule, complianceError.Details))
			}
		} else if complianceError.PolicyField != "" {
			result = append(result, fmt.Sprintf("%s -> [%s] %s", complianceError.PolicyField, rule, complianceError.Details))
		} else {
			result = append(result, complianceError.Details)
		}
	}

	sort.Strings(result)
	return result
}

// ForWorkloads we need the details + the workload field
func (cse *ComplianceErrors) ForWorkloads() []workloadsApi.ComplianceIssuesV2ComplianceIssuesInner {
	result := make([]workloadsApi.ComplianceIssuesV2ComplianceIssuesInner, 0, len(*cse))

	for _, complianceError := range *cse {
		curError := workloadsApi.ComplianceIssuesV2ComplianceIssuesInner{
			Field:   complianceError.WorkloadField,
			Details: complianceError.Details,
		}

		if complianceError.Rule != nil {
			curError.Rule = *workloadsApi.NewNullablePolicyRuleEnum(complianceError.Rule)
		}

		result = append(result, curError)
	}

	return result
}

func (cse *ComplianceErrors) ForWorkloadsV2() []workloadsManagerApi.ComplianceIssuesV2ComplianceIssuesInner {
	result := make([]workloadsManagerApi.ComplianceIssuesV2ComplianceIssuesInner, 0, len(*cse))

	for _, complianceError := range *cse {
		issue := workloadsManagerApi.ComplianceIssuesV2ComplianceIssuesInner{
			Field:   complianceError.WorkloadField,
			Details: complianceError.Details,
		}

		if complianceError.Rule != nil {
			rule := workloadsManagerApi.PolicyRuleEnum(*complianceError.Rule)
			issue.Rule = *workloadsManagerApi.NewNullablePolicyRuleEnum(&rule)
		}

		result = append(result, issue)
	}

	return result
}

// ForAssets we need the details + the policy field, we do not care about the workload field
func (cse *ComplianceErrors) ForAssets() []workloadsApi.ComplianceIssuesComplianceIssuesInner {
	result := make([]workloadsApi.ComplianceIssuesComplianceIssuesInner, 0, len(*cse))

	for _, complianceError := range *cse {
		result = append(result, workloadsApi.ComplianceIssuesComplianceIssuesInner{
			Field:   complianceError.PolicyField,
			Details: complianceError.Details,
		})
	}

	return result
}
