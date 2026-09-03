# NodepoolSyncUpdateFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LabelKey** | Pointer to **NullableString** | Label key for associated nodes to the Node Pool (with value as in labelValue) | [optional] 
**LabelValue** | Pointer to **NullableString** | Label value for associated nodes to the Node Pool (with key as in labelKey) | [optional] 
**OverProvisioningRatio** | Pointer to **NullableInt32** |  | [optional] [default to 1]
**PlacementStrategy** | Pointer to [**NodepoolCreateResponseFieldsPlacementStrategy**](NodepoolCreateResponseFieldsPlacementStrategy.md) |  | [optional] 
**GpuNetworkAccelerationLabelKey** | Pointer to **string** | Label key by which to determine GPUNetworkAccelerationDetection nodes | [optional] 
**GpuNetworkAccelerationDetection** | Pointer to [**NullableNodepoolGPUNetworkAccelerationDetection**](NodepoolGPUNetworkAccelerationDetection.md) |  | [optional] [default to NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO]
**GpuResourceOptimization** | Pointer to [**NullableGPUResourceOptimization**](GPUResourceOptimization.md) |  | [optional] 
**SchedulingConfiguration** | Pointer to [**NullableSchedulingConfiguration**](SchedulingConfiguration.md) |  | [optional] 

## Methods

### NewNodepoolSyncUpdateFields

`func NewNodepoolSyncUpdateFields() *NodepoolSyncUpdateFields`

NewNodepoolSyncUpdateFields instantiates a new NodepoolSyncUpdateFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolSyncUpdateFieldsWithDefaults

`func NewNodepoolSyncUpdateFieldsWithDefaults() *NodepoolSyncUpdateFields`

NewNodepoolSyncUpdateFieldsWithDefaults instantiates a new NodepoolSyncUpdateFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabelKey

`func (o *NodepoolSyncUpdateFields) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NodepoolSyncUpdateFields) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NodepoolSyncUpdateFields) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *NodepoolSyncUpdateFields) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### SetLabelKeyNil

`func (o *NodepoolSyncUpdateFields) SetLabelKeyNil(b bool)`

 SetLabelKeyNil sets the value for LabelKey to be an explicit nil

### UnsetLabelKey
`func (o *NodepoolSyncUpdateFields) UnsetLabelKey()`

UnsetLabelKey ensures that no value is present for LabelKey, not even an explicit nil
### GetLabelValue

`func (o *NodepoolSyncUpdateFields) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NodepoolSyncUpdateFields) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NodepoolSyncUpdateFields) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *NodepoolSyncUpdateFields) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### SetLabelValueNil

`func (o *NodepoolSyncUpdateFields) SetLabelValueNil(b bool)`

 SetLabelValueNil sets the value for LabelValue to be an explicit nil

### UnsetLabelValue
`func (o *NodepoolSyncUpdateFields) UnsetLabelValue()`

UnsetLabelValue ensures that no value is present for LabelValue, not even an explicit nil
### GetOverProvisioningRatio

`func (o *NodepoolSyncUpdateFields) GetOverProvisioningRatio() int32`

GetOverProvisioningRatio returns the OverProvisioningRatio field if non-nil, zero value otherwise.

### GetOverProvisioningRatioOk

`func (o *NodepoolSyncUpdateFields) GetOverProvisioningRatioOk() (*int32, bool)`

GetOverProvisioningRatioOk returns a tuple with the OverProvisioningRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverProvisioningRatio

`func (o *NodepoolSyncUpdateFields) SetOverProvisioningRatio(v int32)`

SetOverProvisioningRatio sets OverProvisioningRatio field to given value.

### HasOverProvisioningRatio

`func (o *NodepoolSyncUpdateFields) HasOverProvisioningRatio() bool`

HasOverProvisioningRatio returns a boolean if a field has been set.

### SetOverProvisioningRatioNil

`func (o *NodepoolSyncUpdateFields) SetOverProvisioningRatioNil(b bool)`

 SetOverProvisioningRatioNil sets the value for OverProvisioningRatio to be an explicit nil

### UnsetOverProvisioningRatio
`func (o *NodepoolSyncUpdateFields) UnsetOverProvisioningRatio()`

UnsetOverProvisioningRatio ensures that no value is present for OverProvisioningRatio, not even an explicit nil
### GetPlacementStrategy

`func (o *NodepoolSyncUpdateFields) GetPlacementStrategy() NodepoolCreateResponseFieldsPlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *NodepoolSyncUpdateFields) GetPlacementStrategyOk() (*NodepoolCreateResponseFieldsPlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *NodepoolSyncUpdateFields) SetPlacementStrategy(v NodepoolCreateResponseFieldsPlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *NodepoolSyncUpdateFields) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.

### GetGpuNetworkAccelerationLabelKey

`func (o *NodepoolSyncUpdateFields) GetGpuNetworkAccelerationLabelKey() string`

GetGpuNetworkAccelerationLabelKey returns the GpuNetworkAccelerationLabelKey field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationLabelKeyOk

`func (o *NodepoolSyncUpdateFields) GetGpuNetworkAccelerationLabelKeyOk() (*string, bool)`

GetGpuNetworkAccelerationLabelKeyOk returns a tuple with the GpuNetworkAccelerationLabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationLabelKey

`func (o *NodepoolSyncUpdateFields) SetGpuNetworkAccelerationLabelKey(v string)`

SetGpuNetworkAccelerationLabelKey sets GpuNetworkAccelerationLabelKey field to given value.

### HasGpuNetworkAccelerationLabelKey

`func (o *NodepoolSyncUpdateFields) HasGpuNetworkAccelerationLabelKey() bool`

HasGpuNetworkAccelerationLabelKey returns a boolean if a field has been set.

### GetGpuNetworkAccelerationDetection

`func (o *NodepoolSyncUpdateFields) GetGpuNetworkAccelerationDetection() NodepoolGPUNetworkAccelerationDetection`

GetGpuNetworkAccelerationDetection returns the GpuNetworkAccelerationDetection field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationDetectionOk

`func (o *NodepoolSyncUpdateFields) GetGpuNetworkAccelerationDetectionOk() (*NodepoolGPUNetworkAccelerationDetection, bool)`

GetGpuNetworkAccelerationDetectionOk returns a tuple with the GpuNetworkAccelerationDetection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationDetection

`func (o *NodepoolSyncUpdateFields) SetGpuNetworkAccelerationDetection(v NodepoolGPUNetworkAccelerationDetection)`

SetGpuNetworkAccelerationDetection sets GpuNetworkAccelerationDetection field to given value.

### HasGpuNetworkAccelerationDetection

`func (o *NodepoolSyncUpdateFields) HasGpuNetworkAccelerationDetection() bool`

HasGpuNetworkAccelerationDetection returns a boolean if a field has been set.

### SetGpuNetworkAccelerationDetectionNil

`func (o *NodepoolSyncUpdateFields) SetGpuNetworkAccelerationDetectionNil(b bool)`

 SetGpuNetworkAccelerationDetectionNil sets the value for GpuNetworkAccelerationDetection to be an explicit nil

### UnsetGpuNetworkAccelerationDetection
`func (o *NodepoolSyncUpdateFields) UnsetGpuNetworkAccelerationDetection()`

UnsetGpuNetworkAccelerationDetection ensures that no value is present for GpuNetworkAccelerationDetection, not even an explicit nil
### GetGpuResourceOptimization

`func (o *NodepoolSyncUpdateFields) GetGpuResourceOptimization() GPUResourceOptimization`

GetGpuResourceOptimization returns the GpuResourceOptimization field if non-nil, zero value otherwise.

### GetGpuResourceOptimizationOk

`func (o *NodepoolSyncUpdateFields) GetGpuResourceOptimizationOk() (*GPUResourceOptimization, bool)`

GetGpuResourceOptimizationOk returns a tuple with the GpuResourceOptimization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuResourceOptimization

`func (o *NodepoolSyncUpdateFields) SetGpuResourceOptimization(v GPUResourceOptimization)`

SetGpuResourceOptimization sets GpuResourceOptimization field to given value.

### HasGpuResourceOptimization

`func (o *NodepoolSyncUpdateFields) HasGpuResourceOptimization() bool`

HasGpuResourceOptimization returns a boolean if a field has been set.

### SetGpuResourceOptimizationNil

`func (o *NodepoolSyncUpdateFields) SetGpuResourceOptimizationNil(b bool)`

 SetGpuResourceOptimizationNil sets the value for GpuResourceOptimization to be an explicit nil

### UnsetGpuResourceOptimization
`func (o *NodepoolSyncUpdateFields) UnsetGpuResourceOptimization()`

UnsetGpuResourceOptimization ensures that no value is present for GpuResourceOptimization, not even an explicit nil
### GetSchedulingConfiguration

`func (o *NodepoolSyncUpdateFields) GetSchedulingConfiguration() SchedulingConfiguration`

GetSchedulingConfiguration returns the SchedulingConfiguration field if non-nil, zero value otherwise.

### GetSchedulingConfigurationOk

`func (o *NodepoolSyncUpdateFields) GetSchedulingConfigurationOk() (*SchedulingConfiguration, bool)`

GetSchedulingConfigurationOk returns a tuple with the SchedulingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingConfiguration

`func (o *NodepoolSyncUpdateFields) SetSchedulingConfiguration(v SchedulingConfiguration)`

SetSchedulingConfiguration sets SchedulingConfiguration field to given value.

### HasSchedulingConfiguration

`func (o *NodepoolSyncUpdateFields) HasSchedulingConfiguration() bool`

HasSchedulingConfiguration returns a boolean if a field has been set.

### SetSchedulingConfigurationNil

`func (o *NodepoolSyncUpdateFields) SetSchedulingConfigurationNil(b bool)`

 SetSchedulingConfigurationNil sets the value for SchedulingConfiguration to be an explicit nil

### UnsetSchedulingConfiguration
`func (o *NodepoolSyncUpdateFields) UnsetSchedulingConfiguration()`

UnsetSchedulingConfiguration ensures that no value is present for SchedulingConfiguration, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


