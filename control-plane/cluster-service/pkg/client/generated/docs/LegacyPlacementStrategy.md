# LegacyPlacementStrategy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Cpu** | Pointer to **string** | scheduling strategy for cpu | [optional] 
**Gpu** | Pointer to **string** | scheduling strategy for gpu | [optional] 

## Methods

### NewLegacyPlacementStrategy

`func NewLegacyPlacementStrategy() *LegacyPlacementStrategy`

NewLegacyPlacementStrategy instantiates a new LegacyPlacementStrategy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyPlacementStrategyWithDefaults

`func NewLegacyPlacementStrategyWithDefaults() *LegacyPlacementStrategy`

NewLegacyPlacementStrategyWithDefaults instantiates a new LegacyPlacementStrategy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpu

`func (o *LegacyPlacementStrategy) GetCpu() string`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *LegacyPlacementStrategy) GetCpuOk() (*string, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *LegacyPlacementStrategy) SetCpu(v string)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *LegacyPlacementStrategy) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetGpu

`func (o *LegacyPlacementStrategy) GetGpu() string`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *LegacyPlacementStrategy) GetGpuOk() (*string, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *LegacyPlacementStrategy) SetGpu(v string)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *LegacyPlacementStrategy) HasGpu() bool`

HasGpu returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


