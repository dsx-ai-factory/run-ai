# Project

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Project Name | [optional] 
**DepartmentId** | Pointer to **int32** | ID of the department owning the project | [optional] 
**DeservedGpus** | Pointer to **float32** | The project GPU allocation that it is deserved to get in case the cluster has those resources | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | An upper limit for the amount of GPUs the project can get (Even if over quota is allowed and resources are available) | [optional] 
**GpuOverQuotaWeight** | Pointer to **float32** | The priority the project gets for over quota resources | [optional] 
**SwapEnabled** | Pointer to **bool** | Flag for enabling thin provisioning for jobs of this project | [optional] 
**DefaultNodePools** | Pointer to **[]string** | Default node pools for workload submission under this project (If a workload doesnt specify node pools list, this list will be used). | [optional] 
**InteractiveJobTimeLimitSecs** | Pointer to **float32** | A limit for the duration of interactive jobs from this project in seconds | [optional] 
**InteractiveJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum time duration in seconds that interactive job can be idle before it will be terminated | [optional] 
**InteractivePreemptibleJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum time duration in seconds that interactive preemptible job can be idle before it will be terminated | [optional] 
**TrainingJobTimeLimitSecs** | Pointer to **float32** | A limit for the duration of training jobs from this project in seconds. Available only from cluster version 2.12 | [optional] 
**TrainingJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum time duration in seconds that training job can be idle before it will be terminated | [optional] 
**NodeAffinity** | Pointer to [**JobsNodeAffinity**](JobsNodeAffinity.md) |  | [optional] 
**Permissions** | Pointer to [**ResourcePermissions**](ResourcePermissions.md) |  | [optional] 
**Resources** | Pointer to [**AssignedResources**](AssignedResources.md) |  | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Project per Node Pool | [optional] 
**Namespace** | Pointer to **string** | The name of an existing namespace to use for the project in the cluster. | [optional] 
**Id** | Pointer to **int32** | The project id | [optional] [readonly] 
**TenantId** | Pointer to **int32** | ID of the tenant owning the project | [optional] 
**ClusterUuid** | Pointer to **string** | ID of the cluster owning the project | [optional] 
**DepartmentName** | Pointer to **string** | Name of the department owning the project | [optional] 
**InteractiveNodeAffinity** | Pointer to **string** |  | [optional] 
**TrainNodeAffinity** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** | The creation date of the project | [optional] 
**Status** | Pointer to [**ProjectStatus**](ProjectStatus.md) |  | [optional] 
**Phase** | Pointer to **string** | project&#39;s phase | [optional] 

## Methods

### NewProject

`func NewProject() *Project`

NewProject instantiates a new Project object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectWithDefaults

`func NewProjectWithDefaults() *Project`

NewProjectWithDefaults instantiates a new Project object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Project) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Project) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Project) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Project) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDepartmentId

`func (o *Project) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Project) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Project) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *Project) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *Project) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *Project) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *Project) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *Project) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *Project) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *Project) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *Project) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *Project) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetGpuOverQuotaWeight

`func (o *Project) GetGpuOverQuotaWeight() float32`

GetGpuOverQuotaWeight returns the GpuOverQuotaWeight field if non-nil, zero value otherwise.

### GetGpuOverQuotaWeightOk

`func (o *Project) GetGpuOverQuotaWeightOk() (*float32, bool)`

GetGpuOverQuotaWeightOk returns a tuple with the GpuOverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuOverQuotaWeight

`func (o *Project) SetGpuOverQuotaWeight(v float32)`

SetGpuOverQuotaWeight sets GpuOverQuotaWeight field to given value.

### HasGpuOverQuotaWeight

`func (o *Project) HasGpuOverQuotaWeight() bool`

HasGpuOverQuotaWeight returns a boolean if a field has been set.

### GetSwapEnabled

`func (o *Project) GetSwapEnabled() bool`

GetSwapEnabled returns the SwapEnabled field if non-nil, zero value otherwise.

### GetSwapEnabledOk

`func (o *Project) GetSwapEnabledOk() (*bool, bool)`

GetSwapEnabledOk returns a tuple with the SwapEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSwapEnabled

`func (o *Project) SetSwapEnabled(v bool)`

SetSwapEnabled sets SwapEnabled field to given value.

### HasSwapEnabled

`func (o *Project) HasSwapEnabled() bool`

HasSwapEnabled returns a boolean if a field has been set.

### GetDefaultNodePools

`func (o *Project) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *Project) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *Project) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *Project) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### GetInteractiveJobTimeLimitSecs

`func (o *Project) GetInteractiveJobTimeLimitSecs() float32`

GetInteractiveJobTimeLimitSecs returns the InteractiveJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetInteractiveJobTimeLimitSecsOk

`func (o *Project) GetInteractiveJobTimeLimitSecsOk() (*float32, bool)`

GetInteractiveJobTimeLimitSecsOk returns a tuple with the InteractiveJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobTimeLimitSecs

`func (o *Project) SetInteractiveJobTimeLimitSecs(v float32)`

SetInteractiveJobTimeLimitSecs sets InteractiveJobTimeLimitSecs field to given value.

### HasInteractiveJobTimeLimitSecs

`func (o *Project) HasInteractiveJobTimeLimitSecs() bool`

HasInteractiveJobTimeLimitSecs returns a boolean if a field has been set.

### GetInteractiveJobMaxIdleDurationSecs

`func (o *Project) GetInteractiveJobMaxIdleDurationSecs() float32`

GetInteractiveJobMaxIdleDurationSecs returns the InteractiveJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractiveJobMaxIdleDurationSecsOk

`func (o *Project) GetInteractiveJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractiveJobMaxIdleDurationSecsOk returns a tuple with the InteractiveJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobMaxIdleDurationSecs

`func (o *Project) SetInteractiveJobMaxIdleDurationSecs(v float32)`

SetInteractiveJobMaxIdleDurationSecs sets InteractiveJobMaxIdleDurationSecs field to given value.

### HasInteractiveJobMaxIdleDurationSecs

`func (o *Project) HasInteractiveJobMaxIdleDurationSecs() bool`

HasInteractiveJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *Project) GetInteractivePreemptibleJobMaxIdleDurationSecs() float32`

GetInteractivePreemptibleJobMaxIdleDurationSecs returns the InteractivePreemptibleJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractivePreemptibleJobMaxIdleDurationSecsOk

`func (o *Project) GetInteractivePreemptibleJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractivePreemptibleJobMaxIdleDurationSecsOk returns a tuple with the InteractivePreemptibleJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *Project) SetInteractivePreemptibleJobMaxIdleDurationSecs(v float32)`

SetInteractivePreemptibleJobMaxIdleDurationSecs sets InteractivePreemptibleJobMaxIdleDurationSecs field to given value.

### HasInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *Project) HasInteractivePreemptibleJobMaxIdleDurationSecs() bool`

HasInteractivePreemptibleJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetTrainingJobTimeLimitSecs

`func (o *Project) GetTrainingJobTimeLimitSecs() float32`

GetTrainingJobTimeLimitSecs returns the TrainingJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetTrainingJobTimeLimitSecsOk

`func (o *Project) GetTrainingJobTimeLimitSecsOk() (*float32, bool)`

GetTrainingJobTimeLimitSecsOk returns a tuple with the TrainingJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobTimeLimitSecs

`func (o *Project) SetTrainingJobTimeLimitSecs(v float32)`

SetTrainingJobTimeLimitSecs sets TrainingJobTimeLimitSecs field to given value.

### HasTrainingJobTimeLimitSecs

`func (o *Project) HasTrainingJobTimeLimitSecs() bool`

HasTrainingJobTimeLimitSecs returns a boolean if a field has been set.

### GetTrainingJobMaxIdleDurationSecs

`func (o *Project) GetTrainingJobMaxIdleDurationSecs() float32`

GetTrainingJobMaxIdleDurationSecs returns the TrainingJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetTrainingJobMaxIdleDurationSecsOk

`func (o *Project) GetTrainingJobMaxIdleDurationSecsOk() (*float32, bool)`

GetTrainingJobMaxIdleDurationSecsOk returns a tuple with the TrainingJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobMaxIdleDurationSecs

`func (o *Project) SetTrainingJobMaxIdleDurationSecs(v float32)`

SetTrainingJobMaxIdleDurationSecs sets TrainingJobMaxIdleDurationSecs field to given value.

### HasTrainingJobMaxIdleDurationSecs

`func (o *Project) HasTrainingJobMaxIdleDurationSecs() bool`

HasTrainingJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetNodeAffinity

`func (o *Project) GetNodeAffinity() JobsNodeAffinity`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *Project) GetNodeAffinityOk() (*JobsNodeAffinity, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *Project) SetNodeAffinity(v JobsNodeAffinity)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *Project) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.

### GetPermissions

`func (o *Project) GetPermissions() ResourcePermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *Project) GetPermissionsOk() (*ResourcePermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *Project) SetPermissions(v ResourcePermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *Project) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetResources

`func (o *Project) GetResources() AssignedResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *Project) GetResourcesOk() (*AssignedResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *Project) SetResources(v AssignedResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *Project) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *Project) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *Project) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *Project) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *Project) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetNamespace

`func (o *Project) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *Project) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *Project) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *Project) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetId

`func (o *Project) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Project) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Project) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *Project) HasId() bool`

HasId returns a boolean if a field has been set.

### GetTenantId

`func (o *Project) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Project) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Project) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Project) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *Project) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *Project) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *Project) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *Project) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetDepartmentName

`func (o *Project) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *Project) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *Project) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *Project) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetInteractiveNodeAffinity

`func (o *Project) GetInteractiveNodeAffinity() string`

GetInteractiveNodeAffinity returns the InteractiveNodeAffinity field if non-nil, zero value otherwise.

### GetInteractiveNodeAffinityOk

`func (o *Project) GetInteractiveNodeAffinityOk() (*string, bool)`

GetInteractiveNodeAffinityOk returns a tuple with the InteractiveNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveNodeAffinity

`func (o *Project) SetInteractiveNodeAffinity(v string)`

SetInteractiveNodeAffinity sets InteractiveNodeAffinity field to given value.

### HasInteractiveNodeAffinity

`func (o *Project) HasInteractiveNodeAffinity() bool`

HasInteractiveNodeAffinity returns a boolean if a field has been set.

### GetTrainNodeAffinity

`func (o *Project) GetTrainNodeAffinity() string`

GetTrainNodeAffinity returns the TrainNodeAffinity field if non-nil, zero value otherwise.

### GetTrainNodeAffinityOk

`func (o *Project) GetTrainNodeAffinityOk() (*string, bool)`

GetTrainNodeAffinityOk returns a tuple with the TrainNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainNodeAffinity

`func (o *Project) SetTrainNodeAffinity(v string)`

SetTrainNodeAffinity sets TrainNodeAffinity field to given value.

### HasTrainNodeAffinity

`func (o *Project) HasTrainNodeAffinity() bool`

HasTrainNodeAffinity returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Project) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Project) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Project) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Project) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetStatus

`func (o *Project) GetStatus() ProjectStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Project) GetStatusOk() (*ProjectStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Project) SetStatus(v ProjectStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Project) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetPhase

`func (o *Project) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *Project) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *Project) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *Project) HasPhase() bool`

HasPhase returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


