# FairShareSimulationGpuSettings

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | Pointer to **NullableFloat64** | The GPU quota; -1 means unlimited. | [optional] 
**Limit** | Pointer to **NullableFloat64** | The maximum GPU allocation; -1 means unlimited. Only while limits is on. | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat64** |  | [optional] 

## Methods

### NewFairShareSimulationGpuSettings

`func NewFairShareSimulationGpuSettings() *FairShareSimulationGpuSettings`

NewFairShareSimulationGpuSettings instantiates a new FairShareSimulationGpuSettings object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationGpuSettingsWithDefaults

`func NewFairShareSimulationGpuSettingsWithDefaults() *FairShareSimulationGpuSettings`

NewFairShareSimulationGpuSettingsWithDefaults instantiates a new FairShareSimulationGpuSettings object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *FairShareSimulationGpuSettings) GetDeserved() float64`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *FairShareSimulationGpuSettings) GetDeservedOk() (*float64, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *FairShareSimulationGpuSettings) SetDeserved(v float64)`

SetDeserved sets Deserved field to given value.

### HasDeserved

`func (o *FairShareSimulationGpuSettings) HasDeserved() bool`

HasDeserved returns a boolean if a field has been set.

### SetDeservedNil

`func (o *FairShareSimulationGpuSettings) SetDeservedNil(b bool)`

 SetDeservedNil sets the value for Deserved to be an explicit nil

### UnsetDeserved
`func (o *FairShareSimulationGpuSettings) UnsetDeserved()`

UnsetDeserved ensures that no value is present for Deserved, not even an explicit nil
### GetLimit

`func (o *FairShareSimulationGpuSettings) GetLimit() float64`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *FairShareSimulationGpuSettings) GetLimitOk() (*float64, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *FairShareSimulationGpuSettings) SetLimit(v float64)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *FairShareSimulationGpuSettings) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### SetLimitNil

`func (o *FairShareSimulationGpuSettings) SetLimitNil(b bool)`

 SetLimitNil sets the value for Limit to be an explicit nil

### UnsetLimit
`func (o *FairShareSimulationGpuSettings) UnsetLimit()`

UnsetLimit ensures that no value is present for Limit, not even an explicit nil
### GetOverQuotaWeight

`func (o *FairShareSimulationGpuSettings) GetOverQuotaWeight() float64`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *FairShareSimulationGpuSettings) GetOverQuotaWeightOk() (*float64, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *FairShareSimulationGpuSettings) SetOverQuotaWeight(v float64)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *FairShareSimulationGpuSettings) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *FairShareSimulationGpuSettings) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *FairShareSimulationGpuSettings) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


