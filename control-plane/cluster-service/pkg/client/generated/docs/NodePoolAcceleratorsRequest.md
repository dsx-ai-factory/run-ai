# NodePoolAcceleratorsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolIds** | **[]string** | IDs of node pools assigned to the project. | 
**Selection** | Pointer to [**NullableAcceleratorSelection**](AcceleratorSelection.md) |  | [optional] 

## Methods

### NewNodePoolAcceleratorsRequest

`func NewNodePoolAcceleratorsRequest(nodePoolIds []string, ) *NodePoolAcceleratorsRequest`

NewNodePoolAcceleratorsRequest instantiates a new NodePoolAcceleratorsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolAcceleratorsRequestWithDefaults

`func NewNodePoolAcceleratorsRequestWithDefaults() *NodePoolAcceleratorsRequest`

NewNodePoolAcceleratorsRequestWithDefaults instantiates a new NodePoolAcceleratorsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolIds

`func (o *NodePoolAcceleratorsRequest) GetNodePoolIds() []string`

GetNodePoolIds returns the NodePoolIds field if non-nil, zero value otherwise.

### GetNodePoolIdsOk

`func (o *NodePoolAcceleratorsRequest) GetNodePoolIdsOk() (*[]string, bool)`

GetNodePoolIdsOk returns a tuple with the NodePoolIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolIds

`func (o *NodePoolAcceleratorsRequest) SetNodePoolIds(v []string)`

SetNodePoolIds sets NodePoolIds field to given value.


### GetSelection

`func (o *NodePoolAcceleratorsRequest) GetSelection() AcceleratorSelection`

GetSelection returns the Selection field if non-nil, zero value otherwise.

### GetSelectionOk

`func (o *NodePoolAcceleratorsRequest) GetSelectionOk() (*AcceleratorSelection, bool)`

GetSelectionOk returns a tuple with the Selection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelection

`func (o *NodePoolAcceleratorsRequest) SetSelection(v AcceleratorSelection)`

SetSelection sets Selection field to given value.

### HasSelection

`func (o *NodePoolAcceleratorsRequest) HasSelection() bool`

HasSelection returns a boolean if a field has been set.

### SetSelectionNil

`func (o *NodePoolAcceleratorsRequest) SetSelectionNil(b bool)`

 SetSelectionNil sets the value for Selection to be an explicit nil

### UnsetSelection
`func (o *NodePoolAcceleratorsRequest) UnsetSelection()`

UnsetSelection ensures that no value is present for Selection, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


