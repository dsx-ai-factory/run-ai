# WorkloadTypeUpdateFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CategoryId** | Pointer to **string** | The unique identifier of the workload category. | [optional] 
**PriorityId** | Pointer to **string** | The unique identifier of the workload priority. | [optional] 
**Preemptibility** | Pointer to [**NullableWorkloadTypePreemptibility**](WorkloadTypePreemptibility.md) |  | [optional] 
**ResourceInterfaces** | Pointer to **[]map[string]interface{}** | Deprecated. Use kartas instead. Lists the versions of the custom resource definition (CRD) supported for this workload type, such as v1, v1beta1, or v1alpha1. This enables the platform to correctly parse, interpret, and manage manifests for this workload type according to the specific structure and schema associated with each listed version. On update, you may only add or remove supported versions, modifying existing version entries is not allowed. | [optional] 
**Kartas** | Pointer to **[]map[string]interface{}** | Lists the versions of the custom resource definition (CRD) supported for this workload type, such as v1, v1beta1, or v1alpha1. This enables the platform to correctly parse, interpret, and manage manifests for this workload type according to the specific structure and schema associated with each listed version. On update, you may only add or remove supported versions, modifying existing version entries is not allowed. | [optional] 

## Methods

### NewWorkloadTypeUpdateFields

`func NewWorkloadTypeUpdateFields() *WorkloadTypeUpdateFields`

NewWorkloadTypeUpdateFields instantiates a new WorkloadTypeUpdateFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeUpdateFieldsWithDefaults

`func NewWorkloadTypeUpdateFieldsWithDefaults() *WorkloadTypeUpdateFields`

NewWorkloadTypeUpdateFieldsWithDefaults instantiates a new WorkloadTypeUpdateFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategoryId

`func (o *WorkloadTypeUpdateFields) GetCategoryId() string`

GetCategoryId returns the CategoryId field if non-nil, zero value otherwise.

### GetCategoryIdOk

`func (o *WorkloadTypeUpdateFields) GetCategoryIdOk() (*string, bool)`

GetCategoryIdOk returns a tuple with the CategoryId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategoryId

`func (o *WorkloadTypeUpdateFields) SetCategoryId(v string)`

SetCategoryId sets CategoryId field to given value.

### HasCategoryId

`func (o *WorkloadTypeUpdateFields) HasCategoryId() bool`

HasCategoryId returns a boolean if a field has been set.

### GetPriorityId

`func (o *WorkloadTypeUpdateFields) GetPriorityId() string`

GetPriorityId returns the PriorityId field if non-nil, zero value otherwise.

### GetPriorityIdOk

`func (o *WorkloadTypeUpdateFields) GetPriorityIdOk() (*string, bool)`

GetPriorityIdOk returns a tuple with the PriorityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityId

`func (o *WorkloadTypeUpdateFields) SetPriorityId(v string)`

SetPriorityId sets PriorityId field to given value.

### HasPriorityId

`func (o *WorkloadTypeUpdateFields) HasPriorityId() bool`

HasPriorityId returns a boolean if a field has been set.

### GetPreemptibility

`func (o *WorkloadTypeUpdateFields) GetPreemptibility() WorkloadTypePreemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadTypeUpdateFields) GetPreemptibilityOk() (*WorkloadTypePreemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadTypeUpdateFields) SetPreemptibility(v WorkloadTypePreemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkloadTypeUpdateFields) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkloadTypeUpdateFields) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkloadTypeUpdateFields) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetResourceInterfaces

`func (o *WorkloadTypeUpdateFields) GetResourceInterfaces() []map[string]interface{}`

GetResourceInterfaces returns the ResourceInterfaces field if non-nil, zero value otherwise.

### GetResourceInterfacesOk

`func (o *WorkloadTypeUpdateFields) GetResourceInterfacesOk() (*[]map[string]interface{}, bool)`

GetResourceInterfacesOk returns a tuple with the ResourceInterfaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceInterfaces

`func (o *WorkloadTypeUpdateFields) SetResourceInterfaces(v []map[string]interface{})`

SetResourceInterfaces sets ResourceInterfaces field to given value.

### HasResourceInterfaces

`func (o *WorkloadTypeUpdateFields) HasResourceInterfaces() bool`

HasResourceInterfaces returns a boolean if a field has been set.

### SetResourceInterfacesNil

`func (o *WorkloadTypeUpdateFields) SetResourceInterfacesNil(b bool)`

 SetResourceInterfacesNil sets the value for ResourceInterfaces to be an explicit nil

### UnsetResourceInterfaces
`func (o *WorkloadTypeUpdateFields) UnsetResourceInterfaces()`

UnsetResourceInterfaces ensures that no value is present for ResourceInterfaces, not even an explicit nil
### GetKartas

`func (o *WorkloadTypeUpdateFields) GetKartas() []map[string]interface{}`

GetKartas returns the Kartas field if non-nil, zero value otherwise.

### GetKartasOk

`func (o *WorkloadTypeUpdateFields) GetKartasOk() (*[]map[string]interface{}, bool)`

GetKartasOk returns a tuple with the Kartas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKartas

`func (o *WorkloadTypeUpdateFields) SetKartas(v []map[string]interface{})`

SetKartas sets Kartas field to given value.

### HasKartas

`func (o *WorkloadTypeUpdateFields) HasKartas() bool`

HasKartas returns a boolean if a field has been set.

### SetKartasNil

`func (o *WorkloadTypeUpdateFields) SetKartasNil(b bool)`

 SetKartasNil sets the value for Kartas to be an explicit nil

### UnsetKartas
`func (o *WorkloadTypeUpdateFields) UnsetKartas()`

UnsetKartas ensures that no value is present for Kartas, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


