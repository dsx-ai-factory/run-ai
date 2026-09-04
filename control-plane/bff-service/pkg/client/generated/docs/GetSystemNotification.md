# GetSystemNotification

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Unique identifier for the admin message | [optional] 
**Msg** | Pointer to **string** | The message content displayed to users. Supports a limited HTML subset for basic formatting | [optional] 
**CreatedAt** | Pointer to **int64** | Time the message was created, as a Unix timestamp | [optional] 
**DisplayName** | Pointer to **string** | Human-readable name for this setting | [optional] [readonly] 
**Description** | Pointer to **string** | Short explanation of what this setting controls, shown alongside the setting | [optional] [readonly] 

## Methods

### NewGetSystemNotification

`func NewGetSystemNotification() *GetSystemNotification`

NewGetSystemNotification instantiates a new GetSystemNotification object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetSystemNotificationWithDefaults

`func NewGetSystemNotificationWithDefaults() *GetSystemNotification`

NewGetSystemNotificationWithDefaults instantiates a new GetSystemNotification object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *GetSystemNotification) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetSystemNotification) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetSystemNotification) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *GetSystemNotification) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMsg

`func (o *GetSystemNotification) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *GetSystemNotification) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *GetSystemNotification) SetMsg(v string)`

SetMsg sets Msg field to given value.

### HasMsg

`func (o *GetSystemNotification) HasMsg() bool`

HasMsg returns a boolean if a field has been set.

### GetCreatedAt

`func (o *GetSystemNotification) GetCreatedAt() int64`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GetSystemNotification) GetCreatedAtOk() (*int64, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GetSystemNotification) SetCreatedAt(v int64)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *GetSystemNotification) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetDisplayName

`func (o *GetSystemNotification) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *GetSystemNotification) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *GetSystemNotification) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *GetSystemNotification) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetDescription

`func (o *GetSystemNotification) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GetSystemNotification) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GetSystemNotification) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GetSystemNotification) HasDescription() bool`

HasDescription returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


