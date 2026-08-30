# NodeStatusConditionDetails

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | Type of node condition. | 
**Reason** | **string** | (brief) reason for the condition&#39;s last transition. | 
**Message** | Pointer to **string** | Human readable message indicating details about last transition. | [optional] 

## Methods

### NewNodeStatusConditionDetails

`func NewNodeStatusConditionDetails(type_ string, reason string, ) *NodeStatusConditionDetails`

NewNodeStatusConditionDetails instantiates a new NodeStatusConditionDetails object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeStatusConditionDetailsWithDefaults

`func NewNodeStatusConditionDetailsWithDefaults() *NodeStatusConditionDetails`

NewNodeStatusConditionDetailsWithDefaults instantiates a new NodeStatusConditionDetails object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *NodeStatusConditionDetails) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *NodeStatusConditionDetails) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *NodeStatusConditionDetails) SetType(v string)`

SetType sets Type field to given value.


### GetReason

`func (o *NodeStatusConditionDetails) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *NodeStatusConditionDetails) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *NodeStatusConditionDetails) SetReason(v string)`

SetReason sets Reason field to given value.


### GetMessage

`func (o *NodeStatusConditionDetails) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *NodeStatusConditionDetails) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *NodeStatusConditionDetails) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *NodeStatusConditionDetails) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


