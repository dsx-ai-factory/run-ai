# SecretInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**MountPath** | Pointer to **NullableString** | Local path within the workload to which the Secret will be mapped to. (mandatory) | [optional] 
**DefaultMode** | Pointer to **NullableString** | File permission mode in octal string format. This value must be a 4-digit octal number, representing the default file mode when mounting a Secret or ConfigMap as a volume.  | [optional] 
**Secret** | Pointer to **NullableString** | The name of the Secret resource. (mandatory) | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewSecretInstance

`func NewSecretInstance() *SecretInstance`

NewSecretInstance instantiates a new SecretInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretInstanceWithDefaults

`func NewSecretInstanceWithDefaults() *SecretInstance`

NewSecretInstanceWithDefaults instantiates a new SecretInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *SecretInstance) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SecretInstance) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SecretInstance) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *SecretInstance) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *SecretInstance) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *SecretInstance) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetMountPath

`func (o *SecretInstance) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *SecretInstance) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *SecretInstance) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *SecretInstance) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *SecretInstance) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *SecretInstance) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetDefaultMode

`func (o *SecretInstance) GetDefaultMode() string`

GetDefaultMode returns the DefaultMode field if non-nil, zero value otherwise.

### GetDefaultModeOk

`func (o *SecretInstance) GetDefaultModeOk() (*string, bool)`

GetDefaultModeOk returns a tuple with the DefaultMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultMode

`func (o *SecretInstance) SetDefaultMode(v string)`

SetDefaultMode sets DefaultMode field to given value.

### HasDefaultMode

`func (o *SecretInstance) HasDefaultMode() bool`

HasDefaultMode returns a boolean if a field has been set.

### SetDefaultModeNil

`func (o *SecretInstance) SetDefaultModeNil(b bool)`

 SetDefaultModeNil sets the value for DefaultMode to be an explicit nil

### UnsetDefaultMode
`func (o *SecretInstance) UnsetDefaultMode()`

UnsetDefaultMode ensures that no value is present for DefaultMode, not even an explicit nil
### GetSecret

`func (o *SecretInstance) GetSecret() string`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *SecretInstance) GetSecretOk() (*string, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *SecretInstance) SetSecret(v string)`

SetSecret sets Secret field to given value.

### HasSecret

`func (o *SecretInstance) HasSecret() bool`

HasSecret returns a boolean if a field has been set.

### SetSecretNil

`func (o *SecretInstance) SetSecretNil(b bool)`

 SetSecretNil sets the value for Secret to be an explicit nil

### UnsetSecret
`func (o *SecretInstance) UnsetSecret()`

UnsetSecret ensures that no value is present for Secret, not even an explicit nil
### GetExclude

`func (o *SecretInstance) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *SecretInstance) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *SecretInstance) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *SecretInstance) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *SecretInstance) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *SecretInstance) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


