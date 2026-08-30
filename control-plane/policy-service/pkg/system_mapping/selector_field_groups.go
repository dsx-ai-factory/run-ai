package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemSelectorFg builds the system-scoped "selector" field group, used by
// workload mappings whose spec carries a top-level LabelSelector at .spec.selector
// (e.g. apps/Deployment/v1, batch/Job/v1). Field paths and item-selector queries
// are rooted at the workload .spec.
func BuildSystemSelectorFg() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
	fields := BuildFieldGroupFieldMap(selectorFgFields())
	selectors := BuildItemSelectorMap(selectorItemSelectorDefs())
	return fields, selectors
}

func selectorItemSelectorDefs() []SystemItemSel {
	return []SystemItemSel{
		{
			Name:  "allMatchExpressions",
			Desc:  "all label-selector match expressions on the workload selector",
			Query: `(.selector.matchExpressions // [])[]`,
		},
		{
			Name:     "matchExpressionByKey",
			Desc:     "the label-selector match expression for a given label key",
			Query:    `(.selector.matchExpressions // [])[] | select(.key == $ARGS.named.matchExpressionKey)`,
			ArgsInfo: map[string]api.MappingFieldType{"matchExpressionKey": api.MAPPINGFIELDTYPE_STRING},
		},
		{
			Name:     "matchExpressionByOperator",
			Desc:     "all label-selector match expressions that use a given operator (In, NotIn, Exists, or DoesNotExist)",
			Query:    `(.selector.matchExpressions // [])[] | select(.operator == $ARGS.named.matchExpressionOperator)`,
			ArgsInfo: map[string]api.MappingFieldType{"matchExpressionOperator": api.MAPPINGFIELDTYPE_STRING},
			ArgsEnum: map[string][]string{"matchExpressionOperator": {"In", "NotIn", "Exists", "DoesNotExist"}},
		},
		// Read companion for matchLabelsSelector (KeyValueMap is set-only).
		{
			Name:     "matchLabelByKey",
			Desc:     "the selector.matchLabels value for a given label key",
			Query:    `.selector.matchLabels[$ARGS.named.key]`,
			ArgsInfo: map[string]api.MappingFieldType{"key": api.MAPPINGFIELDTYPE_STRING},
		},
	}
}

func selectorFgFields() []SystemFgField {
	keyValue := dbmodel.BuiltinFormatterKeyValue
	keyValueMap := dbmodel.BuiltinFormatterKeyValueMap
	selector := SystemSelectorFormatterName

	selectorsMatchExpressions := []string{"allMatchExpressions", "matchExpressionByKey", "matchExpressionByOperator"}

	selectorsMatchLabelByKey := []string{"matchLabelByKey"}

	return []SystemFgField{
		////////////////////////////////////////////////////////
		// matchLabelsSelector (KeyValueMap, set-only) + read companion
		////////////////////////////////////////////////////////
		{
			Name:      "matchLabelsSelector",
			Formatter: keyValueMap,
			SpecRef:   convutils.Ref("selector.matchLabels"),
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "matchLabelsSelector-value",
			Formatter:     selector,
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsMatchLabelByKey,
		},
		////////////////////////////////////////////////////////
		// matchExprSelector (custom rule-builder)
		////////////////////////////////////////////////////////
		{
			Name:      "matchExprSelector",
			Formatter: selector,
			Type:      api.MAPPINGFIELDTYPE_OBJECT,
		},
		{
			Name:          "matchExprSelector-Operator",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("operator"),
			Type:          api.MAPPINGFIELDTYPE_STRING,
			ItemSelectors: selectorsMatchExpressions,
		},
		{
			Name:          "matchExprSelector-Values",
			Formatter:     keyValue,
			SpecRef:       convutils.Ref("values"),
			Type:          api.MAPPINGFIELDTYPE_STRINGS,
			ItemSelectors: selectorsMatchExpressions,
		},
	}
}
