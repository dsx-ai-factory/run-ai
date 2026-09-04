# SchedulingCapabilitiesNullable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AdjustableWeights** | Pointer to **NullableBool** | Whether to allow adjustable weights for the nodepool&#39;s resources. | [optional] [default to false]
**Ranks** | Pointer to **NullableBool** | Whether to allow ranks for the nodepool&#39;s resources. | [optional] [default to true]
**Quota** | Pointer to **NullableBool** | Whether to allow GPU quota for the nodepool&#39;s resources. | [optional] [default to true]
**CpuQuota** | Pointer to **NullableBool** | Whether to allow CPU quota for the nodepool&#39;s resources. Only allowed when allowQuota&#x3D;true. | [optional] [default to false]

## Methods

### NewSchedulingCapabilitiesNullable

`func NewSchedulingCapabilitiesNullable() *SchedulingCapabilitiesNullable`

NewSchedulingCapabilitiesNullable instantiates a new SchedulingCapabilitiesNullable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSchedulingCapabilitiesNullableWithDefaults

`func NewSchedulingCapabilitiesNullableWithDefaults() *SchedulingCapabilitiesNullable`

NewSchedulingCapabilitiesNullableWithDefaults instantiates a new SchedulingCapabilitiesNullable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAdjustableWeights

`func (o *SchedulingCapabilitiesNullable) GetAdjustableWeights() bool`

GetAdjustableWeights returns the AdjustableWeights field if non-nil, zero value otherwise.

### GetAdjustableWeightsOk

`func (o *SchedulingCapabilitiesNullable) GetAdjustableWeightsOk() (*bool, bool)`

GetAdjustableWeightsOk returns a tuple with the AdjustableWeights field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdjustableWeights

`func (o *SchedulingCapabilitiesNullable) SetAdjustableWeights(v bool)`

SetAdjustableWeights sets AdjustableWeights field to given value.

### HasAdjustableWeights

`func (o *SchedulingCapabilitiesNullable) HasAdjustableWeights() bool`

HasAdjustableWeights returns a boolean if a field has been set.

### SetAdjustableWeightsNil

`func (o *SchedulingCapabilitiesNullable) SetAdjustableWeightsNil(b bool)`

 SetAdjustableWeightsNil sets the value for AdjustableWeights to be an explicit nil

### UnsetAdjustableWeights
`func (o *SchedulingCapabilitiesNullable) UnsetAdjustableWeights()`

UnsetAdjustableWeights ensures that no value is present for AdjustableWeights, not even an explicit nil
### GetRanks

`func (o *SchedulingCapabilitiesNullable) GetRanks() bool`

GetRanks returns the Ranks field if non-nil, zero value otherwise.

### GetRanksOk

`func (o *SchedulingCapabilitiesNullable) GetRanksOk() (*bool, bool)`

GetRanksOk returns a tuple with the Ranks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRanks

`func (o *SchedulingCapabilitiesNullable) SetRanks(v bool)`

SetRanks sets Ranks field to given value.

### HasRanks

`func (o *SchedulingCapabilitiesNullable) HasRanks() bool`

HasRanks returns a boolean if a field has been set.

### SetRanksNil

`func (o *SchedulingCapabilitiesNullable) SetRanksNil(b bool)`

 SetRanksNil sets the value for Ranks to be an explicit nil

### UnsetRanks
`func (o *SchedulingCapabilitiesNullable) UnsetRanks()`

UnsetRanks ensures that no value is present for Ranks, not even an explicit nil
### GetQuota

`func (o *SchedulingCapabilitiesNullable) GetQuota() bool`

GetQuota returns the Quota field if non-nil, zero value otherwise.

### GetQuotaOk

`func (o *SchedulingCapabilitiesNullable) GetQuotaOk() (*bool, bool)`

GetQuotaOk returns a tuple with the Quota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuota

`func (o *SchedulingCapabilitiesNullable) SetQuota(v bool)`

SetQuota sets Quota field to given value.

### HasQuota

`func (o *SchedulingCapabilitiesNullable) HasQuota() bool`

HasQuota returns a boolean if a field has been set.

### SetQuotaNil

`func (o *SchedulingCapabilitiesNullable) SetQuotaNil(b bool)`

 SetQuotaNil sets the value for Quota to be an explicit nil

### UnsetQuota
`func (o *SchedulingCapabilitiesNullable) UnsetQuota()`

UnsetQuota ensures that no value is present for Quota, not even an explicit nil
### GetCpuQuota

`func (o *SchedulingCapabilitiesNullable) GetCpuQuota() bool`

GetCpuQuota returns the CpuQuota field if non-nil, zero value otherwise.

### GetCpuQuotaOk

`func (o *SchedulingCapabilitiesNullable) GetCpuQuotaOk() (*bool, bool)`

GetCpuQuotaOk returns a tuple with the CpuQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuQuota

`func (o *SchedulingCapabilitiesNullable) SetCpuQuota(v bool)`

SetCpuQuota sets CpuQuota field to given value.

### HasCpuQuota

`func (o *SchedulingCapabilitiesNullable) HasCpuQuota() bool`

HasCpuQuota returns a boolean if a field has been set.

### SetCpuQuotaNil

`func (o *SchedulingCapabilitiesNullable) SetCpuQuotaNil(b bool)`

 SetCpuQuotaNil sets the value for CpuQuota to be an explicit nil

### UnsetCpuQuota
`func (o *SchedulingCapabilitiesNullable) UnsetCpuQuota()`

UnsetCpuQuota ensures that no value is present for CpuQuota, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


