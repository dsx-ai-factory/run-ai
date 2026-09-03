# NodeInternal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The node name | 
**ClusterUuid** | **string** | The id of the cluster. | 
**NetworkTopologyLabelValues** | Pointer to **[]string** | The topology levels of node (node address) according to nodepool assigned topology | [optional] 

## Methods

### NewNodeInternal

`func NewNodeInternal(name string, clusterUuid string, ) *NodeInternal`

NewNodeInternal instantiates a new NodeInternal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeInternalWithDefaults

`func NewNodeInternalWithDefaults() *NodeInternal`

NewNodeInternalWithDefaults instantiates a new NodeInternal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodeInternal) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodeInternal) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodeInternal) SetName(v string)`

SetName sets Name field to given value.


### GetClusterUuid

`func (o *NodeInternal) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *NodeInternal) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *NodeInternal) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.


### GetNetworkTopologyLabelValues

`func (o *NodeInternal) GetNetworkTopologyLabelValues() []string`

GetNetworkTopologyLabelValues returns the NetworkTopologyLabelValues field if non-nil, zero value otherwise.

### GetNetworkTopologyLabelValuesOk

`func (o *NodeInternal) GetNetworkTopologyLabelValuesOk() (*[]string, bool)`

GetNetworkTopologyLabelValuesOk returns a tuple with the NetworkTopologyLabelValues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNetworkTopologyLabelValues

`func (o *NodeInternal) SetNetworkTopologyLabelValues(v []string)`

SetNetworkTopologyLabelValues sets NetworkTopologyLabelValues field to given value.

### HasNetworkTopologyLabelValues

`func (o *NodeInternal) HasNetworkTopologyLabelValues() bool`

HasNetworkTopologyLabelValues returns a boolean if a field has been set.

### SetNetworkTopologyLabelValuesNil

`func (o *NodeInternal) SetNetworkTopologyLabelValuesNil(b bool)`

 SetNetworkTopologyLabelValuesNil sets the value for NetworkTopologyLabelValues to be an explicit nil

### UnsetNetworkTopologyLabelValues
`func (o *NodeInternal) UnsetNetworkTopologyLabelValues()`

UnsetNetworkTopologyLabelValues ensures that no value is present for NetworkTopologyLabelValues, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


