# ProjectV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | [**[]NodePoolAssignedResourcesV1Response**](NodePoolAssignedResourcesV1Response.md) | Resources assigned to this Project per Node Pool. | 
**DeservedGpus** | **float32** | Deprecated. Use &#39;deserved&#39; for the relevant resource type under &#x60;NodePoolResources&#x60;. The project&#39;s deserved GPU allocation in case the cluster has those resources. | 
**MaxAllowedGpus** | **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. An upper limit for the amount of GPUs the project can get (Even if over quota is allowed and resources are available). | 
**GpuOverQuotaWeight** | **float32** | Deprecated. Instead, use &#x60;overQuotaWeight&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. The priority the project gets for over quota resources. | 
**DefaultNodePools** | Pointer to **[]string** | Default node pools list for workload submission for this project if a workload doesn&#39;t specify a node pools list. | [optional] 
**InteractiveJobTimeLimitSecs** | Pointer to **float32** | A limit (in seconds) for the duration of interactive jobs from this project. | [optional] 
**InteractiveJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that an interactive job can be idle before being terminated. | [optional] 
**InteractivePreemptibleJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that an interactive preemptible job can be idle before being terminated. | [optional] 
**TrainingJobTimeLimitSecs** | Pointer to **float32** | A limit (in seconds) for the duration of training jobs from this project. Available only from cluster version 2.12 | [optional] 
**TrainingJobMaxIdleDurationSecs** | Pointer to **float32** | Maximum duration (in seconds) that a training job can be idle before being terminated. | [optional] 
**NodeAffinity** | Pointer to [**NullableProjectV1NodeAffinityResponse**](ProjectV1NodeAffinityResponse.md) |  | [optional] 
**Permissions** | Pointer to [**ProjectV1ResponseCommonFieldsPermissions**](ProjectV1ResponseCommonFieldsPermissions.md) |  | [optional] 
**Resources** | [**ProjectV1ResponseCommonFieldsResources**](ProjectV1ResponseCommonFieldsResources.md) |  | 
**Name** | Pointer to **string** | Project name. | [optional] 
**Namespace** | **string** | The name of an existing namespace to use for the project in the cluster. Supported only for cluster versions 2.12 or higher. | 
**Id** | Pointer to **int32** | Project id. | [optional] [readonly] 
**DepartmentId** | Pointer to **int32** | ID of the department that owns the project. | [optional] 
**TenantId** | Pointer to **int32** | ID of the tenant where the project is located. | [optional] 
**ClusterUuid** | Pointer to **string** | ID of the cluster where the project is located. | [optional] 
**DepartmentName** | Pointer to **string** | Name of the department where the project is located. | [optional] 
**InteractiveNodeAffinity** | Pointer to **string** |  | [optional] 
**TrainNodeAffinity** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** | Creation date of the project. | [optional] 
**Status** | Pointer to [**ProjectV1AllOfStatus**](ProjectV1AllOfStatus.md) |  | [optional] 
**Phase** | Pointer to **string** | project&#39;s phase | [optional] 
**QuotaStatuses** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewProjectV1

`func NewProjectV1(nodePoolsResources []NodePoolAssignedResourcesV1Response, deservedGpus float32, maxAllowedGpus float32, gpuOverQuotaWeight float32, resources ProjectV1ResponseCommonFieldsResources, namespace string, ) *ProjectV1`

NewProjectV1 instantiates a new ProjectV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectV1WithDefaults

`func NewProjectV1WithDefaults() *ProjectV1`

NewProjectV1WithDefaults instantiates a new ProjectV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *ProjectV1) GetNodePoolsResources() []NodePoolAssignedResourcesV1Response`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *ProjectV1) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesV1Response, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *ProjectV1) SetNodePoolsResources(v []NodePoolAssignedResourcesV1Response)`

SetNodePoolsResources sets NodePoolsResources field to given value.


### GetDeservedGpus

`func (o *ProjectV1) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *ProjectV1) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *ProjectV1) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.


### GetMaxAllowedGpus

`func (o *ProjectV1) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *ProjectV1) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *ProjectV1) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.


### GetGpuOverQuotaWeight

`func (o *ProjectV1) GetGpuOverQuotaWeight() float32`

GetGpuOverQuotaWeight returns the GpuOverQuotaWeight field if non-nil, zero value otherwise.

### GetGpuOverQuotaWeightOk

`func (o *ProjectV1) GetGpuOverQuotaWeightOk() (*float32, bool)`

GetGpuOverQuotaWeightOk returns a tuple with the GpuOverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuOverQuotaWeight

`func (o *ProjectV1) SetGpuOverQuotaWeight(v float32)`

SetGpuOverQuotaWeight sets GpuOverQuotaWeight field to given value.


### GetDefaultNodePools

`func (o *ProjectV1) GetDefaultNodePools() []string`

GetDefaultNodePools returns the DefaultNodePools field if non-nil, zero value otherwise.

### GetDefaultNodePoolsOk

`func (o *ProjectV1) GetDefaultNodePoolsOk() (*[]string, bool)`

GetDefaultNodePoolsOk returns a tuple with the DefaultNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultNodePools

`func (o *ProjectV1) SetDefaultNodePools(v []string)`

SetDefaultNodePools sets DefaultNodePools field to given value.

### HasDefaultNodePools

`func (o *ProjectV1) HasDefaultNodePools() bool`

HasDefaultNodePools returns a boolean if a field has been set.

### GetInteractiveJobTimeLimitSecs

`func (o *ProjectV1) GetInteractiveJobTimeLimitSecs() float32`

GetInteractiveJobTimeLimitSecs returns the InteractiveJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetInteractiveJobTimeLimitSecsOk

`func (o *ProjectV1) GetInteractiveJobTimeLimitSecsOk() (*float32, bool)`

GetInteractiveJobTimeLimitSecsOk returns a tuple with the InteractiveJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobTimeLimitSecs

`func (o *ProjectV1) SetInteractiveJobTimeLimitSecs(v float32)`

SetInteractiveJobTimeLimitSecs sets InteractiveJobTimeLimitSecs field to given value.

### HasInteractiveJobTimeLimitSecs

`func (o *ProjectV1) HasInteractiveJobTimeLimitSecs() bool`

HasInteractiveJobTimeLimitSecs returns a boolean if a field has been set.

### GetInteractiveJobMaxIdleDurationSecs

`func (o *ProjectV1) GetInteractiveJobMaxIdleDurationSecs() float32`

GetInteractiveJobMaxIdleDurationSecs returns the InteractiveJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractiveJobMaxIdleDurationSecsOk

`func (o *ProjectV1) GetInteractiveJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractiveJobMaxIdleDurationSecsOk returns a tuple with the InteractiveJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveJobMaxIdleDurationSecs

`func (o *ProjectV1) SetInteractiveJobMaxIdleDurationSecs(v float32)`

SetInteractiveJobMaxIdleDurationSecs sets InteractiveJobMaxIdleDurationSecs field to given value.

### HasInteractiveJobMaxIdleDurationSecs

`func (o *ProjectV1) HasInteractiveJobMaxIdleDurationSecs() bool`

HasInteractiveJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectV1) GetInteractivePreemptibleJobMaxIdleDurationSecs() float32`

GetInteractivePreemptibleJobMaxIdleDurationSecs returns the InteractivePreemptibleJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetInteractivePreemptibleJobMaxIdleDurationSecsOk

`func (o *ProjectV1) GetInteractivePreemptibleJobMaxIdleDurationSecsOk() (*float32, bool)`

GetInteractivePreemptibleJobMaxIdleDurationSecsOk returns a tuple with the InteractivePreemptibleJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectV1) SetInteractivePreemptibleJobMaxIdleDurationSecs(v float32)`

SetInteractivePreemptibleJobMaxIdleDurationSecs sets InteractivePreemptibleJobMaxIdleDurationSecs field to given value.

### HasInteractivePreemptibleJobMaxIdleDurationSecs

`func (o *ProjectV1) HasInteractivePreemptibleJobMaxIdleDurationSecs() bool`

HasInteractivePreemptibleJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetTrainingJobTimeLimitSecs

`func (o *ProjectV1) GetTrainingJobTimeLimitSecs() float32`

GetTrainingJobTimeLimitSecs returns the TrainingJobTimeLimitSecs field if non-nil, zero value otherwise.

### GetTrainingJobTimeLimitSecsOk

`func (o *ProjectV1) GetTrainingJobTimeLimitSecsOk() (*float32, bool)`

GetTrainingJobTimeLimitSecsOk returns a tuple with the TrainingJobTimeLimitSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobTimeLimitSecs

`func (o *ProjectV1) SetTrainingJobTimeLimitSecs(v float32)`

SetTrainingJobTimeLimitSecs sets TrainingJobTimeLimitSecs field to given value.

### HasTrainingJobTimeLimitSecs

`func (o *ProjectV1) HasTrainingJobTimeLimitSecs() bool`

HasTrainingJobTimeLimitSecs returns a boolean if a field has been set.

### GetTrainingJobMaxIdleDurationSecs

`func (o *ProjectV1) GetTrainingJobMaxIdleDurationSecs() float32`

GetTrainingJobMaxIdleDurationSecs returns the TrainingJobMaxIdleDurationSecs field if non-nil, zero value otherwise.

### GetTrainingJobMaxIdleDurationSecsOk

`func (o *ProjectV1) GetTrainingJobMaxIdleDurationSecsOk() (*float32, bool)`

GetTrainingJobMaxIdleDurationSecsOk returns a tuple with the TrainingJobMaxIdleDurationSecs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainingJobMaxIdleDurationSecs

`func (o *ProjectV1) SetTrainingJobMaxIdleDurationSecs(v float32)`

SetTrainingJobMaxIdleDurationSecs sets TrainingJobMaxIdleDurationSecs field to given value.

### HasTrainingJobMaxIdleDurationSecs

`func (o *ProjectV1) HasTrainingJobMaxIdleDurationSecs() bool`

HasTrainingJobMaxIdleDurationSecs returns a boolean if a field has been set.

### GetNodeAffinity

`func (o *ProjectV1) GetNodeAffinity() ProjectV1NodeAffinityResponse`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *ProjectV1) GetNodeAffinityOk() (*ProjectV1NodeAffinityResponse, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *ProjectV1) SetNodeAffinity(v ProjectV1NodeAffinityResponse)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *ProjectV1) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.

### SetNodeAffinityNil

`func (o *ProjectV1) SetNodeAffinityNil(b bool)`

 SetNodeAffinityNil sets the value for NodeAffinity to be an explicit nil

### UnsetNodeAffinity
`func (o *ProjectV1) UnsetNodeAffinity()`

UnsetNodeAffinity ensures that no value is present for NodeAffinity, not even an explicit nil
### GetPermissions

`func (o *ProjectV1) GetPermissions() ProjectV1ResponseCommonFieldsPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *ProjectV1) GetPermissionsOk() (*ProjectV1ResponseCommonFieldsPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *ProjectV1) SetPermissions(v ProjectV1ResponseCommonFieldsPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *ProjectV1) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetResources

`func (o *ProjectV1) GetResources() ProjectV1ResponseCommonFieldsResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ProjectV1) GetResourcesOk() (*ProjectV1ResponseCommonFieldsResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ProjectV1) SetResources(v ProjectV1ResponseCommonFieldsResources)`

SetResources sets Resources field to given value.


### GetName

`func (o *ProjectV1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectV1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectV1) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProjectV1) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNamespace

`func (o *ProjectV1) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectV1) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectV1) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetId

`func (o *ProjectV1) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectV1) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectV1) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *ProjectV1) HasId() bool`

HasId returns a boolean if a field has been set.

### GetDepartmentId

`func (o *ProjectV1) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *ProjectV1) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *ProjectV1) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *ProjectV1) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetTenantId

`func (o *ProjectV1) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *ProjectV1) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *ProjectV1) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *ProjectV1) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetClusterUuid

`func (o *ProjectV1) GetClusterUuid() string`

GetClusterUuid returns the ClusterUuid field if non-nil, zero value otherwise.

### GetClusterUuidOk

`func (o *ProjectV1) GetClusterUuidOk() (*string, bool)`

GetClusterUuidOk returns a tuple with the ClusterUuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUuid

`func (o *ProjectV1) SetClusterUuid(v string)`

SetClusterUuid sets ClusterUuid field to given value.

### HasClusterUuid

`func (o *ProjectV1) HasClusterUuid() bool`

HasClusterUuid returns a boolean if a field has been set.

### GetDepartmentName

`func (o *ProjectV1) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *ProjectV1) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *ProjectV1) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.

### HasDepartmentName

`func (o *ProjectV1) HasDepartmentName() bool`

HasDepartmentName returns a boolean if a field has been set.

### GetInteractiveNodeAffinity

`func (o *ProjectV1) GetInteractiveNodeAffinity() string`

GetInteractiveNodeAffinity returns the InteractiveNodeAffinity field if non-nil, zero value otherwise.

### GetInteractiveNodeAffinityOk

`func (o *ProjectV1) GetInteractiveNodeAffinityOk() (*string, bool)`

GetInteractiveNodeAffinityOk returns a tuple with the InteractiveNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInteractiveNodeAffinity

`func (o *ProjectV1) SetInteractiveNodeAffinity(v string)`

SetInteractiveNodeAffinity sets InteractiveNodeAffinity field to given value.

### HasInteractiveNodeAffinity

`func (o *ProjectV1) HasInteractiveNodeAffinity() bool`

HasInteractiveNodeAffinity returns a boolean if a field has been set.

### GetTrainNodeAffinity

`func (o *ProjectV1) GetTrainNodeAffinity() string`

GetTrainNodeAffinity returns the TrainNodeAffinity field if non-nil, zero value otherwise.

### GetTrainNodeAffinityOk

`func (o *ProjectV1) GetTrainNodeAffinityOk() (*string, bool)`

GetTrainNodeAffinityOk returns a tuple with the TrainNodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTrainNodeAffinity

`func (o *ProjectV1) SetTrainNodeAffinity(v string)`

SetTrainNodeAffinity sets TrainNodeAffinity field to given value.

### HasTrainNodeAffinity

`func (o *ProjectV1) HasTrainNodeAffinity() bool`

HasTrainNodeAffinity returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ProjectV1) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ProjectV1) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ProjectV1) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ProjectV1) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetStatus

`func (o *ProjectV1) GetStatus() ProjectV1AllOfStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ProjectV1) GetStatusOk() (*ProjectV1AllOfStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ProjectV1) SetStatus(v ProjectV1AllOfStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ProjectV1) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetPhase

`func (o *ProjectV1) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *ProjectV1) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *ProjectV1) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *ProjectV1) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetQuotaStatuses

`func (o *ProjectV1) GetQuotaStatuses() string`

GetQuotaStatuses returns the QuotaStatuses field if non-nil, zero value otherwise.

### GetQuotaStatusesOk

`func (o *ProjectV1) GetQuotaStatusesOk() (*string, bool)`

GetQuotaStatusesOk returns a tuple with the QuotaStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuotaStatuses

`func (o *ProjectV1) SetQuotaStatuses(v string)`

SetQuotaStatuses sets QuotaStatuses field to given value.

### HasQuotaStatuses

`func (o *ProjectV1) HasQuotaStatuses() bool`

HasQuotaStatuses returns a boolean if a field has been set.

### SetQuotaStatusesNil

`func (o *ProjectV1) SetQuotaStatusesNil(b bool)`

 SetQuotaStatusesNil sets the value for QuotaStatuses to be an explicit nil

### UnsetQuotaStatuses
`func (o *ProjectV1) UnsetQuotaStatuses()`

UnsetQuotaStatuses ensures that no value is present for QuotaStatuses, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


