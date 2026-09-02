# NodepoolCreateResponseFieldsPlacementStrategy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Cpu** | Pointer to [**NullablePlacementStrategy**](PlacementStrategy.md) |  | [optional] 
**Gpu** | Pointer to [**NullablePlacementStrategy**](PlacementStrategy.md) |  | [optional] 

## Methods

### NewNodepoolCreateResponseFieldsPlacementStrategy

`func NewNodepoolCreateResponseFieldsPlacementStrategy() *NodepoolCreateResponseFieldsPlacementStrategy`

NewNodepoolCreateResponseFieldsPlacementStrategy instantiates a new NodepoolCreateResponseFieldsPlacementStrategy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolCreateResponseFieldsPlacementStrategyWithDefaults

`func NewNodepoolCreateResponseFieldsPlacementStrategyWithDefaults() *NodepoolCreateResponseFieldsPlacementStrategy`

NewNodepoolCreateResponseFieldsPlacementStrategyWithDefaults instantiates a new NodepoolCreateResponseFieldsPlacementStrategy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpu

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) GetCpu() PlacementStrategy`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) GetCpuOk() (*PlacementStrategy, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) SetCpu(v PlacementStrategy)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### SetCpuNil

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) SetCpuNil(b bool)`

 SetCpuNil sets the value for Cpu to be an explicit nil

### UnsetCpu
`func (o *NodepoolCreateResponseFieldsPlacementStrategy) UnsetCpu()`

UnsetCpu ensures that no value is present for Cpu, not even an explicit nil
### GetGpu

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) GetGpu() PlacementStrategy`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) GetGpuOk() (*PlacementStrategy, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) SetGpu(v PlacementStrategy)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *NodepoolCreateResponseFieldsPlacementStrategy) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *NodepoolCreateResponseFieldsPlacementStrategy) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


