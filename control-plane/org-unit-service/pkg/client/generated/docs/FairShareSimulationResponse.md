# FairShareSimulationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OrgUnits** | [**[]FairShareSimulationResult**](FairShareSimulationResult.md) | One result per requested org unit, in the request&#39;s order. | 

## Methods

### NewFairShareSimulationResponse

`func NewFairShareSimulationResponse(orgUnits []FairShareSimulationResult, ) *FairShareSimulationResponse`

NewFairShareSimulationResponse instantiates a new FairShareSimulationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationResponseWithDefaults

`func NewFairShareSimulationResponseWithDefaults() *FairShareSimulationResponse`

NewFairShareSimulationResponseWithDefaults instantiates a new FairShareSimulationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

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


