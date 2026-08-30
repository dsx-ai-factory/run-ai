# JobsNodeAffinity

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Train** | Pointer to [**JobsNodeAffinityTrain**](JobsNodeAffinityTrain.md) |  | [optional] 
**Interactive** | Pointer to [**JobsNodeAffinityInteractive**](JobsNodeAffinityInteractive.md) |  | [optional] 

## Methods

### NewJobsNodeAffinity

`func NewJobsNodeAffinity() *JobsNodeAffinity`

NewJobsNodeAffinity instantiates a new JobsNodeAffinity object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobsNodeAffinityWithDefaults

`func NewJobsNodeAffinityWithDefaults() *JobsNodeAffinity`

NewJobsNodeAffinityWithDefaults instantiates a new JobsNodeAffinity object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTrain

`func (o *JobsNodeAffinity) GetTrain() JobsNodeAffinityTrain`

GetTrain returns the Train field if non-nil, zero value otherwise.

### GetTrainOk

`func (o *JobsNodeAffinity) GetTrainOk() (*JobsNodeAffinityTrain, bool)`

GetTrainOk returns a tuple with the Train field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrain

`func (o *JobsNodeAffinity) SetTrain(v JobsNodeAffinityTrain)`

SetTrain sets Train field to given value.

### HasTrain

`func (o *JobsNodeAffinity) HasTrain() bool`

HasTrain returns a boolean if a field has been set.

### GetInteractive

`func (o *JobsNodeAffinity) GetInteractive() JobsNodeAffinityInteractive`

GetInteractive returns the Interactive field if non-nil, zero value otherwise.

### GetInteractiveOk

`func (o *JobsNodeAffinity) GetInteractiveOk() (*JobsNodeAffinityInteractive, bool)`

GetInteractiveOk returns a tuple with the Interactive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractive

`func (o *JobsNodeAffinity) SetInteractive(v JobsNodeAffinityInteractive)`

SetInteractive sets Interactive field to given value.

### HasInteractive

`func (o *JobsNodeAffinity) HasInteractive() bool`

HasInteractive returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


