# FailedNodePoolResourcesPatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Payload** | Pointer to [**NodePoolResourcesPatchRequest**](NodePoolResourcesPatchRequest.md) |  | [optional] 
**Code** | Pointer to **int32** | HTTP error code. | [optional] 
**Message** | Pointer to **string** | Error message. | [optional] 

## Methods

### NewFailedNodePoolResourcesPatchRequest

`func NewFailedNodePoolResourcesPatchRequest() *FailedNodePoolResourcesPatchRequest`

NewFailedNodePoolResourcesPatchRequest instantiates a new FailedNodePoolResourcesPatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFailedNodePoolResourcesPatchRequestWithDefaults

`func NewFailedNodePoolResourcesPatchRequestWithDefaults() *FailedNodePoolResourcesPatchRequest`

NewFailedNodePoolResourcesPatchRequestWithDefaults instantiates a new FailedNodePoolResourcesPatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPayload

`func (o *FailedNodePoolResourcesPatchRequest) GetPayload() NodePoolResourcesPatchRequest`

GetPayload returns the Payload field if non-nil, zero value otherwise.

### GetPayloadOk

`func (o *FailedNodePoolResourcesPatchRequest) GetPayloadOk() (*NodePoolResourcesPatchRequest, bool)`

GetPayloadOk returns a tuple with the Payload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPayload

`func (o *FailedNodePoolResourcesPatchRequest) SetPayload(v NodePoolResourcesPatchRequest)`

SetPayload sets Payload field to given value.

### HasPayload

`func (o *FailedNodePoolResourcesPatchRequest) HasPayload() bool`

HasPayload returns a boolean if a field has been set.

### GetCode

`func (o *FailedNodePoolResourcesPatchRequest) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *FailedNodePoolResourcesPatchRequest) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *FailedNodePoolResourcesPatchRequest) SetCode(v int32)`

SetCode sets Code field to given value.

### HasCode

`func (o *FailedNodePoolResourcesPatchRequest) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetMessage

`func (o *FailedNodePoolResourcesPatchRequest) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *FailedNodePoolResourcesPatchRequest) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *FailedNodePoolResourcesPatchRequest) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *FailedNodePoolResourcesPatchRequest) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


