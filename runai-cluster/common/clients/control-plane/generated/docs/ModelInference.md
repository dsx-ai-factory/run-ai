# ModelInference

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadMetaV1**](WorkloadMetaV1.md) |  | 
**Spec** | [**SpecificRunInfoFields**](SpecificRunInfoFields.md) |  | 
**Workload** | Pointer to **map[string]interface{}** | An inference workload ready for submission at the cluster. Further details at https://github.com/run-ai/runai/runai-cluster/tree/master/workload-controller/config/crd\&quot; | [optional] 
**Job** | Pointer to [**DisplayedJob1**](DisplayedJob1.md) |  | [optional] 

## Methods

### NewModelInference

`func NewModelInference(meta WorkloadMetaV1, spec SpecificRunInfoFields, ) *ModelInference`

NewModelInference instantiates a new ModelInference object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelInferenceWithDefaults

`func NewModelInferenceWithDefaults() *ModelInference`

NewModelInferenceWithDefaults instantiates a new ModelInference object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *ModelInference) GetMeta() WorkloadMetaV1`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *ModelInference) GetMetaOk() (*WorkloadMetaV1, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *ModelInference) SetMeta(v WorkloadMetaV1)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *ModelInference) GetSpec() SpecificRunInfoFields`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ModelInference) GetSpecOk() (*SpecificRunInfoFields, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ModelInference) SetSpec(v SpecificRunInfoFields)`

SetSpec sets Spec field to given value.


### GetWorkload

`func (o *ModelInference) GetWorkload() map[string]interface{}`

GetWorkload returns the Workload field if non-nil, zero value otherwise.

### GetWorkloadOk

`func (o *ModelInference) GetWorkloadOk() (*map[string]interface{}, bool)`

GetWorkloadOk returns a tuple with the Workload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkload

`func (o *ModelInference) SetWorkload(v map[string]interface{})`

SetWorkload sets Workload field to given value.

### HasWorkload

`func (o *ModelInference) HasWorkload() bool`

HasWorkload returns a boolean if a field has been set.

### GetJob

`func (o *ModelInference) GetJob() DisplayedJob1`

GetJob returns the Job field if non-nil, zero value otherwise.

### GetJobOk

`func (o *ModelInference) GetJobOk() (*DisplayedJob1, bool)`

GetJobOk returns a tuple with the Job field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJob

`func (o *ModelInference) SetJob(v DisplayedJob1)`

SetJob sets Job field to given value.

### HasJob

`func (o *ModelInference) HasJob() bool`

HasJob returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


