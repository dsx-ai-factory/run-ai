# DistributedInferenceWorkerRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Worker** | Pointer to [**NullableDistributedInferenceLeaderWorkerRulesV2**](DistributedInferenceLeaderWorkerRulesV2.md) | Defines the pod specification for the workers. Required only if the number of workers is greater than 0. | [optional] 

## Methods

### NewDistributedInferenceWorkerRules

`func NewDistributedInferenceWorkerRules() *DistributedInferenceWorkerRules`

NewDistributedInferenceWorkerRules instantiates a new DistributedInferenceWorkerRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceWorkerRulesWithDefaults

`func NewDistributedInferenceWorkerRulesWithDefaults() *DistributedInferenceWorkerRules`

NewDistributedInferenceWorkerRulesWithDefaults instantiates a new DistributedInferenceWorkerRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorker

`func (o *DistributedInferenceWorkerRules) GetWorker() DistributedInferenceLeaderWorkerRulesV2`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *DistributedInferenceWorkerRules) GetWorkerOk() (*DistributedInferenceLeaderWorkerRulesV2, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *DistributedInferenceWorkerRules) SetWorker(v DistributedInferenceLeaderWorkerRulesV2)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *DistributedInferenceWorkerRules) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *DistributedInferenceWorkerRules) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *DistributedInferenceWorkerRules) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


