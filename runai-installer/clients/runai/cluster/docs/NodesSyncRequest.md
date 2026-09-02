# NodesSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodesForSync** | [**[]NodeForSync**](NodeForSync.md) |  | 
**From** | Pointer to **string** | The first node to sync. Will delete all nodes that don&#39;t exist in the request, after this node. If empty, all nodes before the first one sent will be deleted. | [optional] 
**To** | Pointer to **string** | The last node to sync. Will delete all nodes that don&#39;t exist in the request, before this node. If empty, all nodes after the last one sent will be deleted. | [optional] 

## Methods

### NewNodesSyncRequest

`func NewNodesSyncRequest(nodesForSync []NodeForSync, ) *NodesSyncRequest`

NewNodesSyncRequest instantiates a new NodesSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodesSyncRequestWithDefaults

`func NewNodesSyncRequestWithDefaults() *NodesSyncRequest`

NewNodesSyncRequestWithDefaults instantiates a new NodesSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodesForSync

`func (o *NodesSyncRequest) GetNodesForSync() []NodeForSync`

GetNodesForSync returns the NodesForSync field if non-nil, zero value otherwise.

### GetNodesForSyncOk

`func (o *NodesSyncRequest) GetNodesForSyncOk() (*[]NodeForSync, bool)`

GetNodesForSyncOk returns a tuple with the NodesForSync field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodesForSync

`func (o *NodesSyncRequest) SetNodesForSync(v []NodeForSync)`

SetNodesForSync sets NodesForSync field to given value.


### GetFrom

`func (o *NodesSyncRequest) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *NodesSyncRequest) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *NodesSyncRequest) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *NodesSyncRequest) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *NodesSyncRequest) GetTo() string`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *NodesSyncRequest) GetToOk() (*string, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *NodesSyncRequest) SetTo(v string)`

SetTo sets To field to given value.

### HasTo

`func (o *NodesSyncRequest) HasTo() bool`

HasTo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


