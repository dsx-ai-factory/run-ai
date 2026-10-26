# GitCommon

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Repository** | Pointer to **NullableString** | URL to a remote Git repository. The content of this repository will be mapped to the container running the workload. (mandatory) | [optional] 
**Branch** | Pointer to **NullableString** | Specific branch to synchronize the repository from. | [optional] 
**Revision** | Pointer to **NullableString** | Specific revision to synchronize the repository from. | [optional] 
**Path** | Pointer to **NullableString** | Local path within the workload to which the Git repository will be mapped (mandatory). | [optional] 

## Methods

### NewGitCommon

`func NewGitCommon() *GitCommon`

NewGitCommon instantiates a new GitCommon object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitCommonWithDefaults

`func NewGitCommonWithDefaults() *GitCommon`

NewGitCommonWithDefaults instantiates a new GitCommon object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepository

`func (o *GitCommon) GetRepository() string`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *GitCommon) GetRepositoryOk() (*string, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *GitCommon) SetRepository(v string)`

SetRepository sets Repository field to given value.

### HasRepository

`func (o *GitCommon) HasRepository() bool`

HasRepository returns a boolean if a field has been set.

### SetRepositoryNil

`func (o *GitCommon) SetRepositoryNil(b bool)`

 SetRepositoryNil sets the value for Repository to be an explicit nil

### UnsetRepository
`func (o *GitCommon) UnsetRepository()`

UnsetRepository ensures that no value is present for Repository, not even an explicit nil
### GetBranch

`func (o *GitCommon) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GitCommon) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GitCommon) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *GitCommon) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### SetBranchNil

`func (o *GitCommon) SetBranchNil(b bool)`

 SetBranchNil sets the value for Branch to be an explicit nil

### UnsetBranch
`func (o *GitCommon) UnsetBranch()`

UnsetBranch ensures that no value is present for Branch, not even an explicit nil
### GetRevision

`func (o *GitCommon) GetRevision() string`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *GitCommon) GetRevisionOk() (*string, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *GitCommon) SetRevision(v string)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *GitCommon) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### SetRevisionNil

`func (o *GitCommon) SetRevisionNil(b bool)`

 SetRevisionNil sets the value for Revision to be an explicit nil

### UnsetRevision
`func (o *GitCommon) UnsetRevision()`

UnsetRevision ensures that no value is present for Revision, not even an explicit nil
### GetPath

`func (o *GitCommon) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *GitCommon) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *GitCommon) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *GitCommon) HasPath() bool`

HasPath returns a boolean if a field has been set.

### SetPathNil

`func (o *GitCommon) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *GitCommon) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


