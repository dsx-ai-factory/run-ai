# NodepoolConditionDetails

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | Type of Nodepool condition. | 
**Reason** | **string** | (brief) reason for the condition&#39;s last transition. | 
**Message** | Pointer to **string** | Human readable message indicating details about last transition. | [optional] 
**Status** | Pointer to **string** | The status of the condition, can be True or False | [optional] 
**Nodes** | Pointer to **[]string** | Names of the nodes related to the nodepool condition | [optional] 

## Methods

### NewNodepoolConditionDetails

`func NewNodepoolConditionDetails(type_ string, reason string, ) *NodepoolConditionDetails`

NewNodepoolConditionDetails instantiates a new NodepoolConditionDetails object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodepoolConditionDetailsWithDefaults

`func NewNodepoolConditionDetailsWithDefaults() *NodepoolConditionDetails`

NewNodepoolConditionDetailsWithDefaults instantiates a new NodepoolConditionDetails object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *NodepoolConditionDetails) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *NodepoolConditionDetails) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *NodepoolConditionDetails) SetType(v string)`

SetType sets Type field to given value.


### GetReason

`func (o *NodepoolConditionDetails) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *NodepoolConditionDetails) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *NodepoolConditionDetails) SetReason(v string)`

SetReason sets Reason field to given value.


### GetMessage

`func (o *NodepoolConditionDetails) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *NodepoolConditionDetails) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *NodepoolConditionDetails) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *NodepoolConditionDetails) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetStatus

`func (o *NodepoolConditionDetails) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NodepoolConditionDetails) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NodepoolConditionDetails) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NodepoolConditionDetails) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetNodes

`func (o *NodepoolConditionDetails) GetNodes() []string`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *NodepoolConditionDetails) GetNodesOk() (*[]string, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *NodepoolConditionDetails) SetNodes(v []string)`

SetNodes sets Nodes field to given value.

### HasNodes

`func (o *NodepoolConditionDetails) HasNodes() bool`

HasNodes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


