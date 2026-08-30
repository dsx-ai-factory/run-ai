package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

func BuildSystemJobSpecFormatterFields() map[string]api.FormatterField {
	m := make(map[string]api.FormatterField)

	////////////////////////////////////////////////////////
	// completionMode
	////////////////////////////////////////////////////////
	addJobSpecField(m, "completionMode", `{
	  "completionMode": $ARGS.named.value
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"NonIndexed", "Indexed"},
		})

	////////////////////////////////////////////////////////
	// podReplacementPolicy
	////////////////////////////////////////////////////////
	addJobSpecField(m, "podReplacementPolicy", `{
	  "podReplacementPolicy": $ARGS.named.value
	}`, nil, nil, nil,
		map[string][]string{
			"value": {"TerminatingOrFailed", "Failed"},
		})

	////////////////////////////////////////////////////////
	// podFailurePolicyRules
	////////////////////////////////////////////////////////
	addJobSpecField(m, "podFailurePolicyRules", `{
	  "podFailurePolicy": {
		"rules": [
		  {
			"action": $ARGS.named.action,
			"onExitCodes": $ARGS.named.onExitCodes,
			"onPodConditions": $ARGS.named.onPodConditions
		  }
		]
	  }
	}`, nil,
		&api.FormatterArrayKey{ArgName: "action"},
		map[string]api.MappingFieldType{
			"onExitCodes":     api.MAPPINGFIELDTYPE_OBJECT,
			"onPodConditions": api.MAPPINGFIELDTYPE_OBJECT,
		},
		map[string][]string{
			"action": {"FailJob", "Ignore", "Count", "FailIndex"},
		})

	////////////////////////////////////////////////////////
	// podFailurePolicyRules-onExitCodes
	//
	// Rooted at a single rule (selected via item selectors on the field).
	// Sets/defaults the rule's onExitCodes object.
	////////////////////////////////////////////////////////
	addJobSpecField(m, "podFailurePolicyRules-onExitCodes", `{
	  "onExitCodes": {
		"containerName": $ARGS.named.containerName,
		"operator": $ARGS.named.operator,
		"values": $ARGS.named.values
	  }
	}`, nil, nil,
		map[string]api.MappingFieldType{
			"values": api.MAPPINGFIELDTYPE_INTEGERS,
		},
		map[string][]string{
			"operator": {"In", "NotIn"},
		})

	////////////////////////////////////////////////////////
	// podFailurePolicyRules-onPodConditions
	//
	// Rooted at a single rule (selected via item selectors on the field).
	// Sets/defaults one onPodConditions entry, keyed by status.
	////////////////////////////////////////////////////////
	addJobSpecField(m, "podFailurePolicyRules-onPodConditions", `{
	  "onPodConditions": [
		{
		  "type": $ARGS.named.type,
		  "status": $ARGS.named.status
		}
	  ]
	}`, nil,
		&api.FormatterArrayKey{ArgName: "status"}, nil,
		map[string][]string{
			"status": {"True", "False", "Unknown"},
		})

	////////////////////////////////////////////////////////
	// successPolicyRules
	////////////////////////////////////////////////////////
	addJobSpecField(m, "successPolicyRules", `{
	  "successPolicy": {
		"rules": [
		  {
			"succeededCount": $ARGS.named.succeededCount,
			"succeededIndexes": $ARGS.named.succeededIndexes
		  }
		]
	  }
	}`, nil,
		&api.FormatterArrayKey{ArgName: "succeededIndexes"},
		map[string]api.MappingFieldType{
			"succeededCount": api.MAPPINGFIELDTYPE_INTEGER,
		}, nil)

	return m
}

func addJobSpecField(
	dst map[string]api.FormatterField,
	name string,
	setExpr string,
	getExpr *string,
	arrayKey *api.FormatterArrayKey,
	argsInfo map[string]api.MappingFieldType,
	argsEnum map[string][]string,
) {
	ff := api.FormatterField{
		SetExpr:  &setExpr,
		GetExpr:  getExpr,
		ArrayKey: arrayKey,
	}
	if len(argsInfo) > 0 {
		ff.ArgsInfo = argsInfo
	}
	if len(argsEnum) > 0 {
		ff.ArgsEnum = argsEnum
	}
	dst[name] = ff
}
