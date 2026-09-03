# SchedulingConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GpuPodsPlacementStrategy** | Pointer to [**NullableGpuPodsPlacementStrategy**](GpuPodsPlacementStrategy.md) |  | [optional] 
**CpuPodsPlacementStrategy** | Pointer to [**NullableCpuPodsPlacementStrategy**](CpuPodsPlacementStrategy.md) |  | [optional] 
**PlacementStrategy** | Pointer to [**SchedulingConfigurationPlacementStrategy**](SchedulingConfigurationPlacementStrategy.md) |  | [optional] 
**MinGuaranteedRuntime** | Pointer to **NullableString** | Minimum duration (e.g., 5d8h40m) that a preemptible workload is guaranteed to run once scheduled and bound to a node. During this period, it is considered non-preemptible even if a higher-priority workload becomes eligible. Defaults to 0 (immediate preemption allowed). | [optional] 
**TimeBasedFairShare** | Pointer to [**NullableTimeBasedFairShare**](TimeBasedFairShare.md) |  | [optional] 
**NumaAwareScheduling** | Pointer to **NullableBool** | When enabled, the Scheduler places workloads so that GPU, CPU, and CPU memory are allocated within the same NUMA node where possible, minimizing cross-NUMA fragmentation and improving performance. Requires the NodeResourceTopology (NRT) component and a Topology Manager policy other than &#39;none&#39; on the pool&#39;s nodes. When disabled, the Scheduler places workloads using aggregate node capacity only, without considering NUMA topology. | [optional] 

## Methods

### NewSchedulingConfiguration

`func NewSchedulingConfiguration() *SchedulingConfiguration`

NewSchedulingConfiguration instantiates a new SchedulingConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSchedulingConfigurationWithDefaults

`func NewSchedulingConfigurationWithDefaults() *SchedulingConfiguration`

NewSchedulingConfigurationWithDefaults instantiates a new SchedulingConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpuPodsPlacementStrategy

`func (o *SchedulingConfiguration) GetGpuPodsPlacementStrategy() GpuPodsPlacementStrategy`

GetGpuPodsPlacementStrategy returns the GpuPodsPlacementStrategy field if non-nil, zero value otherwise.

### GetGpuPodsPlacementStrategyOk

`func (o *SchedulingConfiguration) GetGpuPodsPlacementStrategyOk() (*GpuPodsPlacementStrategy, bool)`

GetGpuPodsPlacementStrategyOk returns a tuple with the GpuPodsPlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuPodsPlacementStrategy

`func (o *SchedulingConfiguration) SetGpuPodsPlacementStrategy(v GpuPodsPlacementStrategy)`

SetGpuPodsPlacementStrategy sets GpuPodsPlacementStrategy field to given value.

### HasGpuPodsPlacementStrategy

`func (o *SchedulingConfiguration) HasGpuPodsPlacementStrategy() bool`

HasGpuPodsPlacementStrategy returns a boolean if a field has been set.

### SetGpuPodsPlacementStrategyNil

`func (o *SchedulingConfiguration) SetGpuPodsPlacementStrategyNil(b bool)`

 SetGpuPodsPlacementStrategyNil sets the value for GpuPodsPlacementStrategy to be an explicit nil

### UnsetGpuPodsPlacementStrategy
`func (o *SchedulingConfiguration) UnsetGpuPodsPlacementStrategy()`

UnsetGpuPodsPlacementStrategy ensures that no value is present for GpuPodsPlacementStrategy, not even an explicit nil
### GetCpuPodsPlacementStrategy

`func (o *SchedulingConfiguration) GetCpuPodsPlacementStrategy() CpuPodsPlacementStrategy`

GetCpuPodsPlacementStrategy returns the CpuPodsPlacementStrategy field if non-nil, zero value otherwise.

### GetCpuPodsPlacementStrategyOk

`func (o *SchedulingConfiguration) GetCpuPodsPlacementStrategyOk() (*CpuPodsPlacementStrategy, bool)`

GetCpuPodsPlacementStrategyOk returns a tuple with the CpuPodsPlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuPodsPlacementStrategy

`func (o *SchedulingConfiguration) SetCpuPodsPlacementStrategy(v CpuPodsPlacementStrategy)`

SetCpuPodsPlacementStrategy sets CpuPodsPlacementStrategy field to given value.

### HasCpuPodsPlacementStrategy

`func (o *SchedulingConfiguration) HasCpuPodsPlacementStrategy() bool`

HasCpuPodsPlacementStrategy returns a boolean if a field has been set.

### SetCpuPodsPlacementStrategyNil

`func (o *SchedulingConfiguration) SetCpuPodsPlacementStrategyNil(b bool)`

 SetCpuPodsPlacementStrategyNil sets the value for CpuPodsPlacementStrategy to be an explicit nil

### UnsetCpuPodsPlacementStrategy
`func (o *SchedulingConfiguration) UnsetCpuPodsPlacementStrategy()`

UnsetCpuPodsPlacementStrategy ensures that no value is present for CpuPodsPlacementStrategy, not even an explicit nil
### GetPlacementStrategy

`func (o *SchedulingConfiguration) GetPlacementStrategy() SchedulingConfigurationPlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *SchedulingConfiguration) GetPlacementStrategyOk() (*SchedulingConfigurationPlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *SchedulingConfiguration) SetPlacementStrategy(v SchedulingConfigurationPlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *SchedulingConfiguration) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.

### GetMinGuaranteedRuntime

`func (o *SchedulingConfiguration) GetMinGuaranteedRuntime() string`

GetMinGuaranteedRuntime returns the MinGuaranteedRuntime field if non-nil, zero value otherwise.

### GetMinGuaranteedRuntimeOk

`func (o *SchedulingConfiguration) GetMinGuaranteedRuntimeOk() (*string, bool)`

GetMinGuaranteedRuntimeOk returns a tuple with the MinGuaranteedRuntime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinGuaranteedRuntime

`func (o *SchedulingConfiguration) SetMinGuaranteedRuntime(v string)`

SetMinGuaranteedRuntime sets MinGuaranteedRuntime field to given value.

### HasMinGuaranteedRuntime

`func (o *SchedulingConfiguration) HasMinGuaranteedRuntime() bool`

HasMinGuaranteedRuntime returns a boolean if a field has been set.

### SetMinGuaranteedRuntimeNil

`func (o *SchedulingConfiguration) SetMinGuaranteedRuntimeNil(b bool)`

 SetMinGuaranteedRuntimeNil sets the value for MinGuaranteedRuntime to be an explicit nil

### UnsetMinGuaranteedRuntime
`func (o *SchedulingConfiguration) UnsetMinGuaranteedRuntime()`

UnsetMinGuaranteedRuntime ensures that no value is present for MinGuaranteedRuntime, not even an explicit nil
### GetTimeBasedFairShare

`func (o *SchedulingConfiguration) GetTimeBasedFairShare() TimeBasedFairShare`

GetTimeBasedFairShare returns the TimeBasedFairShare field if non-nil, zero value otherwise.

### GetTimeBasedFairShareOk

`func (o *SchedulingConfiguration) GetTimeBasedFairShareOk() (*TimeBasedFairShare, bool)`

GetTimeBasedFairShareOk returns a tuple with the TimeBasedFairShare field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeBasedFairShare

`func (o *SchedulingConfiguration) SetTimeBasedFairShare(v TimeBasedFairShare)`

SetTimeBasedFairShare sets TimeBasedFairShare field to given value.

### HasTimeBasedFairShare

`func (o *SchedulingConfiguration) HasTimeBasedFairShare() bool`

HasTimeBasedFairShare returns a boolean if a field has been set.

### SetTimeBasedFairShareNil

`func (o *SchedulingConfiguration) SetTimeBasedFairShareNil(b bool)`

 SetTimeBasedFairShareNil sets the value for TimeBasedFairShare to be an explicit nil

### UnsetTimeBasedFairShare
`func (o *SchedulingConfiguration) UnsetTimeBasedFairShare()`

UnsetTimeBasedFairShare ensures that no value is present for TimeBasedFairShare, not even an explicit nil
### GetNumaAwareScheduling

`func (o *SchedulingConfiguration) GetNumaAwareScheduling() bool`

GetNumaAwareScheduling returns the NumaAwareScheduling field if non-nil, zero value otherwise.

### GetNumaAwareSchedulingOk

`func (o *SchedulingConfiguration) GetNumaAwareSchedulingOk() (*bool, bool)`

GetNumaAwareSchedulingOk returns a tuple with the NumaAwareScheduling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumaAwareScheduling

`func (o *SchedulingConfiguration) SetNumaAwareScheduling(v bool)`

SetNumaAwareScheduling sets NumaAwareScheduling field to given value.

### HasNumaAwareScheduling

`func (o *SchedulingConfiguration) HasNumaAwareScheduling() bool`

HasNumaAwareScheduling returns a boolean if a field has been set.

### SetNumaAwareSchedulingNil

`func (o *SchedulingConfiguration) SetNumaAwareSchedulingNil(b bool)`

 SetNumaAwareSchedulingNil sets the value for NumaAwareScheduling to be an explicit nil

### UnsetNumaAwareScheduling
`func (o *SchedulingConfiguration) UnsetNumaAwareScheduling()`

UnsetNumaAwareScheduling ensures that no value is present for NumaAwareScheduling, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


