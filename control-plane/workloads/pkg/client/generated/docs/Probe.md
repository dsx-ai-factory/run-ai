# Probe

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InitialDelaySeconds** | Pointer to **NullableInt32** | Number of seconds after the container has started before liveness or readiness probes are initiated. | [optional] 
**PeriodSeconds** | Pointer to **NullableInt32** | How often (in seconds) to perform the probe. | [optional] 
**TimeoutSeconds** | Pointer to **NullableInt32** | Number of seconds after which the probe times out. | [optional] 
**SuccessThreshold** | Pointer to **NullableInt32** | Minimum consecutive successes for the probe to be considered successful after having failed. | [optional] 
**FailureThreshold** | Pointer to **NullableInt32** | When a probe fails, the number of times to try before giving up. | [optional] 
**Handler** | Pointer to [**NullableProbeHandler**](ProbeHandler.md) |  | [optional] 

## Methods

### NewProbe

`func NewProbe() *Probe`

NewProbe instantiates a new Probe object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProbeWithDefaults

`func NewProbeWithDefaults() *Probe`

NewProbeWithDefaults instantiates a new Probe object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInitialDelaySeconds

`func (o *Probe) GetInitialDelaySeconds() int32`

GetInitialDelaySeconds returns the InitialDelaySeconds field if non-nil, zero value otherwise.

### GetInitialDelaySecondsOk

`func (o *Probe) GetInitialDelaySecondsOk() (*int32, bool)`

GetInitialDelaySecondsOk returns a tuple with the InitialDelaySeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitialDelaySeconds

`func (o *Probe) SetInitialDelaySeconds(v int32)`

SetInitialDelaySeconds sets InitialDelaySeconds field to given value.

### HasInitialDelaySeconds

`func (o *Probe) HasInitialDelaySeconds() bool`

HasInitialDelaySeconds returns a boolean if a field has been set.

### SetInitialDelaySecondsNil

`func (o *Probe) SetInitialDelaySecondsNil(b bool)`

 SetInitialDelaySecondsNil sets the value for InitialDelaySeconds to be an explicit nil

### UnsetInitialDelaySeconds
`func (o *Probe) UnsetInitialDelaySeconds()`

UnsetInitialDelaySeconds ensures that no value is present for InitialDelaySeconds, not even an explicit nil
### GetPeriodSeconds

`func (o *Probe) GetPeriodSeconds() int32`

GetPeriodSeconds returns the PeriodSeconds field if non-nil, zero value otherwise.

### GetPeriodSecondsOk

`func (o *Probe) GetPeriodSecondsOk() (*int32, bool)`

GetPeriodSecondsOk returns a tuple with the PeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPeriodSeconds

`func (o *Probe) SetPeriodSeconds(v int32)`

SetPeriodSeconds sets PeriodSeconds field to given value.

### HasPeriodSeconds

`func (o *Probe) HasPeriodSeconds() bool`

HasPeriodSeconds returns a boolean if a field has been set.

### SetPeriodSecondsNil

`func (o *Probe) SetPeriodSecondsNil(b bool)`

 SetPeriodSecondsNil sets the value for PeriodSeconds to be an explicit nil

### UnsetPeriodSeconds
`func (o *Probe) UnsetPeriodSeconds()`

UnsetPeriodSeconds ensures that no value is present for PeriodSeconds, not even an explicit nil
### GetTimeoutSeconds

`func (o *Probe) GetTimeoutSeconds() int32`

GetTimeoutSeconds returns the TimeoutSeconds field if non-nil, zero value otherwise.

### GetTimeoutSecondsOk

`func (o *Probe) GetTimeoutSecondsOk() (*int32, bool)`

GetTimeoutSecondsOk returns a tuple with the TimeoutSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeoutSeconds

`func (o *Probe) SetTimeoutSeconds(v int32)`

SetTimeoutSeconds sets TimeoutSeconds field to given value.

### HasTimeoutSeconds

`func (o *Probe) HasTimeoutSeconds() bool`

HasTimeoutSeconds returns a boolean if a field has been set.

### SetTimeoutSecondsNil

`func (o *Probe) SetTimeoutSecondsNil(b bool)`

 SetTimeoutSecondsNil sets the value for TimeoutSeconds to be an explicit nil

### UnsetTimeoutSeconds
`func (o *Probe) UnsetTimeoutSeconds()`

UnsetTimeoutSeconds ensures that no value is present for TimeoutSeconds, not even an explicit nil
### GetSuccessThreshold

`func (o *Probe) GetSuccessThreshold() int32`

GetSuccessThreshold returns the SuccessThreshold field if non-nil, zero value otherwise.

### GetSuccessThresholdOk

`func (o *Probe) GetSuccessThresholdOk() (*int32, bool)`

GetSuccessThresholdOk returns a tuple with the SuccessThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccessThreshold

`func (o *Probe) SetSuccessThreshold(v int32)`

SetSuccessThreshold sets SuccessThreshold field to given value.

### HasSuccessThreshold

`func (o *Probe) HasSuccessThreshold() bool`

HasSuccessThreshold returns a boolean if a field has been set.

### SetSuccessThresholdNil

`func (o *Probe) SetSuccessThresholdNil(b bool)`

 SetSuccessThresholdNil sets the value for SuccessThreshold to be an explicit nil

### UnsetSuccessThreshold
`func (o *Probe) UnsetSuccessThreshold()`

UnsetSuccessThreshold ensures that no value is present for SuccessThreshold, not even an explicit nil
### GetFailureThreshold

`func (o *Probe) GetFailureThreshold() int32`

GetFailureThreshold returns the FailureThreshold field if non-nil, zero value otherwise.

### GetFailureThresholdOk

`func (o *Probe) GetFailureThresholdOk() (*int32, bool)`

GetFailureThresholdOk returns a tuple with the FailureThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailureThreshold

`func (o *Probe) SetFailureThreshold(v int32)`

SetFailureThreshold sets FailureThreshold field to given value.

### HasFailureThreshold

`func (o *Probe) HasFailureThreshold() bool`

HasFailureThreshold returns a boolean if a field has been set.

### SetFailureThresholdNil

`func (o *Probe) SetFailureThresholdNil(b bool)`

 SetFailureThresholdNil sets the value for FailureThreshold to be an explicit nil

### UnsetFailureThreshold
`func (o *Probe) UnsetFailureThreshold()`

UnsetFailureThreshold ensures that no value is present for FailureThreshold, not even an explicit nil
### GetHandler

`func (o *Probe) GetHandler() ProbeHandler`

GetHandler returns the Handler field if non-nil, zero value otherwise.

### GetHandlerOk

`func (o *Probe) GetHandlerOk() (*ProbeHandler, bool)`

GetHandlerOk returns a tuple with the Handler field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHandler

`func (o *Probe) SetHandler(v ProbeHandler)`

SetHandler sets Handler field to given value.

### HasHandler

`func (o *Probe) HasHandler() bool`

HasHandler returns a boolean if a field has been set.

### SetHandlerNil

`func (o *Probe) SetHandlerNil(b bool)`

 SetHandlerNil sets the value for Handler to be an explicit nil

### UnsetHandler
`func (o *Probe) UnsetHandler()`

UnsetHandler ensures that no value is present for Handler, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


