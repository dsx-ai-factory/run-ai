# WorkloadTypeAdditionalResponseFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CategoryId** | Pointer to **string** | The unique identifier of the workload category. | [optional] 
**PriorityId** | Pointer to **string** | The unique identifier of the workload priority. | [optional] 
**Preemptibility** | Pointer to [**WorkloadTypePreemptibility**](WorkloadTypePreemptibility.md) |  | [optional] 

## Methods

### NewWorkloadTypeAdditionalResponseFields

`func NewWorkloadTypeAdditionalResponseFields() *WorkloadTypeAdditionalResponseFields`

NewWorkloadTypeAdditionalResponseFields instantiates a new WorkloadTypeAdditionalResponseFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeAdditionalResponseFieldsWithDefaults

`func NewWorkloadTypeAdditionalResponseFieldsWithDefaults() *WorkloadTypeAdditionalResponseFields`

NewWorkloadTypeAdditionalResponseFieldsWithDefaults instantiates a new WorkloadTypeAdditionalResponseFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategoryId

`func (o *WorkloadTypeAdditionalResponseFields) GetCategoryId() string`

GetCategoryId returns the CategoryId field if non-nil, zero value otherwise.

### GetCategoryIdOk

`func (o *WorkloadTypeAdditionalResponseFields) GetCategoryIdOk() (*string, bool)`

GetCategoryIdOk returns a tuple with the CategoryId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategoryId

`func (o *WorkloadTypeAdditionalResponseFields) SetCategoryId(v string)`

SetCategoryId sets CategoryId field to given value.

### HasCategoryId

`func (o *WorkloadTypeAdditionalResponseFields) HasCategoryId() bool`

HasCategoryId returns a boolean if a field has been set.

### GetPriorityId

`func (o *WorkloadTypeAdditionalResponseFields) GetPriorityId() string`

GetPriorityId returns the PriorityId field if non-nil, zero value otherwise.

### GetPriorityIdOk

`func (o *WorkloadTypeAdditionalResponseFields) GetPriorityIdOk() (*string, bool)`

GetPriorityIdOk returns a tuple with the PriorityId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityId

`func (o *WorkloadTypeAdditionalResponseFields) SetPriorityId(v string)`

SetPriorityId sets PriorityId field to given value.

### HasPriorityId

`func (o *WorkloadTypeAdditionalResponseFields) HasPriorityId() bool`

HasPriorityId returns a boolean if a field has been set.

### GetPreemptibility

`func (o *WorkloadTypeAdditionalResponseFields) GetPreemptibility() WorkloadTypePreemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadTypeAdditionalResponseFields) GetPreemptibilityOk() (*WorkloadTypePreemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadTypeAdditionalResponseFields) SetPreemptibility(v WorkloadTypePreemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkloadTypeAdditionalResponseFields) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


