# WorkloadTypeAndScope

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadType** | [**GVK**](GVK.md) |  | 
**Scope** | [**ScopeType**](ScopeType.md) |  | 
**ClusterId** | Pointer to **NullableString** | The id of the cluster the policy applies to. Required when the scope is cluster. | [optional] 
**DepartmentId** | Pointer to **NullableString** | The id of the department the policy applies to. Required when the scope is department. | [optional] 
**ProjectId** | Pointer to **NullableString** | The id of the project the policy applies to. Required when the scope is project. | [optional] 

## Methods

### NewWorkloadTypeAndScope

`func NewWorkloadTypeAndScope(workloadType GVK, scope ScopeType, ) *WorkloadTypeAndScope`

NewWorkloadTypeAndScope instantiates a new WorkloadTypeAndScope object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeAndScopeWithDefaults

`func NewWorkloadTypeAndScopeWithDefaults() *WorkloadTypeAndScope`

NewWorkloadTypeAndScopeWithDefaults instantiates a new WorkloadTypeAndScope object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadType

`func (o *WorkloadTypeAndScope) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *WorkloadTypeAndScope) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *WorkloadTypeAndScope) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.


### GetScope

`func (o *WorkloadTypeAndScope) GetScope() ScopeType`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *WorkloadTypeAndScope) GetScopeOk() (*ScopeType, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *WorkloadTypeAndScope) SetScope(v ScopeType)`

SetScope sets Scope field to given value.


### GetClusterId

`func (o *WorkloadTypeAndScope) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadTypeAndScope) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadTypeAndScope) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *WorkloadTypeAndScope) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### SetClusterIdNil

`func (o *WorkloadTypeAndScope) SetClusterIdNil(b bool)`

 SetClusterIdNil sets the value for ClusterId to be an explicit nil

### UnsetClusterId
`func (o *WorkloadTypeAndScope) UnsetClusterId()`

UnsetClusterId ensures that no value is present for ClusterId, not even an explicit nil
### GetDepartmentId

`func (o *WorkloadTypeAndScope) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *WorkloadTypeAndScope) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *WorkloadTypeAndScope) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *WorkloadTypeAndScope) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### SetDepartmentIdNil

`func (o *WorkloadTypeAndScope) SetDepartmentIdNil(b bool)`

 SetDepartmentIdNil sets the value for DepartmentId to be an explicit nil

### UnsetDepartmentId
`func (o *WorkloadTypeAndScope) UnsetDepartmentId()`

UnsetDepartmentId ensures that no value is present for DepartmentId, not even an explicit nil
### GetProjectId

`func (o *WorkloadTypeAndScope) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadTypeAndScope) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadTypeAndScope) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *WorkloadTypeAndScope) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### SetProjectIdNil

`func (o *WorkloadTypeAndScope) SetProjectIdNil(b bool)`

 SetProjectIdNil sets the value for ProjectId to be an explicit nil

### UnsetProjectId
`func (o *WorkloadTypeAndScope) UnsetProjectId()`

UnsetProjectId ensures that no value is present for ProjectId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


