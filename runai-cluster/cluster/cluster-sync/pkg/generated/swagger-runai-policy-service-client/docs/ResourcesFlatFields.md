# ResourcesFlatFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodeType** | Pointer to **NullableString** | Nodes (machines), or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes. For more information, see [Group Nodes](https://docs.run.ai/latest/admin/researcher-setup/limit-to-node-group). When using this flag with with Project-based affinity, it refines the list of allowable node groups set in the Project. For more information, see [Projects](https://docs.run.ai/admin/admin-ui-setup/project-setup). | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workspace on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinity**](PodAffinity.md) |  | [optional] 

## Methods

### NewResourcesFlatFields

`func NewResourcesFlatFields() *ResourcesFlatFields`

NewResourcesFlatFields instantiates a new ResourcesFlatFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesFlatFieldsWithDefaults

`func NewResourcesFlatFieldsWithDefaults() *ResourcesFlatFields`

NewResourcesFlatFieldsWithDefaults instantiates a new ResourcesFlatFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodeType

`func (o *ResourcesFlatFields) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *ResourcesFlatFields) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *ResourcesFlatFields) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *ResourcesFlatFields) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *ResourcesFlatFields) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *ResourcesFlatFields) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *ResourcesFlatFields) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *ResourcesFlatFields) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *ResourcesFlatFields) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *ResourcesFlatFields) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *ResourcesFlatFields) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *ResourcesFlatFields) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *ResourcesFlatFields) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *ResourcesFlatFields) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *ResourcesFlatFields) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *ResourcesFlatFields) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *ResourcesFlatFields) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *ResourcesFlatFields) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


