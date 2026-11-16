# FairShareSimulationSchedulingCapabilities

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Quota** | **bool** |  | 
**Ranks** | **bool** |  | 
**AdjustableWeights** | **bool** |  | 
**Limits** | **bool** | Whether GPU limits apply. Only allowed when quota is true. | 

## Methods

### NewFairShareSimulationSchedulingCapabilities

`func NewFairShareSimulationSchedulingCapabilities(quota bool, ranks bool, adjustableWeights bool, limits bool, ) *FairShareSimulationSchedulingCapabilities`

NewFairShareSimulationSchedulingCapabilities instantiates a new FairShareSimulationSchedulingCapabilities object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationSchedulingCapabilitiesWithDefaults

`func NewFairShareSimulationSchedulingCapabilitiesWithDefaults() *FairShareSimulationSchedulingCapabilities`

NewFairShareSimulationSchedulingCapabilitiesWithDefaults instantiates a new FairShareSimulationSchedulingCapabilities object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetQuota

`func (o *FairShareSimulationSchedulingCapabilities) GetQuota() bool`

GetQuota returns the Quota field if non-nil, zero value otherwise.

### GetQuotaOk

`func (o *FairShareSimulationSchedulingCapabilities) GetQuotaOk() (*bool, bool)`

GetQuotaOk returns a tuple with the Quota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuota

`func (o *FairShareSimulationSchedulingCapabilities) SetQuota(v bool)`

SetQuota sets Quota field to given value.


### GetRanks

`func (o *FairShareSimulationSchedulingCapabilities) GetRanks() bool`

GetRanks returns the Ranks field if non-nil, zero value otherwise.

### GetRanksOk

`func (o *FairShareSimulationSchedulingCapabilities) GetRanksOk() (*bool, bool)`

GetRanksOk returns a tuple with the Ranks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRanks

`func (o *FairShareSimulationSchedulingCapabilities) SetRanks(v bool)`

SetRanks sets Ranks field to given value.


### GetAdjustableWeights

`func (o *FairShareSimulationSchedulingCapabilities) GetAdjustableWeights() bool`

GetAdjustableWeights returns the AdjustableWeights field if non-nil, zero value otherwise.

### GetAdjustableWeightsOk

`func (o *FairShareSimulationSchedulingCapabilities) GetAdjustableWeightsOk() (*bool, bool)`

GetAdjustableWeightsOk returns a tuple with the AdjustableWeights field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdjustableWeights

`func (o *FairShareSimulationSchedulingCapabilities) SetAdjustableWeights(v bool)`

SetAdjustableWeights sets AdjustableWeights field to given value.


### GetLimits

`func (o *FairShareSimulationSchedulingCapabilities) GetLimits() bool`

GetLimits returns the Limits field if non-nil, zero value otherwise.

### GetLimitsOk

`func (o *FairShareSimulationSchedulingCapabilities) GetLimitsOk() (*bool, bool)`

GetLimitsOk returns a tuple with the Limits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimits

`func (o *FairShareSimulationSchedulingCapabilities) SetLimits(v bool)`

SetLimits sets Limits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


