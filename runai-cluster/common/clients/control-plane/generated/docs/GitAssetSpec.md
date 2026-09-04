# GitAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Repository** | Pointer to **NullableString** | URL to a remote Git repository. The content of this repository will be mapped to the container running the workload. (mandatory) | [optional] 
**Branch** | Pointer to **NullableString** | Specific branch to synchronize the repository from. | [optional] 
**Revision** | Pointer to **NullableString** | Specific revision to synchronize the repository from. | [optional] 
**Path** | Pointer to **NullableString** | Local path within the workspace to which the Git repository will be mapped (mandatory). | [optional] 
**PasswordAssetId** | Pointer to **NullableString** | ID of credentials asset of type password. Needed for non public repository which requires authentication. | [optional] 

## Methods

### NewGitAssetSpec

`func NewGitAssetSpec() *GitAssetSpec`

NewGitAssetSpec instantiates a new GitAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitAssetSpecWithDefaults

`func NewGitAssetSpecWithDefaults() *GitAssetSpec`

NewGitAssetSpecWithDefaults instantiates a new GitAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepository

`func (o *GitAssetSpec) GetRepository() string`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *GitAssetSpec) GetRepositoryOk() (*string, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *GitAssetSpec) SetRepository(v string)`

SetRepository sets Repository field to given value.

### HasRepository

`func (o *GitAssetSpec) HasRepository() bool`

HasRepository returns a boolean if a field has been set.

### SetRepositoryNil

`func (o *GitAssetSpec) SetRepositoryNil(b bool)`

 SetRepositoryNil sets the value for Repository to be an explicit nil

### UnsetRepository
`func (o *GitAssetSpec) UnsetRepository()`

UnsetRepository ensures that no value is present for Repository, not even an explicit nil
### GetBranch

`func (o *GitAssetSpec) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GitAssetSpec) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GitAssetSpec) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *GitAssetSpec) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### SetBranchNil

`func (o *GitAssetSpec) SetBranchNil(b bool)`

 SetBranchNil sets the value for Branch to be an explicit nil

### UnsetBranch
`func (o *GitAssetSpec) UnsetBranch()`

UnsetBranch ensures that no value is present for Branch, not even an explicit nil
### GetRevision

`func (o *GitAssetSpec) GetRevision() string`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *GitAssetSpec) GetRevisionOk() (*string, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *GitAssetSpec) SetRevision(v string)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *GitAssetSpec) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### SetRevisionNil

`func (o *GitAssetSpec) SetRevisionNil(b bool)`

 SetRevisionNil sets the value for Revision to be an explicit nil

### UnsetRevision
`func (o *GitAssetSpec) UnsetRevision()`

UnsetRevision ensures that no value is present for Revision, not even an explicit nil
### GetPath

`func (o *GitAssetSpec) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitAssetSpec) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitAssetSpec) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *GitAssetSpec) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *GitAssetSpec) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *GitAssetSpec) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetPasswordAssetId

`func (o *GitAssetSpec) GetPasswordAssetId() string`

GetPasswordAssetId returns the PasswordAssetId field if non-nil, zero value otherwise.

### GetPasswordAssetIdOk

`func (o *GitAssetSpec) GetPasswordAssetIdOk() (*string, bool)`

GetPasswordAssetIdOk returns a tuple with the PasswordAssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordAssetId

`func (o *GitAssetSpec) SetPasswordAssetId(v string)`

SetPasswordAssetId sets PasswordAssetId field to given value.

### HasPasswordAssetId

`func (o *GitAssetSpec) HasPasswordAssetId() bool`

HasPasswordAssetId returns a boolean if a field has been set.

### SetPasswordAssetIdNil

`func (o *GitAssetSpec) SetPasswordAssetIdNil(b bool)`

 SetPasswordAssetIdNil sets the value for PasswordAssetId to be an explicit nil

### UnsetPasswordAssetId
`func (o *GitAssetSpec) UnsetPasswordAssetId()`

UnsetPasswordAssetId ensures that no value is present for PasswordAssetId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


