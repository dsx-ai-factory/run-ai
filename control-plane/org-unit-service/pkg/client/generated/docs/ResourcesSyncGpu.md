# ResourcesSyncGpu

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | **float64** | The GPU quota assigned to this org-unit. | [default to 0]
**Limit** | Pointer to **NullableFloat64** | The maximum GPU device allocation the org-unit can get from this node pool, representing the maximum sum of assigned quota and over-quota GPUs. Set to null for unlimited. | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat64** | The weight assigned to determine the org-unit&#39;s share of over-quota resources within its rank. When multiple org-units share the same rank in a node pool, resources are distributed proportionally based on their weights. Valid values: 1–20. Requires over-quota weight to be enabled. When enabled, the default is 2. See [GPU over quota](/run-ai/self-hosted/documentation/platform-management/manage-ai-initiatives/organization/projects) for more details.  | [optional] 

## Methods

### NewResourcesSyncGpu

`func NewResourcesSyncGpu(deserved float64, ) *ResourcesSyncGpu`

NewResourcesSyncGpu instantiates a new ResourcesSyncGpu object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesSyncGpuWithDefaults

`func NewResourcesSyncGpuWithDefaults() *ResourcesSyncGpu`

NewResourcesSyncGpuWithDefaults instantiates a new ResourcesSyncGpu object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *ResourcesSyncGpu) GetDeserved() float64`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *ResourcesSyncGpu) GetDeservedOk() (*float64, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *ResourcesSyncGpu) SetDeserved(v float64)`

SetDeserved sets Deserved field to given value.


### GetLimit

`func (o *ResourcesSyncGpu) GetLimit() float64`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *ResourcesSyncGpu) GetLimitOk() (*float64, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *ResourcesSyncGpu) SetLimit(v float64)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *ResourcesSyncGpu) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### SetLimitNil

`func (o *ResourcesSyncGpu) SetLimitNil(b bool)`

 SetLimitNil sets the value for Limit to be an explicit nil

### UnsetLimit
`func (o *ResourcesSyncGpu) UnsetLimit()`

UnsetLimit ensures that no value is present for Limit, not even an explicit nil
### GetOverQuotaWeight

`func (o *ResourcesSyncGpu) GetOverQuotaWeight() float64`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *ResourcesSyncGpu) GetOverQuotaWeightOk() (*float64, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *ResourcesSyncGpu) SetOverQuotaWeight(v float64)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *ResourcesSyncGpu) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *ResourcesSyncGpu) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *ResourcesSyncGpu) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


