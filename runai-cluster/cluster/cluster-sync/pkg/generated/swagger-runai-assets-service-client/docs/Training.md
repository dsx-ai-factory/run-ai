# Training

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadMeta**](WorkloadMeta.md) |  | 
**Spec** | [**SpecificRunInfoFields**](SpecificRunInfoFields.md) |  | 
**Workload** | Pointer to **map[string]interface{}** | A training workload ready for submission at the cluster. Further details at https://github.com/dsx-ai-factory/run-ai/runai-cluster/tree/master/workload-controller/config/crd\&quot; | [optional] 
**Job** | Pointer to [**DisplayedJob**](DisplayedJob.md) |  | [optional] 

## Methods

### NewTraining

`func NewTraining(meta WorkloadMeta, spec SpecificRunInfoFields, ) *Training`

NewTraining instantiates a new Training object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingWithDefaults

`func NewTrainingWithDefaults() *Training`

NewTrainingWithDefaults instantiates a new Training object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *Training) GetMeta() WorkloadMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *Training) GetMetaOk() (*WorkloadMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *Training) SetMeta(v WorkloadMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *Training) GetSpec() SpecificRunInfoFields`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *Training) GetSpecOk() (*SpecificRunInfoFields, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *Training) SetSpec(v SpecificRunInfoFields)`

SetSpec sets Spec field to given value.


### GetWorkload

`func (o *Training) GetWorkload() map[string]interface{}`

GetWorkload returns the Workload field if non-nil, zero value otherwise.

### GetWorkloadOk

`func (o *Training) GetWorkloadOk() (*map[string]interface{}, bool)`

GetWorkloadOk returns a tuple with the Workload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkload

`func (o *Training) SetWorkload(v map[string]interface{})`

SetWorkload sets Workload field to given value.

### HasWorkload

`func (o *Training) HasWorkload() bool`

HasWorkload returns a boolean if a field has been set.

### GetJob

`func (o *Training) GetJob() DisplayedJob`

GetJob returns the Job field if non-nil, zero value otherwise.

### GetJobOk

`func (o *Training) GetJobOk() (*DisplayedJob, bool)`

GetJobOk returns a tuple with the Job field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJob

`func (o *Training) SetJob(v DisplayedJob)`

SetJob sets Job field to given value.

### HasJob

`func (o *Training) HasJob() bool`

HasJob returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


