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
**Id** | **string** |  | 
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
**Urls** | Pointer to **[]string** |  | [optional] 
**Datasources** | Pointer to [**[]Datasource**](Datasource.md) |  | [optional] 
**Environments** | Pointer to [**[]Environment**](Environment.md) |  | [optional] 
**ExternalConnections** | Pointer to [**[]Connection1**](Connection1.md) |  | [optional] 
**DistributedFramework** | Pointer to **string** |  | [optional] 
**AdditionalFields** | Pointer to **map[string]interface{}** |  | [optional] 
**Preemptible** | Pointer to **NullableBool** |  | [optional] 
**FullySaasWorkload** | Pointer to **bool** |  | [optional] 

## Methods

### NewWorkload

`func NewWorkload(tenantId int32, runningPods int32, phaseUpdatedAt time.Time, k8sPhaseUpdatedAt time.Time, updatedAt time.Time, source Source, deletedAt NullableTime, type_ string, name string, id string, priorityClassName string, clusterId string, projectName string, projectId string, departmentName string, departmentId string, namespace string, createdAt time.Time, phase Phase, conditions []Condition, k8sPhase string, ) *Workload`

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

`func (o *Workload) GetExternalConnections() []Connection1`

GetExternalConnections returns the ExternalConnections field if non-nil, zero value otherwise.

### GetExternalConnectionsOk

`func (o *Workload) GetExternalConnectionsOk() (*[]Connection1, bool)`

GetExternalConnectionsOk returns a tuple with the ExternalConnections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalConnections

`func (o *Workload) SetExternalConnections(v []Connection1)`

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


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


