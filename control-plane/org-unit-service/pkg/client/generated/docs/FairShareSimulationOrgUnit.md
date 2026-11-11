# FairShareSimulationOrgUnit

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Type** | **string** |  | 
**Rank** | Pointer to **NullableString** | Replaces the stored rank. | [optional] 
**Gpu** | Pointer to [**NullableFairShareSimulationGpuSettings**](FairShareSimulationGpuSettings.md) |  | [optional] 
**Demand** | Pointer to [**NullableFairShareSimulationResources**](FairShareSimulationResources.md) | GPUs the org unit asks for. Without it, the most it can get; for a department, the sum of its projects&#39; demand. | [optional] 

## Methods

### NewFairShareSimulationOrgUnit

`func NewFairShareSimulationOrgUnit(id string, type_ string, ) *FairShareSimulationOrgUnit`

NewFairShareSimulationOrgUnit instantiates a new FairShareSimulationOrgUnit object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationOrgUnitWithDefaults

`func NewFairShareSimulationOrgUnitWithDefaults() *FairShareSimulationOrgUnit`

NewFairShareSimulationOrgUnitWithDefaults instantiates a new FairShareSimulationOrgUnit object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *FairShareSimulationOrgUnit) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *FairShareSimulationOrgUnit) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *FairShareSimulationOrgUnit) SetId(v string)`

SetId sets Id field to given value.


### GetType

`func (o *FairShareSimulationOrgUnit) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *FairShareSimulationOrgUnit) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *FairShareSimulationOrgUnit) SetType(v string)`

SetType sets Type field to given value.


### GetRank

`func (o *FairShareSimulationOrgUnit) GetRank() string`

GetRank returns the Rank field if non-nil, zero value otherwise.

### GetRankOk

`func (o *FairShareSimulationOrgUnit) GetRankOk() (*string, bool)`

GetRankOk returns a tuple with the Rank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRank

`func (o *FairShareSimulationOrgUnit) SetRank(v string)`

SetRank sets Rank field to given value.

### HasRank

`func (o *FairShareSimulationOrgUnit) HasRank() bool`

HasRank returns a boolean if a field has been set.

### SetRankNil

`func (o *FairShareSimulationOrgUnit) SetRankNil(b bool)`

 SetRankNil sets the value for Rank to be an explicit nil

### UnsetRank
`func (o *FairShareSimulationOrgUnit) UnsetRank()`

UnsetRank ensures that no value is present for Rank, not even an explicit nil
### GetGpu

`func (o *FairShareSimulationOrgUnit) GetGpu() FairShareSimulationGpuSettings`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *FairShareSimulationOrgUnit) GetGpuOk() (*FairShareSimulationGpuSettings, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *FairShareSimulationOrgUnit) SetGpu(v FairShareSimulationGpuSettings)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *FairShareSimulationOrgUnit) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *FairShareSimulationOrgUnit) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *FairShareSimulationOrgUnit) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil
### GetDemand

`func (o *FairShareSimulationOrgUnit) GetDemand() FairShareSimulationResources`

GetDemand returns the Demand field if non-nil, zero value otherwise.

### GetDemandOk

`func (o *FairShareSimulationOrgUnit) GetDemandOk() (*FairShareSimulationResources, bool)`

GetDemandOk returns a tuple with the Demand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDemand

`func (o *FairShareSimulationOrgUnit) SetDemand(v FairShareSimulationResources)`

SetDemand sets Demand field to given value.

### HasDemand

`func (o *FairShareSimulationOrgUnit) HasDemand() bool`

HasDemand returns a boolean if a field has been set.

### SetDemandNil

`func (o *FairShareSimulationOrgUnit) SetDemandNil(b bool)`

 SetDemandNil sets the value for Demand to be an explicit nil

### UnsetDemand
`func (o *FairShareSimulationOrgUnit) UnsetDemand()`

UnsetDemand ensures that no value is present for Demand, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


