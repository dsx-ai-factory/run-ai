# DataUpdatableFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** | The description of the project. | [optional] 
**SchedulingRules** | Pointer to [**NullableSchedulingRules**](SchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | An ordered list of node pools that defines the default scheduling preference for workloads in this project. The Scheduler will attempt to schedule workloads on node pools in this order. This setting is enforced if no node pool list is defined in the workload policy. | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 

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

### GetDescription

`func (o *DataUpdatableFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DataUpdatableFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DataUpdatableFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DataUpdatableFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSchedulingRules

`func (o *DataUpdatableFields) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *DataUpdatableFields) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *DataUpdatableFields) SetSchedulingRules(v SchedulingRules)`

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


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


