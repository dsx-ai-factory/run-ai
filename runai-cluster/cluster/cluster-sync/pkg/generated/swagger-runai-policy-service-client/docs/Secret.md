# Secret

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MountPath** | Pointer to **NullableString** | Local path within the workspace to which the Secret will be mapped to. (mandatory) | [optional] 
**Secret** | Pointer to **NullableString** | The name of the Secret resource. (mandatory) | [optional] 

## Methods

### NewSecret

`func NewSecret() *Secret`

NewSecret instantiates a new Secret object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretWithDefaults

`func NewSecretWithDefaults() *Secret`

NewSecretWithDefaults instantiates a new Secret object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMountPath

`func (o *Secret) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *Secret) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *Secret) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *Secret) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *Secret) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *Secret) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetSecret

`func (o *Secret) GetSecret() string`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *Secret) GetSecretOk() (*string, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *Secret) SetSecret(v string)`

SetSecret sets Secret field to given value.

### HasSecret

`func (o *Secret) HasSecret() bool`

HasSecret returns a boolean if a field has been set.

### SetSecretNil

`func (o *Secret) SetSecretNil(b bool)`

 SetSecretNil sets the value for Secret to be an explicit nil

### UnsetSecret
`func (o *Secret) UnsetSecret()`

UnsetSecret ensures that no value is present for Secret, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


