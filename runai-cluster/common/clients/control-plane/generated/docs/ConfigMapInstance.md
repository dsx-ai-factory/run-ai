# ConfigMapInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 
**ConfigMap** | Pointer to **NullableString** | The name of the ConfigMap resource. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | Local path within the workspace to which the ConfigMap will be mapped to. (mandatory) | [optional] 

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

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


