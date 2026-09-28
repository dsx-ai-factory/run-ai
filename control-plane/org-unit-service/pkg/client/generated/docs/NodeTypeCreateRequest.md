# NodeTypeCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**ClusterId** | Pointer to **string** |  | [optional] 

## Methods

### NewNodeTypeCreateRequest

`func NewNodeTypeCreateRequest() *NodeTypeCreateRequest`

NewNodeTypeCreateRequest instantiates a new NodeTypeCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeTypeCreateRequestWithDefaults

`func NewNodeTypeCreateRequestWithDefaults() *NodeTypeCreateRequest`

NewNodeTypeCreateRequestWithDefaults instantiates a new NodeTypeCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodeTypeCreateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodeTypeCreateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodeTypeCreateRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *NodeTypeCreateRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetClusterId

`func (o *NodeTypeCreateRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *NodeTypeCreateRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *NodeTypeCreateRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *NodeTypeCreateRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


