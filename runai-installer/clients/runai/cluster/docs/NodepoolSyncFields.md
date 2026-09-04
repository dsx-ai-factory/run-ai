# NodepoolSyncFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Phase** | Pointer to [**NodepoolPhase**](NodepoolPhase.md) |  | [optional] 
**PhaseMessage** | Pointer to **string** | Message for status of Node Pool | [optional] 
**Status** | Pointer to [**NodepoolSyncFieldsStatus**](NodepoolSyncFieldsStatus.md) |  | [optional] 

## Methods

### NewNodepoolSyncFields

`func NewNodepoolSyncFields(name string, ) *NodepoolSyncFields`

NewNodepoolSyncFields instantiates a new NodepoolSyncFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolSyncFieldsWithDefaults

`func NewNodepoolSyncFieldsWithDefaults() *NodepoolSyncFields`

NewNodepoolSyncFieldsWithDefaults instantiates a new NodepoolSyncFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodepoolSyncFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodepoolSyncFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodepoolSyncFields) SetName(v string)`

SetName sets Name field to given value.


### GetPhase

`func (o *NodepoolSyncFields) GetPhase() NodepoolPhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *NodepoolSyncFields) GetPhaseOk() (*NodepoolPhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *NodepoolSyncFields) SetPhase(v NodepoolPhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *NodepoolSyncFields) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetPhaseMessage

`func (o *NodepoolSyncFields) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *NodepoolSyncFields) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *NodepoolSyncFields) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *NodepoolSyncFields) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetStatus

`func (o *NodepoolSyncFields) GetStatus() NodepoolSyncFieldsStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NodepoolSyncFields) GetStatusOk() (*NodepoolSyncFieldsStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NodepoolSyncFields) SetStatus(v NodepoolSyncFieldsStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NodepoolSyncFields) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


