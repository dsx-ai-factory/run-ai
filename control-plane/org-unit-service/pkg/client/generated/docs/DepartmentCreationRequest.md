# DepartmentCreationRequest

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
**Resources** | [**[]ResourcesNullable**](ResourcesNullable.md) | An array of resource configurations for the department. Each entry defines GPU, CPU, and CPU memory assigned to a specific node pool. | 

## Methods

### NewDepartmentCreationRequest

`func NewDepartmentCreationRequest(name string, clusterId string, resources []ResourcesNullable, ) *DepartmentCreationRequest`

NewDepartmentCreationRequest instantiates a new DepartmentCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentCreationRequestWithDefaults

`func NewDepartmentCreationRequestWithDefaults() *DepartmentCreationRequest`

NewDepartmentCreationRequestWithDefaults instantiates a new DepartmentCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *DepartmentCreationRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DepartmentCreationRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DepartmentCreationRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DepartmentCreationRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSchedulingRules

`func (o *DepartmentCreationRequest) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *DepartmentCreationRequest) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *DepartmentCreationRequest) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *DepartmentCreationRequest) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *DepartmentCreationRequest) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *DepartmentCreationRequest) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *DepartmentCreationRequest) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *DepartmentCreationRequest) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *DepartmentCreationRequest) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *DepartmentCreationRequest) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *DepartmentCreationRequest) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *DepartmentCreationRequest) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *DepartmentCreationRequest) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *DepartmentCreationRequest) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *DepartmentCreationRequest) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *DepartmentCreationRequest) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetName

`func (o *DepartmentCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *DepartmentCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DepartmentCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DepartmentCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetOvertimeData

`func (o *DepartmentCreationRequest) GetOvertimeData() OvertimeData`

GetOvertimeData returns the OvertimeData field if non-nil, zero value otherwise.

### GetOvertimeDataOk

`func (o *DepartmentCreationRequest) GetOvertimeDataOk() (*OvertimeData, bool)`

GetOvertimeDataOk returns a tuple with the OvertimeData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOvertimeData

`func (o *DepartmentCreationRequest) SetOvertimeData(v OvertimeData)`

SetOvertimeData sets OvertimeData field to given value.

### HasOvertimeData

`func (o *DepartmentCreationRequest) HasOvertimeData() bool`

HasOvertimeData returns a boolean if a field has been set.

### GetResources

`func (o *DepartmentCreationRequest) GetResources() []ResourcesNullable`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentCreationRequest) GetResourcesOk() (*[]ResourcesNullable, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentCreationRequest) SetResources(v []ResourcesNullable)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


