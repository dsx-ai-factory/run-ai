# WorkloadMeta

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the workload. | 
**TenantId** | **int32** | The id of the tenant. | 
**ProjectId** | **int32** | The id of the project. | 
**ClusterId** | **string** | The id of the cluster. | 
**Name** | **string** | The name of the workspace or training | 
**DesiredState** | [**WorkloadState**](WorkloadState.md) |  | 
**CurrentState** | [**WorkloadState**](WorkloadState.md) |  | 
**CreatedBy** | **string** | The user who created the workspace or training. | 
**UpdatedBy** | **string** | The recent user who updated the workspace or training. | 

## Methods

### NewWorkloadMeta

`func NewWorkloadMeta(id string, tenantId int32, projectId int32, clusterId string, name string, desiredState WorkloadState, currentState WorkloadState, createdBy string, updatedBy string, ) *WorkloadMeta`

NewWorkloadMeta instantiates a new WorkloadMeta object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadMetaWithDefaults

`func NewWorkloadMetaWithDefaults() *WorkloadMeta`

NewWorkloadMetaWithDefaults instantiates a new WorkloadMeta object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadMeta) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadMeta) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadMeta) SetId(v string)`

SetId sets Id field to given value.


### GetTenantId

`func (o *WorkloadMeta) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *WorkloadMeta) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *WorkloadMeta) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetProjectId

`func (o *WorkloadMeta) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadMeta) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadMeta) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *WorkloadMeta) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadMeta) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadMeta) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetName

`func (o *WorkloadMeta) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadMeta) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadMeta) SetName(v string)`

SetName sets Name field to given value.


### GetDesiredState

`func (o *WorkloadMeta) GetDesiredState() WorkloadState`

GetDesiredState returns the DesiredState field if non-nil, zero value otherwise.

### GetDesiredStateOk

`func (o *WorkloadMeta) GetDesiredStateOk() (*WorkloadState, bool)`

GetDesiredStateOk returns a tuple with the DesiredState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredState

`func (o *WorkloadMeta) SetDesiredState(v WorkloadState)`

SetDesiredState sets DesiredState field to given value.


### GetCurrentState

`func (o *WorkloadMeta) GetCurrentState() WorkloadState`

GetCurrentState returns the CurrentState field if non-nil, zero value otherwise.

### GetCurrentStateOk

`func (o *WorkloadMeta) GetCurrentStateOk() (*WorkloadState, bool)`

GetCurrentStateOk returns a tuple with the CurrentState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentState

`func (o *WorkloadMeta) SetCurrentState(v WorkloadState)`

SetCurrentState sets CurrentState field to given value.


### GetCreatedBy

`func (o *WorkloadMeta) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *WorkloadMeta) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *WorkloadMeta) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetUpdatedBy

`func (o *WorkloadMeta) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *WorkloadMeta) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *WorkloadMeta) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


