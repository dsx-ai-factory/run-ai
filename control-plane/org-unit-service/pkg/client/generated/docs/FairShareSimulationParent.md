# FairShareSimulationParent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Rank** | Pointer to **NullableString** | Highest … Lowest. Only while ranks is on. | [optional] 
**Gpu** | Pointer to [**NullableFairShareSimulationGpuSettings**](FairShareSimulationGpuSettings.md) |  | [optional] 

## Methods

### NewFairShareSimulationParent

`func NewFairShareSimulationParent(id string, ) *FairShareSimulationParent`

NewFairShareSimulationParent instantiates a new FairShareSimulationParent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationParentWithDefaults

`func NewFairShareSimulationParentWithDefaults() *FairShareSimulationParent`

NewFairShareSimulationParentWithDefaults instantiates a new FairShareSimulationParent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *FairShareSimulationParent) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *FairShareSimulationParent) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *FairShareSimulationParent) SetId(v string)`

SetId sets Id field to given value.


### GetRank

`func (o *FairShareSimulationParent) GetRank() string`

GetRank returns the Rank field if non-nil, zero value otherwise.

### GetRankOk

`func (o *FairShareSimulationParent) GetRankOk() (*string, bool)`

GetRankOk returns a tuple with the Rank field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRank

`func (o *FairShareSimulationParent) SetRank(v string)`

SetRank sets Rank field to given value.

### HasRank

`func (o *FairShareSimulationParent) HasRank() bool`

HasRank returns a boolean if a field has been set.

### SetRankNil

`func (o *FairShareSimulationParent) SetRankNil(b bool)`

 SetRankNil sets the value for Rank to be an explicit nil

### UnsetRank
`func (o *FairShareSimulationParent) UnsetRank()`

UnsetRank ensures that no value is present for Rank, not even an explicit nil
### GetGpu

`func (o *FairShareSimulationParent) GetGpu() FairShareSimulationGpuSettings`

GetGpu returns the Gpu field if non-nil, zero value otherwise.

### GetGpuOk

`func (o *FairShareSimulationParent) GetGpuOk() (*FairShareSimulationGpuSettings, bool)`

GetGpuOk returns a tuple with the Gpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGpu

`func (o *FairShareSimulationParent) SetGpu(v FairShareSimulationGpuSettings)`

SetGpu sets Gpu field to given value.

### HasGpu

`func (o *FairShareSimulationParent) HasGpu() bool`

HasGpu returns a boolean if a field has been set.

### SetGpuNil

`func (o *FairShareSimulationParent) SetGpuNil(b bool)`

 SetGpuNil sets the value for Gpu to be an explicit nil

### UnsetGpu
`func (o *FairShareSimulationParent) UnsetGpu()`

UnsetGpu ensures that no value is present for Gpu, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


