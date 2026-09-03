# NodePoolStatusRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to **string** | Status of Node Pool | [optional] 
**StatusMessage** | Pointer to **string** | Message for status of Node Pool | [optional] 
**Nodes** | Pointer to **string** | List of Nodes that are assigned to this nodepool - as json string | [optional] 

## Methods

### NewNodePoolStatusRequest

`func NewNodePoolStatusRequest() *NodePoolStatusRequest`

NewNodePoolStatusRequest instantiates a new NodePoolStatusRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolStatusRequestWithDefaults

`func NewNodePoolStatusRequestWithDefaults() *NodePoolStatusRequest`

NewNodePoolStatusRequestWithDefaults instantiates a new NodePoolStatusRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *NodePoolStatusRequest) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NodePoolStatusRequest) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NodePoolStatusRequest) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NodePoolStatusRequest) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetStatusMessage

`func (o *NodePoolStatusRequest) GetStatusMessage() string`

GetStatusMessage returns the StatusMessage field if non-nil, zero value otherwise.

### GetStatusMessageOk

`func (o *NodePoolStatusRequest) GetStatusMessageOk() (*string, bool)`

GetStatusMessageOk returns a tuple with the StatusMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusMessage

`func (o *NodePoolStatusRequest) SetStatusMessage(v string)`

SetStatusMessage sets StatusMessage field to given value.

### HasStatusMessage

`func (o *NodePoolStatusRequest) HasStatusMessage() bool`

HasStatusMessage returns a boolean if a field has been set.

### GetNodes

`func (o *NodePoolStatusRequest) GetNodes() string`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *NodePoolStatusRequest) GetNodesOk() (*string, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *NodePoolStatusRequest) SetNodes(v string)`

SetNodes sets Nodes field to given value.

### HasNodes

`func (o *NodePoolStatusRequest) HasNodes() bool`

HasNodes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


