# FairShareSimulationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | [**FairShareSimulationNodePool**](FairShareSimulationNodePool.md) |  | 
**Assumptions** | [**FairShareSimulationAssumptions**](FairShareSimulationAssumptions.md) |  | 
**OrgUnitType** | **string** | The type of the requested org units. One type per request. | 
**OrgUnits** | [**[]FairShareSimulationOrgUnit**](FairShareSimulationOrgUnit.md) | The org units to return, with their unsaved changes. Send the rows the screen shows. | 

## Methods

### NewFairShareSimulationRequest

`func NewFairShareSimulationRequest(nodePool FairShareSimulationNodePool, assumptions FairShareSimulationAssumptions, orgUnitType string, orgUnits []FairShareSimulationOrgUnit, ) *FairShareSimulationRequest`

NewFairShareSimulationRequest instantiates a new FairShareSimulationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationRequestWithDefaults

`func NewFairShareSimulationRequestWithDefaults() *FairShareSimulationRequest`

NewFairShareSimulationRequestWithDefaults instantiates a new FairShareSimulationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePool

`func (o *FairShareSimulationRequest) GetNodePool() FairShareSimulationNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *FairShareSimulationRequest) GetNodePoolOk() (*FairShareSimulationNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *FairShareSimulationRequest) SetNodePool(v FairShareSimulationNodePool)`

SetNodePool sets NodePool field to given value.


### GetAssumptions

`func (o *FairShareSimulationRequest) GetAssumptions() FairShareSimulationAssumptions`

GetAssumptions returns the Assumptions field if non-nil, zero value otherwise.

### GetAssumptionsOk

`func (o *FairShareSimulationRequest) GetAssumptionsOk() (*FairShareSimulationAssumptions, bool)`

GetAssumptionsOk returns a tuple with the Assumptions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssumptions

`func (o *FairShareSimulationRequest) SetAssumptions(v FairShareSimulationAssumptions)`

SetAssumptions sets Assumptions field to given value.


### GetOrgUnitType

`func (o *FairShareSimulationRequest) GetOrgUnitType() string`

GetOrgUnitType returns the OrgUnitType field if non-nil, zero value otherwise.

### GetOrgUnitTypeOk

`func (o *FairShareSimulationRequest) GetOrgUnitTypeOk() (*string, bool)`

GetOrgUnitTypeOk returns a tuple with the OrgUnitType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnitType

`func (o *FairShareSimulationRequest) SetOrgUnitType(v string)`

SetOrgUnitType sets OrgUnitType field to given value.


### GetOrgUnits

`func (o *FairShareSimulationRequest) GetOrgUnits() []FairShareSimulationOrgUnit`

GetOrgUnits returns the OrgUnits field if non-nil, zero value otherwise.

### GetOrgUnitsOk

`func (o *FairShareSimulationRequest) GetOrgUnitsOk() (*[]FairShareSimulationOrgUnit, bool)`

GetOrgUnitsOk returns a tuple with the OrgUnits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgUnits

`func (o *FairShareSimulationRequest) SetOrgUnits(v []FairShareSimulationOrgUnit)`

SetOrgUnits sets OrgUnits field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


