# WorkflowGroupsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Todo** | Pointer to **bool** |  | [optional] 
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

### GetTodo

`func (o *WorkflowGroupsInner) GetTodo() bool`

GetTodo returns the Todo field if non-nil, zero value otherwise.

### GetTodoOk

`func (o *WorkflowGroupsInner) GetTodoOk() (*bool, bool)`

GetTodoOk returns a tuple with the Todo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTodo

`func (o *WorkflowGroupsInner) SetTodo(v bool)`

SetTodo sets Todo field to given value.

### HasTodo

`func (o *WorkflowGroupsInner) HasTodo() bool`

HasTodo returns a boolean if a field has been set.

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


