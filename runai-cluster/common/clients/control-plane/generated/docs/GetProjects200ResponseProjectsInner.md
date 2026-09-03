# GetProjects200ResponseProjectsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SchedulingRules** | Pointer to [**NullableDataUpdatableFieldsSchedulingRules**](DataUpdatableFieldsSchedulingRules.md) |  | [optional] 
**DefaultNodePools** | Pointer to **[]string** | default order of node pools for workloads. will be enforced if no list is defined in workload policy | [optional] 
**NodeTypes** | Pointer to [**NodeTypesPerWorkload**](NodeTypesPerWorkload.md) |  | [optional] 
**Resources** | Pointer to [**[]Resources**](Resources.md) | Resources assigned to this Organization per Node Pool | [optional] 
**Name** | **string** |  | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**Id** | **string** |  | 
**ParentId** | Pointer to **NullableString** | department parent uuid | [optional] 
**RequestedNamespace** | Pointer to **NullableString** | the project&#39;s namespace only if requested | [optional] 
**Status** | [**ProjectAllOfStatus**](ProjectAllOfStatus.md) |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 
**Parent** | Pointer to [**NullableDepartmentProjectParent**](DepartmentProjectParent.md) |  | [optional] 

## Methods

### NewGetProjects200ResponseProjectsInner

`func NewGetProjects200ResponseProjectsInner(name string, id string, status ProjectAllOfStatus, createdAt time.Time, updatedAt time.Time, ) *GetProjects200ResponseProjectsInner`

NewGetProjects200ResponseProjectsInner instantiates a new GetProjects200ResponseProjectsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetProjects200ResponseProjectsInnerWithDefaults

`func NewGetProjects200ResponseProjectsInnerWithDefaults() *GetProjects200ResponseProjectsInner`

NewGetProjects200ResponseProjectsInnerWithDefaults instantiates a new GetProjects200ResponseProjectsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSchedulingRules

`func (o *GetProjects200ResponseProjectsInner) GetSchedulingRules() DataUpdatableFieldsSchedulingRules`

GetSchedulingRules returns the SchedulingRules field if non-nil, zero value otherwise.

### GetSchedulingRulesOk

`func (o *GetProjects200ResponseProjectsInner) GetSchedulingRulesOk() (*DataUpdatableFieldsSchedulingRules, bool)`

GetSchedulingRulesOk returns a tuple with the SchedulingRules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingRules

`func (o *GetProjects200ResponseProjectsInner) SetSchedulingRules(v DataUpdatableFieldsSchedulingRules)`

SetSchedulingRules sets SchedulingRules field to given value.

### HasSchedulingRules

`func (o *GetProjects200ResponseProjectsInner) HasSchedulingRules() bool`

HasSchedulingRules returns a boolean if a field has been set.

### SetSchedulingRulesNil

`func (o *GetProjects200ResponseProjectsInner) SetSchedulingRulesNil(b bool)`

 SetSchedulingRulesNil sets the value for SchedulingRules to be an explicit nil

### UnsetSchedulingRules
`func (o *GetProjects200ResponseProjectsInner) UnsetSchedulingRules()`

UnsetSchedulingRules ensures that no value is present for SchedulingRules, not even an explicit nil
### GetDefaultNodePools

`func (o *GetProjects200ResponseProjectsInner) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *GetProjects200ResponseProjectsInner) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *GetProjects200ResponseProjectsInner) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *GetProjects200ResponseProjectsInner) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### SetDefaultNodePoolsNil

`func (o *GetProjects200ResponseProjectsInner) SetDefaultNodePoolsNil(b bool)`

 SetDefaultNodePoolsNil sets the value for DefaultNodePools to be an explicit nil

### UnsetDefaultNodePools
`func (o *GetProjects200ResponseProjectsInner) UnsetDefaultNodePools()`

UnsetDefaultNodePools ensures that no value is present for DefaultNodePools, not even an explicit nil
### GetNodeTypes

`func (o *GetProjects200ResponseProjectsInner) GetNodeTypes() NodeTypesPerWorkload`

GetNodeTypes returns the NodeTypes field if non-nil, zero value otherwise.

### GetNodeTypesOk

`func (o *GetProjects200ResponseProjectsInner) GetNodeTypesOk() (*NodeTypesPerWorkload, bool)`

GetNodeTypesOk returns a tuple with the NodeTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeTypes

`func (o *GetProjects200ResponseProjectsInner) SetNodeTypes(v NodeTypesPerWorkload)`

SetNodeTypes sets NodeTypes field to given value.

### HasNodeTypes

`func (o *GetProjects200ResponseProjectsInner) HasNodeTypes() bool`

HasNodeTypes returns a boolean if a field has been set.

### GetResources

`func (o *GetProjects200ResponseProjectsInner) GetResources() []Resources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *GetProjects200ResponseProjectsInner) GetResourcesOk() (*[]Resources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *GetProjects200ResponseProjectsInner) SetResources(v []Resources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *GetProjects200ResponseProjectsInner) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetName

`func (o *GetProjects200ResponseProjectsInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GetProjects200ResponseProjectsInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GetProjects200ResponseProjectsInner) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *GetProjects200ResponseProjectsInner) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *GetProjects200ResponseProjectsInner) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *GetProjects200ResponseProjectsInner) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *GetProjects200ResponseProjectsInner) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetId

`func (o *GetProjects200ResponseProjectsInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetProjects200ResponseProjectsInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetProjects200ResponseProjectsInner) SetId(v string)`

SetId sets Id field to given value.


### GetParentId

`func (o *GetProjects200ResponseProjectsInner) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *GetProjects200ResponseProjectsInner) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *GetProjects200ResponseProjectsInner) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *GetProjects200ResponseProjectsInner) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### SetParentIdNil

`func (o *GetProjects200ResponseProjectsInner) SetParentIdNil(b bool)`

 SetParentIdNil sets the value for ParentId to be an explicit nil

### UnsetParentId
`func (o *GetProjects200ResponseProjectsInner) UnsetParentId()`

UnsetParentId ensures that no value is present for ParentId, not even an explicit nil
### GetRequestedNamespace

`func (o *GetProjects200ResponseProjectsInner) GetRequestedNamespace() string`

GetRequestedNamespace returns the RequestedNamespace field if non-nil, zero value otherwise.

### GetRequestedNamespaceOk

`func (o *GetProjects200ResponseProjectsInner) GetRequestedNamespaceOk() (*string, bool)`

GetRequestedNamespaceOk returns a tuple with the RequestedNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNamespace

`func (o *GetProjects200ResponseProjectsInner) SetRequestedNamespace(v string)`

SetRequestedNamespace sets RequestedNamespace field to given value.

### HasRequestedNamespace

`func (o *GetProjects200ResponseProjectsInner) HasRequestedNamespace() bool`

HasRequestedNamespace returns a boolean if a field has been set.

### SetRequestedNamespaceNil

`func (o *GetProjects200ResponseProjectsInner) SetRequestedNamespaceNil(b bool)`

 SetRequestedNamespaceNil sets the value for RequestedNamespace to be an explicit nil

### UnsetRequestedNamespace
`func (o *GetProjects200ResponseProjectsInner) UnsetRequestedNamespace()`

UnsetRequestedNamespace ensures that no value is present for RequestedNamespace, not even an explicit nil
### GetStatus

`func (o *GetProjects200ResponseProjectsInner) GetStatus() ProjectAllOfStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *GetProjects200ResponseProjectsInner) GetStatusOk() (*ProjectAllOfStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *GetProjects200ResponseProjectsInner) SetStatus(v ProjectAllOfStatus)`

SetStatus sets Status field to given value.


### GetCreatedAt

`func (o *GetProjects200ResponseProjectsInner) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GetProjects200ResponseProjectsInner) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GetProjects200ResponseProjectsInner) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *GetProjects200ResponseProjectsInner) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GetProjects200ResponseProjectsInner) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GetProjects200ResponseProjectsInner) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetCreatedBy

`func (o *GetProjects200ResponseProjectsInner) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *GetProjects200ResponseProjectsInner) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *GetProjects200ResponseProjectsInner) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *GetProjects200ResponseProjectsInner) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *GetProjects200ResponseProjectsInner) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *GetProjects200ResponseProjectsInner) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *GetProjects200ResponseProjectsInner) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *GetProjects200ResponseProjectsInner) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetParent

`func (o *GetProjects200ResponseProjectsInner) GetParent() DepartmentProjectParent`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *GetProjects200ResponseProjectsInner) GetParentOk() (*DepartmentProjectParent, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *GetProjects200ResponseProjectsInner) SetParent(v DepartmentProjectParent)`

SetParent sets Parent field to given value.

### HasParent

`func (o *GetProjects200ResponseProjectsInner) HasParent() bool`

HasParent returns a boolean if a field has been set.

### SetParentNil

`func (o *GetProjects200ResponseProjectsInner) SetParentNil(b bool)`

 SetParentNil sets the value for Parent to be an explicit nil

### UnsetParent
`func (o *GetProjects200ResponseProjectsInner) UnsetParent()`

UnsetParent ensures that no value is present for Parent, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


