# GitCredentialId

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CredentialId** | Pointer to **NullableString** | UUID of a credential asset of type generic secret or user id / password required for authenticating with private Git repositories. For generic secret, the credential asset must contain a key-value pair where the key is ssh (lowercase) and the value is a valid private SSH key | [optional] 

## Methods

### NewGitCredentialId

`func NewGitCredentialId() *GitCredentialId`

NewGitCredentialId instantiates a new GitCredentialId object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitCredentialIdWithDefaults

`func NewGitCredentialIdWithDefaults() *GitCredentialId`

NewGitCredentialIdWithDefaults instantiates a new GitCredentialId object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCredentialId

`func (o *GitCredentialId) GetCredentialId() string`

GetCredentialId returns the CredentialId field if non-nil, zero value otherwise.

### GetCredentialIdOk

`func (o *GitCredentialId) GetCredentialIdOk() (*string, bool)`

GetCredentialIdOk returns a tuple with the CredentialId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialId

`func (o *GitCredentialId) SetCredentialId(v string)`

SetCredentialId sets CredentialId field to given value.

### HasCredentialId

`func (o *GitCredentialId) HasCredentialId() bool`

HasCredentialId returns a boolean if a field has been set.

### SetCredentialIdNil

`func (o *GitCredentialId) SetCredentialIdNil(b bool)`

 SetCredentialIdNil sets the value for CredentialId to be an explicit nil

### UnsetCredentialId
`func (o *GitCredentialId) UnsetCredentialId()`

UnsetCredentialId ensures that no value is present for CredentialId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


