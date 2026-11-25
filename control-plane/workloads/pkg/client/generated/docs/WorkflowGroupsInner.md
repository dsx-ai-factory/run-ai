# WorkflowGroupsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DependsOn** | Pointer to **[]string** | The names of the groups that must finish successfully before this group starts. If one of them fails, this group doesn&#39;t start. | [optional] 
**Leader** | Pointer to **NullableString** | The name of the group&#39;s lead task. When the leader ends, the group ends with the leader&#39;s result, and the group&#39;s other tasks are stopped. Without a leader, the group ends when all its tasks have ended, and fails if any of them failed. | [optional] 
**Name** | Pointer to **NullableString** | The name of the group. It must be unique in the workflow, and begin with a letter. (mandatory) | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**SynchronizedStart** | Pointer to [**NullableSynchronizedStart**](SynchronizedStart.md) |  | [optional] 
**Tasks** | Pointer to [**[]WorkflowTask**](WorkflowTask.md) |  | [optional] 

## Methods

### NewWorkflowGroupsInner

`func NewWorkflowGroupsInner() *WorkflowGroupsInner`

NewWorkflowGroupsInner instantiates a new WorkflowGroupsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowGroupsInnerWithDefaults

`func NewWorkflowGroupsInnerWithDefaults() *WorkflowGroupsInner`

NewWorkflowGroupsInnerWithDefaults instantiates a new WorkflowGroupsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDependsOn

`func (o *WorkflowGroupsInner) GetDependsOn() []*string`

GetDependsOn returns the DependsOn field if non-nil, zero value otherwise.

### GetDependsOnOk

`func (o *WorkflowGroupsInner) GetDependsOnOk() (*[]*string, bool)`

GetDependsOnOk returns a tuple with the DependsOn field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependsOn

`func (o *WorkflowGroupsInner) SetDependsOn(v []*string)`

SetDependsOn sets DependsOn field to given value.

### HasDependsOn

`func (o *WorkflowGroupsInner) HasDependsOn() bool`

HasDependsOn returns a boolean if a field has been set.

### SetDependsOnNil

`func (o *WorkflowGroupsInner) SetDependsOnNil(b bool)`

 SetDependsOnNil sets the value for DependsOn to be an explicit nil

### UnsetDependsOn
`func (o *WorkflowGroupsInner) UnsetDependsOn()`

UnsetDependsOn ensures that no value is present for DependsOn, not even an explicit nil
### GetLeader

`func (o *WorkflowGroupsInner) GetLeader() string`

GetLeader returns the Leader field if non-nil, zero value otherwise.

### GetLeaderOk

`func (o *WorkflowGroupsInner) GetLeaderOk() (*string, bool)`

GetLeaderOk returns a tuple with the Leader field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLeader

`func (o *WorkflowGroupsInner) SetLeader(v string)`

SetLeader sets Leader field to given value.

### HasLeader

`func (o *WorkflowGroupsInner) HasLeader() bool`

HasLeader returns a boolean if a field has been set.

### SetLeaderNil

`func (o *WorkflowGroupsInner) SetLeaderNil(b bool)`

 SetLeaderNil sets the value for Leader to be an explicit nil

### UnsetLeader
`func (o *WorkflowGroupsInner) UnsetLeader()`

UnsetLeader ensures that no value is present for Leader, not even an explicit nil
### GetName

`func (o *WorkflowGroupsInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkflowGroupsInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkflowGroupsInner) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkflowGroupsInner) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *WorkflowGroupsInner) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *WorkflowGroupsInner) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetNodePools

`func (o *WorkflowGroupsInner) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *WorkflowGroupsInner) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *WorkflowGroupsInner) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *WorkflowGroupsInner) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *WorkflowGroupsInner) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *WorkflowGroupsInner) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetSynchronizedStart

`func (o *WorkflowGroupsInner) GetSynchronizedStart() SynchronizedStart`

GetSynchronizedStart returns the SynchronizedStart field if non-nil, zero value otherwise.

### GetSynchronizedStartOk

`func (o *WorkflowGroupsInner) GetSynchronizedStartOk() (*SynchronizedStart, bool)`

GetSynchronizedStartOk returns a tuple with the SynchronizedStart field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSynchronizedStart

`func (o *WorkflowGroupsInner) SetSynchronizedStart(v SynchronizedStart)`

SetSynchronizedStart sets SynchronizedStart field to given value.

### HasSynchronizedStart

`func (o *WorkflowGroupsInner) HasSynchronizedStart() bool`

HasSynchronizedStart returns a boolean if a field has been set.

### SetSynchronizedStartNil

`func (o *WorkflowGroupsInner) SetSynchronizedStartNil(b bool)`

 SetSynchronizedStartNil sets the value for SynchronizedStart to be an explicit nil

### UnsetSynchronizedStart
`func (o *WorkflowGroupsInner) UnsetSynchronizedStart()`

UnsetSynchronizedStart ensures that no value is present for SynchronizedStart, not even an explicit nil
### GetTasks

`func (o *WorkflowGroupsInner) GetTasks() []WorkflowTask`

GetTasks returns the Tasks field if non-nil, zero value otherwise.

### GetTasksOk

`func (o *WorkflowGroupsInner) GetTasksOk() (*[]WorkflowTask, bool)`

GetTasksOk returns a tuple with the Tasks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTasks

`func (o *WorkflowGroupsInner) SetTasks(v []WorkflowTask)`

SetTasks sets Tasks field to given value.

### HasTasks

`func (o *WorkflowGroupsInner) HasTasks() bool`

HasTasks returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


