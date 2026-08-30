# ConfigMapAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMap** | Pointer to **NullableString** | The name of the ConfigMap resource. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | Local path within the workload to which the ConfigMap will be mapped to. (mandatory) | [optional] 
**SubPath** | Pointer to **NullableString** | Path within the volume from which the container&#39;s volume should be mounted. | [optional] 
**DefaultMode** | Pointer to **NullableString** | File permission mode in octal string format. This value must be a 4-digit octal number, representing the default file mode when mounting a Secret or ConfigMap as a volume.  | [optional] 

## Methods

### NewConfigMapAssetSpec

`func NewConfigMapAssetSpec() *ConfigMapAssetSpec`

NewConfigMapAssetSpec instantiates a new ConfigMapAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigMapAssetSpecWithDefaults

`func NewConfigMapAssetSpecWithDefaults() *ConfigMapAssetSpec`

NewConfigMapAssetSpecWithDefaults instantiates a new ConfigMapAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMap

`func (o *ConfigMapAssetSpec) GetConfigMap() string`

GetConfigMap returns the ConfigMap field if non-nil, zero value otherwise.

### GetConfigMapOk

`func (o *ConfigMapAssetSpec) GetConfigMapOk() (*string, bool)`

GetConfigMapOk returns a tuple with the ConfigMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMap

`func (o *ConfigMapAssetSpec) SetConfigMap(v string)`

SetConfigMap sets ConfigMap field to given value.

### HasConfigMap

`func (o *ConfigMapAssetSpec) HasConfigMap() bool`

HasConfigMap returns a boolean if a field has been set.

### SetConfigMapNil

`func (o *ConfigMapAssetSpec) SetConfigMapNil(b bool)`

 SetConfigMapNil sets the value for ConfigMap to be an explicit nil

### UnsetConfigMap
`func (o *ConfigMapAssetSpec) UnsetConfigMap()`

UnsetConfigMap ensures that no value is present for ConfigMap, not even an explicit nil
### GetMountPath

`func (o *ConfigMapAssetSpec) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *ConfigMapAssetSpec) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *ConfigMapAssetSpec) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *ConfigMapAssetSpec) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *ConfigMapAssetSpec) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *ConfigMapAssetSpec) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetSubPath

`func (o *ConfigMapAssetSpec) GetSubPath() string`

GetSubPath returns the SubPath field if non-nil, zero value otherwise.

### GetSubPathOk

`func (o *ConfigMapAssetSpec) GetSubPathOk() (*string, bool)`

GetSubPathOk returns a tuple with the SubPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubPath

`func (o *ConfigMapAssetSpec) SetSubPath(v string)`

SetSubPath sets SubPath field to given value.

### HasSubPath

`func (o *ConfigMapAssetSpec) HasSubPath() bool`

HasSubPath returns a boolean if a field has been set.

### SetSubPathNil

`func (o *ConfigMapAssetSpec) SetSubPathNil(b bool)`

 SetSubPathNil sets the value for SubPath to be an explicit nil

### UnsetSubPath
`func (o *ConfigMapAssetSpec) UnsetSubPath()`

UnsetSubPath ensures that no value is present for SubPath, not even an explicit nil
### GetDefaultMode

`func (o *ConfigMapAssetSpec) GetDefaultMode() string`

GetDefaultMode returns the DefaultMode field if non-nil, zero value otherwise.

### GetDefaultModeOk

`func (o *ConfigMapAssetSpec) GetDefaultModeOk() (*string, bool)`

GetDefaultModeOk returns a tuple with the DefaultMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultMode

`func (o *ConfigMapAssetSpec) SetDefaultMode(v string)`

SetDefaultMode sets DefaultMode field to given value.

### HasDefaultMode

`func (o *ConfigMapAssetSpec) HasDefaultMode() bool`

HasDefaultMode returns a boolean if a field has been set.

### SetDefaultModeNil

`func (o *ConfigMapAssetSpec) SetDefaultModeNil(b bool)`

 SetDefaultModeNil sets the value for DefaultMode to be an explicit nil

### UnsetDefaultMode
`func (o *ConfigMapAssetSpec) UnsetDefaultMode()`

UnsetDefaultMode ensures that no value is present for DefaultMode, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


