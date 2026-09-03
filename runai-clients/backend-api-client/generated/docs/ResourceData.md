# ResourceData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Quota** | **float32** |  | 
**Allocated** | **float32** |  | 
**Utilization** | Pointer to **float32** |  | [optional] 

## Methods

### NewResourceData

`func NewResourceData(quota float32, allocated float32, ) *ResourceData`

NewResourceData instantiates a new ResourceData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceDataWithDefaults

`func NewResourceDataWithDefaults() *ResourceData`

NewResourceDataWithDefaults instantiates a new ResourceData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetQuota

`func (o *ResourceData) GetQuota() float32`

GetQuota returns the Quota field if non-nil, zero value otherwise.

### GetQuotaOk

`func (o *ResourceData) GetQuotaOk() (*float32, bool)`

GetQuotaOk returns a tuple with the Quota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuota

`func (o *ResourceData) SetQuota(v float32)`

SetQuota sets Quota field to given value.


### GetAllocated

`func (o *ResourceData) GetAllocated() float32`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *ResourceData) GetAllocatedOk() (*float32, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *ResourceData) SetAllocated(v float32)`

SetAllocated sets Allocated field to given value.


### GetUtilization

`func (o *ResourceData) GetUtilization() float32`

GetUtilization returns the Utilization field if non-nil, zero value otherwise.

### GetUtilizationOk

`func (o *ResourceData) GetUtilizationOk() (*float32, bool)`

GetUtilizationOk returns a tuple with the Utilization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUtilization

`func (o *ResourceData) SetUtilization(v float32)`

SetUtilization sets Utilization field to given value.

### HasUtilization

`func (o *ResourceData) HasUtilization() bool`

HasUtilization returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


