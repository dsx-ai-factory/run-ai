# JobResourceData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Allocated** | **float32** |  | 
**Utilization** | Pointer to **float32** | Utilization percentage, values between 0 to 1. | [optional] 

## Methods

### NewJobResourceData

`func NewJobResourceData(allocated float32, ) *JobResourceData`

NewJobResourceData instantiates a new JobResourceData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobResourceDataWithDefaults

`func NewJobResourceDataWithDefaults() *JobResourceData`

NewJobResourceDataWithDefaults instantiates a new JobResourceData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllocated

`func (o *JobResourceData) GetAllocated() float32`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *JobResourceData) GetAllocatedOk() (*float32, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *JobResourceData) SetAllocated(v float32)`

SetAllocated sets Allocated field to given value.


### GetUtilization

`func (o *JobResourceData) GetUtilization() float32`

GetUtilization returns the Utilization field if non-nil, zero value otherwise.

### GetUtilizationOk

`func (o *JobResourceData) GetUtilizationOk() (*float32, bool)`

GetUtilizationOk returns a tuple with the Utilization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUtilization

`func (o *JobResourceData) SetUtilization(v float32)`

SetUtilization sets Utilization field to given value.

### HasUtilization

`func (o *JobResourceData) HasUtilization() bool`

HasUtilization returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


