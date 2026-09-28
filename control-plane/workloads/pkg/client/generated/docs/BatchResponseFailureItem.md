# BatchResponseFailureItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Code** | **int32** | A numeric code representing the error type. | 
**Message** | **string** | The error message. | 

## Methods

### NewBatchResponseFailureItem

`func NewBatchResponseFailureItem(id string, code int32, message string, ) *BatchResponseFailureItem`

NewBatchResponseFailureItem instantiates a new BatchResponseFailureItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBatchResponseFailureItemWithDefaults

`func NewBatchResponseFailureItemWithDefaults() *BatchResponseFailureItem`

NewBatchResponseFailureItemWithDefaults instantiates a new BatchResponseFailureItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *BatchResponseFailureItem) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *BatchResponseFailureItem) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *BatchResponseFailureItem) SetId(v string)`

SetId sets Id field to given value.


### GetCode

`func (o *BatchResponseFailureItem) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *BatchResponseFailureItem) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *BatchResponseFailureItem) SetCode(v int32)`

SetCode sets Code field to given value.


### GetMessage

`func (o *BatchResponseFailureItem) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *BatchResponseFailureItem) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *BatchResponseFailureItem) SetMessage(v string)`

SetMessage sets Message field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


