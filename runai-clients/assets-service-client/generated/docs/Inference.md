# Inference

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadMetaV1**](WorkloadMetaV1.md) |  | 
**Spec** | [**SpecificRunInfoFields**](SpecificRunInfoFields.md) |  | 
**Workload** | Pointer to **map[string]interface{}** | An inference workload ready for submission at the cluster. Further details at https://github.com/run-ai/runai/runai-cluster/tree/master/workload-controller/config/crd\&quot; | [optional] 
**Job** | Pointer to [**DisplayedJob**](DisplayedJob.md) |  | [optional] 

## Methods

### NewInference

`func NewInference(meta WorkloadMetaV1, spec SpecificRunInfoFields, ) *Inference`

NewInference instantiates a new Inference object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceWithDefaults

`func NewInferenceWithDefaults() *Inference`

NewInferenceWithDefaults instantiates a new Inference object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *Inference) GetMeta() WorkloadMetaV1`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *Inference) GetMetaOk() (*WorkloadMetaV1, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *Inference) SetMeta(v WorkloadMetaV1)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *Inference) GetSpec() SpecificRunInfoFields`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *Inference) GetSpecOk() (*SpecificRunInfoFields, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *Inference) SetSpec(v SpecificRunInfoFields)`

SetSpec sets Spec field to given value.


### GetWorkload

`func (o *Inference) GetWorkload() map[string]interface{}`

GetWorkload returns the Workload field if non-nil, zero value otherwise.

### GetWorkloadOk

`func (o *Inference) GetWorkloadOk() (*map[string]interface{}, bool)`

GetWorkloadOk returns a tuple with the Workload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkload

`func (o *Inference) SetWorkload(v map[string]interface{})`

SetWorkload sets Workload field to given value.

### HasWorkload

`func (o *Inference) HasWorkload() bool`

HasWorkload returns a boolean if a field has been set.

### GetJob

`func (o *Inference) GetJob() DisplayedJob`

GetJob returns the Job field if non-nil, zero value otherwise.

### GetJobOk

`func (o *Inference) GetJobOk() (*DisplayedJob, bool)`

GetJobOk returns a tuple with the Job field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJob

`func (o *Inference) SetJob(v DisplayedJob)`

SetJob sets Job field to given value.

### HasJob

`func (o *Inference) HasJob() bool`

HasJob returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


