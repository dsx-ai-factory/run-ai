# Pod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**PriorityClassName** | **string** |  | 
**Id** | **string** |  | 
**WorkloadId** | **string** |  | 
**RevisionId** | Pointer to **NullableString** |  | [optional] 
**ClusterId** | **string** | The id of the cluster. | 
**ProjectId** | Pointer to **string** |  | [optional] 
**NodeName** | Pointer to **NullableString** |  | [optional] 
**CreatedAt** | **time.Time** |  | 
**CompletedAt** | Pointer to **NullableTime** |  | [optional] 
**Containers** | [**[]Container**](Container.md) |  | 
**CurrentNodePool** | Pointer to **NullableString** |  | [optional] 
**RequestedNodePools** | Pointer to **[]string** |  | [optional] 
**RequestedResources** | Pointer to [**NullablePodRequestResources**](PodRequestResources.md) |  | [optional] 
**AllocatedResources** | Pointer to [**NullableAllocatedResources**](AllocatedResources.md) |  | [optional] 
**Tolerations** | Pointer to [**[]PodToleration**](PodToleration.md) |  | [optional] 
**K8sPhase** | **string** |  | 
**Ip** | Pointer to **NullableString** |  | [optional] 
**Task** | Pointer to **NullableString** | Identifies the function of the pod within a distributed workload (for example, leader, worker, or master). | [optional] 
**NodeAffinity** | Pointer to [**PodNodeAffinity**](PodNodeAffinity.md) |  | [optional] 
**Elements** | Pointer to [**[]PodElement**](PodElement.md) | The workload structure elements this pod belongs to, ordered by hierarchy level. | [optional] 
**ResourceClaims** | Pointer to [**[]PodResourceClaim**](PodResourceClaim.md) | Per-pod DRA usage reported as-is from the cluster. One entry per &#x60;pod.spec.resourceClaims[]&#x60; member — either a reference to a &#x60;ResourceClaimTemplate&#x60; or a standalone &#x60;ResourceClaim&#x60; — enriched with the actual bound &#x60;ResourceClaim&#x60; name once published in &#x60;pod.status.resourceClaimStatuses[]&#x60;. Managed/ownership resolution against the k8s-object store happens at API read time, not here. | [optional] 
**TenantId** | **int32** | The id of the tenant. | 
**K8sPhaseUpdatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**ProjectName** | Pointer to **string** |  | [optional] 
**WorkloadName** | Pointer to **string** |  | [optional] 

## Methods

### NewPod

`func NewPod(name string, priorityClassName string, id string, workloadId string, clusterId string, createdAt time.Time, containers []Container, k8sPhase string, tenantId int32, k8sPhaseUpdatedAt time.Time, updatedAt time.Time, ) *Pod`

NewPod instantiates a new Pod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodWithDefaults

`func NewPodWithDefaults() *Pod`

NewPodWithDefaults instantiates a new Pod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Pod) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Pod) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Pod) SetName(v string)`

SetName sets Name field to given value.


### GetPriorityClassName

`func (o *Pod) GetPriorityClassName() string`

GetPriorityClassName returns the PriorityClassName field if non-nil, zero value otherwise.

### GetPriorityClassNameOk

`func (o *Pod) GetPriorityClassNameOk() (*string, bool)`

GetPriorityClassNameOk returns a tuple with the PriorityClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClassName

`func (o *Pod) SetPriorityClassName(v string)`

SetPriorityClassName sets PriorityClassName field to given value.


### GetId

`func (o *Pod) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Pod) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Pod) SetId(v string)`

SetId sets Id field to given value.


### GetWorkloadId

`func (o *Pod) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *Pod) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *Pod) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetRevisionId

`func (o *Pod) GetRevisionId() string`

GetRevisionId returns the RevisionId field if non-nil, zero value otherwise.

### GetRevisionIdOk

`func (o *Pod) GetRevisionIdOk() (*string, bool)`

GetRevisionIdOk returns a tuple with the RevisionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevisionId

`func (o *Pod) SetRevisionId(v string)`

SetRevisionId sets RevisionId field to given value.

### HasRevisionId

`func (o *Pod) HasRevisionId() bool`

HasRevisionId returns a boolean if a field has been set.

### SetRevisionIdNil

`func (o *Pod) SetRevisionIdNil(b bool)`

 SetRevisionIdNil sets the value for RevisionId to be an explicit nil

### UnsetRevisionId
`func (o *Pod) UnsetRevisionId()`

UnsetRevisionId ensures that no value is present for RevisionId, not even an explicit nil
### GetClusterId

`func (o *Pod) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Pod) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Pod) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *Pod) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Pod) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Pod) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *Pod) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetNodeName

`func (o *Pod) GetNodeName() string`

GetNodeName returns the NodeName field if non-nil, zero value otherwise.

### GetNodeNameOk

`func (o *Pod) GetNodeNameOk() (*string, bool)`

GetNodeNameOk returns a tuple with the NodeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeName

`func (o *Pod) SetNodeName(v string)`

SetNodeName sets NodeName field to given value.

### HasNodeName

`func (o *Pod) HasNodeName() bool`

HasNodeName returns a boolean if a field has been set.

### SetNodeNameNil

`func (o *Pod) SetNodeNameNil(b bool)`

 SetNodeNameNil sets the value for NodeName to be an explicit nil

### UnsetNodeName
`func (o *Pod) UnsetNodeName()`

UnsetNodeName ensures that no value is present for NodeName, not even an explicit nil
### GetCreatedAt

`func (o *Pod) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Pod) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Pod) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetCompletedAt

`func (o *Pod) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *Pod) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *Pod) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *Pod) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.

### SetCompletedAtNil

`func (o *Pod) SetCompletedAtNil(b bool)`

 SetCompletedAtNil sets the value for CompletedAt to be an explicit nil

### UnsetCompletedAt
`func (o *Pod) UnsetCompletedAt()`

UnsetCompletedAt ensures that no value is present for CompletedAt, not even an explicit nil
### GetContainers

`func (o *Pod) GetContainers() []Container`

GetContainers returns the Containers field if non-nil, zero value otherwise.

### GetContainersOk

`func (o *Pod) GetContainersOk() (*[]Container, bool)`

GetContainersOk returns a tuple with the Containers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainers

`func (o *Pod) SetContainers(v []Container)`

SetContainers sets Containers field to given value.


### GetCurrentNodePool

`func (o *Pod) GetCurrentNodePool() string`

GetCurrentNodePool returns the CurrentNodePool field if non-nil, zero value otherwise.

### GetCurrentNodePoolOk

`func (o *Pod) GetCurrentNodePoolOk() (*string, bool)`

GetCurrentNodePoolOk returns a tuple with the CurrentNodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentNodePool

`func (o *Pod) SetCurrentNodePool(v string)`

SetCurrentNodePool sets CurrentNodePool field to given value.

### HasCurrentNodePool

`func (o *Pod) HasCurrentNodePool() bool`

HasCurrentNodePool returns a boolean if a field has been set.

### SetCurrentNodePoolNil

`func (o *Pod) SetCurrentNodePoolNil(b bool)`

 SetCurrentNodePoolNil sets the value for CurrentNodePool to be an explicit nil

### UnsetCurrentNodePool
`func (o *Pod) UnsetCurrentNodePool()`

UnsetCurrentNodePool ensures that no value is present for CurrentNodePool, not even an explicit nil
### GetRequestedNodePools

`func (o *Pod) GetRequestedNodePools() []string`

GetRequestedNodePools returns the RequestedNodePools field if non-nil, zero value otherwise.

### GetRequestedNodePoolsOk

`func (o *Pod) GetRequestedNodePoolsOk() (*[]string, bool)`

GetRequestedNodePoolsOk returns a tuple with the RequestedNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNodePools

`func (o *Pod) SetRequestedNodePools(v []string)`

SetRequestedNodePools sets RequestedNodePools field to given value.

### HasRequestedNodePools

`func (o *Pod) HasRequestedNodePools() bool`

HasRequestedNodePools returns a boolean if a field has been set.

### GetRequestedResources

`func (o *Pod) GetRequestedResources() PodRequestResources`

GetRequestedResources returns the RequestedResources field if non-nil, zero value otherwise.

### GetRequestedResourcesOk

`func (o *Pod) GetRequestedResourcesOk() (*PodRequestResources, bool)`

GetRequestedResourcesOk returns a tuple with the RequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedResources

`func (o *Pod) SetRequestedResources(v PodRequestResources)`

SetRequestedResources sets RequestedResources field to given value.

### HasRequestedResources

`func (o *Pod) HasRequestedResources() bool`

HasRequestedResources returns a boolean if a field has been set.

### SetRequestedResourcesNil

`func (o *Pod) SetRequestedResourcesNil(b bool)`

 SetRequestedResourcesNil sets the value for RequestedResources to be an explicit nil

### UnsetRequestedResources
`func (o *Pod) UnsetRequestedResources()`

UnsetRequestedResources ensures that no value is present for RequestedResources, not even an explicit nil
### GetAllocatedResources

`func (o *Pod) GetAllocatedResources() AllocatedResources`

GetAllocatedResources returns the AllocatedResources field if non-nil, zero value otherwise.

### GetAllocatedResourcesOk

`func (o *Pod) GetAllocatedResourcesOk() (*AllocatedResources, bool)`

GetAllocatedResourcesOk returns a tuple with the AllocatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedResources

`func (o *Pod) SetAllocatedResources(v AllocatedResources)`

SetAllocatedResources sets AllocatedResources field to given value.

### HasAllocatedResources

`func (o *Pod) HasAllocatedResources() bool`

HasAllocatedResources returns a boolean if a field has been set.

### SetAllocatedResourcesNil

`func (o *Pod) SetAllocatedResourcesNil(b bool)`

 SetAllocatedResourcesNil sets the value for AllocatedResources to be an explicit nil

### UnsetAllocatedResources
`func (o *Pod) UnsetAllocatedResources()`

UnsetAllocatedResources ensures that no value is present for AllocatedResources, not even an explicit nil
### GetTolerations

`func (o *Pod) GetTolerations() []PodToleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *Pod) GetTolerationsOk() (*[]PodToleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *Pod) SetTolerations(v []PodToleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *Pod) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### GetK8sPhase

`func (o *Pod) GetK8sPhase() string`

GetK8sPhase returns the K8sPhase field if non-nil, zero value otherwise.

### GetK8sPhaseOk

`func (o *Pod) GetK8sPhaseOk() (*string, bool)`

GetK8sPhaseOk returns a tuple with the K8sPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhase

`func (o *Pod) SetK8sPhase(v string)`

SetK8sPhase sets K8sPhase field to given value.


### GetIp

`func (o *Pod) GetIp() string`

GetIp returns the Ip field if non-nil, zero value otherwise.

### GetIpOk

`func (o *Pod) GetIpOk() (*string, bool)`

GetIpOk returns a tuple with the Ip field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIp

`func (o *Pod) SetIp(v string)`

SetIp sets Ip field to given value.

### HasIp

`func (o *Pod) HasIp() bool`

HasIp returns a boolean if a field has been set.

### SetIpNil

`func (o *Pod) SetIpNil(b bool)`

 SetIpNil sets the value for Ip to be an explicit nil

### UnsetIp
`func (o *Pod) UnsetIp()`

UnsetIp ensures that no value is present for Ip, not even an explicit nil
### GetTask

`func (o *Pod) GetTask() string`

GetTask returns the Task field if non-nil, zero value otherwise.

### GetTaskOk

`func (o *Pod) GetTaskOk() (*string, bool)`

GetTaskOk returns a tuple with the Task field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTask

`func (o *Pod) SetTask(v string)`

SetTask sets Task field to given value.

### HasTask

`func (o *Pod) HasTask() bool`

HasTask returns a boolean if a field has been set.

### SetTaskNil

`func (o *Pod) SetTaskNil(b bool)`

 SetTaskNil sets the value for Task to be an explicit nil

### UnsetTask
`func (o *Pod) UnsetTask()`

UnsetTask ensures that no value is present for Task, not even an explicit nil
### GetNodeAffinity

`func (o *Pod) GetNodeAffinity() PodNodeAffinity`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *Pod) GetNodeAffinityOk() (*PodNodeAffinity, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *Pod) SetNodeAffinity(v PodNodeAffinity)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *Pod) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.

### GetElements

`func (o *Pod) GetElements() []PodElement`

GetElements returns the Elements field if non-nil, zero value otherwise.

### GetElementsOk

`func (o *Pod) GetElementsOk() (*[]PodElement, bool)`

GetElementsOk returns a tuple with the Elements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetElements

`func (o *Pod) SetElements(v []PodElement)`

SetElements sets Elements field to given value.

### HasElements

`func (o *Pod) HasElements() bool`

HasElements returns a boolean if a field has been set.

### SetElementsNil

`func (o *Pod) SetElementsNil(b bool)`

 SetElementsNil sets the value for Elements to be an explicit nil

### UnsetElements
`func (o *Pod) UnsetElements()`

UnsetElements ensures that no value is present for Elements, not even an explicit nil
### GetResourceClaims

`func (o *Pod) GetResourceClaims() []PodResourceClaim`

GetResourceClaims returns the ResourceClaims field if non-nil, zero value otherwise.

### GetResourceClaimsOk

`func (o *Pod) GetResourceClaimsOk() (*[]PodResourceClaim, bool)`

GetResourceClaimsOk returns a tuple with the ResourceClaims field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceClaims

`func (o *Pod) SetResourceClaims(v []PodResourceClaim)`

SetResourceClaims sets ResourceClaims field to given value.

### HasResourceClaims

`func (o *Pod) HasResourceClaims() bool`

HasResourceClaims returns a boolean if a field has been set.

### SetResourceClaimsNil

`func (o *Pod) SetResourceClaimsNil(b bool)`

 SetResourceClaimsNil sets the value for ResourceClaims to be an explicit nil

### UnsetResourceClaims
`func (o *Pod) UnsetResourceClaims()`

UnsetResourceClaims ensures that no value is present for ResourceClaims, not even an explicit nil
### GetTenantId

`func (o *Pod) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Pod) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Pod) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetK8sPhaseUpdatedAt

`func (o *Pod) GetK8sPhaseUpdatedAt() time.Time`

GetK8sPhaseUpdatedAt returns the K8sPhaseUpdatedAt field if non-nil, zero value otherwise.

### GetK8sPhaseUpdatedAtOk

`func (o *Pod) GetK8sPhaseUpdatedAtOk() (*time.Time, bool)`

GetK8sPhaseUpdatedAtOk returns a tuple with the K8sPhaseUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhaseUpdatedAt

`func (o *Pod) SetK8sPhaseUpdatedAt(v time.Time)`

SetK8sPhaseUpdatedAt sets K8sPhaseUpdatedAt field to given value.


### GetUpdatedAt

`func (o *Pod) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Pod) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Pod) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *Pod) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *Pod) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *Pod) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *Pod) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *Pod) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *Pod) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetProjectName

`func (o *Pod) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *Pod) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *Pod) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *Pod) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetWorkloadName

`func (o *Pod) GetWorkloadName() string`

GetWorkloadName returns the WorkloadName field if non-nil, zero value otherwise.

### GetWorkloadNameOk

`func (o *Pod) GetWorkloadNameOk() (*string, bool)`

GetWorkloadNameOk returns a tuple with the WorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadName

`func (o *Pod) SetWorkloadName(v string)`

SetWorkloadName sets WorkloadName field to given value.

### HasWorkloadName

`func (o *Pod) HasWorkloadName() bool`

HasWorkloadName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


