# WorkloadElementResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier of the structure element. | 
**Name** | **string** | The display name of the structure element. | 
**HierarchyLevel** | **int32** | The position of the element in the workload structure hierarchy. Valid values are 1 (top-level), 2, and 3. | 
**ParentId** | Pointer to **NullableString** | The unique identifier of the parent element. Empty for level 1 elements, whose parent is implicitly the workload. | [optional] 
**InstanceKey** | Pointer to **NullableString** | Instance identifier from ComponentInstanceSelector (e.g., Frontend, PrefillWorker) | [optional] 
**ReplicaKey** | Pointer to **NullableString** | Replica identifier if applicable | [optional] 
**CreatedAt** | Pointer to **time.Time** | The timestamp when the structure element was created. | [optional] 
**KubernetesRef** | Pointer to [**NullableWorkloadElementKubernetesRef**](WorkloadElementKubernetesRef.md) |  | [optional] 
**PodsRequestedResources** | [**NullableWorkloadRequestResources**](WorkloadRequestResources.md) |  | 
**AllocatedResources** | [**NullableWorkloadAllocatedResources**](WorkloadAllocatedResources.md) |  | 
**Topology** | Pointer to [**NullableWorkloadElementTopology**](WorkloadElementTopology.md) |  | [optional] 

## Methods

### NewWorkloadElementResponse

`func NewWorkloadElementResponse(id string, name string, hierarchyLevel int32, podsRequestedResources NullableWorkloadRequestResources, allocatedResources NullableWorkloadAllocatedResources, ) *WorkloadElementResponse`

NewWorkloadElementResponse instantiates a new WorkloadElementResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadElementResponseWithDefaults

`func NewWorkloadElementResponseWithDefaults() *WorkloadElementResponse`

NewWorkloadElementResponseWithDefaults instantiates a new WorkloadElementResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadElementResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadElementResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadElementResponse) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *WorkloadElementResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadElementResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadElementResponse) SetName(v string)`

SetName sets Name field to given value.


### GetHierarchyLevel

`func (o *WorkloadElementResponse) GetHierarchyLevel() int32`

GetHierarchyLevel returns the HierarchyLevel field if non-nil, zero value otherwise.

### GetHierarchyLevelOk

`func (o *WorkloadElementResponse) GetHierarchyLevelOk() (*int32, bool)`

GetHierarchyLevelOk returns a tuple with the HierarchyLevel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHierarchyLevel

`func (o *WorkloadElementResponse) SetHierarchyLevel(v int32)`

SetHierarchyLevel sets HierarchyLevel field to given value.


### GetParentId

`func (o *WorkloadElementResponse) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *WorkloadElementResponse) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *WorkloadElementResponse) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *WorkloadElementResponse) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### SetParentIdNil

`func (o *WorkloadElementResponse) SetParentIdNil(b bool)`

 SetParentIdNil sets the value for ParentId to be an explicit nil

### UnsetParentId
`func (o *WorkloadElementResponse) UnsetParentId()`

UnsetParentId ensures that no value is present for ParentId, not even an explicit nil
### GetInstanceKey

`func (o *WorkloadElementResponse) GetInstanceKey() string`

GetInstanceKey returns the InstanceKey field if non-nil, zero value otherwise.

### GetInstanceKeyOk

`func (o *WorkloadElementResponse) GetInstanceKeyOk() (*string, bool)`

GetInstanceKeyOk returns a tuple with the InstanceKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceKey

`func (o *WorkloadElementResponse) SetInstanceKey(v string)`

SetInstanceKey sets InstanceKey field to given value.

### HasInstanceKey

`func (o *WorkloadElementResponse) HasInstanceKey() bool`

HasInstanceKey returns a boolean if a field has been set.

### SetInstanceKeyNil

`func (o *WorkloadElementResponse) SetInstanceKeyNil(b bool)`

 SetInstanceKeyNil sets the value for InstanceKey to be an explicit nil

### UnsetInstanceKey
`func (o *WorkloadElementResponse) UnsetInstanceKey()`

UnsetInstanceKey ensures that no value is present for InstanceKey, not even an explicit nil
### GetReplicaKey

`func (o *WorkloadElementResponse) GetReplicaKey() string`

GetReplicaKey returns the ReplicaKey field if non-nil, zero value otherwise.

### GetReplicaKeyOk

`func (o *WorkloadElementResponse) GetReplicaKeyOk() (*string, bool)`

GetReplicaKeyOk returns a tuple with the ReplicaKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicaKey

`func (o *WorkloadElementResponse) SetReplicaKey(v string)`

SetReplicaKey sets ReplicaKey field to given value.

### HasReplicaKey

`func (o *WorkloadElementResponse) HasReplicaKey() bool`

HasReplicaKey returns a boolean if a field has been set.

### SetReplicaKeyNil

`func (o *WorkloadElementResponse) SetReplicaKeyNil(b bool)`

 SetReplicaKeyNil sets the value for ReplicaKey to be an explicit nil

### UnsetReplicaKey
`func (o *WorkloadElementResponse) UnsetReplicaKey()`

UnsetReplicaKey ensures that no value is present for ReplicaKey, not even an explicit nil
### GetCreatedAt

`func (o *WorkloadElementResponse) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadElementResponse) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadElementResponse) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkloadElementResponse) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetKubernetesRef

`func (o *WorkloadElementResponse) GetKubernetesRef() WorkloadElementKubernetesRef`

GetKubernetesRef returns the KubernetesRef field if non-nil, zero value otherwise.

### GetKubernetesRefOk

`func (o *WorkloadElementResponse) GetKubernetesRefOk() (*WorkloadElementKubernetesRef, bool)`

GetKubernetesRefOk returns a tuple with the KubernetesRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKubernetesRef

`func (o *WorkloadElementResponse) SetKubernetesRef(v WorkloadElementKubernetesRef)`

SetKubernetesRef sets KubernetesRef field to given value.

### HasKubernetesRef

`func (o *WorkloadElementResponse) HasKubernetesRef() bool`

HasKubernetesRef returns a boolean if a field has been set.

### SetKubernetesRefNil

`func (o *WorkloadElementResponse) SetKubernetesRefNil(b bool)`

 SetKubernetesRefNil sets the value for KubernetesRef to be an explicit nil

### UnsetKubernetesRef
`func (o *WorkloadElementResponse) UnsetKubernetesRef()`

UnsetKubernetesRef ensures that no value is present for KubernetesRef, not even an explicit nil
### GetPodsRequestedResources

`func (o *WorkloadElementResponse) GetPodsRequestedResources() WorkloadRequestResources`

GetPodsRequestedResources returns the PodsRequestedResources field if non-nil, zero value otherwise.

### GetPodsRequestedResourcesOk

`func (o *WorkloadElementResponse) GetPodsRequestedResourcesOk() (*WorkloadRequestResources, bool)`

GetPodsRequestedResourcesOk returns a tuple with the PodsRequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodsRequestedResources

`func (o *WorkloadElementResponse) SetPodsRequestedResources(v WorkloadRequestResources)`

SetPodsRequestedResources sets PodsRequestedResources field to given value.


### SetPodsRequestedResourcesNil

`func (o *WorkloadElementResponse) SetPodsRequestedResourcesNil(b bool)`

 SetPodsRequestedResourcesNil sets the value for PodsRequestedResources to be an explicit nil

### UnsetPodsRequestedResources
`func (o *WorkloadElementResponse) UnsetPodsRequestedResources()`

UnsetPodsRequestedResources ensures that no value is present for PodsRequestedResources, not even an explicit nil
### GetAllocatedResources

`func (o *WorkloadElementResponse) GetAllocatedResources() WorkloadAllocatedResources`

GetAllocatedResources returns the AllocatedResources field if non-nil, zero value otherwise.

### GetAllocatedResourcesOk

`func (o *WorkloadElementResponse) GetAllocatedResourcesOk() (*WorkloadAllocatedResources, bool)`

GetAllocatedResourcesOk returns a tuple with the AllocatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedResources

`func (o *WorkloadElementResponse) SetAllocatedResources(v WorkloadAllocatedResources)`

SetAllocatedResources sets AllocatedResources field to given value.


### SetAllocatedResourcesNil

`func (o *WorkloadElementResponse) SetAllocatedResourcesNil(b bool)`

 SetAllocatedResourcesNil sets the value for AllocatedResources to be an explicit nil

### UnsetAllocatedResources
`func (o *WorkloadElementResponse) UnsetAllocatedResources()`

UnsetAllocatedResources ensures that no value is present for AllocatedResources, not even an explicit nil
### GetTopology

`func (o *WorkloadElementResponse) GetTopology() WorkloadElementTopology`

GetTopology returns the Topology field if non-nil, zero value otherwise.

### GetTopologyOk

`func (o *WorkloadElementResponse) GetTopologyOk() (*WorkloadElementTopology, bool)`

GetTopologyOk returns a tuple with the Topology field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopology

`func (o *WorkloadElementResponse) SetTopology(v WorkloadElementTopology)`

SetTopology sets Topology field to given value.

### HasTopology

`func (o *WorkloadElementResponse) HasTopology() bool`

HasTopology returns a boolean if a field has been set.

### SetTopologyNil

`func (o *WorkloadElementResponse) SetTopologyNil(b bool)`

 SetTopologyNil sets the value for Topology to be an explicit nil

### UnsetTopology
`func (o *WorkloadElementResponse) UnsetTopology()`

UnsetTopology ensures that no value is present for Topology, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


