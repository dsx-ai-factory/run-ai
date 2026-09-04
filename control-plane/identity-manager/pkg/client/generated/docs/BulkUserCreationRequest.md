# BulkUserCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Users** | [**[]UserCreationRequestBulk**](UserCreationRequestBulk.md) | Array of users to create. Maximum 10 users per batch. | 
**Notify** | Pointer to **NullableBool** | When true, send an email to the user. The email includes a link to reset the password. | [optional] 

## Methods

### NewBulkUserCreationRequest

`func NewBulkUserCreationRequest(users []UserCreationRequestBulk, ) *BulkUserCreationRequest`

NewBulkUserCreationRequest instantiates a new BulkUserCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkUserCreationRequestWithDefaults

`func NewBulkUserCreationRequestWithDefaults() *BulkUserCreationRequest`

NewBulkUserCreationRequestWithDefaults instantiates a new BulkUserCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsers

`func (o *BulkUserCreationRequest) GetUsers() []UserCreationRequestBulk`

GetUsers returns the Users field if non-nil, zero value otherwise.

### GetUsersOk

`func (o *BulkUserCreationRequest) GetUsersOk() (*[]UserCreationRequestBulk, bool)`

GetUsersOk returns a tuple with the Users field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsers

`func (o *BulkUserCreationRequest) SetUsers(v []UserCreationRequestBulk)`

SetUsers sets Users field to given value.


### GetNotify

`func (o *BulkUserCreationRequest) GetNotify() bool`

GetNotify returns the Notify field if non-nil, zero value otherwise.

### GetNotifyOk

`func (o *BulkUserCreationRequest) GetNotifyOk() (*bool, bool)`

GetNotifyOk returns a tuple with the Notify field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotify

`func (o *BulkUserCreationRequest) SetNotify(v bool)`

SetNotify sets Notify field to given value.

### HasNotify

`func (o *BulkUserCreationRequest) HasNotify() bool`

HasNotify returns a boolean if a field has been set.

### SetNotifyNil

`func (o *BulkUserCreationRequest) SetNotifyNil(b bool)`

 SetNotifyNil sets the value for Notify to be an explicit nil

### UnsetNotify
`func (o *BulkUserCreationRequest) UnsetNotify()`

UnsetNotify ensures that no value is present for Notify, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


