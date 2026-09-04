# InternalUserCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** |  | 
**Realm** | **string** |  | 
**Password** | Pointer to **NullableString** |  | [optional] 
**ResetPassword** | **bool** |  | 
**Notify** | Pointer to **NullableBool** | When true, send an email to the user. The email includes a link to reset the password. | [optional] 

## Methods

### NewInternalUserCreationRequest

`func NewInternalUserCreationRequest(email string, realm string, resetPassword bool, ) *InternalUserCreationRequest`

NewInternalUserCreationRequest instantiates a new InternalUserCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalUserCreationRequestWithDefaults

`func NewInternalUserCreationRequestWithDefaults() *InternalUserCreationRequest`

NewInternalUserCreationRequestWithDefaults instantiates a new InternalUserCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *InternalUserCreationRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *InternalUserCreationRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *InternalUserCreationRequest) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetRealm

`func (o *InternalUserCreationRequest) GetRealm() string`

GetRealm returns the Realm field if non-nil, zero value otherwise.

### GetRealmOk

`func (o *InternalUserCreationRequest) GetRealmOk() (*string, bool)`

GetRealmOk returns a tuple with the Realm field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRealm

`func (o *InternalUserCreationRequest) SetRealm(v string)`

SetRealm sets Realm field to given value.


### GetPassword

`func (o *InternalUserCreationRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *InternalUserCreationRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *InternalUserCreationRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *InternalUserCreationRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### SetPasswordNil

`func (o *InternalUserCreationRequest) SetPasswordNil(b bool)`

 SetPasswordNil sets the value for Password to be an explicit nil

### UnsetPassword
`func (o *InternalUserCreationRequest) UnsetPassword()`

UnsetPassword ensures that no value is present for Password, not even an explicit nil
### GetResetPassword

`func (o *InternalUserCreationRequest) GetResetPassword() bool`

GetResetPassword returns the ResetPassword field if non-nil, zero value otherwise.

### GetResetPasswordOk

`func (o *InternalUserCreationRequest) GetResetPasswordOk() (*bool, bool)`

GetResetPasswordOk returns a tuple with the ResetPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResetPassword

`func (o *InternalUserCreationRequest) SetResetPassword(v bool)`

SetResetPassword sets ResetPassword field to given value.


### GetNotify

`func (o *InternalUserCreationRequest) GetNotify() bool`

GetNotify returns the Notify field if non-nil, zero value otherwise.

### GetNotifyOk

`func (o *InternalUserCreationRequest) GetNotifyOk() (*bool, bool)`

GetNotifyOk returns a tuple with the Notify field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotify

`func (o *InternalUserCreationRequest) SetNotify(v bool)`

SetNotify sets Notify field to given value.

### HasNotify

`func (o *InternalUserCreationRequest) HasNotify() bool`

HasNotify returns a boolean if a field has been set.

### SetNotifyNil

`func (o *InternalUserCreationRequest) SetNotifyNil(b bool)`

 SetNotifyNil sets the value for Notify to be an explicit nil

### UnsetNotify
`func (o *InternalUserCreationRequest) UnsetNotify()`

UnsetNotify ensures that no value is present for Notify, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


