# NodePoolResourcesPatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OrgUnitId** | **string** | The project/department ID. | 
**NodePoolId** | **string** |  | 
**Rank** | Pointer to **NullableString** |  | [optional] 
**Resources** | Pointer to [**NullableResourcesPerNodePoolNullable**](ResourcesPerNodePoolNullable.md) |  | [optional] 

## Methods

### NewNodePoolResourcesPatchRequest

`func NewNodePoolResourcesPatchRequest(orgUnitId string, nodePoolId string, ) *NodePoolResourcesPatchRequest`

NewNodePoolResourcesPatchRequest instantiates a new NodePoolResourcesPatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolResourcesPatchRequestWithDefaults

`func NewNodePoolResourcesPatchRequestWithDefaults() *NodePoolResourcesPatchRequest`

NewNodePoolResourcesPatchRequestWithDefaults instantiates a new NodePoolResourcesPatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOrgUnitId

`func (o *NodePoolResourcesPatchRequest) GetOrgUnitId() string`

GetOrgUnitId returns the OrgUnitId field if non-nil, zero value otherwise.

### GetOrgUnitIdOk

`func (o *NodePoolResourcesPatchRequest) GetOrgUnitIdOk() (*string, bool)`

GetOrgUnitIdOk returns a tuple with the OrgUnitId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnitId

`func (o *NodePoolResourcesPatchRequest) SetOrgUnitId(v string)`

SetOrgUnitId sets OrgUnitId field to given value.


### GetNodePoolId

`func (o *NodePoolResourcesPatchRequest) GetNodePoolId() string`

GetNodePoolId returns the NodePoolId field if non-nil, zero value otherwise.

### GetNodePoolIdOk

`func (o *NodePoolResourcesPatchRequest) GetNodePoolIdOk() (*string, bool)`

GetNodePoolIdOk returns a tuple with the NodePoolId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolId

`func (o *NodePoolResourcesPatchRequest) SetNodePoolId(v string)`

SetNodePoolId sets NodePoolId field to given value.


### GetRank

`func (o *NodePoolResourcesPatchRequest) GetRank() string`

GetRank returns the Rank field if non-nil, zero value otherwise.

### GetRankOk

`func (o *NodePoolResourcesPatchRequest) GetRankOk() (*string, bool)`

GetRankOk returns a tuple with the Rank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRank

`func (o *NodePoolResourcesPatchRequest) SetRank(v string)`

SetRank sets Rank field to given value.

### HasRank

`func (o *NodePoolResourcesPatchRequest) HasRank() bool`

HasRank returns a boolean if a field has been set.

### SetRankNil

`func (o *NodePoolResourcesPatchRequest) SetRankNil(b bool)`

 SetRankNil sets the value for Rank to be an explicit nil

### UnsetRank
`func (o *NodePoolResourcesPatchRequest) UnsetRank()`

UnsetRank ensures that no value is present for Rank, not even an explicit nil
### GetResources

`func (o *NodePoolResourcesPatchRequest) GetResources() ResourcesPerNodePoolNullable`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *NodePoolResourcesPatchRequest) GetResourcesOk() (*ResourcesPerNodePoolNullable, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *NodePoolResourcesPatchRequest) SetResources(v ResourcesPerNodePoolNullable)`

SetResources sets Resources field to given value.

### HasResources

`func (o *NodePoolResourcesPatchRequest) HasResources() bool`

HasResources returns a boolean if a field has been set.

### SetResourcesNil

`func (o *NodePoolResourcesPatchRequest) SetResourcesNil(b bool)`

 SetResourcesNil sets the value for Resources to be an explicit nil

### UnsetResources
`func (o *NodePoolResourcesPatchRequest) UnsetResources()`

UnsetResources ensures that no value is present for Resources, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


