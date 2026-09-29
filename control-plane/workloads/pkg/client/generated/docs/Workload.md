# Workload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TenantId** | **int32** | The id of the tenant. | 
**RunningPods** | **int32** |  | 
**PhaseUpdatedAt** | **time.Time** |  | 
**K8sPhaseUpdatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**Source** | [**Source**](Source.md) |  | 
**DeletedAt** | **NullableTime** |  | 
**Type** | **string** |  | 
**Name** | **string** |  | 
**ExternalWorkloadName** | Pointer to **string** |  | [optional] 
**Id** | **string** |  | 
**Priority** | **NullableInt32** |  | 
**PriorityClassName** | **string** |  | 
**SubmittedBy** | Pointer to **string** |  | [optional] 
**ClusterId** | **string** | The id of the cluster. | 
**ProjectName** | **string** |  | 
**ProjectId** | **string** |  | 
**DepartmentName** | **string** |  | 
**DepartmentId** | **string** |  | 
**Namespace** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**WorkloadRequestedResources** | Pointer to [**NullableWorkloadRequestResources**](WorkloadRequestResources.md) |  | [optional] 
**PodsRequestedResources** | Pointer to [**NullableWorkloadRequestResources**](WorkloadRequestResources.md) |  | [optional] 
**AllocatedResources** | Pointer to [**NullableWorkloadAllocatedResources**](WorkloadAllocatedResources.md) |  | [optional] 
**ActionsSupport** | Pointer to [**ActionsSupport**](ActionsSupport.md) |  | [optional] 
**Phase** | [**Phase**](Phase.md) |  | 
**Conditions** | [**[]Condition**](Condition.md) |  | 
**PhaseMessage** | Pointer to **string** |  | [optional] 
**K8sPhase** | **string** |  | 
**RequestedPods** | Pointer to [**RequestedPods**](RequestedPods.md) |  | [optional] 
**RequestedNodePools** | Pointer to **[]string** |  | [optional] 
**CurrentNodePools** | Pointer to **[]string** |  | [optional] 
**CompletedAt** | Pointer to **NullableTime** |  | [optional] 
**Images** | Pointer to **[]string** |  | [optional] 
**ChildrenIds** | Pointer to [**[]WorkloadChildrenIdsInner**](WorkloadChildrenIdsInner.md) |  | [optional] 
**Urls** | Pointer to **[]string** | **DEPRECATED:** This field is deprecated. Use [Get workload endpoints](/latest/#operation/get_workload_endpoints) to retrieve workload network endpoints instead.Requires cluster version 2.25 or above. | [optional] 
**Datasources** | Pointer to [**[]Datasource**](Datasource.md) |  | [optional] 
**Environments** | Pointer to [**[]Environment**](Environment.md) |  | [optional] 
**ExternalConnections** | Pointer to [**[]Connection**](Connection.md) | **DEPRECATED:** This field is deprecated. Use [Get workload endpoints](/latest/#operation/get_workload_endpoints) to retrieve workload network endpoints instead.Requires cluster version 2.25 or above. | [optional] 
**DistributedFramework** | Pointer to **string** |  | [optional] 
**AdditionalFields** | Pointer to **map[string]interface{}** |  | [optional] 
**Preemptible** | Pointer to **NullableBool** |  | [optional] 
**Preemptibility** | Pointer to **NullableString** |  | [optional] 
**FullySaasWorkload** | Pointer to **bool** |  | [optional] 
**EnvironmentVariables** | Pointer to **map[string]string** |  | [optional] 
**Command** | Pointer to **string** |  | [optional] 
**Arguments** | Pointer to **string** |  | [optional] 
**PhaseReason** | Pointer to [**NullablePhaseReason**](PhaseReason.md) |  | [optional] 
**IdleAllocatedGpus** | Pointer to **NullableFloat32** | sum of idle allocated gpus in the workload | [optional] 
**IdleSince** | Pointer to **NullableTime** | The timestamp when the workload became idle. Null if the workload is not currently idle. | [optional] 
**TotalPendingTimeSeconds** | Pointer to **NullableInt32** | The total cumulative time, in seconds, that the workload has spent in the Pending phase since submission. | [optional] 
**TotalRunningTimeSeconds** | Pointer to **NullableInt32** | The total cumulative time, in seconds, that the workload has spent in the Running phase since submission. | [optional] 
**Category** | **string** | Category Description | 
**GuaranteedRuntimeEndsAt** | Pointer to **NullableTime** | A timestamp indicating when the workload will reach its minimum guaranteed runtime, as defined by minGuaranteedRuntime. Until this time, the workload is considered non-preemptible and cannot be interrupted by higher-priority workloads. | [optional] 
**AiApplicationId** | Pointer to **string** |  | [optional] 
**AiApplicationName** | Pointer to **string** |  | [optional] 
**SourceApi** | Pointer to [**NullableSourceApi**](SourceApi.md) |  | [optional] 
**Topology** | Pointer to [**NullableWorkloadTopology**](WorkloadTopology.md) |  | [optional] 
**Endpoints** | Pointer to [**[]Endpoint**](Endpoint.md) | List of network endpoints through which the workload is exposed. | [optional] 
**NodeAffinities** | Pointer to [**[]WorkloadNodeAffinity**](WorkloadNodeAffinity.md) | The node affinities of a workload, which is an array of node affinity rules for different specs (e.g., worker, leader, master). Each item in the array specifies the node affinity for a particular spec. | [optional] 
**KubernetesRef** | Pointer to [**NullableWorkloadKubernetesRef**](WorkloadKubernetesRef.md) |  | [optional] 

## Methods

### NewWorkload

`func NewWorkload(tenantId int32, runningPods int32, phaseUpdatedAt time.Time, k8sPhaseUpdatedAt time.Time, updatedAt time.Time, source Source, deletedAt NullableTime, type_ string, name string, id string, priority NullableInt32, priorityClassName string, clusterId string, projectName string, projectId string, departmentName string, departmentId string, namespace string, createdAt time.Time, phase Phase, conditions []Condition, k8sPhase string, category string, ) *Workload`

NewWorkload instantiates a new Workload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadWithDefaults

`func NewWorkloadWithDefaults() *Workload`

NewWorkloadWithDefaults instantiates a new Workload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenantId

`func (o *Workload) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Workload) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Workload) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetRunningPods

`func (o *Workload) GetRunningPods() int32`

GetRunningPods returns the RunningPods field if non-nil, zero value otherwise.

### GetRunningPodsOk

`func (o *Workload) GetRunningPodsOk() (*int32, bool)`

GetRunningPodsOk returns a tuple with the RunningPods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunningPods

`func (o *Workload) SetRunningPods(v int32)`

SetRunningPods sets RunningPods field to given value.


### GetPhaseUpdatedAt

`func (o *Workload) GetPhaseUpdatedAt() time.Time`

GetPhaseUpdatedAt returns the PhaseUpdatedAt field if non-nil, zero value otherwise.

### GetPhaseUpdatedAtOk

`func (o *Workload) GetPhaseUpdatedAtOk() (*time.Time, bool)`

GetPhaseUpdatedAtOk returns a tuple with the PhaseUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseUpdatedAt

`func (o *Workload) SetPhaseUpdatedAt(v time.Time)`

SetPhaseUpdatedAt sets PhaseUpdatedAt field to given value.


### GetK8sPhaseUpdatedAt

`func (o *Workload) GetK8sPhaseUpdatedAt() time.Time`

GetK8sPhaseUpdatedAt returns the K8sPhaseUpdatedAt field if non-nil, zero value otherwise.

### GetK8sPhaseUpdatedAtOk

`func (o *Workload) GetK8sPhaseUpdatedAtOk() (*time.Time, bool)`

GetK8sPhaseUpdatedAtOk returns a tuple with the K8sPhaseUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhaseUpdatedAt

`func (o *Workload) SetK8sPhaseUpdatedAt(v time.Time)`

SetK8sPhaseUpdatedAt sets K8sPhaseUpdatedAt field to given value.


### GetUpdatedAt

`func (o *Workload) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Workload) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Workload) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetSource

`func (o *Workload) GetSource() Source`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *Workload) GetSourceOk() (*Source, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *Workload) SetSource(v Source)`

SetSource sets Source field to given value.


### GetDeletedAt

`func (o *Workload) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *Workload) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *Workload) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.


### SetDeletedAtNil

`func (o *Workload) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *Workload) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetType

`func (o *Workload) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Workload) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Workload) SetType(v string)`

SetType sets Type field to given value.


### GetName

`func (o *Workload) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Workload) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Workload) SetName(v string)`

SetName sets Name field to given value.


### GetExternalWorkloadName

`func (o *Workload) GetExternalWorkloadName() string`

GetExternalWorkloadName returns the ExternalWorkloadName field if non-nil, zero value otherwise.

### GetExternalWorkloadNameOk

`func (o *Workload) GetExternalWorkloadNameOk() (*string, bool)`

GetExternalWorkloadNameOk returns a tuple with the ExternalWorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalWorkloadName

`func (o *Workload) SetExternalWorkloadName(v string)`

SetExternalWorkloadName sets ExternalWorkloadName field to given value.

### HasExternalWorkloadName

`func (o *Workload) HasExternalWorkloadName() bool`

HasExternalWorkloadName returns a boolean if a field has been set.

### GetId

`func (o *Workload) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Workload) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Workload) SetId(v string)`

SetId sets Id field to given value.


### GetPriority

`func (o *Workload) GetPriority() int32`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *Workload) GetPriorityOk() (*int32, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *Workload) SetPriority(v int32)`

SetPriority sets Priority field to given value.


### SetPriorityNil

`func (o *Workload) SetPriorityNil(b bool)`

 SetPriorityNil sets the value for Priority to be an explicit nil

### UnsetPriority
`func (o *Workload) UnsetPriority()`

UnsetPriority ensures that no value is present for Priority, not even an explicit nil
### GetPriorityClassName

`func (o *Workload) GetPriorityClassName() string`

GetPriorityClassName returns the PriorityClassName field if non-nil, zero value otherwise.

### GetPriorityClassNameOk

`func (o *Workload) GetPriorityClassNameOk() (*string, bool)`

GetPriorityClassNameOk returns a tuple with the PriorityClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClassName

`func (o *Workload) SetPriorityClassName(v string)`

SetPriorityClassName sets PriorityClassName field to given value.


### GetSubmittedBy

`func (o *Workload) GetSubmittedBy() string`

GetSubmittedBy returns the SubmittedBy field if non-nil, zero value otherwise.

### GetSubmittedByOk

`func (o *Workload) GetSubmittedByOk() (*string, bool)`

GetSubmittedByOk returns a tuple with the SubmittedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubmittedBy

`func (o *Workload) SetSubmittedBy(v string)`

SetSubmittedBy sets SubmittedBy field to given value.

### HasSubmittedBy

`func (o *Workload) HasSubmittedBy() bool`

HasSubmittedBy returns a boolean if a field has been set.

### GetClusterId

`func (o *Workload) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Workload) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Workload) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectName

`func (o *Workload) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *Workload) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *Workload) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetProjectId

`func (o *Workload) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Workload) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Workload) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetDepartmentName

`func (o *Workload) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *Workload) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *Workload) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.


### GetDepartmentId

`func (o *Workload) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Workload) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Workload) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.


### GetNamespace

`func (o *Workload) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *Workload) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *Workload) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetCreatedAt

`func (o *Workload) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Workload) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Workload) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetWorkloadRequestedResources

`func (o *Workload) GetWorkloadRequestedResources() WorkloadRequestResources`

GetWorkloadRequestedResources returns the WorkloadRequestedResources field if non-nil, zero value otherwise.

### GetWorkloadRequestedResourcesOk

`func (o *Workload) GetWorkloadRequestedResourcesOk() (*WorkloadRequestResources, bool)`

GetWorkloadRequestedResourcesOk returns a tuple with the WorkloadRequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadRequestedResources

`func (o *Workload) SetWorkloadRequestedResources(v WorkloadRequestResources)`

SetWorkloadRequestedResources sets WorkloadRequestedResources field to given value.

### HasWorkloadRequestedResources

`func (o *Workload) HasWorkloadRequestedResources() bool`

HasWorkloadRequestedResources returns a boolean if a field has been set.

### SetWorkloadRequestedResourcesNil

`func (o *Workload) SetWorkloadRequestedResourcesNil(b bool)`

 SetWorkloadRequestedResourcesNil sets the value for WorkloadRequestedResources to be an explicit nil

### UnsetWorkloadRequestedResources
`func (o *Workload) UnsetWorkloadRequestedResources()`

UnsetWorkloadRequestedResources ensures that no value is present for WorkloadRequestedResources, not even an explicit nil
### GetPodsRequestedResources

`func (o *Workload) GetPodsRequestedResources() WorkloadRequestResources`

GetPodsRequestedResources returns the PodsRequestedResources field if non-nil, zero value otherwise.

### GetPodsRequestedResourcesOk

`func (o *Workload) GetPodsRequestedResourcesOk() (*WorkloadRequestResources, bool)`

GetPodsRequestedResourcesOk returns a tuple with the PodsRequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodsRequestedResources

`func (o *Workload) SetPodsRequestedResources(v WorkloadRequestResources)`

SetPodsRequestedResources sets PodsRequestedResources field to given value.

### HasPodsRequestedResources

`func (o *Workload) HasPodsRequestedResources() bool`

HasPodsRequestedResources returns a boolean if a field has been set.

### SetPodsRequestedResourcesNil

`func (o *Workload) SetPodsRequestedResourcesNil(b bool)`

 SetPodsRequestedResourcesNil sets the value for PodsRequestedResources to be an explicit nil

### UnsetPodsRequestedResources
`func (o *Workload) UnsetPodsRequestedResources()`

UnsetPodsRequestedResources ensures that no value is present for PodsRequestedResources, not even an explicit nil
### GetAllocatedResources

`func (o *Workload) GetAllocatedResources() WorkloadAllocatedResources`

GetAllocatedResources returns the AllocatedResources field if non-nil, zero value otherwise.

### GetAllocatedResourcesOk

`func (o *Workload) GetAllocatedResourcesOk() (*WorkloadAllocatedResources, bool)`

GetAllocatedResourcesOk returns a tuple with the AllocatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedResources

`func (o *Workload) SetAllocatedResources(v WorkloadAllocatedResources)`

SetAllocatedResources sets AllocatedResources field to given value.

### HasAllocatedResources

`func (o *Workload) HasAllocatedResources() bool`

HasAllocatedResources returns a boolean if a field has been set.

### SetAllocatedResourcesNil

`func (o *Workload) SetAllocatedResourcesNil(b bool)`

 SetAllocatedResourcesNil sets the value for AllocatedResources to be an explicit nil

### UnsetAllocatedResources
`func (o *Workload) UnsetAllocatedResources()`

UnsetAllocatedResources ensures that no value is present for AllocatedResources, not even an explicit nil
### GetActionsSupport

`func (o *Workload) GetActionsSupport() ActionsSupport`

GetActionsSupport returns the ActionsSupport field if non-nil, zero value otherwise.

### GetActionsSupportOk

`func (o *Workload) GetActionsSupportOk() (*ActionsSupport, bool)`

GetActionsSupportOk returns a tuple with the ActionsSupport field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionsSupport

`func (o *Workload) SetActionsSupport(v ActionsSupport)`

SetActionsSupport sets ActionsSupport field to given value.

### HasActionsSupport

`func (o *Workload) HasActionsSupport() bool`

HasActionsSupport returns a boolean if a field has been set.

### GetPhase

`func (o *Workload) GetPhase() Phase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *Workload) GetPhaseOk() (*Phase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *Workload) SetPhase(v Phase)`

SetPhase sets Phase field to given value.


### GetConditions

`func (o *Workload) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *Workload) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *Workload) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.


### GetPhaseMessage

`func (o *Workload) GetPhaseMessage() string`

GetPhaseMessage returns the PhaseMessage field if non-nil, zero value otherwise.

### GetPhaseMessageOk

`func (o *Workload) GetPhaseMessageOk() (*string, bool)`

GetPhaseMessageOk returns a tuple with the PhaseMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseMessage

`func (o *Workload) SetPhaseMessage(v string)`

SetPhaseMessage sets PhaseMessage field to given value.

### HasPhaseMessage

`func (o *Workload) HasPhaseMessage() bool`

HasPhaseMessage returns a boolean if a field has been set.

### GetK8sPhase

`func (o *Workload) GetK8sPhase() string`

GetK8sPhase returns the K8sPhase field if non-nil, zero value otherwise.

### GetK8sPhaseOk

`func (o *Workload) GetK8sPhaseOk() (*string, bool)`

GetK8sPhaseOk returns a tuple with the K8sPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhase

`func (o *Workload) SetK8sPhase(v string)`

SetK8sPhase sets K8sPhase field to given value.


### GetRequestedPods

`func (o *Workload) GetRequestedPods() RequestedPods`

GetRequestedPods returns the RequestedPods field if non-nil, zero value otherwise.

### GetRequestedPodsOk

`func (o *Workload) GetRequestedPodsOk() (*RequestedPods, bool)`

GetRequestedPodsOk returns a tuple with the RequestedPods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedPods

`func (o *Workload) SetRequestedPods(v RequestedPods)`

SetRequestedPods sets RequestedPods field to given value.

### HasRequestedPods

`func (o *Workload) HasRequestedPods() bool`

HasRequestedPods returns a boolean if a field has been set.

### GetRequestedNodePools

`func (o *Workload) GetRequestedNodePools() []string`

GetRequestedNodePools returns the RequestedNodePools field if non-nil, zero value otherwise.

### GetRequestedNodePoolsOk

`func (o *Workload) GetRequestedNodePoolsOk() (*[]string, bool)`

GetRequestedNodePoolsOk returns a tuple with the RequestedNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNodePools

`func (o *Workload) SetRequestedNodePools(v []string)`

SetRequestedNodePools sets RequestedNodePools field to given value.

### HasRequestedNodePools

`func (o *Workload) HasRequestedNodePools() bool`

HasRequestedNodePools returns a boolean if a field has been set.

### GetCurrentNodePools

`func (o *Workload) GetCurrentNodePools() []string`

GetCurrentNodePools returns the CurrentNodePools field if non-nil, zero value otherwise.

### GetCurrentNodePoolsOk

`func (o *Workload) GetCurrentNodePoolsOk() (*[]string, bool)`

GetCurrentNodePoolsOk returns a tuple with the CurrentNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentNodePools

`func (o *Workload) SetCurrentNodePools(v []string)`

SetCurrentNodePools sets CurrentNodePools field to given value.

### HasCurrentNodePools

`func (o *Workload) HasCurrentNodePools() bool`

HasCurrentNodePools returns a boolean if a field has been set.

### GetCompletedAt

`func (o *Workload) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *Workload) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *Workload) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *Workload) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.

### SetCompletedAtNil

`func (o *Workload) SetCompletedAtNil(b bool)`

 SetCompletedAtNil sets the value for CompletedAt to be an explicit nil

### UnsetCompletedAt
`func (o *Workload) UnsetCompletedAt()`

UnsetCompletedAt ensures that no value is present for CompletedAt, not even an explicit nil
### GetImages

`func (o *Workload) GetImages() []string`

GetImages returns the Images field if non-nil, zero value otherwise.

### GetImagesOk

`func (o *Workload) GetImagesOk() (*[]string, bool)`

GetImagesOk returns a tuple with the Images field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImages

`func (o *Workload) SetImages(v []string)`

SetImages sets Images field to given value.

### HasImages

`func (o *Workload) HasImages() bool`

HasImages returns a boolean if a field has been set.

### GetChildrenIds

`func (o *Workload) GetChildrenIds() []WorkloadChildrenIdsInner`

GetChildrenIds returns the ChildrenIds field if non-nil, zero value otherwise.

### GetChildrenIdsOk

`func (o *Workload) GetChildrenIdsOk() (*[]WorkloadChildrenIdsInner, bool)`

GetChildrenIdsOk returns a tuple with the ChildrenIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChildrenIds

`func (o *Workload) SetChildrenIds(v []WorkloadChildrenIdsInner)`

SetChildrenIds sets ChildrenIds field to given value.

### HasChildrenIds

`func (o *Workload) HasChildrenIds() bool`

HasChildrenIds returns a boolean if a field has been set.

### GetUrls

`func (o *Workload) GetUrls() []string`

GetUrls returns the Urls field if non-nil, zero value otherwise.

### GetUrlsOk

`func (o *Workload) GetUrlsOk() (*[]string, bool)`

GetUrlsOk returns a tuple with the Urls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrls

`func (o *Workload) SetUrls(v []string)`

SetUrls sets Urls field to given value.

### HasUrls

`func (o *Workload) HasUrls() bool`

HasUrls returns a boolean if a field has been set.

### GetDatasources

`func (o *Workload) GetDatasources() []Datasource`

GetDatasources returns the Datasources field if non-nil, zero value otherwise.

### GetDatasourcesOk

`func (o *Workload) GetDatasourcesOk() (*[]Datasource, bool)`

GetDatasourcesOk returns a tuple with the Datasources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatasources

`func (o *Workload) SetDatasources(v []Datasource)`

SetDatasources sets Datasources field to given value.

### HasDatasources

`func (o *Workload) HasDatasources() bool`

HasDatasources returns a boolean if a field has been set.

### GetEnvironments

`func (o *Workload) GetEnvironments() []Environment`

GetEnvironments returns the Environments field if non-nil, zero value otherwise.

### GetEnvironmentsOk

`func (o *Workload) GetEnvironmentsOk() (*[]Environment, bool)`

GetEnvironmentsOk returns a tuple with the Environments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironments

`func (o *Workload) SetEnvironments(v []Environment)`

SetEnvironments sets Environments field to given value.

### HasEnvironments

`func (o *Workload) HasEnvironments() bool`

HasEnvironments returns a boolean if a field has been set.

### GetExternalConnections

`func (o *Workload) GetExternalConnections() []Connection`

GetExternalConnections returns the ExternalConnections field if non-nil, zero value otherwise.

### GetExternalConnectionsOk

`func (o *Workload) GetExternalConnectionsOk() (*[]Connection, bool)`

GetExternalConnectionsOk returns a tuple with the ExternalConnections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalConnections

`func (o *Workload) SetExternalConnections(v []Connection)`

SetExternalConnections sets ExternalConnections field to given value.

### HasExternalConnections

`func (o *Workload) HasExternalConnections() bool`

HasExternalConnections returns a boolean if a field has been set.

### GetDistributedFramework

`func (o *Workload) GetDistributedFramework() string`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *Workload) GetDistributedFrameworkOk() (*string, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *Workload) SetDistributedFramework(v string)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *Workload) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### GetAdditionalFields

`func (o *Workload) GetAdditionalFields() map[string]interface{}`

GetAdditionalFields returns the AdditionalFields field if non-nil, zero value otherwise.

### GetAdditionalFieldsOk

`func (o *Workload) GetAdditionalFieldsOk() (*map[string]interface{}, bool)`

GetAdditionalFieldsOk returns a tuple with the AdditionalFields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdditionalFields

`func (o *Workload) SetAdditionalFields(v map[string]interface{})`

SetAdditionalFields sets AdditionalFields field to given value.

### HasAdditionalFields

`func (o *Workload) HasAdditionalFields() bool`

HasAdditionalFields returns a boolean if a field has been set.

### GetPreemptible

`func (o *Workload) GetPreemptible() bool`

GetPreemptible returns the Preemptible field if non-nil, zero value otherwise.

### GetPreemptibleOk

`func (o *Workload) GetPreemptibleOk() (*bool, bool)`

GetPreemptibleOk returns a tuple with the Preemptible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptible

`func (o *Workload) SetPreemptible(v bool)`

SetPreemptible sets Preemptible field to given value.

### HasPreemptible

`func (o *Workload) HasPreemptible() bool`

HasPreemptible returns a boolean if a field has been set.

### SetPreemptibleNil

`func (o *Workload) SetPreemptibleNil(b bool)`

 SetPreemptibleNil sets the value for Preemptible to be an explicit nil

### UnsetPreemptible
`func (o *Workload) UnsetPreemptible()`

UnsetPreemptible ensures that no value is present for Preemptible, not even an explicit nil
### GetPreemptibility

`func (o *Workload) GetPreemptibility() string`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *Workload) GetPreemptibilityOk() (*string, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *Workload) SetPreemptibility(v string)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *Workload) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *Workload) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *Workload) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetFullySaasWorkload

`func (o *Workload) GetFullySaasWorkload() bool`

GetFullySaasWorkload returns the FullySaasWorkload field if non-nil, zero value otherwise.

### GetFullySaasWorkloadOk

`func (o *Workload) GetFullySaasWorkloadOk() (*bool, bool)`

GetFullySaasWorkloadOk returns a tuple with the FullySaasWorkload field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullySaasWorkload

`func (o *Workload) SetFullySaasWorkload(v bool)`

SetFullySaasWorkload sets FullySaasWorkload field to given value.

### HasFullySaasWorkload

`func (o *Workload) HasFullySaasWorkload() bool`

HasFullySaasWorkload returns a boolean if a field has been set.

### GetEnvironmentVariables

`func (o *Workload) GetEnvironmentVariables() map[string]string`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *Workload) GetEnvironmentVariablesOk() (*map[string]string, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *Workload) SetEnvironmentVariables(v map[string]string)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *Workload) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### GetCommand

`func (o *Workload) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *Workload) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *Workload) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *Workload) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### GetArguments

`func (o *Workload) GetArguments() string`

GetArguments returns the Arguments field if non-nil, zero value otherwise.

### GetArgumentsOk

`func (o *Workload) GetArgumentsOk() (*string, bool)`

GetArgumentsOk returns a tuple with the Arguments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArguments

`func (o *Workload) SetArguments(v string)`

SetArguments sets Arguments field to given value.

### HasArguments

`func (o *Workload) HasArguments() bool`

HasArguments returns a boolean if a field has been set.

### GetPhaseReason

`func (o *Workload) GetPhaseReason() PhaseReason`

GetPhaseReason returns the PhaseReason field if non-nil, zero value otherwise.

### GetPhaseReasonOk

`func (o *Workload) GetPhaseReasonOk() (*PhaseReason, bool)`

GetPhaseReasonOk returns a tuple with the PhaseReason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseReason

`func (o *Workload) SetPhaseReason(v PhaseReason)`

SetPhaseReason sets PhaseReason field to given value.

### HasPhaseReason

`func (o *Workload) HasPhaseReason() bool`

HasPhaseReason returns a boolean if a field has been set.

### SetPhaseReasonNil

`func (o *Workload) SetPhaseReasonNil(b bool)`

 SetPhaseReasonNil sets the value for PhaseReason to be an explicit nil

### UnsetPhaseReason
`func (o *Workload) UnsetPhaseReason()`

UnsetPhaseReason ensures that no value is present for PhaseReason, not even an explicit nil
### GetIdleAllocatedGpus

`func (o *Workload) GetIdleAllocatedGpus() float32`

GetIdleAllocatedGpus returns the IdleAllocatedGpus field if non-nil, zero value otherwise.

### GetIdleAllocatedGpusOk

`func (o *Workload) GetIdleAllocatedGpusOk() (*float32, bool)`

GetIdleAllocatedGpusOk returns a tuple with the IdleAllocatedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdleAllocatedGpus

`func (o *Workload) SetIdleAllocatedGpus(v float32)`

SetIdleAllocatedGpus sets IdleAllocatedGpus field to given value.

### HasIdleAllocatedGpus

`func (o *Workload) HasIdleAllocatedGpus() bool`

HasIdleAllocatedGpus returns a boolean if a field has been set.

### SetIdleAllocatedGpusNil

`func (o *Workload) SetIdleAllocatedGpusNil(b bool)`

 SetIdleAllocatedGpusNil sets the value for IdleAllocatedGpus to be an explicit nil

### UnsetIdleAllocatedGpus
`func (o *Workload) UnsetIdleAllocatedGpus()`

UnsetIdleAllocatedGpus ensures that no value is present for IdleAllocatedGpus, not even an explicit nil
### GetIdleSince

`func (o *Workload) GetIdleSince() time.Time`

GetIdleSince returns the IdleSince field if non-nil, zero value otherwise.

### GetIdleSinceOk

`func (o *Workload) GetIdleSinceOk() (*time.Time, bool)`

GetIdleSinceOk returns a tuple with the IdleSince field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdleSince

`func (o *Workload) SetIdleSince(v time.Time)`

SetIdleSince sets IdleSince field to given value.

### HasIdleSince

`func (o *Workload) HasIdleSince() bool`

HasIdleSince returns a boolean if a field has been set.

### SetIdleSinceNil

`func (o *Workload) SetIdleSinceNil(b bool)`

 SetIdleSinceNil sets the value for IdleSince to be an explicit nil

### UnsetIdleSince
`func (o *Workload) UnsetIdleSince()`

UnsetIdleSince ensures that no value is present for IdleSince, not even an explicit nil
### GetTotalPendingTimeSeconds

`func (o *Workload) GetTotalPendingTimeSeconds() int32`

GetTotalPendingTimeSeconds returns the TotalPendingTimeSeconds field if non-nil, zero value otherwise.

### GetTotalPendingTimeSecondsOk

`func (o *Workload) GetTotalPendingTimeSecondsOk() (*int32, bool)`

GetTotalPendingTimeSecondsOk returns a tuple with the TotalPendingTimeSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalPendingTimeSeconds

`func (o *Workload) SetTotalPendingTimeSeconds(v int32)`

SetTotalPendingTimeSeconds sets TotalPendingTimeSeconds field to given value.

### HasTotalPendingTimeSeconds

`func (o *Workload) HasTotalPendingTimeSeconds() bool`

HasTotalPendingTimeSeconds returns a boolean if a field has been set.

### SetTotalPendingTimeSecondsNil

`func (o *Workload) SetTotalPendingTimeSecondsNil(b bool)`

 SetTotalPendingTimeSecondsNil sets the value for TotalPendingTimeSeconds to be an explicit nil

### UnsetTotalPendingTimeSeconds
`func (o *Workload) UnsetTotalPendingTimeSeconds()`

UnsetTotalPendingTimeSeconds ensures that no value is present for TotalPendingTimeSeconds, not even an explicit nil
### GetTotalRunningTimeSeconds

`func (o *Workload) GetTotalRunningTimeSeconds() int32`

GetTotalRunningTimeSeconds returns the TotalRunningTimeSeconds field if non-nil, zero value otherwise.

### GetTotalRunningTimeSecondsOk

`func (o *Workload) GetTotalRunningTimeSecondsOk() (*int32, bool)`

GetTotalRunningTimeSecondsOk returns a tuple with the TotalRunningTimeSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalRunningTimeSeconds

`func (o *Workload) SetTotalRunningTimeSeconds(v int32)`

SetTotalRunningTimeSeconds sets TotalRunningTimeSeconds field to given value.

### HasTotalRunningTimeSeconds

`func (o *Workload) HasTotalRunningTimeSeconds() bool`

HasTotalRunningTimeSeconds returns a boolean if a field has been set.

### SetTotalRunningTimeSecondsNil

`func (o *Workload) SetTotalRunningTimeSecondsNil(b bool)`

 SetTotalRunningTimeSecondsNil sets the value for TotalRunningTimeSeconds to be an explicit nil

### UnsetTotalRunningTimeSeconds
`func (o *Workload) UnsetTotalRunningTimeSeconds()`

UnsetTotalRunningTimeSeconds ensures that no value is present for TotalRunningTimeSeconds, not even an explicit nil
### GetCategory

`func (o *Workload) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *Workload) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *Workload) SetCategory(v string)`

SetCategory sets Category field to given value.


### GetGuaranteedRuntimeEndsAt

`func (o *Workload) GetGuaranteedRuntimeEndsAt() time.Time`

GetGuaranteedRuntimeEndsAt returns the GuaranteedRuntimeEndsAt field if non-nil, zero value otherwise.

### GetGuaranteedRuntimeEndsAtOk

`func (o *Workload) GetGuaranteedRuntimeEndsAtOk() (*time.Time, bool)`

GetGuaranteedRuntimeEndsAtOk returns a tuple with the GuaranteedRuntimeEndsAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuaranteedRuntimeEndsAt

`func (o *Workload) SetGuaranteedRuntimeEndsAt(v time.Time)`

SetGuaranteedRuntimeEndsAt sets GuaranteedRuntimeEndsAt field to given value.

### HasGuaranteedRuntimeEndsAt

`func (o *Workload) HasGuaranteedRuntimeEndsAt() bool`

HasGuaranteedRuntimeEndsAt returns a boolean if a field has been set.

### SetGuaranteedRuntimeEndsAtNil

`func (o *Workload) SetGuaranteedRuntimeEndsAtNil(b bool)`

 SetGuaranteedRuntimeEndsAtNil sets the value for GuaranteedRuntimeEndsAt to be an explicit nil

### UnsetGuaranteedRuntimeEndsAt
`func (o *Workload) UnsetGuaranteedRuntimeEndsAt()`

UnsetGuaranteedRuntimeEndsAt ensures that no value is present for GuaranteedRuntimeEndsAt, not even an explicit nil
### GetAiApplicationId

`func (o *Workload) GetAiApplicationId() string`

GetAiApplicationId returns the AiApplicationId field if non-nil, zero value otherwise.

### GetAiApplicationIdOk

`func (o *Workload) GetAiApplicationIdOk() (*string, bool)`

GetAiApplicationIdOk returns a tuple with the AiApplicationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAiApplicationId

`func (o *Workload) SetAiApplicationId(v string)`

SetAiApplicationId sets AiApplicationId field to given value.

### HasAiApplicationId

`func (o *Workload) HasAiApplicationId() bool`

HasAiApplicationId returns a boolean if a field has been set.

### GetAiApplicationName

`func (o *Workload) GetAiApplicationName() string`

GetAiApplicationName returns the AiApplicationName field if non-nil, zero value otherwise.

### GetAiApplicationNameOk

`func (o *Workload) GetAiApplicationNameOk() (*string, bool)`

GetAiApplicationNameOk returns a tuple with the AiApplicationName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAiApplicationName

`func (o *Workload) SetAiApplicationName(v string)`

SetAiApplicationName sets AiApplicationName field to given value.

### HasAiApplicationName

`func (o *Workload) HasAiApplicationName() bool`

HasAiApplicationName returns a boolean if a field has been set.

### GetSourceApi

`func (o *Workload) GetSourceApi() SourceApi`

GetSourceApi returns the SourceApi field if non-nil, zero value otherwise.

### GetSourceApiOk

`func (o *Workload) GetSourceApiOk() (*SourceApi, bool)`

GetSourceApiOk returns a tuple with the SourceApi field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceApi

`func (o *Workload) SetSourceApi(v SourceApi)`

SetSourceApi sets SourceApi field to given value.

### HasSourceApi

`func (o *Workload) HasSourceApi() bool`

HasSourceApi returns a boolean if a field has been set.

### SetSourceApiNil

`func (o *Workload) SetSourceApiNil(b bool)`

 SetSourceApiNil sets the value for SourceApi to be an explicit nil

### UnsetSourceApi
`func (o *Workload) UnsetSourceApi()`

UnsetSourceApi ensures that no value is present for SourceApi, not even an explicit nil
### GetTopology

`func (o *Workload) GetTopology() WorkloadTopology`

GetTopology returns the Topology field if non-nil, zero value otherwise.

### GetTopologyOk

`func (o *Workload) GetTopologyOk() (*WorkloadTopology, bool)`

GetTopologyOk returns a tuple with the Topology field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopology

`func (o *Workload) SetTopology(v WorkloadTopology)`

SetTopology sets Topology field to given value.

### HasTopology

`func (o *Workload) HasTopology() bool`

HasTopology returns a boolean if a field has been set.

### SetTopologyNil

`func (o *Workload) SetTopologyNil(b bool)`

 SetTopologyNil sets the value for Topology to be an explicit nil

### UnsetTopology
`func (o *Workload) UnsetTopology()`

UnsetTopology ensures that no value is present for Topology, not even an explicit nil
### GetEndpoints

`func (o *Workload) GetEndpoints() []Endpoint`

GetEndpoints returns the Endpoints field if non-nil, zero value otherwise.

### GetEndpointsOk

`func (o *Workload) GetEndpointsOk() (*[]Endpoint, bool)`

GetEndpointsOk returns a tuple with the Endpoints field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEndpoints

`func (o *Workload) SetEndpoints(v []Endpoint)`

SetEndpoints sets Endpoints field to given value.

### HasEndpoints

`func (o *Workload) HasEndpoints() bool`

HasEndpoints returns a boolean if a field has been set.

### GetNodeAffinities

`func (o *Workload) GetNodeAffinities() []WorkloadNodeAffinity`

GetNodeAffinities returns the NodeAffinities field if non-nil, zero value otherwise.

### GetNodeAffinitiesOk

`func (o *Workload) GetNodeAffinitiesOk() (*[]WorkloadNodeAffinity, bool)`

GetNodeAffinitiesOk returns a tuple with the NodeAffinities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinities

`func (o *Workload) SetNodeAffinities(v []WorkloadNodeAffinity)`

SetNodeAffinities sets NodeAffinities field to given value.

### HasNodeAffinities

`func (o *Workload) HasNodeAffinities() bool`

HasNodeAffinities returns a boolean if a field has been set.

### GetKubernetesRef

`func (o *Workload) GetKubernetesRef() WorkloadKubernetesRef`

GetKubernetesRef returns the KubernetesRef field if non-nil, zero value otherwise.

### GetKubernetesRefOk

`func (o *Workload) GetKubernetesRefOk() (*WorkloadKubernetesRef, bool)`

GetKubernetesRefOk returns a tuple with the KubernetesRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKubernetesRef

`func (o *Workload) SetKubernetesRef(v WorkloadKubernetesRef)`

SetKubernetesRef sets KubernetesRef field to given value.

### HasKubernetesRef

`func (o *Workload) HasKubernetesRef() bool`

HasKubernetesRef returns a boolean if a field has been set.

### SetKubernetesRefNil

`func (o *Workload) SetKubernetesRefNil(b bool)`

 SetKubernetesRefNil sets the value for KubernetesRef to be an explicit nil

### UnsetKubernetesRef
`func (o *Workload) UnsetKubernetesRef()`

UnsetKubernetesRef ensures that no value is present for KubernetesRef, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


