# FairShareSimulationResult

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Demand** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | The GPUs the org unit asked for in this scenario, after the default and the limit were applied. | 
**FairShare** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | GPUs the scheduler would give the org unit, the sum of guaranteed and overQuota. | 
**Guaranteed** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | The part of the fair share within the org unit&#39;s quota. | 
**OverQuota** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | The part of the fair share above the org unit&#39;s quota. | 
**Unmet** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | demand minus fairShare. | 
**WeightShareInRank** | **float64** | The org unit&#39;s over-quota weight divided by the total weight of its rank among its siblings, including ones the caller can&#39;t read. With ranks off, among all siblings. | 

## Methods

### NewFairShareSimulationResult

`func NewFairShareSimulationResult(id string, demand FairShareSimulationResources, fairShare FairShareSimulationResources, guaranteed FairShareSimulationResources, overQuota FairShareSimulationResources, unmet FairShareSimulationResources, weightShareInRank float64, ) *FairShareSimulationResult`

NewFairShareSimulationResult instantiates a new FairShareSimulationResult object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationResultWithDefaults

`func NewFairShareSimulationResultWithDefaults() *FairShareSimulationResult`

NewFairShareSimulationResultWithDefaults instantiates a new FairShareSimulationResult object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *FairShareSimulationResult) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *FairShareSimulationResult) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *FairShareSimulationResult) SetId(v string)`

SetId sets Id field to given value.


### GetDemand

`func (o *FairShareSimulationResult) GetDemand() FairShareSimulationResources`

GetDemand returns the Demand field if non-nil, zero value otherwise.

### GetDemandOk

`func (o *FairShareSimulationResult) GetDemandOk() (*FairShareSimulationResources, bool)`

GetDemandOk returns a tuple with the Demand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDemand

`func (o *FairShareSimulationResult) SetDemand(v FairShareSimulationResources)`

SetDemand sets Demand field to given value.


### GetFairShare

`func (o *FairShareSimulationResult) GetFairShare() FairShareSimulationResources`

GetFairShare returns the FairShare field if non-nil, zero value otherwise.

### GetFairShareOk

`func (o *FairShareSimulationResult) GetFairShareOk() (*FairShareSimulationResources, bool)`

GetFairShareOk returns a tuple with the FairShare field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFairShare

`func (o *FairShareSimulationResult) SetFairShare(v FairShareSimulationResources)`

SetFairShare sets FairShare field to given value.


### GetGuaranteed

`func (o *FairShareSimulationResult) GetGuaranteed() FairShareSimulationResources`

GetGuaranteed returns the Guaranteed field if non-nil, zero value otherwise.

### GetGuaranteedOk

`func (o *FairShareSimulationResult) GetGuaranteedOk() (*FairShareSimulationResources, bool)`

GetGuaranteedOk returns a tuple with the Guaranteed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuaranteed

`func (o *FairShareSimulationResult) SetGuaranteed(v FairShareSimulationResources)`

SetGuaranteed sets Guaranteed field to given value.


### GetOverQuota

`func (o *FairShareSimulationResult) GetOverQuota() FairShareSimulationResources`

GetOverQuota returns the OverQuota field if non-nil, zero value otherwise.

### GetOverQuotaOk

`func (o *FairShareSimulationResult) GetOverQuotaOk() (*FairShareSimulationResources, bool)`

GetOverQuotaOk returns a tuple with the OverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuota

`func (o *FairShareSimulationResult) SetOverQuota(v FairShareSimulationResources)`

SetOverQuota sets OverQuota field to given value.


### GetUnmet

`func (o *FairShareSimulationResult) GetUnmet() FairShareSimulationResources`

GetUnmet returns the Unmet field if non-nil, zero value otherwise.

### GetUnmetOk

`func (o *FairShareSimulationResult) GetUnmetOk() (*FairShareSimulationResources, bool)`

GetUnmetOk returns a tuple with the Unmet field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnmet

`func (o *FairShareSimulationResult) SetUnmet(v FairShareSimulationResources)`

SetUnmet sets Unmet field to given value.


### GetWeightShareInRank

`func (o *FairShareSimulationResult) GetWeightShareInRank() float64`

GetWeightShareInRank returns the WeightShareInRank field if non-nil, zero value otherwise.

### GetWeightShareInRankOk

`func (o *FairShareSimulationResult) GetWeightShareInRankOk() (*float64, bool)`

GetWeightShareInRankOk returns a tuple with the WeightShareInRank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWeightShareInRank

`func (o *FairShareSimulationResult) SetWeightShareInRank(v float64)`

SetWeightShareInRank sets WeightShareInRank field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


