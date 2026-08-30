package enforcer

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"

	jsonq "github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/rules"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/third_party/validation"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	LevelWorkload = "workload"
	LevelPolicy   = "policy"
)

//	  a merge stack consists of:
//			2 to 4 levelElements:
//	        * Optionally, the Master Spec
//	        * The Workload Spec
//			* Optionally, base Workload
//			* The Policy. After merging, this level contains the merged values
type MergeStack struct {
	levelElement  []*jsonq.JSONQ
	levelDesc     []string     // description of each level
	policyRules   *jsonq.JSONQ // we keep the original policy because the merge overrides the values
	runaiRules    *jsonq.JSONQ // additional set of rules, imposed by run:ai
	usedForSubmit bool         // is this stack holds a workload with usage=Submit
}

//	  initialize a merge stack
//	     workloadQ: the "topmost" workload
//			name: the name of the workload
//	     usedForSubmit: indication if we're building the stack for a usedForSubmit job
//	  the caller will have to append the levels of base-workload (if applicable) and the policy
func NewMergeStack(workloadQ *jsonq.JSONQ, workloadName string, runaiRules *jsonq.JSONQ, usedForSubmit bool) *MergeStack {
	result := MergeStack{
		levelElement:  make([]*jsonq.JSONQ, 0, 4),
		levelDesc:     make([]string, 0, 4),
		runaiRules:    runaiRules,
		policyRules:   nil, // will be set when AddLevel below adds the level containing the policy
		usedForSubmit: usedForSubmit,
	}
	result.AddLevel(workloadQ, LevelWorkload, workloadName)
	return &result
}

// AddLevel - after initializing the mergeStack, AddLevel adds the base workload (optionally) and the policy
// note that when we validate the policy itself, we add an empty workload with empty name.
func (self *MergeStack) AddLevel(element *jsonq.JSONQ, resourceType string, name string) {
	self.levelElement = append(self.levelElement, element)
	if name != "" {
		self.levelDesc = append(self.levelDesc, resourceType+":"+name)
		if strings.HasPrefix(resourceType, LevelPolicy) {
			//   make a copy of the policy, before the merge overrides it with values from the workload(s)
			self.policyRules, _ = utils.NewJsonQ(runtime.DeepCopyJSONValue(element.Get()))
		}
	} else {
		self.levelDesc = append(self.levelDesc, "")
	}
}

type MergeAndValidateOptions struct {
	DisableSignining bool // this option is for testing/debugging
	DisableEnforce   bool // this option is for testing/debugging
}

// MergeAndValidate merge the workload(s) and the policy values and validate them against the policy rules
func (self *MergeStack) MergeAndValidate(ctx context.Context, options MergeAndValidateOptions) error {

	//
	//   sign the fields of each level with the index of that level in the stack. after we'll merge everything,
	//   this will allow us to know, for each field, if we received it from the policy or from a workload
	//
	if !options.DisableSignining {
		self.signEachLevel(ctx)
	}

	//
	//   merge all the levels to form a complete set of values
	//
	err := self.merge(ctx)
	if err != nil {
		return err
	}

	//
	//   apply default of itemized fields from DEFAULTS item, if provided
	//
	self.applyItemDefaults()

	//
	//    enforce the rules on the set of values
	//
	if !options.DisableEnforce {
		err = self.enforce(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateSyntaxOfWorkload performs syntactic validations for workloads. we run here validation that we cannot define using open3api:
// (#1) no rules in a workload
//
//		validate that a workload does not contain any rules. we need this cause we use the same
//		structure for both, so we have to check it programmatically here
//		the json structure we're expecting is:
//	     field:
//	       rules: <-- forbidden
//	           ...
//	       itemRules: <-- forbidden
//	           ...
//	       items:
//	           key:
//	               rules:  <-- forbidden
//	                   ...
//
// (#2) either "deleted" or "value" for an item, not both
func ValidateSyntaxOfWorkload(prefix string, templateFlds *jsonq.JSONQ) error {

	for fldName := range utils.MapMe(templateFlds.Get()) {

		fldDef := templateFlds.JumpTo(fldName)

		if fldDef.JumpToTop().Find(common.Rules) != nil {
			return fmt.Errorf("%s.%s: Element '%s' is forbidden", prefix, fldName, common.Rules)
		}
		if fldDef.JumpToTop().Find(common.ItemRules) != nil {
			return fmt.Errorf("%s.%s: Element '%s' is forbidden", prefix, fldName, common.ItemRules)
		}
		if fldDef.JumpToTop().Find(common.KeyRules) != nil {
			return fmt.Errorf("%s.%s: Element '%s' is forbidden", prefix, fldName, common.KeyRules)
		}
		if fldDef.JumpToTop().Find(common.Value) != nil && fldDef.JumpToTop().WhereEqual(common.Deleted, true).Count() > 0 {
			return fmt.Errorf("%s.%s: '%s' and '%s' are mutual exclusive", prefix, fldName, common.Value, common.Deleted)
		}

		//
		//   items of itemized fields can have their own rules, so make sure to check them as well
		//
		items := fldDef.JumpTo(common.Items)
		if items.Get() != nil {
			err := ValidateSyntaxOfWorkload(utils.DotUs(prefix, fldName), items)
			if err != nil {
				return err
			}
		}

		//
		//   values containing member fields, each can have their own rules, so make sure to check them as well
		//
		values := fldDef.JumpTo(common.Values)
		if values.Get() != nil {
			err := ValidateSyntaxOfWorkload(utils.DotUs(prefix, fldName), values)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateSyntaxOfPolicy performs syntactic validations for policy. we run here validations that we cannot define using open3api:
// (#1) item cannot be "deleted" in policy. only workload can delete an item from the
//
//	baseTemplate or from the policy defaults.
//
// (#2) if an item has a rule, then this item also requires a value
// (#3) min/max rules in the policy must comply with min/max rules of run:ai
// (#4) canEdit cannot appear in itemRules. admin can only prevent editing of an entire item.
//
//	volumes:
//	   itemRules:
//	      sourcePath:
//	        rules:
//	          canEdit: false  // this is not ok, can't limit editing of sourcePath only
//	   item:
//	     "myvolume":
//	        rules:
//	          canEdit: false   // this is ok, cant edit source path and all other members
func ValidateSyntaxOfPolicy(prefix string, policyFlds *jsonq.JSONQ, runaiFlds *jsonq.JSONQ) error {

	for fldName := range utils.MapMe(policyFlds.Get()) {

		fldTag := utils.DotUs(prefix, fldName)

		policyFldDef := policyFlds.JumpTo(fldName) // definitions of the field in the policy
		runaiFldDef := runaiFlds.JumpTo(fldName)   // definitions of the field in the run:ai rules

		policyFldRules := policyFldDef.JumpTo(common.Rules)
		runaiFldRules := runaiFldDef.JumpTo(common.Rules)

		//
		//    if rules contain min/max, need to validate them against run:ai min/max rules. for example,
		//    if run:ai rules state that max processes is 100 and the policy has rules allowing
		//    max processes to be 200, then need to tell the administrator that this rule cannot be applied
		//
		if policyFldRules.Get() != nil {
			err := validateMinMaxRules(fldTag, policyFldRules, runaiFldRules)
			if err != nil {
				return err
			}
		}

		//
		//   validation on the items of an itemized field and member of "values"
		//
		for _, clause := range []string{common.Items, common.Values} {
			policyFldItems := policyFldDef.JumpTo(clause)

			for itemName := range utils.MapMe(policyFldItems.Get()) {

				itemTag := utils.DotUs(prefix, fldName, itemName)
				policyItemDef := policyFldItems.JumpTo(itemName)

				//
				//    make sure no deletion of items in the policy
				//
				if policyItemDef.JumpToTop().WhereEqual(common.Deleted, true).Count() > 0 {
					return fmt.Errorf("%s: '%s' is not allowed in a policy", itemTag, common.Deleted)
				}

				//
				//    if there are rules for an item - there must be value for this item
				//
				if policyItemDef.JumpToTop().Find(common.Rules) != nil {
					if policyItemDef.JumpToTop().Find(common.Value) == nil {
						return fmt.Errorf("%s: '%s' is missing", itemTag, common.Value)
					}

					policyItemRules := policyItemDef.JumpTo(common.Rules)
					runaiItemRules := runaiFldDef.JumpTo(utils.DotUs(clause, itemName, common.Rules))

					err := validateMinMaxRules(itemTag, policyItemRules, runaiItemRules)
					if err != nil {
						return err
					}
				}
			}
		}

		//
		//    if itemRules contain min/max rules, need to validate them against run:ai rules.
		//    for example if policy says maxPort = 90000 but run:ai allows port to have maximum
		//    of 30000, than need to tell the administrator that its policy rule is not sensible
		//
		policyFldItemRules := policyFldDef.JumpTo(common.ItemRules)
		runaiFldItemRules := runaiFldDef.JumpTo(common.ItemRules)

		if policyFldItemRules.Get() != nil {

			if isComplexField(policyFldItemRules) {
				//
				//   for complex field, we validate each member. for example for volume
				//   there are item rules for sourcePath, targetPath, etc
				//
				for memberName := range utils.MapMe(policyFldItemRules.Get()) {

					memberItemRulesTag := utils.DotUs("", fldName, common.ItemRules, memberName)

					policyMemberItemRules := policyFldItemRules.JumpTo(memberName)
					runaiMemberItemRules := runaiFldItemRules.JumpTo(memberName)

					err := validateMinMaxRules(memberItemRulesTag, policyMemberItemRules, runaiMemberItemRules)
					if err != nil {
						return err
					}
				}
			} else {
				itemRulesTag := utils.DotUs(fldName, common.ItemRules)

				err := validateMinMaxRules(itemRulesTag, policyFldItemRules, runaiFldItemRules)
				if err != nil {
					return err
				}
			}

		}
	}
	return nil
}

// validate the policy min/max rules against the runai min/max rules
func validateMinMaxRules(prefix string, policyRules *jsonq.JSONQ, runaiRules *jsonq.JSONQ) error {
	var ok bool
	var err error

	//
	//   regardless of run:ai rules, if both min and max are provided - make sure min <= max
	//
	policyMin := policyRules.JumpToTop().Find(rules.Min)
	policyMax := policyRules.JumpToTop().Find(rules.Max)

	if policyMin != nil && policyMax != nil {
		ok, _ = isOk(validation.Validate(policyMin, validation.Max(policyMax)))
		if !ok {
			return fmt.Errorf("%s must be no greater than %s", utils.DotUs(prefix, rules.Min), utils.DotUs(prefix, rules.Max))
		}
	}

	runaiMin := runaiRules.JumpToTop().Find(rules.Min)
	runaiMax := runaiRules.JumpToTop().Find(rules.Max)

	if runaiMax != nil {
		allowedMax := validation.Max(runaiMax)

		ok, err = isOk(validation.Validate(policyMin, allowedMax))
		if !ok {
			return addPrefixToErrorMessage(err, utils.DotUs(prefix, rules.Min))
		}

		ok, err = isOk(validation.Validate(policyMax, allowedMax))
		if !ok {
			return addPrefixToErrorMessage(err, utils.DotUs(prefix, rules.Max))
		}
	}

	if runaiMin != nil {
		allowedMin := validation.Min(runaiMin)

		ok, err = isOk(validation.Validate(policyMin, allowedMin))
		if !ok {
			return addPrefixToErrorMessage(err, utils.DotUs(prefix, rules.Min))
		}

		ok, err = isOk(validation.Validate(policyMax, allowedMin))
		if !ok {
			return addPrefixToErrorMessage(err, utils.DotUs(prefix, rules.Max))
		}
	}

	return nil
}

// check itemRules to figure out if their field is complex (like volumes, ports, etc)
// or is a simple value (like arguments or environment variables whose value is a string)
func isComplexField(itemRules *jsonq.JSONQ) bool {
	for _, value := range utils.MapMe(itemRules.Get()) {
		switch value.(type) {
		case map[string]interface{}:
			return true
		default:
			return false
		}
	}
	return false
}

// this function apply defaults of itemized field from DEFAULTS item, if given
// for example:
//
//	pvcs:
//	  items:
//	    DEFAULTS:
//	      ...
//	    pvc-a:
//	      ... <-- apply defaults for missing fields from DEFAULTS
func (self *MergeStack) applyItemDefaults() {
	//
	//    get the merged values from the lowest level of the stack
	//
	mergedValues := self.MergedValues().JumpToTop()

	//
	//    for each field which has rules...
	//
	for fldName := range utils.MapMe(mergedValues.Get()) {

		//
		//   get the merged value of the field
		//
		mergedFldDef := mergedValues.JumpTo(fldName)

		//
		//   if this is a primitive value - skip it, we only need to handle itemized fields
		//
		items := mergedFldDef.JumpTo(common.Items)
		if items.Get() == nil {
			continue
		}
		itemsMap := utils.MapMe(items.Get())

		//
		//   check if there is a DEFAULTS item from which we can load defaults for the other items.
		//
		fldDefaults := items.JumpTo(common.ItemDefaults)
		if fldDefaults.Get() == nil {
			continue
		}

		for itemName := range itemsMap {
			if itemName == common.ItemDefaults {
				continue
			}

			thisItemDef := items.JumpTo(itemName)

			//
			//    there are two cases with items. items which contain primitive value (like environment variable
			//    whose value is a simple string) or a "complex" item, containing multiple members, like pvcs, s3, etc.
			//
			switch fldDefaults.JumpToTop().Find(common.Value).(type) {
			case map[string]interface{}:
				defaultMembers := fldDefaults.JumpTo(common.Value)
				thisItemMembers := thisItemDef.JumpTo(common.Value)

				//
				//   it is possible that the item does not even have a "value' field, and in this case we
				//   will copy all the values from the defaults
				//
				if thisItemMembers.Get() == nil {
					thisItemMap := utils.MapMe(thisItemDef.JumpToTop().Get())
					for k, v := range utils.MapMe(fldDefaults.JumpToTop().Get()) {
						thisItemMap[k] = v
					}
					continue
				}

				//
				//    for each member that does not have a value - check if there is a default value and apply it
				//
				for memberName := range utils.MapMe(defaultMembers.Get()) {
					memberValue := thisItemMembers.JumpToTop().Find(memberName)
					if memberValue != nil {
						continue // member has a value so no need to apply defaults
					}

					defaultValue := defaultMembers.JumpToTop().Find(memberName)
					if defaultValue != nil {
						// apply the value from the defaults.
						itemValuesMap := utils.MapMe(thisItemMembers.JumpToTop().Get())
						itemValuesMap[memberName] = defaultValue
					}
				}
			default:
				thisItemValue := thisItemDef.JumpToTop().Find(common.Value)
				if thisItemValue != nil {
					continue // field has a value so no need to apply default
				}

				defaultValue := fldDefaults.JumpToTop().Find(common.Value)
				if defaultValue != nil {
					thisItemMap := utils.MapMe(thisItemDef.JumpToTop().Get())
					for k, v := range utils.MapMe(fldDefaults.JumpToTop().Get()) {
						thisItemMap[k] = v
					}
					continue
				}
			}
		}

		//
		//   remove the DEFAULTS item from the merged values, once its values have been used we do not need it
		//   and leaving it will interfere with the enforcer and builders
		//
		delete(utils.MapMe(items.JumpToTop().Get()), common.ItemDefaults)
	}
}
