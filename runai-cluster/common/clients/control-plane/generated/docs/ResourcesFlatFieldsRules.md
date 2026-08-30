# ResourcesFlatFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodeType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**NodePools** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinityRules**](PodAffinityRules.md) |  | [optional] 

## Methods

### NewResourcesFlatFieldsRules

`func NewResourcesFlatFieldsRules() *ResourcesFlatFieldsRules`

NewResourcesFlatFieldsRules instantiates a new ResourcesFlatFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesFlatFieldsRulesWithDefaults

`func NewResourcesFlatFieldsRulesWithDefaults() *ResourcesFlatFieldsRules`

NewResourcesFlatFieldsRulesWithDefaults instantiates a new ResourcesFlatFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodeType

`func (o *ResourcesFlatFieldsRules) GetNodeType() StringRules`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *ResourcesFlatFieldsRules) GetNodeTypeOk() (*StringRules, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *ResourcesFlatFieldsRules) SetNodeType(v StringRules)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *ResourcesFlatFieldsRules) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *ResourcesFlatFieldsRules) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *ResourcesFlatFieldsRules) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *ResourcesFlatFieldsRules) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *ResourcesFlatFieldsRules) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *ResourcesFlatFieldsRules) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *ResourcesFlatFieldsRules) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *ResourcesFlatFieldsRules) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *ResourcesFlatFieldsRules) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *ResourcesFlatFieldsRules) GetPodAffinity() PodAffinityRules`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *ResourcesFlatFieldsRules) GetPodAffinityOk() (*PodAffinityRules, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *ResourcesFlatFieldsRules) SetPodAffinity(v PodAffinityRules)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *ResourcesFlatFieldsRules) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *ResourcesFlatFieldsRules) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *ResourcesFlatFieldsRules) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


