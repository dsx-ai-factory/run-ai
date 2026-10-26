# PodSyncFields

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

## Methods

### NewPodSyncFields

`func NewPodSyncFields(name string, priorityClassName string, id string, workloadId string, clusterId string, createdAt time.Time, containers []Container, k8sPhase string, ) *PodSyncFields`

NewPodSyncFields instantiates a new PodSyncFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodSyncFieldsWithDefaults

`func NewPodSyncFieldsWithDefaults() *PodSyncFields`

NewPodSyncFieldsWithDefaults instantiates a new PodSyncFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PodSyncFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PodSyncFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PodSyncFields) SetName(v string)`

SetName sets Name field to given value.


### GetPriorityClassName

`func (o *PodSyncFields) GetPriorityClassName() string`

GetPriorityClassName returns the PriorityClassName field if non-nil, zero value otherwise.

### GetPriorityClassNameOk

`func (o *PodSyncFields) GetPriorityClassNameOk() (*string, bool)`

GetPriorityClassNameOk returns a tuple with the PriorityClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClassName

`func (o *PodSyncFields) SetPriorityClassName(v string)`

SetPriorityClassName sets PriorityClassName field to given value.


### GetId

`func (o *PodSyncFields) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PodSyncFields) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PodSyncFields) SetId(v string)`

SetId sets Id field to given value.


### GetWorkloadId

`func (o *PodSyncFields) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *PodSyncFields) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *PodSyncFields) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetRevisionId

`func (o *PodSyncFields) GetRevisionId() string`

GetRevisionId returns the RevisionId field if non-nil, zero value otherwise.

### GetRevisionIdOk

`func (o *PodSyncFields) GetRevisionIdOk() (*string, bool)`

GetRevisionIdOk returns a tuple with the RevisionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevisionId

`func (o *PodSyncFields) SetRevisionId(v string)`

SetRevisionId sets RevisionId field to given value.

### HasRevisionId

`func (o *PodSyncFields) HasRevisionId() bool`

HasRevisionId returns a boolean if a field has been set.

### SetRevisionIdNil

`func (o *PodSyncFields) SetRevisionIdNil(b bool)`

 SetRevisionIdNil sets the value for RevisionId to be an explicit nil

### UnsetRevisionId
`func (o *PodSyncFields) UnsetRevisionId()`

UnsetRevisionId ensures that no value is present for RevisionId, not even an explicit nil
### GetClusterId

`func (o *PodSyncFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PodSyncFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PodSyncFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *PodSyncFields) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *PodSyncFields) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *PodSyncFields) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *PodSyncFields) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetNodeName

`func (o *PodSyncFields) GetNodeName() string`

GetNodeName returns the NodeName field if non-nil, zero value otherwise.

### GetNodeNameOk

`func (o *PodSyncFields) GetNodeNameOk() (*string, bool)`

GetNodeNameOk returns a tuple with the NodeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeName

`func (o *PodSyncFields) SetNodeName(v string)`

SetNodeName sets NodeName field to given value.

### HasNodeName

`func (o *PodSyncFields) HasNodeName() bool`

HasNodeName returns a boolean if a field has been set.

### SetNodeNameNil

`func (o *PodSyncFields) SetNodeNameNil(b bool)`

 SetNodeNameNil sets the value for NodeName to be an explicit nil

### UnsetNodeName
`func (o *PodSyncFields) UnsetNodeName()`

UnsetNodeName ensures that no value is present for NodeName, not even an explicit nil
### GetCreatedAt

`func (o *PodSyncFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *PodSyncFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *PodSyncFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetCompletedAt

`func (o *PodSyncFields) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *PodSyncFields) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *PodSyncFields) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *PodSyncFields) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.

### SetCompletedAtNil

`func (o *PodSyncFields) SetCompletedAtNil(b bool)`

 SetCompletedAtNil sets the value for CompletedAt to be an explicit nil

### UnsetCompletedAt
`func (o *PodSyncFields) UnsetCompletedAt()`

UnsetCompletedAt ensures that no value is present for CompletedAt, not even an explicit nil
### GetContainers

`func (o *PodSyncFields) GetContainers() []Container`

GetContainers returns the Containers field if non-nil, zero value otherwise.

### GetContainersOk

`func (o *PodSyncFields) GetContainersOk() (*[]Container, bool)`

GetContainersOk returns a tuple with the Containers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainers

`func (o *PodSyncFields) SetContainers(v []Container)`

SetContainers sets Containers field to given value.


### GetCurrentNodePool

`func (o *PodSyncFields) GetCurrentNodePool() string`

GetCurrentNodePool returns the CurrentNodePool field if non-nil, zero value otherwise.

### GetCurrentNodePoolOk

`func (o *PodSyncFields) GetCurrentNodePoolOk() (*string, bool)`

GetCurrentNodePoolOk returns a tuple with the CurrentNodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentNodePool

`func (o *PodSyncFields) SetCurrentNodePool(v string)`

SetCurrentNodePool sets CurrentNodePool field to given value.

### HasCurrentNodePool

`func (o *PodSyncFields) HasCurrentNodePool() bool`

HasCurrentNodePool returns a boolean if a field has been set.

### SetCurrentNodePoolNil

`func (o *PodSyncFields) SetCurrentNodePoolNil(b bool)`

 SetCurrentNodePoolNil sets the value for CurrentNodePool to be an explicit nil

### UnsetCurrentNodePool
`func (o *PodSyncFields) UnsetCurrentNodePool()`

UnsetCurrentNodePool ensures that no value is present for CurrentNodePool, not even an explicit nil
### GetRequestedNodePools

`func (o *PodSyncFields) GetRequestedNodePools() []string`

GetRequestedNodePools returns the RequestedNodePools field if non-nil, zero value otherwise.

### GetRequestedNodePoolsOk

`func (o *PodSyncFields) GetRequestedNodePoolsOk() (*[]string, bool)`

GetRequestedNodePoolsOk returns a tuple with the RequestedNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNodePools

`func (o *PodSyncFields) SetRequestedNodePools(v []string)`

SetRequestedNodePools sets RequestedNodePools field to given value.

### HasRequestedNodePools

`func (o *PodSyncFields) HasRequestedNodePools() bool`

HasRequestedNodePools returns a boolean if a field has been set.

### GetRequestedResources

`func (o *PodSyncFields) GetRequestedResources() PodRequestResources`

GetRequestedResources returns the RequestedResources field if non-nil, zero value otherwise.

### GetRequestedResourcesOk

`func (o *PodSyncFields) GetRequestedResourcesOk() (*PodRequestResources, bool)`

GetRequestedResourcesOk returns a tuple with the RequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedResources

`func (o *PodSyncFields) SetRequestedResources(v PodRequestResources)`

SetRequestedResources sets RequestedResources field to given value.

### HasRequestedResources

`func (o *PodSyncFields) HasRequestedResources() bool`

HasRequestedResources returns a boolean if a field has been set.

### SetRequestedResourcesNil

`func (o *PodSyncFields) SetRequestedResourcesNil(b bool)`

 SetRequestedResourcesNil sets the value for RequestedResources to be an explicit nil

### UnsetRequestedResources
`func (o *PodSyncFields) UnsetRequestedResources()`

UnsetRequestedResources ensures that no value is present for RequestedResources, not even an explicit nil
### GetAllocatedResources

`func (o *PodSyncFields) GetAllocatedResources() AllocatedResources`

GetAllocatedResources returns the AllocatedResources field if non-nil, zero value otherwise.

### GetAllocatedResourcesOk

`func (o *PodSyncFields) GetAllocatedResourcesOk() (*AllocatedResources, bool)`

GetAllocatedResourcesOk returns a tuple with the AllocatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedResources

`func (o *PodSyncFields) SetAllocatedResources(v AllocatedResources)`

SetAllocatedResources sets AllocatedResources field to given value.

### HasAllocatedResources

`func (o *PodSyncFields) HasAllocatedResources() bool`

HasAllocatedResources returns a boolean if a field has been set.

### SetAllocatedResourcesNil

`func (o *PodSyncFields) SetAllocatedResourcesNil(b bool)`

 SetAllocatedResourcesNil sets the value for AllocatedResources to be an explicit nil

### UnsetAllocatedResources
`func (o *PodSyncFields) UnsetAllocatedResources()`

UnsetAllocatedResources ensures that no value is present for AllocatedResources, not even an explicit nil
### GetTolerations

`func (o *PodSyncFields) GetTolerations() []PodToleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *PodSyncFields) GetTolerationsOk() (*[]PodToleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *PodSyncFields) SetTolerations(v []PodToleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *PodSyncFields) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### GetK8sPhase

`func (o *PodSyncFields) GetK8sPhase() string`

GetK8sPhase returns the K8sPhase field if non-nil, zero value otherwise.

### GetK8sPhaseOk

`func (o *PodSyncFields) GetK8sPhaseOk() (*string, bool)`

GetK8sPhaseOk returns a tuple with the K8sPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhase

`func (o *PodSyncFields) SetK8sPhase(v string)`

SetK8sPhase sets K8sPhase field to given value.


### GetIp

`func (o *PodSyncFields) GetIp() string`

GetIp returns the Ip field if non-nil, zero value otherwise.

### GetIpOk

`func (o *PodSyncFields) GetIpOk() (*string, bool)`

GetIpOk returns a tuple with the Ip field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIp

`func (o *PodSyncFields) SetIp(v string)`

SetIp sets Ip field to given value.

### HasIp

`func (o *PodSyncFields) HasIp() bool`

HasIp returns a boolean if a field has been set.

### SetIpNil

`func (o *PodSyncFields) SetIpNil(b bool)`

 SetIpNil sets the value for Ip to be an explicit nil

### UnsetIp
`func (o *PodSyncFields) UnsetIp()`

UnsetIp ensures that no value is present for Ip, not even an explicit nil
### GetTask

`func (o *PodSyncFields) GetTask() string`

GetTask returns the Task field if non-nil, zero value otherwise.

### GetTaskOk

`func (o *PodSyncFields) GetTaskOk() (*string, bool)`

GetTaskOk returns a tuple with the Task field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTask

`func (o *PodSyncFields) SetTask(v string)`

SetTask sets Task field to given value.

### HasTask

`func (o *PodSyncFields) HasTask() bool`

HasTask returns a boolean if a field has been set.

### SetTaskNil

`func (o *PodSyncFields) SetTaskNil(b bool)`

 SetTaskNil sets the value for Task to be an explicit nil

### UnsetTask
`func (o *PodSyncFields) UnsetTask()`

UnsetTask ensures that no value is present for Task, not even an explicit nil
### GetNodeAffinity

`func (o *PodSyncFields) GetNodeAffinity() PodNodeAffinity`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *PodSyncFields) GetNodeAffinityOk() (*PodNodeAffinity, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *PodSyncFields) SetNodeAffinity(v PodNodeAffinity)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *PodSyncFields) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.

### GetElements

`func (o *PodSyncFields) GetElements() []PodElement`

GetElements returns the Elements field if non-nil, zero value otherwise.

### GetElementsOk

`func (o *PodSyncFields) GetElementsOk() (*[]PodElement, bool)`

GetElementsOk returns a tuple with the Elements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetElements

`func (o *PodSyncFields) SetElements(v []PodElement)`

SetElements sets Elements field to given value.

### HasElements

`func (o *PodSyncFields) HasElements() bool`

HasElements returns a boolean if a field has been set.

### SetElementsNil

`func (o *PodSyncFields) SetElementsNil(b bool)`

 SetElementsNil sets the value for Elements to be an explicit nil

### UnsetElements
`func (o *PodSyncFields) UnsetElements()`

UnsetElements ensures that no value is present for Elements, not even an explicit nil
### GetResourceClaims

`func (o *PodSyncFields) GetResourceClaims() []PodResourceClaim`

GetResourceClaims returns the ResourceClaims field if non-nil, zero value otherwise.

### GetResourceClaimsOk

`func (o *PodSyncFields) GetResourceClaimsOk() (*[]PodResourceClaim, bool)`

GetResourceClaimsOk returns a tuple with the ResourceClaims field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceClaims

`func (o *PodSyncFields) SetResourceClaims(v []PodResourceClaim)`

SetResourceClaims sets ResourceClaims field to given value.

### HasResourceClaims

`func (o *PodSyncFields) HasResourceClaims() bool`

HasResourceClaims returns a boolean if a field has been set.

### SetResourceClaimsNil

`func (o *PodSyncFields) SetResourceClaimsNil(b bool)`

 SetResourceClaimsNil sets the value for ResourceClaims to be an explicit nil

### UnsetResourceClaims
`func (o *PodSyncFields) UnsetResourceClaims()`

UnsetResourceClaims ensures that no value is present for ResourceClaims, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


