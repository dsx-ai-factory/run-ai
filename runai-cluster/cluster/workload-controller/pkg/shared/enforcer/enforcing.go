package enforcer

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/stretchr/testify/assert"

	jsonq "github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/rules"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/third_party/validation"
)

// validate all values are aligned with the rules stated by
// (a) the admin policy (b) run:ai global rules
func (self *MergeStack) enforce(ctx context.Context) error {

	//
	//    validate against the admin policy - this is necessary only for workloads
	//
	if self.policyRules != nil {
		err := self.enforceRules(ctx, self.policyRules.JumpToTop())
		if err != nil {
			return err
		}
	}

	//
	//    validate against run:ai rules
	//
	err := self.enforceRules(ctx, self.runaiRules)
	if err != nil {
		return err
	}

	return nil
}

// validate a set of rules against the merged set of values
func (self *MergeStack) enforceRules(ctx context.Context, rulesToEnforce *jsonq.JSONQ) error {

	//
	//    get the merged values from the lowest level of the stack
	//
	mergedValues := self.MergedValues().JumpToTop()

	//
	//    for each field which has rules...
	//
	for ruleFldName := range utils.MapMe(rulesToEnforce.Get()) {

		//
		//   get the merged value of the field
		//
		mergedFldDef := mergedValues.JumpTo(ruleFldName)

		//
		//   for values, apply primitive rules for each of the members
		//
		mergedFldMembers := mergedFldDef.JumpTo(common.Values)
		if mergedFldMembers.Get() != nil {
			members := utils.MapMe(mergedFldMembers.Get())
			prefix := utils.DotUs(ruleFldName, common.Values)

			for memberFldName := range members {
				memberFullName := utils.DotUs(prefix, memberFldName)

				thisMemberDef := mergedFldMembers.JumpTo(memberFldName)
				thisMemberRules := rulesToEnforce.JumpTo(utils.DotUs(memberFullName, common.Rules))
				defaultValue := thisMemberDef.From(common.Value).Get()

				err := self.enforcePrimitiveRules(ctx, memberFullName, thisMemberRules, thisMemberDef, defaultValue)
				if err != nil {
					return err
				}
			}
			continue
		}

		//
		//   get the rules of the field
		//
		ruleFldDef := rulesToEnforce.JumpTo(ruleFldName)
		ruleFldRules := ruleFldDef.JumpTo(common.Rules)

		defaultsItem := utils.MapMe(ruleFldDef.JumpTo(utils.DotUs(common.Items, common.ItemDefaults, common.Value)).Get())

		//
		//   if this is a primitive value, we need to validate its scalar value against the rules
		//
		mergedFldItems := mergedFldDef.JumpTo(common.Items)
		if mergedFldItems.Get() == nil {
			if ruleFldRules.Get() != nil {
				prefix := utils.DotUs(ruleFldName, common.Value)
				defaultValue := ruleFldDef.JumpToTop().From(common.Value).Get()
				err := self.enforcePrimitiveRules(ctx, prefix, ruleFldRules, mergedFldDef, defaultValue)
				if err != nil {
					return err
				}
			}
			continue
		}

		//
		//    if the itemized field is required as a whole, make sure there is at least one item
		//
		items := utils.MapMe(mergedFldItems.Get())
		if len(items) == 0 {
			if ruleFldRules.JumpToTop().WhereEqual(rules.Required, true).Count() > 0 && self.levelDesc[0] != "" {
				return fmt.Errorf("value is not provided required: %s.%s", ruleFldName, common.Value)
			}
		}

		//
		//   for an itemized field, we have to validate both item-related rules (adding/removing items) ]
		//   and also to validate the value of each specific item
		//
		for itemName := range items {

			itemFullName := fmt.Sprintf("%s.%s", ruleFldName, itemName)

			//
			//    get item's definitions and rules
			//
			thisItemDef := mergedFldItems.JumpTo(itemName)
			thisItemRules := thisItemDef.JumpTo(common.Rules)

			//
			//   if there is a keyRule, it means we also have to check the name of the item
			//
			keyRules := ruleFldDef.JumpTo(common.KeyRules)
			if keyRules.Get() != nil {
				keyValue, _ := utils.NewJsonQ(map[string]interface{}{
					common.Value: itemName,
				})
				err := self.enforcePrimitiveRules(ctx, itemFullName, keyRules, keyValue, nil)
				if err != nil {
					return err
				}
			}

			//
			//    an item can be deleted in the base template then added in the topmost template, or it
			//    can be the other way around. to figure out what was the effective action, we can compare
			//    the level in which it was deleted to the level it has been added. the smaller one is the last
			//
			deletionLevel := thisItemDef.JumpToTop().Find(common.DeletedLevel)
			valueFromLevel := thisItemDef.JumpToTop().Find(common.ValueFromLevel)

			if deletionLevel != nil && valueFromLevel != nil {
				if deletionLevel.(int) < valueFromLevel.(int) {
					delete(utils.MapMe(thisItemDef.JumpToTop().Get()), common.Value)
					delete(utils.MapMe(thisItemDef.JumpToTop().Get()), common.ValueFromLevel)
				} else {
					delete(utils.MapMe(thisItemDef.JumpToTop().Get()), common.Deleted)
					delete(utils.MapMe(thisItemDef.JumpToTop().Get()), common.DeletedLevel)
				}
			}

			//
			//    if the value has been deleted, make sure the rules allow to remove it. if yes,
			//    nothing further to check cause deleted value = no value...
			//
			if thisItemDef.JumpToTop().WhereEqual(common.Deleted, true).Count() > 0 {
				if thisItemRules.WhereEqual(rules.CanRemove, false).Count() > 0 {
					return fmt.Errorf("the administrator mandated that %s cannot be removed", itemFullName)
				}
				continue
			}

			//
			//   check if this item is added by the workload (or found in policy, in this
			//   case the workload attempts to override it)
			//
			var itemInPolicy interface{}
			if self.policyRules != nil {
				policyItems := self.policyRules.JumpTo(utils.DotUs(ruleFldName, common.Items))
				itemInPolicy = policyItems.JumpToTop().Find(itemName)
			}

			if itemInPolicy == nil {
				// item is not included in the policy items of this field, which indited tht workload
				// attempts to add an item. we need to check if canAdd is permitted by the rules
				if ruleFldRules.JumpToTop().WhereEqual(rules.CanAdd, false).Count() > 0 {
					return fmt.Errorf("%s: the administrator prohibited adding to %s", itemName, ruleFldName)
				}
			} else {
				// overriding item - need to check if canEdit is permitted by the rules
				// in this case we have to check the value has not changed
				if thisItemRules.WhereEqual(rules.CanEdit, false).Count() > 0 {
					itemDefInPolicy := utils.MapMe(itemInPolicy)
					err := self.validateValueHasNotChanged(itemFullName, thisItemDef, itemDefInPolicy[common.Value])
					if err != nil {
						return err
					}
				}
			}

			var err error
			itemValueFromLevel := thisItemDef.JumpToTop().Find(common.ValueFromLevel)
			if itemValueFromLevel == nil {
				continue
			}

			//
			//    there are two cases with items. items which contain primitive value (like environment variable
			//    whose value is a simple string) or a "complex" item, containing multiple members, like volume
			//    or port which can contain host port, external port, etc...
			//
			switch thisItemDef.JumpToTop().Find(common.Value).(type) {
			case map[string]interface{}:
				//
				//    for complex items we have to run validation for each member specifically
				//
				itemValues := thisItemDef.JumpTo(common.Value)
				err = self.enforceItemRules(ctx, itemFullName, ruleFldDef.JumpTo(common.ItemRules), itemValues, itemValueFromLevel.(int), defaultsItem)
			default:
				//
				//    for simple items we just have to check their values against their "primitive" rules
				//
				itemValueDef, _ := utils.NewJsonQ(map[string]interface{}{
					common.Value:          thisItemDef.JumpToTop().Find(common.Value),
					common.ValueFromLevel: itemValueFromLevel,
				})
				err = self.enforcePrimitiveRules(ctx, itemFullName, ruleFldDef.JumpTo(common.ItemRules), itemValueDef, nil)
			}
			if err != nil {
				return err
			}

		}

	}
	return nil
}

// get an item with multiple members (like volume) and apply primitive rules for each of its members
func (self *MergeStack) enforceItemRules(ctx context.Context, itemFullName string,
	itemRules *jsonq.JSONQ, memberValues *jsonq.JSONQ, valueFromLevel int, defaultsItem map[string]interface{}) error {

	//
	//    for each member...
	//
	for memberName := range utils.MapMe(itemRules.Get()) {

		memberRules := itemRules.JumpTo(memberName)
		memberFullName := fmt.Sprintf("%s.%s", itemFullName, memberName)

		//
		//   construct the map containing the value and its origin
		//
		memberValueDef, _ := utils.NewJsonQ(map[string]interface{}{
			common.Value:          memberValues.JumpToTop().Find(memberName),
			common.ValueFromLevel: valueFromLevel,
		})

		//
		//   run the primitive rules for this member
		//
		err := self.enforcePrimitiveRules(ctx, memberFullName, memberRules, memberValueDef, defaultsItem[memberName])
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *MergeStack) enforcePrimitiveRules(_ context.Context, prefix string, setOfRules *jsonq.JSONQ,
	mergedFldDef *jsonq.JSONQ, defaultValue interface{}) error {

	//
	//   no rules - nothing to do
	//
	if setOfRules.Get() == nil {
		return nil
	}
	value := mergedFldDef.JumpToTop().Find(common.Value)

	//
	//   the first thing we check is if value has been provided for this field
	//   If not, and it is a mandatory field, reject the request. Otherwise return,
	//   cause with no value there is nothing else to validate
	//
	isNil, _ := isOk(validation.Validate(value, validation.Nil))
	if isNil {
		//
		//   note an exception of validating policy (=no workload yet) and workload used as templates:
		//   in those case we allow required values to be missing. only for workloads used for Submit all the values
		//   which are required are validated.
		//
		if setOfRules.JumpToTop().WhereEqual(rules.Required, true).Count() > 0 && self.usedForSubmit {
			return fmt.Errorf("missing value for required: %s", prefix)
		}
		return nil
	}

	//
	//   validate canEdit rule
	//
	if setOfRules.JumpToTop().WhereEqual(rules.CanEdit, false).Count() > 0 {
		err := self.validateValueHasNotChanged(prefix, mergedFldDef, defaultValue)
		if err != nil {
			return err
		}
	}

	var err error
	var ok = true

	rulesMap := utils.MapMe(setOfRules.JumpToTop().Get())

	for ruleName, ruleDef := range rulesMap {
		switch ruleName {

		case rules.Required:
			continue // already handled
		case rules.CanEdit:
			continue // already handled

		case rules.Max:
			ok, err = isOk(validation.Validate(value, validation.Max(ruleDef)))
		case rules.Min:
			ok, err = isOk(validation.Validate(value, validation.Min(ruleDef)))

		case rules.Step:
			ok, err = isOk(validation.Validate(value, Step(rulesMap)))

		case rules.Options:
			optionList := jsonq.New().FromInterface(ruleDef)
			foundOption := utils.MapMe(optionList.WhereEqual(common.OptionsValue, value).First())
			if len(foundOption) == 0 || utils.BoolMe(foundOption[common.OptionsDisabled], false) {
				err = fmt.Errorf("value must be one of: %s, received: %s", optionsToString(optionList), value)
				ok = false
			}

		case workload_fields.Format:
			ok, err = enforceFormat(ruleDef.(string), value)

		default:
			fmt.Printf("Unsupported rule %s\n", ruleName)
			continue
		}

		if !ok {
			return addPrefixToErrorMessage(err, prefix)
		}
	}
	return err
}

// figure out if a workload attempts to change the default value provided in the policy of the administrator
func (self *MergeStack) validateValueHasNotChanged(prefix string, mergedFldDef *jsonq.JSONQ, defaultValue interface{}) error {
	//
	//   check if researcher provided any value for the parameter. if not, then the
	//   policy is honored and no need to compare the value itself
	//
	policyLevel := self.PolicyLevel()
	valueFromLevel, err := utils.JsonNumberToInt(mergedFldDef.JumpToTop().Find(common.ValueFromLevel))
	if err != nil {
		return err
	}
	if valueFromLevel == nil || *valueFromLevel == policyLevel {
		return nil
	}

	//
	//   researcher specified value for this field, we need to make sure this is the same
	//   value as in the policy
	//
	givenValue := mergedFldDef.JumpToTop().Find(common.Value)

	if givenValue != nil && defaultValue != nil {
		isSameValue := assert.ObjectsAreEqualValues(givenValue, defaultValue)
		if isSameValue {
			return nil
		}
	}

	return fmt.Errorf("the administrator prohibited overriding the value of %s", prefix)
}

// add prefix to an error message
func addPrefixToErrorMessage(err error, prefix string) error {
	return fmt.Errorf("%s: %v", prefix, err.Error())
}

func isOk(err error) (bool, error) {
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

// convert an option list of format: name;value;disabled... to a string of values
// take only non-disabled values
func optionsToString(options *jsonq.JSONQ) string {
	optionValues := options.JumpToTop()
	result := make([]string, 0, options.Count())
	for i := 1; i <= optionValues.Count(); i++ {
		curOption := utils.MapMe(optionValues.Nth(i))
		if !utils.BoolMe(curOption[common.OptionsDisabled], false) {
			result = append(result, fmt.Sprintf("%v", curOption[common.OptionsValue]))
		}
	}
	return strings.Join(result, ", ")
}

// StepRule is the implementation of the step rule, which is not provided by the validation third party
type StepRule struct {
	minVal string
	step   string
}

// when working with float we cannot depend on equality, but on an allowed
// difference. for example 1.0000001 and 0.999999 are considered equal to 1
const maxAllowedDiff = float64(0.0001)

func Step(fldRules interface{}) StepRule {

	fldRulesMap := utils.MapMe(fldRules)

	result := StepRule{
		step:   fldRulesMap[rules.Step].(string),
		minVal: "0.0",
	}
	if fldRulesMap[rules.Min] != nil {
		result.minVal = fldRulesMap[rules.Min].(string)
	}

	return result
}

func (r StepRule) Validate(value interface{}) error {

	minValue, err := validation.ToFloat(r.minVal)
	if err != nil {
		return err
	}

	step, err := validation.ToFloat(r.step)
	if err != nil {
		return err
	}

	numericValue, err := validation.ToFloat(value)
	if err != nil {
		return err
	}

	steps := (numericValue - minValue) / step
	diff := math.Abs(steps - math.Round(steps))
	if diff > maxAllowedDiff {
		return fmt.Errorf("value must follow %v steps from %v (%.2f, %.2f, etc.)",
			r.step, minValue, minValue, minValue+step)
	}

	return nil
}
