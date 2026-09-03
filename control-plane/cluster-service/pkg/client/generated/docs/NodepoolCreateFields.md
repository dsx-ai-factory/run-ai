# NodepoolCreateFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**LabelKey** | **string** | Label key for associated nodes to the Node Pool (with value as in labelValue) | 
**LabelValue** | **string** | Label value for associated nodes to the Node Pool (with key as in labelKey) | 
**ClusterId** | **string** |  | 
**SchedulingCapabilities** | Pointer to [**SchedulingCapabilitiesNullable**](SchedulingCapabilitiesNullable.md) |  | [optional] 
**GpuNetworkAccelerationLabelKey** | Pointer to **NullableString** | Label key by which to determine GPUNetworkAccelerationDetection nodes | [optional] 
**GpuNetworkAccelerationDetection** | Pointer to [**NullableNodepoolGPUNetworkAccelerationDetection**](NodepoolGPUNetworkAccelerationDetection.md) |  | [optional] [default to NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO]
**SchedulingConfiguration** | Pointer to [**NullableSchedulingConfiguration**](SchedulingConfiguration.md) |  | [optional] 
**NetworkTopologyId** | Pointer to **NullableString** | The unique identifier for the network topology | [optional] 

## Methods

### NewNodepoolCreateFields

`func NewNodepoolCreateFields(name string, labelKey string, labelValue string, clusterId string, ) *NodepoolCreateFields`

NewNodepoolCreateFields instantiates a new NodepoolCreateFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolCreateFieldsWithDefaults

`func NewNodepoolCreateFieldsWithDefaults() *NodepoolCreateFields`

NewNodepoolCreateFieldsWithDefaults instantiates a new NodepoolCreateFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodepoolCreateFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodepoolCreateFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodepoolCreateFields) SetName(v string)`

SetName sets Name field to given value.


### GetLabelKey

`func (o *NodepoolCreateFields) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NodepoolCreateFields) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NodepoolCreateFields) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.


### GetLabelValue

`func (o *NodepoolCreateFields) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NodepoolCreateFields) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NodepoolCreateFields) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.


### GetClusterId

`func (o *NodepoolCreateFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *NodepoolCreateFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *NodepoolCreateFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetSchedulingCapabilities

`func (o *NodepoolCreateFields) GetSchedulingCapabilities() SchedulingCapabilitiesNullable`

GetSchedulingCapabilities returns the SchedulingCapabilities field if non-nil, zero value otherwise.

### GetSchedulingCapabilitiesOk

`func (o *NodepoolCreateFields) GetSchedulingCapabilitiesOk() (*SchedulingCapabilitiesNullable, bool)`

GetSchedulingCapabilitiesOk returns a tuple with the SchedulingCapabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingCapabilities

`func (o *NodepoolCreateFields) SetSchedulingCapabilities(v SchedulingCapabilitiesNullable)`

SetSchedulingCapabilities sets SchedulingCapabilities field to given value.

### HasSchedulingCapabilities

`func (o *NodepoolCreateFields) HasSchedulingCapabilities() bool`

HasSchedulingCapabilities returns a boolean if a field has been set.

### GetGpuNetworkAccelerationLabelKey

`func (o *NodepoolCreateFields) GetGpuNetworkAccelerationLabelKey() string`

GetGpuNetworkAccelerationLabelKey returns the GpuNetworkAccelerationLabelKey field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationLabelKeyOk

`func (o *NodepoolCreateFields) GetGpuNetworkAccelerationLabelKeyOk() (*string, bool)`

GetGpuNetworkAccelerationLabelKeyOk returns a tuple with the GpuNetworkAccelerationLabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationLabelKey

`func (o *NodepoolCreateFields) SetGpuNetworkAccelerationLabelKey(v string)`

SetGpuNetworkAccelerationLabelKey sets GpuNetworkAccelerationLabelKey field to given value.

### HasGpuNetworkAccelerationLabelKey

`func (o *NodepoolCreateFields) HasGpuNetworkAccelerationLabelKey() bool`

HasGpuNetworkAccelerationLabelKey returns a boolean if a field has been set.

### SetGpuNetworkAccelerationLabelKeyNil

`func (o *NodepoolCreateFields) SetGpuNetworkAccelerationLabelKeyNil(b bool)`

 SetGpuNetworkAccelerationLabelKeyNil sets the value for GpuNetworkAccelerationLabelKey to be an explicit nil

### UnsetGpuNetworkAccelerationLabelKey
`func (o *NodepoolCreateFields) UnsetGpuNetworkAccelerationLabelKey()`

UnsetGpuNetworkAccelerationLabelKey ensures that no value is present for GpuNetworkAccelerationLabelKey, not even an explicit nil
### GetGpuNetworkAccelerationDetection

`func (o *NodepoolCreateFields) GetGpuNetworkAccelerationDetection() NodepoolGPUNetworkAccelerationDetection`

GetGpuNetworkAccelerationDetection returns the GpuNetworkAccelerationDetection field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationDetectionOk

`func (o *NodepoolCreateFields) GetGpuNetworkAccelerationDetectionOk() (*NodepoolGPUNetworkAccelerationDetection, bool)`

GetGpuNetworkAccelerationDetectionOk returns a tuple with the GpuNetworkAccelerationDetection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationDetection

`func (o *NodepoolCreateFields) SetGpuNetworkAccelerationDetection(v NodepoolGPUNetworkAccelerationDetection)`

SetGpuNetworkAccelerationDetection sets GpuNetworkAccelerationDetection field to given value.

### HasGpuNetworkAccelerationDetection

`func (o *NodepoolCreateFields) HasGpuNetworkAccelerationDetection() bool`

HasGpuNetworkAccelerationDetection returns a boolean if a field has been set.

### SetGpuNetworkAccelerationDetectionNil

`func (o *NodepoolCreateFields) SetGpuNetworkAccelerationDetectionNil(b bool)`

 SetGpuNetworkAccelerationDetectionNil sets the value for GpuNetworkAccelerationDetection to be an explicit nil

### UnsetGpuNetworkAccelerationDetection
`func (o *NodepoolCreateFields) UnsetGpuNetworkAccelerationDetection()`

UnsetGpuNetworkAccelerationDetection ensures that no value is present for GpuNetworkAccelerationDetection, not even an explicit nil
### GetSchedulingConfiguration

`func (o *NodepoolCreateFields) GetSchedulingConfiguration() SchedulingConfiguration`

GetSchedulingConfiguration returns the SchedulingConfiguration field if non-nil, zero value otherwise.

### GetSchedulingConfigurationOk

`func (o *NodepoolCreateFields) GetSchedulingConfigurationOk() (*SchedulingConfiguration, bool)`

GetSchedulingConfigurationOk returns a tuple with the SchedulingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingConfiguration

`func (o *NodepoolCreateFields) SetSchedulingConfiguration(v SchedulingConfiguration)`

SetSchedulingConfiguration sets SchedulingConfiguration field to given value.

### HasSchedulingConfiguration

`func (o *NodepoolCreateFields) HasSchedulingConfiguration() bool`

HasSchedulingConfiguration returns a boolean if a field has been set.

### SetSchedulingConfigurationNil

`func (o *NodepoolCreateFields) SetSchedulingConfigurationNil(b bool)`

 SetSchedulingConfigurationNil sets the value for SchedulingConfiguration to be an explicit nil

### UnsetSchedulingConfiguration
`func (o *NodepoolCreateFields) UnsetSchedulingConfiguration()`

UnsetSchedulingConfiguration ensures that no value is present for SchedulingConfiguration, not even an explicit nil
### GetNetworkTopologyId

`func (o *NodepoolCreateFields) GetNetworkTopologyId() string`

GetNetworkTopologyId returns the NetworkTopologyId field if non-nil, zero value otherwise.

### GetNetworkTopologyIdOk

`func (o *NodepoolCreateFields) GetNetworkTopologyIdOk() (*string, bool)`

GetNetworkTopologyIdOk returns a tuple with the NetworkTopologyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyId

`func (o *NodepoolCreateFields) SetNetworkTopologyId(v string)`

SetNetworkTopologyId sets NetworkTopologyId field to given value.

### HasNetworkTopologyId

`func (o *NodepoolCreateFields) HasNetworkTopologyId() bool`

HasNetworkTopologyId returns a boolean if a field has been set.

### SetNetworkTopologyIdNil

`func (o *NodepoolCreateFields) SetNetworkTopologyIdNil(b bool)`

 SetNetworkTopologyIdNil sets the value for NetworkTopologyId to be an explicit nil

### UnsetNetworkTopologyId
`func (o *NodepoolCreateFields) UnsetNetworkTopologyId()`

UnsetNetworkTopologyId ensures that no value is present for NetworkTopologyId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


