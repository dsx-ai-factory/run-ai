# WorkflowStorageRefInput

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Group** | Pointer to **NullableString** | The name of the group. It must be unique in the workflow, and begin with a letter. (mandatory) | [optional] 
**Task** | Pointer to **NullableString** | The name of the task. It must be unique in its group, and begin with a letter. It can be omitted in a group with a single task, which then takes the group&#39;s name. | [optional] 
**MountPath** | Pointer to **NullableString** | The path of the folder in the task&#39;s container. (mandatory) | [optional] 

## Methods

### NewWorkflowStorageRefInput

`func NewWorkflowStorageRefInput() *WorkflowStorageRefInput`

NewWorkflowStorageRefInput instantiates a new WorkflowStorageRefInput object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowStorageRefInputWithDefaults

`func NewWorkflowStorageRefInputWithDefaults() *WorkflowStorageRefInput`

NewWorkflowStorageRefInputWithDefaults instantiates a new WorkflowStorageRefInput object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGroup

`func (o *WorkflowStorageRefInput) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *WorkflowStorageRefInput) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *WorkflowStorageRefInput) SetGroup(v string)`

SetGroup sets Group field to given value.

### HasGroup

`func (o *WorkflowStorageRefInput) HasGroup() bool`

HasGroup returns a boolean if a field has been set.

### SetGroupNil

`func (o *WorkflowStorageRefInput) SetGroupNil(b bool)`

 SetGroupNil sets the value for Group to be an explicit nil

### UnsetGroup
`func (o *WorkflowStorageRefInput) UnsetGroup()`

UnsetGroup ensures that no value is present for Group, not even an explicit nil
### GetTask

`func (o *WorkflowStorageRefInput) GetTask() string`

GetTask returns the Task field if non-nil, zero value otherwise.

### GetTaskOk

`func (o *WorkflowStorageRefInput) GetTaskOk() (*string, bool)`

GetTaskOk returns a tuple with the Task field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTask

`func (o *WorkflowStorageRefInput) SetTask(v string)`

SetTask sets Task field to given value.

### HasTask

`func (o *WorkflowStorageRefInput) HasTask() bool`

HasTask returns a boolean if a field has been set.

### SetTaskNil

`func (o *WorkflowStorageRefInput) SetTaskNil(b bool)`

 SetTaskNil sets the value for Task to be an explicit nil

### UnsetTask
`func (o *WorkflowStorageRefInput) UnsetTask()`

UnsetTask ensures that no value is present for Task, not even an explicit nil
### GetMountPath

`func (o *WorkflowStorageRefInput) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *WorkflowStorageRefInput) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *WorkflowStorageRefInput) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *WorkflowStorageRefInput) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *WorkflowStorageRefInput) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *WorkflowStorageRefInput) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


