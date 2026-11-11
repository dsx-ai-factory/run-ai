# FairShareSimulationResult

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Type** | **string** |  | 
**ParentId** | Pointer to **NullableString** |  | [optional] 
**FairShare** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | Resources the scheduler would give the org unit when the node pool is fully busy, the sum of guaranteed and overQuota. | 
**Guaranteed** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | The part of the fair share that comes from the org unit&#39;s quota. | 
**OverQuota** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | The part of the fair share above the org unit&#39;s quota. | 
**Unmet** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | Demand the fair share does not cover. Without a demand, measured against the org unit&#39;s limit, else its parent&#39;s limit, else the node pool&#39;s capacity. | 
**WeightShareInRank** | **float64** | The org unit&#39;s over-quota weight divided by the total weight of its rank among its siblings, including the ones the caller can&#39;t read. With ranks off, among all its siblings. | 

## Methods

### NewFairShareSimulationResult

`func NewFairShareSimulationResult(id string, type_ string, fairShare FairShareSimulationResources, guaranteed FairShareSimulationResources, overQuota FairShareSimulationResources, unmet FairShareSimulationResources, weightShareInRank float64, ) *FairShareSimulationResult`

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


### GetType

`func (o *FairShareSimulationResult) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *FairShareSimulationResult) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *FairShareSimulationResult) SetType(v string)`

SetType sets Type field to given value.


### GetParentId

`func (o *FairShareSimulationResult) GetParentId() string`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *FairShareSimulationResult) GetParentIdOk() (*string, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *FairShareSimulationResult) SetParentId(v string)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *FairShareSimulationResult) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### SetParentIdNil

`func (o *FairShareSimulationResult) SetParentIdNil(b bool)`

 SetParentIdNil sets the value for ParentId to be an explicit nil

### UnsetParentId
`func (o *FairShareSimulationResult) UnsetParentId()`

UnsetParentId ensures that no value is present for ParentId, not even an explicit nil
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


