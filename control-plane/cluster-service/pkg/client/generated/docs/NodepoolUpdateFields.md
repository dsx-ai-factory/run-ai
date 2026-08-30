# NodepoolUpdateFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LabelKey** | Pointer to **NullableString** | Label key for associated nodes to the Node Pool (with value as in labelValue) | [optional] 
**LabelValue** | Pointer to **NullableString** | Label value for associated nodes to the Node Pool (with key as in labelKey) | [optional] 
**SchedulingCapabilities** | Pointer to [**SchedulingCapabilitiesNullable**](SchedulingCapabilitiesNullable.md) |  | [optional] 
**GpuNetworkAccelerationLabelKey** | Pointer to **NullableString** | Label key by which to determine GPUNetworkAccelerationDetection nodes | [optional] 
**GpuNetworkAccelerationDetection** | Pointer to [**NullableNodepoolGPUNetworkAccelerationDetection**](NodepoolGPUNetworkAccelerationDetection.md) |  | [optional] [default to NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO]
**SchedulingConfiguration** | Pointer to [**NullableSchedulingConfiguration**](SchedulingConfiguration.md) |  | [optional] 
**NetworkTopologyId** | Pointer to **NullableString** | The unique identifier for the network topology | [optional] 

## Methods

### NewNodepoolUpdateFields

`func NewNodepoolUpdateFields() *NodepoolUpdateFields`

NewNodepoolUpdateFields instantiates a new NodepoolUpdateFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolUpdateFieldsWithDefaults

`func NewNodepoolUpdateFieldsWithDefaults() *NodepoolUpdateFields`

NewNodepoolUpdateFieldsWithDefaults instantiates a new NodepoolUpdateFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabelKey

`func (o *NodepoolUpdateFields) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NodepoolUpdateFields) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NodepoolUpdateFields) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *NodepoolUpdateFields) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### SetLabelKeyNil

`func (o *NodepoolUpdateFields) SetLabelKeyNil(b bool)`

 SetLabelKeyNil sets the value for LabelKey to be an explicit nil

### UnsetLabelKey
`func (o *NodepoolUpdateFields) UnsetLabelKey()`

UnsetLabelKey ensures that no value is present for LabelKey, not even an explicit nil
### GetLabelValue

`func (o *NodepoolUpdateFields) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NodepoolUpdateFields) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NodepoolUpdateFields) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *NodepoolUpdateFields) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### SetLabelValueNil

`func (o *NodepoolUpdateFields) SetLabelValueNil(b bool)`

 SetLabelValueNil sets the value for LabelValue to be an explicit nil

### UnsetLabelValue
`func (o *NodepoolUpdateFields) UnsetLabelValue()`

UnsetLabelValue ensures that no value is present for LabelValue, not even an explicit nil
### GetSchedulingCapabilities

`func (o *NodepoolUpdateFields) GetSchedulingCapabilities() SchedulingCapabilitiesNullable`

GetSchedulingCapabilities returns the SchedulingCapabilities field if non-nil, zero value otherwise.

### GetSchedulingCapabilitiesOk

`func (o *NodepoolUpdateFields) GetSchedulingCapabilitiesOk() (*SchedulingCapabilitiesNullable, bool)`

GetSchedulingCapabilitiesOk returns a tuple with the SchedulingCapabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingCapabilities

`func (o *NodepoolUpdateFields) SetSchedulingCapabilities(v SchedulingCapabilitiesNullable)`

SetSchedulingCapabilities sets SchedulingCapabilities field to given value.

### HasSchedulingCapabilities

`func (o *NodepoolUpdateFields) HasSchedulingCapabilities() bool`

HasSchedulingCapabilities returns a boolean if a field has been set.

### GetGpuNetworkAccelerationLabelKey

`func (o *NodepoolUpdateFields) GetGpuNetworkAccelerationLabelKey() string`

GetGpuNetworkAccelerationLabelKey returns the GpuNetworkAccelerationLabelKey field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationLabelKeyOk

`func (o *NodepoolUpdateFields) GetGpuNetworkAccelerationLabelKeyOk() (*string, bool)`

GetGpuNetworkAccelerationLabelKeyOk returns a tuple with the GpuNetworkAccelerationLabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationLabelKey

`func (o *NodepoolUpdateFields) SetGpuNetworkAccelerationLabelKey(v string)`

SetGpuNetworkAccelerationLabelKey sets GpuNetworkAccelerationLabelKey field to given value.

### HasGpuNetworkAccelerationLabelKey

`func (o *NodepoolUpdateFields) HasGpuNetworkAccelerationLabelKey() bool`

HasGpuNetworkAccelerationLabelKey returns a boolean if a field has been set.

### SetGpuNetworkAccelerationLabelKeyNil

`func (o *NodepoolUpdateFields) SetGpuNetworkAccelerationLabelKeyNil(b bool)`

 SetGpuNetworkAccelerationLabelKeyNil sets the value for GpuNetworkAccelerationLabelKey to be an explicit nil

### UnsetGpuNetworkAccelerationLabelKey
`func (o *NodepoolUpdateFields) UnsetGpuNetworkAccelerationLabelKey()`

UnsetGpuNetworkAccelerationLabelKey ensures that no value is present for GpuNetworkAccelerationLabelKey, not even an explicit nil
### GetGpuNetworkAccelerationDetection

`func (o *NodepoolUpdateFields) GetGpuNetworkAccelerationDetection() NodepoolGPUNetworkAccelerationDetection`

GetGpuNetworkAccelerationDetection returns the GpuNetworkAccelerationDetection field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationDetectionOk

`func (o *NodepoolUpdateFields) GetGpuNetworkAccelerationDetectionOk() (*NodepoolGPUNetworkAccelerationDetection, bool)`

GetGpuNetworkAccelerationDetectionOk returns a tuple with the GpuNetworkAccelerationDetection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationDetection

`func (o *NodepoolUpdateFields) SetGpuNetworkAccelerationDetection(v NodepoolGPUNetworkAccelerationDetection)`

SetGpuNetworkAccelerationDetection sets GpuNetworkAccelerationDetection field to given value.

### HasGpuNetworkAccelerationDetection

`func (o *NodepoolUpdateFields) HasGpuNetworkAccelerationDetection() bool`

HasGpuNetworkAccelerationDetection returns a boolean if a field has been set.

### SetGpuNetworkAccelerationDetectionNil

`func (o *NodepoolUpdateFields) SetGpuNetworkAccelerationDetectionNil(b bool)`

 SetGpuNetworkAccelerationDetectionNil sets the value for GpuNetworkAccelerationDetection to be an explicit nil

### UnsetGpuNetworkAccelerationDetection
`func (o *NodepoolUpdateFields) UnsetGpuNetworkAccelerationDetection()`

UnsetGpuNetworkAccelerationDetection ensures that no value is present for GpuNetworkAccelerationDetection, not even an explicit nil
### GetSchedulingConfiguration

`func (o *NodepoolUpdateFields) GetSchedulingConfiguration() SchedulingConfiguration`

GetSchedulingConfiguration returns the SchedulingConfiguration field if non-nil, zero value otherwise.

### GetSchedulingConfigurationOk

`func (o *NodepoolUpdateFields) GetSchedulingConfigurationOk() (*SchedulingConfiguration, bool)`

GetSchedulingConfigurationOk returns a tuple with the SchedulingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingConfiguration

`func (o *NodepoolUpdateFields) SetSchedulingConfiguration(v SchedulingConfiguration)`

SetSchedulingConfiguration sets SchedulingConfiguration field to given value.

### HasSchedulingConfiguration

`func (o *NodepoolUpdateFields) HasSchedulingConfiguration() bool`

HasSchedulingConfiguration returns a boolean if a field has been set.

### SetSchedulingConfigurationNil

`func (o *NodepoolUpdateFields) SetSchedulingConfigurationNil(b bool)`

 SetSchedulingConfigurationNil sets the value for SchedulingConfiguration to be an explicit nil

### UnsetSchedulingConfiguration
`func (o *NodepoolUpdateFields) UnsetSchedulingConfiguration()`

UnsetSchedulingConfiguration ensures that no value is present for SchedulingConfiguration, not even an explicit nil
### GetNetworkTopologyId

`func (o *NodepoolUpdateFields) GetNetworkTopologyId() string`

GetNetworkTopologyId returns the NetworkTopologyId field if non-nil, zero value otherwise.

### GetNetworkTopologyIdOk

`func (o *NodepoolUpdateFields) GetNetworkTopologyIdOk() (*string, bool)`

GetNetworkTopologyIdOk returns a tuple with the NetworkTopologyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyId

`func (o *NodepoolUpdateFields) SetNetworkTopologyId(v string)`

SetNetworkTopologyId sets NetworkTopologyId field to given value.

### HasNetworkTopologyId

`func (o *NodepoolUpdateFields) HasNetworkTopologyId() bool`

HasNetworkTopologyId returns a boolean if a field has been set.

### SetNetworkTopologyIdNil

`func (o *NodepoolUpdateFields) SetNetworkTopologyIdNil(b bool)`

 SetNetworkTopologyIdNil sets the value for NetworkTopologyId to be an explicit nil

### UnsetNetworkTopologyId
`func (o *NodepoolUpdateFields) UnsetNetworkTopologyId()`

UnsetNetworkTopologyId ensures that no value is present for NetworkTopologyId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


