# WorkflowStorageProfile

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**S3** | Pointer to [**[]WorkflowS3Instance**](WorkflowS3Instance.md) | Set of S3 buckets that the tasks of the workflow can use. | [optional] 

## Methods

### NewWorkflowStorageProfile

`func NewWorkflowStorageProfile() *WorkflowStorageProfile`

NewWorkflowStorageProfile instantiates a new WorkflowStorageProfile object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowStorageProfileWithDefaults

`func NewWorkflowStorageProfileWithDefaults() *WorkflowStorageProfile`

NewWorkflowStorageProfileWithDefaults instantiates a new WorkflowStorageProfile object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetS3

`func (o *WorkflowStorageProfile) GetS3() []WorkflowS3Instance`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *WorkflowStorageProfile) GetS3Ok() (*[]WorkflowS3Instance, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *WorkflowStorageProfile) SetS3(v []WorkflowS3Instance)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *WorkflowStorageProfile) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *WorkflowStorageProfile) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *WorkflowStorageProfile) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


