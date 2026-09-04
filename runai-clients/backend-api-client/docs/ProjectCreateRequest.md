# ProjectCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DeservedGpus** | Pointer to **float32** | Deprecated. Use &#39;deserved&#39; for the relevant resource type under &#x60;NodePoolResources&#x60;. The project&#39;s deserved GPU allocation in case the cluster has those resources. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. An upper limit for the amount of GPUs the project can get (Even if over quota is allowed and resources are available). | [optional] 
**GpuOverQuotaWeight** | Pointer to **float32** | Deprecated. Instead, use &#x60;overQuotaWeight&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. The priority the project gets for over quota resources. | [optional] 
**DefaultNodePools** | Pointer to **[]string** | Default node pools list for workload submission for this project if a workload doesn&#39;t specify a node pools list. | [optional] 
**InteractiveJobTimeLimitSecs** | Pointer to **float32** | A limit (in seconds) for the duration of interactive jobs from this project. | [optional] 
**InteractiveJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that an interactive job can be idle before being terminated. | [optional] 
**InteractivePreemptibleJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that an interactive preemptible job can be idle before being terminated. | [optional] 
**TrainingJobTimeLimitSecs** | Pointer to **float32** | A limit (in seconds) for the duration of training jobs from this project. Available only from cluster version 2.12 | [optional] 
**TrainingJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that a training job can be idle before being terminated. | [optional] 
**NodeAffinity** | Pointer to [**JobsNodeAffinity**](JobsNodeAffinity.md) | Node affinity configuration for jobs in the project. | [optional] 
**Permissions** | Pointer to [**ResourcePermissions**](ResourcePermissions.md) | Deprecated. Instead, use the &#x60;accessRules&#x60; API to add permissions to a specific subject in the project scope. | [optional] 
**Resources** | Pointer to [**AssignedResources**](AssignedResources.md) | Deprecated. Instead, use &#x60;nodePoolsResources&#x60;. Total resources assigned to the Project. Can only be used in PUT/POST when there is a single Node Pool in the system. The resources returned in &#x60;GET&#x60; are the sum of all Node Pool Resources. | [optional] 
**Name** | **string** | Project Name | 
**DepartmentId** | **int32** | ID of the department that owns the project. | 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesCreate**](NodePoolAssignedResourcesCreate.md) | Resources assigned to this Project per Node Pool. | [optional] 
**Namespace** | Pointer to **string** | The name of an existing namespace to use for the project in the cluster. Supported only for cluster versions 2.12 or higher. | [optional] 

## Methods

### NewProjectCreateRequest

`func NewProjectCreateRequest(name string, departmentId int32, ) *ProjectCreateRequest`

NewProjectCreateRequest instantiates a new ProjectCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectCreateRequestWithDefaults

`func NewProjectCreateRequestWithDefaults() *ProjectCreateRequest`

NewProjectCreateRequestWithDefaults instantiates a new ProjectCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeservedGpus

`func (o *ProjectCreateRequest) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *ProjectCreateRequest) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *ProjectCreateRequest) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *ProjectCreateRequest) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *ProjectCreateRequest) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *ProjectCreateRequest) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *ProjectCreateRequest) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *ProjectCreateRequest) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetGpuOverQuotaWeight

`func (o *ProjectCreateRequest) GetGpuOverQuotaWeight() float32`

GetGpuOverQuotaWeight returns the GpuOverQuotaWeight field if non-nil, zero value otherwise.

### GetGpuOverQuotaWeightOk

`func (o *ProjectCreateRequest) GetGpuOverQuotaWeightOk() (*float32, bool)`

GetGpuOverQuotaWeightOk returns a tuple with the GpuOverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuOverQuotaWeight

`func (o *ProjectCreateRequest) SetGpuOverQuotaWeight(v float32)`

SetGpuOverQuotaWeight sets GpuOverQuotaWeight field to given value.

### HasGpuOverQuotaWeight

`func (o *ProjectCreateRequest) HasGpuOverQuotaWeight() bool`

HasGpuOverQuotaWeight returns a boolean if a field has been set.

### GetDefaultNodePools

`func (o *ProjectCreateRequest) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *ProjectCreateRequest) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *ProjectCreateRequest) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *ProjectCreateRequest) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### GetInteractiveJobTimeLimitSecs

`func (o *ProjectCreateRequest) GetInteractiveJobTimeLimitSecs() float32`

GetInteractiveJobTimeLimitSecs returns the InteractiveJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetInteractiveJobTimeLimitSecsOk

`func (o *ProjectCreateRequest) GetInteractiveJobTimeLimitSecsOk() (*float32, bool)`

GetInteractiveJobTimeLimitSecsOk returns a tuple with the InteractiveJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobTimeLimitSecs

`func (o *ProjectCreateRequest) SetInteractiveJobTimeLimitSecs(v float32)`

SetInteractiveJobTimeLimitSecs sets InteractiveJobTimeLimitSecs field to given value.

### HasInteractiveJobTimeLimitSecs

`func (o *ProjectCreateRequest) HasInteractiveJobTimeLimitSecs() bool`

HasInteractiveJobTimeLimitSecs returns a boolean if a field has been set.

### GetInteractiveJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) GetInteractiveJobMaxIdleDurationSecs() float32`

GetInteractiveJobMaxIdleDurationSecs returns the InteractiveJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractiveJobMaxIdleDurationSecsOk

`func (o *ProjectCreateRequest) GetInteractiveJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractiveJobMaxIdleDurationSecsOk returns a tuple with the InteractiveJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) SetInteractiveJobMaxIdleDurationSecs(v float32)`

SetInteractiveJobMaxIdleDurationSecs sets InteractiveJobMaxIdleDurationSecs field to given value.

### HasInteractiveJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) HasInteractiveJobMaxIdleDurationSecs() bool`

HasInteractiveJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) GetInteractivePreemptibleJobMaxIdleDurationSecs() float32`

GetInteractivePreemptibleJobMaxIdleDurationSecs returns the InteractivePreemptibleJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractivePreemptibleJobMaxIdleDurationSecsOk

`func (o *ProjectCreateRequest) GetInteractivePreemptibleJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractivePreemptibleJobMaxIdleDurationSecsOk returns a tuple with the InteractivePreemptibleJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) SetInteractivePreemptibleJobMaxIdleDurationSecs(v float32)`

SetInteractivePreemptibleJobMaxIdleDurationSecs sets InteractivePreemptibleJobMaxIdleDurationSecs field to given value.

### HasInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) HasInteractivePreemptibleJobMaxIdleDurationSecs() bool`

HasInteractivePreemptibleJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetTrainingJobTimeLimitSecs

`func (o *ProjectCreateRequest) GetTrainingJobTimeLimitSecs() float32`

GetTrainingJobTimeLimitSecs returns the TrainingJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetTrainingJobTimeLimitSecsOk

`func (o *ProjectCreateRequest) GetTrainingJobTimeLimitSecsOk() (*float32, bool)`

GetTrainingJobTimeLimitSecsOk returns a tuple with the TrainingJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobTimeLimitSecs

`func (o *ProjectCreateRequest) SetTrainingJobTimeLimitSecs(v float32)`

SetTrainingJobTimeLimitSecs sets TrainingJobTimeLimitSecs field to given value.

### HasTrainingJobTimeLimitSecs

`func (o *ProjectCreateRequest) HasTrainingJobTimeLimitSecs() bool`

HasTrainingJobTimeLimitSecs returns a boolean if a field has been set.

### GetTrainingJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) GetTrainingJobMaxIdleDurationSecs() float32`

GetTrainingJobMaxIdleDurationSecs returns the TrainingJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetTrainingJobMaxIdleDurationSecsOk

`func (o *ProjectCreateRequest) GetTrainingJobMaxIdleDurationSecsOk() (*float32, bool)`

GetTrainingJobMaxIdleDurationSecsOk returns a tuple with the TrainingJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) SetTrainingJobMaxIdleDurationSecs(v float32)`

SetTrainingJobMaxIdleDurationSecs sets TrainingJobMaxIdleDurationSecs field to given value.

### HasTrainingJobMaxIdleDurationSecs

`func (o *ProjectCreateRequest) HasTrainingJobMaxIdleDurationSecs() bool`

HasTrainingJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetNodeAffinity

`func (o *ProjectCreateRequest) GetNodeAffinity() JobsNodeAffinity`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *ProjectCreateRequest) GetNodeAffinityOk() (*JobsNodeAffinity, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *ProjectCreateRequest) SetNodeAffinity(v JobsNodeAffinity)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *ProjectCreateRequest) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.

### GetPermissions

`func (o *ProjectCreateRequest) GetPermissions() ResourcePermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *ProjectCreateRequest) GetPermissionsOk() (*ResourcePermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *ProjectCreateRequest) SetPermissions(v ResourcePermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *ProjectCreateRequest) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetResources

`func (o *ProjectCreateRequest) GetResources() AssignedResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ProjectCreateRequest) GetResourcesOk() (*AssignedResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ProjectCreateRequest) SetResources(v AssignedResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *ProjectCreateRequest) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetName

`func (o *ProjectCreateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectCreateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectCreateRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDepartmentId

`func (o *ProjectCreateRequest) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *ProjectCreateRequest) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *ProjectCreateRequest) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.


### GetNodePoolsResources

`func (o *ProjectCreateRequest) GetNodePoolsResources() []NodePoolAssignedResourcesCreate`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *ProjectCreateRequest) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesCreate, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *ProjectCreateRequest) SetNodePoolsResources(v []NodePoolAssignedResourcesCreate)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *ProjectCreateRequest) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetNamespace

`func (o *ProjectCreateRequest) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectCreateRequest) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectCreateRequest) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ProjectCreateRequest) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


