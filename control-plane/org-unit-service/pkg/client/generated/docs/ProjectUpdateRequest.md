# ProjectUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** | The description of the project. | [optional] 
**SchedulingRules** | Pointer to [**NullableSchedulingRules**](SchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | An ordered list of node pools that defines the default scheduling preference for workloads in this project. The Scheduler will attempt to schedule workloads on node pools in this order. This setting is enforced if no node pool list is defined in the workload policy. | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 
**EnforceRunaiScheduler** | Pointer to **NullableBool** | a flag to enforce that all pods in this namespace will be scheduled with runai-scheduler. Supported for cluster v2.21+ | [optional] 
**Resources** | [**[]ResourcesNullable**](ResourcesNullable.md) | Resources assigned to this Organization per Node Pool | 

## Methods

### NewProjectUpdateRequest

`func NewProjectUpdateRequest(resources []ResourcesNullable, ) *ProjectUpdateRequest`

NewProjectUpdateRequest instantiates a new ProjectUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectUpdateRequestWithDefaults

`func NewProjectUpdateRequestWithDefaults() *ProjectUpdateRequest`

NewProjectUpdateRequestWithDefaults instantiates a new ProjectUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *ProjectUpdateRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ProjectUpdateRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ProjectUpdateRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ProjectUpdateRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSchedulingRules

`func (o *ProjectUpdateRequest) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *ProjectUpdateRequest) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *ProjectUpdateRequest) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *ProjectUpdateRequest) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *ProjectUpdateRequest) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *ProjectUpdateRequest) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *ProjectUpdateRequest) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *ProjectUpdateRequest) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *ProjectUpdateRequest) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *ProjectUpdateRequest) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *ProjectUpdateRequest) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *ProjectUpdateRequest) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *ProjectUpdateRequest) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *ProjectUpdateRequest) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *ProjectUpdateRequest) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *ProjectUpdateRequest) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetEnforceRunaiScheduler

`func (o *ProjectUpdateRequest) GetEnforceRunaiScheduler() bool`

GetEnforceRunaiScheduler returns the EnforceRunaiScheduler field if non-nil, zero value otherwise.

### GetEnforceRunaiSchedulerOk

`func (o *ProjectUpdateRequest) GetEnforceRunaiSchedulerOk() (*bool, bool)`

GetEnforceRunaiSchedulerOk returns a tuple with the EnforceRunaiScheduler field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforceRunaiScheduler

`func (o *ProjectUpdateRequest) SetEnforceRunaiScheduler(v bool)`

SetEnforceRunaiScheduler sets EnforceRunaiScheduler field to given value.

### HasEnforceRunaiScheduler

`func (o *ProjectUpdateRequest) HasEnforceRunaiScheduler() bool`

HasEnforceRunaiScheduler returns a boolean if a field has been set.

### SetEnforceRunaiSchedulerNil

`func (o *ProjectUpdateRequest) SetEnforceRunaiSchedulerNil(b bool)`

 SetEnforceRunaiSchedulerNil sets the value for EnforceRunaiScheduler to be an explicit nil

### UnsetEnforceRunaiScheduler
`func (o *ProjectUpdateRequest) UnsetEnforceRunaiScheduler()`

UnsetEnforceRunaiScheduler ensures that no value is present for EnforceRunaiScheduler, not even an explicit nil
### GetResources

`func (o *ProjectUpdateRequest) GetResources() []ResourcesNullable`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ProjectUpdateRequest) GetResourcesOk() (*[]ResourcesNullable, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ProjectUpdateRequest) SetResources(v []ResourcesNullable)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


