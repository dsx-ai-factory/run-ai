# WorkloadTypeCreateFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CategoryId** | **string** | The unique identifier of the workload category. | 
**PriorityId** | **string** | The unique identifier of the workload priority. | 
**Preemptibility** | Pointer to [**NullableWorkloadTypePreemptibility**](WorkloadTypePreemptibility.md) |  | [optional] 
**Name** | **string** | The unique name of the workload type. This value must exactly match the Kubernetes Kind that represents the workload type. | 
**Group** | **string** | The Kubernetes group associated with the workload resource. | 
**ResourceInterfaces** | Pointer to **[]map[string]interface{}** | Deprecated. Use kartas instead. Lists the versions of the custom resource definition (CRD) supported for this workload type, such as v1, v1beta1, or v1alpha1. This enables the platform to correctly parse, interpret, and manage manifests for this workload type according to the specific structure and schema associated with each listed version. On update, you may only add or remove supported versions, modifying existing version entries is not allowed. | [optional] 
**Kartas** | Pointer to **[]map[string]interface{}** | Lists the versions of the custom resource definition (CRD) supported for this workload type, such as v1, v1beta1, or v1alpha1. This enables the platform to correctly parse, interpret, and manage manifests for this workload type according to the specific structure and schema associated with each listed version. On update, you may only add or remove supported versions, modifying existing version entries is not allowed. | [optional] 

## Methods

### NewWorkloadTypeCreateFields

`func NewWorkloadTypeCreateFields(categoryId string, priorityId string, name string, group string, ) *WorkloadTypeCreateFields`

NewWorkloadTypeCreateFields instantiates a new WorkloadTypeCreateFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeCreateFieldsWithDefaults

`func NewWorkloadTypeCreateFieldsWithDefaults() *WorkloadTypeCreateFields`

NewWorkloadTypeCreateFieldsWithDefaults instantiates a new WorkloadTypeCreateFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategoryId

`func (o *WorkloadTypeCreateFields) GetCategoryId() string`

GetCategoryId returns the CategoryId field if non-nil, zero value otherwise.

### GetCategoryIdOk

`func (o *WorkloadTypeCreateFields) GetCategoryIdOk() (*string, bool)`

GetCategoryIdOk returns a tuple with the CategoryId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategoryId

`func (o *WorkloadTypeCreateFields) SetCategoryId(v string)`

SetCategoryId sets CategoryId field to given value.


### GetPriorityId

`func (o *WorkloadTypeCreateFields) GetPriorityId() string`

GetPriorityId returns the PriorityId field if non-nil, zero value otherwise.

### GetPriorityIdOk

`func (o *WorkloadTypeCreateFields) GetPriorityIdOk() (*string, bool)`

GetPriorityIdOk returns a tuple with the PriorityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityId

`func (o *WorkloadTypeCreateFields) SetPriorityId(v string)`

SetPriorityId sets PriorityId field to given value.


### GetPreemptibility

`func (o *WorkloadTypeCreateFields) GetPreemptibility() WorkloadTypePreemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadTypeCreateFields) GetPreemptibilityOk() (*WorkloadTypePreemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadTypeCreateFields) SetPreemptibility(v WorkloadTypePreemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkloadTypeCreateFields) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkloadTypeCreateFields) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkloadTypeCreateFields) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetName

`func (o *WorkloadTypeCreateFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadTypeCreateFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadTypeCreateFields) SetName(v string)`

SetName sets Name field to given value.


### GetGroup

`func (o *WorkloadTypeCreateFields) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *WorkloadTypeCreateFields) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *WorkloadTypeCreateFields) SetGroup(v string)`

SetGroup sets Group field to given value.


### GetResourceInterfaces

`func (o *WorkloadTypeCreateFields) GetResourceInterfaces() []map[string]interface{}`

GetResourceInterfaces returns the ResourceInterfaces field if non-nil, zero value otherwise.

### GetResourceInterfacesOk

`func (o *WorkloadTypeCreateFields) GetResourceInterfacesOk() (*[]map[string]interface{}, bool)`

GetResourceInterfacesOk returns a tuple with the ResourceInterfaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceInterfaces

`func (o *WorkloadTypeCreateFields) SetResourceInterfaces(v []map[string]interface{})`

SetResourceInterfaces sets ResourceInterfaces field to given value.

### HasResourceInterfaces

`func (o *WorkloadTypeCreateFields) HasResourceInterfaces() bool`

HasResourceInterfaces returns a boolean if a field has been set.

### GetKartas

`func (o *WorkloadTypeCreateFields) GetKartas() []map[string]interface{}`

GetKartas returns the Kartas field if non-nil, zero value otherwise.

### GetKartasOk

`func (o *WorkloadTypeCreateFields) GetKartasOk() (*[]map[string]interface{}, bool)`

GetKartasOk returns a tuple with the Kartas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKartas

`func (o *WorkloadTypeCreateFields) SetKartas(v []map[string]interface{})`

SetKartas sets Kartas field to given value.

### HasKartas

`func (o *WorkloadTypeCreateFields) HasKartas() bool`

HasKartas returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


