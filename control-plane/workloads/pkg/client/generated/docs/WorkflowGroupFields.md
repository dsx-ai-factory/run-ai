# WorkflowGroupFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DependsOn** | Pointer to **[]string** | The names of the groups that must finish successfully before this group starts. If one of them fails, this group doesn&#39;t start. | [optional] 
**Leader** | Pointer to **NullableString** | The name of the group&#39;s lead task. When the leader ends, the group ends with the leader&#39;s result, and the group&#39;s other tasks are stopped. Without a leader, the group ends when all its tasks have ended, and fails if any of them failed. | [optional] 
**Name** | Pointer to **NullableString** | The name of the group. It must be unique in the workflow, and begin with a letter. (mandatory) | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**SynchronizedStart** | Pointer to [**NullableSynchronizedStart**](SynchronizedStart.md) |  | [optional] 

## Methods

### NewWorkflowGroupFields

`func NewWorkflowGroupFields() *WorkflowGroupFields`

NewWorkflowGroupFields instantiates a new WorkflowGroupFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowGroupFieldsWithDefaults

`func NewWorkflowGroupFieldsWithDefaults() *WorkflowGroupFields`

NewWorkflowGroupFieldsWithDefaults instantiates a new WorkflowGroupFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDependsOn

`func (o *WorkflowGroupFields) GetDependsOn() []*string`

GetDependsOn returns the DependsOn field if non-nil, zero value otherwise.

### GetDependsOnOk

`func (o *WorkflowGroupFields) GetDependsOnOk() (*[]*string, bool)`

GetDependsOnOk returns a tuple with the DependsOn field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependsOn

`func (o *WorkflowGroupFields) SetDependsOn(v []*string)`

SetDependsOn sets DependsOn field to given value.

### HasDependsOn

`func (o *WorkflowGroupFields) HasDependsOn() bool`

HasDependsOn returns a boolean if a field has been set.

### SetDependsOnNil

`func (o *WorkflowGroupFields) SetDependsOnNil(b bool)`

 SetDependsOnNil sets the value for DependsOn to be an explicit nil

### UnsetDependsOn
`func (o *WorkflowGroupFields) UnsetDependsOn()`

UnsetDependsOn ensures that no value is present for DependsOn, not even an explicit nil
### GetLeader

`func (o *WorkflowGroupFields) GetLeader() string`

GetLeader returns the Leader field if non-nil, zero value otherwise.

### GetLeaderOk

`func (o *WorkflowGroupFields) GetLeaderOk() (*string, bool)`

GetLeaderOk returns a tuple with the Leader field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLeader

`func (o *WorkflowGroupFields) SetLeader(v string)`

SetLeader sets Leader field to given value.

### HasLeader

`func (o *WorkflowGroupFields) HasLeader() bool`

HasLeader returns a boolean if a field has been set.

### SetLeaderNil

`func (o *WorkflowGroupFields) SetLeaderNil(b bool)`

 SetLeaderNil sets the value for Leader to be an explicit nil

### UnsetLeader
`func (o *WorkflowGroupFields) UnsetLeader()`

UnsetLeader ensures that no value is present for Leader, not even an explicit nil
### GetName

`func (o *WorkflowGroupFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkflowGroupFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkflowGroupFields) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkflowGroupFields) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *WorkflowGroupFields) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *WorkflowGroupFields) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetNodePools

`func (o *WorkflowGroupFields) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *WorkflowGroupFields) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *WorkflowGroupFields) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *WorkflowGroupFields) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *WorkflowGroupFields) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *WorkflowGroupFields) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetSynchronizedStart

`func (o *WorkflowGroupFields) GetSynchronizedStart() SynchronizedStart`

GetSynchronizedStart returns the SynchronizedStart field if non-nil, zero value otherwise.

### GetSynchronizedStartOk

`func (o *WorkflowGroupFields) GetSynchronizedStartOk() (*SynchronizedStart, bool)`

GetSynchronizedStartOk returns a tuple with the SynchronizedStart field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSynchronizedStart

`func (o *WorkflowGroupFields) SetSynchronizedStart(v SynchronizedStart)`

SetSynchronizedStart sets SynchronizedStart field to given value.

### HasSynchronizedStart

`func (o *WorkflowGroupFields) HasSynchronizedStart() bool`

HasSynchronizedStart returns a boolean if a field has been set.

### SetSynchronizedStartNil

`func (o *WorkflowGroupFields) SetSynchronizedStartNil(b bool)`

 SetSynchronizedStartNil sets the value for SynchronizedStart to be an explicit nil

### UnsetSynchronizedStart
`func (o *WorkflowGroupFields) UnsetSynchronizedStart()`

UnsetSynchronizedStart ensures that no value is present for SynchronizedStart, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


