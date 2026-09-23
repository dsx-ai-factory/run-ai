# Department

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | **string** | The description of the department. | 
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
**ProjectsAggregatedResources** | Pointer to [**[]AggregatedResources**](AggregatedResources.md) |  | [optional] 
**Status** | Pointer to [**DepartmentAllOfStatus**](DepartmentAllOfStatus.md) |  | [optional] 
**TotalResources** | [**TotalResources**](TotalResources.md) |  | 
**Resources** | [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | 

## Methods

### NewDepartment

`func NewDepartment(description string, name string, clusterId string, id string, createdAt time.Time, updatedAt time.Time, totalResources TotalResources, resources []Resources, ) *Department`

NewDepartment instantiates a new Department object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentWithDefaults

`func NewDepartmentWithDefaults() *Department`

NewDepartmentWithDefaults instantiates a new Department object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *Department) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Department) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Department) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetSchedulingRules

`func (o *Department) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *Department) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *Department) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *Department) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *Department) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *Department) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *Department) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *Department) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *Department) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *Department) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *Department) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *Department) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *Department) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *Department) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *Department) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *Department) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetName

`func (o *Department) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Department) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Department) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *Department) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Department) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Department) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetOvertimeData

`func (o *Department) GetOvertimeData() OvertimeData`

GetOvertimeData returns the OvertimeData field if non-nil, zero value otherwise.

### GetOvertimeDataOk

`func (o *Department) GetOvertimeDataOk() (*OvertimeData, bool)`

GetOvertimeDataOk returns a tuple with the OvertimeData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOvertimeData

`func (o *Department) SetOvertimeData(v OvertimeData)`

SetOvertimeData sets OvertimeData field to given value.

### HasOvertimeData

`func (o *Department) HasOvertimeData() bool`

HasOvertimeData returns a boolean if a field has been set.

### GetId

`func (o *Department) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Department) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Department) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *Department) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Department) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Department) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *Department) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Department) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Department) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetCreatedBy

`func (o *Department) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Department) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Department) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *Department) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *Department) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *Department) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *Department) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *Department) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetChildren

`func (o *Department) GetChildren() []DepartmentProjectChild`

GetChildren returns the Children field if non-nil, zero value otherwise.

### GetChildrenOk

`func (o *Department) GetChildrenOk() (*[]DepartmentProjectChild, bool)`

GetChildrenOk returns a tuple with the Children field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildren

`func (o *Department) SetChildren(v []DepartmentProjectChild)`

SetChildren sets Children field to given value.

### HasChildren

`func (o *Department) HasChildren() bool`

HasChildren returns a boolean if a field has been set.

### SetChildrenNil

`func (o *Department) SetChildrenNil(b bool)`

 SetChildrenNil sets the value for Children to be an explicit nil

### UnsetChildren
`func (o *Department) UnsetChildren()`

UnsetChildren ensures that no value is present for Children, not even an explicit nil
### GetProjectsAggregatedResources

`func (o *Department) GetProjectsAggregatedResources() []AggregatedResources`

GetProjectsAggregatedResources returns the ProjectsAggregatedResources field if non-nil, zero value otherwise.

### GetProjectsAggregatedResourcesOk

`func (o *Department) GetProjectsAggregatedResourcesOk() (*[]AggregatedResources, bool)`

GetProjectsAggregatedResourcesOk returns a tuple with the ProjectsAggregatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectsAggregatedResources

`func (o *Department) SetProjectsAggregatedResources(v []AggregatedResources)`

SetProjectsAggregatedResources sets ProjectsAggregatedResources field to given value.

### HasProjectsAggregatedResources

`func (o *Department) HasProjectsAggregatedResources() bool`

HasProjectsAggregatedResources returns a boolean if a field has been set.

### SetProjectsAggregatedResourcesNil

`func (o *Department) SetProjectsAggregatedResourcesNil(b bool)`

 SetProjectsAggregatedResourcesNil sets the value for ProjectsAggregatedResources to be an explicit nil

### UnsetProjectsAggregatedResources
`func (o *Department) UnsetProjectsAggregatedResources()`

UnsetProjectsAggregatedResources ensures that no value is present for ProjectsAggregatedResources, not even an explicit nil
### GetStatus

`func (o *Department) GetStatus() DepartmentAllOfStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Department) GetStatusOk() (*DepartmentAllOfStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Department) SetStatus(v DepartmentAllOfStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Department) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetTotalResources

`func (o *Department) GetTotalResources() TotalResources`

GetTotalResources returns the TotalResources field if non-nil, zero value otherwise.

### GetTotalResourcesOk

`func (o *Department) GetTotalResourcesOk() (*TotalResources, bool)`

GetTotalResourcesOk returns a tuple with the TotalResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalResources

`func (o *Department) SetTotalResources(v TotalResources)`

SetTotalResources sets TotalResources field to given value.


### GetResources

`func (o *Department) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *Department) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *Department) SetResources(v []Resources)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


