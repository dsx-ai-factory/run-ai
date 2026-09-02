# Secret2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Secret** | Pointer to **NullableString** | The name of the Secret resource. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | Local path within the workspace to which the Secret will be mapped to. (mandatory) | [optional] 

## Methods

### NewSecret2

`func NewSecret2() *Secret2`

NewSecret2 instantiates a new Secret2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecret2WithDefaults

`func NewSecret2WithDefaults() *Secret2`

NewSecret2WithDefaults instantiates a new Secret2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSecret

`func (o *Secret2) GetSecret() string`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *Secret2) GetSecretOk() (*string, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *Secret2) SetSecret(v string)`

SetSecret sets Secret field to given value.

### HasSecret

`func (o *Secret2) HasSecret() bool`

HasSecret returns a boolean if a field has been set.

### SetSecretNil

`func (o *Secret2) SetSecretNil(b bool)`

 SetSecretNil sets the value for Secret to be an explicit nil

### UnsetSecret
`func (o *Secret2) UnsetSecret()`

UnsetSecret ensures that no value is present for Secret, not even an explicit nil
### GetMountPath

`func (o *Secret2) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *Secret2) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *Secret2) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *Secret2) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *Secret2) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *Secret2) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


