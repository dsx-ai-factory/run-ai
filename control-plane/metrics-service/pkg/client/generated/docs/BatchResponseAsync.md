# BatchResponseAsync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The ID of the batch operation. | 
**Accepted** | Pointer to **[]string** | List of ids to perform the action on. | [optional] 
**Failed** | Pointer to [**[]BatchResponseFailureItem**](BatchResponseFailureItem.md) | The IDs that failed to update. | [optional] 

## Methods

### NewBatchResponseAsync

`func NewBatchResponseAsync(id string, ) *BatchResponseAsync`

NewBatchResponseAsync instantiates a new BatchResponseAsync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBatchResponseAsyncWithDefaults

`func NewBatchResponseAsyncWithDefaults() *BatchResponseAsync`

NewBatchResponseAsyncWithDefaults instantiates a new BatchResponseAsync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *BatchResponseAsync) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *BatchResponseAsync) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *BatchResponseAsync) SetId(v string)`

SetId sets Id field to given value.


### GetAccepted

`func (o *BatchResponseAsync) GetAccepted() []string`

GetAccepted returns the Accepted field if non-nil, zero value otherwise.

### GetAcceptedOk

`func (o *BatchResponseAsync) GetAcceptedOk() (*[]string, bool)`

GetAcceptedOk returns a tuple with the Accepted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccepted

`func (o *BatchResponseAsync) SetAccepted(v []string)`

SetAccepted sets Accepted field to given value.

### HasAccepted

`func (o *BatchResponseAsync) HasAccepted() bool`

HasAccepted returns a boolean if a field has been set.

### GetFailed

`func (o *BatchResponseAsync) GetFailed() []BatchResponseFailureItem`

GetFailed returns the Failed field if non-nil, zero value otherwise.

### GetFailedOk

`func (o *BatchResponseAsync) GetFailedOk() (*[]BatchResponseFailureItem, bool)`

GetFailedOk returns a tuple with the Failed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailed

`func (o *BatchResponseAsync) SetFailed(v []BatchResponseFailureItem)`

SetFailed sets Failed field to given value.

### HasFailed

`func (o *BatchResponseAsync) HasFailed() bool`

HasFailed returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


