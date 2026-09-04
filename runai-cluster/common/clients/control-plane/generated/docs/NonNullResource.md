# NonNullResource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | **float32** | The project resource allocation that it is deserved to get in case the cluster has those resources | 
**Limit** | Pointer to **NullableFloat32** | Maximum amount of resources the project can get | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat32** | The priority the project gets for over quota resources | [optional] 

## Methods

### NewNonNullResource

`func NewNonNullResource(deserved float32, ) *NonNullResource`

NewNonNullResource instantiates a new NonNullResource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNonNullResourceWithDefaults

`func NewNonNullResourceWithDefaults() *NonNullResource`

NewNonNullResourceWithDefaults instantiates a new NonNullResource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *NonNullResource) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *NonNullResource) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *NonNullResource) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.


### GetLimit

`func (o *NonNullResource) GetLimit() float32`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *NonNullResource) GetLimitOk() (*float32, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *NonNullResource) SetLimit(v float32)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *NonNullResource) HasLimit() bool`

HasLimit returns a boolean if a field has been set.

### SetLimitNil

`func (o *NonNullResource) SetLimitNil(b bool)`

 SetLimitNil sets the value for Limit to be an explicit nil

### UnsetLimit
`func (o *NonNullResource) UnsetLimit()`

UnsetLimit ensures that no value is present for Limit, not even an explicit nil
### GetOverQuotaWeight

`func (o *NonNullResource) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *NonNullResource) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *NonNullResource) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *NonNullResource) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *NonNullResource) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *NonNullResource) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


