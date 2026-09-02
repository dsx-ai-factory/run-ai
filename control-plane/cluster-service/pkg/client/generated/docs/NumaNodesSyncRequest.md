# NumaNodesSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Nodes** | [**[]NodeNumaForSync**](NodeNumaForSync.md) | Physical nodes and their complete NUMA-node snapshots. | 
**DeleteMissing** | Pointer to **bool** | When true, delete NUMA-node details for physical nodes in the requested range that are missing from nodes. Used for full-cluster reconciliation. | [optional] [default to false]
**From** | Pointer to **string** | The inclusive lower physical-node-name boundary for deletion when deleteMissing is true. If omitted, the range is unbounded below. | [optional] 
**To** | Pointer to **string** | The inclusive upper physical-node-name boundary for deletion when deleteMissing is true. If omitted, the range is unbounded above. | [optional] 

## Methods

### NewNumaNodesSyncRequest

`func NewNumaNodesSyncRequest(nodes []NodeNumaForSync, ) *NumaNodesSyncRequest`

NewNumaNodesSyncRequest instantiates a new NumaNodesSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNumaNodesSyncRequestWithDefaults

`func NewNumaNodesSyncRequestWithDefaults() *NumaNodesSyncRequest`

NewNumaNodesSyncRequestWithDefaults instantiates a new NumaNodesSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodes

`func (o *NumaNodesSyncRequest) GetNodes() []NodeNumaForSync`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *NumaNodesSyncRequest) GetNodesOk() (*[]NodeNumaForSync, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *NumaNodesSyncRequest) SetNodes(v []NodeNumaForSync)`

SetNodes sets Nodes field to given value.


### GetDeleteMissing

`func (o *NumaNodesSyncRequest) GetDeleteMissing() bool`

GetDeleteMissing returns the DeleteMissing field if non-nil, zero value otherwise.

### GetDeleteMissingOk

`func (o *NumaNodesSyncRequest) GetDeleteMissingOk() (*bool, bool)`

GetDeleteMissingOk returns a tuple with the DeleteMissing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleteMissing

`func (o *NumaNodesSyncRequest) SetDeleteMissing(v bool)`

SetDeleteMissing sets DeleteMissing field to given value.

### HasDeleteMissing

`func (o *NumaNodesSyncRequest) HasDeleteMissing() bool`

HasDeleteMissing returns a boolean if a field has been set.

### GetFrom

`func (o *NumaNodesSyncRequest) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *NumaNodesSyncRequest) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *NumaNodesSyncRequest) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *NumaNodesSyncRequest) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *NumaNodesSyncRequest) GetTo() string`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *NumaNodesSyncRequest) GetToOk() (*string, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *NumaNodesSyncRequest) SetTo(v string)`

SetTo sets To field to given value.

### HasTo

`func (o *NumaNodesSyncRequest) HasTo() bool`

HasTo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


