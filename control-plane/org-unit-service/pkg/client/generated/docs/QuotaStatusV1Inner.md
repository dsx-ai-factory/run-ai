# QuotaStatusV1Inner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolName** | Pointer to **string** | The node pool name which the current quota status refers to. | [optional] 
**Allocated** | Pointer to [**QuotaStatusResourceListV1**](QuotaStatusResourceListV1.md) |  | [optional] 
**AllocatedNonPreemptible** | Pointer to [**QuotaStatusResourceListV1**](QuotaStatusResourceListV1.md) |  | [optional] 
**Requested** | Pointer to [**QuotaStatusResourceListV1**](QuotaStatusResourceListV1.md) |  | [optional] 

## Methods

### NewQuotaStatusV1Inner

`func NewQuotaStatusV1Inner() *QuotaStatusV1Inner`

NewQuotaStatusV1Inner instantiates a new QuotaStatusV1Inner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQuotaStatusV1InnerWithDefaults

`func NewQuotaStatusV1InnerWithDefaults() *QuotaStatusV1Inner`

NewQuotaStatusV1InnerWithDefaults instantiates a new QuotaStatusV1Inner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolName

`func (o *QuotaStatusV1Inner) GetNodePoolName() string`

GetNodePoolName returns the NodePoolName field if non-nil, zero value otherwise.

### GetNodePoolNameOk

`func (o *QuotaStatusV1Inner) GetNodePoolNameOk() (*string, bool)`

GetNodePoolNameOk returns a tuple with the NodePoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolName

`func (o *QuotaStatusV1Inner) SetNodePoolName(v string)`

SetNodePoolName sets NodePoolName field to given value.

### HasNodePoolName

`func (o *QuotaStatusV1Inner) HasNodePoolName() bool`

HasNodePoolName returns a boolean if a field has been set.

### GetAllocated

`func (o *QuotaStatusV1Inner) GetAllocated() QuotaStatusResourceListV1`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *QuotaStatusV1Inner) GetAllocatedOk() (*QuotaStatusResourceListV1, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *QuotaStatusV1Inner) SetAllocated(v QuotaStatusResourceListV1)`

SetAllocated sets Allocated field to given value.

### HasAllocated

`func (o *QuotaStatusV1Inner) HasAllocated() bool`

HasAllocated returns a boolean if a field has been set.

### GetAllocatedNonPreemptible

`func (o *QuotaStatusV1Inner) GetAllocatedNonPreemptible() QuotaStatusResourceListV1`

GetAllocatedNonPreemptible returns the AllocatedNonPreemptible field if non-nil, zero value otherwise.

### GetAllocatedNonPreemptibleOk

`func (o *QuotaStatusV1Inner) GetAllocatedNonPreemptibleOk() (*QuotaStatusResourceListV1, bool)`

GetAllocatedNonPreemptibleOk returns a tuple with the AllocatedNonPreemptible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedNonPreemptible

`func (o *QuotaStatusV1Inner) SetAllocatedNonPreemptible(v QuotaStatusResourceListV1)`

SetAllocatedNonPreemptible sets AllocatedNonPreemptible field to given value.

### HasAllocatedNonPreemptible

`func (o *QuotaStatusV1Inner) HasAllocatedNonPreemptible() bool`

HasAllocatedNonPreemptible returns a boolean if a field has been set.

### GetRequested

`func (o *QuotaStatusV1Inner) GetRequested() QuotaStatusResourceListV1`

GetRequested returns the Requested field if non-nil, zero value otherwise.

### GetRequestedOk

`func (o *QuotaStatusV1Inner) GetRequestedOk() (*QuotaStatusResourceListV1, bool)`

GetRequestedOk returns a tuple with the Requested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequested

`func (o *QuotaStatusV1Inner) SetRequested(v QuotaStatusResourceListV1)`

SetRequested sets Requested field to given value.

### HasRequested

`func (o *QuotaStatusV1Inner) HasRequested() bool`

HasRequested returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


