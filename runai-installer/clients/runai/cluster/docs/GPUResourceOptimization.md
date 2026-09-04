# GPUResourceOptimization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SwapEnabled** | Pointer to **NullableBool** | Indicates if swap is enabled on the node pool | [optional] [default to false]
**CpuSwapMemorySize** | Pointer to **NullableString** | Size of the swap cpu memory. Only allowed when swapEnabled is true | [optional] [default to "100G"]
**ReservedGpuMemoryForSwapOperations** | Pointer to **NullableString** | Reserved GPU memory for swap operations. Only allowed when swapEnabled is true | [optional] [default to "2G"]
**NodeLevelSchedulerEnabled** | Pointer to **NullableBool** | Indicates if node level scheduler is enabled on the node pool. Replaces the deprecated &#39;overProvisioningRatio&#39; field | [optional] [default to false]

## Methods

### NewGPUResourceOptimization

`func NewGPUResourceOptimization() *GPUResourceOptimization`

NewGPUResourceOptimization instantiates a new GPUResourceOptimization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGPUResourceOptimizationWithDefaults

`func NewGPUResourceOptimizationWithDefaults() *GPUResourceOptimization`

NewGPUResourceOptimizationWithDefaults instantiates a new GPUResourceOptimization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSwapEnabled

`func (o *GPUResourceOptimization) GetSwapEnabled() bool`

GetSwapEnabled returns the SwapEnabled field if non-nil, zero value otherwise.

### GetSwapEnabledOk

`func (o *GPUResourceOptimization) GetSwapEnabledOk() (*bool, bool)`

GetSwapEnabledOk returns a tuple with the SwapEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSwapEnabled

`func (o *GPUResourceOptimization) SetSwapEnabled(v bool)`

SetSwapEnabled sets SwapEnabled field to given value.

### HasSwapEnabled

`func (o *GPUResourceOptimization) HasSwapEnabled() bool`

HasSwapEnabled returns a boolean if a field has been set.

### SetSwapEnabledNil

`func (o *GPUResourceOptimization) SetSwapEnabledNil(b bool)`

 SetSwapEnabledNil sets the value for SwapEnabled to be an explicit nil

### UnsetSwapEnabled
`func (o *GPUResourceOptimization) UnsetSwapEnabled()`

UnsetSwapEnabled ensures that no value is present for SwapEnabled, not even an explicit nil
### GetCpuSwapMemorySize

`func (o *GPUResourceOptimization) GetCpuSwapMemorySize() string`

GetCpuSwapMemorySize returns the CpuSwapMemorySize field if non-nil, zero value otherwise.

### GetCpuSwapMemorySizeOk

`func (o *GPUResourceOptimization) GetCpuSwapMemorySizeOk() (*string, bool)`

GetCpuSwapMemorySizeOk returns a tuple with the CpuSwapMemorySize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuSwapMemorySize

`func (o *GPUResourceOptimization) SetCpuSwapMemorySize(v string)`

SetCpuSwapMemorySize sets CpuSwapMemorySize field to given value.

### HasCpuSwapMemorySize

`func (o *GPUResourceOptimization) HasCpuSwapMemorySize() bool`

HasCpuSwapMemorySize returns a boolean if a field has been set.

### SetCpuSwapMemorySizeNil

`func (o *GPUResourceOptimization) SetCpuSwapMemorySizeNil(b bool)`

 SetCpuSwapMemorySizeNil sets the value for CpuSwapMemorySize to be an explicit nil

### UnsetCpuSwapMemorySize
`func (o *GPUResourceOptimization) UnsetCpuSwapMemorySize()`

UnsetCpuSwapMemorySize ensures that no value is present for CpuSwapMemorySize, not even an explicit nil
### GetReservedGpuMemoryForSwapOperations

`func (o *GPUResourceOptimization) GetReservedGpuMemoryForSwapOperations() string`

GetReservedGpuMemoryForSwapOperations returns the ReservedGpuMemoryForSwapOperations field if non-nil, zero value otherwise.

### GetReservedGpuMemoryForSwapOperationsOk

`func (o *GPUResourceOptimization) GetReservedGpuMemoryForSwapOperationsOk() (*string, bool)`

GetReservedGpuMemoryForSwapOperationsOk returns a tuple with the ReservedGpuMemoryForSwapOperations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReservedGpuMemoryForSwapOperations

`func (o *GPUResourceOptimization) SetReservedGpuMemoryForSwapOperations(v string)`

SetReservedGpuMemoryForSwapOperations sets ReservedGpuMemoryForSwapOperations field to given value.

### HasReservedGpuMemoryForSwapOperations

`func (o *GPUResourceOptimization) HasReservedGpuMemoryForSwapOperations() bool`

HasReservedGpuMemoryForSwapOperations returns a boolean if a field has been set.

### SetReservedGpuMemoryForSwapOperationsNil

`func (o *GPUResourceOptimization) SetReservedGpuMemoryForSwapOperationsNil(b bool)`

 SetReservedGpuMemoryForSwapOperationsNil sets the value for ReservedGpuMemoryForSwapOperations to be an explicit nil

### UnsetReservedGpuMemoryForSwapOperations
`func (o *GPUResourceOptimization) UnsetReservedGpuMemoryForSwapOperations()`

UnsetReservedGpuMemoryForSwapOperations ensures that no value is present for ReservedGpuMemoryForSwapOperations, not even an explicit nil
### GetNodeLevelSchedulerEnabled

`func (o *GPUResourceOptimization) GetNodeLevelSchedulerEnabled() bool`

GetNodeLevelSchedulerEnabled returns the NodeLevelSchedulerEnabled field if non-nil, zero value otherwise.

### GetNodeLevelSchedulerEnabledOk

`func (o *GPUResourceOptimization) GetNodeLevelSchedulerEnabledOk() (*bool, bool)`

GetNodeLevelSchedulerEnabledOk returns a tuple with the NodeLevelSchedulerEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeLevelSchedulerEnabled

`func (o *GPUResourceOptimization) SetNodeLevelSchedulerEnabled(v bool)`

SetNodeLevelSchedulerEnabled sets NodeLevelSchedulerEnabled field to given value.

### HasNodeLevelSchedulerEnabled

`func (o *GPUResourceOptimization) HasNodeLevelSchedulerEnabled() bool`

HasNodeLevelSchedulerEnabled returns a boolean if a field has been set.

### SetNodeLevelSchedulerEnabledNil

`func (o *GPUResourceOptimization) SetNodeLevelSchedulerEnabledNil(b bool)`

 SetNodeLevelSchedulerEnabledNil sets the value for NodeLevelSchedulerEnabled to be an explicit nil

### UnsetNodeLevelSchedulerEnabled
`func (o *GPUResourceOptimization) UnsetNodeLevelSchedulerEnabled()`

UnsetNodeLevelSchedulerEnabled ensures that no value is present for NodeLevelSchedulerEnabled, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


