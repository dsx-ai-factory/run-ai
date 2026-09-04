# SchedulingConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PlacementStrategy** | Pointer to [**SchedulingConfigurationPlacementStrategy**](SchedulingConfigurationPlacementStrategy.md) |  | [optional] 
**MinGuaranteedRuntime** | Pointer to **NullableString** | Minimum duration (e.g., 5d8h40m) that a preemptible workload is guaranteed to run once scheduled and bound to a node. During this period, it is considered non-preemptible even if a higher-priority workload becomes eligible. Defaults to 0 (immediate preemption allowed). | [optional] 
**TimeBasedFairShare** | Pointer to [**NullableTimeBasedFairShare**](TimeBasedFairShare.md) |  | [optional] 

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

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


