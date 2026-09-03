# DatasourceListResponseAssetSpecGit

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Repository** | **NullableString** | URL to a remote git repository. The content of this repository will be mapped to the container running the workload. Repository name is mandatory for creating a workspace. | 
**Branch** | Pointer to **NullableString** | Specific branch to synchronize the repository from. | [optional] 
**Revision** | Pointer to **NullableString** | Specific revision to synchronize the repository from. | [optional] 
**Path** | **NullableString** | Local path within the workspace to which the S3 bucket will be mapped. Path is mandatory for creating a workspace. | 
**PasswordAssetId** | Pointer to **NullableString** | ID of credentials asset of type password. Needed for non public repository which requires authentication. | [optional] 

## Methods

### NewDatasourceListResponseAssetSpecGit

`func NewDatasourceListResponseAssetSpecGit(repository NullableString, path NullableString, ) *DatasourceListResponseAssetSpecGit`

NewDatasourceListResponseAssetSpecGit instantiates a new DatasourceListResponseAssetSpecGit object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceListResponseAssetSpecGitWithDefaults

`func NewDatasourceListResponseAssetSpecGitWithDefaults() *DatasourceListResponseAssetSpecGit`

NewDatasourceListResponseAssetSpecGitWithDefaults instantiates a new DatasourceListResponseAssetSpecGit object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepository

`func (o *DatasourceListResponseAssetSpecGit) GetRepository() string`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *DatasourceListResponseAssetSpecGit) GetRepositoryOk() (*string, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *DatasourceListResponseAssetSpecGit) SetRepository(v string)`

SetRepository sets Repository field to given value.


### SetRepositoryNil

`func (o *DatasourceListResponseAssetSpecGit) SetRepositoryNil(b bool)`

 SetRepositoryNil sets the value for Repository to be an explicit nil

### UnsetRepository
`func (o *DatasourceListResponseAssetSpecGit) UnsetRepository()`

UnsetRepository ensures that no value is present for Repository, not even an explicit nil
### GetBranch

`func (o *DatasourceListResponseAssetSpecGit) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *DatasourceListResponseAssetSpecGit) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *DatasourceListResponseAssetSpecGit) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *DatasourceListResponseAssetSpecGit) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### SetBranchNil

`func (o *DatasourceListResponseAssetSpecGit) SetBranchNil(b bool)`

 SetBranchNil sets the value for Branch to be an explicit nil

### UnsetBranch
`func (o *DatasourceListResponseAssetSpecGit) UnsetBranch()`

UnsetBranch ensures that no value is present for Branch, not even an explicit nil
### GetRevision

`func (o *DatasourceListResponseAssetSpecGit) GetRevision() string`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *DatasourceListResponseAssetSpecGit) GetRevisionOk() (*string, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *DatasourceListResponseAssetSpecGit) SetRevision(v string)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *DatasourceListResponseAssetSpecGit) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### SetRevisionNil

`func (o *DatasourceListResponseAssetSpecGit) SetRevisionNil(b bool)`

 SetRevisionNil sets the value for Revision to be an explicit nil

### UnsetRevision
`func (o *DatasourceListResponseAssetSpecGit) UnsetRevision()`

UnsetRevision ensures that no value is present for Revision, not even an explicit nil
### GetPath

`func (o *DatasourceListResponseAssetSpecGit) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *DatasourceListResponseAssetSpecGit) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *DatasourceListResponseAssetSpecGit) SetPath(v string)`

SetPath sets Path field to given value.


### SetPathNil

`func (o *DatasourceListResponseAssetSpecGit) SetPathNil(b bool)`

 SetPathNil sets the value for Path to be an explicit nil

### UnsetPath
`func (o *DatasourceListResponseAssetSpecGit) UnsetPath()`

UnsetPath ensures that no value is present for Path, not even an explicit nil
### GetPasswordAssetId

`func (o *DatasourceListResponseAssetSpecGit) GetPasswordAssetId() string`

GetPasswordAssetId returns the PasswordAssetId field if non-nil, zero value otherwise.

### GetPasswordAssetIdOk

`func (o *DatasourceListResponseAssetSpecGit) GetPasswordAssetIdOk() (*string, bool)`

GetPasswordAssetIdOk returns a tuple with the PasswordAssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordAssetId

`func (o *DatasourceListResponseAssetSpecGit) SetPasswordAssetId(v string)`

SetPasswordAssetId sets PasswordAssetId field to given value.

### HasPasswordAssetId

`func (o *DatasourceListResponseAssetSpecGit) HasPasswordAssetId() bool`

HasPasswordAssetId returns a boolean if a field has been set.

### SetPasswordAssetIdNil

`func (o *DatasourceListResponseAssetSpecGit) SetPasswordAssetIdNil(b bool)`

 SetPasswordAssetIdNil sets the value for PasswordAssetId to be an explicit nil

### UnsetPasswordAssetId
`func (o *DatasourceListResponseAssetSpecGit) UnsetPasswordAssetId()`

UnsetPasswordAssetId ensures that no value is present for PasswordAssetId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


