# WorkloadCategory

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier of the workload category. | 
**Name** | **string** | A unique, human-readable name for the workload category. | 
**Description** | Pointer to **string** | An optional description providing additional details about the category. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The date and time when the workload category was created. | [optional] 
**CreatedBy** | Pointer to **string** | The user who created the workload category. | [optional] 

## Methods

### NewWorkloadCategory

`func NewWorkloadCategory(id string, name string, ) *WorkloadCategory`

NewWorkloadCategory instantiates a new WorkloadCategory object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadCategoryWithDefaults

`func NewWorkloadCategoryWithDefaults() *WorkloadCategory`

NewWorkloadCategoryWithDefaults instantiates a new WorkloadCategory object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadCategory) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadCategory) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadCategory) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *WorkloadCategory) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadCategory) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadCategory) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *WorkloadCategory) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *WorkloadCategory) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *WorkloadCategory) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *WorkloadCategory) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetCreatedAt

`func (o *WorkloadCategory) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadCategory) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadCategory) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkloadCategory) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *WorkloadCategory) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *WorkloadCategory) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *WorkloadCategory) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *WorkloadCategory) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


