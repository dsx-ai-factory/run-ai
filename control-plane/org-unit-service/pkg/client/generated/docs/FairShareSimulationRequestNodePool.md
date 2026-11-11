# FairShareSimulationRequestNodePool

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Capacity** | Pointer to [**NullableQuotaStatusResource**](QuotaStatusResource.md) | Resources to split. Defaults to the node pool&#39;s capacity. | [optional] 
**SchedulingCapabilities** | Pointer to [**NullableFairShareSimulationSchedulingCapabilities**](FairShareSimulationSchedulingCapabilities.md) |  | [optional] 

## Methods

### NewFairShareSimulationRequestNodePool

`func NewFairShareSimulationRequestNodePool() *FairShareSimulationRequestNodePool`

NewFairShareSimulationRequestNodePool instantiates a new FairShareSimulationRequestNodePool object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFairShareSimulationRequestNodePoolWithDefaults

`func NewFairShareSimulationRequestNodePoolWithDefaults() *FairShareSimulationRequestNodePool`

NewFairShareSimulationRequestNodePoolWithDefaults instantiates a new FairShareSimulationRequestNodePool object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCapacity

`func (o *FairShareSimulationRequestNodePool) GetCapacity() QuotaStatusResource`

GetCapacity returns the Capacity field if non-nil, zero value otherwise.

### GetCapacityOk

`func (o *FairShareSimulationRequestNodePool) GetCapacityOk() (*QuotaStatusResource, bool)`

GetCapacityOk returns a tuple with the Capacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapacity

`func (o *FairShareSimulationRequestNodePool) SetCapacity(v QuotaStatusResource)`

SetCapacity sets Capacity field to given value.

### HasCapacity

`func (o *FairShareSimulationRequestNodePool) HasCapacity() bool`

HasCapacity returns a boolean if a field has been set.

### SetCapacityNil

`func (o *FairShareSimulationRequestNodePool) SetCapacityNil(b bool)`

 SetCapacityNil sets the value for Capacity to be an explicit nil

### UnsetCapacity
`func (o *FairShareSimulationRequestNodePool) UnsetCapacity()`

UnsetCapacity ensures that no value is present for Capacity, not even an explicit nil
### GetSchedulingCapabilities

`func (o *FairShareSimulationRequestNodePool) GetSchedulingCapabilities() FairShareSimulationSchedulingCapabilities`

GetSchedulingCapabilities returns the SchedulingCapabilities field if non-nil, zero value otherwise.

### GetSchedulingCapabilitiesOk

`func (o *FairShareSimulationRequestNodePool) GetSchedulingCapabilitiesOk() (*FairShareSimulationSchedulingCapabilities, bool)`

GetSchedulingCapabilitiesOk returns a tuple with the SchedulingCapabilities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedulingCapabilities

`func (o *FairShareSimulationRequestNodePool) SetSchedulingCapabilities(v FairShareSimulationSchedulingCapabilities)`

SetSchedulingCapabilities sets SchedulingCapabilities field to given value.

### HasSchedulingCapabilities

`func (o *FairShareSimulationRequestNodePool) HasSchedulingCapabilities() bool`

HasSchedulingCapabilities returns a boolean if a field has been set.

### SetSchedulingCapabilitiesNil

`func (o *FairShareSimulationRequestNodePool) SetSchedulingCapabilitiesNil(b bool)`

 SetSchedulingCapabilitiesNil sets the value for SchedulingCapabilities to be an explicit nil

### UnsetSchedulingCapabilities
`func (o *FairShareSimulationRequestNodePool) UnsetSchedulingCapabilities()`

UnsetSchedulingCapabilities ensures that no value is present for SchedulingCapabilities, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


