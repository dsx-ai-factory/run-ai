# UserCreationRequestBulk

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** | The email of the user. Must be a valid email address. | 
**ResetPassword** | Pointer to **NullableBool** |  | [optional] 

## Methods

### NewUserCreationRequestBulk

`func NewUserCreationRequestBulk(email string, ) *UserCreationRequestBulk`

NewUserCreationRequestBulk instantiates a new UserCreationRequestBulk object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserCreationRequestBulkWithDefaults

`func NewUserCreationRequestBulkWithDefaults() *UserCreationRequestBulk`

NewUserCreationRequestBulkWithDefaults instantiates a new UserCreationRequestBulk object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *UserCreationRequestBulk) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserCreationRequestBulk) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserCreationRequestBulk) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetResetPassword

`func (o *UserCreationRequestBulk) GetResetPassword() bool`

GetResetPassword returns the ResetPassword field if non-nil, zero value otherwise.

### GetResetPasswordOk

`func (o *UserCreationRequestBulk) GetResetPasswordOk() (*bool, bool)`

GetResetPasswordOk returns a tuple with the ResetPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResetPassword

`func (o *UserCreationRequestBulk) SetResetPassword(v bool)`

SetResetPassword sets ResetPassword field to given value.

### HasResetPassword

`func (o *UserCreationRequestBulk) HasResetPassword() bool`

HasResetPassword returns a boolean if a field has been set.

### SetResetPasswordNil

`func (o *UserCreationRequestBulk) SetResetPasswordNil(b bool)`

 SetResetPasswordNil sets the value for ResetPassword to be an explicit nil

### UnsetResetPassword
`func (o *UserCreationRequestBulk) UnsetResetPassword()`

UnsetResetPassword ensures that no value is present for ResetPassword, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


