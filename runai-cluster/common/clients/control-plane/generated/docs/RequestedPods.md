# RequestedPods

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Number** | Pointer to **NullableInt32** |  | [optional] 
**Min** | Pointer to **NullableInt32** |  | [optional] 
**Max** | Pointer to **NullableInt32** |  | [optional] 
**Parallelism** | Pointer to **NullableInt32** | specifies how many Pods can run in parallel | [optional] 
**Completions** | Pointer to **NullableInt32** | specifies how many Pods should terminate successfully before the Workload is completed | [optional] 

## Methods

### NewRequestedPods

`func NewRequestedPods() *RequestedPods`

NewRequestedPods instantiates a new RequestedPods object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRequestedPodsWithDefaults

`func NewRequestedPodsWithDefaults() *RequestedPods`

NewRequestedPodsWithDefaults instantiates a new RequestedPods object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumber

`func (o *RequestedPods) GetNumber() int32`

GetNumber returns the Number field if non-nil, zero value otherwise.

### GetNumberOk

`func (o *RequestedPods) GetNumberOk() (*int32, bool)`

GetNumberOk returns a tuple with the Number field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumber

`func (o *RequestedPods) SetNumber(v int32)`

SetNumber sets Number field to given value.

### HasNumber

`func (o *RequestedPods) HasNumber() bool`

HasNumber returns a boolean if a field has been set.

### SetNumberNil

`func (o *RequestedPods) SetNumberNil(b bool)`

 SetNumberNil sets the value for Number to be an explicit nil

### UnsetNumber
`func (o *RequestedPods) UnsetNumber()`

UnsetNumber ensures that no value is present for Number, not even an explicit nil
### GetMin

`func (o *RequestedPods) GetMin() int32`

GetMin returns the Min field if non-nil, zero value otherwise.

### GetMinOk

`func (o *RequestedPods) GetMinOk() (*int32, bool)`

GetMinOk returns a tuple with the Min field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMin

`func (o *RequestedPods) SetMin(v int32)`

SetMin sets Min field to given value.

### HasMin

`func (o *RequestedPods) HasMin() bool`

HasMin returns a boolean if a field has been set.

### SetMinNil

`func (o *RequestedPods) SetMinNil(b bool)`

 SetMinNil sets the value for Min to be an explicit nil

### UnsetMin
`func (o *RequestedPods) UnsetMin()`

UnsetMin ensures that no value is present for Min, not even an explicit nil
### GetMax

`func (o *RequestedPods) GetMax() int32`

GetMax returns the Max field if non-nil, zero value otherwise.

### GetMaxOk

`func (o *RequestedPods) GetMaxOk() (*int32, bool)`

GetMaxOk returns a tuple with the Max field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMax

`func (o *RequestedPods) SetMax(v int32)`

SetMax sets Max field to given value.

### HasMax

`func (o *RequestedPods) HasMax() bool`

HasMax returns a boolean if a field has been set.

### SetMaxNil

`func (o *RequestedPods) SetMaxNil(b bool)`

 SetMaxNil sets the value for Max to be an explicit nil

### UnsetMax
`func (o *RequestedPods) UnsetMax()`

UnsetMax ensures that no value is present for Max, not even an explicit nil
### GetParallelism

`func (o *RequestedPods) GetParallelism() int32`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *RequestedPods) GetParallelismOk() (*int32, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *RequestedPods) SetParallelism(v int32)`

SetParallelism sets Parallelism field to given value.

### HasParallelism

`func (o *RequestedPods) HasParallelism() bool`

HasParallelism returns a boolean if a field has been set.

### SetParallelismNil

`func (o *RequestedPods) SetParallelismNil(b bool)`

 SetParallelismNil sets the value for Parallelism to be an explicit nil

### UnsetParallelism
`func (o *RequestedPods) UnsetParallelism()`

UnsetParallelism ensures that no value is present for Parallelism, not even an explicit nil
### GetCompletions

`func (o *RequestedPods) GetCompletions() int32`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *RequestedPods) GetCompletionsOk() (*int32, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *RequestedPods) SetCompletions(v int32)`

SetCompletions sets Completions field to given value.

### HasCompletions

`func (o *RequestedPods) HasCompletions() bool`

HasCompletions returns a boolean if a field has been set.

### SetCompletionsNil

`func (o *RequestedPods) SetCompletionsNil(b bool)`

 SetCompletionsNil sets the value for Completions to be an explicit nil

### UnsetCompletions
`func (o *RequestedPods) UnsetCompletions()`

UnsetCompletions ensures that no value is present for Completions, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


