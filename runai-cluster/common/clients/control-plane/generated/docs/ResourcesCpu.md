# ResourcesCpu

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | Pointer to **NullableFloat32** | The project resource allocation that it is deserved to get in case the cluster has those resources | [optional] 
**Limit** | Pointer to **NullableFloat32** | Maximum amount of resources the project can get | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat32** | The priority the project gets for over quota resources | [optional] 

## Methods

### NewResourcesCpu

`func NewResourcesCpu() *ResourcesCpu`

NewResourcesCpu instantiates a new ResourcesCpu object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcesCpuWithDefaults

`func NewResourcesCpuWithDefaults() *ResourcesCpu`

NewResourcesCpuWithDefaults instantiates a new ResourcesCpu object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *ResourcesCpu) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *ResourcesCpu) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *ResourcesCpu) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.

### HasDeserved

`func (o *ResourcesCpu) HasDeserved() bool`

HasDeserved returns a boolean if a field has been set.

### SetDeservedNil

`func (o *ResourcesCpu) SetDeservedNil(b bool)`

 SetDeservedNil sets the value for Deserved to be an explicit nil

### UnsetDeserved
`func (o *ResourcesCpu) UnsetDeserved()`

UnsetDeserved ensures that no value is present for Deserved, not even an explicit nil
### GetLimit

`func (o *ResourcesCpu) GetLimit() float32`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *ResourcesCpu) GetLimitOk() (*float32, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *ResourcesCpu) SetLimit(v float32)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *ResourcesCpu) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### SetLimitNil

`func (o *ResourcesCpu) SetLimitNil(b bool)`

 SetLimitNil sets the value for Limit to be an explicit nil

### UnsetLimit
`func (o *ResourcesCpu) UnsetLimit()`

UnsetLimit ensures that no value is present for Limit, not even an explicit nil
### GetOverQuotaWeight

`func (o *ResourcesCpu) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *ResourcesCpu) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *ResourcesCpu) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *ResourcesCpu) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *ResourcesCpu) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *ResourcesCpu) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


