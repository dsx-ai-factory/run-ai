# WorkloadHistorySpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadEvent** | Pointer to [**WorkloadEvent**](WorkloadEvent.md) |  | [optional] 
**StatusChange** | Pointer to [**StatusChange**](StatusChange.md) |  | [optional] 

## Methods

### NewWorkloadHistorySpec

`func NewWorkloadHistorySpec() *WorkloadHistorySpec`

NewWorkloadHistorySpec instantiates a new WorkloadHistorySpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadHistorySpecWithDefaults

`func NewWorkloadHistorySpecWithDefaults() *WorkloadHistorySpec`

NewWorkloadHistorySpecWithDefaults instantiates a new WorkloadHistorySpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadEvent

`func (o *WorkloadHistorySpec) GetWorkloadEvent() WorkloadEvent`

GetWorkloadEvent returns the WorkloadEvent field if non-nil, zero value otherwise.

### GetWorkloadEventOk

`func (o *WorkloadHistorySpec) GetWorkloadEventOk() (*WorkloadEvent, bool)`

GetWorkloadEventOk returns a tuple with the WorkloadEvent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadEvent

`func (o *WorkloadHistorySpec) SetWorkloadEvent(v WorkloadEvent)`

SetWorkloadEvent sets WorkloadEvent field to given value.

### HasWorkloadEvent

`func (o *WorkloadHistorySpec) HasWorkloadEvent() bool`

HasWorkloadEvent returns a boolean if a field has been set.

### GetStatusChange

`func (o *WorkloadHistorySpec) GetStatusChange() StatusChange`

GetStatusChange returns the StatusChange field if non-nil, zero value otherwise.

### GetStatusChangeOk

`func (o *WorkloadHistorySpec) GetStatusChangeOk() (*StatusChange, bool)`

GetStatusChangeOk returns a tuple with the StatusChange field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusChange

`func (o *WorkloadHistorySpec) SetStatusChange(v StatusChange)`

SetStatusChange sets StatusChange field to given value.

### HasStatusChange

`func (o *WorkloadHistorySpec) HasStatusChange() bool`

HasStatusChange returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


