# DatavolumeStatusInternal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | [**DatavolumePhase**](DatavolumePhase.md) |  | 
**PhaseMessage** | Pointer to **string** | Message explaining the phase of the Datavolume in the cluster | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**DatavolumePvcName** | Pointer to **string** | The name of the copied PVC that is created in the cluster in the project namespace | [optional] 
**DatavolumePvName** | Pointer to **string** | The name of the PV that is created in the cluster - copied from the one that was attached to the original pvc | [optional] 
**IsDeleted** | Pointer to **NullableBool** | If true, the datavolume is deleted in the cluster and &#39;phase&#39; field is irrelevant. | [optional] 

## Methods

### NewDatavolumeStatusInternal

`func NewDatavolumeStatusInternal(phase DatavolumePhase, ) *DatavolumeStatusInternal`

NewDatavolumeStatusInternal instantiates a new DatavolumeStatusInternal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeStatusInternalWithDefaults

`func NewDatavolumeStatusInternalWithDefaults() *DatavolumeStatusInternal`

NewDatavolumeStatusInternalWithDefaults instantiates a new DatavolumeStatusInternal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *DatavolumeStatusInternal) GetPhase() DatavolumePhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *DatavolumeStatusInternal) GetPhaseOk() (*DatavolumePhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *DatavolumeStatusInternal) SetPhase(v DatavolumePhase)`

SetPhase sets Phase field to given value.


### GetPhaseMessage

`func (o *DatavolumeStatusInternal) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *DatavolumeStatusInternal) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *DatavolumeStatusInternal) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *DatavolumeStatusInternal) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetConditions

`func (o *DatavolumeStatusInternal) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *DatavolumeStatusInternal) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *DatavolumeStatusInternal) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *DatavolumeStatusInternal) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetDatavolumePvcName

`func (o *DatavolumeStatusInternal) GetDatavolumePvcName() string`

GetDatavolumePvcName returns the DatavolumePvcName field if non-nil, zero value otherwise.

### GetDatavolumePvcNameOk

`func (o *DatavolumeStatusInternal) GetDatavolumePvcNameOk() (*string, bool)`

GetDatavolumePvcNameOk returns a tuple with the DatavolumePvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatavolumePvcName

`func (o *DatavolumeStatusInternal) SetDatavolumePvcName(v string)`

SetDatavolumePvcName sets DatavolumePvcName field to given value.

### HasDatavolumePvcName

`func (o *DatavolumeStatusInternal) HasDatavolumePvcName() bool`

HasDatavolumePvcName returns a boolean if a field has been set.

### GetDatavolumePvName

`func (o *DatavolumeStatusInternal) GetDatavolumePvName() string`

GetDatavolumePvName returns the DatavolumePvName field if non-nil, zero value otherwise.

### GetDatavolumePvNameOk

`func (o *DatavolumeStatusInternal) GetDatavolumePvNameOk() (*string, bool)`

GetDatavolumePvNameOk returns a tuple with the DatavolumePvName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatavolumePvName

`func (o *DatavolumeStatusInternal) SetDatavolumePvName(v string)`

SetDatavolumePvName sets DatavolumePvName field to given value.

### HasDatavolumePvName

`func (o *DatavolumeStatusInternal) HasDatavolumePvName() bool`

HasDatavolumePvName returns a boolean if a field has been set.

### GetIsDeleted

`func (o *DatavolumeStatusInternal) GetIsDeleted() bool`

GetIsDeleted returns the IsDeleted field if non-nil, zero value otherwise.

### GetIsDeletedOk

`func (o *DatavolumeStatusInternal) GetIsDeletedOk() (*bool, bool)`

GetIsDeletedOk returns a tuple with the IsDeleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeleted

`func (o *DatavolumeStatusInternal) SetIsDeleted(v bool)`

SetIsDeleted sets IsDeleted field to given value.

### HasIsDeleted

`func (o *DatavolumeStatusInternal) HasIsDeleted() bool`

HasIsDeleted returns a boolean if a field has been set.

### SetIsDeletedNil

`func (o *DatavolumeStatusInternal) SetIsDeletedNil(b bool)`

 SetIsDeletedNil sets the value for IsDeleted to be an explicit nil

### UnsetIsDeleted
`func (o *DatavolumeStatusInternal) UnsetIsDeleted()`

UnsetIsDeleted ensures that no value is present for IsDeleted, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


