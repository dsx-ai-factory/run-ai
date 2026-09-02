# ConfigMapsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableConfigMapInstance**](ConfigMapInstance.md) |  | [optional] 
**Instances** | Pointer to [**[]ConfigMapInstance**](ConfigMapInstance.md) | Set of config map volumes to use in the workload | [optional] 

## Methods

### NewConfigMapsDefaults

`func NewConfigMapsDefaults() *ConfigMapsDefaults`

NewConfigMapsDefaults instantiates a new ConfigMapsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigMapsDefaultsWithDefaults

`func NewConfigMapsDefaultsWithDefaults() *ConfigMapsDefaults`

NewConfigMapsDefaultsWithDefaults instantiates a new ConfigMapsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *ConfigMapsDefaults) GetAttributes() ConfigMapInstance`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *ConfigMapsDefaults) GetAttributesOk() (*ConfigMapInstance, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *ConfigMapsDefaults) SetAttributes(v ConfigMapInstance)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *ConfigMapsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *ConfigMapsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *ConfigMapsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *ConfigMapsDefaults) GetInstances() []ConfigMapInstance`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *ConfigMapsDefaults) GetInstancesOk() (*[]ConfigMapInstance, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *ConfigMapsDefaults) SetInstances(v []ConfigMapInstance)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *ConfigMapsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *ConfigMapsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *ConfigMapsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


