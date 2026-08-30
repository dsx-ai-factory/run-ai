# NetworkTopologySyncGroveCondition

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to **string** | Type of the condition | [optional] 
**Status** | Pointer to **string** | Status of the condition | [optional] 
**Reason** | Pointer to **string** | Indicates the reason for the condition&#39;s last transition | [optional] 
**Message** | Pointer to **string** | A human-readable message describing the condition | [optional] 
**ObservedGeneration** | Pointer to **int64** | Generation of the topology observed when this condition was last updated | [optional] 

## Methods

### NewNetworkTopologySyncGroveCondition

`func NewNetworkTopologySyncGroveCondition() *NetworkTopologySyncGroveCondition`

NewNetworkTopologySyncGroveCondition instantiates a new NetworkTopologySyncGroveCondition object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologySyncGroveConditionWithDefaults

`func NewNetworkTopologySyncGroveConditionWithDefaults() *NetworkTopologySyncGroveCondition`

NewNetworkTopologySyncGroveConditionWithDefaults instantiates a new NetworkTopologySyncGroveCondition object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *NetworkTopologySyncGroveCondition) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *NetworkTopologySyncGroveCondition) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *NetworkTopologySyncGroveCondition) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *NetworkTopologySyncGroveCondition) HasType() bool`

HasType returns a boolean if a field has been set.

### GetStatus

`func (o *NetworkTopologySyncGroveCondition) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NetworkTopologySyncGroveCondition) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NetworkTopologySyncGroveCondition) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NetworkTopologySyncGroveCondition) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetReason

`func (o *NetworkTopologySyncGroveCondition) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *NetworkTopologySyncGroveCondition) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *NetworkTopologySyncGroveCondition) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *NetworkTopologySyncGroveCondition) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetMessage

`func (o *NetworkTopologySyncGroveCondition) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *NetworkTopologySyncGroveCondition) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *NetworkTopologySyncGroveCondition) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *NetworkTopologySyncGroveCondition) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetObservedGeneration

`func (o *NetworkTopologySyncGroveCondition) GetObservedGeneration() int64`

GetObservedGeneration returns the ObservedGeneration field if non-nil, zero value otherwise.

### GetObservedGenerationOk

`func (o *NetworkTopologySyncGroveCondition) GetObservedGenerationOk() (*int64, bool)`

GetObservedGenerationOk returns a tuple with the ObservedGeneration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedGeneration

`func (o *NetworkTopologySyncGroveCondition) SetObservedGeneration(v int64)`

SetObservedGeneration sets ObservedGeneration field to given value.

### HasObservedGeneration

`func (o *NetworkTopologySyncGroveCondition) HasObservedGeneration() bool`

HasObservedGeneration returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


