# AccessRulesBatchCreateResponseFailedInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Payload** | Pointer to [**AccessRuleCreationFields**](AccessRuleCreationFields.md) |  | [optional] 
**Code** | Pointer to **int32** | A numeric code representing the error type. | [optional] 
**Message** | Pointer to **string** | A human-readable message describing the access rule IDs that failed to create. | [optional] 

## Methods

### NewAccessRulesBatchCreateResponseFailedInner

`func NewAccessRulesBatchCreateResponseFailedInner() *AccessRulesBatchCreateResponseFailedInner`

NewAccessRulesBatchCreateResponseFailedInner instantiates a new AccessRulesBatchCreateResponseFailedInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRulesBatchCreateResponseFailedInnerWithDefaults

`func NewAccessRulesBatchCreateResponseFailedInnerWithDefaults() *AccessRulesBatchCreateResponseFailedInner`

NewAccessRulesBatchCreateResponseFailedInnerWithDefaults instantiates a new AccessRulesBatchCreateResponseFailedInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPayload

`func (o *AccessRulesBatchCreateResponseFailedInner) GetPayload() AccessRuleCreationFields`

GetPayload returns the Payload field if non-nil, zero value otherwise.

### GetPayloadOk

`func (o *AccessRulesBatchCreateResponseFailedInner) GetPayloadOk() (*AccessRuleCreationFields, bool)`

GetPayloadOk returns a tuple with the Payload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPayload

`func (o *AccessRulesBatchCreateResponseFailedInner) SetPayload(v AccessRuleCreationFields)`

SetPayload sets Payload field to given value.

### HasPayload

`func (o *AccessRulesBatchCreateResponseFailedInner) HasPayload() bool`

HasPayload returns a boolean if a field has been set.

### GetCode

`func (o *AccessRulesBatchCreateResponseFailedInner) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *AccessRulesBatchCreateResponseFailedInner) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *AccessRulesBatchCreateResponseFailedInner) SetCode(v int32)`

SetCode sets Code field to given value.

### HasCode

`func (o *AccessRulesBatchCreateResponseFailedInner) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetMessage

`func (o *AccessRulesBatchCreateResponseFailedInner) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *AccessRulesBatchCreateResponseFailedInner) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *AccessRulesBatchCreateResponseFailedInner) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *AccessRulesBatchCreateResponseFailedInner) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


