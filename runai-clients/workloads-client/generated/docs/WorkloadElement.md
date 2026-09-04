# WorkloadElement

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

## Methods

### NewWorkloadElement

`func NewWorkloadElement(id string, name string, hierarchyLevel int32, ) *WorkloadElement`

NewWorkloadElement instantiates a new WorkloadElement object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadElementWithDefaults

`func NewWorkloadElementWithDefaults() *WorkloadElement`

NewWorkloadElementWithDefaults instantiates a new WorkloadElement object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadElement) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadElement) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadElement) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *WorkloadElement) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadElement) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadElement) SetName(v string)`

SetName sets Name field to given value.


### GetHierarchyLevel

`func (o *WorkloadElement) GetHierarchyLevel() int32`

GetHierarchyLevel returns the HierarchyLevel field if non-nil, zero value otherwise.

### GetHierarchyLevelOk

`func (o *WorkloadElement) GetHierarchyLevelOk() (*int32, bool)`

GetHierarchyLevelOk returns a tuple with the HierarchyLevel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHierarchyLevel

`func (o *WorkloadElement) SetHierarchyLevel(v int32)`

SetHierarchyLevel sets HierarchyLevel field to given value.


### GetParentId

`func (o *WorkloadElement) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *WorkloadElement) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *WorkloadElement) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *WorkloadElement) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### SetParentIdNil

`func (o *WorkloadElement) SetParentIdNil(b bool)`

 SetParentIdNil sets the value for ParentId to be an explicit nil

### UnsetParentId
`func (o *WorkloadElement) UnsetParentId()`

UnsetParentId ensures that no value is present for ParentId, not even an explicit nil
### GetInstanceKey

`func (o *WorkloadElement) GetInstanceKey() string`

GetInstanceKey returns the InstanceKey field if non-nil, zero value otherwise.

### GetInstanceKeyOk

`func (o *WorkloadElement) GetInstanceKeyOk() (*string, bool)`

GetInstanceKeyOk returns a tuple with the InstanceKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceKey

`func (o *WorkloadElement) SetInstanceKey(v string)`

SetInstanceKey sets InstanceKey field to given value.

### HasInstanceKey

`func (o *WorkloadElement) HasInstanceKey() bool`

HasInstanceKey returns a boolean if a field has been set.

### SetInstanceKeyNil

`func (o *WorkloadElement) SetInstanceKeyNil(b bool)`

 SetInstanceKeyNil sets the value for InstanceKey to be an explicit nil

### UnsetInstanceKey
`func (o *WorkloadElement) UnsetInstanceKey()`

UnsetInstanceKey ensures that no value is present for InstanceKey, not even an explicit nil
### GetReplicaKey

`func (o *WorkloadElement) GetReplicaKey() string`

GetReplicaKey returns the ReplicaKey field if non-nil, zero value otherwise.

### GetReplicaKeyOk

`func (o *WorkloadElement) GetReplicaKeyOk() (*string, bool)`

GetReplicaKeyOk returns a tuple with the ReplicaKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicaKey

`func (o *WorkloadElement) SetReplicaKey(v string)`

SetReplicaKey sets ReplicaKey field to given value.

### HasReplicaKey

`func (o *WorkloadElement) HasReplicaKey() bool`

HasReplicaKey returns a boolean if a field has been set.

### SetReplicaKeyNil

`func (o *WorkloadElement) SetReplicaKeyNil(b bool)`

 SetReplicaKeyNil sets the value for ReplicaKey to be an explicit nil

### UnsetReplicaKey
`func (o *WorkloadElement) UnsetReplicaKey()`

UnsetReplicaKey ensures that no value is present for ReplicaKey, not even an explicit nil
### GetCreatedAt

`func (o *WorkloadElement) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadElement) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadElement) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkloadElement) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetKubernetesRef

`func (o *WorkloadElement) GetKubernetesRef() WorkloadElementKubernetesRef`

GetKubernetesRef returns the KubernetesRef field if non-nil, zero value otherwise.

### GetKubernetesRefOk

`func (o *WorkloadElement) GetKubernetesRefOk() (*WorkloadElementKubernetesRef, bool)`

GetKubernetesRefOk returns a tuple with the KubernetesRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKubernetesRef

`func (o *WorkloadElement) SetKubernetesRef(v WorkloadElementKubernetesRef)`

SetKubernetesRef sets KubernetesRef field to given value.

### HasKubernetesRef

`func (o *WorkloadElement) HasKubernetesRef() bool`

HasKubernetesRef returns a boolean if a field has been set.

### SetKubernetesRefNil

`func (o *WorkloadElement) SetKubernetesRefNil(b bool)`

 SetKubernetesRefNil sets the value for KubernetesRef to be an explicit nil

### UnsetKubernetesRef
`func (o *WorkloadElement) UnsetKubernetesRef()`

UnsetKubernetesRef ensures that no value is present for KubernetesRef, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


