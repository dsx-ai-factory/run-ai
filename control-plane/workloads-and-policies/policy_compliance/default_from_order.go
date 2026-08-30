package policy_compliance

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

// DefaultFromOrder is a map containing order of fields
// - order of all flat fields is grouped with the key ""
// - for each itemized fields we keep the order of its attributes
type DefaultFromOrder map[string][]string

func NewDefaultsFromOrder(rulesAsMap map[string]interface{}) (DefaultFromOrder, error) {
	result := make(DefaultFromOrder)

	rules := GetRulesForCompliance(rulesAsMap)

	err := determineDefaultsFromOrder(rules, "", &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func determineDefaultsFromOrder(rules map[string]interface{}, key string, result *DefaultFromOrder) error {
	g := datastructs.NewDependenciesGraph[string]()

	for fldName, fldRules := range rules {
		constraints := convutils.MapMe(fldRules)

		defaultFrom := convutils.MapMe(constraints[PolicyDefaultFrom])["field"]
		if defaultFrom != nil {
			if fldName == defaultFrom.(string) {
				return fmt.Errorf("%s: defaultFrom.field must be a different field", fldName)
			}

			g.AddDependency(fldName, defaultFrom.(string))
			continue
		}

		// for itemized fields, we need to build the order of their attributes
		attributes := constraints[itemized.PolicyAttributes]
		if attributes != nil {
			err := determineDefaultsFromOrder(convutils.MapMe(attributes), fldName, result)
			if err != nil {
				return err
			}
		}
	}

	orderedNodes, err := g.FindOrder()
	if err != nil {
		return err
	}

	(*result)[key] = orderedNodes

	return nil
}
