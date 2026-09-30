# BatchPatchProjectResourcesResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique ID of the batch operation. | 
**Succeeded** | Pointer to [**[]ProjectNodePoolResources**](ProjectNodePoolResources.md) | Successfully patched project node pool resources results. | [optional] 
**Failed** | Pointer to [**[]FailedNodePoolResourcesPatchRequest**](FailedNodePoolResourcesPatchRequest.md) | Failed patch operations with the original payload and error details. | [optional] 
**TotalRequested** | **int32** | Total number of items requested in the batch. | 
**TotalSuccessful** | **int32** | Number of items successfully processed. | 
**TotalFailed** | **int32** | Number of items that failed. | 

## Methods

### NewBatchPatchProjectResourcesResponse

`func NewBatchPatchProjectResourcesResponse(id string, totalRequested int32, totalSuccessful int32, totalFailed int32, ) *BatchPatchProjectResourcesResponse`

NewBatchPatchProjectResourcesResponse instantiates a new BatchPatchProjectResourcesResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBatchPatchProjectResourcesResponseWithDefaults

`func NewBatchPatchProjectResourcesResponseWithDefaults() *BatchPatchProjectResourcesResponse`

NewBatchPatchProjectResourcesResponseWithDefaults instantiates a new BatchPatchProjectResourcesResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *BatchPatchProjectResourcesResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *BatchPatchProjectResourcesResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *BatchPatchProjectResourcesResponse) SetId(v string)`

SetId sets Id field to given value.


### GetSucceeded

`func (o *BatchPatchProjectResourcesResponse) GetSucceeded() []ProjectNodePoolResources`

GetSucceeded returns the Succeeded field if non-nil, zero value otherwise.

### GetSucceededOk

`func (o *BatchPatchProjectResourcesResponse) GetSucceededOk() (*[]ProjectNodePoolResources, bool)`

GetSucceededOk returns a tuple with the Succeeded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSucceeded

`func (o *BatchPatchProjectResourcesResponse) SetSucceeded(v []ProjectNodePoolResources)`

SetSucceeded sets Succeeded field to given value.

### HasSucceeded

`func (o *BatchPatchProjectResourcesResponse) HasSucceeded() bool`

HasSucceeded returns a boolean if a field has been set.

### GetFailed

`func (o *BatchPatchProjectResourcesResponse) GetFailed() []FailedNodePoolResourcesPatchRequest`

GetFailed returns the Failed field if non-nil, zero value otherwise.

### GetFailedOk

`func (o *BatchPatchProjectResourcesResponse) GetFailedOk() (*[]FailedNodePoolResourcesPatchRequest, bool)`

GetFailedOk returns a tuple with the Failed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailed

`func (o *BatchPatchProjectResourcesResponse) SetFailed(v []FailedNodePoolResourcesPatchRequest)`

SetFailed sets Failed field to given value.

### HasFailed

`func (o *BatchPatchProjectResourcesResponse) HasFailed() bool`

HasFailed returns a boolean if a field has been set.

### GetTotalRequested

`func (o *BatchPatchProjectResourcesResponse) GetTotalRequested() int32`

GetTotalRequested returns the TotalRequested field if non-nil, zero value otherwise.

### GetTotalRequestedOk

`func (o *BatchPatchProjectResourcesResponse) GetTotalRequestedOk() (*int32, bool)`

GetTotalRequestedOk returns a tuple with the TotalRequested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRequested

`func (o *BatchPatchProjectResourcesResponse) SetTotalRequested(v int32)`

SetTotalRequested sets TotalRequested field to given value.


### GetTotalSuccessful

`func (o *BatchPatchProjectResourcesResponse) GetTotalSuccessful() int32`

GetTotalSuccessful returns the TotalSuccessful field if non-nil, zero value otherwise.

### GetTotalSuccessfulOk

`func (o *BatchPatchProjectResourcesResponse) GetTotalSuccessfulOk() (*int32, bool)`

GetTotalSuccessfulOk returns a tuple with the TotalSuccessful field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalSuccessful

`func (o *BatchPatchProjectResourcesResponse) SetTotalSuccessful(v int32)`

SetTotalSuccessful sets TotalSuccessful field to given value.


### GetTotalFailed

`func (o *BatchPatchProjectResourcesResponse) GetTotalFailed() int32`

GetTotalFailed returns the TotalFailed field if non-nil, zero value otherwise.

### GetTotalFailedOk

`func (o *BatchPatchProjectResourcesResponse) GetTotalFailedOk() (*int32, bool)`

GetTotalFailedOk returns a tuple with the TotalFailed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalFailed

`func (o *BatchPatchProjectResourcesResponse) SetTotalFailed(v int32)`

SetTotalFailed sets TotalFailed field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


