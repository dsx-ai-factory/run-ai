# UserCreationRequest1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** |  | 
**ResetPassword** | Pointer to **NullableBool** |  | [optional] 

## Methods

### NewUserCreationRequest1

`func NewUserCreationRequest1(email string, ) *UserCreationRequest1`

NewUserCreationRequest1 instantiates a new UserCreationRequest1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserCreationRequest1WithDefaults

`func NewUserCreationRequest1WithDefaults() *UserCreationRequest1`

NewUserCreationRequest1WithDefaults instantiates a new UserCreationRequest1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *UserCreationRequest1) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserCreationRequest1) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserCreationRequest1) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetResetPassword

`func (o *UserCreationRequest1) GetResetPassword() bool`

GetResetPassword returns the ResetPassword field if non-nil, zero value otherwise.

### GetResetPasswordOk

`func (o *UserCreationRequest1) GetResetPasswordOk() (*bool, bool)`

GetResetPasswordOk returns a tuple with the ResetPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResetPassword

`func (o *UserCreationRequest1) SetResetPassword(v bool)`

SetResetPassword sets ResetPassword field to given value.

### HasResetPassword

`func (o *UserCreationRequest1) HasResetPassword() bool`

HasResetPassword returns a boolean if a field has been set.

### SetResetPasswordNil

`func (o *UserCreationRequest1) SetResetPasswordNil(b bool)`

 SetResetPasswordNil sets the value for ResetPassword to be an explicit nil

### UnsetResetPassword
`func (o *UserCreationRequest1) UnsetResetPassword()`

UnsetResetPassword ensures that no value is present for ResetPassword, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


