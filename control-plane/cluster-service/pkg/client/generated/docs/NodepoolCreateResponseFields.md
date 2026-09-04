# NodepoolCreateResponseFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**LabelKey** | **string** | Label key for associated nodes to the Node Pool (with value as in labelValue) | 
**LabelValue** | **string** | Label value for associated nodes to the Node Pool (with key as in labelKey) | 
**ClusterId** | **string** |  | 
**SchedulingCapabilities** | Pointer to [**SchedulingCapabilities**](SchedulingCapabilities.md) |  | [optional] 
**GpuNetworkAccelerationLabelKey** | Pointer to **string** | Label key by which to determine GPUNetworkAccelerationDetection nodes | [optional] 
**GpuNetworkAccelerationDetection** | Pointer to [**NullableNodepoolGPUNetworkAccelerationDetection**](NodepoolGPUNetworkAccelerationDetection.md) |  | [optional] [default to NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO]
**NetworkTopologyName** | Pointer to **NullableString** | Name of the network topology | [optional] 
**NetworkTopologyId** | Pointer to **NullableString** | The unique identifier for the network topology | [optional] 
**SchedulingConfiguration** | Pointer to [**NullableSchedulingConfiguration**](SchedulingConfiguration.md) |  | [optional] 

## Methods

### NewNodepoolCreateResponseFields

`func NewNodepoolCreateResponseFields(name string, labelKey string, labelValue string, clusterId string, ) *NodepoolCreateResponseFields`

NewNodepoolCreateResponseFields instantiates a new NodepoolCreateResponseFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolCreateResponseFieldsWithDefaults

`func NewNodepoolCreateResponseFieldsWithDefaults() *NodepoolCreateResponseFields`

NewNodepoolCreateResponseFieldsWithDefaults instantiates a new NodepoolCreateResponseFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodepoolCreateResponseFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodepoolCreateResponseFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodepoolCreateResponseFields) SetName(v string)`

SetName sets Name field to given value.


### GetLabelKey

`func (o *NodepoolCreateResponseFields) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NodepoolCreateResponseFields) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NodepoolCreateResponseFields) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.


### GetLabelValue

`func (o *NodepoolCreateResponseFields) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NodepoolCreateResponseFields) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NodepoolCreateResponseFields) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.


### GetClusterId

`func (o *NodepoolCreateResponseFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *NodepoolCreateResponseFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *NodepoolCreateResponseFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetSchedulingCapabilities

`func (o *NodepoolCreateResponseFields) GetSchedulingCapabilities() SchedulingCapabilities`

GetSchedulingCapabilities returns the SchedulingCapabilities field if non-nil, zero value otherwise.

### GetSchedulingCapabilitiesOk

`func (o *NodepoolCreateResponseFields) GetSchedulingCapabilitiesOk() (*SchedulingCapabilities, bool)`

GetSchedulingCapabilitiesOk returns a tuple with the SchedulingCapabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingCapabilities

`func (o *NodepoolCreateResponseFields) SetSchedulingCapabilities(v SchedulingCapabilities)`

SetSchedulingCapabilities sets SchedulingCapabilities field to given value.

### HasSchedulingCapabilities

`func (o *NodepoolCreateResponseFields) HasSchedulingCapabilities() bool`

HasSchedulingCapabilities returns a boolean if a field has been set.

### GetGpuNetworkAccelerationLabelKey

`func (o *NodepoolCreateResponseFields) GetGpuNetworkAccelerationLabelKey() string`

GetGpuNetworkAccelerationLabelKey returns the GpuNetworkAccelerationLabelKey field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationLabelKeyOk

`func (o *NodepoolCreateResponseFields) GetGpuNetworkAccelerationLabelKeyOk() (*string, bool)`

GetGpuNetworkAccelerationLabelKeyOk returns a tuple with the GpuNetworkAccelerationLabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationLabelKey

`func (o *NodepoolCreateResponseFields) SetGpuNetworkAccelerationLabelKey(v string)`

SetGpuNetworkAccelerationLabelKey sets GpuNetworkAccelerationLabelKey field to given value.

### HasGpuNetworkAccelerationLabelKey

`func (o *NodepoolCreateResponseFields) HasGpuNetworkAccelerationLabelKey() bool`

HasGpuNetworkAccelerationLabelKey returns a boolean if a field has been set.

### GetGpuNetworkAccelerationDetection

`func (o *NodepoolCreateResponseFields) GetGpuNetworkAccelerationDetection() NodepoolGPUNetworkAccelerationDetection`

GetGpuNetworkAccelerationDetection returns the GpuNetworkAccelerationDetection field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationDetectionOk

`func (o *NodepoolCreateResponseFields) GetGpuNetworkAccelerationDetectionOk() (*NodepoolGPUNetworkAccelerationDetection, bool)`

GetGpuNetworkAccelerationDetectionOk returns a tuple with the GpuNetworkAccelerationDetection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationDetection

`func (o *NodepoolCreateResponseFields) SetGpuNetworkAccelerationDetection(v NodepoolGPUNetworkAccelerationDetection)`

SetGpuNetworkAccelerationDetection sets GpuNetworkAccelerationDetection field to given value.

### HasGpuNetworkAccelerationDetection

`func (o *NodepoolCreateResponseFields) HasGpuNetworkAccelerationDetection() bool`

HasGpuNetworkAccelerationDetection returns a boolean if a field has been set.

### SetGpuNetworkAccelerationDetectionNil

`func (o *NodepoolCreateResponseFields) SetGpuNetworkAccelerationDetectionNil(b bool)`

 SetGpuNetworkAccelerationDetectionNil sets the value for GpuNetworkAccelerationDetection to be an explicit nil

### UnsetGpuNetworkAccelerationDetection
`func (o *NodepoolCreateResponseFields) UnsetGpuNetworkAccelerationDetection()`

UnsetGpuNetworkAccelerationDetection ensures that no value is present for GpuNetworkAccelerationDetection, not even an explicit nil
### GetNetworkTopologyName

`func (o *NodepoolCreateResponseFields) GetNetworkTopologyName() string`

GetNetworkTopologyName returns the NetworkTopologyName field if non-nil, zero value otherwise.

### GetNetworkTopologyNameOk

`func (o *NodepoolCreateResponseFields) GetNetworkTopologyNameOk() (*string, bool)`

GetNetworkTopologyNameOk returns a tuple with the NetworkTopologyName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyName

`func (o *NodepoolCreateResponseFields) SetNetworkTopologyName(v string)`

SetNetworkTopologyName sets NetworkTopologyName field to given value.

### HasNetworkTopologyName

`func (o *NodepoolCreateResponseFields) HasNetworkTopologyName() bool`

HasNetworkTopologyName returns a boolean if a field has been set.

### SetNetworkTopologyNameNil

`func (o *NodepoolCreateResponseFields) SetNetworkTopologyNameNil(b bool)`

 SetNetworkTopologyNameNil sets the value for NetworkTopologyName to be an explicit nil

### UnsetNetworkTopologyName
`func (o *NodepoolCreateResponseFields) UnsetNetworkTopologyName()`

UnsetNetworkTopologyName ensures that no value is present for NetworkTopologyName, not even an explicit nil
### GetNetworkTopologyId

`func (o *NodepoolCreateResponseFields) GetNetworkTopologyId() string`

GetNetworkTopologyId returns the NetworkTopologyId field if non-nil, zero value otherwise.

### GetNetworkTopologyIdOk

`func (o *NodepoolCreateResponseFields) GetNetworkTopologyIdOk() (*string, bool)`

GetNetworkTopologyIdOk returns a tuple with the NetworkTopologyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyId

`func (o *NodepoolCreateResponseFields) SetNetworkTopologyId(v string)`

SetNetworkTopologyId sets NetworkTopologyId field to given value.

### HasNetworkTopologyId

`func (o *NodepoolCreateResponseFields) HasNetworkTopologyId() bool`

HasNetworkTopologyId returns a boolean if a field has been set.

### SetNetworkTopologyIdNil

`func (o *NodepoolCreateResponseFields) SetNetworkTopologyIdNil(b bool)`

 SetNetworkTopologyIdNil sets the value for NetworkTopologyId to be an explicit nil

### UnsetNetworkTopologyId
`func (o *NodepoolCreateResponseFields) UnsetNetworkTopologyId()`

UnsetNetworkTopologyId ensures that no value is present for NetworkTopologyId, not even an explicit nil
### GetSchedulingConfiguration

`func (o *NodepoolCreateResponseFields) GetSchedulingConfiguration() SchedulingConfiguration`

GetSchedulingConfiguration returns the SchedulingConfiguration field if non-nil, zero value otherwise.

### GetSchedulingConfigurationOk

`func (o *NodepoolCreateResponseFields) GetSchedulingConfigurationOk() (*SchedulingConfiguration, bool)`

GetSchedulingConfigurationOk returns a tuple with the SchedulingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingConfiguration

`func (o *NodepoolCreateResponseFields) SetSchedulingConfiguration(v SchedulingConfiguration)`

SetSchedulingConfiguration sets SchedulingConfiguration field to given value.

### HasSchedulingConfiguration

`func (o *NodepoolCreateResponseFields) HasSchedulingConfiguration() bool`

HasSchedulingConfiguration returns a boolean if a field has been set.

### SetSchedulingConfigurationNil

`func (o *NodepoolCreateResponseFields) SetSchedulingConfigurationNil(b bool)`

 SetSchedulingConfigurationNil sets the value for SchedulingConfiguration to be an explicit nil

### UnsetSchedulingConfiguration
`func (o *NodepoolCreateResponseFields) UnsetSchedulingConfiguration()`

UnsetSchedulingConfiguration ensures that no value is present for SchedulingConfiguration, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


