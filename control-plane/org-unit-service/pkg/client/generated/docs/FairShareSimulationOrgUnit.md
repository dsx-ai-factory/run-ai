# FairShareSimulationOrgUnit

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**Type** | **string** |  | 
**Rank** | Pointer to **string** | Sent together with resources. Replaces the stored rank. | [optional] 
**Resources** | Pointer to [**NullableResourcesPerNodePool**](ResourcesPerNodePool.md) | Sent together with rank. Replaces the stored resources; -1 means unlimited. | [optional] 
**Demand** | Pointer to [**NullableQuotaStatusResource**](QuotaStatusResource.md) | Resources a project asks for. Without it the project asks for the most it can get. Not allowed on departments, which ask for the sum of their projects&#39; demand. | [optional] 

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

### GetResources

`func (o *FairShareSimulationOrgUnit) GetResources() ResourcesPerNodePool`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *FairShareSimulationOrgUnit) GetResourcesOk() (*ResourcesPerNodePool, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *FairShareSimulationOrgUnit) SetResources(v ResourcesPerNodePool)`

SetResources sets Resources field to given value.

### HasResources

`func (o *FairShareSimulationOrgUnit) HasResources() bool`

HasResources returns a boolean if a field has been set.

### SetResourcesNil

`func (o *FairShareSimulationOrgUnit) SetResourcesNil(b bool)`

 SetResourcesNil sets the value for Resources to be an explicit nil

### UnsetResources
`func (o *FairShareSimulationOrgUnit) UnsetResources()`

UnsetResources ensures that no value is present for Resources, not even an explicit nil
### GetDemand

`func (o *FairShareSimulationOrgUnit) GetDemand() QuotaStatusResource`

GetDemand returns the Demand field if non-nil, zero value otherwise.

### GetDemandOk

`func (o *FairShareSimulationOrgUnit) GetDemandOk() (*QuotaStatusResource, bool)`

GetDemandOk returns a tuple with the Demand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDemand

`func (o *FairShareSimulationOrgUnit) SetDemand(v QuotaStatusResource)`

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


