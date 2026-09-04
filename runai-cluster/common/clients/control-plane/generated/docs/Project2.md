# Project2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Project name. | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Project per Node Pool. | [optional] 
**Namespace** | Pointer to **string** | The name of an existing namespace to use for the project in the cluster. Supported only for cluster versions 2.12 or higher. | [optional] 
**Id** | Pointer to **int32** | Project id. | [optional] [readonly] 
**DepartmentId** | Pointer to **int32** | ID of the department that owns the project. | [optional] 
**TenantId** | Pointer to **int32** | ID of the tenant where the project is located. | [optional] 
**ClusterUuid** | Pointer to **string** | ID of the cluster where the project is located. | [optional] 
**DepartmentName** | Pointer to **string** | Name of the department where the project is located. | [optional] 
**InteractiveNodeAffinity** | Pointer to **string** |  | [optional] 
**TrainNodeAffinity** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** | Creation date of the project. | [optional] 
**Status** | Pointer to [**ProjectStatus**](ProjectStatus.md) |  | [optional] 
**Phase** | Pointer to **string** | project&#39;s phase | [optional] 
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

### NewProject2

`func NewProject2() *Project2`

NewProject2 instantiates a new Project2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProject2WithDefaults

`func NewProject2WithDefaults() *Project2`

NewProject2WithDefaults instantiates a new Project2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Project2) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Project2) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Project2) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Project2) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *Project2) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *Project2) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *Project2) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *Project2) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetNamespace

`func (o *Project2) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *Project2) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *Project2) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *Project2) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetId

`func (o *Project2) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Project2) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Project2) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *Project2) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDepartmentId

`func (o *Project2) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Project2) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Project2) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *Project2) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetTenantId

`func (o *Project2) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Project2) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Project2) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Project2) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *Project2) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *Project2) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *Project2) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *Project2) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetDepartmentName

`func (o *Project2) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *Project2) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *Project2) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *Project2) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetInteractiveNodeAffinity

`func (o *Project2) GetInteractiveNodeAffinity() string`

GetInteractiveNodeAffinity returns the InteractiveNodeAffinity field if non-nil, zero value otherwise.

### GetInteractiveNodeAffinityOk

`func (o *Project2) GetInteractiveNodeAffinityOk() (*string, bool)`

GetInteractiveNodeAffinityOk returns a tuple with the InteractiveNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveNodeAffinity

`func (o *Project2) SetInteractiveNodeAffinity(v string)`

SetInteractiveNodeAffinity sets InteractiveNodeAffinity field to given value.

### HasInteractiveNodeAffinity

`func (o *Project2) HasInteractiveNodeAffinity() bool`

HasInteractiveNodeAffinity returns a boolean if a field has been set.

### GetTrainNodeAffinity

`func (o *Project2) GetTrainNodeAffinity() string`

GetTrainNodeAffinity returns the TrainNodeAffinity field if non-nil, zero value otherwise.

### GetTrainNodeAffinityOk

`func (o *Project2) GetTrainNodeAffinityOk() (*string, bool)`

GetTrainNodeAffinityOk returns a tuple with the TrainNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainNodeAffinity

`func (o *Project2) SetTrainNodeAffinity(v string)`

SetTrainNodeAffinity sets TrainNodeAffinity field to given value.

### HasTrainNodeAffinity

`func (o *Project2) HasTrainNodeAffinity() bool`

HasTrainNodeAffinity returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Project2) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Project2) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Project2) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Project2) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetStatus

`func (o *Project2) GetStatus() ProjectStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Project2) GetStatusOk() (*ProjectStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Project2) SetStatus(v ProjectStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Project2) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetPhase

`func (o *Project2) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *Project2) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *Project2) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *Project2) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *Project2) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *Project2) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *Project2) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *Project2) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *Project2) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *Project2) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *Project2) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *Project2) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetGpuOverQuotaWeight

`func (o *Project2) GetGpuOverQuotaWeight() float32`

GetGpuOverQuotaWeight returns the GpuOverQuotaWeight field if non-nil, zero value otherwise.

### GetGpuOverQuotaWeightOk

`func (o *Project2) GetGpuOverQuotaWeightOk() (*float32, bool)`

GetGpuOverQuotaWeightOk returns a tuple with the GpuOverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuOverQuotaWeight

`func (o *Project2) SetGpuOverQuotaWeight(v float32)`

SetGpuOverQuotaWeight sets GpuOverQuotaWeight field to given value.

### HasGpuOverQuotaWeight

`func (o *Project2) HasGpuOverQuotaWeight() bool`

HasGpuOverQuotaWeight returns a boolean if a field has been set.

### GetDefaultNodePools

`func (o *Project2) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *Project2) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *Project2) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *Project2) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### GetInteractiveJobTimeLimitSecs

`func (o *Project2) GetInteractiveJobTimeLimitSecs() float32`

GetInteractiveJobTimeLimitSecs returns the InteractiveJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetInteractiveJobTimeLimitSecsOk

`func (o *Project2) GetInteractiveJobTimeLimitSecsOk() (*float32, bool)`

GetInteractiveJobTimeLimitSecsOk returns a tuple with the InteractiveJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobTimeLimitSecs

`func (o *Project2) SetInteractiveJobTimeLimitSecs(v float32)`

SetInteractiveJobTimeLimitSecs sets InteractiveJobTimeLimitSecs field to given value.

### HasInteractiveJobTimeLimitSecs

`func (o *Project2) HasInteractiveJobTimeLimitSecs() bool`

HasInteractiveJobTimeLimitSecs returns a boolean if a field has been set.

### GetInteractiveJobMaxIdleDurationSecs

`func (o *Project2) GetInteractiveJobMaxIdleDurationSecs() float32`

GetInteractiveJobMaxIdleDurationSecs returns the InteractiveJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractiveJobMaxIdleDurationSecsOk

`func (o *Project2) GetInteractiveJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractiveJobMaxIdleDurationSecsOk returns a tuple with the InteractiveJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobMaxIdleDurationSecs

`func (o *Project2) SetInteractiveJobMaxIdleDurationSecs(v float32)`

SetInteractiveJobMaxIdleDurationSecs sets InteractiveJobMaxIdleDurationSecs field to given value.

### HasInteractiveJobMaxIdleDurationSecs

`func (o *Project2) HasInteractiveJobMaxIdleDurationSecs() bool`

HasInteractiveJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *Project2) GetInteractivePreemptibleJobMaxIdleDurationSecs() float32`

GetInteractivePreemptibleJobMaxIdleDurationSecs returns the InteractivePreemptibleJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractivePreemptibleJobMaxIdleDurationSecsOk

`func (o *Project2) GetInteractivePreemptibleJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractivePreemptibleJobMaxIdleDurationSecsOk returns a tuple with the InteractivePreemptibleJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *Project2) SetInteractivePreemptibleJobMaxIdleDurationSecs(v float32)`

SetInteractivePreemptibleJobMaxIdleDurationSecs sets InteractivePreemptibleJobMaxIdleDurationSecs field to given value.

### HasInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *Project2) HasInteractivePreemptibleJobMaxIdleDurationSecs() bool`

HasInteractivePreemptibleJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetTrainingJobTimeLimitSecs

`func (o *Project2) GetTrainingJobTimeLimitSecs() float32`

GetTrainingJobTimeLimitSecs returns the TrainingJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetTrainingJobTimeLimitSecsOk

`func (o *Project2) GetTrainingJobTimeLimitSecsOk() (*float32, bool)`

GetTrainingJobTimeLimitSecsOk returns a tuple with the TrainingJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobTimeLimitSecs

`func (o *Project2) SetTrainingJobTimeLimitSecs(v float32)`

SetTrainingJobTimeLimitSecs sets TrainingJobTimeLimitSecs field to given value.

### HasTrainingJobTimeLimitSecs

`func (o *Project2) HasTrainingJobTimeLimitSecs() bool`

HasTrainingJobTimeLimitSecs returns a boolean if a field has been set.

### GetTrainingJobMaxIdleDurationSecs

`func (o *Project2) GetTrainingJobMaxIdleDurationSecs() float32`

GetTrainingJobMaxIdleDurationSecs returns the TrainingJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetTrainingJobMaxIdleDurationSecsOk

`func (o *Project2) GetTrainingJobMaxIdleDurationSecsOk() (*float32, bool)`

GetTrainingJobMaxIdleDurationSecsOk returns a tuple with the TrainingJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobMaxIdleDurationSecs

`func (o *Project2) SetTrainingJobMaxIdleDurationSecs(v float32)`

SetTrainingJobMaxIdleDurationSecs sets TrainingJobMaxIdleDurationSecs field to given value.

### HasTrainingJobMaxIdleDurationSecs

`func (o *Project2) HasTrainingJobMaxIdleDurationSecs() bool`

HasTrainingJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetNodeAffinity

`func (o *Project2) GetNodeAffinity() ProjectCommonFieldsNodeAffinity`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *Project2) GetNodeAffinityOk() (*ProjectCommonFieldsNodeAffinity, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *Project2) SetNodeAffinity(v ProjectCommonFieldsNodeAffinity)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *Project2) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.

### GetPermissions

`func (o *Project2) GetPermissions() ProjectCommonFieldsPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *Project2) GetPermissionsOk() (*ProjectCommonFieldsPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *Project2) SetPermissions(v ProjectCommonFieldsPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *Project2) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetResources

`func (o *Project2) GetResources() ProjectCommonFieldsResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *Project2) GetResourcesOk() (*ProjectCommonFieldsResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *Project2) SetResources(v ProjectCommonFieldsResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *Project2) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


