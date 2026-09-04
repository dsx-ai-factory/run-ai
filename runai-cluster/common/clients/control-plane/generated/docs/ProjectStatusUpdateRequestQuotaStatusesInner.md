# ProjectStatusUpdateRequestQuotaStatusesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolName** | Pointer to **string** | The node pool name which the current quota status refers to. | [optional] 
**Allocated** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 
**AllocatedNonPreemptible** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 
**Requested** | Pointer to [**QuotaStatusResource**](QuotaStatusResource.md) |  | [optional] 

## Methods

### NewProjectStatusUpdateRequestQuotaStatusesInner

`func NewProjectStatusUpdateRequestQuotaStatusesInner() *ProjectStatusUpdateRequestQuotaStatusesInner`

NewProjectStatusUpdateRequestQuotaStatusesInner instantiates a new ProjectStatusUpdateRequestQuotaStatusesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectStatusUpdateRequestQuotaStatusesInnerWithDefaults

`func NewProjectStatusUpdateRequestQuotaStatusesInnerWithDefaults() *ProjectStatusUpdateRequestQuotaStatusesInner`

NewProjectStatusUpdateRequestQuotaStatusesInnerWithDefaults instantiates a new ProjectStatusUpdateRequestQuotaStatusesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolName

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) GetNodePoolName() string`

GetNodePoolName returns the NodePoolName field if non-nil, zero value otherwise.

### GetNodePoolNameOk

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) GetNodePoolNameOk() (*string, bool)`

GetNodePoolNameOk returns a tuple with the NodePoolName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolName

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) SetNodePoolName(v string)`

SetNodePoolName sets NodePoolName field to given value.

### HasNodePoolName

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) HasNodePoolName() bool`

HasNodePoolName returns a boolean if a field has been set.

### GetAllocated

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) GetAllocated() QuotaStatusResource`

GetAllocated returns the Allocated field if non-nil, zero value otherwise.

### GetAllocatedOk

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) GetAllocatedOk() (*QuotaStatusResource, bool)`

GetAllocatedOk returns a tuple with the Allocated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocated

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) SetAllocated(v QuotaStatusResource)`

SetAllocated sets Allocated field to given value.

### HasAllocated

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) HasAllocated() bool`

HasAllocated returns a boolean if a field has been set.

### GetAllocatedNonPreemptible

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) GetAllocatedNonPreemptible() QuotaStatusResource`

GetAllocatedNonPreemptible returns the AllocatedNonPreemptible field if non-nil, zero value otherwise.

### GetAllocatedNonPreemptibleOk

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) GetAllocatedNonPreemptibleOk() (*QuotaStatusResource, bool)`

GetAllocatedNonPreemptibleOk returns a tuple with the AllocatedNonPreemptible field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllocatedNonPreemptible

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) SetAllocatedNonPreemptible(v QuotaStatusResource)`

SetAllocatedNonPreemptible sets AllocatedNonPreemptible field to given value.

### HasAllocatedNonPreemptible

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) HasAllocatedNonPreemptible() bool`

HasAllocatedNonPreemptible returns a boolean if a field has been set.

### GetRequested

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) GetRequested() QuotaStatusResource`

GetRequested returns the Requested field if non-nil, zero value otherwise.

### GetRequestedOk

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) GetRequestedOk() (*QuotaStatusResource, bool)`

GetRequestedOk returns a tuple with the Requested field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequested

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) SetRequested(v QuotaStatusResource)`

SetRequested sets Requested field to given value.

### HasRequested

`func (o *ProjectStatusUpdateRequestQuotaStatusesInner) HasRequested() bool`

HasRequested returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


