# FairShareSimulationNodePool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Capacity** | [**FairShareSimulationResources**](FairShareSimulationResources.md) | GPUs to split. 0 is valid; quotas are still granted, as the scheduler does, and nothing is given above them. | 
**SchedulingCapabilities** | [**FairShareSimulationSchedulingCapabilities**](FairShareSimulationSchedulingCapabilities.md) |  | 

## Methods

### NewFairShareSimulationNodePool

`func NewFairShareSimulationNodePool(capacity FairShareSimulationResources, schedulingCapabilities FairShareSimulationSchedulingCapabilities, ) *FairShareSimulationNodePool`

NewFairShareSimulationNodePool instantiates a new FairShareSimulationNodePool object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationNodePoolWithDefaults

`func NewFairShareSimulationNodePoolWithDefaults() *FairShareSimulationNodePool`

NewFairShareSimulationNodePoolWithDefaults instantiates a new FairShareSimulationNodePool object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCapacity

`func (o *FairShareSimulationNodePool) GetCapacity() FairShareSimulationResources`

GetCapacity returns the Capacity field if non-nil, zero value otherwise.

### GetCapacityOk

`func (o *FairShareSimulationNodePool) GetCapacityOk() (*FairShareSimulationResources, bool)`

GetCapacityOk returns a tuple with the Capacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapacity

`func (o *FairShareSimulationNodePool) SetCapacity(v FairShareSimulationResources)`

SetCapacity sets Capacity field to given value.


### GetSchedulingCapabilities

`func (o *FairShareSimulationNodePool) GetSchedulingCapabilities() FairShareSimulationSchedulingCapabilities`

GetSchedulingCapabilities returns the SchedulingCapabilities field if non-nil, zero value otherwise.

### GetSchedulingCapabilitiesOk

`func (o *FairShareSimulationNodePool) GetSchedulingCapabilitiesOk() (*FairShareSimulationSchedulingCapabilities, bool)`

GetSchedulingCapabilitiesOk returns a tuple with the SchedulingCapabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingCapabilities

`func (o *FairShareSimulationNodePool) SetSchedulingCapabilities(v FairShareSimulationSchedulingCapabilities)`

SetSchedulingCapabilities sets SchedulingCapabilities field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


