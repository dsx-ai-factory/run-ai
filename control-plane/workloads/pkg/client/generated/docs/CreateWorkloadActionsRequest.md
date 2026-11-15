# CreateWorkloadActionsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | [**[]WorkloadActionRecordInput**](WorkloadActionRecordInput.md) |  | 

## Methods

### NewCreateWorkloadActionsRequest

`func NewCreateWorkloadActionsRequest(actions []WorkloadActionRecordInput, ) *CreateWorkloadActionsRequest`

NewCreateWorkloadActionsRequest instantiates a new CreateWorkloadActionsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateWorkloadActionsRequestWithDefaults

`func NewCreateWorkloadActionsRequestWithDefaults() *CreateWorkloadActionsRequest`

NewCreateWorkloadActionsRequestWithDefaults instantiates a new CreateWorkloadActionsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *CreateWorkloadActionsRequest) GetActions() []WorkloadActionRecordInput`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *CreateWorkloadActionsRequest) GetActionsOk() (*[]WorkloadActionRecordInput, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *CreateWorkloadActionsRequest) SetActions(v []WorkloadActionRecordInput)`

SetActions sets Actions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


