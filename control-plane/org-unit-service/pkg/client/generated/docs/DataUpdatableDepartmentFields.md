# DataUpdatableDepartmentFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** | The description of the department. | [optional] 
**SchedulingRules** | Pointer to [**NullableSchedulingRules**](SchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | An ordered list of node pools that defines the default scheduling preference for workloads in this department. The Scheduler will attempt to schedule workloads on node pools in this order. This setting is enforced if no node pool list is defined in the workload policy. | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 

## Methods

### NewDataUpdatableDepartmentFields

`func NewDataUpdatableDepartmentFields() *DataUpdatableDepartmentFields`

NewDataUpdatableDepartmentFields instantiates a new DataUpdatableDepartmentFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataUpdatableDepartmentFieldsWithDefaults

`func NewDataUpdatableDepartmentFieldsWithDefaults() *DataUpdatableDepartmentFields`

NewDataUpdatableDepartmentFieldsWithDefaults instantiates a new DataUpdatableDepartmentFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *DataUpdatableDepartmentFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DataUpdatableDepartmentFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DataUpdatableDepartmentFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DataUpdatableDepartmentFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSchedulingRules

`func (o *DataUpdatableDepartmentFields) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *DataUpdatableDepartmentFields) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *DataUpdatableDepartmentFields) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *DataUpdatableDepartmentFields) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *DataUpdatableDepartmentFields) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *DataUpdatableDepartmentFields) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *DataUpdatableDepartmentFields) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *DataUpdatableDepartmentFields) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *DataUpdatableDepartmentFields) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *DataUpdatableDepartmentFields) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *DataUpdatableDepartmentFields) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *DataUpdatableDepartmentFields) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *DataUpdatableDepartmentFields) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *DataUpdatableDepartmentFields) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *DataUpdatableDepartmentFields) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *DataUpdatableDepartmentFields) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


