# Workspace

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadMetaV1**](WorkloadMetaV1.md) |  | 
**Spec** | [**SpecificRunInfoFields**](SpecificRunInfoFields.md) |  | 
**Workload** | Pointer to **map[string]interface{}** | An interactive workload ready for submission at the cluster. Further details at https://github.com/run-ai/runai/runai-cluster/tree/master/workload-controller/config/crd\&quot; | [optional] 
**Job** | Pointer to [**DisplayedJob**](DisplayedJob.md) |  | [optional] 

## Methods

### NewWorkspace

`func NewWorkspace(meta WorkloadMetaV1, spec SpecificRunInfoFields, ) *Workspace`

NewWorkspace instantiates a new Workspace object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceWithDefaults

`func NewWorkspaceWithDefaults() *Workspace`

NewWorkspaceWithDefaults instantiates a new Workspace object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *Workspace) GetMeta() WorkloadMetaV1`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *Workspace) GetMetaOk() (*WorkloadMetaV1, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *Workspace) SetMeta(v WorkloadMetaV1)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *Workspace) GetSpec() SpecificRunInfoFields`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *Workspace) GetSpecOk() (*SpecificRunInfoFields, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *Workspace) SetSpec(v SpecificRunInfoFields)`

SetSpec sets Spec field to given value.


### GetWorkload

`func (o *Workspace) GetWorkload() map[string]interface{}`

GetWorkload returns the Workload field if non-nil, zero value otherwise.

### GetWorkloadOk

`func (o *Workspace) GetWorkloadOk() (*map[string]interface{}, bool)`

GetWorkloadOk returns a tuple with the Workload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkload

`func (o *Workspace) SetWorkload(v map[string]interface{})`

SetWorkload sets Workload field to given value.

### HasWorkload

`func (o *Workspace) HasWorkload() bool`

HasWorkload returns a boolean if a field has been set.

### GetJob

`func (o *Workspace) GetJob() DisplayedJob`

GetJob returns the Job field if non-nil, zero value otherwise.

### GetJobOk

`func (o *Workspace) GetJobOk() (*DisplayedJob, bool)`

GetJobOk returns a tuple with the Job field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJob

`func (o *Workspace) SetJob(v DisplayedJob)`

SetJob sets Job field to given value.

### HasJob

`func (o *Workspace) HasJob() bool`

HasJob returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


