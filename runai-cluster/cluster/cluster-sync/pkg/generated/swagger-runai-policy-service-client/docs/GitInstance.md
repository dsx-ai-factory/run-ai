# GitInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**Repository** | Pointer to **NullableString** | URL to a remote Git repository. The content of this repository will be mapped to the container running the workload. (mandatory) | [optional] 
**Branch** | Pointer to **NullableString** | Specific branch to synchronize the repository from. | [optional] 
**Revision** | Pointer to **NullableString** | Specific revision to synchronize the repository from. | [optional] 
**Path** | Pointer to **NullableString** | Local path within the workspace to which the Git repository will be mapped (mandatory). | [optional] 
**PasswordSecret** | Pointer to **NullableString** | Secret containing the credentials of the repository (needed for non public repository which requires authentication). | [optional] 
**SecretKeyOfUser** | Pointer to **NullableString** | The key to use for loading the user name from the secret. The default is &#x60;User&#x60;. | [optional] 
**SecretKeyOfPassword** | Pointer to **NullableString** | The key to use for loading the password from the secret. The default is &#x60;Password&#x60;. | [optional] 

## Methods

### NewGitInstance

`func NewGitInstance() *GitInstance`

NewGitInstance instantiates a new GitInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitInstanceWithDefaults

`func NewGitInstanceWithDefaults() *GitInstance`

NewGitInstanceWithDefaults instantiates a new GitInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *GitInstance) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GitInstance) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GitInstance) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GitInstance) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *GitInstance) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *GitInstance) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetRepository

`func (o *GitInstance) GetRepository() string`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *GitInstance) GetRepositoryOk() (*string, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *GitInstance) SetRepository(v string)`

SetRepository sets Repository field to given value.

### HasRepository

`func (o *GitInstance) HasRepository() bool`

HasRepository returns a boolean if a field has been set.

### SetRepositoryNil

`func (o *GitInstance) SetRepositoryNil(b bool)`

 SetRepositoryNil sets the value for Repository to be an explicit nil

### UnsetRepository
`func (o *GitInstance) UnsetRepository()`

UnsetRepository ensures that no value is present for Repository, not even an explicit nil
### GetBranch

`func (o *GitInstance) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GitInstance) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GitInstance) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *GitInstance) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### SetBranchNil

`func (o *GitInstance) SetBranchNil(b bool)`

 SetBranchNil sets the value for Branch to be an explicit nil

### UnsetBranch
`func (o *GitInstance) UnsetBranch()`

UnsetBranch ensures that no value is present for Branch, not even an explicit nil
### GetRevision

`func (o *GitInstance) GetRevision() string`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *GitInstance) GetRevisionOk() (*string, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *GitInstance) SetRevision(v string)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *GitInstance) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### SetRevisionNil

`func (o *GitInstance) SetRevisionNil(b bool)`

 SetRevisionNil sets the value for Revision to be an explicit nil

### UnsetRevision
`func (o *GitInstance) UnsetRevision()`

UnsetRevision ensures that no value is present for Revision, not even an explicit nil
### GetPath

`func (o *GitInstance) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitInstance) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitInstance) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *GitInstance) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *GitInstance) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *GitInstance) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetPasswordSecret

`func (o *GitInstance) GetPasswordSecret() string`

GetPasswordSecret returns the PasswordSecret field if non-nil, zero value otherwise.

### GetPasswordSecretOk

`func (o *GitInstance) GetPasswordSecretOk() (*string, bool)`

GetPasswordSecretOk returns a tuple with the PasswordSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordSecret

`func (o *GitInstance) SetPasswordSecret(v string)`

SetPasswordSecret sets PasswordSecret field to given value.

### HasPasswordSecret

`func (o *GitInstance) HasPasswordSecret() bool`

HasPasswordSecret returns a boolean if a field has been set.

### SetPasswordSecretNil

`func (o *GitInstance) SetPasswordSecretNil(b bool)`

 SetPasswordSecretNil sets the value for PasswordSecret to be an explicit nil

### UnsetPasswordSecret
`func (o *GitInstance) UnsetPasswordSecret()`

UnsetPasswordSecret ensures that no value is present for PasswordSecret, not even an explicit nil
### GetSecretKeyOfUser

`func (o *GitInstance) GetSecretKeyOfUser() string`

GetSecretKeyOfUser returns the SecretKeyOfUser field if non-nil, zero value otherwise.

### GetSecretKeyOfUserOk

`func (o *GitInstance) GetSecretKeyOfUserOk() (*string, bool)`

GetSecretKeyOfUserOk returns a tuple with the SecretKeyOfUser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfUser

`func (o *GitInstance) SetSecretKeyOfUser(v string)`

SetSecretKeyOfUser sets SecretKeyOfUser field to given value.

### HasSecretKeyOfUser

`func (o *GitInstance) HasSecretKeyOfUser() bool`

HasSecretKeyOfUser returns a boolean if a field has been set.

### SetSecretKeyOfUserNil

`func (o *GitInstance) SetSecretKeyOfUserNil(b bool)`

 SetSecretKeyOfUserNil sets the value for SecretKeyOfUser to be an explicit nil

### UnsetSecretKeyOfUser
`func (o *GitInstance) UnsetSecretKeyOfUser()`

UnsetSecretKeyOfUser ensures that no value is present for SecretKeyOfUser, not even an explicit nil
### GetSecretKeyOfPassword

`func (o *GitInstance) GetSecretKeyOfPassword() string`

GetSecretKeyOfPassword returns the SecretKeyOfPassword field if non-nil, zero value otherwise.

### GetSecretKeyOfPasswordOk

`func (o *GitInstance) GetSecretKeyOfPasswordOk() (*string, bool)`

GetSecretKeyOfPasswordOk returns a tuple with the SecretKeyOfPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretKeyOfPassword

`func (o *GitInstance) SetSecretKeyOfPassword(v string)`

SetSecretKeyOfPassword sets SecretKeyOfPassword field to given value.

### HasSecretKeyOfPassword

`func (o *GitInstance) HasSecretKeyOfPassword() bool`

HasSecretKeyOfPassword returns a boolean if a field has been set.

### SetSecretKeyOfPasswordNil

`func (o *GitInstance) SetSecretKeyOfPasswordNil(b bool)`

 SetSecretKeyOfPasswordNil sets the value for SecretKeyOfPassword to be an explicit nil

### UnsetSecretKeyOfPassword
`func (o *GitInstance) UnsetSecretKeyOfPassword()`

UnsetSecretKeyOfPassword ensures that no value is present for SecretKeyOfPassword, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


