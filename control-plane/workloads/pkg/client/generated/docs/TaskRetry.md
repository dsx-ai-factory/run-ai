# TaskRetry

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Limit** | Pointer to **NullableInt32** | The maximum number of retries. The task is retried when it ends with one of &#x60;onExitCodes&#x60;, when it&#39;s preempted, and after an infrastructure error such as a lost node. 0 means that the task isn&#39;t retried, and fails on its first failure or preemption. | [optional] 
**OnExitCodes** | Pointer to [**[]TaskExitCodeRange**](TaskExitCodeRange.md) | Ranges of exit codes that retry the task. | [optional] 

## Methods

### NewTaskRetry

`func NewTaskRetry() *TaskRetry`

NewTaskRetry instantiates a new TaskRetry object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskRetryWithDefaults

`func NewTaskRetryWithDefaults() *TaskRetry`

NewTaskRetryWithDefaults instantiates a new TaskRetry object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLimit

`func (o *TaskRetry) GetLimit() int32`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *TaskRetry) GetLimitOk() (*int32, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *TaskRetry) SetLimit(v int32)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *TaskRetry) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### SetLimitNil

`func (o *TaskRetry) SetLimitNil(b bool)`

 SetLimitNil sets the value for Limit to be an explicit nil

### UnsetLimit
`func (o *TaskRetry) UnsetLimit()`

UnsetLimit ensures that no value is present for Limit, not even an explicit nil
### GetOnExitCodes

`func (o *TaskRetry) GetOnExitCodes() []TaskExitCodeRange`

GetOnExitCodes returns the OnExitCodes field if non-nil, zero value otherwise.

### GetOnExitCodesOk

`func (o *TaskRetry) GetOnExitCodesOk() (*[]TaskExitCodeRange, bool)`

GetOnExitCodesOk returns a tuple with the OnExitCodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnExitCodes

`func (o *TaskRetry) SetOnExitCodes(v []TaskExitCodeRange)`

SetOnExitCodes sets OnExitCodes field to given value.

### HasOnExitCodes

`func (o *TaskRetry) HasOnExitCodes() bool`

HasOnExitCodes returns a boolean if a field has been set.

### SetOnExitCodesNil

`func (o *TaskRetry) SetOnExitCodesNil(b bool)`

 SetOnExitCodesNil sets the value for OnExitCodes to be an explicit nil

### UnsetOnExitCodes
`func (o *TaskRetry) UnsetOnExitCodes()`

UnsetOnExitCodes ensures that no value is present for OnExitCodes, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


