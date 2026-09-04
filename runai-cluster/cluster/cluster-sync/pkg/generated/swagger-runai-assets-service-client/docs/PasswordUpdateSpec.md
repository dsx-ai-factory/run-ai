# PasswordUpdateSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExistingSecretName** | Pointer to **NullableString** | optional name of existing secret in the cluster from which the user name and password should be taken. If omitted, you will have to provide those credentials in the user and password fields. The provided credentials are encrypted into the control plane database and cloned to a kubernetes secret in the cluster. | [optional] 
**User** | Pointer to **NullableString** | The name of the user, required only when not using existing secret. | [optional] 
**Password** | Pointer to **NullableString** | The password, required only when not using existing secret. | [optional] 

## Methods

### NewPasswordUpdateSpec

`func NewPasswordUpdateSpec() *PasswordUpdateSpec`

NewPasswordUpdateSpec instantiates a new PasswordUpdateSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPasswordUpdateSpecWithDefaults

`func NewPasswordUpdateSpecWithDefaults() *PasswordUpdateSpec`

NewPasswordUpdateSpecWithDefaults instantiates a new PasswordUpdateSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExistingSecretName

`func (o *PasswordUpdateSpec) GetExistingSecretName() string`

GetExistingSecretName returns the ExistingSecretName field if non-nil, zero value otherwise.

### GetExistingSecretNameOk

`func (o *PasswordUpdateSpec) GetExistingSecretNameOk() (*string, bool)`

GetExistingSecretNameOk returns a tuple with the ExistingSecretName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingSecretName

`func (o *PasswordUpdateSpec) SetExistingSecretName(v string)`

SetExistingSecretName sets ExistingSecretName field to given value.

### HasExistingSecretName

`func (o *PasswordUpdateSpec) HasExistingSecretName() bool`

HasExistingSecretName returns a boolean if a field has been set.

### SetExistingSecretNameNil

`func (o *PasswordUpdateSpec) SetExistingSecretNameNil(b bool)`

 SetExistingSecretNameNil sets the value for ExistingSecretName to be an explicit nil

### UnsetExistingSecretName
`func (o *PasswordUpdateSpec) UnsetExistingSecretName()`

UnsetExistingSecretName ensures that no value is present for ExistingSecretName, not even an explicit nil
### GetUser

`func (o *PasswordUpdateSpec) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *PasswordUpdateSpec) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *PasswordUpdateSpec) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *PasswordUpdateSpec) HasUser() bool`

HasUser returns a boolean if a field has been set.

### SetUserNil

`func (o *PasswordUpdateSpec) SetUserNil(b bool)`

 SetUserNil sets the value for User to be an explicit nil

### UnsetUser
`func (o *PasswordUpdateSpec) UnsetUser()`

UnsetUser ensures that no value is present for User, not even an explicit nil
### GetPassword

`func (o *PasswordUpdateSpec) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *PasswordUpdateSpec) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *PasswordUpdateSpec) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *PasswordUpdateSpec) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### SetPasswordNil

`func (o *PasswordUpdateSpec) SetPasswordNil(b bool)`

 SetPasswordNil sets the value for Password to be an explicit nil

### UnsetPassword
`func (o *PasswordUpdateSpec) UnsetPassword()`

UnsetPassword ensures that no value is present for Password, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


