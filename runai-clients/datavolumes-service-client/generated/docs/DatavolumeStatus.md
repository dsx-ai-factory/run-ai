# DatavolumeStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | [**DatavolumePhase**](DatavolumePhase.md) |  | 
**PhaseMessage** | Pointer to **string** | Message explaining the phase of the Datavolume in the cluster | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**DatavolumePvcName** | Pointer to **string** | The name of the copied PVC that is created in the cluster in the project namespace | [optional] 
**DatavolumePvName** | Pointer to **string** | The name of the PV that is created in the cluster - copied from the one that was attached to the original pvc | [optional] 

## Methods

### NewDatavolumeStatus

`func NewDatavolumeStatus(phase DatavolumePhase, ) *DatavolumeStatus`

NewDatavolumeStatus instantiates a new DatavolumeStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeStatusWithDefaults

`func NewDatavolumeStatusWithDefaults() *DatavolumeStatus`

NewDatavolumeStatusWithDefaults instantiates a new DatavolumeStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *DatavolumeStatus) GetPhase() DatavolumePhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *DatavolumeStatus) GetPhaseOk() (*DatavolumePhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *DatavolumeStatus) SetPhase(v DatavolumePhase)`

SetPhase sets Phase field to given value.


### GetPhaseMessage

`func (o *DatavolumeStatus) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *DatavolumeStatus) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *DatavolumeStatus) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *DatavolumeStatus) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetConditions

`func (o *DatavolumeStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *DatavolumeStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *DatavolumeStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *DatavolumeStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetDatavolumePvcName

`func (o *DatavolumeStatus) GetDatavolumePvcName() string`

GetDatavolumePvcName returns the DatavolumePvcName field if non-nil, zero value otherwise.

### GetDatavolumePvcNameOk

`func (o *DatavolumeStatus) GetDatavolumePvcNameOk() (*string, bool)`

GetDatavolumePvcNameOk returns a tuple with the DatavolumePvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatavolumePvcName

`func (o *DatavolumeStatus) SetDatavolumePvcName(v string)`

SetDatavolumePvcName sets DatavolumePvcName field to given value.

### HasDatavolumePvcName

`func (o *DatavolumeStatus) HasDatavolumePvcName() bool`

HasDatavolumePvcName returns a boolean if a field has been set.

### GetDatavolumePvName

`func (o *DatavolumeStatus) GetDatavolumePvName() string`

GetDatavolumePvName returns the DatavolumePvName field if non-nil, zero value otherwise.

### GetDatavolumePvNameOk

`func (o *DatavolumeStatus) GetDatavolumePvNameOk() (*string, bool)`

GetDatavolumePvNameOk returns a tuple with the DatavolumePvName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatavolumePvName

`func (o *DatavolumeStatus) SetDatavolumePvName(v string)`

SetDatavolumePvName sets DatavolumePvName field to given value.

### HasDatavolumePvName

`func (o *DatavolumeStatus) HasDatavolumePvName() bool`

HasDatavolumePvName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


