# ConfigMapAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMap** | Pointer to **NullableString** | The name of the ConfigMap resource. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | Local path within the workspace to which the ConfigMap will be mapped to. (mandatory) | [optional] 

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

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


