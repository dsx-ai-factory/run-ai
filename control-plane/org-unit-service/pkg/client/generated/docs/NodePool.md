# NodePool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Uid** | Pointer to **float32** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**ClusterId** | **string** | The id of the cluster. | 
**Operation** | **string** |  | 

## Methods

### NewNodePool

`func NewNodePool(clusterId string, operation string, ) *NodePool`

NewNodePool instantiates a new NodePool object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolWithDefaults

`func NewNodePoolWithDefaults() *NodePool`

NewNodePoolWithDefaults instantiates a new NodePool object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUid

`func (o *NodePool) GetUid() float32`

GetUid returns the Uid field if non-nil, zero value otherwise.

### GetUidOk

`func (o *NodePool) GetUidOk() (*float32, bool)`

GetUidOk returns a tuple with the Uid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUid

`func (o *NodePool) SetUid(v float32)`

SetUid sets Uid field to given value.

### HasUid

`func (o *NodePool) HasUid() bool`

HasUid returns a boolean if a field has been set.

### GetId

`func (o *NodePool) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodePool) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodePool) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *NodePool) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *NodePool) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodePool) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodePool) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *NodePool) HasName() bool`

HasName returns a boolean if a field has been set.

### GetClusterId

`func (o *NodePool) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *NodePool) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *NodePool) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetOperation

`func (o *NodePool) GetOperation() string`

GetOperation returns the Operation field if non-nil, zero value otherwise.

### GetOperationOk

`func (o *NodePool) GetOperationOk() (*string, bool)`

GetOperationOk returns a tuple with the Operation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperation

`func (o *NodePool) SetOperation(v string)`

SetOperation sets Operation field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


