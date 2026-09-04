# NodeNumaForSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the physical node. | 
**NumaNodes** | [**[]NumaNodeDetails**](NumaNodeDetails.md) | The complete NUMA-node snapshot for this physical node. An empty array removes its existing NUMA-node details. | 

## Methods

### NewNodeNumaForSync

`func NewNodeNumaForSync(name string, numaNodes []NumaNodeDetails, ) *NodeNumaForSync`

NewNodeNumaForSync instantiates a new NodeNumaForSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeNumaForSyncWithDefaults

`func NewNodeNumaForSyncWithDefaults() *NodeNumaForSync`

NewNodeNumaForSyncWithDefaults instantiates a new NodeNumaForSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodeNumaForSync) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodeNumaForSync) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodeNumaForSync) SetName(v string)`

SetName sets Name field to given value.


### GetNumaNodes

`func (o *NodeNumaForSync) GetNumaNodes() []NumaNodeDetails`

GetNumaNodes returns the NumaNodes field if non-nil, zero value otherwise.

### GetNumaNodesOk

`func (o *NodeNumaForSync) GetNumaNodesOk() (*[]NumaNodeDetails, bool)`

GetNumaNodesOk returns a tuple with the NumaNodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumaNodes

`func (o *NodeNumaForSync) SetNumaNodes(v []NumaNodeDetails)`

SetNumaNodes sets NumaNodes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


