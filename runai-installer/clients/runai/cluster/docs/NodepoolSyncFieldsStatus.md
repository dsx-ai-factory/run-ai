# NodepoolSyncFieldsStatus

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
**Conditions** | Pointer to [**[]NodepoolConditionDetails**](NodepoolConditionDetails.md) |  | [optional] 
**GpuNetworkAccelerationDetected** | Pointer to **bool** | Signals to all relevant parties about GPUNetworkAcceleration being detected in some nodepool&#39;s nodes | [optional] 
**Nodes** | Pointer to **[]string** |  | [optional] 

## Methods

### NewNodepoolSyncFieldsStatus

`func NewNodepoolSyncFieldsStatus() *NodepoolSyncFieldsStatus`

NewNodepoolSyncFieldsStatus instantiates a new NodepoolSyncFieldsStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolSyncFieldsStatusWithDefaults

`func NewNodepoolSyncFieldsStatusWithDefaults() *NodepoolSyncFieldsStatus`

NewNodepoolSyncFieldsStatusWithDefaults instantiates a new NodepoolSyncFieldsStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabelKey

`func (o *NodepoolSyncFieldsStatus) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NodepoolSyncFieldsStatus) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NodepoolSyncFieldsStatus) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *NodepoolSyncFieldsStatus) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### SetLabelKeyNil

`func (o *NodepoolSyncFieldsStatus) SetLabelKeyNil(b bool)`

 SetLabelKeyNil sets the value for LabelKey to be an explicit nil

### UnsetLabelKey
`func (o *NodepoolSyncFieldsStatus) UnsetLabelKey()`

UnsetLabelKey ensures that no value is present for LabelKey, not even an explicit nil
### GetLabelValue

`func (o *NodepoolSyncFieldsStatus) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NodepoolSyncFieldsStatus) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NodepoolSyncFieldsStatus) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *NodepoolSyncFieldsStatus) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### SetLabelValueNil

`func (o *NodepoolSyncFieldsStatus) SetLabelValueNil(b bool)`

 SetLabelValueNil sets the value for LabelValue to be an explicit nil

### UnsetLabelValue
`func (o *NodepoolSyncFieldsStatus) UnsetLabelValue()`

UnsetLabelValue ensures that no value is present for LabelValue, not even an explicit nil
### GetOverProvisioningRatio

`func (o *NodepoolSyncFieldsStatus) GetOverProvisioningRatio() int32`

GetOverProvisioningRatio returns the OverProvisioningRatio field if non-nil, zero value otherwise.

### GetOverProvisioningRatioOk

`func (o *NodepoolSyncFieldsStatus) GetOverProvisioningRatioOk() (*int32, bool)`

GetOverProvisioningRatioOk returns a tuple with the OverProvisioningRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverProvisioningRatio

`func (o *NodepoolSyncFieldsStatus) SetOverProvisioningRatio(v int32)`

SetOverProvisioningRatio sets OverProvisioningRatio field to given value.

### HasOverProvisioningRatio

`func (o *NodepoolSyncFieldsStatus) HasOverProvisioningRatio() bool`

HasOverProvisioningRatio returns a boolean if a field has been set.

### SetOverProvisioningRatioNil

`func (o *NodepoolSyncFieldsStatus) SetOverProvisioningRatioNil(b bool)`

 SetOverProvisioningRatioNil sets the value for OverProvisioningRatio to be an explicit nil

### UnsetOverProvisioningRatio
`func (o *NodepoolSyncFieldsStatus) UnsetOverProvisioningRatio()`

UnsetOverProvisioningRatio ensures that no value is present for OverProvisioningRatio, not even an explicit nil
### GetPlacementStrategy

`func (o *NodepoolSyncFieldsStatus) GetPlacementStrategy() NodepoolCreateResponseFieldsPlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *NodepoolSyncFieldsStatus) GetPlacementStrategyOk() (*NodepoolCreateResponseFieldsPlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *NodepoolSyncFieldsStatus) SetPlacementStrategy(v NodepoolCreateResponseFieldsPlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *NodepoolSyncFieldsStatus) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.

### GetGpuNetworkAccelerationLabelKey

`func (o *NodepoolSyncFieldsStatus) GetGpuNetworkAccelerationLabelKey() string`

GetGpuNetworkAccelerationLabelKey returns the GpuNetworkAccelerationLabelKey field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationLabelKeyOk

`func (o *NodepoolSyncFieldsStatus) GetGpuNetworkAccelerationLabelKeyOk() (*string, bool)`

GetGpuNetworkAccelerationLabelKeyOk returns a tuple with the GpuNetworkAccelerationLabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationLabelKey

`func (o *NodepoolSyncFieldsStatus) SetGpuNetworkAccelerationLabelKey(v string)`

SetGpuNetworkAccelerationLabelKey sets GpuNetworkAccelerationLabelKey field to given value.

### HasGpuNetworkAccelerationLabelKey

`func (o *NodepoolSyncFieldsStatus) HasGpuNetworkAccelerationLabelKey() bool`

HasGpuNetworkAccelerationLabelKey returns a boolean if a field has been set.

### GetGpuNetworkAccelerationDetection

`func (o *NodepoolSyncFieldsStatus) GetGpuNetworkAccelerationDetection() NodepoolGPUNetworkAccelerationDetection`

GetGpuNetworkAccelerationDetection returns the GpuNetworkAccelerationDetection field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationDetectionOk

`func (o *NodepoolSyncFieldsStatus) GetGpuNetworkAccelerationDetectionOk() (*NodepoolGPUNetworkAccelerationDetection, bool)`

GetGpuNetworkAccelerationDetectionOk returns a tuple with the GpuNetworkAccelerationDetection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationDetection

`func (o *NodepoolSyncFieldsStatus) SetGpuNetworkAccelerationDetection(v NodepoolGPUNetworkAccelerationDetection)`

SetGpuNetworkAccelerationDetection sets GpuNetworkAccelerationDetection field to given value.

### HasGpuNetworkAccelerationDetection

`func (o *NodepoolSyncFieldsStatus) HasGpuNetworkAccelerationDetection() bool`

HasGpuNetworkAccelerationDetection returns a boolean if a field has been set.

### SetGpuNetworkAccelerationDetectionNil

`func (o *NodepoolSyncFieldsStatus) SetGpuNetworkAccelerationDetectionNil(b bool)`

 SetGpuNetworkAccelerationDetectionNil sets the value for GpuNetworkAccelerationDetection to be an explicit nil

### UnsetGpuNetworkAccelerationDetection
`func (o *NodepoolSyncFieldsStatus) UnsetGpuNetworkAccelerationDetection()`

UnsetGpuNetworkAccelerationDetection ensures that no value is present for GpuNetworkAccelerationDetection, not even an explicit nil
### GetGpuResourceOptimization

`func (o *NodepoolSyncFieldsStatus) GetGpuResourceOptimization() GPUResourceOptimization`

GetGpuResourceOptimization returns the GpuResourceOptimization field if non-nil, zero value otherwise.

### GetGpuResourceOptimizationOk

`func (o *NodepoolSyncFieldsStatus) GetGpuResourceOptimizationOk() (*GPUResourceOptimization, bool)`

GetGpuResourceOptimizationOk returns a tuple with the GpuResourceOptimization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuResourceOptimization

`func (o *NodepoolSyncFieldsStatus) SetGpuResourceOptimization(v GPUResourceOptimization)`

SetGpuResourceOptimization sets GpuResourceOptimization field to given value.

### HasGpuResourceOptimization

`func (o *NodepoolSyncFieldsStatus) HasGpuResourceOptimization() bool`

HasGpuResourceOptimization returns a boolean if a field has been set.

### SetGpuResourceOptimizationNil

`func (o *NodepoolSyncFieldsStatus) SetGpuResourceOptimizationNil(b bool)`

 SetGpuResourceOptimizationNil sets the value for GpuResourceOptimization to be an explicit nil

### UnsetGpuResourceOptimization
`func (o *NodepoolSyncFieldsStatus) UnsetGpuResourceOptimization()`

UnsetGpuResourceOptimization ensures that no value is present for GpuResourceOptimization, not even an explicit nil
### GetSchedulingConfiguration

`func (o *NodepoolSyncFieldsStatus) GetSchedulingConfiguration() SchedulingConfiguration`

GetSchedulingConfiguration returns the SchedulingConfiguration field if non-nil, zero value otherwise.

### GetSchedulingConfigurationOk

`func (o *NodepoolSyncFieldsStatus) GetSchedulingConfigurationOk() (*SchedulingConfiguration, bool)`

GetSchedulingConfigurationOk returns a tuple with the SchedulingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingConfiguration

`func (o *NodepoolSyncFieldsStatus) SetSchedulingConfiguration(v SchedulingConfiguration)`

SetSchedulingConfiguration sets SchedulingConfiguration field to given value.

### HasSchedulingConfiguration

`func (o *NodepoolSyncFieldsStatus) HasSchedulingConfiguration() bool`

HasSchedulingConfiguration returns a boolean if a field has been set.

### SetSchedulingConfigurationNil

`func (o *NodepoolSyncFieldsStatus) SetSchedulingConfigurationNil(b bool)`

 SetSchedulingConfigurationNil sets the value for SchedulingConfiguration to be an explicit nil

### UnsetSchedulingConfiguration
`func (o *NodepoolSyncFieldsStatus) UnsetSchedulingConfiguration()`

UnsetSchedulingConfiguration ensures that no value is present for SchedulingConfiguration, not even an explicit nil
### GetConditions

`func (o *NodepoolSyncFieldsStatus) GetConditions() []NodepoolConditionDetails`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *NodepoolSyncFieldsStatus) GetConditionsOk() (*[]NodepoolConditionDetails, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *NodepoolSyncFieldsStatus) SetConditions(v []NodepoolConditionDetails)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *NodepoolSyncFieldsStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetGpuNetworkAccelerationDetected

`func (o *NodepoolSyncFieldsStatus) GetGpuNetworkAccelerationDetected() bool`

GetGpuNetworkAccelerationDetected returns the GpuNetworkAccelerationDetected field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationDetectedOk

`func (o *NodepoolSyncFieldsStatus) GetGpuNetworkAccelerationDetectedOk() (*bool, bool)`

GetGpuNetworkAccelerationDetectedOk returns a tuple with the GpuNetworkAccelerationDetected field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationDetected

`func (o *NodepoolSyncFieldsStatus) SetGpuNetworkAccelerationDetected(v bool)`

SetGpuNetworkAccelerationDetected sets GpuNetworkAccelerationDetected field to given value.

### HasGpuNetworkAccelerationDetected

`func (o *NodepoolSyncFieldsStatus) HasGpuNetworkAccelerationDetected() bool`

HasGpuNetworkAccelerationDetected returns a boolean if a field has been set.

### GetNodes

`func (o *NodepoolSyncFieldsStatus) GetNodes() []string`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *NodepoolSyncFieldsStatus) GetNodesOk() (*[]string, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *NodepoolSyncFieldsStatus) SetNodes(v []string)`

SetNodes sets Nodes field to given value.

### HasNodes

`func (o *NodepoolSyncFieldsStatus) HasNodes() bool`

HasNodes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


