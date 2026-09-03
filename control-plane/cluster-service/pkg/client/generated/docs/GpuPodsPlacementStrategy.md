# GpuPodsPlacementStrategy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Node** | Pointer to [**NullablePlacementStrategy**](PlacementStrategy.md) | Specifies how GPU pods are placed across nodes.  - binpack - Place as many GPU pods as possible on each node to use fewer nodes and maximize available capacity.  - spread - Distribute GPU pods across as many nodes as possible to balance load across nodes.  | [optional] 
**Device** | Pointer to [**NullablePlacementStrategy**](PlacementStrategy.md) | Specifies how GPU pods are placed across GPU devices within a node when using GPU fractions.  - binpack - Place as many GPU pods as possible on each GPU device to use fewer devices and maximize available capacity, minimizing GPU fraction fragmentation.  - spread - Distribute GPU pods across as many GPU devices as possible to balance load across devices.  | [optional] 

## Methods

### NewGpuPodsPlacementStrategy

`func NewGpuPodsPlacementStrategy() *GpuPodsPlacementStrategy`

NewGpuPodsPlacementStrategy instantiates a new GpuPodsPlacementStrategy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGpuPodsPlacementStrategyWithDefaults

`func NewGpuPodsPlacementStrategyWithDefaults() *GpuPodsPlacementStrategy`

NewGpuPodsPlacementStrategyWithDefaults instantiates a new GpuPodsPlacementStrategy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNode

`func (o *GpuPodsPlacementStrategy) GetNode() PlacementStrategy`

GetNode returns the Node field if non-nil, zero value otherwise.

### GetNodeOk

`func (o *GpuPodsPlacementStrategy) GetNodeOk() (*PlacementStrategy, bool)`

GetNodeOk returns a tuple with the Node field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNode

`func (o *GpuPodsPlacementStrategy) SetNode(v PlacementStrategy)`

SetNode sets Node field to given value.

### HasNode

`func (o *GpuPodsPlacementStrategy) HasNode() bool`

HasNode returns a boolean if a field has been set.

### SetNodeNil

`func (o *GpuPodsPlacementStrategy) SetNodeNil(b bool)`

 SetNodeNil sets the value for Node to be an explicit nil

### UnsetNode
`func (o *GpuPodsPlacementStrategy) UnsetNode()`

UnsetNode ensures that no value is present for Node, not even an explicit nil
### GetDevice

`func (o *GpuPodsPlacementStrategy) GetDevice() PlacementStrategy`

GetDevice returns the Device field if non-nil, zero value otherwise.

### GetDeviceOk

`func (o *GpuPodsPlacementStrategy) GetDeviceOk() (*PlacementStrategy, bool)`

GetDeviceOk returns a tuple with the Device field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDevice

`func (o *GpuPodsPlacementStrategy) SetDevice(v PlacementStrategy)`

SetDevice sets Device field to given value.

### HasDevice

`func (o *GpuPodsPlacementStrategy) HasDevice() bool`

HasDevice returns a boolean if a field has been set.

### SetDeviceNil

`func (o *GpuPodsPlacementStrategy) SetDeviceNil(b bool)`

 SetDeviceNil sets the value for Device to be an explicit nil

### UnsetDevice
`func (o *GpuPodsPlacementStrategy) UnsetDevice()`

UnsetDevice ensures that no value is present for Device, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


