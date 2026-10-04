# ResourcesPerNodePoolGpu

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | Pointer to **NullableFloat64** | The resource quota assigned to this org-unit. | [optional] 
**Limit** | Pointer to **NullableFloat64** | The maximum resource allocation for this org-unit. Set to null for unlimited. | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat64** | The weight assigned to determine the org-unit&#39;s share of over-quota resources within its rank. When multiple org-units share the same rank in a node pool, resources are distributed proportionally based on their weights. Valid values: 1–20. Requires over-quota weight to be enabled. When enabled, the default is 2. See [GPU over quota](/run-ai/self-hosted/documentation/platform-management/manage-ai-initiatives/organization/projects) for more details.  | [optional] 

## Methods

### NewResourcesPerNodePoolGpu

`func NewResourcesPerNodePoolGpu() *ResourcesPerNodePoolGpu`

NewResourcesPerNodePoolGpu instantiates a new ResourcesPerNodePoolGpu object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesPerNodePoolGpuWithDefaults

`func NewResourcesPerNodePoolGpuWithDefaults() *ResourcesPerNodePoolGpu`

NewResourcesPerNodePoolGpuWithDefaults instantiates a new ResourcesPerNodePoolGpu object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *ResourcesPerNodePoolGpu) GetDeserved() float64`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *ResourcesPerNodePoolGpu) GetDeservedOk() (*float64, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *ResourcesPerNodePoolGpu) SetDeserved(v float64)`

SetDeserved sets Deserved field to given value.

### HasDeserved

`func (o *ResourcesPerNodePoolGpu) HasDeserved() bool`

HasDeserved returns a boolean if a field has been set.

### SetDeservedNil

`func (o *ResourcesPerNodePoolGpu) SetDeservedNil(b bool)`

 SetDeservedNil sets the value for Deserved to be an explicit nil

### UnsetDeserved
`func (o *ResourcesPerNodePoolGpu) UnsetDeserved()`

UnsetDeserved ensures that no value is present for Deserved, not even an explicit nil
### GetLimit

`func (o *ResourcesPerNodePoolGpu) GetLimit() float64`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *ResourcesPerNodePoolGpu) GetLimitOk() (*float64, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *ResourcesPerNodePoolGpu) SetLimit(v float64)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *ResourcesPerNodePoolGpu) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### SetLimitNil

`func (o *ResourcesPerNodePoolGpu) SetLimitNil(b bool)`

 SetLimitNil sets the value for Limit to be an explicit nil

### UnsetLimit
`func (o *ResourcesPerNodePoolGpu) UnsetLimit()`

UnsetLimit ensures that no value is present for Limit, not even an explicit nil
### GetOverQuotaWeight

`func (o *ResourcesPerNodePoolGpu) GetOverQuotaWeight() float64`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *ResourcesPerNodePoolGpu) GetOverQuotaWeightOk() (*float64, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *ResourcesPerNodePoolGpu) SetOverQuotaWeight(v float64)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *ResourcesPerNodePoolGpu) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *ResourcesPerNodePoolGpu) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *ResourcesPerNodePoolGpu) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


