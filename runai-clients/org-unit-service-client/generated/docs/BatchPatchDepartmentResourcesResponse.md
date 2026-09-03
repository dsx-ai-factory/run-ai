# BatchPatchDepartmentResourcesResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique ID of the batch operation. | 
**Succeeded** | Pointer to [**[]DepartmentNodePoolResources**](DepartmentNodePoolResources.md) | Successfully patched department node pool resources results. | [optional] 
**Failed** | Pointer to [**[]FailedNodePoolResourcesPatchRequest**](FailedNodePoolResourcesPatchRequest.md) | Failed patch operations with the original payload and error details. | [optional] 
**TotalRequested** | **int32** | Total number of items requested in the batch. | 
**TotalSuccessful** | **int32** | Number of items successfully processed. | 
**TotalFailed** | **int32** | Number of items that failed. | 

## Methods

### NewBatchPatchDepartmentResourcesResponse

`func NewBatchPatchDepartmentResourcesResponse(id string, totalRequested int32, totalSuccessful int32, totalFailed int32, ) *BatchPatchDepartmentResourcesResponse`

NewBatchPatchDepartmentResourcesResponse instantiates a new BatchPatchDepartmentResourcesResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBatchPatchDepartmentResourcesResponseWithDefaults

`func NewBatchPatchDepartmentResourcesResponseWithDefaults() *BatchPatchDepartmentResourcesResponse`

NewBatchPatchDepartmentResourcesResponseWithDefaults instantiates a new BatchPatchDepartmentResourcesResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *BatchPatchDepartmentResourcesResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *BatchPatchDepartmentResourcesResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *BatchPatchDepartmentResourcesResponse) SetId(v string)`

SetId sets Id field to given value.


### GetSucceeded

`func (o *BatchPatchDepartmentResourcesResponse) GetSucceeded() []DepartmentNodePoolResources`

GetSucceeded returns the Succeeded field if non-nil, zero value otherwise.

### GetSucceededOk

`func (o *BatchPatchDepartmentResourcesResponse) GetSucceededOk() (*[]DepartmentNodePoolResources, bool)`

GetSucceededOk returns a tuple with the Succeeded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSucceeded

`func (o *BatchPatchDepartmentResourcesResponse) SetSucceeded(v []DepartmentNodePoolResources)`

SetSucceeded sets Succeeded field to given value.

### HasSucceeded

`func (o *BatchPatchDepartmentResourcesResponse) HasSucceeded() bool`

HasSucceeded returns a boolean if a field has been set.

### GetFailed

`func (o *BatchPatchDepartmentResourcesResponse) GetFailed() []FailedNodePoolResourcesPatchRequest`

GetFailed returns the Failed field if non-nil, zero value otherwise.

### GetFailedOk

`func (o *BatchPatchDepartmentResourcesResponse) GetFailedOk() (*[]FailedNodePoolResourcesPatchRequest, bool)`

GetFailedOk returns a tuple with the Failed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailed

`func (o *BatchPatchDepartmentResourcesResponse) SetFailed(v []FailedNodePoolResourcesPatchRequest)`

SetFailed sets Failed field to given value.

### HasFailed

`func (o *BatchPatchDepartmentResourcesResponse) HasFailed() bool`

HasFailed returns a boolean if a field has been set.

### GetTotalRequested

`func (o *BatchPatchDepartmentResourcesResponse) GetTotalRequested() int32`

GetTotalRequested returns the TotalRequested field if non-nil, zero value otherwise.

### GetTotalRequestedOk

`func (o *BatchPatchDepartmentResourcesResponse) GetTotalRequestedOk() (*int32, bool)`

GetTotalRequestedOk returns a tuple with the TotalRequested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequested

`func (o *BatchPatchDepartmentResourcesResponse) SetTotalRequested(v int32)`

SetTotalRequested sets TotalRequested field to given value.


### GetTotalSuccessful

`func (o *BatchPatchDepartmentResourcesResponse) GetTotalSuccessful() int32`

GetTotalSuccessful returns the TotalSuccessful field if non-nil, zero value otherwise.

### GetTotalSuccessfulOk

`func (o *BatchPatchDepartmentResourcesResponse) GetTotalSuccessfulOk() (*int32, bool)`

GetTotalSuccessfulOk returns a tuple with the TotalSuccessful field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalSuccessful

`func (o *BatchPatchDepartmentResourcesResponse) SetTotalSuccessful(v int32)`

SetTotalSuccessful sets TotalSuccessful field to given value.


### GetTotalFailed

`func (o *BatchPatchDepartmentResourcesResponse) GetTotalFailed() int32`

GetTotalFailed returns the TotalFailed field if non-nil, zero value otherwise.

### GetTotalFailedOk

`func (o *BatchPatchDepartmentResourcesResponse) GetTotalFailedOk() (*int32, bool)`

GetTotalFailedOk returns a tuple with the TotalFailed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalFailed

`func (o *BatchPatchDepartmentResourcesResponse) SetTotalFailed(v int32)`

SetTotalFailed sets TotalFailed field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


