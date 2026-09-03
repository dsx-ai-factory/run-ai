# Pod1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**PriorityClassName** | **string** |  | 
**Id** | **string** |  | 
**WorkloadId** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**ProjectId** | Pointer to **string** |  | [optional] 
**NodeName** | Pointer to **NullableString** |  | [optional] 
**CreatedAt** | **time.Time** |  | 
**CompletedAt** | Pointer to **NullableTime** |  | [optional] 
**Containers** | [**[]Container1**](Container1.md) |  | 
**CurrentNodePool** | Pointer to **NullableString** |  | [optional] 
**RequestedNodePools** | Pointer to **[]string** |  | [optional] 
**RequestedResources** | Pointer to [**NullablePodRequestResources**](PodRequestResources.md) |  | [optional] 
**AllocatedResources** | Pointer to [**NullableAllocatedResources**](AllocatedResources.md) |  | [optional] 
**Tolerations** | Pointer to [**[]PodToleration**](PodToleration.md) |  | [optional] 
**K8sPhase** | **string** |  | 
**TenantId** | **int32** | The id of the tenant. | 
**K8sPhaseUpdatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**ProjectName** | Pointer to **string** |  | [optional] 
**WorkloadName** | Pointer to **string** |  | [optional] 

## Methods

### NewPod1

`func NewPod1(name string, priorityClassName string, id string, workloadId string, clusterId string, createdAt time.Time, containers []Container1, k8sPhase string, tenantId int32, k8sPhaseUpdatedAt time.Time, updatedAt time.Time, ) *Pod1`

NewPod1 instantiates a new Pod1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPod1WithDefaults

`func NewPod1WithDefaults() *Pod1`

NewPod1WithDefaults instantiates a new Pod1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Pod1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Pod1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Pod1) SetName(v string)`

SetName sets Name field to given value.


### GetPriorityClassName

`func (o *Pod1) GetPriorityClassName() string`

GetPriorityClassName returns the PriorityClassName field if non-nil, zero value otherwise.

### GetPriorityClassNameOk

`func (o *Pod1) GetPriorityClassNameOk() (*string, bool)`

GetPriorityClassNameOk returns a tuple with the PriorityClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClassName

`func (o *Pod1) SetPriorityClassName(v string)`

SetPriorityClassName sets PriorityClassName field to given value.


### GetId

`func (o *Pod1) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Pod1) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Pod1) SetId(v string)`

SetId sets Id field to given value.


### GetWorkloadId

`func (o *Pod1) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *Pod1) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *Pod1) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetClusterId

`func (o *Pod1) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Pod1) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Pod1) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *Pod1) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Pod1) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Pod1) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *Pod1) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetNodeName

`func (o *Pod1) GetNodeName() string`

GetNodeName returns the NodeName field if non-nil, zero value otherwise.

### GetNodeNameOk

`func (o *Pod1) GetNodeNameOk() (*string, bool)`

GetNodeNameOk returns a tuple with the NodeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeName

`func (o *Pod1) SetNodeName(v string)`

SetNodeName sets NodeName field to given value.

### HasNodeName

`func (o *Pod1) HasNodeName() bool`

HasNodeName returns a boolean if a field has been set.

### SetNodeNameNil

`func (o *Pod1) SetNodeNameNil(b bool)`

 SetNodeNameNil sets the value for NodeName to be an explicit nil

### UnsetNodeName
`func (o *Pod1) UnsetNodeName()`

UnsetNodeName ensures that no value is present for NodeName, not even an explicit nil
### GetCreatedAt

`func (o *Pod1) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Pod1) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Pod1) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetCompletedAt

`func (o *Pod1) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *Pod1) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *Pod1) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *Pod1) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.

### SetCompletedAtNil

`func (o *Pod1) SetCompletedAtNil(b bool)`

 SetCompletedAtNil sets the value for CompletedAt to be an explicit nil

### UnsetCompletedAt
`func (o *Pod1) UnsetCompletedAt()`

UnsetCompletedAt ensures that no value is present for CompletedAt, not even an explicit nil
### GetContainers

`func (o *Pod1) GetContainers() []Container1`

GetContainers returns the Containers field if non-nil, zero value otherwise.

### GetContainersOk

`func (o *Pod1) GetContainersOk() (*[]Container1, bool)`

GetContainersOk returns a tuple with the Containers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainers

`func (o *Pod1) SetContainers(v []Container1)`

SetContainers sets Containers field to given value.


### GetCurrentNodePool

`func (o *Pod1) GetCurrentNodePool() string`

GetCurrentNodePool returns the CurrentNodePool field if non-nil, zero value otherwise.

### GetCurrentNodePoolOk

`func (o *Pod1) GetCurrentNodePoolOk() (*string, bool)`

GetCurrentNodePoolOk returns a tuple with the CurrentNodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentNodePool

`func (o *Pod1) SetCurrentNodePool(v string)`

SetCurrentNodePool sets CurrentNodePool field to given value.

### HasCurrentNodePool

`func (o *Pod1) HasCurrentNodePool() bool`

HasCurrentNodePool returns a boolean if a field has been set.

### SetCurrentNodePoolNil

`func (o *Pod1) SetCurrentNodePoolNil(b bool)`

 SetCurrentNodePoolNil sets the value for CurrentNodePool to be an explicit nil

### UnsetCurrentNodePool
`func (o *Pod1) UnsetCurrentNodePool()`

UnsetCurrentNodePool ensures that no value is present for CurrentNodePool, not even an explicit nil
### GetRequestedNodePools

`func (o *Pod1) GetRequestedNodePools() []string`

GetRequestedNodePools returns the RequestedNodePools field if non-nil, zero value otherwise.

### GetRequestedNodePoolsOk

`func (o *Pod1) GetRequestedNodePoolsOk() (*[]string, bool)`

GetRequestedNodePoolsOk returns a tuple with the RequestedNodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedNodePools

`func (o *Pod1) SetRequestedNodePools(v []string)`

SetRequestedNodePools sets RequestedNodePools field to given value.

### HasRequestedNodePools

`func (o *Pod1) HasRequestedNodePools() bool`

HasRequestedNodePools returns a boolean if a field has been set.

### GetRequestedResources

`func (o *Pod1) GetRequestedResources() PodRequestResources`

GetRequestedResources returns the RequestedResources field if non-nil, zero value otherwise.

### GetRequestedResourcesOk

`func (o *Pod1) GetRequestedResourcesOk() (*PodRequestResources, bool)`

GetRequestedResourcesOk returns a tuple with the RequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedResources

`func (o *Pod1) SetRequestedResources(v PodRequestResources)`

SetRequestedResources sets RequestedResources field to given value.

### HasRequestedResources

`func (o *Pod1) HasRequestedResources() bool`

HasRequestedResources returns a boolean if a field has been set.

### SetRequestedResourcesNil

`func (o *Pod1) SetRequestedResourcesNil(b bool)`

 SetRequestedResourcesNil sets the value for RequestedResources to be an explicit nil

### UnsetRequestedResources
`func (o *Pod1) UnsetRequestedResources()`

UnsetRequestedResources ensures that no value is present for RequestedResources, not even an explicit nil
### GetAllocatedResources

`func (o *Pod1) GetAllocatedResources() AllocatedResources`

GetAllocatedResources returns the AllocatedResources field if non-nil, zero value otherwise.

### GetAllocatedResourcesOk

`func (o *Pod1) GetAllocatedResourcesOk() (*AllocatedResources, bool)`

GetAllocatedResourcesOk returns a tuple with the AllocatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedResources

`func (o *Pod1) SetAllocatedResources(v AllocatedResources)`

SetAllocatedResources sets AllocatedResources field to given value.

### HasAllocatedResources

`func (o *Pod1) HasAllocatedResources() bool`

HasAllocatedResources returns a boolean if a field has been set.

### SetAllocatedResourcesNil

`func (o *Pod1) SetAllocatedResourcesNil(b bool)`

 SetAllocatedResourcesNil sets the value for AllocatedResources to be an explicit nil

### UnsetAllocatedResources
`func (o *Pod1) UnsetAllocatedResources()`

UnsetAllocatedResources ensures that no value is present for AllocatedResources, not even an explicit nil
### GetTolerations

`func (o *Pod1) GetTolerations() []PodToleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *Pod1) GetTolerationsOk() (*[]PodToleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *Pod1) SetTolerations(v []PodToleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *Pod1) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### GetK8sPhase

`func (o *Pod1) GetK8sPhase() string`

GetK8sPhase returns the K8sPhase field if non-nil, zero value otherwise.

### GetK8sPhaseOk

`func (o *Pod1) GetK8sPhaseOk() (*string, bool)`

GetK8sPhaseOk returns a tuple with the K8sPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhase

`func (o *Pod1) SetK8sPhase(v string)`

SetK8sPhase sets K8sPhase field to given value.


### GetTenantId

`func (o *Pod1) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Pod1) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Pod1) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetK8sPhaseUpdatedAt

`func (o *Pod1) GetK8sPhaseUpdatedAt() time.Time`

GetK8sPhaseUpdatedAt returns the K8sPhaseUpdatedAt field if non-nil, zero value otherwise.

### GetK8sPhaseUpdatedAtOk

`func (o *Pod1) GetK8sPhaseUpdatedAtOk() (*time.Time, bool)`

GetK8sPhaseUpdatedAtOk returns a tuple with the K8sPhaseUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetK8sPhaseUpdatedAt

`func (o *Pod1) SetK8sPhaseUpdatedAt(v time.Time)`

SetK8sPhaseUpdatedAt sets K8sPhaseUpdatedAt field to given value.


### GetUpdatedAt

`func (o *Pod1) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Pod1) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Pod1) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *Pod1) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *Pod1) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *Pod1) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *Pod1) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *Pod1) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *Pod1) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetProjectName

`func (o *Pod1) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *Pod1) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *Pod1) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *Pod1) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetWorkloadName

`func (o *Pod1) GetWorkloadName() string`

GetWorkloadName returns the WorkloadName field if non-nil, zero value otherwise.

### GetWorkloadNameOk

`func (o *Pod1) GetWorkloadNameOk() (*string, bool)`

GetWorkloadNameOk returns a tuple with the WorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadName

`func (o *Pod1) SetWorkloadName(v string)`

SetWorkloadName sets WorkloadName field to given value.

### HasWorkloadName

`func (o *Pod1) HasWorkloadName() bool`

HasWorkloadName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


