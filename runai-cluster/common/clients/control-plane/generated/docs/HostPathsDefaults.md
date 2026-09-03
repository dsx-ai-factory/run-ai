# HostPathsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullableHostPathInstance**](HostPathInstance.md) |  | [optional] 
**Instances** | Pointer to [**[]HostPathInstance**](HostPathInstance.md) | Set of host paths to use in the workload. | [optional] 

## Methods

### NewHostPathsDefaults

`func NewHostPathsDefaults() *HostPathsDefaults`

NewHostPathsDefaults instantiates a new HostPathsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHostPathsDefaultsWithDefaults

`func NewHostPathsDefaultsWithDefaults() *HostPathsDefaults`

NewHostPathsDefaultsWithDefaults instantiates a new HostPathsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *HostPathsDefaults) GetAttributes() HostPathInstance`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *HostPathsDefaults) GetAttributesOk() (*HostPathInstance, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *HostPathsDefaults) SetAttributes(v HostPathInstance)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *HostPathsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *HostPathsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *HostPathsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *HostPathsDefaults) GetInstances() []HostPathInstance`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *HostPathsDefaults) GetInstancesOk() (*[]HostPathInstance, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *HostPathsDefaults) SetInstances(v []HostPathInstance)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *HostPathsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *HostPathsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *HostPathsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


