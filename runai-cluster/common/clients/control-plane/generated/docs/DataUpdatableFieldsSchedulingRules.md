# DataUpdatableFieldsSchedulingRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InteractiveJobTimeLimitSeconds** | Pointer to **NullableInt32** | interactive job time limit in seconds | [optional] 
**InteractiveJobMaxIdleDurationSeconds** | Pointer to **NullableInt32** | interactive job max idle duration in seconds | [optional] 
**InteractiveJobPreemptIdleDurationSeconds** | Pointer to **NullableInt32** | interactive job preempt idle duration in seconds | [optional] 
**TrainingJobMaxIdleDurationSeconds** | Pointer to **NullableInt32** | training job max idle duration in seconds | [optional] 
**TrainingJobTimeLimitSeconds** | Pointer to **NullableInt32** | training job time limit in seconds | [optional] 

## Methods

### NewDataUpdatableFieldsSchedulingRules

`func NewDataUpdatableFieldsSchedulingRules() *DataUpdatableFieldsSchedulingRules`

NewDataUpdatableFieldsSchedulingRules instantiates a new DataUpdatableFieldsSchedulingRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataUpdatableFieldsSchedulingRulesWithDefaults

`func NewDataUpdatableFieldsSchedulingRulesWithDefaults() *DataUpdatableFieldsSchedulingRules`

NewDataUpdatableFieldsSchedulingRulesWithDefaults instantiates a new DataUpdatableFieldsSchedulingRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInteractiveJobTimeLimitSeconds

`func (o *DataUpdatableFieldsSchedulingRules) GetInteractiveJobTimeLimitSeconds() int32`

GetInteractiveJobTimeLimitSeconds returns the InteractiveJobTimeLimitSeconds field if non-nil, zero value otherwise.

### GetInteractiveJobTimeLimitSecondsOk

`func (o *DataUpdatableFieldsSchedulingRules) GetInteractiveJobTimeLimitSecondsOk() (*int32, bool)`

GetInteractiveJobTimeLimitSecondsOk returns a tuple with the InteractiveJobTimeLimitSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobTimeLimitSeconds

`func (o *DataUpdatableFieldsSchedulingRules) SetInteractiveJobTimeLimitSeconds(v int32)`

SetInteractiveJobTimeLimitSeconds sets InteractiveJobTimeLimitSeconds field to given value.

### HasInteractiveJobTimeLimitSeconds

`func (o *DataUpdatableFieldsSchedulingRules) HasInteractiveJobTimeLimitSeconds() bool`

HasInteractiveJobTimeLimitSeconds returns a boolean if a field has been set.

### SetInteractiveJobTimeLimitSecondsNil

`func (o *DataUpdatableFieldsSchedulingRules) SetInteractiveJobTimeLimitSecondsNil(b bool)`

 SetInteractiveJobTimeLimitSecondsNil sets the value for InteractiveJobTimeLimitSeconds to be an explicit nil

### UnsetInteractiveJobTimeLimitSeconds
`func (o *DataUpdatableFieldsSchedulingRules) UnsetInteractiveJobTimeLimitSeconds()`

UnsetInteractiveJobTimeLimitSeconds ensures that no value is present for InteractiveJobTimeLimitSeconds, not even an explicit nil
### GetInteractiveJobMaxIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) GetInteractiveJobMaxIdleDurationSeconds() int32`

GetInteractiveJobMaxIdleDurationSeconds returns the InteractiveJobMaxIdleDurationSeconds field if non-nil, zero value otherwise.

### GetInteractiveJobMaxIdleDurationSecondsOk

`func (o *DataUpdatableFieldsSchedulingRules) GetInteractiveJobMaxIdleDurationSecondsOk() (*int32, bool)`

GetInteractiveJobMaxIdleDurationSecondsOk returns a tuple with the InteractiveJobMaxIdleDurationSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobMaxIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) SetInteractiveJobMaxIdleDurationSeconds(v int32)`

SetInteractiveJobMaxIdleDurationSeconds sets InteractiveJobMaxIdleDurationSeconds field to given value.

### HasInteractiveJobMaxIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) HasInteractiveJobMaxIdleDurationSeconds() bool`

HasInteractiveJobMaxIdleDurationSeconds returns a boolean if a field has been set.

### SetInteractiveJobMaxIdleDurationSecondsNil

`func (o *DataUpdatableFieldsSchedulingRules) SetInteractiveJobMaxIdleDurationSecondsNil(b bool)`

 SetInteractiveJobMaxIdleDurationSecondsNil sets the value for InteractiveJobMaxIdleDurationSeconds to be an explicit nil

### UnsetInteractiveJobMaxIdleDurationSeconds
`func (o *DataUpdatableFieldsSchedulingRules) UnsetInteractiveJobMaxIdleDurationSeconds()`

UnsetInteractiveJobMaxIdleDurationSeconds ensures that no value is present for InteractiveJobMaxIdleDurationSeconds, not even an explicit nil
### GetInteractiveJobPreemptIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) GetInteractiveJobPreemptIdleDurationSeconds() int32`

GetInteractiveJobPreemptIdleDurationSeconds returns the InteractiveJobPreemptIdleDurationSeconds field if non-nil, zero value otherwise.

### GetInteractiveJobPreemptIdleDurationSecondsOk

`func (o *DataUpdatableFieldsSchedulingRules) GetInteractiveJobPreemptIdleDurationSecondsOk() (*int32, bool)`

GetInteractiveJobPreemptIdleDurationSecondsOk returns a tuple with the InteractiveJobPreemptIdleDurationSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobPreemptIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) SetInteractiveJobPreemptIdleDurationSeconds(v int32)`

SetInteractiveJobPreemptIdleDurationSeconds sets InteractiveJobPreemptIdleDurationSeconds field to given value.

### HasInteractiveJobPreemptIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) HasInteractiveJobPreemptIdleDurationSeconds() bool`

HasInteractiveJobPreemptIdleDurationSeconds returns a boolean if a field has been set.

### SetInteractiveJobPreemptIdleDurationSecondsNil

`func (o *DataUpdatableFieldsSchedulingRules) SetInteractiveJobPreemptIdleDurationSecondsNil(b bool)`

 SetInteractiveJobPreemptIdleDurationSecondsNil sets the value for InteractiveJobPreemptIdleDurationSeconds to be an explicit nil

### UnsetInteractiveJobPreemptIdleDurationSeconds
`func (o *DataUpdatableFieldsSchedulingRules) UnsetInteractiveJobPreemptIdleDurationSeconds()`

UnsetInteractiveJobPreemptIdleDurationSeconds ensures that no value is present for InteractiveJobPreemptIdleDurationSeconds, not even an explicit nil
### GetTrainingJobMaxIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) GetTrainingJobMaxIdleDurationSeconds() int32`

GetTrainingJobMaxIdleDurationSeconds returns the TrainingJobMaxIdleDurationSeconds field if non-nil, zero value otherwise.

### GetTrainingJobMaxIdleDurationSecondsOk

`func (o *DataUpdatableFieldsSchedulingRules) GetTrainingJobMaxIdleDurationSecondsOk() (*int32, bool)`

GetTrainingJobMaxIdleDurationSecondsOk returns a tuple with the TrainingJobMaxIdleDurationSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobMaxIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) SetTrainingJobMaxIdleDurationSeconds(v int32)`

SetTrainingJobMaxIdleDurationSeconds sets TrainingJobMaxIdleDurationSeconds field to given value.

### HasTrainingJobMaxIdleDurationSeconds

`func (o *DataUpdatableFieldsSchedulingRules) HasTrainingJobMaxIdleDurationSeconds() bool`

HasTrainingJobMaxIdleDurationSeconds returns a boolean if a field has been set.

### SetTrainingJobMaxIdleDurationSecondsNil

`func (o *DataUpdatableFieldsSchedulingRules) SetTrainingJobMaxIdleDurationSecondsNil(b bool)`

 SetTrainingJobMaxIdleDurationSecondsNil sets the value for TrainingJobMaxIdleDurationSeconds to be an explicit nil

### UnsetTrainingJobMaxIdleDurationSeconds
`func (o *DataUpdatableFieldsSchedulingRules) UnsetTrainingJobMaxIdleDurationSeconds()`

UnsetTrainingJobMaxIdleDurationSeconds ensures that no value is present for TrainingJobMaxIdleDurationSeconds, not even an explicit nil
### GetTrainingJobTimeLimitSeconds

`func (o *DataUpdatableFieldsSchedulingRules) GetTrainingJobTimeLimitSeconds() int32`

GetTrainingJobTimeLimitSeconds returns the TrainingJobTimeLimitSeconds field if non-nil, zero value otherwise.

### GetTrainingJobTimeLimitSecondsOk

`func (o *DataUpdatableFieldsSchedulingRules) GetTrainingJobTimeLimitSecondsOk() (*int32, bool)`

GetTrainingJobTimeLimitSecondsOk returns a tuple with the TrainingJobTimeLimitSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobTimeLimitSeconds

`func (o *DataUpdatableFieldsSchedulingRules) SetTrainingJobTimeLimitSeconds(v int32)`

SetTrainingJobTimeLimitSeconds sets TrainingJobTimeLimitSeconds field to given value.

### HasTrainingJobTimeLimitSeconds

`func (o *DataUpdatableFieldsSchedulingRules) HasTrainingJobTimeLimitSeconds() bool`

HasTrainingJobTimeLimitSeconds returns a boolean if a field has been set.

### SetTrainingJobTimeLimitSecondsNil

`func (o *DataUpdatableFieldsSchedulingRules) SetTrainingJobTimeLimitSecondsNil(b bool)`

 SetTrainingJobTimeLimitSecondsNil sets the value for TrainingJobTimeLimitSeconds to be an explicit nil

### UnsetTrainingJobTimeLimitSeconds
`func (o *DataUpdatableFieldsSchedulingRules) UnsetTrainingJobTimeLimitSeconds()`

UnsetTrainingJobTimeLimitSeconds ensures that no value is present for TrainingJobTimeLimitSeconds, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


