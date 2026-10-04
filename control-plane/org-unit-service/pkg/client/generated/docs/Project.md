# Project

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | **string** | The description of the project. | 
**SchedulingRules** | Pointer to [**NullableSchedulingRules**](SchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | An ordered list of node pools that defines the default scheduling preference for workloads in this project. The Scheduler will attempt to schedule workloads on node pools in this order. This setting is enforced if no node pool list is defined in the workload policy. | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 
**Id** | **string** |  | 
**Name** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**ParentId** | **string** | department parent uuid | 
**RequestedNamespace** | Pointer to **NullableString** | the project&#39;s namespace only if requested | [optional] 
**EnforceRunaiScheduler** | **bool** | a flag to enforce that all pods in this namespace will be scheduled with runai-scheduler. Supported for cluster v2.21+ | 
**Status** | [**ProjectStatus**](ProjectStatus.md) |  | 
**TotalResources** | [**TotalResources**](TotalResources.md) |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 
**Parent** | Pointer to [**NullableDepartmentProjectParent**](DepartmentProjectParent.md) |  | [optional] 
**Effective** | [**ProjectEffective**](ProjectEffective.md) |  | 
**OvertimeData** | Pointer to [**OvertimeData**](OvertimeData.md) |  | [optional] 
**Deletion** | Pointer to [**NullableDeletion**](Deletion.md) |  | [optional] 
**Resources** | [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | 

## Methods

### NewProject

`func NewProject(description string, id string, name string, clusterId string, parentId string, enforceRunaiScheduler bool, status ProjectStatus, totalResources TotalResources, createdAt time.Time, updatedAt time.Time, effective ProjectEffective, resources []Resources, ) *Project`

NewProject instantiates a new Project object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectWithDefaults

`func NewProjectWithDefaults() *Project`

NewProjectWithDefaults instantiates a new Project object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *Project) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Project) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Project) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetSchedulingRules

`func (o *Project) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *Project) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *Project) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *Project) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *Project) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *Project) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *Project) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *Project) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *Project) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *Project) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *Project) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *Project) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *Project) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *Project) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *Project) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *Project) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetId

`func (o *Project) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Project) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Project) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *Project) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Project) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Project) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *Project) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Project) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Project) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetParentId

`func (o *Project) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *Project) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *Project) SetParentId(v string)`

SetParentId sets ParentId field to given value.


### GetRequestedNamespace

`func (o *Project) GetRequestedNamespace() string`

GetRequestedNamespace returns the RequestedNamespace field if non-nil, zero value otherwise.

### GetRequestedNamespaceOk

`func (o *Project) GetRequestedNamespaceOk() (*string, bool)`

GetRequestedNamespaceOk returns a tuple with the RequestedNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNamespace

`func (o *Project) SetRequestedNamespace(v string)`

SetRequestedNamespace sets RequestedNamespace field to given value.

### HasRequestedNamespace

`func (o *Project) HasRequestedNamespace() bool`

HasRequestedNamespace returns a boolean if a field has been set.

### SetRequestedNamespaceNil

`func (o *Project) SetRequestedNamespaceNil(b bool)`

 SetRequestedNamespaceNil sets the value for RequestedNamespace to be an explicit nil

### UnsetRequestedNamespace
`func (o *Project) UnsetRequestedNamespace()`

UnsetRequestedNamespace ensures that no value is present for RequestedNamespace, not even an explicit nil
### GetEnforceRunaiScheduler

`func (o *Project) GetEnforceRunaiScheduler() bool`

GetEnforceRunaiScheduler returns the EnforceRunaiScheduler field if non-nil, zero value otherwise.

### GetEnforceRunaiSchedulerOk

`func (o *Project) GetEnforceRunaiSchedulerOk() (*bool, bool)`

GetEnforceRunaiSchedulerOk returns a tuple with the EnforceRunaiScheduler field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforceRunaiScheduler

`func (o *Project) SetEnforceRunaiScheduler(v bool)`

SetEnforceRunaiScheduler sets EnforceRunaiScheduler field to given value.


### GetStatus

`func (o *Project) GetStatus() ProjectStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Project) GetStatusOk() (*ProjectStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Project) SetStatus(v ProjectStatus)`

SetStatus sets Status field to given value.


### GetTotalResources

`func (o *Project) GetTotalResources() TotalResources`

GetTotalResources returns the TotalResources field if non-nil, zero value otherwise.

### GetTotalResourcesOk

`func (o *Project) GetTotalResourcesOk() (*TotalResources, bool)`

GetTotalResourcesOk returns a tuple with the TotalResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalResources

`func (o *Project) SetTotalResources(v TotalResources)`

SetTotalResources sets TotalResources field to given value.


### GetCreatedAt

`func (o *Project) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Project) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Project) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *Project) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Project) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Project) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetCreatedBy

`func (o *Project) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Project) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Project) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *Project) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *Project) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *Project) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *Project) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *Project) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetParent

`func (o *Project) GetParent() DepartmentProjectParent`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *Project) GetParentOk() (*DepartmentProjectParent, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *Project) SetParent(v DepartmentProjectParent)`

SetParent sets Parent field to given value.

### HasParent

`func (o *Project) HasParent() bool`

HasParent returns a boolean if a field has been set.

### SetParentNil

`func (o *Project) SetParentNil(b bool)`

 SetParentNil sets the value for Parent to be an explicit nil

### UnsetParent
`func (o *Project) UnsetParent()`

UnsetParent ensures that no value is present for Parent, not even an explicit nil
### GetEffective

`func (o *Project) GetEffective() ProjectEffective`

GetEffective returns the Effective field if non-nil, zero value otherwise.

### GetEffectiveOk

`func (o *Project) GetEffectiveOk() (*ProjectEffective, bool)`

GetEffectiveOk returns a tuple with the Effective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffective

`func (o *Project) SetEffective(v ProjectEffective)`

SetEffective sets Effective field to given value.


### GetOvertimeData

`func (o *Project) GetOvertimeData() OvertimeData`

GetOvertimeData returns the OvertimeData field if non-nil, zero value otherwise.

### GetOvertimeDataOk

`func (o *Project) GetOvertimeDataOk() (*OvertimeData, bool)`

GetOvertimeDataOk returns a tuple with the OvertimeData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOvertimeData

`func (o *Project) SetOvertimeData(v OvertimeData)`

SetOvertimeData sets OvertimeData field to given value.

### HasOvertimeData

`func (o *Project) HasOvertimeData() bool`

HasOvertimeData returns a boolean if a field has been set.

### GetDeletion

`func (o *Project) GetDeletion() Deletion`

GetDeletion returns the Deletion field if non-nil, zero value otherwise.

### GetDeletionOk

`func (o *Project) GetDeletionOk() (*Deletion, bool)`

GetDeletionOk returns a tuple with the Deletion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletion

`func (o *Project) SetDeletion(v Deletion)`

SetDeletion sets Deletion field to given value.

### HasDeletion

`func (o *Project) HasDeletion() bool`

HasDeletion returns a boolean if a field has been set.

### SetDeletionNil

`func (o *Project) SetDeletionNil(b bool)`

 SetDeletionNil sets the value for Deletion to be an explicit nil

### UnsetDeletion
`func (o *Project) UnsetDeletion()`

UnsetDeletion ensures that no value is present for Deletion, not even an explicit nil
### GetResources

`func (o *Project) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *Project) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *Project) SetResources(v []Resources)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


