/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/modern-go/reflect2"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/validation"
	"github.com/stretchr/testify/assert"

	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"github.com/PaesslerAG/jsonpath"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const jsonPathUnknownFieldError = "unknown key"

// note! this uuid pattern only test basic structure, do NOT use it for checking VALIDITY of uuid!!
var uuidPattern = "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[0-9a-f]{12}$"
var uuidRegex = regexp.MustCompile(uuidPattern)

// regular expression for removing instances like "[27736]" from the string
var removeIndexPattern = `\[\d+\]`
var removeIndexRegexp = regexp.MustCompile(removeIndexPattern)

const (
	DefaultForRequired = false
	DefaultForCanEdit  = true
	DefaultForCanAdd   = true

	ExclusionFlag           = "exclude"
	DefaultForExclusionFlag = false
)

type ValidateComplianceOptions struct {
	IgnoreRequiredFields bool

	// Enforce checking that a default value is set when canEdit=false and the field is not required.
	// This applies only when validating the effective policy (not the given policy).
	CanEditWithNoDefaultValidation bool

	//
	//  policy compliance can be invoked in two different ways:
	//  1) when we want to validate the values of a workload against a policy, the function is invoked with defaults
	//     taken from the policy, and spec taken from the workload being validated.
	//  2) when we want to validate the defaults of a policy itself against the rules of this policy (we do it when
	//     adding a policy and want to check that the defaults themselves are sensible), in this case PolicySelfValidation
	//     option should be true
	PolicySelfValidation bool
}

var defaultOptions = &ValidateComplianceOptions{}

// ValidateCompliance validates defaults of an applied policy against the effective rules of this policy
// it adds any error that it finds to the output array complianceErrors
// for itemized fields the function is invoked recursively. in this case, the fieldPrefix contains the jsonpath
// prefix of the parent field.
func ValidateCompliance(fieldPrefix string, rulesToCheck map[string]interface{}, defaults map[string]interface{},
	valuesAsMap map[string]interface{}, defaultFromOrder DefaultFromOrder,
	fieldsWithCantEditRule *[]string, options *ValidateComplianceOptions, complianceErrors *ComplianceErrors) {

	if options == nil {
		options = defaultOptions
	}

	if defaultFromOrder == nil {
		defaultFromOrder = make(DefaultFromOrder)
	}

	if ok := validateItemizedFields(valuesAsMap, "", complianceErrors, *options); !ok {
		return
	}

	rulesMap := GetRulesForCompliance(rulesToCheck)

	if !options.PolicySelfValidation {
		// for itemized fields we need the field name without the added index. e.g. for ports[2] we need ports
		indexesStripped := removeIndexRegexp.ReplaceAllString(fieldPrefix, "")

		order := defaultFromOrder[indexesStripped]
		for _, fieldName := range order {
			fieldRules := rulesMap[fieldName]
			constraintsMap := convutils.MapMe(fieldRules)
			defaultsFromAsMap := convutils.MapMe(constraintsMap[PolicyDefaultFrom])
			if len(defaultsFromAsMap) == 0 {
				continue
			}

			_, fieldValueFound := getFieldValue(valuesAsMap, fieldName)
			if fieldValueFound {
				continue
			}

			var defaultFrom openapi.DefaultFromRule
			json.Unmarshal(gojsonq.ToBytes(defaultsFromAsMap), &defaultFrom)

			sourceValue, sourceValueFound := getFieldValue(valuesAsMap, *defaultFrom.Field.Get())
			if !sourceValueFound {
				sourceValue, sourceValueFound = getFieldValue(defaults, *defaultFrom.Field.Get())
			}
			if sourceValueFound {
				err := setFieldValue(valuesAsMap, fieldName, sourceValue, defaultFrom.Factor.Get())
				if err != nil {
					logging.Errorf("%s", err.Error())
				}
			}
		}
	}

	for fieldName, fieldRules := range rulesMap {
		constraintsMap := convutils.MapMe(fieldRules)

		policyFieldName := composePolicyFieldName(fieldPrefix, fieldName)

		//
		//   sanity validation of the min/max rule
		//
		minPolicy := constraintsMap[string(PolicyMin)]
		maxPolicy := constraintsMap[string(PolicyMax)]
		if minPolicy != nil && maxPolicy != nil {
			_, err := IsOk(validation.Validate(minPolicy, validation.Max(maxPolicy)))
			if err != nil {
				complianceErrors.Add(&PolicyMax, "max value must be no less than min value",
					convutils.DotUs(fieldPrefix, fieldName), policyFieldName)
			}
		}

		//
		// we use the jsonpath expression $.fff to load the value of field 'fff' from the spec
		//
		value, err := jsonpath.Get(fmt.Sprintf("$.%s", fieldName), valuesAsMap)
		if err != nil && !strings.Contains(err.Error(), jsonPathUnknownFieldError) {
			logging.Errorf("failed to load the value of %s", fieldName)
			continue
		}

		if value == nil {
			//
			// if we arrived here it means that the field does not have a value in the spec, so we need to validate
			// a limited set of rules. the obvious rule to check -> is this field is required
			//
			required := convutils.BoolMe(constraintsMap[string(PolicyRequired)], DefaultForRequired)
			if required {
				if !options.PolicySelfValidation {
					//
					//  a required field must have either a value or a default
					//
					defaultValue, err := jsonpath.Get(fmt.Sprintf("$.%s", fieldName), defaults)
					if err != nil || defaultValue == nil {
						if !options.IgnoreRequiredFields {
							complianceErrors.Add(&PolicyRequired, "value must be provided for required fields", convutils.DotUs(fieldPrefix, fieldName), policyFieldName)
						}
						continue
					}
				}
			}

			//
			//   when checking the defaults of the policy itself: withCanEdit == false must have a
			//   value in its spec, to serve as default of the policy
			//
			if options.PolicySelfValidation && options.CanEditWithNoDefaultValidation {
				if !convutils.BoolMe(constraintsMap[string(PolicyCanEdit)], DefaultForCanEdit) {
					complianceErrors.Add(&PolicyCanEdit, "default value must be provided for policy fields having canEdit=false",
						convutils.DotUs(fieldPrefix, fieldName), policyFieldName)
					continue
				}

				attributes, hasAttributes := constraintsMap[itemized.PolicyAttributes]

				if hasAttributes {
					if _, isItemized := itemized.GetItemizedFieldInfo(fieldName); !isItemized {
						logging.Errorf("attribures are not supported for field %s", fieldName)
					} else {
						//  itemized fields attributes are assumed to be flat fields, so we just send the entire valuesAndMap to the
						//  validateAttributesRules function. assumption is that itemized sections are flat
						//
						validateAttributesRules(fieldName, convutils.MapMe(attributes), defaults, valuesAsMap, defaultFromOrder, complianceErrors, *options)
					}
				}
			}

			//  itemized fields are assumed to be flat, so we just send the entire valuesAndMap to the
			//  validateInstancesRules function. assumption is that itemized sections are flat
			//
			if itemRules, hasItemRules := constraintsMap[itemized.PolicyInstances]; hasItemRules {
				validateInstancesRules(fieldName, convutils.MapMe(itemRules), defaults, valuesAsMap, complianceErrors, *options)
			}

			continue
		}

		for ruleKind, ruleDetails := range constraintsMap {
			var err error

			switch ruleKind {
			case string(PolicyRequired):
				continue // there is a value, so required is satisfied
			case string(PolicyMin):
				_, err = IsOk(validation.Validate(value, validation.Min(ruleDetails)))
			case string(PolicyMax):
				_, err = IsOk(validation.Validate(value, validation.Max(ruleDetails)))
			case string(PolicyStep):
				_, err = IsOk(validation.Validate(value, Step(constraintsMap)))
			case string(PolicyOptions):
				err = validateOptionsRule(ruleDetails, value)
			case string(PolicyCanEdit):
				canEdit := ruleDetails.(bool)
				if !canEdit {
					if value != nil {
						defaultValue, defaultError := jsonpath.Get(fmt.Sprintf("$.%s", fieldName), defaults)
						if defaultValue != nil && defaultError == nil {
							isSameValue := assert.ObjectsAreEqualValues(value, defaultValue)
							if !isSameValue {
								err = fmt.Errorf("the administrator prohibited changing the value of this field")
							}
						}
					}
					*fieldsWithCantEditRule = append(*fieldsWithCantEditRule, fieldName)
				}
			case string(PolicyBlocked):
				blocked := ruleDetails.(bool)
				if blocked && !reflect2.IsNil(value) {
					if options.PolicySelfValidation {
						err = fmt.Errorf("default value should not be provided for a field with 'blocked' rule")
					} else {
						err = fmt.Errorf("value should not be provided for this field, it has been blocked by the administrator")
					}
				}
			case itemized.PolicyInstances:
				validateInstancesRules(fieldName, convutils.MapMe(ruleDetails), defaults, valuesAsMap, complianceErrors, *options)
			case itemized.PolicyAttributes:
				if _, isItemized := itemized.GetItemizedFieldInfo(fieldName); !isItemized {
					err = fmt.Errorf("attribures are not supported for field %s", fieldName)
				} else {
					//  itemized fields attributes are assumed to be flat fields, so we just send the entire valuesAndMap to the
					//  validateAttributesRules function. assumption is that itemized sections are flat
					//
					validateAttributesRules(fieldName, convutils.MapMe(ruleDetails), defaults, valuesAsMap, defaultFromOrder, complianceErrors, *options)
				}
			case PolicyDefaultFrom:
				continue // already taken care of
			case PolicySourceOfRule:
				continue // nothing to validate
			default:
				logging.Errorf("unsupported policy %s for field %s%s", ruleKind, fieldPrefix, fieldName)
				continue
			}
			if err != nil {
				complianceErrors.Add(convutils.Ref(workloadsApi.PolicyRuleEnum(ruleKind)),
					strings.TrimSuffix(err.Error(), "\n"),
					convutils.DotUs(fieldPrefix, fieldName), policyFieldName)
			}
		}
	}
}

// ValidateGivenPolicyDefaults checks the defaults of the given policy. the only validation we need is the
// validation of the itemized fields. we do not need to check rules, we check rules against the effective
// not against the given
func ValidateGivenPolicyDefaults(defaults map[string]interface{}, complianceErrors *ComplianceErrors) {
	validateItemizedFields(defaults, "", complianceErrors, ValidateComplianceOptions{
		PolicySelfValidation: true,
	})
}

func IsOk(err error) (bool, error) {
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

// function to pretty print the options from array [ "ubuntu", "linux" ] to string "ubuntu, linux"
// and return them as a set
func optionsInfo(options *gojsonq.JSONQ) (string, datastructs.Set[string]) {
	optionValues := options.JumpToTop()
	result := make([]string, 0, options.Count())
	for i := 1; i <= optionValues.Count(); i++ {
		curOption := convutils.MapMe(optionValues.Nth(i))
		result = append(result, fmt.Sprintf("%v", curOption["value"]))
	}
	return strings.Join(result, ", "), datastructs.NewSet(result...)
}

// validateItemizedFields validates item fields syntactically.
// 1. when checking the defaults of the policy itself: deletion flag iot allowed. it is allowed only for workloads.
// 2. validate that the key attribute is provided for all the items.
// 3. validate that there are no duplicate keys (two environment variables with the same name).
func validateItemizedFields(value map[string]interface{}, section string, complianceErrors *ComplianceErrors, options ValidateComplianceOptions) bool {
	for fieldName, fieldValue := range value {
		fullFieldName := convutils.DotUs(section, fieldName)

		if itemized.NestedPolicySections.Has(fullFieldName) {
			isOk := validateItemizedFields(convutils.MapMe(value[fieldName]), fullFieldName, complianceErrors, options)
			if !isOk {
				return false
			}
		}

		itemizedFieldInfo, _ := itemized.GetItemizedFieldInfo(fullFieldName)

		keys := datastructs.NewSet[string]()

		switch realType := fieldValue.(type) {
		case []interface{}:
			// in workloads the items are provided as an array
			for _, instance := range realType {
				attributes := convutils.MapMe(instance)

				isOk := validateItemizedInstance(attributes, fullFieldName, &itemizedFieldInfo, &keys, complianceErrors, options)
				if !isOk {
					return false
				}
			}
		case map[string]interface{}:
			// in policy, we need to check both attributes and instances, and we need
			// to check them both, but for attributes key validation is not needed
			defaultAttributes := convutils.MapMe(realType[itemized.PolicyAttributes])
			defaultInstances := convutils.ArrayMe(realType[itemized.PolicyInstances])

			isOk := validateItemizedInstance(defaultAttributes, fullFieldName, nil, nil, complianceErrors, options)
			if !isOk {
				return false
			}

			for _, instance := range defaultInstances {
				attributes := convutils.MapMe(instance)

				isOk := validateItemizedInstance(attributes, fullFieldName, &itemizedFieldInfo, &keys, complianceErrors, options)
				if !isOk {
					return false
				}
			}
		}
	}

	return true
}

func validateItemizedInstance(attributes map[string]interface{},
	fieldName string, itemizedFieldInfo *itemized.ItemizedFieldsInfo, keys *datastructs.Set[string],
	complianceErrors *ComplianceErrors, options ValidateComplianceOptions) bool {

	//  in policy validation, we have to add .instances to the name of the field so the user will be able to find them
	errFieldName := fieldName
	if options.PolicySelfValidation {
		errFieldName += itemized.InstancesSuffix
	}

	if options.PolicySelfValidation && convutils.BoolMe(attributes[ExclusionFlag], DefaultForExclusionFlag) {
		complianceErrors.Add(nil, fmt.Sprintf("'%s' is not allowed in policy defaults", ExclusionFlag),
			errFieldName, fieldName+itemized.InstancesSuffix)
		return false
	}

	if itemizedFieldInfo != nil && itemizedFieldInfo.KeyAttribute != "" {
		keyAttribute := itemizedFieldInfo.KeyAttribute

		//
		// in workload -> keys can be optional, otherwise their value must be provided
		//
		if !itemizedFieldInfo.KeyIsOptional {
			if attributes[keyAttribute] == nil || fmt.Sprintf("%v", attributes[keyAttribute]) == "" {
				complianceErrors.Add(nil, fmt.Sprintf("attribute '%s' must be provided to all instances", keyAttribute),
					errFieldName, fieldName+itemized.InstancesSuffix)
				return false
			}
		}

		keyValue, hasKey := attributes[keyAttribute]
		if hasKey {
			keyAsString := fmt.Sprintf("%v", keyValue)
			if keys.Has(keyAsString) {
				complianceErrors.Add(nil, fmt.Sprintf("duplicate %s '%s'", keyAttribute, keyAsString),
					fmt.Sprintf("%s[%s]", errFieldName, keyValue), fieldName+itemized.InstancesSuffix)
				return false
			}

			keys.Add(keyAsString)
		}
	}

	return true
}

// validateInstancesRulesF validates the itemRules of itemized fields such as environment variables or extended resources
func validateInstancesRules(fieldName string, itemRules map[string]interface{}, defaults map[string]interface{},
	spec map[string]interface{}, complianceErrors *ComplianceErrors, options ValidateComplianceOptions) {

	//  in policy validation, we have to add .instances to the name of the field so the user will be able to find them
	errFieldName := fieldName
	if options.PolicySelfValidation {
		errFieldName += itemized.InstancesSuffix
	}

	//
	//   key attribute depends on the field. for environment variables it is called "name" for extended
	//   resources "resource" and capabilities do not even have a key attribute, it's a simple strings array
	//
	itemizedInfo, _ := itemized.GetItemizedFieldInfo(fieldName)
	keyAttribute := itemizedInfo.KeyAttribute

	//
	//   load the itemRules details
	//
	canAdd := DefaultForCanAdd
	locked := datastructs.NewSet[string]()

	for itemRuleType, itemRule := range itemRules {
		switch itemRuleType {
		case string(PolicyCanAdd):
			canAdd = convutils.BoolMe(itemRule, DefaultForCanAdd)
		case string(PolicyLocked):
			for _, key := range itemRule.([]interface{}) {
				locked.Add(fmt.Sprintf("%v", key))
			}
		}
	}

	keysOfTheSpec, instancesWithNoKey := itemKeys(spec, fieldName, keyAttribute, "", options.PolicySelfValidation)
	keysOfDefault, _ := itemKeys(defaults, fieldName, keyAttribute, "", options.PolicySelfValidation)

	//
	//   for policy self validation, we need to make sure locked keys have default instances.
	//
	if options.PolicySelfValidation {
		for key := range locked {
			if !keysOfTheSpec.Has(key) {
				complianceErrors.Add(nil, fmt.Sprintf("default value is missing for locked item '%v'", key),
					fmt.Sprintf("%s[%s]", errFieldName, key), fieldName+itemized.InstancesSuffix)
			}
		}
		return
	}

	//
	//  for workload VS policy, check that workload does not attempt to add keys in case canAdd=false
	//
	if !canAdd {
		if instancesWithNoKey != 0 {
			complianceErrors.Add(&PolicyCanAdd, "the administrator prohibited adding new items", errFieldName, fieldName+itemized.InstancesSuffix)
		} else {
			for key := range keysOfTheSpec {
				if !keysOfDefault.Has(key) {
					complianceErrors.Add(&PolicyCanAdd, "the administrator prohibited adding new items", errFieldName, fieldName+itemized.InstancesSuffix)
					break
				}
			}
		}
	}

	//
	//   if no locked keys, non more things to validate
	//
	if len(locked) == 0 {
		return
	}

	//
	//   for checking locked keys, we need to find the items from the spec and from the defaults
	//
	valuesOfTheSpec := loadInstances(spec, fieldName, keyAttribute)
	valuesOfTheDefault := loadInstances(defaults, fieldName, keyAttribute)

	//
	//   make sure the workload does not attempt to change any of the locked items
	//
	for key := range keysOfTheSpec {
		if locked.Has(key) {
			valueOfTheSpec := valuesOfTheSpec[key]
			valueOfTheDefault := valuesOfTheDefault[key]

			isSameValue := assert.ObjectsAreEqualValues(valueOfTheDefault, valueOfTheSpec)
			if !isSameValue {
				complianceErrors.Add(&PolicyLocked, fmt.Sprintf("the administrator prohibited modifying item '%s'", key),
					fmt.Sprintf("%s[%s]", errFieldName, key), fieldName+itemized.InstancesSuffix)
			}
		}
	}
}

// load the instances of a given field to a map key -> instance
func loadInstances(spec map[string]interface{}, fieldName string, keyAttribute string) map[string]interface{} {
	instancesMap := make(map[string]interface{})

	//
	//   load the instances of the field. we attempt to load then with .instance suffix and without the suffix
	//   because the function should work for both workload spec and defaults spec
	//
	instances, err := jsonpath.Get(fmt.Sprintf("$.%s.instances", fieldName), spec)
	if err != nil && !strings.Contains(err.Error(), jsonPathUnknownFieldError) {
		instances, err = jsonpath.Get(fmt.Sprintf("$.%s", fieldName), spec)
	}
	if err != nil && !strings.Contains(err.Error(), jsonPathUnknownFieldError) {
		logging.Errorf("failed to load the instances of %s", fieldName)
		return instancesMap
	}

	if instances != nil {
		for _, instance := range convutils.ArrayMe(instances) {
			itemAsMap := convutils.MapMe(instance)
			instancesMap[fmt.Sprintf("%s", itemAsMap[keyAttribute])] = instance
		}
	}

	return instancesMap
}

// for itemized fields, there are specific rules for each attribute that we have to validate
// recall that the structure of an itemized field is
//
//		field:
//		 - name: aaa
//	    value: bbb
//		 - name: kkk
//	   value mmm
func validateAttributesRules(fieldName string, attributeRules map[string]interface{}, defaults map[string]interface{},
	values map[string]interface{}, defaultFromOrder DefaultFromOrder, complianceErrors *ComplianceErrors, options ValidateComplianceOptions) {

	//
	//   in the example above 'name' is the key attribute, but it depends on each field, so we have a table for the key
	//   attribute of each field. fields can be itemized without having keys (they will not have "lock ïnstance"
	//   rule but all the rest should work)
	//
	itemizedInfo, isItemized := itemized.GetItemizedFieldInfo(fieldName)
	if !isItemized {
		return
	}

	//
	//   default values for itemized fields are stored under attributes. in order to find this element,
	//   we need the following jsonpath expression (assuming 'fff' is the field, and 'name' is the key attribute)
	//        $.fff.attributes
	//   note that when validating policy itself the defaults is received as nil, and we have to continue
	//   sending it as nil to indicate policy validation.
	//
	var attributesDefaultsMap map[string]interface{}

	var loadDefaultsFrom map[string]interface{}
	defaultsExpr := fmt.Sprintf("$.%s.%s", fieldName, itemized.PolicyAttributes)

	if options.PolicySelfValidation {
		loadDefaultsFrom = values
	} else {
		loadDefaultsFrom = defaults
	}

	if loadDefaultsFrom != nil {
		attributesDefaultsMap = make(map[string]interface{})

		attributesDefaultsRaw, _ := jsonpath.Get(defaultsExpr, loadDefaultsFrom)
		if !reflect2.IsNil(attributesDefaultsRaw) {
			switch realType := attributesDefaultsRaw.(type) {
			case map[string]interface{}:
				attributesDefaultsMap = realType
			default:
				logging.Errorf("cannot load attributes of %s as map", defaultsExpr)
			}
		}
	}

	//
	//   load the items into an array. in the example above we'll get
	//     [0] { name: aaa, value: bbb }
	//     [1] { name: kkk, value: mmm }
	//   when checking the defaults of the policy this array will be under "instances:"
	//
	var specExpr string
	if options.PolicySelfValidation {
		specExpr = fmt.Sprintf("$.%s.%s[*]", fieldName, itemized.PolicyInstances)
	} else {
		specExpr = fmt.Sprintf("$.%s[*]", fieldName)
	}
	instancesSpec, jpErr := jsonpath.Get(specExpr, values)
	if jpErr != nil {
		logging.Errorf("failed to load the instances of field %s: %v", fieldName, jpErr)
		return
	}

	//
	//  for each item recursively check the primitive rules for its attributes
	//
	unused := make([]string, 0)
	instancesSpecArray := instancesSpec.([]interface{})

	for idx, instanceSpec := range instancesSpecArray {
		instanceSpecMap := convutils.MapMe(instanceSpec)

		keyAsString := fmt.Sprintf("%d", idx+1)

		keyAttribute := itemizedInfo.KeyAttribute
		if keyAttribute != "" {
			key := instanceSpecMap[keyAttribute]
			if key == nil {
				if !itemizedInfo.KeyIsOptional {
					complianceErrors.Add(nil, fmt.Sprintf("missing key attribute '%s' for item #%s", keyAttribute, keyAsString),
						fieldName, fieldName+itemized.InstancesSuffix)
				}
			} else {
				keyAsString = fmt.Sprintf("%v", key)
			}
		}

		//
		//  if there is more than one instance we need to add the key or index of the instance to the field name
		//
		fieldPrefix := fieldName

		if options.PolicySelfValidation {
			fieldPrefix += itemized.InstancesSuffix
		}

		//
		//   we want to include the index in two cases
		//   1. there is more than one instance
		//   2. there is a single instance but its key is uuid, which means it is an imposed asset
		//
		toAddInstanceIndex := len(instancesSpecArray) > 1
		if len(instancesSpecArray) == 1 && uuidRegex.MatchString(keyAsString) {
			toAddInstanceIndex = true
		}

		if toAddInstanceIndex {
			fieldPrefix += "[" + keyAsString + "]"
		}

		ValidateCompliance(fieldPrefix, attributeRules, attributesDefaultsMap, instanceSpecMap, defaultFromOrder, &unused, &options, complianceErrors)
	}

	//
	//    when validating defaults, we also need to validate the rules again the defaults values given to the attributes:
	//       extendedResources:
	//         attributes:
	//           quantity: 10M <- this needs to be validated against the attributes rules
	//
	if options.PolicySelfValidation {
		attributesSpec, jpErr := jsonpath.Get(defaultsExpr, values)
		if jpErr != nil {
			if !strings.HasPrefix(jpErr.Error(), jsonPathUnknownFieldError) {
				logging.Errorf("failed to load the attributes of field %s: %v", fieldName, jpErr)
				return
			}
		}

		ValidateCompliance(fieldName+itemized.AttributesSuffix, attributeRules, attributesDefaultsMap, convutils.MapMe(attributesSpec), defaultFromOrder, &unused, &options, complianceErrors)
	}
}

// itemKeys finds the keys of an itemized field. key attribute is optional, as fields like capabilities
// do not have a key, but just an array of strings
// subAttribute is needed for cases that the instances are not directly under the field.
// the function returns the set of keys, and optionally the number of instances not having keys
func itemKeys(rawItem map[string]interface{}, fieldName string, keyAttribute string, subAttribute string, validatingPolicy bool) (datastructs.Set[string], int) {
	resultKeys := datastructs.NewSet[string]()
	resultNonKeys := 0

	if subAttribute != "" {
		subAttribute = convutils.DotUs(".", subAttribute)
	}

	//
	//   in policy, the instances are under sub-field named instances
	//
	if validatingPolicy {
		subAttribute += itemized.InstancesSuffix
	}

	//
	//   for field 'fff' the jsonpath expression is $.fff[*] to get the instances
	//
	expr := fmt.Sprintf("$.%s%s[*]", fieldName, subAttribute)

	instances, err := jsonpath.Get(expr, rawItem)
	if err != nil {
		logging.Errorf("failed to load item keys of %s: %v\n", fieldName, err)
		return resultKeys, resultNonKeys
	}

	instancesArray := instances.([]interface{})

	if keyAttribute == "" {
		return resultKeys, len(instancesArray)
	}

	//
	//   collect all the nodes which were found with keys, and append their string value to the result keys
	//   count all the instances with no keys
	//
	for _, instance := range instancesArray {

		instanceAsMap := convutils.MapMe(instance)

		if instanceAsMap[keyAttribute] != nil {
			resultKeys.Add(fmt.Sprintf("%v", instanceAsMap[keyAttribute]))
		} else {
			resultNonKeys += 1
		}
	}

	return resultKeys, resultNonKeys
}

// composePolicyFieldName wipes workload related parts of the field and make sure
// to add attributes after the suffix, so if we get a.instances[123] as prefix, b as fieldName
// the result will be a.attributes.b
var removeInstanceKeyPattern = "\\[.*\\]"
var reRemoveInstanceKey = regexp.MustCompile(removeInstanceKeyPattern)

func composePolicyFieldName(fieldPrefix string, fieldName string) string {
	if fieldPrefix == "" {
		return convutils.DotUs(fieldPrefix, fieldName)
	}

	fieldPrefix = reRemoveInstanceKey.ReplaceAllString(fieldPrefix, "")
	fieldPrefix = strings.ReplaceAll(fieldPrefix, itemized.InstancesSuffix, "")
	fieldPrefix = strings.ReplaceAll(fieldPrefix, itemized.AttributesSuffix, "")

	return convutils.DotUs(fieldPrefix+itemized.AttributesSuffix, fieldName)
}

func getFieldValue(aMap map[string]interface{}, fieldName string) (interface{}, bool) {
	parts := strings.Split(fieldName, ".")
	var idx int
	for idx = 0; idx < len(parts)-1; idx++ {
		if node, found := aMap[parts[idx]]; found {
			aMap = convutils.MapMe(node)
		} else {
			return nil, false
		}
	}
	result, found := aMap[parts[idx]]

	return result, found
}

func setFieldValue(aMap map[string]interface{}, fieldName string, value interface{}, factor *float64) error {
	var factorizedValue interface{}
	if factor == nil || *factor == 1 {
		factorizedValue = value
	} else {
		var convErr error

		switch realValue := value.(type) {
		case int32, int64:
			intValue, convErr := validation.ToInt(value)
			if convErr == nil {
				factorizedValue = int64(float64(intValue) * (*factor))
			}
		case string:
			factorizedValue, convErr = MultiplyResourceQuantity(realValue, *factor)
		case float32, float64:
			floatValue, convErr := validation.ToFloat(value)
			if convErr == nil {
				factorizedValue = floatValue * (*factor)
			}
		case json.Number:
			floatValue, convErr := realValue.Float64()
			if convErr == nil {
				factorizedValue = floatValue * (*factor)
			}
		default:
			convErr = fmt.Errorf("field '%s' is not supported by defaultFrom", fieldName)
		}
		if convErr != nil {
			return convErr
		}
	}

	parts := strings.Split(fieldName, ".")
	var idx int
	for idx = 0; idx < len(parts)-1; idx++ {
		aMap = jsonutils.AddSection(aMap, parts[idx])
	}
	aMap[parts[idx]] = factorizedValue

	return nil
}

func MultiplyResourceQuantity(quantityStr string, factor float64) (string, error) {
	// Parse the original quantity string into a Quantity object
	quantity, err := resource.ParseQuantity(quantityStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse quantity: %v", err)
	}

	newQuantity := quantity.AsApproximateFloat64() * factor

	// Convert the new quantity back to a Quantity object
	result := resource.NewQuantity(int64(newQuantity), quantity.Format)
	return result.String(), nil
}

func validateOptionsRule(ruleDetails interface{}, value interface{}) error {
	optionList := gojsonq.New().FromInterface(ruleDetails)

	asString, asSet := optionsInfo(optionList)

	//
	//   value can be either single string or array
	//
	switch realValue := value.(type) {
	case string:
		if !asSet.Has(realValue) {
			return fmt.Errorf("the value of this field must be one of: %+v", asString)
		}
	case []interface{}:
		for idx, curValue := range realValue {
			if !asSet.Has(curValue.(string)) {
				return fmt.Errorf("item #%d has a value of '%s', which is not one of: %+v", idx+1, curValue, asString)
			}
		}
	default:
		logging.Errorf("unsupported value for array %v", realValue)
	}

	return nil
}
