# Nodepool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**LabelKey** | **string** | Label key for associated nodes to the Node Pool (with value as in labelValue) | 
**LabelValue** | **string** | Label value for associated nodes to the Node Pool (with key as in labelKey) | 
**ClusterId** | **string** |  | 
**SchedulingCapabilities** | Pointer to [**SchedulingCapabilities**](SchedulingCapabilities.md) |  | [optional] 
**OverProvisioningRatio** | Pointer to **NullableInt32** |  | [optional] [default to 1]
**PlacementStrategy** | Pointer to [**NodepoolCreateResponseFieldsPlacementStrategy**](NodepoolCreateResponseFieldsPlacementStrategy.md) |  | [optional] 
**GpuNetworkAccelerationLabelKey** | Pointer to **string** | Label key by which to determine GPUNetworkAccelerationDetection nodes | [optional] 
**GpuNetworkAccelerationDetection** | Pointer to [**NullableNodepoolGPUNetworkAccelerationDetection**](NodepoolGPUNetworkAccelerationDetection.md) |  | [optional] [default to NODEPOOLGPUNETWORKACCELERATIONDETECTION_AUTO]
**NetworkTopologyName** | Pointer to **NullableString** | Name of the network topology | [optional] 
**NetworkTopologyId** | Pointer to **NullableString** | The unique identifier for the network topology | [optional] 
**GpuResourceOptimization** | Pointer to [**NullableGPUResourceOptimization**](GPUResourceOptimization.md) |  | [optional] 
**SchedulingConfiguration** | Pointer to [**NullableSchedulingConfiguration**](SchedulingConfiguration.md) |  | [optional] 
**Phase** | Pointer to [**NodepoolPhase**](NodepoolPhase.md) |  | [optional] 
**PhaseMessage** | Pointer to **string** | Message for status of Node Pool | [optional] 
**Status** | Pointer to [**NodepoolSyncFieldsStatus**](NodepoolSyncFieldsStatus.md) |  | [optional] 
**Id** | **string** | The NodePool id | [readonly] 
**TenantId** | **int32** | The id of the tenant. | 
**ClusterName** | **string** |  | 
**CreatedBy** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedBy** | **string** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedBy** | Pointer to **string** |  | [optional] 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**IsDefault** | **bool** | Is the Node Pool is the cluster default | 

## Methods

### NewNodepool

`func NewNodepool(name string, labelKey string, labelValue string, clusterId string, id string, tenantId int32, clusterName string, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time, isDefault bool, ) *Nodepool`

NewNodepool instantiates a new Nodepool object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolWithDefaults

`func NewNodepoolWithDefaults() *Nodepool`

NewNodepoolWithDefaults instantiates a new Nodepool object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Nodepool) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Nodepool) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Nodepool) SetName(v string)`

SetName sets Name field to given value.


### GetLabelKey

`func (o *Nodepool) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *Nodepool) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *Nodepool) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.


### GetLabelValue

`func (o *Nodepool) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *Nodepool) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *Nodepool) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.


### GetClusterId

`func (o *Nodepool) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Nodepool) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Nodepool) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetSchedulingCapabilities

`func (o *Nodepool) GetSchedulingCapabilities() SchedulingCapabilities`

GetSchedulingCapabilities returns the SchedulingCapabilities field if non-nil, zero value otherwise.

### GetSchedulingCapabilitiesOk

`func (o *Nodepool) GetSchedulingCapabilitiesOk() (*SchedulingCapabilities, bool)`

GetSchedulingCapabilitiesOk returns a tuple with the SchedulingCapabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingCapabilities

`func (o *Nodepool) SetSchedulingCapabilities(v SchedulingCapabilities)`

SetSchedulingCapabilities sets SchedulingCapabilities field to given value.

### HasSchedulingCapabilities

`func (o *Nodepool) HasSchedulingCapabilities() bool`

HasSchedulingCapabilities returns a boolean if a field has been set.

### GetOverProvisioningRatio

`func (o *Nodepool) GetOverProvisioningRatio() int32`

GetOverProvisioningRatio returns the OverProvisioningRatio field if non-nil, zero value otherwise.

### GetOverProvisioningRatioOk

`func (o *Nodepool) GetOverProvisioningRatioOk() (*int32, bool)`

GetOverProvisioningRatioOk returns a tuple with the OverProvisioningRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverProvisioningRatio

`func (o *Nodepool) SetOverProvisioningRatio(v int32)`

SetOverProvisioningRatio sets OverProvisioningRatio field to given value.

### HasOverProvisioningRatio

`func (o *Nodepool) HasOverProvisioningRatio() bool`

HasOverProvisioningRatio returns a boolean if a field has been set.

### SetOverProvisioningRatioNil

`func (o *Nodepool) SetOverProvisioningRatioNil(b bool)`

 SetOverProvisioningRatioNil sets the value for OverProvisioningRatio to be an explicit nil

### UnsetOverProvisioningRatio
`func (o *Nodepool) UnsetOverProvisioningRatio()`

UnsetOverProvisioningRatio ensures that no value is present for OverProvisioningRatio, not even an explicit nil
### GetPlacementStrategy

`func (o *Nodepool) GetPlacementStrategy() NodepoolCreateResponseFieldsPlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *Nodepool) GetPlacementStrategyOk() (*NodepoolCreateResponseFieldsPlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *Nodepool) SetPlacementStrategy(v NodepoolCreateResponseFieldsPlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *Nodepool) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.

### GetGpuNetworkAccelerationLabelKey

`func (o *Nodepool) GetGpuNetworkAccelerationLabelKey() string`

GetGpuNetworkAccelerationLabelKey returns the GpuNetworkAccelerationLabelKey field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationLabelKeyOk

`func (o *Nodepool) GetGpuNetworkAccelerationLabelKeyOk() (*string, bool)`

GetGpuNetworkAccelerationLabelKeyOk returns a tuple with the GpuNetworkAccelerationLabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationLabelKey

`func (o *Nodepool) SetGpuNetworkAccelerationLabelKey(v string)`

SetGpuNetworkAccelerationLabelKey sets GpuNetworkAccelerationLabelKey field to given value.

### HasGpuNetworkAccelerationLabelKey

`func (o *Nodepool) HasGpuNetworkAccelerationLabelKey() bool`

HasGpuNetworkAccelerationLabelKey returns a boolean if a field has been set.

### GetGpuNetworkAccelerationDetection

`func (o *Nodepool) GetGpuNetworkAccelerationDetection() NodepoolGPUNetworkAccelerationDetection`

GetGpuNetworkAccelerationDetection returns the GpuNetworkAccelerationDetection field if non-nil, zero value otherwise.

### GetGpuNetworkAccelerationDetectionOk

`func (o *Nodepool) GetGpuNetworkAccelerationDetectionOk() (*NodepoolGPUNetworkAccelerationDetection, bool)`

GetGpuNetworkAccelerationDetectionOk returns a tuple with the GpuNetworkAccelerationDetection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuNetworkAccelerationDetection

`func (o *Nodepool) SetGpuNetworkAccelerationDetection(v NodepoolGPUNetworkAccelerationDetection)`

SetGpuNetworkAccelerationDetection sets GpuNetworkAccelerationDetection field to given value.

### HasGpuNetworkAccelerationDetection

`func (o *Nodepool) HasGpuNetworkAccelerationDetection() bool`

HasGpuNetworkAccelerationDetection returns a boolean if a field has been set.

### SetGpuNetworkAccelerationDetectionNil

`func (o *Nodepool) SetGpuNetworkAccelerationDetectionNil(b bool)`

 SetGpuNetworkAccelerationDetectionNil sets the value for GpuNetworkAccelerationDetection to be an explicit nil

### UnsetGpuNetworkAccelerationDetection
`func (o *Nodepool) UnsetGpuNetworkAccelerationDetection()`

UnsetGpuNetworkAccelerationDetection ensures that no value is present for GpuNetworkAccelerationDetection, not even an explicit nil
### GetNetworkTopologyName

`func (o *Nodepool) GetNetworkTopologyName() string`

GetNetworkTopologyName returns the NetworkTopologyName field if non-nil, zero value otherwise.

### GetNetworkTopologyNameOk

`func (o *Nodepool) GetNetworkTopologyNameOk() (*string, bool)`

GetNetworkTopologyNameOk returns a tuple with the NetworkTopologyName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyName

`func (o *Nodepool) SetNetworkTopologyName(v string)`

SetNetworkTopologyName sets NetworkTopologyName field to given value.

### HasNetworkTopologyName

`func (o *Nodepool) HasNetworkTopologyName() bool`

HasNetworkTopologyName returns a boolean if a field has been set.

### SetNetworkTopologyNameNil

`func (o *Nodepool) SetNetworkTopologyNameNil(b bool)`

 SetNetworkTopologyNameNil sets the value for NetworkTopologyName to be an explicit nil

### UnsetNetworkTopologyName
`func (o *Nodepool) UnsetNetworkTopologyName()`

UnsetNetworkTopologyName ensures that no value is present for NetworkTopologyName, not even an explicit nil
### GetNetworkTopologyId

`func (o *Nodepool) GetNetworkTopologyId() string`

GetNetworkTopologyId returns the NetworkTopologyId field if non-nil, zero value otherwise.

### GetNetworkTopologyIdOk

`func (o *Nodepool) GetNetworkTopologyIdOk() (*string, bool)`

GetNetworkTopologyIdOk returns a tuple with the NetworkTopologyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyId

`func (o *Nodepool) SetNetworkTopologyId(v string)`

SetNetworkTopologyId sets NetworkTopologyId field to given value.

### HasNetworkTopologyId

`func (o *Nodepool) HasNetworkTopologyId() bool`

HasNetworkTopologyId returns a boolean if a field has been set.

### SetNetworkTopologyIdNil

`func (o *Nodepool) SetNetworkTopologyIdNil(b bool)`

 SetNetworkTopologyIdNil sets the value for NetworkTopologyId to be an explicit nil

### UnsetNetworkTopologyId
`func (o *Nodepool) UnsetNetworkTopologyId()`

UnsetNetworkTopologyId ensures that no value is present for NetworkTopologyId, not even an explicit nil
### GetGpuResourceOptimization

`func (o *Nodepool) GetGpuResourceOptimization() GPUResourceOptimization`

GetGpuResourceOptimization returns the GpuResourceOptimization field if non-nil, zero value otherwise.

### GetGpuResourceOptimizationOk

`func (o *Nodepool) GetGpuResourceOptimizationOk() (*GPUResourceOptimization, bool)`

GetGpuResourceOptimizationOk returns a tuple with the GpuResourceOptimization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuResourceOptimization

`func (o *Nodepool) SetGpuResourceOptimization(v GPUResourceOptimization)`

SetGpuResourceOptimization sets GpuResourceOptimization field to given value.

### HasGpuResourceOptimization

`func (o *Nodepool) HasGpuResourceOptimization() bool`

HasGpuResourceOptimization returns a boolean if a field has been set.

### SetGpuResourceOptimizationNil

`func (o *Nodepool) SetGpuResourceOptimizationNil(b bool)`

 SetGpuResourceOptimizationNil sets the value for GpuResourceOptimization to be an explicit nil

### UnsetGpuResourceOptimization
`func (o *Nodepool) UnsetGpuResourceOptimization()`

UnsetGpuResourceOptimization ensures that no value is present for GpuResourceOptimization, not even an explicit nil
### GetSchedulingConfiguration

`func (o *Nodepool) GetSchedulingConfiguration() SchedulingConfiguration`

GetSchedulingConfiguration returns the SchedulingConfiguration field if non-nil, zero value otherwise.

### GetSchedulingConfigurationOk

`func (o *Nodepool) GetSchedulingConfigurationOk() (*SchedulingConfiguration, bool)`

GetSchedulingConfigurationOk returns a tuple with the SchedulingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingConfiguration

`func (o *Nodepool) SetSchedulingConfiguration(v SchedulingConfiguration)`

SetSchedulingConfiguration sets SchedulingConfiguration field to given value.

### HasSchedulingConfiguration

`func (o *Nodepool) HasSchedulingConfiguration() bool`

HasSchedulingConfiguration returns a boolean if a field has been set.

### SetSchedulingConfigurationNil

`func (o *Nodepool) SetSchedulingConfigurationNil(b bool)`

 SetSchedulingConfigurationNil sets the value for SchedulingConfiguration to be an explicit nil

### UnsetSchedulingConfiguration
`func (o *Nodepool) UnsetSchedulingConfiguration()`

UnsetSchedulingConfiguration ensures that no value is present for SchedulingConfiguration, not even an explicit nil
### GetPhase

`func (o *Nodepool) GetPhase() NodepoolPhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *Nodepool) GetPhaseOk() (*NodepoolPhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *Nodepool) SetPhase(v NodepoolPhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *Nodepool) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetPhaseMessage

`func (o *Nodepool) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *Nodepool) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *Nodepool) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *Nodepool) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetStatus

`func (o *Nodepool) GetStatus() NodepoolSyncFieldsStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Nodepool) GetStatusOk() (*NodepoolSyncFieldsStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Nodepool) SetStatus(v NodepoolSyncFieldsStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Nodepool) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetId

`func (o *Nodepool) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Nodepool) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Nodepool) SetId(v string)`

SetId sets Id field to given value.


### GetTenantId

`func (o *Nodepool) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Nodepool) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Nodepool) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetClusterName

`func (o *Nodepool) GetClusterName() string`

GetClusterName returns the ClusterName field if non-nil, zero value otherwise.

### GetClusterNameOk

`func (o *Nodepool) GetClusterNameOk() (*string, bool)`

GetClusterNameOk returns a tuple with the ClusterName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterName

`func (o *Nodepool) SetClusterName(v string)`

SetClusterName sets ClusterName field to given value.


### GetCreatedBy

`func (o *Nodepool) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Nodepool) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Nodepool) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *Nodepool) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Nodepool) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Nodepool) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedBy

`func (o *Nodepool) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *Nodepool) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *Nodepool) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetUpdatedAt

`func (o *Nodepool) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Nodepool) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Nodepool) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedBy

`func (o *Nodepool) GetDeletedBy() string`

GetDeletedBy returns the DeletedBy field if non-nil, zero value otherwise.

### GetDeletedByOk

`func (o *Nodepool) GetDeletedByOk() (*string, bool)`

GetDeletedByOk returns a tuple with the DeletedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedBy

`func (o *Nodepool) SetDeletedBy(v string)`

SetDeletedBy sets DeletedBy field to given value.

### HasDeletedBy

`func (o *Nodepool) HasDeletedBy() bool`

HasDeletedBy returns a boolean if a field has been set.

### GetDeletedAt

`func (o *Nodepool) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *Nodepool) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *Nodepool) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *Nodepool) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *Nodepool) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *Nodepool) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetIsDefault

`func (o *Nodepool) GetIsDefault() bool`

GetIsDefault returns the IsDefault field if non-nil, zero value otherwise.

### GetIsDefaultOk

`func (o *Nodepool) GetIsDefaultOk() (*bool, bool)`

GetIsDefaultOk returns a tuple with the IsDefault field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDefault

`func (o *Nodepool) SetIsDefault(v bool)`

SetIsDefault sets IsDefault field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


