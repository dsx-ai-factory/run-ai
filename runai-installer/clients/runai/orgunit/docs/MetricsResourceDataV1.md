# MetricsResourceDataV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Quota** | **float32** |  | 
**Allocated** | **float32** |  | 
**Utilization** | Pointer to **float32** |  | [optional] 

## Methods

### NewMetricsResourceDataV1

`func NewMetricsResourceDataV1(quota float32, allocated float32, ) *MetricsResourceDataV1`

NewMetricsResourceDataV1 instantiates a new MetricsResourceDataV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMetricsResourceDataV1WithDefaults

`func NewMetricsResourceDataV1WithDefaults() *MetricsResourceDataV1`

NewMetricsResourceDataV1WithDefaults instantiates a new MetricsResourceDataV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetQuota

`func (o *MetricsResourceDataV1) GetQuota() float32`

GetQuota returns the Quota field if non-nil, zero value otherwise.

### GetQuotaOk

`func (o *MetricsResourceDataV1) GetQuotaOk() (*float32, bool)`

GetQuotaOk returns a tuple with the Quota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuota

`func (o *MetricsResourceDataV1) SetQuota(v float32)`

SetQuota sets Quota field to given value.


### GetAllocated

`func (o *MetricsResourceDataV1) GetAllocated() float32`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *MetricsResourceDataV1) GetAllocatedOk() (*float32, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *MetricsResourceDataV1) SetAllocated(v float32)`

SetAllocated sets Allocated field to given value.


### GetUtilization

`func (o *MetricsResourceDataV1) GetUtilization() float32`

GetUtilization returns the Utilization field if non-nil, zero value otherwise.

### GetUtilizationOk

`func (o *MetricsResourceDataV1) GetUtilizationOk() (*float32, bool)`

GetUtilizationOk returns a tuple with the Utilization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUtilization

`func (o *MetricsResourceDataV1) SetUtilization(v float32)`

SetUtilization sets Utilization field to given value.

### HasUtilization

`func (o *MetricsResourceDataV1) HasUtilization() bool`

HasUtilization returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


