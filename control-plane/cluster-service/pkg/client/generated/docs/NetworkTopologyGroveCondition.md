# NetworkTopologyGroveCondition

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to **string** | Type of the condition | [optional] 
**Status** | Pointer to **string** | Status of the condition | [optional] 
**Reason** | Pointer to **string** | Indicates the reason for the condition&#39;s last transition | [optional] 
**Message** | Pointer to **string** | A human-readable message describing the condition | [optional] 
**ObservedGeneration** | Pointer to **int64** | Generation of the topology observed when this condition was last updated | [optional] 

## Methods

### NewNetworkTopologyGroveCondition

`func NewNetworkTopologyGroveCondition() *NetworkTopologyGroveCondition`

NewNetworkTopologyGroveCondition instantiates a new NetworkTopologyGroveCondition object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologyGroveConditionWithDefaults

`func NewNetworkTopologyGroveConditionWithDefaults() *NetworkTopologyGroveCondition`

NewNetworkTopologyGroveConditionWithDefaults instantiates a new NetworkTopologyGroveCondition object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *NetworkTopologyGroveCondition) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *NetworkTopologyGroveCondition) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *NetworkTopologyGroveCondition) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *NetworkTopologyGroveCondition) HasType() bool`

HasType returns a boolean if a field has been set.

### GetStatus

`func (o *NetworkTopologyGroveCondition) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NetworkTopologyGroveCondition) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NetworkTopologyGroveCondition) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NetworkTopologyGroveCondition) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetReason

`func (o *NetworkTopologyGroveCondition) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *NetworkTopologyGroveCondition) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *NetworkTopologyGroveCondition) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *NetworkTopologyGroveCondition) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetMessage

`func (o *NetworkTopologyGroveCondition) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *NetworkTopologyGroveCondition) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *NetworkTopologyGroveCondition) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *NetworkTopologyGroveCondition) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetObservedGeneration

`func (o *NetworkTopologyGroveCondition) GetObservedGeneration() int64`

GetObservedGeneration returns the ObservedGeneration field if non-nil, zero value otherwise.

### GetObservedGenerationOk

`func (o *NetworkTopologyGroveCondition) GetObservedGenerationOk() (*int64, bool)`

GetObservedGenerationOk returns a tuple with the ObservedGeneration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedGeneration

`func (o *NetworkTopologyGroveCondition) SetObservedGeneration(v int64)`

SetObservedGeneration sets ObservedGeneration field to given value.

### HasObservedGeneration

`func (o *NetworkTopologyGroveCondition) HasObservedGeneration() bool`

HasObservedGeneration returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


