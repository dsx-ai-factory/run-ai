# NonNullResourceV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Deserved** | **float32** | The amount of resources guaranteed to be allocated in case the cluster has those resources. | 
**MaxAllowed** | Pointer to **NullableFloat32** | Maximum amount of resources that can be allocated. If equal to deserved, no over-quota will be allowed. Use \&quot;-1\&quot; for unlimited over quota. | [optional] 
**OverQuotaWeight** | Pointer to **NullableFloat32** | The priority for over quota resources. | [optional] 

## Methods

### NewNonNullResourceV1

`func NewNonNullResourceV1(deserved float32, ) *NonNullResourceV1`

NewNonNullResourceV1 instantiates a new NonNullResourceV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNonNullResourceV1WithDefaults

`func NewNonNullResourceV1WithDefaults() *NonNullResourceV1`

NewNonNullResourceV1WithDefaults instantiates a new NonNullResourceV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDeserved

`func (o *NonNullResourceV1) GetDeserved() float32`

GetDeserved returns the Deserved field if non-nil, zero value otherwise.

### GetDeservedOk

`func (o *NonNullResourceV1) GetDeservedOk() (*float32, bool)`

GetDeservedOk returns a tuple with the Deserved field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeserved

`func (o *NonNullResourceV1) SetDeserved(v float32)`

SetDeserved sets Deserved field to given value.


### GetMaxAllowed

`func (o *NonNullResourceV1) GetMaxAllowed() float32`

GetMaxAllowed returns the MaxAllowed field if non-nil, zero value otherwise.

### GetMaxAllowedOk

`func (o *NonNullResourceV1) GetMaxAllowedOk() (*float32, bool)`

GetMaxAllowedOk returns a tuple with the MaxAllowed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowed

`func (o *NonNullResourceV1) SetMaxAllowed(v float32)`

SetMaxAllowed sets MaxAllowed field to given value.

### HasMaxAllowed

`func (o *NonNullResourceV1) HasMaxAllowed() bool`

HasMaxAllowed returns a boolean if a field has been set.

### SetMaxAllowedNil

`func (o *NonNullResourceV1) SetMaxAllowedNil(b bool)`

 SetMaxAllowedNil sets the value for MaxAllowed to be an explicit nil

### UnsetMaxAllowed
`func (o *NonNullResourceV1) UnsetMaxAllowed()`

UnsetMaxAllowed ensures that no value is present for MaxAllowed, not even an explicit nil
### GetOverQuotaWeight

`func (o *NonNullResourceV1) GetOverQuotaWeight() float32`

GetOverQuotaWeight returns the OverQuotaWeight field if non-nil, zero value otherwise.

### GetOverQuotaWeightOk

`func (o *NonNullResourceV1) GetOverQuotaWeightOk() (*float32, bool)`

GetOverQuotaWeightOk returns a tuple with the OverQuotaWeight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverQuotaWeight

`func (o *NonNullResourceV1) SetOverQuotaWeight(v float32)`

SetOverQuotaWeight sets OverQuotaWeight field to given value.

### HasOverQuotaWeight

`func (o *NonNullResourceV1) HasOverQuotaWeight() bool`

HasOverQuotaWeight returns a boolean if a field has been set.

### SetOverQuotaWeightNil

`func (o *NonNullResourceV1) SetOverQuotaWeightNil(b bool)`

 SetOverQuotaWeightNil sets the value for OverQuotaWeight to be an explicit nil

### UnsetOverQuotaWeight
`func (o *NonNullResourceV1) UnsetOverQuotaWeight()`

UnsetOverQuotaWeight ensures that no value is present for OverQuotaWeight, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


