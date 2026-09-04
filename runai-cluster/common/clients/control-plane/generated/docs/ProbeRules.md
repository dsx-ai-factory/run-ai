# ProbeRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InitialDelaySeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**PeriodSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**TimeoutSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**SuccessThreshold** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**FailureThreshold** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Handler** | Pointer to [**NullableProbeHandlerRules**](ProbeHandlerRules.md) |  | [optional] 

## Methods

### NewProbeRules

`func NewProbeRules() *ProbeRules`

NewProbeRules instantiates a new ProbeRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProbeRulesWithDefaults

`func NewProbeRulesWithDefaults() *ProbeRules`

NewProbeRulesWithDefaults instantiates a new ProbeRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInitialDelaySeconds

`func (o *ProbeRules) GetInitialDelaySeconds() IntegerRules`

GetInitialDelaySeconds returns the InitialDelaySeconds field if non-nil, zero value otherwise.

### GetInitialDelaySecondsOk

`func (o *ProbeRules) GetInitialDelaySecondsOk() (*IntegerRules, bool)`

GetInitialDelaySecondsOk returns a tuple with the InitialDelaySeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitialDelaySeconds

`func (o *ProbeRules) SetInitialDelaySeconds(v IntegerRules)`

SetInitialDelaySeconds sets InitialDelaySeconds field to given value.

### HasInitialDelaySeconds

`func (o *ProbeRules) HasInitialDelaySeconds() bool`

HasInitialDelaySeconds returns a boolean if a field has been set.

### SetInitialDelaySecondsNil

`func (o *ProbeRules) SetInitialDelaySecondsNil(b bool)`

 SetInitialDelaySecondsNil sets the value for InitialDelaySeconds to be an explicit nil

### UnsetInitialDelaySeconds
`func (o *ProbeRules) UnsetInitialDelaySeconds()`

UnsetInitialDelaySeconds ensures that no value is present for InitialDelaySeconds, not even an explicit nil
### GetPeriodSeconds

`func (o *ProbeRules) GetPeriodSeconds() IntegerRules`

GetPeriodSeconds returns the PeriodSeconds field if non-nil, zero value otherwise.

### GetPeriodSecondsOk

`func (o *ProbeRules) GetPeriodSecondsOk() (*IntegerRules, bool)`

GetPeriodSecondsOk returns a tuple with the PeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPeriodSeconds

`func (o *ProbeRules) SetPeriodSeconds(v IntegerRules)`

SetPeriodSeconds sets PeriodSeconds field to given value.

### HasPeriodSeconds

`func (o *ProbeRules) HasPeriodSeconds() bool`

HasPeriodSeconds returns a boolean if a field has been set.

### SetPeriodSecondsNil

`func (o *ProbeRules) SetPeriodSecondsNil(b bool)`

 SetPeriodSecondsNil sets the value for PeriodSeconds to be an explicit nil

### UnsetPeriodSeconds
`func (o *ProbeRules) UnsetPeriodSeconds()`

UnsetPeriodSeconds ensures that no value is present for PeriodSeconds, not even an explicit nil
### GetTimeoutSeconds

`func (o *ProbeRules) GetTimeoutSeconds() IntegerRules`

GetTimeoutSeconds returns the TimeoutSeconds field if non-nil, zero value otherwise.

### GetTimeoutSecondsOk

`func (o *ProbeRules) GetTimeoutSecondsOk() (*IntegerRules, bool)`

GetTimeoutSecondsOk returns a tuple with the TimeoutSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeoutSeconds

`func (o *ProbeRules) SetTimeoutSeconds(v IntegerRules)`

SetTimeoutSeconds sets TimeoutSeconds field to given value.

### HasTimeoutSeconds

`func (o *ProbeRules) HasTimeoutSeconds() bool`

HasTimeoutSeconds returns a boolean if a field has been set.

### SetTimeoutSecondsNil

`func (o *ProbeRules) SetTimeoutSecondsNil(b bool)`

 SetTimeoutSecondsNil sets the value for TimeoutSeconds to be an explicit nil

### UnsetTimeoutSeconds
`func (o *ProbeRules) UnsetTimeoutSeconds()`

UnsetTimeoutSeconds ensures that no value is present for TimeoutSeconds, not even an explicit nil
### GetSuccessThreshold

`func (o *ProbeRules) GetSuccessThreshold() IntegerRules`

GetSuccessThreshold returns the SuccessThreshold field if non-nil, zero value otherwise.

### GetSuccessThresholdOk

`func (o *ProbeRules) GetSuccessThresholdOk() (*IntegerRules, bool)`

GetSuccessThresholdOk returns a tuple with the SuccessThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccessThreshold

`func (o *ProbeRules) SetSuccessThreshold(v IntegerRules)`

SetSuccessThreshold sets SuccessThreshold field to given value.

### HasSuccessThreshold

`func (o *ProbeRules) HasSuccessThreshold() bool`

HasSuccessThreshold returns a boolean if a field has been set.

### SetSuccessThresholdNil

`func (o *ProbeRules) SetSuccessThresholdNil(b bool)`

 SetSuccessThresholdNil sets the value for SuccessThreshold to be an explicit nil

### UnsetSuccessThreshold
`func (o *ProbeRules) UnsetSuccessThreshold()`

UnsetSuccessThreshold ensures that no value is present for SuccessThreshold, not even an explicit nil
### GetFailureThreshold

`func (o *ProbeRules) GetFailureThreshold() IntegerRules`

GetFailureThreshold returns the FailureThreshold field if non-nil, zero value otherwise.

### GetFailureThresholdOk

`func (o *ProbeRules) GetFailureThresholdOk() (*IntegerRules, bool)`

GetFailureThresholdOk returns a tuple with the FailureThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailureThreshold

`func (o *ProbeRules) SetFailureThreshold(v IntegerRules)`

SetFailureThreshold sets FailureThreshold field to given value.

### HasFailureThreshold

`func (o *ProbeRules) HasFailureThreshold() bool`

HasFailureThreshold returns a boolean if a field has been set.

### SetFailureThresholdNil

`func (o *ProbeRules) SetFailureThresholdNil(b bool)`

 SetFailureThresholdNil sets the value for FailureThreshold to be an explicit nil

### UnsetFailureThreshold
`func (o *ProbeRules) UnsetFailureThreshold()`

UnsetFailureThreshold ensures that no value is present for FailureThreshold, not even an explicit nil
### GetHandler

`func (o *ProbeRules) GetHandler() ProbeHandlerRules`

GetHandler returns the Handler field if non-nil, zero value otherwise.

### GetHandlerOk

`func (o *ProbeRules) GetHandlerOk() (*ProbeHandlerRules, bool)`

GetHandlerOk returns a tuple with the Handler field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHandler

`func (o *ProbeRules) SetHandler(v ProbeHandlerRules)`

SetHandler sets Handler field to given value.

### HasHandler

`func (o *ProbeRules) HasHandler() bool`

HasHandler returns a boolean if a field has been set.

### SetHandlerNil

`func (o *ProbeRules) SetHandlerNil(b bool)`

 SetHandlerNil sets the value for Handler to be an explicit nil

### UnsetHandler
`func (o *ProbeRules) UnsetHandler()`

UnsetHandler ensures that no value is present for Handler, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


