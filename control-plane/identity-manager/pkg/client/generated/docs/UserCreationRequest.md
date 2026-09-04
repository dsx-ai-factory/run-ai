# UserCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** | The email of the user. Must be a valid email address. | 
**ResetPassword** | Pointer to **NullableBool** |  | [optional] 
**Notify** | Pointer to **NullableBool** | When true, send an email to the user. The email includes a link to reset the password. | [optional] 
**FirstName** | Pointer to **string** |  | [optional] 
**LastName** | Pointer to **string** |  | [optional] 

## Methods

### NewUserCreationRequest

`func NewUserCreationRequest(email string, ) *UserCreationRequest`

NewUserCreationRequest instantiates a new UserCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserCreationRequestWithDefaults

`func NewUserCreationRequestWithDefaults() *UserCreationRequest`

NewUserCreationRequestWithDefaults instantiates a new UserCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *UserCreationRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserCreationRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserCreationRequest) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetResetPassword

`func (o *UserCreationRequest) GetResetPassword() bool`

GetResetPassword returns the ResetPassword field if non-nil, zero value otherwise.

### GetResetPasswordOk

`func (o *UserCreationRequest) GetResetPasswordOk() (*bool, bool)`

GetResetPasswordOk returns a tuple with the ResetPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResetPassword

`func (o *UserCreationRequest) SetResetPassword(v bool)`

SetResetPassword sets ResetPassword field to given value.

### HasResetPassword

`func (o *UserCreationRequest) HasResetPassword() bool`

HasResetPassword returns a boolean if a field has been set.

### SetResetPasswordNil

`func (o *UserCreationRequest) SetResetPasswordNil(b bool)`

 SetResetPasswordNil sets the value for ResetPassword to be an explicit nil

### UnsetResetPassword
`func (o *UserCreationRequest) UnsetResetPassword()`

UnsetResetPassword ensures that no value is present for ResetPassword, not even an explicit nil
### GetNotify

`func (o *UserCreationRequest) GetNotify() bool`

GetNotify returns the Notify field if non-nil, zero value otherwise.

### GetNotifyOk

`func (o *UserCreationRequest) GetNotifyOk() (*bool, bool)`

GetNotifyOk returns a tuple with the Notify field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotify

`func (o *UserCreationRequest) SetNotify(v bool)`

SetNotify sets Notify field to given value.

### HasNotify

`func (o *UserCreationRequest) HasNotify() bool`

HasNotify returns a boolean if a field has been set.

### SetNotifyNil

`func (o *UserCreationRequest) SetNotifyNil(b bool)`

 SetNotifyNil sets the value for Notify to be an explicit nil

### UnsetNotify
`func (o *UserCreationRequest) UnsetNotify()`

UnsetNotify ensures that no value is present for Notify, not even an explicit nil
### GetFirstName

`func (o *UserCreationRequest) GetFirstName() string`

GetFirstName returns the FirstName field if non-nil, zero value otherwise.

### GetFirstNameOk

`func (o *UserCreationRequest) GetFirstNameOk() (*string, bool)`

GetFirstNameOk returns a tuple with the FirstName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFirstName

`func (o *UserCreationRequest) SetFirstName(v string)`

SetFirstName sets FirstName field to given value.

### HasFirstName

`func (o *UserCreationRequest) HasFirstName() bool`

HasFirstName returns a boolean if a field has been set.

### GetLastName

`func (o *UserCreationRequest) GetLastName() string`

GetLastName returns the LastName field if non-nil, zero value otherwise.

### GetLastNameOk

`func (o *UserCreationRequest) GetLastNameOk() (*string, bool)`

GetLastNameOk returns a tuple with the LastName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastName

`func (o *UserCreationRequest) SetLastName(v string)`

SetLastName sets LastName field to given value.

### HasLastName

`func (o *UserCreationRequest) HasLastName() bool`

HasLastName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


