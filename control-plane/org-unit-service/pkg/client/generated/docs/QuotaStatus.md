# QuotaStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Allocated** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 
**AllocatedNonPreemptible** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 
**Requested** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 

## Methods

### NewQuotaStatus

`func NewQuotaStatus() *QuotaStatus`

NewQuotaStatus instantiates a new QuotaStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQuotaStatusWithDefaults

`func NewQuotaStatusWithDefaults() *QuotaStatus`

NewQuotaStatusWithDefaults instantiates a new QuotaStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllocated

`func (o *QuotaStatus) GetAllocated() QuotaStatusResource`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *QuotaStatus) GetAllocatedOk() (*QuotaStatusResource, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *QuotaStatus) SetAllocated(v QuotaStatusResource)`

SetAllocated sets Allocated field to given value.

### HasAllocated

`func (o *QuotaStatus) HasAllocated() bool`

HasAllocated returns a boolean if a field has been set.

### GetAllocatedNonPreemptible

`func (o *QuotaStatus) GetAllocatedNonPreemptible() QuotaStatusResource`

GetAllocatedNonPreemptible returns the AllocatedNonPreemptible field if non-nil, zero value otherwise.

### GetAllocatedNonPreemptibleOk

`func (o *QuotaStatus) GetAllocatedNonPreemptibleOk() (*QuotaStatusResource, bool)`

GetAllocatedNonPreemptibleOk returns a tuple with the AllocatedNonPreemptible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedNonPreemptible

`func (o *QuotaStatus) SetAllocatedNonPreemptible(v QuotaStatusResource)`

SetAllocatedNonPreemptible sets AllocatedNonPreemptible field to given value.

### HasAllocatedNonPreemptible

`func (o *QuotaStatus) HasAllocatedNonPreemptible() bool`

HasAllocatedNonPreemptible returns a boolean if a field has been set.

### GetRequested

`func (o *QuotaStatus) GetRequested() QuotaStatusResource`

GetRequested returns the Requested field if non-nil, zero value otherwise.

### GetRequestedOk

`func (o *QuotaStatus) GetRequestedOk() (*QuotaStatusResource, bool)`

GetRequestedOk returns a tuple with the Requested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequested

`func (o *QuotaStatus) SetRequested(v QuotaStatusResource)`

SetRequested sets Requested field to given value.

### HasRequested

`func (o *QuotaStatus) HasRequested() bool`

HasRequested returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


