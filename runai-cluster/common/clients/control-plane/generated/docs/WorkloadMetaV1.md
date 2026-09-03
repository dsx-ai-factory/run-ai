# WorkloadMetaV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | A unique ID of the workload. | 
**TenantId** | **int32** | The id of the tenant. | 
**ProjectId** | **int32** | The id of the project. | 
**ClusterId** | **string** | The id of the cluster. | 
**Name** | **string** | The name of the workload. | 
**DesiredState** | [**WorkloadState**](WorkloadState.md) |  | 
**CurrentState** | [**WorkloadState**](WorkloadState.md) |  | 
**CreatedBy** | **string** | The user who created the workspace or training. | 
**UpdatedBy** | **string** | The recent user who updated the workspace or training. | 

## Methods

### NewWorkloadMetaV1

`func NewWorkloadMetaV1(id string, tenantId int32, projectId int32, clusterId string, name string, desiredState WorkloadState, currentState WorkloadState, createdBy string, updatedBy string, ) *WorkloadMetaV1`

NewWorkloadMetaV1 instantiates a new WorkloadMetaV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadMetaV1WithDefaults

`func NewWorkloadMetaV1WithDefaults() *WorkloadMetaV1`

NewWorkloadMetaV1WithDefaults instantiates a new WorkloadMetaV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadMetaV1) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadMetaV1) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadMetaV1) SetId(v string)`

SetId sets Id field to given value.


### GetTenantId

`func (o *WorkloadMetaV1) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *WorkloadMetaV1) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *WorkloadMetaV1) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetProjectId

`func (o *WorkloadMetaV1) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadMetaV1) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadMetaV1) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *WorkloadMetaV1) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadMetaV1) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadMetaV1) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetName

`func (o *WorkloadMetaV1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadMetaV1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadMetaV1) SetName(v string)`

SetName sets Name field to given value.


### GetDesiredState

`func (o *WorkloadMetaV1) GetDesiredState() WorkloadState`

GetDesiredState returns the DesiredState field if non-nil, zero value otherwise.

### GetDesiredStateOk

`func (o *WorkloadMetaV1) GetDesiredStateOk() (*WorkloadState, bool)`

GetDesiredStateOk returns a tuple with the DesiredState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredState

`func (o *WorkloadMetaV1) SetDesiredState(v WorkloadState)`

SetDesiredState sets DesiredState field to given value.


### GetCurrentState

`func (o *WorkloadMetaV1) GetCurrentState() WorkloadState`

GetCurrentState returns the CurrentState field if non-nil, zero value otherwise.

### GetCurrentStateOk

`func (o *WorkloadMetaV1) GetCurrentStateOk() (*WorkloadState, bool)`

GetCurrentStateOk returns a tuple with the CurrentState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentState

`func (o *WorkloadMetaV1) SetCurrentState(v WorkloadState)`

SetCurrentState sets CurrentState field to given value.


### GetCreatedBy

`func (o *WorkloadMetaV1) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *WorkloadMetaV1) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *WorkloadMetaV1) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetUpdatedBy

`func (o *WorkloadMetaV1) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *WorkloadMetaV1) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *WorkloadMetaV1) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


