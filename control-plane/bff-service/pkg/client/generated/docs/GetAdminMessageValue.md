# GetAdminMessageValue

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Unique identifier for the admin message | [optional] 
**Msg** | Pointer to **string** | The message content displayed to users. Supports a limited HTML subset for basic formatting | [optional] 
**CreatedAt** | Pointer to **int64** | Time the message was created, as a Unix timestamp | [optional] 
**EnableDismiss** | Pointer to **bool** | When true, users can dismiss the message from; when false, the message persists until it is updated or removed | [optional] 

## Methods

### NewGetAdminMessageValue

`func NewGetAdminMessageValue() *GetAdminMessageValue`

NewGetAdminMessageValue instantiates a new GetAdminMessageValue object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetAdminMessageValueWithDefaults

`func NewGetAdminMessageValueWithDefaults() *GetAdminMessageValue`

NewGetAdminMessageValueWithDefaults instantiates a new GetAdminMessageValue object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *GetAdminMessageValue) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetAdminMessageValue) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetAdminMessageValue) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *GetAdminMessageValue) HasId() bool`

HasId returns a boolean if a field has been set.

### GetMsg

`func (o *GetAdminMessageValue) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *GetAdminMessageValue) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *GetAdminMessageValue) SetMsg(v string)`

SetMsg sets Msg field to given value.

### HasMsg

`func (o *GetAdminMessageValue) HasMsg() bool`

HasMsg returns a boolean if a field has been set.

### GetCreatedAt

`func (o *GetAdminMessageValue) GetCreatedAt() int64`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GetAdminMessageValue) GetCreatedAtOk() (*int64, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GetAdminMessageValue) SetCreatedAt(v int64)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *GetAdminMessageValue) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetEnableDismiss

`func (o *GetAdminMessageValue) GetEnableDismiss() bool`

GetEnableDismiss returns the EnableDismiss field if non-nil, zero value otherwise.

### GetEnableDismissOk

`func (o *GetAdminMessageValue) GetEnableDismissOk() (*bool, bool)`

GetEnableDismissOk returns a tuple with the EnableDismiss field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnableDismiss

`func (o *GetAdminMessageValue) SetEnableDismiss(v bool)`

SetEnableDismiss sets EnableDismiss field to given value.

### HasEnableDismiss

`func (o *GetAdminMessageValue) HasEnableDismiss() bool`

HasEnableDismiss returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


