# DistributedInferenceWorkerTemplateSpecFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Worker** | Pointer to [**NullableDistributedInferenceLeaderWorkerTemplateSpec**](DistributedInferenceLeaderWorkerTemplateSpec.md) | Defines the pod specification for the workers. | [optional] 

## Methods

### NewDistributedInferenceWorkerTemplateSpecFields

`func NewDistributedInferenceWorkerTemplateSpecFields() *DistributedInferenceWorkerTemplateSpecFields`

NewDistributedInferenceWorkerTemplateSpecFields instantiates a new DistributedInferenceWorkerTemplateSpecFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceWorkerTemplateSpecFieldsWithDefaults

`func NewDistributedInferenceWorkerTemplateSpecFieldsWithDefaults() *DistributedInferenceWorkerTemplateSpecFields`

NewDistributedInferenceWorkerTemplateSpecFieldsWithDefaults instantiates a new DistributedInferenceWorkerTemplateSpecFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorker

`func (o *DistributedInferenceWorkerTemplateSpecFields) GetWorker() DistributedInferenceLeaderWorkerTemplateSpec`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *DistributedInferenceWorkerTemplateSpecFields) GetWorkerOk() (*DistributedInferenceLeaderWorkerTemplateSpec, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *DistributedInferenceWorkerTemplateSpecFields) SetWorker(v DistributedInferenceLeaderWorkerTemplateSpec)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *DistributedInferenceWorkerTemplateSpecFields) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *DistributedInferenceWorkerTemplateSpecFields) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *DistributedInferenceWorkerTemplateSpecFields) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


