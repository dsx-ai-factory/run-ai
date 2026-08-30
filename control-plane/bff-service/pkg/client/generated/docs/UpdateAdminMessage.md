# UpdateAdminMessage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Msg** | Pointer to **NullableString** | The message content displayed to users. Supports a limited HTML subset for basic formatting | [optional] 
**EnableDismiss** | Pointer to **NullableBool** | When true, users can dismiss the message from; when false, the message persists until it is updated or removed | [optional] 

## Methods

### NewUpdateAdminMessage

`func NewUpdateAdminMessage() *UpdateAdminMessage`

NewUpdateAdminMessage instantiates a new UpdateAdminMessage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateAdminMessageWithDefaults

`func NewUpdateAdminMessageWithDefaults() *UpdateAdminMessage`

NewUpdateAdminMessageWithDefaults instantiates a new UpdateAdminMessage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMsg

`func (o *UpdateAdminMessage) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *UpdateAdminMessage) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *UpdateAdminMessage) SetMsg(v string)`

SetMsg sets Msg field to given value.

### HasMsg

`func (o *UpdateAdminMessage) HasMsg() bool`

HasMsg returns a boolean if a field has been set.

### SetMsgNil

`func (o *UpdateAdminMessage) SetMsgNil(b bool)`

 SetMsgNil sets the value for Msg to be an explicit nil

### UnsetMsg
`func (o *UpdateAdminMessage) UnsetMsg()`

UnsetMsg ensures that no value is present for Msg, not even an explicit nil
### GetEnableDismiss

`func (o *UpdateAdminMessage) GetEnableDismiss() bool`

GetEnableDismiss returns the EnableDismiss field if non-nil, zero value otherwise.

### GetEnableDismissOk

`func (o *UpdateAdminMessage) GetEnableDismissOk() (*bool, bool)`

GetEnableDismissOk returns a tuple with the EnableDismiss field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnableDismiss

`func (o *UpdateAdminMessage) SetEnableDismiss(v bool)`

SetEnableDismiss sets EnableDismiss field to given value.

### HasEnableDismiss

`func (o *UpdateAdminMessage) HasEnableDismiss() bool`

HasEnableDismiss returns a boolean if a field has been set.

### SetEnableDismissNil

`func (o *UpdateAdminMessage) SetEnableDismissNil(b bool)`

 SetEnableDismissNil sets the value for EnableDismiss to be an explicit nil

### UnsetEnableDismiss
`func (o *UpdateAdminMessage) UnsetEnableDismiss()`

UnsetEnableDismiss ensures that no value is present for EnableDismiss, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


