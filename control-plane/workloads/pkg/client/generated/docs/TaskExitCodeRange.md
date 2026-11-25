# TaskExitCodeRange

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**From** | Pointer to **NullableInt32** | The first exit code of the range. (mandatory) | [optional] 
**To** | Pointer to **NullableInt32** | The last exit code of the range. (mandatory) | [optional] 

## Methods

### NewTaskExitCodeRange

`func NewTaskExitCodeRange() *TaskExitCodeRange`

NewTaskExitCodeRange instantiates a new TaskExitCodeRange object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTaskExitCodeRangeWithDefaults

`func NewTaskExitCodeRangeWithDefaults() *TaskExitCodeRange`

NewTaskExitCodeRangeWithDefaults instantiates a new TaskExitCodeRange object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFrom

`func (o *TaskExitCodeRange) GetFrom() int32`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *TaskExitCodeRange) GetFromOk() (*int32, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *TaskExitCodeRange) SetFrom(v int32)`

SetFrom sets From field to given value.

### HasFrom

`func (o *TaskExitCodeRange) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### SetFromNil

`func (o *TaskExitCodeRange) SetFromNil(b bool)`

 SetFromNil sets the value for From to be an explicit nil

### UnsetFrom
`func (o *TaskExitCodeRange) UnsetFrom()`

UnsetFrom ensures that no value is present for From, not even an explicit nil
### GetTo

`func (o *TaskExitCodeRange) GetTo() int32`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *TaskExitCodeRange) GetToOk() (*int32, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *TaskExitCodeRange) SetTo(v int32)`

SetTo sets To field to given value.

### HasTo

`func (o *TaskExitCodeRange) HasTo() bool`

HasTo returns a boolean if a field has been set.

### SetToNil

`func (o *TaskExitCodeRange) SetToNil(b bool)`

 SetToNil sets the value for To to be an explicit nil

### UnsetTo
`func (o *TaskExitCodeRange) UnsetTo()`

UnsetTo ensures that no value is present for To, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


