# GitPasswordAsset

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PasswordAssetId** | Pointer to **NullableString** | UUID of a credential asset of type userid / password. This credential is required for authenticating with private Git repositories using SSH. (deprecated - use credentialId instead). | [optional] 

## Methods

### NewGitPasswordAsset

`func NewGitPasswordAsset() *GitPasswordAsset`

NewGitPasswordAsset instantiates a new GitPasswordAsset object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitPasswordAssetWithDefaults

`func NewGitPasswordAssetWithDefaults() *GitPasswordAsset`

NewGitPasswordAssetWithDefaults instantiates a new GitPasswordAsset object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPasswordAssetId

`func (o *GitPasswordAsset) GetPasswordAssetId() string`

GetPasswordAssetId returns the PasswordAssetId field if non-nil, zero value otherwise.

### GetPasswordAssetIdOk

`func (o *GitPasswordAsset) GetPasswordAssetIdOk() (*string, bool)`

GetPasswordAssetIdOk returns a tuple with the PasswordAssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordAssetId

`func (o *GitPasswordAsset) SetPasswordAssetId(v string)`

SetPasswordAssetId sets PasswordAssetId field to given value.

### HasPasswordAssetId

`func (o *GitPasswordAsset) HasPasswordAssetId() bool`

HasPasswordAssetId returns a boolean if a field has been set.

### SetPasswordAssetIdNil

`func (o *GitPasswordAsset) SetPasswordAssetIdNil(b bool)`

 SetPasswordAssetIdNil sets the value for PasswordAssetId to be an explicit nil

### UnsetPasswordAssetId
`func (o *GitPasswordAsset) UnsetPasswordAssetId()`

UnsetPasswordAssetId ensures that no value is present for PasswordAssetId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


