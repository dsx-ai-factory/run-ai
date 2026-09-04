# NodePoolQuotaStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolName** | Pointer to **string** | The node pool name which the current quota status refers to. | [optional] 
**Allocated** | Pointer to [**QuotaStatusResourceList**](QuotaStatusResourceList.md) |  | [optional] 
**AllocatedNonPreemptible** | Pointer to [**QuotaStatusResourceList**](QuotaStatusResourceList.md) |  | [optional] 
**Requested** | Pointer to [**QuotaStatusResourceList**](QuotaStatusResourceList.md) |  | [optional] 

## Methods

### NewNodePoolQuotaStatus

`func NewNodePoolQuotaStatus() *NodePoolQuotaStatus`

NewNodePoolQuotaStatus instantiates a new NodePoolQuotaStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolQuotaStatusWithDefaults

`func NewNodePoolQuotaStatusWithDefaults() *NodePoolQuotaStatus`

NewNodePoolQuotaStatusWithDefaults instantiates a new NodePoolQuotaStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolName

`func (o *NodePoolQuotaStatus) GetNodePoolName() string`

GetNodePoolName returns the NodePoolName field if non-nil, zero value otherwise.

### GetNodePoolNameOk

`func (o *NodePoolQuotaStatus) GetNodePoolNameOk() (*string, bool)`

GetNodePoolNameOk returns a tuple with the NodePoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolName

`func (o *NodePoolQuotaStatus) SetNodePoolName(v string)`

SetNodePoolName sets NodePoolName field to given value.

### HasNodePoolName

`func (o *NodePoolQuotaStatus) HasNodePoolName() bool`

HasNodePoolName returns a boolean if a field has been set.

### GetAllocated

`func (o *NodePoolQuotaStatus) GetAllocated() QuotaStatusResourceList`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *NodePoolQuotaStatus) GetAllocatedOk() (*QuotaStatusResourceList, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *NodePoolQuotaStatus) SetAllocated(v QuotaStatusResourceList)`

SetAllocated sets Allocated field to given value.

### HasAllocated

`func (o *NodePoolQuotaStatus) HasAllocated() bool`

HasAllocated returns a boolean if a field has been set.

### GetAllocatedNonPreemptible

`func (o *NodePoolQuotaStatus) GetAllocatedNonPreemptible() QuotaStatusResourceList`

GetAllocatedNonPreemptible returns the AllocatedNonPreemptible field if non-nil, zero value otherwise.

### GetAllocatedNonPreemptibleOk

`func (o *NodePoolQuotaStatus) GetAllocatedNonPreemptibleOk() (*QuotaStatusResourceList, bool)`

GetAllocatedNonPreemptibleOk returns a tuple with the AllocatedNonPreemptible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedNonPreemptible

`func (o *NodePoolQuotaStatus) SetAllocatedNonPreemptible(v QuotaStatusResourceList)`

SetAllocatedNonPreemptible sets AllocatedNonPreemptible field to given value.

### HasAllocatedNonPreemptible

`func (o *NodePoolQuotaStatus) HasAllocatedNonPreemptible() bool`

HasAllocatedNonPreemptible returns a boolean if a field has been set.

### GetRequested

`func (o *NodePoolQuotaStatus) GetRequested() QuotaStatusResourceList`

GetRequested returns the Requested field if non-nil, zero value otherwise.

### GetRequestedOk

`func (o *NodePoolQuotaStatus) GetRequestedOk() (*QuotaStatusResourceList, bool)`

GetRequestedOk returns a tuple with the Requested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequested

`func (o *NodePoolQuotaStatus) SetRequested(v QuotaStatusResourceList)`

SetRequested sets Requested field to given value.

### HasRequested

`func (o *NodePoolQuotaStatus) HasRequested() bool`

HasRequested returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


