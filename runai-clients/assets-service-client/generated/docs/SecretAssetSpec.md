# SecretAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CredentialAssetId** | Pointer to **NullableString** | ID of credentials asset (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | Local path within the workspace to which the Secret will be mapped to (mandatory) | [optional] 
**DefaultMode** | Pointer to **NullableString** | File permission mode in octal string format. This value must be a 4-digit octal number, representing the default file mode when mounting a Secret or ConfigMap as a volume.  | [optional] 

## Methods

### NewSecretAssetSpec

`func NewSecretAssetSpec() *SecretAssetSpec`

NewSecretAssetSpec instantiates a new SecretAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretAssetSpecWithDefaults

`func NewSecretAssetSpecWithDefaults() *SecretAssetSpec`

NewSecretAssetSpecWithDefaults instantiates a new SecretAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCredentialAssetId

`func (o *SecretAssetSpec) GetCredentialAssetId() string`

GetCredentialAssetId returns the CredentialAssetId field if non-nil, zero value otherwise.

### GetCredentialAssetIdOk

`func (o *SecretAssetSpec) GetCredentialAssetIdOk() (*string, bool)`

GetCredentialAssetIdOk returns a tuple with the CredentialAssetId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialAssetId

`func (o *SecretAssetSpec) SetCredentialAssetId(v string)`

SetCredentialAssetId sets CredentialAssetId field to given value.

### HasCredentialAssetId

`func (o *SecretAssetSpec) HasCredentialAssetId() bool`

HasCredentialAssetId returns a boolean if a field has been set.

### SetCredentialAssetIdNil

`func (o *SecretAssetSpec) SetCredentialAssetIdNil(b bool)`

 SetCredentialAssetIdNil sets the value for CredentialAssetId to be an explicit nil

### UnsetCredentialAssetId
`func (o *SecretAssetSpec) UnsetCredentialAssetId()`

UnsetCredentialAssetId ensures that no value is present for CredentialAssetId, not even an explicit nil
### GetMountPath

`func (o *SecretAssetSpec) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *SecretAssetSpec) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *SecretAssetSpec) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *SecretAssetSpec) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *SecretAssetSpec) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *SecretAssetSpec) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetDefaultMode

`func (o *SecretAssetSpec) GetDefaultMode() string`

GetDefaultMode returns the DefaultMode field if non-nil, zero value otherwise.

### GetDefaultModeOk

`func (o *SecretAssetSpec) GetDefaultModeOk() (*string, bool)`

GetDefaultModeOk returns a tuple with the DefaultMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultMode

`func (o *SecretAssetSpec) SetDefaultMode(v string)`

SetDefaultMode sets DefaultMode field to given value.

### HasDefaultMode

`func (o *SecretAssetSpec) HasDefaultMode() bool`

HasDefaultMode returns a boolean if a field has been set.

### SetDefaultModeNil

`func (o *SecretAssetSpec) SetDefaultModeNil(b bool)`

 SetDefaultModeNil sets the value for DefaultMode to be an explicit nil

### UnsetDefaultMode
`func (o *SecretAssetSpec) UnsetDefaultMode()`

UnsetDefaultMode ensures that no value is present for DefaultMode, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


