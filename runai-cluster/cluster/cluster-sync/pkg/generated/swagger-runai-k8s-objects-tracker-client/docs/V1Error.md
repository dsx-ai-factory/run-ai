# V1Error

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | **int32** |  | 
**Message** | **string** |  | 
**Details** | Pointer to **string** |  | [optional] 

## Methods

### NewV1Error

`func NewV1Error(code int32, message string, ) *V1Error`

NewV1Error instantiates a new V1Error object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewV1ErrorWithDefaults

`func NewV1ErrorWithDefaults() *V1Error`

NewV1ErrorWithDefaults instantiates a new V1Error object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *V1Error) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *V1Error) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *V1Error) SetCode(v int32)`

SetCode sets Code field to given value.


### GetMessage

`func (o *V1Error) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *V1Error) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *V1Error) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetDetails

`func (o *V1Error) GetDetails() string`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *V1Error) GetDetailsOk() (*string, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *V1Error) SetDetails(v string)`

SetDetails sets Details field to given value.

### HasDetails

`func (o *V1Error) HasDetails() bool`

HasDetails returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


