# ProjectV1CreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Project Name | 
**DepartmentId** | **int32** | ID of the department that owns the project. | 
**Namespace** | Pointer to **string** | The name of an existing namespace to use for the project in the cluster. Supported only for cluster versions 2.12 or higher. | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesV1**](NodePoolAssignedResourcesV1.md) | Resources assigned to this Project per Node Pool. | [optional] 
**DeservedGpus** | Pointer to **float32** | Deprecated. Use &#39;deserved&#39; for the relevant resource type under &#x60;NodePoolResources&#x60;. The project&#39;s deserved GPU allocation in case the cluster has those resources. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. An upper limit for the amount of GPUs the project can get (Even if over quota is allowed and resources are available). | [optional] 
**GpuOverQuotaWeight** | Pointer to **float32** | Deprecated. Instead, use &#x60;overQuotaWeight&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. The priority the project gets for over quota resources. | [optional] 
**DefaultNodePools** | Pointer to **[]string** | Default node pools list for workload submission for this project if a workload doesn&#39;t specify a node pools list. | [optional] 
**InteractiveJobTimeLimitSecs** | Pointer to **float32** | A limit (in seconds) for the duration of interactive jobs from this project. | [optional] 
**InteractiveJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that an interactive job can be idle before being terminated. | [optional] 
**InteractivePreemptibleJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that an interactive preemptible job can be idle before being terminated. | [optional] 
**TrainingJobTimeLimitSecs** | Pointer to **float32** | A limit (in seconds) for the duration of training jobs from this project. Available only from cluster version 2.12 | [optional] 
**TrainingJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that a training job can be idle before being terminated. | [optional] 
**NodeAffinity** | Pointer to [**NullableProjectV1NodeAffinity**](ProjectV1NodeAffinity.md) |  | [optional] 
**Permissions** | Pointer to [**ProjectV1CreationRequestPermissions**](ProjectV1CreationRequestPermissions.md) |  | [optional] 
**Resources** | Pointer to [**ProjectV1CreationRequestResources**](ProjectV1CreationRequestResources.md) |  | [optional] 

## Methods

### NewProjectV1CreationRequest

`func NewProjectV1CreationRequest(name string, departmentId int32, ) *ProjectV1CreationRequest`

NewProjectV1CreationRequest instantiates a new ProjectV1CreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectV1CreationRequestWithDefaults

`func NewProjectV1CreationRequestWithDefaults() *ProjectV1CreationRequest`

NewProjectV1CreationRequestWithDefaults instantiates a new ProjectV1CreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProjectV1CreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectV1CreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectV1CreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDepartmentId

`func (o *ProjectV1CreationRequest) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *ProjectV1CreationRequest) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *ProjectV1CreationRequest) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.


### GetNamespace

`func (o *ProjectV1CreationRequest) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectV1CreationRequest) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectV1CreationRequest) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ProjectV1CreationRequest) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *ProjectV1CreationRequest) GetNodePoolsResources() []NodePoolAssignedResourcesV1`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *ProjectV1CreationRequest) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesV1, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *ProjectV1CreationRequest) SetNodePoolsResources(v []NodePoolAssignedResourcesV1)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *ProjectV1CreationRequest) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *ProjectV1CreationRequest) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *ProjectV1CreationRequest) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *ProjectV1CreationRequest) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *ProjectV1CreationRequest) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *ProjectV1CreationRequest) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *ProjectV1CreationRequest) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *ProjectV1CreationRequest) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *ProjectV1CreationRequest) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetGpuOverQuotaWeight

`func (o *ProjectV1CreationRequest) GetGpuOverQuotaWeight() float32`

GetGpuOverQuotaWeight returns the GpuOverQuotaWeight field if non-nil, zero value otherwise.

### GetGpuOverQuotaWeightOk

`func (o *ProjectV1CreationRequest) GetGpuOverQuotaWeightOk() (*float32, bool)`

GetGpuOverQuotaWeightOk returns a tuple with the GpuOverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuOverQuotaWeight

`func (o *ProjectV1CreationRequest) SetGpuOverQuotaWeight(v float32)`

SetGpuOverQuotaWeight sets GpuOverQuotaWeight field to given value.

### HasGpuOverQuotaWeight

`func (o *ProjectV1CreationRequest) HasGpuOverQuotaWeight() bool`

HasGpuOverQuotaWeight returns a boolean if a field has been set.

### GetDefaultNodePools

`func (o *ProjectV1CreationRequest) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *ProjectV1CreationRequest) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *ProjectV1CreationRequest) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *ProjectV1CreationRequest) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### GetInteractiveJobTimeLimitSecs

`func (o *ProjectV1CreationRequest) GetInteractiveJobTimeLimitSecs() float32`

GetInteractiveJobTimeLimitSecs returns the InteractiveJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetInteractiveJobTimeLimitSecsOk

`func (o *ProjectV1CreationRequest) GetInteractiveJobTimeLimitSecsOk() (*float32, bool)`

GetInteractiveJobTimeLimitSecsOk returns a tuple with the InteractiveJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobTimeLimitSecs

`func (o *ProjectV1CreationRequest) SetInteractiveJobTimeLimitSecs(v float32)`

SetInteractiveJobTimeLimitSecs sets InteractiveJobTimeLimitSecs field to given value.

### HasInteractiveJobTimeLimitSecs

`func (o *ProjectV1CreationRequest) HasInteractiveJobTimeLimitSecs() bool`

HasInteractiveJobTimeLimitSecs returns a boolean if a field has been set.

### GetInteractiveJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) GetInteractiveJobMaxIdleDurationSecs() float32`

GetInteractiveJobMaxIdleDurationSecs returns the InteractiveJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractiveJobMaxIdleDurationSecsOk

`func (o *ProjectV1CreationRequest) GetInteractiveJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractiveJobMaxIdleDurationSecsOk returns a tuple with the InteractiveJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) SetInteractiveJobMaxIdleDurationSecs(v float32)`

SetInteractiveJobMaxIdleDurationSecs sets InteractiveJobMaxIdleDurationSecs field to given value.

### HasInteractiveJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) HasInteractiveJobMaxIdleDurationSecs() bool`

HasInteractiveJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) GetInteractivePreemptibleJobMaxIdleDurationSecs() float32`

GetInteractivePreemptibleJobMaxIdleDurationSecs returns the InteractivePreemptibleJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractivePreemptibleJobMaxIdleDurationSecsOk

`func (o *ProjectV1CreationRequest) GetInteractivePreemptibleJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractivePreemptibleJobMaxIdleDurationSecsOk returns a tuple with the InteractivePreemptibleJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) SetInteractivePreemptibleJobMaxIdleDurationSecs(v float32)`

SetInteractivePreemptibleJobMaxIdleDurationSecs sets InteractivePreemptibleJobMaxIdleDurationSecs field to given value.

### HasInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) HasInteractivePreemptibleJobMaxIdleDurationSecs() bool`

HasInteractivePreemptibleJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetTrainingJobTimeLimitSecs

`func (o *ProjectV1CreationRequest) GetTrainingJobTimeLimitSecs() float32`

GetTrainingJobTimeLimitSecs returns the TrainingJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetTrainingJobTimeLimitSecsOk

`func (o *ProjectV1CreationRequest) GetTrainingJobTimeLimitSecsOk() (*float32, bool)`

GetTrainingJobTimeLimitSecsOk returns a tuple with the TrainingJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobTimeLimitSecs

`func (o *ProjectV1CreationRequest) SetTrainingJobTimeLimitSecs(v float32)`

SetTrainingJobTimeLimitSecs sets TrainingJobTimeLimitSecs field to given value.

### HasTrainingJobTimeLimitSecs

`func (o *ProjectV1CreationRequest) HasTrainingJobTimeLimitSecs() bool`

HasTrainingJobTimeLimitSecs returns a boolean if a field has been set.

### GetTrainingJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) GetTrainingJobMaxIdleDurationSecs() float32`

GetTrainingJobMaxIdleDurationSecs returns the TrainingJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetTrainingJobMaxIdleDurationSecsOk

`func (o *ProjectV1CreationRequest) GetTrainingJobMaxIdleDurationSecsOk() (*float32, bool)`

GetTrainingJobMaxIdleDurationSecsOk returns a tuple with the TrainingJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) SetTrainingJobMaxIdleDurationSecs(v float32)`

SetTrainingJobMaxIdleDurationSecs sets TrainingJobMaxIdleDurationSecs field to given value.

### HasTrainingJobMaxIdleDurationSecs

`func (o *ProjectV1CreationRequest) HasTrainingJobMaxIdleDurationSecs() bool`

HasTrainingJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetNodeAffinity

`func (o *ProjectV1CreationRequest) GetNodeAffinity() ProjectV1NodeAffinity`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *ProjectV1CreationRequest) GetNodeAffinityOk() (*ProjectV1NodeAffinity, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *ProjectV1CreationRequest) SetNodeAffinity(v ProjectV1NodeAffinity)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *ProjectV1CreationRequest) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.

### SetNodeAffinityNil

`func (o *ProjectV1CreationRequest) SetNodeAffinityNil(b bool)`

 SetNodeAffinityNil sets the value for NodeAffinity to be an explicit nil

### UnsetNodeAffinity
`func (o *ProjectV1CreationRequest) UnsetNodeAffinity()`

UnsetNodeAffinity ensures that no value is present for NodeAffinity, not even an explicit nil
### GetPermissions

`func (o *ProjectV1CreationRequest) GetPermissions() ProjectV1CreationRequestPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *ProjectV1CreationRequest) GetPermissionsOk() (*ProjectV1CreationRequestPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *ProjectV1CreationRequest) SetPermissions(v ProjectV1CreationRequestPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *ProjectV1CreationRequest) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetResources

`func (o *ProjectV1CreationRequest) GetResources() ProjectV1CreationRequestResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ProjectV1CreationRequest) GetResourcesOk() (*ProjectV1CreationRequestResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ProjectV1CreationRequest) SetResources(v ProjectV1CreationRequestResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *ProjectV1CreationRequest) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


