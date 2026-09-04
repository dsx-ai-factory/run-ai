# DataUpdatableFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SchedulingRules** | Pointer to [**NullableDataUpdatableFieldsSchedulingRules**](DataUpdatableFieldsSchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | default order of node pools for workloads. will be enforced if no list is defined in workload policy | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 
**Resources** | Pointer to [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | [optional] 

## Methods

### NewDataUpdatableFields

`func NewDataUpdatableFields() *DataUpdatableFields`

NewDataUpdatableFields instantiates a new DataUpdatableFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataUpdatableFieldsWithDefaults

`func NewDataUpdatableFieldsWithDefaults() *DataUpdatableFields`

NewDataUpdatableFieldsWithDefaults instantiates a new DataUpdatableFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSchedulingRules

`func (o *DataUpdatableFields) GetSchedulingRules() DataUpdatableFieldsSchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *DataUpdatableFields) GetSchedulingRulesOk() (*DataUpdatableFieldsSchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *DataUpdatableFields) SetSchedulingRules(v DataUpdatableFieldsSchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *DataUpdatableFields) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *DataUpdatableFields) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *DataUpdatableFields) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *DataUpdatableFields) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *DataUpdatableFields) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *DataUpdatableFields) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *DataUpdatableFields) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *DataUpdatableFields) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *DataUpdatableFields) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *DataUpdatableFields) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *DataUpdatableFields) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *DataUpdatableFields) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *DataUpdatableFields) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetResources

`func (o *DataUpdatableFields) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DataUpdatableFields) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DataUpdatableFields) SetResources(v []Resources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DataUpdatableFields) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


