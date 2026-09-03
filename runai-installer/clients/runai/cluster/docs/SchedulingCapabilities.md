# SchedulingCapabilities

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AdjustableWeights** | **bool** | Whether to allow adjustable weights for the nodepool&#39;s resources. | [default to false]
**Ranks** | **bool** | Whether to allow ranks for the nodepool&#39;s resources. | [default to true]
**Quota** | **bool** | Whether to allow GPU quota for the nodepool&#39;s resources. | [default to true]
**CpuQuota** | **bool** | Whether to allow CPU quota for the nodepool&#39;s resources. Only allowed when allowQuota&#x3D;true. | [default to false]

## Methods

### NewSchedulingCapabilities

`func NewSchedulingCapabilities(adjustableWeights bool, ranks bool, quota bool, cpuQuota bool, ) *SchedulingCapabilities`

NewSchedulingCapabilities instantiates a new SchedulingCapabilities object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSchedulingCapabilitiesWithDefaults

`func NewSchedulingCapabilitiesWithDefaults() *SchedulingCapabilities`

NewSchedulingCapabilitiesWithDefaults instantiates a new SchedulingCapabilities object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAdjustableWeights

`func (o *SchedulingCapabilities) GetAdjustableWeights() bool`

GetAdjustableWeights returns the AdjustableWeights field if non-nil, zero value otherwise.

### GetAdjustableWeightsOk

`func (o *SchedulingCapabilities) GetAdjustableWeightsOk() (*bool, bool)`

GetAdjustableWeightsOk returns a tuple with the AdjustableWeights field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdjustableWeights

`func (o *SchedulingCapabilities) SetAdjustableWeights(v bool)`

SetAdjustableWeights sets AdjustableWeights field to given value.


### GetRanks

`func (o *SchedulingCapabilities) GetRanks() bool`

GetRanks returns the Ranks field if non-nil, zero value otherwise.

### GetRanksOk

`func (o *SchedulingCapabilities) GetRanksOk() (*bool, bool)`

GetRanksOk returns a tuple with the Ranks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRanks

`func (o *SchedulingCapabilities) SetRanks(v bool)`

SetRanks sets Ranks field to given value.


### GetQuota

`func (o *SchedulingCapabilities) GetQuota() bool`

GetQuota returns the Quota field if non-nil, zero value otherwise.

### GetQuotaOk

`func (o *SchedulingCapabilities) GetQuotaOk() (*bool, bool)`

GetQuotaOk returns a tuple with the Quota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuota

`func (o *SchedulingCapabilities) SetQuota(v bool)`

SetQuota sets Quota field to given value.


### GetCpuQuota

`func (o *SchedulingCapabilities) GetCpuQuota() bool`

GetCpuQuota returns the CpuQuota field if non-nil, zero value otherwise.

### GetCpuQuotaOk

`func (o *SchedulingCapabilities) GetCpuQuotaOk() (*bool, bool)`

GetCpuQuotaOk returns a tuple with the CpuQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuQuota

`func (o *SchedulingCapabilities) SetCpuQuota(v bool)`

SetCpuQuota sets CpuQuota field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


