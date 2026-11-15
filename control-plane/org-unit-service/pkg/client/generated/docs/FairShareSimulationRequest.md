# FairShareSimulationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | [**NullableFairShareSimulationNodePool**](FairShareSimulationNodePool.md) | Required; null is rejected. | 
**OrgUnitType** | **string** | The type of the requested org units. One type per request. | 
**Parent** | Pointer to [**NullableFairShareSimulationParent**](FairShareSimulationParent.md) | Required for projects; not allowed for departments. | [optional] 
**OrgUnits** | [**[]FairShareSimulationOrgUnit**](FairShareSimulationOrgUnit.md) | The org units to return, with their unsaved changes. Send the rows the screen shows and every edited row: an edit that isn&#39;t sent is lost.  | 

## Methods

### NewFairShareSimulationRequest

`func NewFairShareSimulationRequest(nodePool NullableFairShareSimulationNodePool, orgUnitType string, orgUnits []FairShareSimulationOrgUnit, ) *FairShareSimulationRequest`

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


### SetNodePoolNil

`func (o *FairShareSimulationRequest) SetNodePoolNil(b bool)`

 SetNodePoolNil sets the value for NodePool to be an explicit nil

### UnsetNodePool
`func (o *FairShareSimulationRequest) UnsetNodePool()`

UnsetNodePool ensures that no value is present for NodePool, not even an explicit nil
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


### GetParent

`func (o *FairShareSimulationRequest) GetParent() FairShareSimulationParent`

GetParent returns the Parent field if non-nil, zero value otherwise.

### GetParentOk

`func (o *FairShareSimulationRequest) GetParentOk() (*FairShareSimulationParent, bool)`

GetParentOk returns a tuple with the Parent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParent

`func (o *FairShareSimulationRequest) SetParent(v FairShareSimulationParent)`

SetParent sets Parent field to given value.

### HasParent

`func (o *FairShareSimulationRequest) HasParent() bool`

HasParent returns a boolean if a field has been set.

### SetParentNil

`func (o *FairShareSimulationRequest) SetParentNil(b bool)`

 SetParentNil sets the value for Parent to be an explicit nil

### UnsetParent
`func (o *FairShareSimulationRequest) UnsetParent()`

UnsetParent ensures that no value is present for Parent, not even an explicit nil
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


