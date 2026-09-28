# DataDepartmentFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** | The description of the department. | [optional] 
**SchedulingRules** | Pointer to [**NullableSchedulingRules**](SchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | An ordered list of node pools that defines the default scheduling preference for workloads in this department. The Scheduler will attempt to schedule workloads on node pools in this order. This setting is enforced if no node pool list is defined in the workload policy. | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 
**Name** | **string** | The unique name of the department within the cluster. Must be a valid Kubernetes name (lowercase, alphanumeric, hyphens allowed). | 
**ClusterId** | **string** | The id of the cluster. | 
**OvertimeData** | Pointer to [**OvertimeData**](OvertimeData.md) |  | [optional] 

## Methods

### NewDataDepartmentFields

`func NewDataDepartmentFields(name string, clusterId string, ) *DataDepartmentFields`

NewDataDepartmentFields instantiates a new DataDepartmentFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataDepartmentFieldsWithDefaults

`func NewDataDepartmentFieldsWithDefaults() *DataDepartmentFields`

NewDataDepartmentFieldsWithDefaults instantiates a new DataDepartmentFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *DataDepartmentFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DataDepartmentFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DataDepartmentFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DataDepartmentFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSchedulingRules

`func (o *DataDepartmentFields) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *DataDepartmentFields) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *DataDepartmentFields) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *DataDepartmentFields) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *DataDepartmentFields) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *DataDepartmentFields) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *DataDepartmentFields) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *DataDepartmentFields) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *DataDepartmentFields) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *DataDepartmentFields) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *DataDepartmentFields) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *DataDepartmentFields) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *DataDepartmentFields) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *DataDepartmentFields) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *DataDepartmentFields) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *DataDepartmentFields) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetName

`func (o *DataDepartmentFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DataDepartmentFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DataDepartmentFields) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *DataDepartmentFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DataDepartmentFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DataDepartmentFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetOvertimeData

`func (o *DataDepartmentFields) GetOvertimeData() OvertimeData`

GetOvertimeData returns the OvertimeData field if non-nil, zero value otherwise.

### GetOvertimeDataOk

`func (o *DataDepartmentFields) GetOvertimeDataOk() (*OvertimeData, bool)`

GetOvertimeDataOk returns a tuple with the OvertimeData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOvertimeData

`func (o *DataDepartmentFields) SetOvertimeData(v OvertimeData)`

SetOvertimeData sets OvertimeData field to given value.

### HasOvertimeData

`func (o *DataDepartmentFields) HasOvertimeData() bool`

HasOvertimeData returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


