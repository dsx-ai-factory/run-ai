# PortsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attributes** | Pointer to [**NullablePort**](Port.md) |  | [optional] 
**Instances** | Pointer to [**[]Port**](Port.md) | Set of container ports that the workload exposes. | [optional] 

## Methods

### NewPortsDefaults

`func NewPortsDefaults() *PortsDefaults`

NewPortsDefaults instantiates a new PortsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPortsDefaultsWithDefaults

`func NewPortsDefaultsWithDefaults() *PortsDefaults`

NewPortsDefaultsWithDefaults instantiates a new PortsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttributes

`func (o *PortsDefaults) GetAttributes() Port`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *PortsDefaults) GetAttributesOk() (*Port, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *PortsDefaults) SetAttributes(v Port)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *PortsDefaults) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### SetAttributesNil

`func (o *PortsDefaults) SetAttributesNil(b bool)`

 SetAttributesNil sets the value for Attributes to be an explicit nil

### UnsetAttributes
`func (o *PortsDefaults) UnsetAttributes()`

UnsetAttributes ensures that no value is present for Attributes, not even an explicit nil
### GetInstances

`func (o *PortsDefaults) GetInstances() []Port`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *PortsDefaults) GetInstancesOk() (*[]Port, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *PortsDefaults) SetInstances(v []Port)`

SetInstances sets Instances field to given value.

### HasInstances

`func (o *PortsDefaults) HasInstances() bool`

HasInstances returns a boolean if a field has been set.

### SetInstancesNil

`func (o *PortsDefaults) SetInstancesNil(b bool)`

 SetInstancesNil sets the value for Instances to be an explicit nil

### UnsetInstances
`func (o *PortsDefaults) UnsetInstances()`

UnsetInstances ensures that no value is present for Instances, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


