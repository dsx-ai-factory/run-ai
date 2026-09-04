# GitPassword

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PasswordSecret** | Pointer to **NullableString** | Secret containing the credentials of the repository (needed for non public repository which requires authentication). (deprecated) | [optional] 
**SecretKeyOfUser** | Pointer to **NullableString** | The key to use for loading the user name from the secret. The default is &#x60;User&#x60;. (deprecated) | [optional] 
**SecretKeyOfPassword** | Pointer to **NullableString** | The key to use for loading the password from the secret. The default is &#x60;Password&#x60;. (deprecated) | [optional] 

## Methods

### NewGitPassword

`func NewGitPassword() *GitPassword`

NewGitPassword instantiates a new GitPassword object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitPasswordWithDefaults

`func NewGitPasswordWithDefaults() *GitPassword`

NewGitPasswordWithDefaults instantiates a new GitPassword object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPasswordSecret

`func (o *GitPassword) GetPasswordSecret() string`

GetPasswordSecret returns the PasswordSecret field if non-nil, zero value otherwise.

### GetPasswordSecretOk

`func (o *GitPassword) GetPasswordSecretOk() (*string, bool)`

GetPasswordSecretOk returns a tuple with the PasswordSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordSecret

`func (o *GitPassword) SetPasswordSecret(v string)`

SetPasswordSecret sets PasswordSecret field to given value.

### HasPasswordSecret

`func (o *GitPassword) HasPasswordSecret() bool`

HasPasswordSecret returns a boolean if a field has been set.

### SetPasswordSecretNil

`func (o *GitPassword) SetPasswordSecretNil(b bool)`

 SetPasswordSecretNil sets the value for PasswordSecret to be an explicit nil

### UnsetPasswordSecret
`func (o *GitPassword) UnsetPasswordSecret()`

UnsetPasswordSecret ensures that no value is present for PasswordSecret, not even an explicit nil
### GetSecretKeyOfUser

`func (o *GitPassword) GetSecretKeyOfUser() string`

GetSecretKeyOfUser returns the SecretKeyOfUser field if non-nil, zero value otherwise.

### GetSecretKeyOfUserOk

`func (o *GitPassword) GetSecretKeyOfUserOk() (*string, bool)`

GetSecretKeyOfUserOk returns a tuple with the SecretKeyOfUser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfUser

`func (o *GitPassword) SetSecretKeyOfUser(v string)`

SetSecretKeyOfUser sets SecretKeyOfUser field to given value.

### HasSecretKeyOfUser

`func (o *GitPassword) HasSecretKeyOfUser() bool`

HasSecretKeyOfUser returns a boolean if a field has been set.

### SetSecretKeyOfUserNil

`func (o *GitPassword) SetSecretKeyOfUserNil(b bool)`

 SetSecretKeyOfUserNil sets the value for SecretKeyOfUser to be an explicit nil

### UnsetSecretKeyOfUser
`func (o *GitPassword) UnsetSecretKeyOfUser()`

UnsetSecretKeyOfUser ensures that no value is present for SecretKeyOfUser, not even an explicit nil
### GetSecretKeyOfPassword

`func (o *GitPassword) GetSecretKeyOfPassword() string`

GetSecretKeyOfPassword returns the SecretKeyOfPassword field if non-nil, zero value otherwise.

### GetSecretKeyOfPasswordOk

`func (o *GitPassword) GetSecretKeyOfPasswordOk() (*string, bool)`

GetSecretKeyOfPasswordOk returns a tuple with the SecretKeyOfPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfPassword

`func (o *GitPassword) SetSecretKeyOfPassword(v string)`

SetSecretKeyOfPassword sets SecretKeyOfPassword field to given value.

### HasSecretKeyOfPassword

`func (o *GitPassword) HasSecretKeyOfPassword() bool`

HasSecretKeyOfPassword returns a boolean if a field has been set.

### SetSecretKeyOfPasswordNil

`func (o *GitPassword) SetSecretKeyOfPasswordNil(b bool)`

 SetSecretKeyOfPasswordNil sets the value for SecretKeyOfPassword to be an explicit nil

### UnsetSecretKeyOfPassword
`func (o *GitPassword) UnsetSecretKeyOfPassword()`

UnsetSecretKeyOfPassword ensures that no value is present for SecretKeyOfPassword, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


