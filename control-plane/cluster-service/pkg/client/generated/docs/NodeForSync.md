# NodeForSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the node | 
**NodeInfo** | [**NodeInfo**](NodeInfo.md) |  | 

## Methods

### NewNodeForSync

`func NewNodeForSync(name string, nodeInfo NodeInfo, ) *NodeForSync`

NewNodeForSync instantiates a new NodeForSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeForSyncWithDefaults

`func NewNodeForSyncWithDefaults() *NodeForSync`

NewNodeForSyncWithDefaults instantiates a new NodeForSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodeForSync) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodeForSync) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodeForSync) SetName(v string)`

SetName sets Name field to given value.


### GetNodeInfo

`func (o *NodeForSync) GetNodeInfo() NodeInfo`

GetNodeInfo returns the NodeInfo field if non-nil, zero value otherwise.

### GetNodeInfoOk

`func (o *NodeForSync) GetNodeInfoOk() (*NodeInfo, bool)`

GetNodeInfoOk returns a tuple with the NodeInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeInfo

`func (o *NodeForSync) SetNodeInfo(v NodeInfo)`

SetNodeInfo sets NodeInfo field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


