# WorkloadTypeConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CategoryId** | **string** | The unique identifier of the workload category. | 
**PriorityId** | **string** | The unique identifier of the workload priority. | 
**Preemptibility** | [**WorkloadTypePreemptibility**](WorkloadTypePreemptibility.md) |  | 
**Name** | **string** | The unique name of the workload type. This value must exactly match the Kubernetes Kind that represents the workload type. | 
**Group** | **string** | The Kubernetes group associated with the workload resource. | 
**Id** | Pointer to **string** | The unique identifier of the workload type. | [optional] 
**CategoryName** | Pointer to **string** | The name of the workload category. | [optional] 
**PriorityName** | Pointer to **string** | The name of the workload priority. | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The timestamp for the last time the workload type was updated. | [optional] 
**UpdatedBy** | Pointer to **string** | Identifier of the user who last updated the workload type. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The timestamp for when the workload type was created. | [optional] 
**CreatedBy** | Pointer to **string** | Identifier of the user who created the workload type. | [optional] 
**ResourceInterfaces** | Pointer to [**[]WorkloadTypeResourceInterfacesWithVersionInner**](WorkloadTypeResourceInterfacesWithVersionInner.md) | Deprecated. Use kartas instead. Lists the versions of the custom resource definition (CRD) supported for this workload type, such as v1, v1beta1, or v1alpha1. This enables the platform to correctly parse, interpret, and manage manifests for this workload type according to the specific structure and schema associated with each listed version. | [optional] 
**Kartas** | Pointer to [**[]WorkloadTypeKartasWithVersionInner**](WorkloadTypeKartasWithVersionInner.md) | Lists the versions of the custom resource definition (CRD) supported for this workload type, such as v1, v1beta1, or v1alpha1. This enables the platform to correctly parse, interpret, and manage manifests for this workload type according to the specific structure and schema associated with each listed version. | [optional] 
**Source** | Pointer to **string** | The source of the workload type - &#39;native&#39;, &#39;custom&#39; or &#39;external&#39;. | [optional] 
**ClusterStatuses** | Pointer to [**[]WorkloadTypeStatus**](WorkloadTypeStatus.md) |  | [optional] 

## Methods

### NewWorkloadTypeConfig

`func NewWorkloadTypeConfig(categoryId string, priorityId string, preemptibility WorkloadTypePreemptibility, name string, group string, ) *WorkloadTypeConfig`

NewWorkloadTypeConfig instantiates a new WorkloadTypeConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeConfigWithDefaults

`func NewWorkloadTypeConfigWithDefaults() *WorkloadTypeConfig`

NewWorkloadTypeConfigWithDefaults instantiates a new WorkloadTypeConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategoryId

`func (o *WorkloadTypeConfig) GetCategoryId() string`

GetCategoryId returns the CategoryId field if non-nil, zero value otherwise.

### GetCategoryIdOk

`func (o *WorkloadTypeConfig) GetCategoryIdOk() (*string, bool)`

GetCategoryIdOk returns a tuple with the CategoryId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategoryId

`func (o *WorkloadTypeConfig) SetCategoryId(v string)`

SetCategoryId sets CategoryId field to given value.


### GetPriorityId

`func (o *WorkloadTypeConfig) GetPriorityId() string`

GetPriorityId returns the PriorityId field if non-nil, zero value otherwise.

### GetPriorityIdOk

`func (o *WorkloadTypeConfig) GetPriorityIdOk() (*string, bool)`

GetPriorityIdOk returns a tuple with the PriorityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityId

`func (o *WorkloadTypeConfig) SetPriorityId(v string)`

SetPriorityId sets PriorityId field to given value.


### GetPreemptibility

`func (o *WorkloadTypeConfig) GetPreemptibility() WorkloadTypePreemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadTypeConfig) GetPreemptibilityOk() (*WorkloadTypePreemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadTypeConfig) SetPreemptibility(v WorkloadTypePreemptibility)`

SetPreemptibility sets Preemptibility field to given value.


### GetName

`func (o *WorkloadTypeConfig) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadTypeConfig) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadTypeConfig) SetName(v string)`

SetName sets Name field to given value.


### GetGroup

`func (o *WorkloadTypeConfig) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *WorkloadTypeConfig) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *WorkloadTypeConfig) SetGroup(v string)`

SetGroup sets Group field to given value.


### GetId

`func (o *WorkloadTypeConfig) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadTypeConfig) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadTypeConfig) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *WorkloadTypeConfig) HasId() bool`

HasId returns a boolean if a field has been set.

### GetCategoryName

`func (o *WorkloadTypeConfig) GetCategoryName() string`

GetCategoryName returns the CategoryName field if non-nil, zero value otherwise.

### GetCategoryNameOk

`func (o *WorkloadTypeConfig) GetCategoryNameOk() (*string, bool)`

GetCategoryNameOk returns a tuple with the CategoryName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategoryName

`func (o *WorkloadTypeConfig) SetCategoryName(v string)`

SetCategoryName sets CategoryName field to given value.

### HasCategoryName

`func (o *WorkloadTypeConfig) HasCategoryName() bool`

HasCategoryName returns a boolean if a field has been set.

### GetPriorityName

`func (o *WorkloadTypeConfig) GetPriorityName() string`

GetPriorityName returns the PriorityName field if non-nil, zero value otherwise.

### GetPriorityNameOk

`func (o *WorkloadTypeConfig) GetPriorityNameOk() (*string, bool)`

GetPriorityNameOk returns a tuple with the PriorityName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityName

`func (o *WorkloadTypeConfig) SetPriorityName(v string)`

SetPriorityName sets PriorityName field to given value.

### HasPriorityName

`func (o *WorkloadTypeConfig) HasPriorityName() bool`

HasPriorityName returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkloadTypeConfig) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkloadTypeConfig) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkloadTypeConfig) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkloadTypeConfig) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *WorkloadTypeConfig) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *WorkloadTypeConfig) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *WorkloadTypeConfig) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *WorkloadTypeConfig) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetCreatedAt

`func (o *WorkloadTypeConfig) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadTypeConfig) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadTypeConfig) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkloadTypeConfig) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *WorkloadTypeConfig) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *WorkloadTypeConfig) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *WorkloadTypeConfig) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *WorkloadTypeConfig) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetResourceInterfaces

`func (o *WorkloadTypeConfig) GetResourceInterfaces() []WorkloadTypeResourceInterfacesWithVersionInner`

GetResourceInterfaces returns the ResourceInterfaces field if non-nil, zero value otherwise.

### GetResourceInterfacesOk

`func (o *WorkloadTypeConfig) GetResourceInterfacesOk() (*[]WorkloadTypeResourceInterfacesWithVersionInner, bool)`

GetResourceInterfacesOk returns a tuple with the ResourceInterfaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceInterfaces

`func (o *WorkloadTypeConfig) SetResourceInterfaces(v []WorkloadTypeResourceInterfacesWithVersionInner)`

SetResourceInterfaces sets ResourceInterfaces field to given value.

### HasResourceInterfaces

`func (o *WorkloadTypeConfig) HasResourceInterfaces() bool`

HasResourceInterfaces returns a boolean if a field has been set.

### GetKartas

`func (o *WorkloadTypeConfig) GetKartas() []WorkloadTypeKartasWithVersionInner`

GetKartas returns the Kartas field if non-nil, zero value otherwise.

### GetKartasOk

`func (o *WorkloadTypeConfig) GetKartasOk() (*[]WorkloadTypeKartasWithVersionInner, bool)`

GetKartasOk returns a tuple with the Kartas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKartas

`func (o *WorkloadTypeConfig) SetKartas(v []WorkloadTypeKartasWithVersionInner)`

SetKartas sets Kartas field to given value.

### HasKartas

`func (o *WorkloadTypeConfig) HasKartas() bool`

HasKartas returns a boolean if a field has been set.

### GetSource

`func (o *WorkloadTypeConfig) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *WorkloadTypeConfig) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *WorkloadTypeConfig) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *WorkloadTypeConfig) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetClusterStatuses

`func (o *WorkloadTypeConfig) GetClusterStatuses() []WorkloadTypeStatus`

GetClusterStatuses returns the ClusterStatuses field if non-nil, zero value otherwise.

### GetClusterStatusesOk

`func (o *WorkloadTypeConfig) GetClusterStatusesOk() (*[]WorkloadTypeStatus, bool)`

GetClusterStatusesOk returns a tuple with the ClusterStatuses field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterStatuses

`func (o *WorkloadTypeConfig) SetClusterStatuses(v []WorkloadTypeStatus)`

SetClusterStatuses sets ClusterStatuses field to given value.

### HasClusterStatuses

`func (o *WorkloadTypeConfig) HasClusterStatuses() bool`

HasClusterStatuses returns a boolean if a field has been set.

### SetClusterStatusesNil

`func (o *WorkloadTypeConfig) SetClusterStatusesNil(b bool)`

 SetClusterStatusesNil sets the value for ClusterStatuses to be an explicit nil

### UnsetClusterStatuses
`func (o *WorkloadTypeConfig) UnsetClusterStatuses()`

UnsetClusterStatuses ensures that no value is present for ClusterStatuses, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


