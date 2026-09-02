# TotalResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GpuQuota** | Pointer to **NullableFloat64** | Total gpu quota of the org unit | [optional] 
**CpuQuota** | Pointer to **NullableFloat64** | Total cpu quota of the org unit in Millicores | [optional] 
**MemoryQuota** | Pointer to **NullableFloat64** | Total memory quota of the org unit in MB | [optional] 

## Methods

### NewTotalResources

`func NewTotalResources() *TotalResources`

NewTotalResources instantiates a new TotalResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTotalResourcesWithDefaults

`func NewTotalResourcesWithDefaults() *TotalResources`

NewTotalResourcesWithDefaults instantiates a new TotalResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGpuQuota

`func (o *TotalResources) GetGpuQuota() float64`

GetGpuQuota returns the GpuQuota field if non-nil, zero value otherwise.

### GetGpuQuotaOk

`func (o *TotalResources) GetGpuQuotaOk() (*float64, bool)`

GetGpuQuotaOk returns a tuple with the GpuQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpuQuota

`func (o *TotalResources) SetGpuQuota(v float64)`

SetGpuQuota sets GpuQuota field to given value.

### HasGpuQuota

`func (o *TotalResources) HasGpuQuota() bool`

HasGpuQuota returns a boolean if a field has been set.

### SetGpuQuotaNil

`func (o *TotalResources) SetGpuQuotaNil(b bool)`

 SetGpuQuotaNil sets the value for GpuQuota to be an explicit nil

### UnsetGpuQuota
`func (o *TotalResources) UnsetGpuQuota()`

UnsetGpuQuota ensures that no value is present for GpuQuota, not even an explicit nil
### GetCpuQuota

`func (o *TotalResources) GetCpuQuota() float64`

GetCpuQuota returns the CpuQuota field if non-nil, zero value otherwise.

### GetCpuQuotaOk

`func (o *TotalResources) GetCpuQuotaOk() (*float64, bool)`

GetCpuQuotaOk returns a tuple with the CpuQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuQuota

`func (o *TotalResources) SetCpuQuota(v float64)`

SetCpuQuota sets CpuQuota field to given value.

### HasCpuQuota

`func (o *TotalResources) HasCpuQuota() bool`

HasCpuQuota returns a boolean if a field has been set.

### SetCpuQuotaNil

`func (o *TotalResources) SetCpuQuotaNil(b bool)`

 SetCpuQuotaNil sets the value for CpuQuota to be an explicit nil

### UnsetCpuQuota
`func (o *TotalResources) UnsetCpuQuota()`

UnsetCpuQuota ensures that no value is present for CpuQuota, not even an explicit nil
### GetMemoryQuota

`func (o *TotalResources) GetMemoryQuota() float64`

GetMemoryQuota returns the MemoryQuota field if non-nil, zero value otherwise.

### GetMemoryQuotaOk

`func (o *TotalResources) GetMemoryQuotaOk() (*float64, bool)`

GetMemoryQuotaOk returns a tuple with the MemoryQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemoryQuota

`func (o *TotalResources) SetMemoryQuota(v float64)`

SetMemoryQuota sets MemoryQuota field to given value.

### HasMemoryQuota

`func (o *TotalResources) HasMemoryQuota() bool`

HasMemoryQuota returns a boolean if a field has been set.

### SetMemoryQuotaNil

`func (o *TotalResources) SetMemoryQuotaNil(b bool)`

 SetMemoryQuotaNil sets the value for MemoryQuota to be an explicit nil

### UnsetMemoryQuota
`func (o *TotalResources) UnsetMemoryQuota()`

UnsetMemoryQuota ensures that no value is present for MemoryQuota, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


