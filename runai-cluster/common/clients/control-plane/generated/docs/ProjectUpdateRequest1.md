# ProjectUpdateRequest1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Project per Node Pool. | [optional] 
**DeservedGpus** | Pointer to **float32** | Deprecated. Use &#39;deserved&#39; for the relevant resource type under &#x60;NodePoolResources&#x60;. The project&#39;s deserved GPU allocation in case the cluster has those resources. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. An upper limit for the amount of GPUs the project can get (Even if over quota is allowed and resources are available). | [optional] 
**GpuOverQuotaWeight** | Pointer to **float32** | Deprecated. Instead, use &#x60;overQuotaWeight&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. The priority the project gets for over quota resources. | [optional] 
**DefaultNodePools** | Pointer to **[]string** | Default node pools list for workload submission for this project if a workload doesn&#39;t specify a node pools list. | [optional] 
**InteractiveJobTimeLimitSecs** | Pointer to **float32** | A limit (in seconds) for the duration of interactive jobs from this project. | [optional] 
**InteractiveJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that an interactive job can be idle before being terminated. | [optional] 
**InteractivePreemptibleJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that an interactive preemptible job can be idle before being terminated. | [optional] 
**TrainingJobTimeLimitSecs** | Pointer to **float32** | A limit (in seconds) for the duration of training jobs from this project. Available only from cluster version 2.12 | [optional] 
**TrainingJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that a training job can be idle before being terminated. | [optional] 
**NodeAffinity** | Pointer to [**ProjectCommonFieldsNodeAffinity**](ProjectCommonFieldsNodeAffinity.md) |  | [optional] 
**Permissions** | Pointer to [**ProjectCommonFieldsPermissions**](ProjectCommonFieldsPermissions.md) |  | [optional] 
**Resources** | Pointer to [**ProjectCommonFieldsResources**](ProjectCommonFieldsResources.md) |  | [optional] 

## Methods

### NewProjectUpdateRequest1

`func NewProjectUpdateRequest1() *ProjectUpdateRequest1`

NewProjectUpdateRequest1 instantiates a new ProjectUpdateRequest1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectUpdateRequest1WithDefaults

`func NewProjectUpdateRequest1WithDefaults() *ProjectUpdateRequest1`

NewProjectUpdateRequest1WithDefaults instantiates a new ProjectUpdateRequest1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *ProjectUpdateRequest1) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *ProjectUpdateRequest1) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *ProjectUpdateRequest1) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *ProjectUpdateRequest1) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *ProjectUpdateRequest1) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *ProjectUpdateRequest1) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *ProjectUpdateRequest1) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *ProjectUpdateRequest1) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *ProjectUpdateRequest1) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *ProjectUpdateRequest1) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *ProjectUpdateRequest1) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *ProjectUpdateRequest1) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetGpuOverQuotaWeight

`func (o *ProjectUpdateRequest1) GetGpuOverQuotaWeight() float32`

GetGpuOverQuotaWeight returns the GpuOverQuotaWeight field if non-nil, zero value otherwise.

### GetGpuOverQuotaWeightOk

`func (o *ProjectUpdateRequest1) GetGpuOverQuotaWeightOk() (*float32, bool)`

GetGpuOverQuotaWeightOk returns a tuple with the GpuOverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuOverQuotaWeight

`func (o *ProjectUpdateRequest1) SetGpuOverQuotaWeight(v float32)`

SetGpuOverQuotaWeight sets GpuOverQuotaWeight field to given value.

### HasGpuOverQuotaWeight

`func (o *ProjectUpdateRequest1) HasGpuOverQuotaWeight() bool`

HasGpuOverQuotaWeight returns a boolean if a field has been set.

### GetDefaultNodePools

`func (o *ProjectUpdateRequest1) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *ProjectUpdateRequest1) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *ProjectUpdateRequest1) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *ProjectUpdateRequest1) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### GetInteractiveJobTimeLimitSecs

`func (o *ProjectUpdateRequest1) GetInteractiveJobTimeLimitSecs() float32`

GetInteractiveJobTimeLimitSecs returns the InteractiveJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetInteractiveJobTimeLimitSecsOk

`func (o *ProjectUpdateRequest1) GetInteractiveJobTimeLimitSecsOk() (*float32, bool)`

GetInteractiveJobTimeLimitSecsOk returns a tuple with the InteractiveJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobTimeLimitSecs

`func (o *ProjectUpdateRequest1) SetInteractiveJobTimeLimitSecs(v float32)`

SetInteractiveJobTimeLimitSecs sets InteractiveJobTimeLimitSecs field to given value.

### HasInteractiveJobTimeLimitSecs

`func (o *ProjectUpdateRequest1) HasInteractiveJobTimeLimitSecs() bool`

HasInteractiveJobTimeLimitSecs returns a boolean if a field has been set.

### GetInteractiveJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) GetInteractiveJobMaxIdleDurationSecs() float32`

GetInteractiveJobMaxIdleDurationSecs returns the InteractiveJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractiveJobMaxIdleDurationSecsOk

`func (o *ProjectUpdateRequest1) GetInteractiveJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractiveJobMaxIdleDurationSecsOk returns a tuple with the InteractiveJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) SetInteractiveJobMaxIdleDurationSecs(v float32)`

SetInteractiveJobMaxIdleDurationSecs sets InteractiveJobMaxIdleDurationSecs field to given value.

### HasInteractiveJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) HasInteractiveJobMaxIdleDurationSecs() bool`

HasInteractiveJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) GetInteractivePreemptibleJobMaxIdleDurationSecs() float32`

GetInteractivePreemptibleJobMaxIdleDurationSecs returns the InteractivePreemptibleJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractivePreemptibleJobMaxIdleDurationSecsOk

`func (o *ProjectUpdateRequest1) GetInteractivePreemptibleJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractivePreemptibleJobMaxIdleDurationSecsOk returns a tuple with the InteractivePreemptibleJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) SetInteractivePreemptibleJobMaxIdleDurationSecs(v float32)`

SetInteractivePreemptibleJobMaxIdleDurationSecs sets InteractivePreemptibleJobMaxIdleDurationSecs field to given value.

### HasInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) HasInteractivePreemptibleJobMaxIdleDurationSecs() bool`

HasInteractivePreemptibleJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetTrainingJobTimeLimitSecs

`func (o *ProjectUpdateRequest1) GetTrainingJobTimeLimitSecs() float32`

GetTrainingJobTimeLimitSecs returns the TrainingJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetTrainingJobTimeLimitSecsOk

`func (o *ProjectUpdateRequest1) GetTrainingJobTimeLimitSecsOk() (*float32, bool)`

GetTrainingJobTimeLimitSecsOk returns a tuple with the TrainingJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobTimeLimitSecs

`func (o *ProjectUpdateRequest1) SetTrainingJobTimeLimitSecs(v float32)`

SetTrainingJobTimeLimitSecs sets TrainingJobTimeLimitSecs field to given value.

### HasTrainingJobTimeLimitSecs

`func (o *ProjectUpdateRequest1) HasTrainingJobTimeLimitSecs() bool`

HasTrainingJobTimeLimitSecs returns a boolean if a field has been set.

### GetTrainingJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) GetTrainingJobMaxIdleDurationSecs() float32`

GetTrainingJobMaxIdleDurationSecs returns the TrainingJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetTrainingJobMaxIdleDurationSecsOk

`func (o *ProjectUpdateRequest1) GetTrainingJobMaxIdleDurationSecsOk() (*float32, bool)`

GetTrainingJobMaxIdleDurationSecsOk returns a tuple with the TrainingJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) SetTrainingJobMaxIdleDurationSecs(v float32)`

SetTrainingJobMaxIdleDurationSecs sets TrainingJobMaxIdleDurationSecs field to given value.

### HasTrainingJobMaxIdleDurationSecs

`func (o *ProjectUpdateRequest1) HasTrainingJobMaxIdleDurationSecs() bool`

HasTrainingJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetNodeAffinity

`func (o *ProjectUpdateRequest1) GetNodeAffinity() ProjectCommonFieldsNodeAffinity`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *ProjectUpdateRequest1) GetNodeAffinityOk() (*ProjectCommonFieldsNodeAffinity, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *ProjectUpdateRequest1) SetNodeAffinity(v ProjectCommonFieldsNodeAffinity)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *ProjectUpdateRequest1) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.

### GetPermissions

`func (o *ProjectUpdateRequest1) GetPermissions() ProjectCommonFieldsPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *ProjectUpdateRequest1) GetPermissionsOk() (*ProjectCommonFieldsPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *ProjectUpdateRequest1) SetPermissions(v ProjectCommonFieldsPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *ProjectUpdateRequest1) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetResources

`func (o *ProjectUpdateRequest1) GetResources() ProjectCommonFieldsResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ProjectUpdateRequest1) GetResourcesOk() (*ProjectCommonFieldsResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ProjectUpdateRequest1) SetResources(v ProjectCommonFieldsResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *ProjectUpdateRequest1) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


