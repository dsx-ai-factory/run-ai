package policy_compliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	policyApi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	"github.com/run-ai/runai/runai-application-packages/jsonutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils/conjungo"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var distPolicyJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedPolicyDefaultsV2{})
var masterDefaultsSection = distPolicyJsonCracker.NameOf(&distPolicyJsonCracker.Struct().Master)
var workerDefaultsSection = distPolicyJsonCracker.NameOf(&distPolicyJsonCracker.Struct().Worker)

var distInferencePolicyJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedInferenceDefaultsV2{})
var leaderDefaultsSection = distInferencePolicyJsonCracker.NameOf(&distInferencePolicyJsonCracker.Struct().Leader)
var workerDistInferenceDefaultsSection = distInferencePolicyJsonCracker.NameOf(&distInferencePolicyJsonCracker.Struct().Worker)
var distInferenceWorkers = distInferencePolicyJsonCracker.NameOf(&distInferencePolicyJsonCracker.Struct().Workers)

const DummyKeyForFieldName = "_field_"

type PerReplicaField struct {
	ReplicaType openapi.ReplicaType
	FieldName   string
}

type PreProcessingOptions struct {
	AddFieldNames      bool    // add dummy instance with the field name. we need it for merging and mutation
	ReformatToWorkload bool    // remove instances/attributes sections, in order to reformat it to workload payload
	NamingIndicator    *string // if we want all instances to have a nick, provide the prefix for it.for example pol will yield pol-0, pol-1,...
}

// PrepareDefaultsForMutation prepares a policy to be merged with the workload
// For this purpose it does the following
//
//  1. adds policy key (a.k.a "nick") for any itemized instance that does not have
//     key provided by the user. it is also
//
//  2. removes the instances section and puts the instances as array just
//     like it appears in the workload
//
//  3. add the name of the field as additional dummy instance
//
//  4. returns the attributes for itemized field for later use
func PrepareDefaultsForMutation(defaults map[string]interface{}) map[PerReplicaField]interface{} {
	numbering := make(itemized.ItemizedNumbering)

	masterSection, hasMAsterSection := defaults[masterDefaultsSection]
	workerSection, hasWorkerSection := defaults[workerDefaultsSection]
	leaderSection, hasLeaderSection := defaults[leaderDefaultsSection]
	workerInferenceSection, hasWorkerDistInferenceSection := defaults[workerDistInferenceDefaultsSection]

	attributesDefaults := make(map[PerReplicaField]interface{})

	options := PreProcessingOptions{
		ReformatToWorkload: true,
		AddFieldNames:      true,
		NamingIndicator:    convutils.Ref(itemized.PolicyNamingIndicator),
	}

	if hasMAsterSection || hasWorkerSection || hasLeaderSection || hasWorkerDistInferenceSection {
		preprocessDefaults(convutils.MapMe(workerSection), numbering, "", convutils.Ref(openapi.REPLICATYPE_WORKER), attributesDefaults, options)
		preprocessDefaults(convutils.MapMe(masterSection), numbering, "", convutils.Ref(openapi.REPLICATYPE_MASTER), attributesDefaults, options)
		preprocessDefaults(convutils.MapMe(workerInferenceSection), numbering, "", convutils.Ref(openapi.REPLICATYPE_WORKER), attributesDefaults, options)
		preprocessDefaults(convutils.MapMe(leaderSection), numbering, "", convutils.Ref(openapi.REPLICATYPE_MASTER), attributesDefaults, options)
	} else {
		preprocessDefaults(defaults, numbering, "", nil, attributesDefaults, options)
	}

	return attributesDefaults
}

// PrepareTemplateForMutation is similar to prepare defaults for merge, only need
// to add the dummy instance
func PrepareTemplateForMutation(templateSpec map[string]interface{}) {
	PrepareDefaultsForMerge(templateSpec, convutils.Ref(itemized.TemplateNamingIndicator))
}

// PrepareDefaultsForMerge prepares a policy to be merged with defaults of another policy
// For this purpose, it only needs to add the name of the field as additional dummy instance
func PrepareDefaultsForMerge(defaults map[string]interface{}, namingIndicator *string) {
	numbering := make(itemized.ItemizedNumbering)

	masterSection, hasMAsterSection := defaults[masterDefaultsSection]
	workerSection, hasWorkerSection := defaults[workerDefaultsSection]
	leaderSection, hasLeaderSection := defaults[leaderDefaultsSection]
	workerInferenceSection, hasWorkerDistInferenceSection := defaults[workerDistInferenceDefaultsSection]

	options := PreProcessingOptions{
		AddFieldNames:   true,
		NamingIndicator: namingIndicator,
	}

	attributesDefaults := make(map[PerReplicaField]interface{})

	if hasMAsterSection || hasWorkerSection || hasLeaderSection || hasWorkerDistInferenceSection {
		preprocessDefaults(convutils.MapMe(masterSection), numbering, "", convutils.Ref(openapi.REPLICATYPE_MASTER), attributesDefaults, options)
		preprocessDefaults(convutils.MapMe(workerSection), numbering, "", convutils.Ref(openapi.REPLICATYPE_WORKER), attributesDefaults, options)
		preprocessDefaults(convutils.MapMe(workerInferenceSection), numbering, "", convutils.Ref(openapi.REPLICATYPE_WORKER), attributesDefaults, options)
		preprocessDefaults(convutils.MapMe(leaderSection), numbering, "", convutils.Ref(openapi.REPLICATYPE_MASTER), attributesDefaults, options)
	} else {
		preprocessDefaults(defaults, numbering, "", nil, attributesDefaults, options)
	}
}

func preprocessDefaults(defaults map[string]interface{}, numbering itemized.ItemizedNumbering,
	section string, replicaType *openapi.ReplicaType, attributesDefaults map[PerReplicaField]interface{},
	options PreProcessingOptions) {

	for fieldName, fieldValue := range defaults {
		fullFieldName := convutils.DotUs(section, fieldName)

		//
		//    recursively call the function for nested sections (compute, storage, etc...)
		//
		if itemized.NestedPolicySections.Has(fullFieldName) {
			preprocessDefaults(convutils.MapMe(defaults[fieldName]), numbering, fullFieldName, replicaType, attributesDefaults, options)
			continue
		}

		//
		//   check if the current field is itemized, otherwise we leave it as is
		//
		itemizedFieldInfo, isItemized := itemized.GetItemizedFieldInfo(fullFieldName)
		if !isItemized {
			continue
		}

		keyField := itemizedFieldInfo.KeyAttribute

		//
		//   in defaults, the itemized field is a map with up to two entries, one for instances
		//   the other for attributes
		//
		var defaultInstances []interface{}

		switch realType := fieldValue.(type) {
		case []interface{}:
			defaultInstances = realType
		case map[string]interface{}:
			defaultInstances = convutils.ArrayMe(realType[itemized.PolicyInstances])

			//
			//   if there is attributes section, keep it aside for future apply
			//   of attribute defaults.
			//
			defaultAttributes, attributeDefaultsFound := realType[itemized.PolicyAttributes]
			if attributeDefaultsFound {
				key := PerReplicaField{
					ReplicaType: convutils.Deref(replicaType, openapi.REPLICATYPE_WORKER),
					FieldName:   fullFieldName,
				}
				attributesDefaults[key] = defaultAttributes
			}
		}

		if defaultInstances != nil {
			if options.NamingIndicator != nil {
				//
				//   for the instances, we want to make sure all have instance key
				//
				for _, instance := range defaultInstances {
					//
					//  for any instance, no matter if it has a key or not, call the numbering manager
					//  needs to be aware of all the instances in order to assign correct numbers for them
					//
					policyNick := numbering.DefaultPolicyNick(*options.NamingIndicator, itemizedFieldInfo.NicknamePrefix, replicaType)

					instanceFields := convutils.MapMe(instance)
					if instanceFields[DummyKeyForFieldName] != nil {
						continue // no need to handle dummy keys
					}

					var itemizedKeyValue *string
					if instanceFields[keyField] != nil {
						itemizedKeyValue = convutils.Ref(instanceFields[keyField].(string))
					}

					//
					//   if the instance does not have provided key, give it the key that the
					//   numbering manager determined. this mimics the key that it would get in the cluster
					//
					if itemizedKeyValue == nil {
						instanceFields[keyField] = policyNick
					}
				}
			}

			if options.AddFieldNames {
				fieldNameAlreadyAdded := false

				if len(defaultInstances) > 0 {
					asMap, ok := defaultInstances[len(defaultInstances)-1].(map[string]interface{})
					if ok && asMap[DummyKeyForFieldName] != nil {
						fieldNameAlreadyAdded = true
					}
				}

				//
				//   add additional instance with the field name. we need it as a cookie for the
				//   conjungo merging function
				//
				if !fieldNameAlreadyAdded {
					defaultInstances = append(defaultInstances, map[string]interface{}{DummyKeyForFieldName: fullFieldName})
				}
			}

			switch realType := fieldValue.(type) {
			case []interface{}:
				defaults[fieldName] = defaultInstances
			case map[string]interface{}:
				if options.ReformatToWorkload {
					//
					//   put the instances on top of the field, this removes the instances and attributes
					//   sections that we do not want during merging
					//
					defaults[fieldName] = defaultInstances
				} else {
					realType[itemized.PolicyInstances] = defaultInstances
				}
			}
		}
	}
}

// shouldIgnoreWorkerPolicyDefaults determines whether to skip applying worker-related policy defaults
// to the workload specification. Policy defaults should be ignored in the following cases:
// 1. If workers section doesn't exist in spec, OR
// 2. If workers field is set AND evaluates to 0
func shouldIgnoreWorkerPolicyDefaults(workersFieldExist bool, workers interface{}) bool {
	if !workersFieldExist {
		return true
	}
	if workers == nil {
		return false
	}

	n, err := workers.(json.Number).Int64()
	// Return true to ignore policy defaults if either:
	// - Failed to parse workers number (invalid format)
	// - Or workers count is 0 (no workers needed)
	if err != nil || n == 0 {
		return true
	}

	return false
}

func ApplyDefaults(spec map[string]interface{}, defaults map[string]interface{}, attributesDefaults map[PerReplicaField]interface{}) *openapi.Error {
	// set merge options, as we need to handle the itemized our own way.
	mergeOptions := conjungo.NewOptions()
	mergeOptions.SetKindMergeFunc(reflect.Slice, WorkloadSlicesMergeFunc)

	// if the workload does not have master at all (not explicitly specified and not same as worker)
	// then we do not want to apply policy on the master section
	if spec[masterDefaultsSection] == nil {
		delete(defaults, masterDefaultsSection)
	}

	// Similarly for distributed inference: if the workload does not have worker at all,
	// then we do not want to apply policy on the worker section
	if spec[workerDistInferenceDefaultsSection] == nil && spec[leaderDefaultsSection] != nil {
		workers, workersFieldExist := spec[distInferenceWorkers]
		if shouldIgnoreWorkerPolicyDefaults(workersFieldExist, workers) {
			delete(defaults, workerDefaultsSection)
		}

	}

	err := conjungo.Merge(&defaults, spec, mergeOptions)
	if err != nil {
		return openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to apply policy defaults: %v", err))
	}

	// now we have all the instances sorted out, so we can apply the attribute defaults
	// wherever needed.
	masterSection, hasMAsterSection := defaults[masterDefaultsSection]
	workerSection, hasWorkerSection := defaults[workerDefaultsSection]
	leaderSection, hasLeaderSection := defaults[leaderDefaultsSection]
	workerInferenceSection, hasWorkerDistInferenceSection := defaults[workerDistInferenceDefaultsSection]

	if hasMAsterSection || hasWorkerSection || hasLeaderSection || hasWorkerDistInferenceSection {
		applyAttributesDefaults(convutils.MapMe(masterSection), "", openapi.REPLICATYPE_MASTER, attributesDefaults)
		applyAttributesDefaults(convutils.MapMe(workerSection), "", openapi.REPLICATYPE_WORKER, attributesDefaults)
		applyAttributesDefaults(convutils.MapMe(leaderSection), "", openapi.REPLICATYPE_MASTER, attributesDefaults)
		applyAttributesDefaults(convutils.MapMe(workerInferenceSection), "", openapi.REPLICATYPE_WORKER, attributesDefaults)
	} else {
		applyAttributesDefaults(defaults, "", openapi.REPLICATYPE_WORKER, attributesDefaults)
	}

	return nil
}

func WorkloadSlicesMergeFunc(valueFromDefaults, valueFromWorkload reflect.Value, o *conjungo.Options) (reflect.Value, error) {
	// look at the defaults instances (this is also our target for merge. we always
	// force the workload values -> upon the defaults, and so the defaults becomes
	// the result of the merge with the original defaults modified by workload defaults
	defaultsInstances := valueFromDefaults.Interface().([]interface{})

	// a slice of itemized field contains at least two elements, the value and the
	// dummy instance that we add that contains the field name. any array smaller than
	// that is not an itemized (it can be a capabilities array containing only one value).
	// for such fields we just return the value from the workload
	if len(defaultsInstances) == 0 {
		return valueFromWorkload, nil
	}

	// assuming this is a slice containing itemized instances, obtain the field name
	// from the last instance. not having this dummy instance indicate, again, that
	// this is a slice unrelated to itemized field. for such fields we return the
	// value from the workload
	lastDefaultsInstance := len(defaultsInstances) - 1
	lastInstance := convutils.MapMe(defaultsInstances[lastDefaultsInstance])

	fullFieldName, isItemized := lastInstance[DummyKeyForFieldName].(string)
	if !isItemized {
		return valueFromWorkload, nil
	}

	// obtain information about this itemized field
	itemizedFieldInfo, isItemized := itemized.GetItemizedFieldInfo(fullFieldName)
	if !isItemized {
		return valueFromWorkload, nil
	}

	keyField := itemizedFieldInfo.KeyAttribute

	// now create a map with the workload instances that have a key (instance key is
	// mandatory for env variables, extended resources, etc, but it is optional for
	// storage instances so we only take those that do have a key
	workloadInstances := valueFromWorkload.Interface().([]interface{})
	workloadKeyValueMap := make(map[string]interface{})

	workloadInstancesWithNoKey := 0

	for _, workloadInstance := range workloadInstances {
		keyValue := convutils.MapMe(workloadInstance)[keyField]
		if keyValue == nil {
			workloadInstancesWithNoKey += 1
		} else {
			keyValueAsString := fmt.Sprintf("%v", keyValue)
			workloadKeyValueMap[keyValueAsString] = workloadInstance
		}
	}

	// now traverse the defaults instances and replace the value for those
	// that have an instance in the workload
	for i := 0; i < len(defaultsInstances)-1; i++ {
		defaultsInstance := defaultsInstances[i]
		keyValue := convutils.MapMe(defaultsInstance)[keyField]
		keyValueAsString := fmt.Sprintf("%v", keyValue)

		if workloadInstance, hasWorkloadInstance := workloadKeyValueMap[keyValueAsString]; hasWorkloadInstance {
			defaultsInstances[i] = workloadInstance
			delete(workloadKeyValueMap, keyValueAsString)
		}
	}

	// the last thing for us to do is to add workloads element that do not appear in policy
	// the defaults instances. before doing that we have to move the last element (the dummy
	// element that holds the field name) to the end of the array, so we first add empty the
	// number of elements we need to the array, and then we will fill them with the workload
	// additional items
	for i := 0; i < workloadInstancesWithNoKey+len(workloadKeyValueMap); i++ {
		valueFromDefaults = reflect.Append(valueFromDefaults, reflect.ValueOf(lastInstance))
	}

	defaultsInstances = valueFromDefaults.Interface().([]interface{})

	// now we add all the workload instances that (a) did not have a key at all (b) had a key
	// hat is not associated with any policy key. both cases indicate workload-only instance
	for _, workloadInstance := range workloadInstances {
		keyValue := convutils.MapMe(workloadInstance)[keyField]

		workloadOnly := false
		if keyValue == nil {
			workloadOnly = true // (a)
		} else {
			keyValueAsString := fmt.Sprintf("%v", keyValue)
			_, workloadOnly = workloadKeyValueMap[keyValueAsString] // (b)
		}

		if workloadOnly {
			defaultsInstances[lastDefaultsInstance] = workloadInstance
			lastDefaultsInstance += 1
		}
	}

	return valueFromDefaults, nil
}

func applyAttributesDefaults(defaults map[string]interface{}, section string,
	replicaType openapi.ReplicaType, attributesDefaults map[PerReplicaField]interface{}) {

	for fieldName, fieldValue := range defaults {
		fullFieldName := convutils.DotUs(section, fieldName)

		//
		//    recursively call the function for nested sections (compute, storage, etc...)
		//
		if itemized.NestedPolicySections.Has(fullFieldName) {
			applyAttributesDefaults(convutils.MapMe(defaults[fieldName]), fullFieldName, replicaType, attributesDefaults)
			continue
		}

		//
		//    check if there are attribute defaults for this field in this replica
		//
		defaultsForField := attributesDefaults[PerReplicaField{
			ReplicaType: replicaType,
			FieldName:   fullFieldName,
		}]

		switch instances := fieldValue.(type) {
		case []interface{}:
			// for each attribute with defaults scan all instances and add the default for items
			// that needs it
			for attribute, defaultForAttribute := range convutils.MapMe(defaultsForField) {
				for _, instance := range instances {
					instanceAttributes := convutils.MapMe(instance)
					_, instanceHasTheAttribute := instanceAttributes[attribute]
					if !instanceHasTheAttribute {
						instanceAttributes[attribute] = defaultForAttribute
					}
				}
			}

			// remove the dummy element, we no longer need it
			if len(instances) > 0 {
				lastInstance := len(instances) - 1
				if _, hasDummyElement := convutils.MapMe(instances[lastInstance])[DummyKeyForFieldName]; hasDummyElement {
					defaults[fieldName] = instances[0 : len(instances)-1]
				}
			}
		}
	}
}

func RemoveFieldNameInstance(defaults map[string]interface{}) {
	masterSection, hasMAsterSection := defaults[masterDefaultsSection]
	workerSection, hasWorkerSection := defaults[workerDefaultsSection]
	leaderSection, hasLeaderSection := defaults[leaderDefaultsSection]
	workerInferenceSection, hasWorkerDistInferenceSection := defaults[workerDistInferenceDefaultsSection]

	if hasMAsterSection || hasWorkerSection || hasLeaderSection || hasWorkerDistInferenceSection {
		removeFiledNameInstance(convutils.MapMe(masterSection), "")
		removeFiledNameInstance(convutils.MapMe(workerSection), "")
		removeFiledNameInstance(convutils.MapMe(leaderSection), "")
		removeFiledNameInstance(convutils.MapMe(workerInferenceSection), "")
	} else {
		removeFiledNameInstance(defaults, "")
	}
}

func removeFiledNameInstance(defaults map[string]interface{}, section string) {
	for fieldName, fieldValue := range defaults {
		fullFieldName := convutils.DotUs(section, fieldName)

		//
		//    recursively call the function for nested sections (compute, storage, etc...)
		//
		if itemized.NestedPolicySections.Has(fullFieldName) {
			removeFiledNameInstance(convutils.MapMe(defaults[fieldName]), fullFieldName)
			continue
		}

		switch instances := fieldValue.(type) {
		case []interface{}:
			if len(instances) >= 1 {
				// this one handles workload that holds the instances directly under the field
				lastInstance := len(instances) - 1
				if _, hasDummyElement := convutils.MapMe(instances[lastInstance])[DummyKeyForFieldName]; hasDummyElement {
					defaults[fieldName] = instances[0 : len(instances)-1]
				}
			}

		case map[string]interface{}:
			// this one handles policy which has a subsection instances, that holds the instances array
			value, hasInstances := instances[itemized.PolicyInstances]
			if hasInstances {
				instancesArr := value.([]interface{})
				if len(instancesArr) >= 1 {
					lastInstance := len(instancesArr) - 1
					if _, hasDummyElement := convutils.MapMe(instancesArr[lastInstance])[DummyKeyForFieldName]; hasDummyElement {
						instances[itemized.PolicyInstances] = instancesArr[0 : len(instancesArr)-1]
					}
				}
			}
		}

	}
}

///////////////////////////////////////////////////////////////////////////
// Note! the functions below are meant to work on the spec of the workload
// For distributed, they receive the master-spec or worker-spec.
///////////////////////////////////////////////////////////////////////////

// MakeSureAllItemizedHaveKey make sure that all the instances of itemized fields
// in aa workload have reference-key. it does not change anything else in the workload
func MakeSureAllItemizedHaveKey(spec map[string]interface{}) {
	numbering := make(itemized.ItemizedNumbering)

	makeSureAllItemizedHaveKey(spec, "", numbering)
}

func makeSureAllItemizedHaveKey(spec map[string]interface{}, section string, numbering itemized.ItemizedNumbering) {
	for fieldName, fieldValue := range spec {
		fullFieldName := convutils.DotUs(section, fieldName)

		//
		//    recursively call the function for nested sections (compute, storage, etc...)
		//
		if itemized.NestedPolicySections.Has(fullFieldName) {
			makeSureAllItemizedHaveKey(convutils.MapMe(spec[fieldName]), fullFieldName, numbering)
			continue
		}

		//
		//   check if the current field is itemized, otherwise we leave it as is
		//
		itemizedFieldInfo, isItemized := itemized.GetItemizedFieldInfo(fullFieldName)
		if !isItemized {
			continue
		}

		keyField := itemizedFieldInfo.KeyAttribute

		//
		//   in defaults, the itemized field is a map with up to two entries, one for instances
		//   the other for attributes
		//
		switch instances := fieldValue.(type) {
		case []interface{}:

			//
			//   for the instances, we want to make sure all have instance key
			//
			for _, instance := range instances {
				instanceFields := convutils.MapMe(instance)

				var itemizedKeyValue *string
				if instanceFields[keyField] != nil {
					itemizedKeyValue = convutils.Ref(instanceFields[keyField].(string))
				}

				//
				//   if the instance does not have provided key, give it the key that the
				//   numbering manager determined. this mimics the key that it would get in the cluster
				//
				if itemizedKeyValue == nil {
					availableNick := numbering.DummyNick(itemizedFieldInfo.NicknamePrefix)
					instanceFields[keyField] = availableNick
				}
			}
		}
	}
}
