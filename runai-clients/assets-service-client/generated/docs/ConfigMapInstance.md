# ConfigMapInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**ConfigMap** | Pointer to **NullableString** | The name of the ConfigMap resource. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | Local path within the workload to which the ConfigMap will be mapped to. (mandatory) | [optional] 
**SubPath** | Pointer to **NullableString** | Path within the volume from which the container&#39;s volume should be mounted. | [optional] 
**DefaultMode** | Pointer to **NullableString** | File permission mode in octal string format. This value must be a 4-digit octal number, representing the default file mode when mounting a Secret or ConfigMap as a volume.  | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewConfigMapInstance

`func NewConfigMapInstance() *ConfigMapInstance`

NewConfigMapInstance instantiates a new ConfigMapInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigMapInstanceWithDefaults

`func NewConfigMapInstanceWithDefaults() *ConfigMapInstance`

NewConfigMapInstanceWithDefaults instantiates a new ConfigMapInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ConfigMapInstance) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ConfigMapInstance) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ConfigMapInstance) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ConfigMapInstance) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *ConfigMapInstance) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *ConfigMapInstance) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetConfigMap

`func (o *ConfigMapInstance) GetConfigMap() string`

GetConfigMap returns the ConfigMap field if non-nil, zero value otherwise.

### GetConfigMapOk

`func (o *ConfigMapInstance) GetConfigMapOk() (*string, bool)`

GetConfigMapOk returns a tuple with the ConfigMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMap

`func (o *ConfigMapInstance) SetConfigMap(v string)`

SetConfigMap sets ConfigMap field to given value.

### HasConfigMap

`func (o *ConfigMapInstance) HasConfigMap() bool`

HasConfigMap returns a boolean if a field has been set.

### SetConfigMapNil

`func (o *ConfigMapInstance) SetConfigMapNil(b bool)`

 SetConfigMapNil sets the value for ConfigMap to be an explicit nil

### UnsetConfigMap
`func (o *ConfigMapInstance) UnsetConfigMap()`

UnsetConfigMap ensures that no value is present for ConfigMap, not even an explicit nil
### GetMountPath

`func (o *ConfigMapInstance) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *ConfigMapInstance) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *ConfigMapInstance) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *ConfigMapInstance) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *ConfigMapInstance) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *ConfigMapInstance) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil
### GetSubPath

`func (o *ConfigMapInstance) GetSubPath() string`

GetSubPath returns the SubPath field if non-nil, zero value otherwise.

### GetSubPathOk

`func (o *ConfigMapInstance) GetSubPathOk() (*string, bool)`

GetSubPathOk returns a tuple with the SubPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubPath

`func (o *ConfigMapInstance) SetSubPath(v string)`

SetSubPath sets SubPath field to given value.

### HasSubPath

`func (o *ConfigMapInstance) HasSubPath() bool`

HasSubPath returns a boolean if a field has been set.

### SetSubPathNil

`func (o *ConfigMapInstance) SetSubPathNil(b bool)`

 SetSubPathNil sets the value for SubPath to be an explicit nil

### UnsetSubPath
`func (o *ConfigMapInstance) UnsetSubPath()`

UnsetSubPath ensures that no value is present for SubPath, not even an explicit nil
### GetDefaultMode

`func (o *ConfigMapInstance) GetDefaultMode() string`

GetDefaultMode returns the DefaultMode field if non-nil, zero value otherwise.

### GetDefaultModeOk

`func (o *ConfigMapInstance) GetDefaultModeOk() (*string, bool)`

GetDefaultModeOk returns a tuple with the DefaultMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultMode

`func (o *ConfigMapInstance) SetDefaultMode(v string)`

SetDefaultMode sets DefaultMode field to given value.

### HasDefaultMode

`func (o *ConfigMapInstance) HasDefaultMode() bool`

HasDefaultMode returns a boolean if a field has been set.

### SetDefaultModeNil

`func (o *ConfigMapInstance) SetDefaultModeNil(b bool)`

 SetDefaultModeNil sets the value for DefaultMode to be an explicit nil

### UnsetDefaultMode
`func (o *ConfigMapInstance) UnsetDefaultMode()`

UnsetDefaultMode ensures that no value is present for DefaultMode, not even an explicit nil
### GetExclude

`func (o *ConfigMapInstance) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *ConfigMapInstance) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *ConfigMapInstance) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *ConfigMapInstance) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *ConfigMapInstance) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *ConfigMapInstance) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


