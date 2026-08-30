# CpuPodsPlacementStrategy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Node** | Pointer to [**NullablePlacementStrategy**](PlacementStrategy.md) | Specifies how CPU pods are placed across nodes.  - binpack - Place as many CPU pods as possible on each node to use fewer nodes and maximize available capacity.  - spread - Distribute CPU pods across as many nodes as possible to balance load across nodes.  | [optional] 

## Methods

### NewCpuPodsPlacementStrategy

`func NewCpuPodsPlacementStrategy() *CpuPodsPlacementStrategy`

NewCpuPodsPlacementStrategy instantiates a new CpuPodsPlacementStrategy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCpuPodsPlacementStrategyWithDefaults

`func NewCpuPodsPlacementStrategyWithDefaults() *CpuPodsPlacementStrategy`

NewCpuPodsPlacementStrategyWithDefaults instantiates a new CpuPodsPlacementStrategy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNode

`func (o *CpuPodsPlacementStrategy) GetNode() PlacementStrategy`

GetNode returns the Node field if non-nil, zero value otherwise.

### GetNodeOk

`func (o *CpuPodsPlacementStrategy) GetNodeOk() (*PlacementStrategy, bool)`

GetNodeOk returns a tuple with the Node field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNode

`func (o *CpuPodsPlacementStrategy) SetNode(v PlacementStrategy)`

SetNode sets Node field to given value.

### HasNode

`func (o *CpuPodsPlacementStrategy) HasNode() bool`

HasNode returns a boolean if a field has been set.

### SetNodeNil

`func (o *CpuPodsPlacementStrategy) SetNodeNil(b bool)`

 SetNodeNil sets the value for Node to be an explicit nil

### UnsetNode
`func (o *CpuPodsPlacementStrategy) UnsetNode()`

UnsetNode ensures that no value is present for Node, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


