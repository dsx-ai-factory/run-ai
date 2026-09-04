# DepartmentForList

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
**Id** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 
**Children** | Pointer to [**[]DepartmentProjectChild**](DepartmentProjectChild.md) |  | [optional] 
**Status** | Pointer to [**NullableDepartmentForListAllOfStatus**](DepartmentForListAllOfStatus.md) |  | [optional] 
**TotalResources** | [**TotalResources**](TotalResources.md) |  | 
**Resources** | [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | 

## Methods

### NewDepartmentForList

`func NewDepartmentForList(name string, clusterId string, id string, createdAt time.Time, updatedAt time.Time, totalResources TotalResources, resources []Resources, ) *DepartmentForList`

NewDepartmentForList instantiates a new DepartmentForList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentForListWithDefaults

`func NewDepartmentForListWithDefaults() *DepartmentForList`

NewDepartmentForListWithDefaults instantiates a new DepartmentForList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *DepartmentForList) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DepartmentForList) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DepartmentForList) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DepartmentForList) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetSchedulingRules

`func (o *DepartmentForList) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *DepartmentForList) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *DepartmentForList) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *DepartmentForList) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *DepartmentForList) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *DepartmentForList) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *DepartmentForList) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *DepartmentForList) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *DepartmentForList) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *DepartmentForList) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *DepartmentForList) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *DepartmentForList) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *DepartmentForList) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *DepartmentForList) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *DepartmentForList) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *DepartmentForList) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetName

`func (o *DepartmentForList) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentForList) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentForList) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *DepartmentForList) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DepartmentForList) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DepartmentForList) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetOvertimeData

`func (o *DepartmentForList) GetOvertimeData() OvertimeData`

GetOvertimeData returns the OvertimeData field if non-nil, zero value otherwise.

### GetOvertimeDataOk

`func (o *DepartmentForList) GetOvertimeDataOk() (*OvertimeData, bool)`

GetOvertimeDataOk returns a tuple with the OvertimeData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOvertimeData

`func (o *DepartmentForList) SetOvertimeData(v OvertimeData)`

SetOvertimeData sets OvertimeData field to given value.

### HasOvertimeData

`func (o *DepartmentForList) HasOvertimeData() bool`

HasOvertimeData returns a boolean if a field has been set.

### GetId

`func (o *DepartmentForList) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DepartmentForList) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DepartmentForList) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *DepartmentForList) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DepartmentForList) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DepartmentForList) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *DepartmentForList) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *DepartmentForList) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *DepartmentForList) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetCreatedBy

`func (o *DepartmentForList) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *DepartmentForList) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *DepartmentForList) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *DepartmentForList) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *DepartmentForList) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *DepartmentForList) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *DepartmentForList) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *DepartmentForList) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetChildren

`func (o *DepartmentForList) GetChildren() []DepartmentProjectChild`

GetChildren returns the Children field if non-nil, zero value otherwise.

### GetChildrenOk

`func (o *DepartmentForList) GetChildrenOk() (*[]DepartmentProjectChild, bool)`

GetChildrenOk returns a tuple with the Children field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildren

`func (o *DepartmentForList) SetChildren(v []DepartmentProjectChild)`

SetChildren sets Children field to given value.

### HasChildren

`func (o *DepartmentForList) HasChildren() bool`

HasChildren returns a boolean if a field has been set.

### SetChildrenNil

`func (o *DepartmentForList) SetChildrenNil(b bool)`

 SetChildrenNil sets the value for Children to be an explicit nil

### UnsetChildren
`func (o *DepartmentForList) UnsetChildren()`

UnsetChildren ensures that no value is present for Children, not even an explicit nil
### GetStatus

`func (o *DepartmentForList) GetStatus() DepartmentForListAllOfStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DepartmentForList) GetStatusOk() (*DepartmentForListAllOfStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DepartmentForList) SetStatus(v DepartmentForListAllOfStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DepartmentForList) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *DepartmentForList) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *DepartmentForList) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetTotalResources

`func (o *DepartmentForList) GetTotalResources() TotalResources`

GetTotalResources returns the TotalResources field if non-nil, zero value otherwise.

### GetTotalResourcesOk

`func (o *DepartmentForList) GetTotalResourcesOk() (*TotalResources, bool)`

GetTotalResourcesOk returns a tuple with the TotalResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalResources

`func (o *DepartmentForList) SetTotalResources(v TotalResources)`

SetTotalResources sets TotalResources field to given value.


### GetResources

`func (o *DepartmentForList) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentForList) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentForList) SetResources(v []Resources)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


