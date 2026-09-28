# WorkloadTypeCreateUpdateInnerFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CategoryId** | Pointer to **string** | The unique identifier of the workload category. | [optional] 
**PriorityId** | Pointer to **string** | The unique identifier of the workload priority. | [optional] 
**Preemptibility** | Pointer to [**NullableWorkloadTypePreemptibility**](WorkloadTypePreemptibility.md) |  | [optional] 

## Methods

### NewWorkloadTypeCreateUpdateInnerFields

`func NewWorkloadTypeCreateUpdateInnerFields() *WorkloadTypeCreateUpdateInnerFields`

NewWorkloadTypeCreateUpdateInnerFields instantiates a new WorkloadTypeCreateUpdateInnerFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeCreateUpdateInnerFieldsWithDefaults

`func NewWorkloadTypeCreateUpdateInnerFieldsWithDefaults() *WorkloadTypeCreateUpdateInnerFields`

NewWorkloadTypeCreateUpdateInnerFieldsWithDefaults instantiates a new WorkloadTypeCreateUpdateInnerFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategoryId

`func (o *WorkloadTypeCreateUpdateInnerFields) GetCategoryId() string`

GetCategoryId returns the CategoryId field if non-nil, zero value otherwise.

### GetCategoryIdOk

`func (o *WorkloadTypeCreateUpdateInnerFields) GetCategoryIdOk() (*string, bool)`

GetCategoryIdOk returns a tuple with the CategoryId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategoryId

`func (o *WorkloadTypeCreateUpdateInnerFields) SetCategoryId(v string)`

SetCategoryId sets CategoryId field to given value.

### HasCategoryId

`func (o *WorkloadTypeCreateUpdateInnerFields) HasCategoryId() bool`

HasCategoryId returns a boolean if a field has been set.

### GetPriorityId

`func (o *WorkloadTypeCreateUpdateInnerFields) GetPriorityId() string`

GetPriorityId returns the PriorityId field if non-nil, zero value otherwise.

### GetPriorityIdOk

`func (o *WorkloadTypeCreateUpdateInnerFields) GetPriorityIdOk() (*string, bool)`

GetPriorityIdOk returns a tuple with the PriorityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityId

`func (o *WorkloadTypeCreateUpdateInnerFields) SetPriorityId(v string)`

SetPriorityId sets PriorityId field to given value.

### HasPriorityId

`func (o *WorkloadTypeCreateUpdateInnerFields) HasPriorityId() bool`

HasPriorityId returns a boolean if a field has been set.

### GetPreemptibility

`func (o *WorkloadTypeCreateUpdateInnerFields) GetPreemptibility() WorkloadTypePreemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadTypeCreateUpdateInnerFields) GetPreemptibilityOk() (*WorkloadTypePreemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadTypeCreateUpdateInnerFields) SetPreemptibility(v WorkloadTypePreemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkloadTypeCreateUpdateInnerFields) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkloadTypeCreateUpdateInnerFields) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkloadTypeCreateUpdateInnerFields) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


