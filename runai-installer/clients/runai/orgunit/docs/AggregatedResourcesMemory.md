# AggregatedResourcesMemory

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | Pointer to **NullableFloat64** | The project resource allocation that it is deserved to get in case the cluster has those resources | [optional] 
**Units** | [**Units**](Units.md) |  | 

## Methods

### NewAggregatedResourcesMemory

`func NewAggregatedResourcesMemory(units Units, ) *AggregatedResourcesMemory`

NewAggregatedResourcesMemory instantiates a new AggregatedResourcesMemory object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAggregatedResourcesMemoryWithDefaults

`func NewAggregatedResourcesMemoryWithDefaults() *AggregatedResourcesMemory`

NewAggregatedResourcesMemoryWithDefaults instantiates a new AggregatedResourcesMemory object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *AggregatedResourcesMemory) GetDeserved() float64`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *AggregatedResourcesMemory) GetDeservedOk() (*float64, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *AggregatedResourcesMemory) SetDeserved(v float64)`

SetDeserved sets Deserved field to given value.

### HasDeserved

`func (o *AggregatedResourcesMemory) HasDeserved() bool`

HasDeserved returns a boolean if a field has been set.

### SetDeservedNil

`func (o *AggregatedResourcesMemory) SetDeservedNil(b bool)`

 SetDeservedNil sets the value for Deserved to be an explicit nil

### UnsetDeserved
`func (o *AggregatedResourcesMemory) UnsetDeserved()`

UnsetDeserved ensures that no value is present for Deserved, not even an explicit nil
### GetUnits

`func (o *AggregatedResourcesMemory) GetUnits() Units`

GetUnits returns the Units field if non-nil, zero value otherwise.

### GetUnitsOk

`func (o *AggregatedResourcesMemory) GetUnitsOk() (*Units, bool)`

GetUnitsOk returns a tuple with the Units field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnits

`func (o *AggregatedResourcesMemory) SetUnits(v Units)`

SetUnits sets Units field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


