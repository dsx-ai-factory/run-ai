# ProjectEffective

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SchedulingRules** | Pointer to [**NullableSchedulingRules**](SchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | default order of node pools for workloads. will be enforced if no list is defined in workload policy | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 

## Methods

### NewProjectEffective

`func NewProjectEffective() *ProjectEffective`

NewProjectEffective instantiates a new ProjectEffective object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectEffectiveWithDefaults

`func NewProjectEffectiveWithDefaults() *ProjectEffective`

NewProjectEffectiveWithDefaults instantiates a new ProjectEffective object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSchedulingRules

`func (o *ProjectEffective) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *ProjectEffective) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *ProjectEffective) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *ProjectEffective) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *ProjectEffective) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *ProjectEffective) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *ProjectEffective) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *ProjectEffective) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *ProjectEffective) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *ProjectEffective) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *ProjectEffective) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *ProjectEffective) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *ProjectEffective) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *ProjectEffective) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *ProjectEffective) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *ProjectEffective) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


