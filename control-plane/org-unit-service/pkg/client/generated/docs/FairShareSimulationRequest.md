# FairShareSimulationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | Pointer to [**NullableFairShareSimulationRequestNodePool**](FairShareSimulationRequestNodePool.md) |  | [optional] 
**OrgUnits** | Pointer to [**[]FairShareSimulationOrgUnit**](FairShareSimulationOrgUnit.md) | Org units to simulate with other values than their stored ones. Others keep their stored values and ask for the most GPUs they can get. | [optional] 

## Methods

### NewFairShareSimulationRequest

`func NewFairShareSimulationRequest() *FairShareSimulationRequest`

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

`func (o *FairShareSimulationRequest) GetNodePool() FairShareSimulationRequestNodePool`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *FairShareSimulationRequest) GetNodePoolOk() (*FairShareSimulationRequestNodePool, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *FairShareSimulationRequest) SetNodePool(v FairShareSimulationRequestNodePool)`

SetNodePool sets NodePool field to given value.

### HasNodePool

`func (o *FairShareSimulationRequest) HasNodePool() bool`

HasNodePool returns a boolean if a field has been set.

### SetNodePoolNil

`func (o *FairShareSimulationRequest) SetNodePoolNil(b bool)`

 SetNodePoolNil sets the value for NodePool to be an explicit nil

### UnsetNodePool
`func (o *FairShareSimulationRequest) UnsetNodePool()`

UnsetNodePool ensures that no value is present for NodePool, not even an explicit nil
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

### HasOrgUnits

`func (o *FairShareSimulationRequest) HasOrgUnits() bool`

HasOrgUnits returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


