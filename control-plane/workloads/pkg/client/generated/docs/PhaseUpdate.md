# PhaseUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | [**Phase**](Phase.md) |  | 
**PhaseMessage** | **string** |  | 

## Methods

### NewPhaseUpdate

`func NewPhaseUpdate(phase Phase, phaseMessage string, ) *PhaseUpdate`

NewPhaseUpdate instantiates a new PhaseUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPhaseUpdateWithDefaults

`func NewPhaseUpdateWithDefaults() *PhaseUpdate`

NewPhaseUpdateWithDefaults instantiates a new PhaseUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *PhaseUpdate) GetPhase() Phase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PhaseUpdate) GetPhaseOk() (*Phase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PhaseUpdate) SetPhase(v Phase)`

SetPhase sets Phase field to given value.


### GetPhaseMessage

`func (o *PhaseUpdate) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *PhaseUpdate) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *PhaseUpdate) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


