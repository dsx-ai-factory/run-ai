# NodePoolResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **float32** | Node pool id. | 
**Name** | Pointer to **string** | Node pool name. | [optional] 

## Methods

### NewNodePoolResources

`func NewNodePoolResources(id float32, ) *NodePoolResources`

NewNodePoolResources instantiates a new NodePoolResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolResourcesWithDefaults

`func NewNodePoolResourcesWithDefaults() *NodePoolResources`

NewNodePoolResourcesWithDefaults instantiates a new NodePoolResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NodePoolResources) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodePoolResources) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodePoolResources) SetId(v float32)`

SetId sets Id field to given value.


### GetName

`func (o *NodePoolResources) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodePoolResources) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodePoolResources) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *NodePoolResources) HasName() bool`

HasName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


