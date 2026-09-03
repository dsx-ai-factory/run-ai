# NimServiceMultiNode

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Workers** | Pointer to **NullableInt32** | Specifies the number of worker nodes to use when running the NIM service in multi-node. | [optional] 

## Methods

### NewNimServiceMultiNode

`func NewNimServiceMultiNode() *NimServiceMultiNode`

NewNimServiceMultiNode instantiates a new NimServiceMultiNode object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceMultiNodeWithDefaults

`func NewNimServiceMultiNodeWithDefaults() *NimServiceMultiNode`

NewNimServiceMultiNodeWithDefaults instantiates a new NimServiceMultiNode object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkers

`func (o *NimServiceMultiNode) GetWorkers() int32`

GetWorkers returns the Workers field if non-nil, zero value otherwise.

### GetWorkersOk

`func (o *NimServiceMultiNode) GetWorkersOk() (*int32, bool)`

GetWorkersOk returns a tuple with the Workers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkers

`func (o *NimServiceMultiNode) SetWorkers(v int32)`

SetWorkers sets Workers field to given value.

### HasWorkers

`func (o *NimServiceMultiNode) HasWorkers() bool`

HasWorkers returns a boolean if a field has been set.

### SetWorkersNil

`func (o *NimServiceMultiNode) SetWorkersNil(b bool)`

 SetWorkersNil sets the value for Workers to be an explicit nil

### UnsetWorkers
`func (o *NimServiceMultiNode) UnsetWorkers()`

UnsetWorkers ensures that no value is present for Workers, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


