# SynchronizedStart

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** | Turns on the synchronized start for the group. The default is false. | [optional] [default to false]
**Port** | Pointer to **NullableInt32** | The port that the tasks of the group use to reach each other before they start. It must be a port that the group&#39;s tasks don&#39;t use. (mandatory when enabled) | [optional] 
**TimeoutSeconds** | Pointer to **NullableInt32** | How long, in seconds, each task waits for the group&#39;s other tasks. (mandatory when enabled) | [optional] 

## Methods

### NewSynchronizedStart

`func NewSynchronizedStart() *SynchronizedStart`

NewSynchronizedStart instantiates a new SynchronizedStart object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSynchronizedStartWithDefaults

`func NewSynchronizedStartWithDefaults() *SynchronizedStart`

NewSynchronizedStartWithDefaults instantiates a new SynchronizedStart object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *SynchronizedStart) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *SynchronizedStart) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *SynchronizedStart) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *SynchronizedStart) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *SynchronizedStart) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *SynchronizedStart) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetPort

`func (o *SynchronizedStart) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *SynchronizedStart) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *SynchronizedStart) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *SynchronizedStart) HasPort() bool`

HasPort returns a boolean if a field has been set.

### SetPortNil

`func (o *SynchronizedStart) SetPortNil(b bool)`

 SetPortNil sets the value for Port to be an explicit nil

### UnsetPort
`func (o *SynchronizedStart) UnsetPort()`

UnsetPort ensures that no value is present for Port, not even an explicit nil
### GetTimeoutSeconds

`func (o *SynchronizedStart) GetTimeoutSeconds() int32`

GetTimeoutSeconds returns the TimeoutSeconds field if non-nil, zero value otherwise.

### GetTimeoutSecondsOk

`func (o *SynchronizedStart) GetTimeoutSecondsOk() (*int32, bool)`

GetTimeoutSecondsOk returns a tuple with the TimeoutSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeoutSeconds

`func (o *SynchronizedStart) SetTimeoutSeconds(v int32)`

SetTimeoutSeconds sets TimeoutSeconds field to given value.

### HasTimeoutSeconds

`func (o *SynchronizedStart) HasTimeoutSeconds() bool`

HasTimeoutSeconds returns a boolean if a field has been set.

### SetTimeoutSecondsNil

`func (o *SynchronizedStart) SetTimeoutSecondsNil(b bool)`

 SetTimeoutSecondsNil sets the value for TimeoutSeconds to be an explicit nil

### UnsetTimeoutSeconds
`func (o *SynchronizedStart) UnsetTimeoutSeconds()`

UnsetTimeoutSeconds ensures that no value is present for TimeoutSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


