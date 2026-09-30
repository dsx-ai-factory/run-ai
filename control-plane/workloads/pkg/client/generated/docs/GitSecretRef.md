# GitSecretRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Name of the Secret containing the credentials of the repository. | [optional] 
**AuthenticationMethod** | Pointer to [**GitAuthenticationMethod**](GitAuthenticationMethod.md) |  | [optional] 
**SecretKeyOfUser** | Pointer to **NullableString** | The key in the Secret that contains the Git username (used for &#x60;password&#x60; authentication). | [optional] 
**SecretKeyOfPassword** | Pointer to **NullableString** | The key in the Secret that contains the Git password (used for &#x60;password&#x60; authentication). | [optional] 

## Methods

### NewGitSecretRef

`func NewGitSecretRef() *GitSecretRef`

NewGitSecretRef instantiates a new GitSecretRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitSecretRefWithDefaults

`func NewGitSecretRefWithDefaults() *GitSecretRef`

NewGitSecretRefWithDefaults instantiates a new GitSecretRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *GitSecretRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GitSecretRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GitSecretRef) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GitSecretRef) HasName() bool`

HasName returns a boolean if a field has been set.

### GetAuthenticationMethod

`func (o *GitSecretRef) GetAuthenticationMethod() GitAuthenticationMethod`

GetAuthenticationMethod returns the AuthenticationMethod field if non-nil, zero value otherwise.

### GetAuthenticationMethodOk

`func (o *GitSecretRef) GetAuthenticationMethodOk() (*GitAuthenticationMethod, bool)`

GetAuthenticationMethodOk returns a tuple with the AuthenticationMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthenticationMethod

`func (o *GitSecretRef) SetAuthenticationMethod(v GitAuthenticationMethod)`

SetAuthenticationMethod sets AuthenticationMethod field to given value.

### HasAuthenticationMethod

`func (o *GitSecretRef) HasAuthenticationMethod() bool`

HasAuthenticationMethod returns a boolean if a field has been set.

### GetSecretKeyOfUser

`func (o *GitSecretRef) GetSecretKeyOfUser() string`

GetSecretKeyOfUser returns the SecretKeyOfUser field if non-nil, zero value otherwise.

### GetSecretKeyOfUserOk

`func (o *GitSecretRef) GetSecretKeyOfUserOk() (*string, bool)`

GetSecretKeyOfUserOk returns a tuple with the SecretKeyOfUser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfUser

`func (o *GitSecretRef) SetSecretKeyOfUser(v string)`

SetSecretKeyOfUser sets SecretKeyOfUser field to given value.

### HasSecretKeyOfUser

`func (o *GitSecretRef) HasSecretKeyOfUser() bool`

HasSecretKeyOfUser returns a boolean if a field has been set.

### SetSecretKeyOfUserNil

`func (o *GitSecretRef) SetSecretKeyOfUserNil(b bool)`

 SetSecretKeyOfUserNil sets the value for SecretKeyOfUser to be an explicit nil

### UnsetSecretKeyOfUser
`func (o *GitSecretRef) UnsetSecretKeyOfUser()`

UnsetSecretKeyOfUser ensures that no value is present for SecretKeyOfUser, not even an explicit nil
### GetSecretKeyOfPassword

`func (o *GitSecretRef) GetSecretKeyOfPassword() string`

GetSecretKeyOfPassword returns the SecretKeyOfPassword field if non-nil, zero value otherwise.

### GetSecretKeyOfPasswordOk

`func (o *GitSecretRef) GetSecretKeyOfPasswordOk() (*string, bool)`

GetSecretKeyOfPasswordOk returns a tuple with the SecretKeyOfPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfPassword

`func (o *GitSecretRef) SetSecretKeyOfPassword(v string)`

SetSecretKeyOfPassword sets SecretKeyOfPassword field to given value.

### HasSecretKeyOfPassword

`func (o *GitSecretRef) HasSecretKeyOfPassword() bool`

HasSecretKeyOfPassword returns a boolean if a field has been set.

### SetSecretKeyOfPasswordNil

`func (o *GitSecretRef) SetSecretKeyOfPasswordNil(b bool)`

 SetSecretKeyOfPasswordNil sets the value for SecretKeyOfPassword to be an explicit nil

### UnsetSecretKeyOfPassword
`func (o *GitSecretRef) UnsetSecretKeyOfPassword()`

UnsetSecretKeyOfPassword ensures that no value is present for SecretKeyOfPassword, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


