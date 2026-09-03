# BatchPatchNodePoolResourcesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PatchRequests** | [**[]NodePoolResourcesPatchRequest**](NodePoolResourcesPatchRequest.md) | Array of node pool resources patch requests. | 

## Methods

### NewBatchPatchNodePoolResourcesRequest

`func NewBatchPatchNodePoolResourcesRequest(patchRequests []NodePoolResourcesPatchRequest, ) *BatchPatchNodePoolResourcesRequest`

NewBatchPatchNodePoolResourcesRequest instantiates a new BatchPatchNodePoolResourcesRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBatchPatchNodePoolResourcesRequestWithDefaults

`func NewBatchPatchNodePoolResourcesRequestWithDefaults() *BatchPatchNodePoolResourcesRequest`

NewBatchPatchNodePoolResourcesRequestWithDefaults instantiates a new BatchPatchNodePoolResourcesRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPatchRequests

`func (o *BatchPatchNodePoolResourcesRequest) GetPatchRequests() []NodePoolResourcesPatchRequest`

GetPatchRequests returns the PatchRequests field if non-nil, zero value otherwise.

### GetPatchRequestsOk

`func (o *BatchPatchNodePoolResourcesRequest) GetPatchRequestsOk() (*[]NodePoolResourcesPatchRequest, bool)`

GetPatchRequestsOk returns a tuple with the PatchRequests field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPatchRequests

`func (o *BatchPatchNodePoolResourcesRequest) SetPatchRequests(v []NodePoolResourcesPatchRequest)`

SetPatchRequests sets PatchRequests field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


