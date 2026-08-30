# ResourcesMemory

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | Pointer to **NullableFloat32** | The project resource allocation that it is deserved to get in case the cluster has those resources | [optional] 
**Limit** | Pointer to **NullableFloat32** | Maximum amount of resources the project can get | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat32** | The priority the project gets for over quota resources | [optional] 
**Units** | Pointer to **string** |  | [optional] 

## Methods

### NewResourcesMemory

`func NewResourcesMemory() *ResourcesMemory`

NewResourcesMemory instantiates a new ResourcesMemory object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesMemoryWithDefaults

`func NewResourcesMemoryWithDefaults() *ResourcesMemory`

NewResourcesMemoryWithDefaults instantiates a new ResourcesMemory object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *ResourcesMemory) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *ResourcesMemory) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *ResourcesMemory) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.

### HasDeserved

`func (o *ResourcesMemory) HasDeserved() bool`

HasDeserved returns a boolean if a field has been set.

### SetDeservedNil

`func (o *ResourcesMemory) SetDeservedNil(b bool)`

 SetDeservedNil sets the value for Deserved to be an explicit nil

### UnsetDeserved
`func (o *ResourcesMemory) UnsetDeserved()`

UnsetDeserved ensures that no value is present for Deserved, not even an explicit nil
### GetLimit

`func (o *ResourcesMemory) GetLimit() float32`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *ResourcesMemory) GetLimitOk() (*float32, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *ResourcesMemory) SetLimit(v float32)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *ResourcesMemory) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### SetLimitNil

`func (o *ResourcesMemory) SetLimitNil(b bool)`

 SetLimitNil sets the value for Limit to be an explicit nil

### UnsetLimit
`func (o *ResourcesMemory) UnsetLimit()`

UnsetLimit ensures that no value is present for Limit, not even an explicit nil
### GetOverQuotaWeight

`func (o *ResourcesMemory) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *ResourcesMemory) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *ResourcesMemory) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *ResourcesMemory) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *ResourcesMemory) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *ResourcesMemory) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil
### GetUnits

`func (o *ResourcesMemory) GetUnits() string`

GetUnits returns the Units field if non-nil, zero value otherwise.

### GetUnitsOk

`func (o *ResourcesMemory) GetUnitsOk() (*string, bool)`

GetUnitsOk returns a tuple with the Units field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUnits

`func (o *ResourcesMemory) SetUnits(v string)`

SetUnits sets Units field to given value.

### HasUnits

`func (o *ResourcesMemory) HasUnits() bool`

HasUnits returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


