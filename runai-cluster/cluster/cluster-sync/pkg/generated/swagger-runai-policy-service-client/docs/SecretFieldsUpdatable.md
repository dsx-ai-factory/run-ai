# SecretFieldsUpdatable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MountPath** | Pointer to **NullableString** | Local path within the workspace to which the Secret will be mapped to. (mandatory) | [optional] 

## Methods

### NewSecretFieldsUpdatable

`func NewSecretFieldsUpdatable() *SecretFieldsUpdatable`

NewSecretFieldsUpdatable instantiates a new SecretFieldsUpdatable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretFieldsUpdatableWithDefaults

`func NewSecretFieldsUpdatableWithDefaults() *SecretFieldsUpdatable`

NewSecretFieldsUpdatableWithDefaults instantiates a new SecretFieldsUpdatable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMountPath

`func (o *SecretFieldsUpdatable) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *SecretFieldsUpdatable) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *SecretFieldsUpdatable) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *SecretFieldsUpdatable) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *SecretFieldsUpdatable) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *SecretFieldsUpdatable) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


