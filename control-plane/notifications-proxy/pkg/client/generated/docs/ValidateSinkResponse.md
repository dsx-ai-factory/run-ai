# ValidateSinkResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | **bool** |  | 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewValidateSinkResponse

`func NewValidateSinkResponse(status bool, ) *ValidateSinkResponse`

NewValidateSinkResponse instantiates a new ValidateSinkResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewValidateSinkResponseWithDefaults

`func NewValidateSinkResponseWithDefaults() *ValidateSinkResponse`

NewValidateSinkResponseWithDefaults instantiates a new ValidateSinkResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *ValidateSinkResponse) GetStatus() bool`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ValidateSinkResponse) GetStatusOk() (*bool, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ValidateSinkResponse) SetStatus(v bool)`

SetStatus sets Status field to given value.


### GetMessage

`func (o *ValidateSinkResponse) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ValidateSinkResponse) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ValidateSinkResponse) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ValidateSinkResponse) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


