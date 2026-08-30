# QuotaStatusNodePool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Allocated** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 
**AllocatedNonPreemptible** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 
**Requested** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 
**NodePoolName** | **string** | The node pool name which the current quota status refers to | 
**NodePoolId** | **string** | The node pool id which the current quota status refers to | 

## Methods

### NewQuotaStatusNodePool

`func NewQuotaStatusNodePool(nodePoolName string, nodePoolId string, ) *QuotaStatusNodePool`

NewQuotaStatusNodePool instantiates a new QuotaStatusNodePool object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQuotaStatusNodePoolWithDefaults

`func NewQuotaStatusNodePoolWithDefaults() *QuotaStatusNodePool`

NewQuotaStatusNodePoolWithDefaults instantiates a new QuotaStatusNodePool object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllocated

`func (o *QuotaStatusNodePool) GetAllocated() QuotaStatusResource`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *QuotaStatusNodePool) GetAllocatedOk() (*QuotaStatusResource, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *QuotaStatusNodePool) SetAllocated(v QuotaStatusResource)`

SetAllocated sets Allocated field to given value.

### HasAllocated

`func (o *QuotaStatusNodePool) HasAllocated() bool`

HasAllocated returns a boolean if a field has been set.

### GetAllocatedNonPreemptible

`func (o *QuotaStatusNodePool) GetAllocatedNonPreemptible() QuotaStatusResource`

GetAllocatedNonPreemptible returns the AllocatedNonPreemptible field if non-nil, zero value otherwise.

### GetAllocatedNonPreemptibleOk

`func (o *QuotaStatusNodePool) GetAllocatedNonPreemptibleOk() (*QuotaStatusResource, bool)`

GetAllocatedNonPreemptibleOk returns a tuple with the AllocatedNonPreemptible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedNonPreemptible

`func (o *QuotaStatusNodePool) SetAllocatedNonPreemptible(v QuotaStatusResource)`

SetAllocatedNonPreemptible sets AllocatedNonPreemptible field to given value.

### HasAllocatedNonPreemptible

`func (o *QuotaStatusNodePool) HasAllocatedNonPreemptible() bool`

HasAllocatedNonPreemptible returns a boolean if a field has been set.

### GetRequested

`func (o *QuotaStatusNodePool) GetRequested() QuotaStatusResource`

GetRequested returns the Requested field if non-nil, zero value otherwise.

### GetRequestedOk

`func (o *QuotaStatusNodePool) GetRequestedOk() (*QuotaStatusResource, bool)`

GetRequestedOk returns a tuple with the Requested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequested

`func (o *QuotaStatusNodePool) SetRequested(v QuotaStatusResource)`

SetRequested sets Requested field to given value.

### HasRequested

`func (o *QuotaStatusNodePool) HasRequested() bool`

HasRequested returns a boolean if a field has been set.

### GetNodePoolName

`func (o *QuotaStatusNodePool) GetNodePoolName() string`

GetNodePoolName returns the NodePoolName field if non-nil, zero value otherwise.

### GetNodePoolNameOk

`func (o *QuotaStatusNodePool) GetNodePoolNameOk() (*string, bool)`

GetNodePoolNameOk returns a tuple with the NodePoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolName

`func (o *QuotaStatusNodePool) SetNodePoolName(v string)`

SetNodePoolName sets NodePoolName field to given value.


### GetNodePoolId

`func (o *QuotaStatusNodePool) GetNodePoolId() string`

GetNodePoolId returns the NodePoolId field if non-nil, zero value otherwise.

### GetNodePoolIdOk

`func (o *QuotaStatusNodePool) GetNodePoolIdOk() (*string, bool)`

GetNodePoolIdOk returns a tuple with the NodePoolId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolId

`func (o *QuotaStatusNodePool) SetNodePoolId(v string)`

SetNodePoolId sets NodePoolId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


