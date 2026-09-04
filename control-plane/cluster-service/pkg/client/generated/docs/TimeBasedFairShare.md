# TimeBasedFairShare

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** | Enables time-based fairshare for the node pool. When true, the Scheduler adjusts fairshare based on historical resource usage over a defined time window, allowing past consumption to influence scheduling order. When false, scheduling is based only on the current queue state. [Experimental]  | [optional] [default to false]
**HistoricalUsageWeight** | Pointer to **NullableFloat32** | Set the weight value that determines how a project’s historical resource usage affects its fairshare.  - 0.0: Ignore historical usage data (100% current usage).  - 1.0: Historical usage data has higher impact, it may cause fairshare to be 0.  - Values between 0.0 and 1.0 combine historical and current usage to maintain a balanced fairshare  - Values above 1 will cause historical usage to have an even higher impact on fairshare.  | [optional] [default to 1]
**DecayHalfLife** | Pointer to **NullableString** | Set the half-life time to define the decay rate of a project’s historical resource usage. The half-life defines how long it takes for past usage to diminish to half of its original influence on a project&#39;s fairshare.  - 0: disables decay; all historical usage is weighted equally.  - &gt; 0: older usage gradually contributes less over time (e.g., with a 24-hour half-life, usage from 24 hours ago contributes 50%, from 48 hours ago 25%, and so on).  | [optional] [default to "0h"]
**WindowType** | Pointer to [**NullableTimeBasedFairShareWindowType**](TimeBasedFairShareWindowType.md) |  | [optional] [default to TIMEBASEDFAIRSHAREWINDOWTYPE_SLIDING]
**WindowDuration** | Pointer to **NullableString** | The time period over which historical usage is evaluated. Only usage within this window contributes to fairshare. For sliding windows, this is a moving period (e.g., 7d &#x3D; always the last 7 days). For tumbling windows, this is the fixed cycle duration after which the window resets.  | [optional] [default to "7d"]
**TumblingWindowStartTime** | Pointer to **NullableTime** | An RFC-3339 timestamp that defines the start time for all tumbling windows. Used only when windowType &#x3D; Tumbling. This timestamp marks the beginning of the first fixed window, and subsequent windows repeat at intervals equal to the configured WindowDuration. For example, with a 7-day window, an anchor of 2025-01-06T08:00:00Z produces weekly windows starting every Monday at 08:00 UTC.  | [optional] 

## Methods

### NewTimeBasedFairShare

`func NewTimeBasedFairShare() *TimeBasedFairShare`

NewTimeBasedFairShare instantiates a new TimeBasedFairShare object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTimeBasedFairShareWithDefaults

`func NewTimeBasedFairShareWithDefaults() *TimeBasedFairShare`

NewTimeBasedFairShareWithDefaults instantiates a new TimeBasedFairShare object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *TimeBasedFairShare) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *TimeBasedFairShare) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *TimeBasedFairShare) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *TimeBasedFairShare) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *TimeBasedFairShare) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *TimeBasedFairShare) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetHistoricalUsageWeight

`func (o *TimeBasedFairShare) GetHistoricalUsageWeight() float32`

GetHistoricalUsageWeight returns the HistoricalUsageWeight field if non-nil, zero value otherwise.

### GetHistoricalUsageWeightOk

`func (o *TimeBasedFairShare) GetHistoricalUsageWeightOk() (*float32, bool)`

GetHistoricalUsageWeightOk returns a tuple with the HistoricalUsageWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHistoricalUsageWeight

`func (o *TimeBasedFairShare) SetHistoricalUsageWeight(v float32)`

SetHistoricalUsageWeight sets HistoricalUsageWeight field to given value.

### HasHistoricalUsageWeight

`func (o *TimeBasedFairShare) HasHistoricalUsageWeight() bool`

HasHistoricalUsageWeight returns a boolean if a field has been set.

### SetHistoricalUsageWeightNil

`func (o *TimeBasedFairShare) SetHistoricalUsageWeightNil(b bool)`

 SetHistoricalUsageWeightNil sets the value for HistoricalUsageWeight to be an explicit nil

### UnsetHistoricalUsageWeight
`func (o *TimeBasedFairShare) UnsetHistoricalUsageWeight()`

UnsetHistoricalUsageWeight ensures that no value is present for HistoricalUsageWeight, not even an explicit nil
### GetDecayHalfLife

`func (o *TimeBasedFairShare) GetDecayHalfLife() string`

GetDecayHalfLife returns the DecayHalfLife field if non-nil, zero value otherwise.

### GetDecayHalfLifeOk

`func (o *TimeBasedFairShare) GetDecayHalfLifeOk() (*string, bool)`

GetDecayHalfLifeOk returns a tuple with the DecayHalfLife field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDecayHalfLife

`func (o *TimeBasedFairShare) SetDecayHalfLife(v string)`

SetDecayHalfLife sets DecayHalfLife field to given value.

### HasDecayHalfLife

`func (o *TimeBasedFairShare) HasDecayHalfLife() bool`

HasDecayHalfLife returns a boolean if a field has been set.

### SetDecayHalfLifeNil

`func (o *TimeBasedFairShare) SetDecayHalfLifeNil(b bool)`

 SetDecayHalfLifeNil sets the value for DecayHalfLife to be an explicit nil

### UnsetDecayHalfLife
`func (o *TimeBasedFairShare) UnsetDecayHalfLife()`

UnsetDecayHalfLife ensures that no value is present for DecayHalfLife, not even an explicit nil
### GetWindowType

`func (o *TimeBasedFairShare) GetWindowType() TimeBasedFairShareWindowType`

GetWindowType returns the WindowType field if non-nil, zero value otherwise.

### GetWindowTypeOk

`func (o *TimeBasedFairShare) GetWindowTypeOk() (*TimeBasedFairShareWindowType, bool)`

GetWindowTypeOk returns a tuple with the WindowType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowType

`func (o *TimeBasedFairShare) SetWindowType(v TimeBasedFairShareWindowType)`

SetWindowType sets WindowType field to given value.

### HasWindowType

`func (o *TimeBasedFairShare) HasWindowType() bool`

HasWindowType returns a boolean if a field has been set.

### SetWindowTypeNil

`func (o *TimeBasedFairShare) SetWindowTypeNil(b bool)`

 SetWindowTypeNil sets the value for WindowType to be an explicit nil

### UnsetWindowType
`func (o *TimeBasedFairShare) UnsetWindowType()`

UnsetWindowType ensures that no value is present for WindowType, not even an explicit nil
### GetWindowDuration

`func (o *TimeBasedFairShare) GetWindowDuration() string`

GetWindowDuration returns the WindowDuration field if non-nil, zero value otherwise.

### GetWindowDurationOk

`func (o *TimeBasedFairShare) GetWindowDurationOk() (*string, bool)`

GetWindowDurationOk returns a tuple with the WindowDuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowDuration

`func (o *TimeBasedFairShare) SetWindowDuration(v string)`

SetWindowDuration sets WindowDuration field to given value.

### HasWindowDuration

`func (o *TimeBasedFairShare) HasWindowDuration() bool`

HasWindowDuration returns a boolean if a field has been set.

### SetWindowDurationNil

`func (o *TimeBasedFairShare) SetWindowDurationNil(b bool)`

 SetWindowDurationNil sets the value for WindowDuration to be an explicit nil

### UnsetWindowDuration
`func (o *TimeBasedFairShare) UnsetWindowDuration()`

UnsetWindowDuration ensures that no value is present for WindowDuration, not even an explicit nil
### GetTumblingWindowStartTime

`func (o *TimeBasedFairShare) GetTumblingWindowStartTime() time.Time`

GetTumblingWindowStartTime returns the TumblingWindowStartTime field if non-nil, zero value otherwise.

### GetTumblingWindowStartTimeOk

`func (o *TimeBasedFairShare) GetTumblingWindowStartTimeOk() (*time.Time, bool)`

GetTumblingWindowStartTimeOk returns a tuple with the TumblingWindowStartTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTumblingWindowStartTime

`func (o *TimeBasedFairShare) SetTumblingWindowStartTime(v time.Time)`

SetTumblingWindowStartTime sets TumblingWindowStartTime field to given value.

### HasTumblingWindowStartTime

`func (o *TimeBasedFairShare) HasTumblingWindowStartTime() bool`

HasTumblingWindowStartTime returns a boolean if a field has been set.

### SetTumblingWindowStartTimeNil

`func (o *TimeBasedFairShare) SetTumblingWindowStartTimeNil(b bool)`

 SetTumblingWindowStartTimeNil sets the value for TumblingWindowStartTime to be an explicit nil

### UnsetTumblingWindowStartTime
`func (o *TimeBasedFairShare) UnsetTumblingWindowStartTime()`

UnsetTumblingWindowStartTime ensures that no value is present for TumblingWindowStartTime, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


