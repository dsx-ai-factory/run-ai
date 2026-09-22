# WorkloadPatchFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | [**Phase**](Phase.md) |  | 
**PhaseMessage** | Pointer to **string** |  | [optional] 
**SoftDelete** | Pointer to **bool** |  | [optional] 

## Methods

### NewWorkloadPatchFields

`func NewWorkloadPatchFields(phase Phase, ) *WorkloadPatchFields`

NewWorkloadPatchFields instantiates a new WorkloadPatchFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadPatchFieldsWithDefaults

`func NewWorkloadPatchFieldsWithDefaults() *WorkloadPatchFields`

NewWorkloadPatchFieldsWithDefaults instantiates a new WorkloadPatchFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *WorkloadPatchFields) GetPhase() Phase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *WorkloadPatchFields) GetPhaseOk() (*Phase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *WorkloadPatchFields) SetPhase(v Phase)`

SetPhase sets Phase field to given value.


### GetPhaseMessage

`func (o *WorkloadPatchFields) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *WorkloadPatchFields) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *WorkloadPatchFields) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *WorkloadPatchFields) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetSoftDelete

`func (o *WorkloadPatchFields) GetSoftDelete() bool`

GetSoftDelete returns the SoftDelete field if non-nil, zero value otherwise.

### GetSoftDeleteOk

`func (o *WorkloadPatchFields) GetSoftDeleteOk() (*bool, bool)`

GetSoftDeleteOk returns a tuple with the SoftDelete field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSoftDelete

`func (o *WorkloadPatchFields) SetSoftDelete(v bool)`

SetSoftDelete sets SoftDelete field to given value.

### HasSoftDelete

`func (o *WorkloadPatchFields) HasSoftDelete() bool`

HasSoftDelete returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


