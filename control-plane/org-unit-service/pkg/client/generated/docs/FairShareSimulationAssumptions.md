# FairShareSimulationAssumptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DefaultDemand** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | GPUs every org unit without a demand in the request asks for, capped by its limit. The node pool form sends half the node pool&#39;s capacity. In the node pool form it is also a department&#39;s demand; in the department and project forms a department asks for the sum of its projects&#39; demand.  | 
**OtherDepartments** | **string** | What departments other than focusDepartmentId ask for. &#x60;default&#x60;: the default demand. &#x60;idle&#x60;: nothing, so the focus department can take the whole node pool up to its limit. The department and project forms send &#x60;idle&#x60; and say so in a footnote.  | 
**FocusDepartmentId** | Pointer to **NullableString** | The department the requested org units belong to. Required with &#x60;idle&#x60; and when orgUnitType is project; the requested org units must belong to it.  | [optional] 

## Methods

### NewFairShareSimulationAssumptions

`func NewFairShareSimulationAssumptions(defaultDemand FairShareSimulationResources, otherDepartments string, ) *FairShareSimulationAssumptions`

NewFairShareSimulationAssumptions instantiates a new FairShareSimulationAssumptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationAssumptionsWithDefaults

`func NewFairShareSimulationAssumptionsWithDefaults() *FairShareSimulationAssumptions`

NewFairShareSimulationAssumptionsWithDefaults instantiates a new FairShareSimulationAssumptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaultDemand

`func (o *FairShareSimulationAssumptions) GetDefaultDemand() FairShareSimulationResources`

GetDefaultDemand returns the DefaultDemand field if non-nil, zero value otherwise.

### GetDefaultDemandOk

`func (o *FairShareSimulationAssumptions) GetDefaultDemandOk() (*FairShareSimulationResources, bool)`

GetDefaultDemandOk returns a tuple with the DefaultDemand field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultDemand

`func (o *FairShareSimulationAssumptions) SetDefaultDemand(v FairShareSimulationResources)`

SetDefaultDemand sets DefaultDemand field to given value.


### GetOtherDepartments

`func (o *FairShareSimulationAssumptions) GetOtherDepartments() string`

GetOtherDepartments returns the OtherDepartments field if non-nil, zero value otherwise.

### GetOtherDepartmentsOk

`func (o *FairShareSimulationAssumptions) GetOtherDepartmentsOk() (*string, bool)`

GetOtherDepartmentsOk returns a tuple with the OtherDepartments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOtherDepartments

`func (o *FairShareSimulationAssumptions) SetOtherDepartments(v string)`

SetOtherDepartments sets OtherDepartments field to given value.


### GetFocusDepartmentId

`func (o *FairShareSimulationAssumptions) GetFocusDepartmentId() string`

GetFocusDepartmentId returns the FocusDepartmentId field if non-nil, zero value otherwise.

### GetFocusDepartmentIdOk

`func (o *FairShareSimulationAssumptions) GetFocusDepartmentIdOk() (*string, bool)`

GetFocusDepartmentIdOk returns a tuple with the FocusDepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFocusDepartmentId

`func (o *FairShareSimulationAssumptions) SetFocusDepartmentId(v string)`

SetFocusDepartmentId sets FocusDepartmentId field to given value.

### HasFocusDepartmentId

`func (o *FairShareSimulationAssumptions) HasFocusDepartmentId() bool`

HasFocusDepartmentId returns a boolean if a field has been set.

### SetFocusDepartmentIdNil

`func (o *FairShareSimulationAssumptions) SetFocusDepartmentIdNil(b bool)`

 SetFocusDepartmentIdNil sets the value for FocusDepartmentId to be an explicit nil

### UnsetFocusDepartmentId
`func (o *FairShareSimulationAssumptions) UnsetFocusDepartmentId()`

UnsetFocusDepartmentId ensures that no value is present for FocusDepartmentId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


