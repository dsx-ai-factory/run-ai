# HistoryRecordSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Event** | Pointer to [**NullableEvent**](Event.md) |  | [optional] 
**PhaseUpdate** | Pointer to [**NullablePhaseUpdate**](PhaseUpdate.md) |  | [optional] 

## Methods

### NewHistoryRecordSpec

`func NewHistoryRecordSpec() *HistoryRecordSpec`

NewHistoryRecordSpec instantiates a new HistoryRecordSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHistoryRecordSpecWithDefaults

`func NewHistoryRecordSpecWithDefaults() *HistoryRecordSpec`

NewHistoryRecordSpecWithDefaults instantiates a new HistoryRecordSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEvent

`func (o *HistoryRecordSpec) GetEvent() Event`

GetEvent returns the Event field if non-nil, zero value otherwise.

### GetEventOk

`func (o *HistoryRecordSpec) GetEventOk() (*Event, bool)`

GetEventOk returns a tuple with the Event field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvent

`func (o *HistoryRecordSpec) SetEvent(v Event)`

SetEvent sets Event field to given value.

### HasEvent

`func (o *HistoryRecordSpec) HasEvent() bool`

HasEvent returns a boolean if a field has been set.

### SetEventNil

`func (o *HistoryRecordSpec) SetEventNil(b bool)`

 SetEventNil sets the value for Event to be an explicit nil

### UnsetEvent
`func (o *HistoryRecordSpec) UnsetEvent()`

UnsetEvent ensures that no value is present for Event, not even an explicit nil
### GetPhaseUpdate

`func (o *HistoryRecordSpec) GetPhaseUpdate() PhaseUpdate`

GetPhaseUpdate returns the PhaseUpdate field if non-nil, zero value otherwise.

### GetPhaseUpdateOk

`func (o *HistoryRecordSpec) GetPhaseUpdateOk() (*PhaseUpdate, bool)`

GetPhaseUpdateOk returns a tuple with the PhaseUpdate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseUpdate

`func (o *HistoryRecordSpec) SetPhaseUpdate(v PhaseUpdate)`

SetPhaseUpdate sets PhaseUpdate field to given value.

### HasPhaseUpdate

`func (o *HistoryRecordSpec) HasPhaseUpdate() bool`

HasPhaseUpdate returns a boolean if a field has been set.

### SetPhaseUpdateNil

`func (o *HistoryRecordSpec) SetPhaseUpdateNil(b bool)`

 SetPhaseUpdateNil sets the value for PhaseUpdate to be an explicit nil

### UnsetPhaseUpdate
`func (o *HistoryRecordSpec) UnsetPhaseUpdate()`

UnsetPhaseUpdate ensures that no value is present for PhaseUpdate, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


