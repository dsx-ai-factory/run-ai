# FairShareSimulationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Capacity** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | The GPUs the simulation split. | 
**Parent** | Pointer to [**NullableFairShareSimulationResult**](FairShareSimulationResult.md) | When the requested org units are projects: their department&#39;s own result. Null when the caller can&#39;t read the department.  | [optional] 
**OrgUnits** | [**[]FairShareSimulationResult**](FairShareSimulationResult.md) | One result per requested org unit, in the request&#39;s order. | 

## Methods

### NewFairShareSimulationResponse

`func NewFairShareSimulationResponse(capacity FairShareSimulationResources, orgUnits []FairShareSimulationResult, ) *FairShareSimulationResponse`

NewFairShareSimulationResponse instantiates a new FairShareSimulationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationResponseWithDefaults

`func NewFairShareSimulationResponseWithDefaults() *FairShareSimulationResponse`

NewFairShareSimulationResponseWithDefaults instantiates a new FairShareSimulationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCapacity

`func (o *FairShareSimulationResponse) GetCapacity() FairShareSimulationResources`

GetCapacity returns the Capacity field if non-nil, zero value otherwise.

### GetCapacityOk

`func (o *FairShareSimulationResponse) GetCapacityOk() (*FairShareSimulationResources, bool)`

GetCapacityOk returns a tuple with the Capacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapacity

`func (o *FairShareSimulationResponse) SetCapacity(v FairShareSimulationResources)`

SetCapacity sets Capacity field to given value.


### GetParent

`func (o *FairShareSimulationResponse) GetParent() FairShareSimulationResult`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *FairShareSimulationResponse) GetParentOk() (*FairShareSimulationResult, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *FairShareSimulationResponse) SetParent(v FairShareSimulationResult)`

SetParent sets Parent field to given value.

### HasParent

`func (o *FairShareSimulationResponse) HasParent() bool`

HasParent returns a boolean if a field has been set.

### SetParentNil

`func (o *FairShareSimulationResponse) SetParentNil(b bool)`

 SetParentNil sets the value for Parent to be an explicit nil

### UnsetParent
`func (o *FairShareSimulationResponse) UnsetParent()`

UnsetParent ensures that no value is present for Parent, not even an explicit nil
### GetOrgUnits

`func (o *FairShareSimulationResponse) GetOrgUnits() []FairShareSimulationResult`

GetOrgUnits returns the OrgUnits field if non-nil, zero value otherwise.

### GetOrgUnitsOk

`func (o *FairShareSimulationResponse) GetOrgUnitsOk() (*[]FairShareSimulationResult, bool)`

GetOrgUnitsOk returns a tuple with the OrgUnits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnits

`func (o *FairShareSimulationResponse) SetOrgUnits(v []FairShareSimulationResult)`

SetOrgUnits sets OrgUnits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


