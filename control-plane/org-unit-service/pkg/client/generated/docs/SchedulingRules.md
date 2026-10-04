# SchedulingRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InteractiveJobTimeLimitSeconds** | Pointer to **NullableInt32** | interactive job time limit in seconds | [optional] 
**InteractiveJobMaxIdleDurationSeconds** | Pointer to **NullableInt32** | interactive job max idle duration in seconds | [optional] 
**InteractiveJobPreemptIdleDurationSeconds** | Pointer to **NullableInt32** | interactive job preempt idle duration in seconds | [optional] 
**TrainingJobMaxIdleDurationSeconds** | Pointer to **NullableInt32** | training job max idle duration in seconds | [optional] 
**TrainingJobTimeLimitSeconds** | Pointer to **NullableInt32** | training job time limit in seconds | [optional] 

## Methods

### NewSchedulingRules

`func NewSchedulingRules() *SchedulingRules`

NewSchedulingRules instantiates a new SchedulingRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSchedulingRulesWithDefaults

`func NewSchedulingRulesWithDefaults() *SchedulingRules`

NewSchedulingRulesWithDefaults instantiates a new SchedulingRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInteractiveJobTimeLimitSeconds

`func (o *SchedulingRules) GetInteractiveJobTimeLimitSeconds() int32`

GetInteractiveJobTimeLimitSeconds returns the InteractiveJobTimeLimitSeconds field if non-nil, zero value otherwise.

### GetInteractiveJobTimeLimitSecondsOk

`func (o *SchedulingRules) GetInteractiveJobTimeLimitSecondsOk() (*int32, bool)`

GetInteractiveJobTimeLimitSecondsOk returns a tuple with the InteractiveJobTimeLimitSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobTimeLimitSeconds

`func (o *SchedulingRules) SetInteractiveJobTimeLimitSeconds(v int32)`

SetInteractiveJobTimeLimitSeconds sets InteractiveJobTimeLimitSeconds field to given value.

### HasInteractiveJobTimeLimitSeconds

`func (o *SchedulingRules) HasInteractiveJobTimeLimitSeconds() bool`

HasInteractiveJobTimeLimitSeconds returns a boolean if a field has been set.

### SetInteractiveJobTimeLimitSecondsNil

`func (o *SchedulingRules) SetInteractiveJobTimeLimitSecondsNil(b bool)`

 SetInteractiveJobTimeLimitSecondsNil sets the value for InteractiveJobTimeLimitSeconds to be an explicit nil

### UnsetInteractiveJobTimeLimitSeconds
`func (o *SchedulingRules) UnsetInteractiveJobTimeLimitSeconds()`

UnsetInteractiveJobTimeLimitSeconds ensures that no value is present for InteractiveJobTimeLimitSeconds, not even an explicit nil
### GetInteractiveJobMaxIdleDurationSeconds

`func (o *SchedulingRules) GetInteractiveJobMaxIdleDurationSeconds() int32`

GetInteractiveJobMaxIdleDurationSeconds returns the InteractiveJobMaxIdleDurationSeconds field if non-nil, zero value otherwise.

### GetInteractiveJobMaxIdleDurationSecondsOk

`func (o *SchedulingRules) GetInteractiveJobMaxIdleDurationSecondsOk() (*int32, bool)`

GetInteractiveJobMaxIdleDurationSecondsOk returns a tuple with the InteractiveJobMaxIdleDurationSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobMaxIdleDurationSeconds

`func (o *SchedulingRules) SetInteractiveJobMaxIdleDurationSeconds(v int32)`

SetInteractiveJobMaxIdleDurationSeconds sets InteractiveJobMaxIdleDurationSeconds field to given value.

### HasInteractiveJobMaxIdleDurationSeconds

`func (o *SchedulingRules) HasInteractiveJobMaxIdleDurationSeconds() bool`

HasInteractiveJobMaxIdleDurationSeconds returns a boolean if a field has been set.

### SetInteractiveJobMaxIdleDurationSecondsNil

`func (o *SchedulingRules) SetInteractiveJobMaxIdleDurationSecondsNil(b bool)`

 SetInteractiveJobMaxIdleDurationSecondsNil sets the value for InteractiveJobMaxIdleDurationSeconds to be an explicit nil

### UnsetInteractiveJobMaxIdleDurationSeconds
`func (o *SchedulingRules) UnsetInteractiveJobMaxIdleDurationSeconds()`

UnsetInteractiveJobMaxIdleDurationSeconds ensures that no value is present for InteractiveJobMaxIdleDurationSeconds, not even an explicit nil
### GetInteractiveJobPreemptIdleDurationSeconds

`func (o *SchedulingRules) GetInteractiveJobPreemptIdleDurationSeconds() int32`

GetInteractiveJobPreemptIdleDurationSeconds returns the InteractiveJobPreemptIdleDurationSeconds field if non-nil, zero value otherwise.

### GetInteractiveJobPreemptIdleDurationSecondsOk

`func (o *SchedulingRules) GetInteractiveJobPreemptIdleDurationSecondsOk() (*int32, bool)`

GetInteractiveJobPreemptIdleDurationSecondsOk returns a tuple with the InteractiveJobPreemptIdleDurationSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobPreemptIdleDurationSeconds

`func (o *SchedulingRules) SetInteractiveJobPreemptIdleDurationSeconds(v int32)`

SetInteractiveJobPreemptIdleDurationSeconds sets InteractiveJobPreemptIdleDurationSeconds field to given value.

### HasInteractiveJobPreemptIdleDurationSeconds

`func (o *SchedulingRules) HasInteractiveJobPreemptIdleDurationSeconds() bool`

HasInteractiveJobPreemptIdleDurationSeconds returns a boolean if a field has been set.

### SetInteractiveJobPreemptIdleDurationSecondsNil

`func (o *SchedulingRules) SetInteractiveJobPreemptIdleDurationSecondsNil(b bool)`

 SetInteractiveJobPreemptIdleDurationSecondsNil sets the value for InteractiveJobPreemptIdleDurationSeconds to be an explicit nil

### UnsetInteractiveJobPreemptIdleDurationSeconds
`func (o *SchedulingRules) UnsetInteractiveJobPreemptIdleDurationSeconds()`

UnsetInteractiveJobPreemptIdleDurationSeconds ensures that no value is present for InteractiveJobPreemptIdleDurationSeconds, not even an explicit nil
### GetTrainingJobMaxIdleDurationSeconds

`func (o *SchedulingRules) GetTrainingJobMaxIdleDurationSeconds() int32`

GetTrainingJobMaxIdleDurationSeconds returns the TrainingJobMaxIdleDurationSeconds field if non-nil, zero value otherwise.

### GetTrainingJobMaxIdleDurationSecondsOk

`func (o *SchedulingRules) GetTrainingJobMaxIdleDurationSecondsOk() (*int32, bool)`

GetTrainingJobMaxIdleDurationSecondsOk returns a tuple with the TrainingJobMaxIdleDurationSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobMaxIdleDurationSeconds

`func (o *SchedulingRules) SetTrainingJobMaxIdleDurationSeconds(v int32)`

SetTrainingJobMaxIdleDurationSeconds sets TrainingJobMaxIdleDurationSeconds field to given value.

### HasTrainingJobMaxIdleDurationSeconds

`func (o *SchedulingRules) HasTrainingJobMaxIdleDurationSeconds() bool`

HasTrainingJobMaxIdleDurationSeconds returns a boolean if a field has been set.

### SetTrainingJobMaxIdleDurationSecondsNil

`func (o *SchedulingRules) SetTrainingJobMaxIdleDurationSecondsNil(b bool)`

 SetTrainingJobMaxIdleDurationSecondsNil sets the value for TrainingJobMaxIdleDurationSeconds to be an explicit nil

### UnsetTrainingJobMaxIdleDurationSeconds
`func (o *SchedulingRules) UnsetTrainingJobMaxIdleDurationSeconds()`

UnsetTrainingJobMaxIdleDurationSeconds ensures that no value is present for TrainingJobMaxIdleDurationSeconds, not even an explicit nil
### GetTrainingJobTimeLimitSeconds

`func (o *SchedulingRules) GetTrainingJobTimeLimitSeconds() int32`

GetTrainingJobTimeLimitSeconds returns the TrainingJobTimeLimitSeconds field if non-nil, zero value otherwise.

### GetTrainingJobTimeLimitSecondsOk

`func (o *SchedulingRules) GetTrainingJobTimeLimitSecondsOk() (*int32, bool)`

GetTrainingJobTimeLimitSecondsOk returns a tuple with the TrainingJobTimeLimitSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobTimeLimitSeconds

`func (o *SchedulingRules) SetTrainingJobTimeLimitSeconds(v int32)`

SetTrainingJobTimeLimitSeconds sets TrainingJobTimeLimitSeconds field to given value.

### HasTrainingJobTimeLimitSeconds

`func (o *SchedulingRules) HasTrainingJobTimeLimitSeconds() bool`

HasTrainingJobTimeLimitSeconds returns a boolean if a field has been set.

### SetTrainingJobTimeLimitSecondsNil

`func (o *SchedulingRules) SetTrainingJobTimeLimitSecondsNil(b bool)`

 SetTrainingJobTimeLimitSecondsNil sets the value for TrainingJobTimeLimitSeconds to be an explicit nil

### UnsetTrainingJobTimeLimitSeconds
`func (o *SchedulingRules) UnsetTrainingJobTimeLimitSeconds()`

UnsetTrainingJobTimeLimitSeconds ensures that no value is present for TrainingJobTimeLimitSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


