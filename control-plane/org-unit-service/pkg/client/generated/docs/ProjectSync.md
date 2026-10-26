# ProjectSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Name** | **string** |  | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**RequestedNamespace** | Pointer to **NullableString** | the project&#39;s namespace only if requested | [optional] 
**EnforceRunaiScheduler** | Pointer to **bool** | a flag to enforce that all pods in this namespace will be scheduled with runai-scheduler. Supported for cluster v2.21+ | [optional] 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**Parent** | [**NullableDepartmentProjectParent**](DepartmentProjectParent.md) |  | 
**SchedulingRules** | Pointer to [**NullableSchedulingRules**](SchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | default order of node pools for workloads. will be enforced if no list is defined in workload policy | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 
**Resources** | Pointer to [**[]ResourcesSync**](ResourcesSync.md) | Resources assigned to this Organization per Node Pool | [optional] 
**Deletion** | Pointer to [**NullableDeletion**](Deletion.md) |  | [optional] 
**Permissions** | Pointer to [**ProjectSyncPermissions**](ProjectSyncPermissions.md) |  | [optional] 

## Methods

### NewProjectSync

`func NewProjectSync(id string, name string, createdAt time.Time, updatedAt time.Time, parent NullableDepartmentProjectParent, ) *ProjectSync`

NewProjectSync instantiates a new ProjectSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectSyncWithDefaults

`func NewProjectSyncWithDefaults() *ProjectSync`

NewProjectSyncWithDefaults instantiates a new ProjectSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ProjectSync) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectSync) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectSync) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *ProjectSync) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectSync) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectSync) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *ProjectSync) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ProjectSync) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ProjectSync) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *ProjectSync) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetRequestedNamespace

`func (o *ProjectSync) GetRequestedNamespace() string`

GetRequestedNamespace returns the RequestedNamespace field if non-nil, zero value otherwise.

### GetRequestedNamespaceOk

`func (o *ProjectSync) GetRequestedNamespaceOk() (*string, bool)`

GetRequestedNamespaceOk returns a tuple with the RequestedNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNamespace

`func (o *ProjectSync) SetRequestedNamespace(v string)`

SetRequestedNamespace sets RequestedNamespace field to given value.

### HasRequestedNamespace

`func (o *ProjectSync) HasRequestedNamespace() bool`

HasRequestedNamespace returns a boolean if a field has been set.

### SetRequestedNamespaceNil

`func (o *ProjectSync) SetRequestedNamespaceNil(b bool)`

 SetRequestedNamespaceNil sets the value for RequestedNamespace to be an explicit nil

### UnsetRequestedNamespace
`func (o *ProjectSync) UnsetRequestedNamespace()`

UnsetRequestedNamespace ensures that no value is present for RequestedNamespace, not even an explicit nil
### GetEnforceRunaiScheduler

`func (o *ProjectSync) GetEnforceRunaiScheduler() bool`

GetEnforceRunaiScheduler returns the EnforceRunaiScheduler field if non-nil, zero value otherwise.

### GetEnforceRunaiSchedulerOk

`func (o *ProjectSync) GetEnforceRunaiSchedulerOk() (*bool, bool)`

GetEnforceRunaiSchedulerOk returns a tuple with the EnforceRunaiScheduler field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforceRunaiScheduler

`func (o *ProjectSync) SetEnforceRunaiScheduler(v bool)`

SetEnforceRunaiScheduler sets EnforceRunaiScheduler field to given value.

### HasEnforceRunaiScheduler

`func (o *ProjectSync) HasEnforceRunaiScheduler() bool`

HasEnforceRunaiScheduler returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ProjectSync) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ProjectSync) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ProjectSync) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *ProjectSync) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *ProjectSync) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *ProjectSync) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *ProjectSync) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *ProjectSync) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *ProjectSync) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *ProjectSync) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *ProjectSync) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *ProjectSync) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetParent

`func (o *ProjectSync) GetParent() DepartmentProjectParent`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *ProjectSync) GetParentOk() (*DepartmentProjectParent, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *ProjectSync) SetParent(v DepartmentProjectParent)`

SetParent sets Parent field to given value.


### SetParentNil

`func (o *ProjectSync) SetParentNil(b bool)`

 SetParentNil sets the value for Parent to be an explicit nil

### UnsetParent
`func (o *ProjectSync) UnsetParent()`

UnsetParent ensures that no value is present for Parent, not even an explicit nil
### GetSchedulingRules

`func (o *ProjectSync) GetSchedulingRules() SchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *ProjectSync) GetSchedulingRulesOk() (*SchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *ProjectSync) SetSchedulingRules(v SchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *ProjectSync) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *ProjectSync) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *ProjectSync) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *ProjectSync) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *ProjectSync) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *ProjectSync) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *ProjectSync) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *ProjectSync) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *ProjectSync) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *ProjectSync) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *ProjectSync) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *ProjectSync) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *ProjectSync) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetResources

`func (o *ProjectSync) GetResources() []ResourcesSync`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ProjectSync) GetResourcesOk() (*[]ResourcesSync, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ProjectSync) SetResources(v []ResourcesSync)`

SetResources sets Resources field to given value.

### HasResources

`func (o *ProjectSync) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetDeletion

`func (o *ProjectSync) GetDeletion() Deletion`

GetDeletion returns the Deletion field if non-nil, zero value otherwise.

### GetDeletionOk

`func (o *ProjectSync) GetDeletionOk() (*Deletion, bool)`

GetDeletionOk returns a tuple with the Deletion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletion

`func (o *ProjectSync) SetDeletion(v Deletion)`

SetDeletion sets Deletion field to given value.

### HasDeletion

`func (o *ProjectSync) HasDeletion() bool`

HasDeletion returns a boolean if a field has been set.

### SetDeletionNil

`func (o *ProjectSync) SetDeletionNil(b bool)`

 SetDeletionNil sets the value for Deletion to be an explicit nil

### UnsetDeletion
`func (o *ProjectSync) UnsetDeletion()`

UnsetDeletion ensures that no value is present for Deletion, not even an explicit nil
### GetPermissions

`func (o *ProjectSync) GetPermissions() ProjectSyncPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *ProjectSync) GetPermissionsOk() (*ProjectSyncPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *ProjectSync) SetPermissions(v ProjectSyncPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *ProjectSync) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


